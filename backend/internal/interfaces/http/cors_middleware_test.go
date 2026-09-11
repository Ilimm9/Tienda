package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewarePermitePreflightPatchDesdeFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware("http://localhost:4200"))
	router.PATCH("/api/v1/negocios/:id", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/negocios/negocio-1", nil)
	request.Header.Set("Origin", "http://localhost:4200")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, se esperaba 204", response.Code)
	}
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Methods"), http.MethodPatch) {
		t.Fatalf("métodos permitidos = %q", response.Header().Get("Access-Control-Allow-Methods"))
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4200" {
		t.Fatalf("origen permitido = %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
	if response.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("credenciales CORS no fueron habilitadas")
	}
}

func TestCORSMiddlewareNoAutorizaOrigenAjeno(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORSMiddleware("http://localhost:4200"))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://sitio-ajeno.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("origen ajeno fue autorizado: %q", response.Header().Get("Access-Control-Allow-Origin"))
	}
}
