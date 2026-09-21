package cuenta

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
)

var ErrInvalidSession = errors.New("sesión inválida")

type SessionRepository interface {
	CreateSession(session *cuentadomain.SesionUsuario) error
	FindSessionByHash(hash string) (*cuentadomain.SesionUsuario, *cuentadomain.Usuario, error)
	TouchSession(id uuid.UUID, activityAt time.Time) error
	RevokeSession(hash string, revokedAt time.Time, reason string) error
	RevokeAllSessions(userID uuid.UUID, revokedAt time.Time, reason string) error
	DeleteInactiveSessions(before time.Time) error
}

type SessionConfig struct {
	Duration              time.Duration
	RememberDuration      time.Duration
	RememberIdleDuration  time.Duration
	ActivityTouchInterval time.Duration
}

type CreatedSession struct {
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}

type AuthenticatedSession struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Email    string
	CSRFHash string
}

type SessionService struct {
	repository SessionRepository
	config     SessionConfig
	now        func() time.Time
}

func NewSessionService(repository SessionRepository, cfg SessionConfig) *SessionService {
	return &SessionService{repository: repository, config: cfg, now: time.Now}
}

func (s *SessionService) Create(userID uuid.UUID, remember bool, ipAddress, userAgent string) (CreatedSession, error) {
	token, err := randomToken()
	if err != nil {
		return CreatedSession{}, err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return CreatedSession{}, err
	}
	now := s.now().UTC()
	duration := s.config.Duration
	if remember {
		duration = s.config.RememberDuration
	}
	session := &cuentadomain.SesionUsuario{
		UsuarioID:         userID,
		HashToken:         HashSessionToken(token),
		HashCSRF:          HashSessionToken(csrfToken),
		DireccionIP:       truncate(strings.TrimSpace(ipAddress), 45),
		AgenteUsuario:     truncate(strings.TrimSpace(userAgent), 512),
		Recordarme:        remember,
		EmitidoEn:         now,
		ExpiraEn:          now.Add(duration),
		UltimaActividadEn: now,
	}
	if err := s.repository.CreateSession(session); err != nil {
		return CreatedSession{}, err
	}
	return CreatedSession{Token: token, CSRFToken: csrfToken, ExpiresAt: session.ExpiraEn}, nil
}

func (s *SessionService) Authenticate(token string) (AuthenticatedSession, error) {
	if strings.TrimSpace(token) == "" {
		return AuthenticatedSession{}, ErrInvalidSession
	}
	session, user, err := s.repository.FindSessionByHash(HashSessionToken(token))
	if err != nil || session == nil || user == nil {
		return AuthenticatedSession{}, ErrInvalidSession
	}
	now := s.now().UTC()
	if session.RevocadoEn != nil || !session.ExpiraEn.After(now) || user.Estado != "activo" || user.DeshabilitadoEn != nil || user.CorreoVerificadoEn == nil {
		return AuthenticatedSession{}, ErrInvalidSession
	}
	if session.Recordarme && s.config.RememberIdleDuration > 0 && !session.UltimaActividadEn.Add(s.config.RememberIdleDuration).After(now) {
		return AuthenticatedSession{}, ErrInvalidSession
	}
	if s.config.ActivityTouchInterval > 0 && !session.UltimaActividadEn.Add(s.config.ActivityTouchInterval).After(now) {
		if err := s.repository.TouchSession(session.ID, now); err != nil {
			return AuthenticatedSession{}, ErrInvalidSession
		}
	}
	return AuthenticatedSession{ID: session.ID, UserID: user.ID, Email: user.Correo, CSRFHash: session.HashCSRF}, nil
}

func (s *SessionService) Revoke(token, reason string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	return s.repository.RevokeSession(HashSessionToken(token), s.now().UTC(), truncate(reason, 120))
}

func (s *SessionService) RevokeAll(userID uuid.UUID, reason string) error {
	return s.repository.RevokeAllSessions(userID, s.now().UTC(), truncate(reason, 120))
}

func (s *SessionService) CleanupInactive(retention time.Duration) error {
	return s.repository.DeleteInactiveSessions(s.now().UTC().Add(-retention))
}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
