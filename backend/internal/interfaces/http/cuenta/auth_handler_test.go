package cuenta

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	application "tienda/backend/internal/application/cuenta"
	"tienda/backend/internal/config"
	cuentadomain "tienda/backend/internal/domain/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUserRepositoryStub struct{ user *cuentadomain.Usuario }

func (r *authUserRepositoryStub) FindByEmail(string) (*cuentadomain.Usuario, error) {
	return r.user, nil
}
func (*authUserRepositoryStub) Save(*cuentadomain.Usuario) error { return nil }
func (*authUserRepositoryStub) CreateAccount(*cuentadomain.Usuario, *cuentadomain.PerfilUsuario) error {
	return nil
}

type authSessionRepositoryStub struct{ created *cuentadomain.SesionUsuario }

func (r *authSessionRepositoryStub) CreateSession(session *cuentadomain.SesionUsuario) error {
	r.created = session
	return nil
}
func (*authSessionRepositoryStub) FindSessionByHash(string) (*cuentadomain.SesionUsuario, *cuentadomain.Usuario, error) {
	return nil, nil, application.ErrInvalidSession
}
func (*authSessionRepositoryStub) TouchSession(uuid.UUID, time.Time) error { return nil }
func (*authSessionRepositoryStub) RevokeSession(string, time.Time, string) error {
	return nil
}
func (*authSessionRepositoryStub) RevokeAllSessions(uuid.UUID, time.Time, string) error { return nil }
func (*authSessionRepositoryStub) DeleteInactiveSessions(time.Time) error               { return nil }

func TestLoginUsesOpaqueProductionCookiesWithStrictAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	verifiedAt := time.Now()
	user := &cuentadomain.Usuario{ID: uuid.New(), Correo: "user@example.com", HashContrasena: string(hash), Estado: "activo", CorreoVerificadoEn: &verifiedAt}
	sessionRepository := &authSessionRepositoryStub{}
	sessionService := application.NewSessionService(sessionRepository, application.SessionConfig{
		Duration: 24 * time.Hour, RememberDuration: 30 * 24 * time.Hour,
		RememberIdleDuration: 7 * 24 * time.Hour, ActivityTouchInterval: 5 * time.Minute,
	})
	cfg := config.Config{AppEnv: "production", SessionDuration: 24 * time.Hour}
	handler := NewAuthHandler(application.NewAuthService(&authUserRepositoryStub{user: user}), sessionService, nil, cfg)
	router := gin.New()
	router.POST("/login", handler.Login)
	request := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"correo":"user@example.com","contrasena":"correct-password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if sessionRepository.created == nil || len(sessionRepository.created.HashToken) != 64 {
		t.Fatal("la sesión opaca no fue persistida como hash")
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies=%d", len(cookies))
	}
	byName := map[string]*http.Cookie{}
	for _, cookie := range cookies {
		byName[cookie.Name] = cookie
	}
	sessionCookie := byName["__Host-tienda_session"]
	if sessionCookie == nil || !sessionCookie.HttpOnly || !sessionCookie.Secure || sessionCookie.SameSite != http.SameSiteStrictMode || sessionCookie.Path != "/" || sessionCookie.Domain != "" {
		t.Fatalf("cookie de sesión insegura: %+v", sessionCookie)
	}
	csrfCookie := byName["XSRF-TOKEN"]
	if csrfCookie == nil || csrfCookie.HttpOnly || !csrfCookie.Secure || csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie CSRF inválida: %+v", csrfCookie)
	}
}
