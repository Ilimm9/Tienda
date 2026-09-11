package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type sucursalRepositoryStub struct {
	access        domain.ContextoNegocioSucursal
	items         []domain.SucursalResumen
	detail        domain.SucursalDetalle
	createdInput  domain.CrearSucursalInput
	updatedInput  domain.ActualizarSucursalInput
	createCalled  bool
	updateCalled  bool
	archiveCalled bool
	restoreCalled bool
	err           error
}

func (r *sucursalRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	if r.err != nil {
		return domain.ContextoNegocioSucursal{}, r.err
	}
	return r.access, nil
}

func (r *sucursalRepositoryStub) Listar(context.Context, uuid.UUID, string, string) ([]domain.SucursalResumen, error) {
	return r.items, r.err
}

func (r *sucursalRepositoryStub) Obtener(context.Context, uuid.UUID, uuid.UUID) (domain.SucursalDetalle, error) {
	if r.detail.ID == uuid.Nil {
		return domain.SucursalDetalle{}, ErrSucursalNoEncontrada
	}
	return r.detail, nil
}

func (r *sucursalRepositoryStub) Crear(_ context.Context, _ uuid.UUID, input domain.CrearSucursalInput) (uuid.UUID, error) {
	r.createCalled = true
	r.createdInput = input
	return r.detail.ID, r.err
}

func (r *sucursalRepositoryStub) Actualizar(_ context.Context, _, _ uuid.UUID, input domain.ActualizarSucursalInput) error {
	r.updateCalled = true
	r.updatedInput = input
	return r.err
}

func (r *sucursalRepositoryStub) Archivar(context.Context, uuid.UUID, uuid.UUID) error {
	r.archiveCalled = true
	return r.err
}

func (r *sucursalRepositoryStub) Restaurar(context.Context, uuid.UUID, uuid.UUID) error {
	r.restoreCalled = true
	return r.err
}

func TestSucursalServiceCrearNormalizaDatos(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
		detail: domain.SucursalDetalle{ID: uuid.New()},
	}
	service := NewSucursalService(repository)
	phone := "  +52 55 1234 5678  "
	city := "  Mérida "

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearSucursalInput{
		Codigo: " suc-001 ", Nombre: "  Matriz Centro ", Telefono: &phone,
		Direccion: &domain.DireccionInput{Ciudad: &city},
	})
	if err != nil {
		t.Fatalf("Crear() error = %v", err)
	}
	if !repository.createCalled {
		t.Fatal("repository debía crear")
	}
	if repository.createdInput.Codigo != "SUC-001" || repository.createdInput.Nombre != "Matriz Centro" {
		t.Fatalf("input no normalizado: %#v", repository.createdInput)
	}
	if repository.createdInput.Telefono == nil || *repository.createdInput.Telefono != "+52 55 1234 5678" {
		t.Fatalf("teléfono no normalizado: %#v", repository.createdInput.Telefono)
	}
	if repository.createdInput.Direccion == nil || repository.createdInput.Direccion.Ciudad == nil || *repository.createdInput.Direccion.Ciudad != "Mérida" {
		t.Fatal("dirección no normalizada")
	}
}

func TestSucursalServiceRechazaCodigoInvalido(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
	}
	service := NewSucursalService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearSucursalInput{
		Codigo: "código con espacios", Nombre: "Matriz",
	})
	var validation *ErrorValidacion
	if !errors.As(err, &validation) || validation.Campos["codigo"] == "" {
		t.Fatalf("error = %#v, se esperaba validación de código", err)
	}
	if repository.createCalled {
		t.Fatal("repository no debía crear")
	}
}

func TestSucursalServiceListaParaMiembro(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		items:  []domain.SucursalResumen{{ID: uuid.New()}},
	}
	items, err := NewSucursalService(repository).Listar(context.Background(), uuid.New(), uuid.New(), "activo", "")
	if err != nil {
		t.Fatalf("Listar() error = %v", err)
	}
	if len(items) != 1 || items[0].TipoMiembro != "miembro" {
		t.Fatalf("items = %#v", items)
	}
}

func TestSucursalServiceRechazaMutacionDeMiembro(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
	}
	err := NewSucursalService(repository).Archivar(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrSucursalProhibida) {
		t.Fatalf("error = %v, se esperaba ErrSucursalProhibida", err)
	}
	if repository.archiveCalled {
		t.Fatal("repository no debía archivar")
	}
}

func TestSucursalServiceRechazaMutacionDeNegocioArchivado(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "archivado", TipoMiembro: "propietario"},
	}
	err := NewSucursalService(repository).Archivar(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrEstadoNegocio) {
		t.Fatalf("error = %v, se esperaba ErrEstadoNegocio", err)
	}
}

func TestSucursalServiceRechazaActualizacionVacia(t *testing.T) {
	repository := &sucursalRepositoryStub{
		access: domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "propietario"},
	}
	_, err := NewSucursalService(repository).Actualizar(
		context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ActualizarSucursalInput{},
	)
	if !errors.Is(err, ErrSucursalSinCambios) {
		t.Fatalf("error = %v, se esperaba ErrSucursalSinCambios", err)
	}
	if repository.updateCalled {
		t.Fatal("repository no debía actualizar")
	}
}
