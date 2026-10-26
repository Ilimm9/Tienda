package plataforma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cuentaapplication "tienda/backend/internal/application/cuenta"
	application "tienda/backend/internal/application/plataforma"
	domain "tienda/backend/internal/domain/plataforma"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type middlewareRepository struct {
	roles      []string
	activities int
}

func (r *middlewareRepository) RolesUsuario(context.Context, uuid.UUID) ([]string, error) {
	return r.roles, nil
}
func (*middlewareRepository) Resumen(context.Context, time.Time, time.Time) (domain.Resumen, error) {
	return domain.Resumen{}, nil
}
func (*middlewareRepository) ListarNegocios(context.Context, string, string, int, int) (domain.PaginaNegocios, error) {
	return domain.PaginaNegocios{}, nil
}
func (*middlewareRepository) ObtenerNegocio(context.Context, uuid.UUID, time.Time, time.Time) (domain.NegocioDetalle, error) {
	return domain.NegocioDetalle{}, nil
}
func (*middlewareRepository) ListarUsuarios(context.Context, uuid.UUID) ([]domain.UsuarioResumen, error) {
	return nil, nil
}
func (*middlewareRepository) ListarCuentas(context.Context, string, string, int, int) (domain.PaginaCuentas, error) {
	return domain.PaginaCuentas{}, nil
}
func (r *middlewareRepository) RegistrarActividad(context.Context, uuid.UUID, uuid.UUID, time.Time) error {
	r.activities++
	return nil
}
func (*middlewareRepository) LimpiarActividad(context.Context, time.Time) error { return nil }

func platformRouter(roles []string, withActivity bool) (*gin.Engine, *http.Cookie, *middlewareRepository) {
	gin.SetMode(gin.TestMode)
	repository := &middlewareRepository{roles: roles}
	service := application.NewService(repository)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(transporthttp.RequireAuth(transporthttp.SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{UserID: uuid.New()}, nil
	}), "session"))
	if withActivity {
		group.Use(RegistrarActividadNegocio(service))
	} else {
		group.Use(RequireAdminLectura(service))
	}
	group.GET("/negocios/:negocioId", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router, &http.Cookie{Name: "session", Value: "token"}, repository
}

func TestRequireAdminLecturaSeparaRolesTenant(t *testing.T) {
	for _, tc := range []struct {
		roles []string
		want  int
	}{{[]string{domain.RolAdminLectura}, 200}, {[]string{"PROPIETARIO"}, 403}} {
		router, cookie, _ := platformRouter(tc.roles, false)
		request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+uuid.NewString(), nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Fatalf("roles %v: esperaba %d, obtuvo %d", tc.roles, tc.want, response.Code)
		}
	}
}

func TestActividadSoloRegistraRespuestaExitosa(t *testing.T) {
	router, cookie, repository := platformRouter(nil, true)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+uuid.NewString(), nil)
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if repository.activities != 1 {
		t.Fatalf("esperaba una actividad, obtuvo %d", repository.activities)
	}
}

func TestRejectPlatformAdminImpideOperarTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &middlewareRepository{roles: []string{domain.RolAdminLectura}}
	service := application.NewService(repository)
	router := gin.New()
	router.Use(transporthttp.RequireAuth(transporthttp.SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{UserID: uuid.New()}, nil
	}), "session"))
	router.Use(RejectPlatformAdmin(service))
	router.GET("/operacion", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/operacion", nil)
	request.AddCookie(&http.Cookie{Name: "session", Value: "token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("un administrador global no debe operar tenant: %d", response.Code)
	}
}
