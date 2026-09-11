package negocio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	application "tienda/backend/internal/application/negocio"
	"tienda/backend/internal/config"
	domain "tienda/backend/internal/domain/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type handlerSucursalRepositoryStub struct {
	access domain.ContextoNegocioSucursal
	detail domain.SucursalDetalle
}

func (r *handlerSucursalRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return r.access, nil
}
func (*handlerSucursalRepositoryStub) Listar(context.Context, uuid.UUID, string, string) ([]domain.SucursalResumen, error) {
	return []domain.SucursalResumen{}, nil
}
func (r *handlerSucursalRepositoryStub) Obtener(context.Context, uuid.UUID, uuid.UUID) (domain.SucursalDetalle, error) {
	return r.detail, nil
}
func (r *handlerSucursalRepositoryStub) Crear(context.Context, uuid.UUID, domain.CrearSucursalInput) (uuid.UUID, error) {
	return r.detail.ID, nil
}
func (*handlerSucursalRepositoryStub) Actualizar(context.Context, uuid.UUID, uuid.UUID, domain.ActualizarSucursalInput) error {
	return nil
}
func (*handlerSucursalRepositoryStub) Archivar(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (*handlerSucursalRepositoryStub) Restaurar(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestSucursalHandlerCrearRespondeContrato(t *testing.T) {
	businessID := uuid.New()
	branchID := uuid.New()
	repository := &handlerSucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
		detail: domain.SucursalDetalle{ID: branchID, NegocioID: businessID, Codigo: "SUC-001", Nombre: "Matriz"},
	}
	router, cookie := testSucursalRouter(t, repository)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/negocios/"+businessID.String()+"/administracion/sucursales", strings.NewReader(`{"codigo":"suc-001","nombre":"Matriz"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || !strings.Contains(response.Body.String(), `"codigo":"SUC-001"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestSucursalHandlerValidaUUIDYCampos(t *testing.T) {
	businessID := uuid.New()
	repository := &handlerSucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
	}
	router, cookie := testSucursalRouter(t, repository)

	cases := []struct {
		name, path, body, code string
		status                 int
	}{
		{name: "uuid", path: "/api/v1/negocios/invalido/administracion/sucursales", body: `{}`, status: http.StatusBadRequest, code: "IDENTIFICADOR_INVALIDO"},
		{name: "validacion", path: "/api/v1/negocios/" + businessID.String() + "/administracion/sucursales", body: `{"codigo":"x","nombre":""}`, status: http.StatusUnprocessableEntity, code: "DATOS_INVALIDOS"},
		{name: "campo desconocido", path: "/api/v1/negocios/" + businessID.String() + "/administracion/sucursales", body: `{"codigo":"SUC-001","nombre":"Matriz","activo":false}`, status: http.StatusBadRequest, code: "DATOS_INVALIDOS"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(cookie)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.code) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestSucursalHandlerMiembroNoPuedeMutar(t *testing.T) {
	businessID := uuid.New()
	repository := &handlerSucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
	}
	router, cookie := testSucursalRouter(t, repository)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/negocios/"+businessID.String()+"/administracion/sucursales/"+uuid.NewString(), nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "ACCESO_DENEGADO") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestSucursalHandlerNoAceptaCodigoEnActualizacion(t *testing.T) {
	businessID := uuid.New()
	branchID := uuid.New()
	repository := &handlerSucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
	}
	router, cookie := testSucursalRouter(t, repository)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/negocios/"+businessID.String()+"/administracion/sucursales/"+branchID.String(), strings.NewReader(`{"codigo":"OTRO","nombre":"Matriz"}`))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "DATOS_INVALIDOS") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func testSucursalRouter(t *testing.T, repository application.SucursalRepository) (*gin.Engine, *http.Cookie) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	secret := "phase-two-test-secret"
	userID := uuid.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(), "exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewSucursalHandler(application.NewSucursalService(repository))
	router := gin.New()
	group := router.Group("/api/v1/negocios")
	group.Use(transporthttp.RequireAuth(config.Config{JWTSecret: secret}))
	group.POST("/:negocioId/administracion/sucursales", handler.Crear)
	group.PATCH("/:negocioId/administracion/sucursales/:sucursalId", handler.Actualizar)
	group.DELETE("/:negocioId/administracion/sucursales/:sucursalId", handler.Archivar)
	return router, &http.Cookie{Name: "tienda_session", Value: signed}
}
