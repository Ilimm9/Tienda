package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cuentaapplication "tienda/backend/internal/application/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireTrustedOriginRejectsUnsafeForeignOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequireTrustedOrigin("https://tienda.example"))
	router.POST("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestRequireCSRFValidatesCookieHeaderAndSessionBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	csrfToken := "csrf-token"
	authenticator := SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{UserID: uuid.New(), CSRFHash: cuentaapplication.HashSessionToken(csrfToken)}, nil
	})
	router := gin.New()
	router.Use(RequireAuth(authenticator, "tienda_session"), RequireCSRF())
	router.PATCH("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	valid := httptest.NewRequest(http.MethodPatch, "/", nil)
	valid.AddCookie(&http.Cookie{Name: "tienda_session", Value: "session-token"})
	valid.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	valid.Header.Set("X-XSRF-TOKEN", csrfToken)
	validResponse := httptest.NewRecorder()
	router.ServeHTTP(validResponse, valid)
	if validResponse.Code != http.StatusNoContent {
		t.Fatalf("token válido status=%d body=%s", validResponse.Code, validResponse.Body.String())
	}

	invalid := httptest.NewRequest(http.MethodPatch, "/", nil)
	invalid.AddCookie(&http.Cookie{Name: "tienda_session", Value: "session-token"})
	invalid.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: csrfToken})
	invalid.Header.Set("X-XSRF-TOKEN", "otro-token")
	invalidResponse := httptest.NewRecorder()
	router.ServeHTTP(invalidResponse, invalid)
	if invalidResponse.Code != http.StatusForbidden {
		t.Fatalf("token inválido status=%d", invalidResponse.Code)
	}
}
