package negocio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cuentaapplication "tienda/backend/internal/application/cuenta"
	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextoRepositoryStub struct {
	opciones   []domain.ContextoNegocio
	negocioOK  bool
	sucursalOK bool
}

func (r *contextoRepositoryStub) ListarOpcionesContexto(context.Context, uuid.UUID) ([]domain.ContextoNegocio, error) {
	return r.opciones, nil
}

func (r *contextoRepositoryStub) NegocioActivoAccesible(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.negocioOK, nil
}

func (r *contextoRepositoryStub) SucursalActivaDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.sucursalOK, nil
}

func testContextoRouter(t *testing.T, repository *contextoRepositoryStub) (*gin.Engine, *http.Cookie) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := application.NewContextoService(repository)
	handler := NewContextoHandler(service)
	router := gin.New()
	protegido := router.Group("/api/v1")
	protegido.Use(transporthttp.RequireAuth(transporthttp.SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{UserID: userID}, nil
	}), "tienda_session"))
	protegido.GET("/contexto/opciones", handler.Opciones)
	porNegocio := protegido.Group("/negocios")
	porNegocio.Use(RequireNegocioActivo(service))
	porNegocio.GET("/:negocioId/prueba", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return router, &http.Cookie{Name: "tienda_session", Value: "opaque-test-token"}
}

func TestContextoHandlerOpcionesRequiereSesion(t *testing.T) {
	router, _ := testContextoRouter(t, &contextoRepositoryStub{})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/contexto/opciones", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("sesión ausente debe responder 401, respondió %d", response.Code)
	}
}

func TestContextoHandlerOpcionesDevuelveContrato(t *testing.T) {
	negocioID := uuid.New()
	sucursalID := uuid.New()
	router, cookie := testContextoRouter(t, &contextoRepositoryStub{opciones: []domain.ContextoNegocio{{
		ID: negocioID, Slug: "tienda-centro", NombreComercial: "Tienda Centro", TipoMiembro: "propietario",
		Sucursales: []domain.ContextoSucursal{{ID: sucursalID, Codigo: "SUC-001", Nombre: "Matriz", EsPrincipal: true}},
	}}})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/contexto/opciones", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	body := response.Body.String()
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, body)
	}
	for _, esperado := range []string{`"total":1`, `"nombre_comercial":"Tienda Centro"`, `"tipo_miembro":"propietario"`, `"es_principal":true`, `"codigo":"SUC-001"`} {
		if !strings.Contains(body, esperado) {
			t.Fatalf("falta %s en %s", esperado, body)
		}
	}
	if strings.Contains(body, "permisos") || strings.Contains(body, "roles") {
		t.Fatalf("el contrato no debe exponer roles ni permisos antes de fase 4: %s", body)
	}
}

func TestContextoHandlerCuentaSinNegociosDevuelveListaVacia(t *testing.T) {
	router, cookie := testContextoRouter(t, &contextoRepositoryStub{opciones: []domain.ContextoNegocio{}})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/contexto/opciones", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRequireNegocioActivoRechazaUUIDInvalido(t *testing.T) {
	router, cookie := testContextoRouter(t, &contextoRepositoryStub{negocioOK: true})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/no-es-uuid/prueba", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("UUID mal formado debe responder 400, respondió %d", response.Code)
	}
}

func TestRequireNegocioActivoRechazaNegocioAjenoCon404(t *testing.T) {
	router, cookie := testContextoRouter(t, &contextoRepositoryStub{negocioOK: false})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+uuid.New().String()+"/prueba", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("negocio ajeno debe responder 404 indistinguible, respondió %d", response.Code)
	}
}

func TestRequireNegocioActivoRequiereSesion(t *testing.T) {
	router, _ := testContextoRouter(t, &contextoRepositoryStub{negocioOK: true})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+uuid.New().String()+"/prueba", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("ruta por negocio ya no debe ser pública, respondió %d", response.Code)
	}
}

func TestRequireNegocioActivoPermiteMembresiaVigente(t *testing.T) {
	router, cookie := testContextoRouter(t, &contextoRepositoryStub{negocioOK: true})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+uuid.New().String()+"/prueba", nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("membresía activa debe pasar, respondió %d", response.Code)
	}
}
