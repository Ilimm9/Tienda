package plataforma

import (
	"context"
	"testing"
	"time"

	domain "tienda/backend/internal/domain/plataforma"

	"github.com/google/uuid"
)

type repositoryStub struct {
	roles      []string
	query      string
	estado     string
	page       int
	pageSize   int
	activityAt time.Time
}

func (r *repositoryStub) RolesUsuario(context.Context, uuid.UUID) ([]string, error) {
	return r.roles, nil
}
func (*repositoryStub) Resumen(context.Context, time.Time, time.Time) (domain.Resumen, error) {
	return domain.Resumen{}, nil
}
func (r *repositoryStub) ListarNegocios(_ context.Context, query, estado string, page, pageSize int) (domain.PaginaNegocios, error) {
	r.query, r.estado, r.page, r.pageSize = query, estado, page, pageSize
	return domain.PaginaNegocios{}, nil
}
func (*repositoryStub) ObtenerNegocio(context.Context, uuid.UUID, time.Time, time.Time) (domain.NegocioDetalle, error) {
	return domain.NegocioDetalle{}, nil
}
func (*repositoryStub) ListarUsuarios(context.Context, uuid.UUID) ([]domain.UsuarioResumen, error) {
	return nil, nil
}
func (r *repositoryStub) RegistrarActividad(_ context.Context, _, _ uuid.UUID, at time.Time) error {
	r.activityAt = at
	return nil
}
func (*repositoryStub) LimpiarActividad(context.Context, time.Time) error { return nil }

func TestAutorizarExigeRolGlobal(t *testing.T) {
	userID := uuid.New()
	allowed := NewService(&repositoryStub{roles: []string{domain.RolAdminLectura}})
	if err := allowed.Autorizar(context.Background(), userID); err != nil {
		t.Fatalf("ADMIN_LECTURA debe pasar: %v", err)
	}
	denied := NewService(&repositoryStub{roles: []string{"PROPIETARIO"}})
	if err := denied.Autorizar(context.Background(), userID); err != ErrAccesoDenegado {
		t.Fatalf("un rol tenant no debe pasar: %v", err)
	}
}

func TestListarNegociosNormalizaFiltrosYPaginacion(t *testing.T) {
	repo := &repositoryStub{}
	service := NewService(repo)
	_, _ = service.ListarNegocios(context.Background(), "  Tienda  ", "desconocido", 0, 1000)
	if repo.query != "Tienda" || repo.estado != "" || repo.page != 1 || repo.pageSize != 100 {
		t.Fatalf("filtros no normalizados: %#v", repo)
	}
}

type assignmentStub struct {
	assigned AdminAsignado
	calls    int
}

func (s *assignmentStub) AsignarAdminLectura(context.Context, string, string) (AdminAsignado, error) {
	s.calls++
	return s.assigned, nil
}
func (s *assignmentStub) RetirarAdminLectura(context.Context, string, string) (AdminAsignado, error) {
	s.calls++
	return s.assigned, nil
}
func (*assignmentStub) ListarAdmins(context.Context) ([]AdminAsignado, error) { return nil, nil }

func TestAsignacionExigeMotivo(t *testing.T) {
	repo := &assignmentStub{}
	service := NewAssignmentService(repo)
	if _, err := service.Asignar(context.Background(), "admin@ejemplo.com", " "); err != ErrMotivoRequerido {
		t.Fatalf("esperaba motivo requerido, obtuve %v", err)
	}
	if repo.calls != 0 {
		t.Fatal("no debe tocar repositorio sin motivo")
	}
}
