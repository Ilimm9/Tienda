package application

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"tienda/backend/internal/domain"
)

var ErrInvalidCredentials = errors.New("credenciales inválidas")
var ErrAccountUnavailable = errors.New("cuenta no disponible")
var ErrEmailAlreadyExists = errors.New("el correo electrónico ya está registrado")

type UserRepository interface {
	FindByEmail(email string) (*domain.Usuario, error)
	Save(user *domain.Usuario) error
	CreateAccount(user *domain.Usuario, profile *domain.PerfilUsuario) error
}

type AuthService struct{ users UserRepository }

func NewAuthService(users UserRepository) *AuthService { return &AuthService{users: users} }

func (s *AuthService) Login(email, password string) (*domain.Usuario, error) {
	user, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	now := time.Now()
	if user.DeshabilitadoEn != nil || user.Estado != "activo" || (user.BloqueadoHasta != nil && user.BloqueadoHasta.After(now)) {
		return nil, ErrAccountUnavailable
	}
	if bcrypt.CompareHashAndPassword([]byte(user.HashContrasena), []byte(password)) != nil {
		user.IntentosInicioSesionFallidos++
		if user.IntentosInicioSesionFallidos >= 5 {
			until := now.Add(15 * time.Minute)
			user.BloqueadoHasta = &until
		}
		_ = s.users.Save(user)
		return nil, ErrInvalidCredentials
	}
	user.IntentosInicioSesionFallidos = 0
	user.BloqueadoHasta = nil
	user.UltimoInicioSesionEn = &now
	if err := s.users.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Register(fullName, email, phone, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := s.users.FindByEmail(email); err == nil {
		return ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return errors.New("el nombre completo es obligatorio")
	}
	names := parts[0]
	surnames := ""
	if len(parts) > 1 {
		surnames = strings.Join(parts[1:], " ")
	}

	user := &domain.Usuario{Correo: email, HashContrasena: string(hash), Estado: "activo"}
	profile := &domain.PerfilUsuario{UsuarioID: user.ID, Nombres: names, Apellidos: surnames}
	if phone != "" {
		cleanPhone := strings.TrimSpace(phone)
		profile.Telefono = &cleanPhone
	}
	if err := s.users.CreateAccount(user, profile); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}
