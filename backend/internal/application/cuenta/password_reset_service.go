package cuenta

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordResetInvalid = errors.New("el enlace no es válido o expiró")
	ErrPasswordPolicy       = errors.New("la contraseña debe tener al menos 8 caracteres y como máximo 72 bytes")
)

type PasswordResetRepository interface {
	FindChallenge(id uuid.UUID) (*cuentadomain.DesafioAutenticacion, error)
	IssueChallenge(challenge *cuentadomain.DesafioAutenticacion) error
	CountChallengesByUserSince(userID uuid.UUID, since time.Time) (int64, error)
	CountChallengesByIPSince(ip string, since time.Time) (int64, error)
	LastChallengeByUserAndPurpose(userID uuid.UUID, purpose string) (*cuentadomain.DesafioAutenticacion, error)
	IncrementFailedAttempts(id uuid.UUID, maxAttempts int) error
	ResetPasswordWithChallenge(challengeID, userID uuid.UUID, passwordHash string, now time.Time, maxAttempts int) error
}

type PasswordResetUserRepository interface {
	FindByEmail(email string) (*cuentadomain.Usuario, error)
	FindByID(id uuid.UUID) (*cuentadomain.Usuario, error)
}

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, recipient, challengeID, token string, expiresAt time.Time) error
	SendPasswordChanged(ctx context.Context, recipient string) error
}

type PasswordResetConfig struct {
	HMACSecret    string
	TTL           time.Duration
	MaxAttempts   int
	ResendWait    time.Duration
	HourlySendMax int
}

// PasswordResetService emite enlaces de un solo uso. Solicitar recuperación
// nunca revela si la cuenta existe: todos los caminos esperados devuelven nil.
type PasswordResetService struct {
	users      PasswordResetUserRepository
	challenges PasswordResetRepository
	mailer     PasswordResetMailer
	config     PasswordResetConfig
	now        func() time.Time
}

func NewPasswordResetService(users PasswordResetUserRepository, challenges PasswordResetRepository, mailer PasswordResetMailer, cfg PasswordResetConfig) *PasswordResetService {
	return &PasswordResetService{users: users, challenges: challenges, mailer: mailer, config: cfg, now: time.Now}
}

func (s *PasswordResetService) Request(ctx context.Context, email, ip string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.users.FindByEmail(email)
	if err != nil || !canResetPassword(user) {
		return nil
	}
	now := s.now().UTC()
	ip = truncate(strings.TrimSpace(ip), 45)
	if limited, err := s.limited(user.ID, ip, now); err != nil {
		return err
	} else if limited {
		log.Printf("evento=recuperacion_limitada usuario=%s", user.ID)
		return nil
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	challenge := &cuentadomain.DesafioAutenticacion{
		ID: uuid.New(), UsuarioID: user.ID, Proposito: cuentadomain.PropositoRecuperacionContrasena,
		DireccionIP: ip, UltimoEnvioEn: now, ExpiraEn: now.Add(s.config.TTL),
	}
	challenge.HashOTP = otpHMAC(s.config.HMACSecret, challenge.ID, token)
	if err := s.challenges.IssueChallenge(challenge); err != nil {
		return err
	}
	// El fallo de entrega ya queda registrado por el servicio de correo; la respuesta sigue siendo uniforme.
	_ = s.mailer.SendPasswordReset(ctx, user.Correo, challenge.ID.String(), token, challenge.ExpiraEn)
	return nil
}

func (s *PasswordResetService) Reset(ctx context.Context, challengeID uuid.UUID, token, password string) error {
	if utf8.RuneCountInString(password) < 8 || len(password) > 72 {
		return ErrPasswordPolicy
	}
	challenge, err := s.challenges.FindChallenge(challengeID)
	if err != nil {
		return ErrPasswordResetInvalid
	}
	now := s.now().UTC()
	if challenge.Proposito != cuentadomain.PropositoRecuperacionContrasena || challenge.UsadoEn != nil || !challenge.ExpiraEn.After(now) || challenge.IntentosFallidos >= s.config.MaxAttempts {
		return ErrPasswordResetInvalid
	}
	expected := otpHMAC(s.config.HMACSecret, challenge.ID, strings.TrimSpace(token))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge.HashOTP)) != 1 {
		_ = s.challenges.IncrementFailedAttempts(challenge.ID, s.config.MaxAttempts)
		return ErrPasswordResetInvalid
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.challenges.ResetPasswordWithChallenge(challenge.ID, challenge.UsuarioID, string(hash), now, s.config.MaxAttempts); err != nil {
		return ErrPasswordResetInvalid
	}
	if user, err := s.users.FindByID(challenge.UsuarioID); err == nil {
		_ = s.mailer.SendPasswordChanged(ctx, user.Correo)
	}
	return nil
}

func (s *PasswordResetService) limited(userID uuid.UUID, ip string, now time.Time) (bool, error) {
	last, err := s.challenges.LastChallengeByUserAndPurpose(userID, cuentadomain.PropositoRecuperacionContrasena)
	if err == nil && last.UltimoEnvioEn.Add(s.config.ResendWait).After(now) {
		return true, nil
	}
	since := now.Add(-time.Hour)
	userCount, err := s.challenges.CountChallengesByUserSince(userID, since)
	if err != nil {
		return false, err
	}
	ipCount, err := s.challenges.CountChallengesByIPSince(ip, since)
	if err != nil {
		return false, err
	}
	return userCount >= int64(s.config.HourlySendMax) || ipCount >= int64(s.config.HourlySendMax), nil
}

func canResetPassword(user *cuentadomain.Usuario) bool {
	return user.Estado == "activo" && user.CorreoVerificadoEn != nil && user.DeshabilitadoEn == nil
}
