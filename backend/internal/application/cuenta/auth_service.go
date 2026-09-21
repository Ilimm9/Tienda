package cuenta

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	cuentadomain "tienda/backend/internal/domain/cuenta"
)

var ErrInvalidCredentials = errors.New("credenciales inválidas")
var ErrAccountUnavailable = errors.New("cuenta no disponible")

type UserRepository interface {
	FindByEmail(email string) (*cuentadomain.Usuario, error)
	Save(user *cuentadomain.Usuario) error
}

type AuthService struct{ users UserRepository }

func NewAuthService(users UserRepository) *AuthService { return &AuthService{users: users} }

func (s *AuthService) Login(email, password string) (*cuentadomain.Usuario, error) {
	user, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	now := time.Now()
	if user.DeshabilitadoEn != nil || user.Estado != "activo" || user.CorreoVerificadoEn == nil || (user.BloqueadoHasta != nil && user.BloqueadoHasta.After(now)) {
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
