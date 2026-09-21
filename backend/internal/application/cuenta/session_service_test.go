package cuenta

import (
	"errors"
	"testing"
	"time"

	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/google/uuid"
)

type sessionRepositoryStub struct {
	session     *cuentadomain.SesionUsuario
	user        *cuentadomain.Usuario
	created     *cuentadomain.SesionUsuario
	revokedHash string
	touchedAt   time.Time
	findError   error
}

func (r *sessionRepositoryStub) CreateSession(session *cuentadomain.SesionUsuario) error {
	copy := *session
	r.created = &copy
	return nil
}
func (r *sessionRepositoryStub) FindSessionByHash(string) (*cuentadomain.SesionUsuario, *cuentadomain.Usuario, error) {
	return r.session, r.user, r.findError
}
func (r *sessionRepositoryStub) TouchSession(_ uuid.UUID, at time.Time) error {
	r.touchedAt = at
	return nil
}
func (r *sessionRepositoryStub) RevokeSession(hash string, _ time.Time, _ string) error {
	r.revokedHash = hash
	return nil
}
func (*sessionRepositoryStub) RevokeAllSessions(uuid.UUID, time.Time, string) error { return nil }
func (*sessionRepositoryStub) DeleteInactiveSessions(time.Time) error               { return nil }

func sessionTestConfig() SessionConfig {
	return SessionConfig{Duration: 24 * time.Hour, RememberDuration: 30 * 24 * time.Hour, RememberIdleDuration: 7 * 24 * time.Hour, ActivityTouchInterval: 5 * time.Minute}
}

func TestSessionCreatePersistsOnlyHashesAndUses256BitTokens(t *testing.T) {
	repository := &sessionRepositoryStub{}
	service := NewSessionService(repository, sessionTestConfig())
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	created, err := service.Create(uuid.New(), false, "127.0.0.1", "browser")
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Token) < 43 || len(created.CSRFToken) < 43 {
		t.Fatalf("los tokens no contienen 256 bits codificados: sesión=%d csrf=%d", len(created.Token), len(created.CSRFToken))
	}
	if repository.created.HashToken != HashSessionToken(created.Token) || repository.created.HashCSRF != HashSessionToken(created.CSRFToken) {
		t.Fatal("el repositorio no recibió los hashes esperados")
	}
	if repository.created.HashToken == created.Token || repository.created.HashCSRF == created.CSRFToken {
		t.Fatal("un token en claro llegó a persistencia")
	}
	if !created.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("expiración=%s", created.ExpiresAt)
	}
}

func TestSessionAuthenticateRejectsExpiredRevokedAndDisabled(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-time.Hour)
	activeUser := &cuentadomain.Usuario{ID: uuid.New(), Correo: "user@example.com", Estado: "activo", CorreoVerificadoEn: &verifiedAt}
	tests := []struct {
		name    string
		session *cuentadomain.SesionUsuario
		user    *cuentadomain.Usuario
	}{
		{name: "expired", session: &cuentadomain.SesionUsuario{ExpiraEn: now}, user: activeUser},
		{name: "revoked", session: &cuentadomain.SesionUsuario{ExpiraEn: now.Add(time.Hour), RevocadoEn: &now}, user: activeUser},
		{name: "disabled", session: &cuentadomain.SesionUsuario{ExpiraEn: now.Add(time.Hour)}, user: &cuentadomain.Usuario{ID: uuid.New(), Estado: "deshabilitado", CorreoVerificadoEn: &verifiedAt}},
		{name: "unverified", session: &cuentadomain.SesionUsuario{ExpiraEn: now.Add(time.Hour)}, user: &cuentadomain.Usuario{ID: uuid.New(), Estado: "activo"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &sessionRepositoryStub{session: test.session, user: test.user}
			service := NewSessionService(repository, sessionTestConfig())
			service.now = func() time.Time { return now }
			if _, err := service.Authenticate("token"); !errors.Is(err, ErrInvalidSession) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestSessionRememberedExpiresAfterIdleLimit(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	verifiedAt := now.Add(-time.Hour)
	repository := &sessionRepositoryStub{
		session: &cuentadomain.SesionUsuario{Recordarme: true, ExpiraEn: now.Add(24 * time.Hour), UltimaActividadEn: now.Add(-7 * 24 * time.Hour)},
		user:    &cuentadomain.Usuario{ID: uuid.New(), Estado: "activo", CorreoVerificadoEn: &verifiedAt},
	}
	service := NewSessionService(repository, sessionTestConfig())
	service.now = func() time.Time { return now }
	if _, err := service.Authenticate("token"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("error=%v", err)
	}
}

func TestSessionLogoutRevokesHashNotToken(t *testing.T) {
	repository := &sessionRepositoryStub{}
	service := NewSessionService(repository, sessionTestConfig())
	if err := service.Revoke("captured-token", "logout"); err != nil {
		t.Fatal(err)
	}
	if repository.revokedHash != HashSessionToken("captured-token") || repository.revokedHash == "captured-token" {
		t.Fatalf("hash revocado=%q", repository.revokedHash)
	}
}
