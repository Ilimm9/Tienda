package cuenta

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrVerificationInvalid = errors.New("el código no es válido o expiró")
	ErrVerificationTooSoon = errors.New("debes esperar antes de solicitar otro código")
	ErrVerificationLimited = errors.New("se alcanzó el límite temporal de envíos")
	ErrEmailDelivery       = errors.New("no fue posible enviar el correo de verificación")
)

type VerificationRepository interface {
	FindChallenge(id uuid.UUID) (*cuentadomain.DesafioAutenticacion, error)
	IssueChallenge(challenge *cuentadomain.DesafioAutenticacion) error
	IssueChallengeForPendingAccount(user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario, challenge *cuentadomain.DesafioAutenticacion) error
	CountChallengesByUserSince(userID uuid.UUID, since time.Time) (int64, error)
	CountChallengesByIPSince(ip string, since time.Time) (int64, error)
	LastChallengeByUser(userID uuid.UUID) (*cuentadomain.DesafioAutenticacion, error)
	IncrementFailedAttempts(id uuid.UUID, maxAttempts int) error
	ActivateUserWithChallenge(challengeID, userID uuid.UUID, now time.Time, maxAttempts int) error
}

type VerificationUserRepository interface {
	FindByEmail(email string) (*cuentadomain.Usuario, error)
	FindByID(id uuid.UUID) (*cuentadomain.Usuario, error)
	CreateAccount(user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario) error
}

type VerificationMailer interface {
	SendVerificationOTP(ctx context.Context, recipient, code string, expiresAt time.Time) error
}

type VerificationConfig struct {
	HMACSecret    string
	TTL           time.Duration
	MaxAttempts   int
	ResendWait    time.Duration
	HourlySendMax int
}

type RegistrationResult struct {
	ChallengeID      uuid.UUID
	MaskedEmail      string
	ResendAfter      time.Duration
	DeliveryDeferred bool
}

type VerificationService struct {
	users      VerificationUserRepository
	challenges VerificationRepository
	mailer     VerificationMailer
	config     VerificationConfig
	now        func() time.Time
}

func NewVerificationService(users VerificationUserRepository, challenges VerificationRepository, mailer VerificationMailer, cfg VerificationConfig) *VerificationService {
	return &VerificationService{users: users, challenges: challenges, mailer: mailer, config: cfg, now: time.Now}
}

func (s *VerificationService) Register(ctx context.Context, fullName, email, phone, password, ip string) (RegistrationResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegistrationResult{}, err
	}
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return RegistrationResult{}, errors.New("el nombre completo es obligatorio")
	}

	user, findErr := s.users.FindByEmail(email)
	if findErr == nil && (user.Estado != "pendiente_verificacion" || user.CorreoVerificadoEn != nil) {
		return RegistrationResult{ChallengeID: uuid.New(), MaskedEmail: maskEmail(email), ResendAfter: s.config.ResendWait}, nil
	}
	if findErr != nil {
		user = &cuentadomain.Usuario{Correo: email, HashContrasena: string(hash), Estado: "pendiente_verificacion"}
		profile := &cuentadomain.PerfilUsuario{Nombres: parts[0], Apellidos: strings.Join(parts[1:], " ")}
		if cleanPhone := strings.TrimSpace(phone); cleanPhone != "" {
			profile.Telefono = &cleanPhone
		}
		if err := s.users.CreateAccount(user, profile); err != nil {
			return RegistrationResult{}, err
		}
		return s.issueAndSend(ctx, user, ip, false)
	}
	user.HashContrasena = string(hash)
	profile := &cuentadomain.PerfilUsuario{UsuarioID: user.ID, Nombres: parts[0], Apellidos: strings.Join(parts[1:], " ")}
	if cleanPhone := strings.TrimSpace(phone); cleanPhone != "" {
		profile.Telefono = &cleanPhone
	}
	return s.issueAndSendForPendingAccount(ctx, user, profile, ip)
}

func (s *VerificationService) Resend(ctx context.Context, challengeID uuid.UUID, ip string) (RegistrationResult, error) {
	challenge, err := s.challenges.FindChallenge(challengeID)
	if err != nil {
		return RegistrationResult{ChallengeID: uuid.New(), MaskedEmail: "***@***", ResendAfter: s.config.ResendWait}, nil
	}
	user, err := s.users.FindByID(challenge.UsuarioID)
	if err != nil || user.Estado != "pendiente_verificacion" || user.CorreoVerificadoEn != nil {
		return RegistrationResult{ChallengeID: uuid.New(), MaskedEmail: "***@***", ResendAfter: s.config.ResendWait}, nil
	}
	return s.issueAndSend(ctx, user, ip, true)
}

func (s *VerificationService) Verify(ctx context.Context, challengeID uuid.UUID, code string) (*cuentadomain.Usuario, error) {
	challenge, err := s.challenges.FindChallenge(challengeID)
	if err != nil {
		return nil, ErrVerificationInvalid
	}
	now := s.now().UTC()
	if challenge.Proposito != cuentadomain.PropositoVerificacionCorreo || challenge.UsadoEn != nil || !challenge.ExpiraEn.After(now) || challenge.IntentosFallidos >= s.config.MaxAttempts {
		return nil, ErrVerificationInvalid
	}
	expected := otpHMAC(s.config.HMACSecret, challenge.ID, strings.TrimSpace(code))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(challenge.HashOTP)) != 1 {
		_ = s.challenges.IncrementFailedAttempts(challenge.ID, s.config.MaxAttempts)
		return nil, ErrVerificationInvalid
	}
	if err := s.challenges.ActivateUserWithChallenge(challenge.ID, challenge.UsuarioID, now, s.config.MaxAttempts); err != nil {
		return nil, ErrVerificationInvalid
	}
	user, err := s.users.FindByID(challenge.UsuarioID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *VerificationService) issueAndSend(ctx context.Context, user *cuentadomain.Usuario, ip string, enforceWait bool) (RegistrationResult, error) {
	return s.issueAndSendInternal(ctx, user, nil, ip, enforceWait)
}

func (s *VerificationService) issueAndSendForPendingAccount(ctx context.Context, user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario, ip string) (RegistrationResult, error) {
	return s.issueAndSendInternal(ctx, user, profile, ip, true)
}

func (s *VerificationService) issueAndSendInternal(ctx context.Context, user *cuentadomain.Usuario, profile *cuentadomain.PerfilUsuario, ip string, enforceWait bool) (RegistrationResult, error) {
	now := s.now().UTC()
	if err := s.ensureMayIssue(user, ip, enforceWait); err != nil {
		return RegistrationResult{}, err
	}
	code, err := generateOTP()
	if err != nil {
		return RegistrationResult{}, err
	}
	challenge := &cuentadomain.DesafioAutenticacion{
		UsuarioID: user.ID, Proposito: cuentadomain.PropositoVerificacionCorreo,
		DireccionIP: truncate(strings.TrimSpace(ip), 45), UltimoEnvioEn: now, ExpiraEn: now.Add(s.config.TTL),
	}
	challenge.ID = uuid.New()
	challenge.HashOTP = otpHMAC(s.config.HMACSecret, challenge.ID, code)
	if profile != nil {
		err = s.challenges.IssueChallengeForPendingAccount(user, profile, challenge)
	} else {
		err = s.challenges.IssueChallenge(challenge)
	}
	if err != nil {
		return RegistrationResult{}, err
	}
	result := RegistrationResult{ChallengeID: challenge.ID, MaskedEmail: maskEmail(user.Correo), ResendAfter: s.config.ResendWait}
	if err := s.mailer.SendVerificationOTP(ctx, user.Correo, code, challenge.ExpiraEn); err != nil {
		result.DeliveryDeferred = true
		return result, ErrEmailDelivery
	}
	return result, nil
}

func (s *VerificationService) ensureMayIssue(user *cuentadomain.Usuario, ip string, enforceWait bool) error {
	now := s.now().UTC()
	if enforceWait {
		last, err := s.challenges.LastChallengeByUser(user.ID)
		if err == nil && last.UltimoEnvioEn.Add(s.config.ResendWait).After(now) {
			return ErrVerificationTooSoon
		}
	}
	since := now.Add(-time.Hour)
	userCount, err := s.challenges.CountChallengesByUserSince(user.ID, since)
	if err != nil {
		return err
	}
	ipCount, err := s.challenges.CountChallengesByIPSince(ip, since)
	if err != nil {
		return err
	}
	if userCount >= int64(s.config.HourlySendMax) || ipCount >= int64(s.config.HourlySendMax) {
		return ErrVerificationLimited
	}
	return nil
}

func otpHMAC(secret string, challengeID uuid.UUID, code string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(challengeID.String() + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateOTP() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "***@***"
	}
	visible := string([]rune(parts[0])[:1])
	return visible + "***@" + parts[1]
}
