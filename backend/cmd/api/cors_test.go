package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareAllowsPatchPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(transporthttp.CORSMiddleware("http://localhost:4200"))
	router.OPTIONS("/api/v1/catalogo/marcas/123", func(c *gin.Context) {
		t.Fatal("the preflight request must be aborted before reaching the route")
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/catalogo/marcas/123", nil)
	request.Header.Set("Origin", "http://localhost:4200")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4200" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PATCH, DELETE, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q", got)
	}
}
