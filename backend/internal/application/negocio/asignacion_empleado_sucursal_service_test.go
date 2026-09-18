package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type asignacionRepositoryStub struct {
	access          domain.ContextoNegocioSucursal
	permisos        []string
	empleado        domain.EmpleadoDetalle
	sucursalValida  bool
	existeActiva    bool
	asignoLlamado   bool
	asignado        domain.AsignarSucursalInput
	errorEmpleado   error
	errorAsignacion error
}

func (r *asignacionRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return r.access, nil
}
func (r *asignacionRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return r.permisos, nil
}
func (r *asignacionRepositoryStub) ObtenerEmpleado(context.Context, uuid.UUID, uuid.UUID) (domain.EmpleadoDetalle, error) {
	return r.empleado, r.errorEmpleado
}
func (r *asignacionRepositoryStub) SucursalActivaDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.sucursalValida, nil
}
func (r *asignacionRepositoryStub) Listar(context.Context, uuid.UUID, uuid.UUID, bool) ([]domain.AsignacionResumen, error) {
	return []domain.AsignacionResumen{}, nil
}
func (r *asignacionRepositoryStub) ExisteActiva(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	return r.existeActiva, nil
}
func (r *asignacionRepositoryStub) Asignar(_ context.Context, _, _ uuid.UUID, input domain.AsignarSucursalInput) (uuid.UUID, error) {
	r.asignoLlamado = true
	r.asignado = input
	return uuid.New(), nil
}
func (r *asignacionRepositoryStub) EstablecerPrincipal(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return r.errorAsignacion
}
func (r *asignacionRepositoryStub) Finalizar(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return r.errorAsignacion
}

func asignacionStub(permisos ...string) *asignacionRepositoryStub {
	return &asignacionRepositoryStub{
		access:         domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		permisos:       permisos,
		empleado:       domain.EmpleadoDetalle{ID: uuid.New()},
		sucursalValida: true,
	}
}

func TestAsignacionServiceListarExigePermiso(t *testing.T) {
	service := NewAsignacionService(asignacionStub())

	_, err := service.Listar(context.Background(), uuid.New(), uuid.New(), uuid.New(), false)

	if !errors.Is(err, ErrAsignacionProhibida) {
		t.Fatalf("sin permiso de lectura debe rechazarse: %v", err)
	}
}

func TestAsignacionServiceAsignarRechazaSucursalDeOtroNegocio(t *testing.T) {
	repository := asignacionStub(domain.PermisoAsignacionEditar)
	repository.sucursalValida = false
	service := NewAsignacionService(repository)

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{SucursalID: uuid.New()})

	if !errors.Is(err, ErrAsignacionSucursalAjena) || repository.asignoLlamado {
		t.Fatalf("una sucursal ajena o archivada no debe asignarse: %v", err)
	}
}

func TestAsignacionServiceAsignarExigeSucursal(t *testing.T) {
	service := NewAsignacionService(asignacionStub(domain.PermisoAsignacionEditar))

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || validacion.Campos["sucursal_id"] == "" {
		t.Fatalf("la sucursal es obligatoria: %v", err)
	}
}

func TestAsignacionServiceAsignarRechazaDuplicado(t *testing.T) {
	repository := asignacionStub(domain.PermisoAsignacionEditar)
	repository.existeActiva = true
	service := NewAsignacionService(repository)

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{SucursalID: uuid.New()})

	if !errors.Is(err, ErrAsignacionDuplicada) || repository.asignoLlamado {
		t.Fatalf("no debe duplicarse una asignación activa: %v", err)
	}
}

func TestAsignacionServiceAsignarConservaPrincipalSolicitada(t *testing.T) {
	repository := asignacionStub(domain.PermisoAsignacionEditar)
	service := NewAsignacionService(repository)

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{SucursalID: uuid.New(), EsPrincipal: true})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !repository.asignado.EsPrincipal {
		t.Fatal("la solicitud de principal debe llegar al repositorio")
	}
}

func TestAsignacionServiceEmpleadoAjenoNoSeEncuentra(t *testing.T) {
	repository := asignacionStub(domain.PermisoAsignacionEditar)
	repository.errorEmpleado = ErrEmpleadoNoEncontrado
	service := NewAsignacionService(repository)

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{SucursalID: uuid.New()})

	if !errors.Is(err, ErrEmpleadoNoEncontrado) {
		t.Fatalf("un empleado de otro negocio no debe encontrarse: %v", err)
	}
}

func TestAsignacionServiceNegocioArchivadoBloqueaEscritura(t *testing.T) {
	repository := asignacionStub(domain.PermisoAsignacionEditar)
	repository.access.EstadoNegocio = "archivado"
	service := NewAsignacionService(repository)

	_, err := service.Asignar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.AsignarSucursalInput{SucursalID: uuid.New()})

	if !errors.Is(err, ErrEstadoNegocio) {
		t.Fatalf("un negocio archivado no acepta asignaciones: %v", err)
	}
}

func TestAsignacionServiceLecturaNoRequierePermisoDeEdicion(t *testing.T) {
	service := NewAsignacionService(asignacionStub(domain.PermisoAsignacionVer))

	if _, err := service.Listar(context.Background(), uuid.New(), uuid.New(), uuid.New(), true); err != nil {
		t.Fatalf("el permiso de lectura debe bastar para consultar: %v", err)
	}
}

func TestAsignacionServiceFinalizarExigePermisoDeEdicion(t *testing.T) {
	service := NewAsignacionService(asignacionStub(domain.PermisoAsignacionVer))

	err := service.Finalizar(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrAsignacionProhibida) {
		t.Fatalf("finalizar requiere permiso de edición: %v", err)
	}
}
