package negocio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	cuentaapplication "tienda/backend/internal/application/cuenta"
	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type permisoRepositoryStub struct {
	permisos []string
}

func (r *permisoRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"}, nil
}
func (r *permisoRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return r.permisos, nil
}
func (*permisoRepositoryStub) ListarPermisos(context.Context) ([]domain.Permiso, error) {
	return domain.CatalogoPermisos, nil
}
func (*permisoRepositoryStub) ListarRoles(context.Context, uuid.UUID, bool) ([]domain.RolResumen, error) {
	return []domain.RolResumen{}, nil
}
func (*permisoRepositoryStub) ObtenerRol(context.Context, uuid.UUID, uuid.UUID) (domain.RolDetalle, error) {
	return domain.RolDetalle{}, application.ErrRolNoEncontrado
}
func (*permisoRepositoryStub) ExisteCodigoRol(context.Context, uuid.UUID, string) (bool, error) {
	return false, nil
}
func (*permisoRepositoryStub) CrearRol(context.Context, uuid.UUID, uuid.UUID, domain.CrearRolInput) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (*permisoRepositoryStub) ActualizarRol(context.Context, uuid.UUID, uuid.UUID, domain.ActualizarRolInput) error {
	return nil
}
func (*permisoRepositoryStub) EliminarRol(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (*permisoRepositoryStub) ContarMiembrosConRol(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}
func (*permisoRepositoryStub) ListarMiembros(context.Context, uuid.UUID) ([]domain.MiembroRoles, error) {
	return []domain.MiembroRoles{}, nil
}
func (*permisoRepositoryStub) ReemplazarRolesDeMembresia(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID, uuid.UUID) error {
	return nil
}
func (*permisoRepositoryStub) MembresiaPerteneceANegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}
func (*permisoRepositoryStub) QuedaPropietarioConRolSistema(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) (bool, error) {
	return true, nil
}

func testPermisoRouter(t *testing.T, permisos []string) (*gin.Engine, *http.Cookie) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	service := application.NewRolService(&permisoRepositoryStub{permisos: permisos})
	router := gin.New()
	grupo := router.Group("/api/v1/negocios")
	grupo.Use(transporthttp.RequireAuth(transporthttp.SessionAuthenticatorFunc(func(string) (cuentaapplication.AuthenticatedSession, error) {
		return cuentaapplication.AuthenticatedSession{UserID: userID}, nil
	}), "tienda_session"))
	grupo.GET("/:negocioId/protegido",
		RequierePermiso(service, domain.PermisoRolGestionar),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return router, &http.Cookie{Name: "tienda_session", Value: "opaque-test-token"}
}

func ejecutarPermiso(t *testing.T, permisos []string, conSesion bool, negocioID string) int {
	t.Helper()
	router, cookie := testPermisoRouter(t, permisos)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/negocios/"+negocioID+"/protegido", nil)
	if conSesion {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code
}

func TestRequierePermisoAceptaMembresiaConPermiso(t *testing.T) {
	code := ejecutarPermiso(t, []string{domain.PermisoRolGestionar}, true, uuid.New().String())

	if code != http.StatusOK {
		t.Fatalf("un miembro con el permiso debe pasar, respondió %d", code)
	}
}

func TestRequierePermisoRechazaSinPermisoCon403(t *testing.T) {
	code := ejecutarPermiso(t, []string{domain.PermisoRolVer}, true, uuid.New().String())

	if code != http.StatusForbidden {
		t.Fatalf("sin el permiso debe responder 403, respondió %d", code)
	}
}

func TestRequierePermisoRechazaSesionAusente(t *testing.T) {
	code := ejecutarPermiso(t, []string{domain.PermisoRolGestionar}, false, uuid.New().String())

	if code != http.StatusUnauthorized {
		t.Fatalf("sin sesión debe responder 401, respondió %d", code)
	}
}

func TestRequierePermisoRechazaUUIDInvalido(t *testing.T) {
	code := ejecutarPermiso(t, []string{domain.PermisoRolGestionar}, true, "no-es-uuid")

	if code != http.StatusBadRequest {
		t.Fatalf("un UUID mal formado debe responder 400, respondió %d", code)
	}
}
