package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cuentaapplication "tienda/backend/internal/application/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestRequireAuthAceptaSesionValida(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	router := gin.New()
	router.Use(RequireAuth(SessionAuthenticatorFunc(func(token string) (cuentaapplication.AuthenticatedSession, error) {
		if token != "opaque-token" {
			return cuentaapplication.AuthenticatedSession{}, errors.New("invalid")
		}
		return cuentaapplication.AuthenticatedSession{UserID: userID, Email: "test@example.com"}, nil
	}), "tienda_session"))
	router.GET("/", func(c *gin.Context) {
		actual, ok := AuthenticatedUserID(c)
		if !ok || actual != userID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "tienda_session", Value: "opaque-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestRequireAuthRechazaSesionInvalida(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequireAuth(SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{}, errors.New("invalid")
	}), "tienda_session"))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "tienda_session", Value: "invalid-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401", response.Code)
	}
}
