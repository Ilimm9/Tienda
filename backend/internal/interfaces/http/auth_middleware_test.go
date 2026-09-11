package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tienda/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestRequireAuthAceptaSesionValida(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(), "exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(RequireAuth(config.Config{JWTSecret: secret}))
	router.GET("/", func(c *gin.Context) {
		actual, ok := AuthenticatedUserID(c)
		if !ok || actual != userID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "tienda_session", Value: signed})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestRequireAuthRechazaSesionSinExpiracion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": uuid.NewString()})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(RequireAuth(config.Config{JWTSecret: secret}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "tienda_session", Value: signed})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, se esperaba 401", response.Code)
	}
}
