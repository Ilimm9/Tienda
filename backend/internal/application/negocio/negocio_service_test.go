package negocio

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type negocioRepositoryStub struct {
	items         []domain.NegocioResumen
	detalle       domain.NegocioDetalle
	slugExiste    bool
	createdSlug   string
	createdInput  domain.CrearNegocioInput
	updatedInput  domain.ActualizarNegocioInput
	createErr     error
	updateCalled  bool
	archiveCalled bool
	restoreCalled bool
	permisos      []string
}

// PermisosEfectivos devuelve el catálogo completo para un propietario, que es exactamente lo que
// concentra su rol de sistema, salvo que la prueba restrinja los permisos explícitamente.
func (r *negocioRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	if r.permisos != nil {
		return r.permisos, nil
	}
	if r.detalle.TipoMiembro != "propietario" {
		return []string{}, nil
	}
	codigos := make([]string, 0, len(domain.CatalogoPermisos))
	for _, permiso := range domain.CatalogoPermisos {
		codigos = append(codigos, permiso.Codigo)
	}
	return codigos, nil
}

func (r *negocioRepositoryStub) Listar(context.Context, uuid.UUID, string) ([]domain.NegocioResumen, error) {
	return r.items, nil
}
func (r *negocioRepositoryStub) ObtenerAccesible(context.Context, uuid.UUID, uuid.UUID) (domain.NegocioDetalle, error) {
	if r.detalle.ID == uuid.Nil {
		return domain.NegocioDetalle{}, ErrNegocioNoEncontrado
	}
	return r.detalle, nil
}
func (r *negocioRepositoryStub) ExisteSlug(context.Context, string) (bool, error) {
	return r.slugExiste, nil
}
func (r *negocioRepositoryStub) Crear(_ context.Context, _ uuid.UUID, slug string, input domain.CrearNegocioInput) (domain.NegocioDetalle, error) {
	r.createdSlug = slug
	r.createdInput = input
	return r.detalle, r.createErr
}
func (r *negocioRepositoryStub) Actualizar(_ context.Context, _, _ uuid.UUID, input domain.ActualizarNegocioInput) error {
	r.updateCalled = true
	r.updatedInput = input
	return nil
}
func (r *negocioRepositoryStub) Archivar(context.Context, uuid.UUID, uuid.UUID) error {
	r.archiveCalled = true
	return nil
}
func (r *negocioRepositoryStub) Restaurar(context.Context, uuid.UUID, uuid.UUID) error {
	r.restoreCalled = true
	return nil
}

func TestNegocioServiceCrearNormalizaDatos(t *testing.T) {
	repository := &negocioRepositoryStub{detalle: domain.NegocioDetalle{ID: uuid.New()}}
	service := NewNegocioService(repository)
	direccionEstado := "  Jalisco "
	rfc := " abc010101abc "
	correo := " PERSONA@EJEMPLO.COM "

	_, err := service.Crear(context.Background(), uuid.New(), domain.CrearNegocioInput{
		NombreComercial: "  Café del Centro  ", RFC: &rfc, Correo: &correo,
		Direccion: &domain.DireccionInput{Estado: &direccionEstado},
	})
	if err != nil {
		t.Fatalf("Crear() error = %v", err)
	}
	if repository.createdSlug != "cafe-del-centro" {
		t.Fatalf("slug = %q", repository.createdSlug)
	}
	if repository.createdInput.NombreComercial != "Café del Centro" {
		t.Fatalf("nombre = %q", repository.createdInput.NombreComercial)
	}
	if repository.createdInput.RFC == nil || *repository.createdInput.RFC != "ABC010101ABC" {
		t.Fatalf("RFC no normalizado: %#v", repository.createdInput.RFC)
	}
	if repository.createdInput.Correo == nil || *repository.createdInput.Correo != "persona@ejemplo.com" {
		t.Fatalf("correo no normalizado: %#v", repository.createdInput.Correo)
	}
	if repository.createdInput.CodigoMoneda != "MXN" || repository.createdInput.ZonaHoraria != "America/Mexico_City" {
		t.Fatal("no se aplicaron valores regionales predeterminados")
	}
	if repository.createdInput.Direccion == nil || repository.createdInput.Direccion.Estado == nil || *repository.createdInput.Direccion.Estado != "Jalisco" {
		t.Fatal("dirección no normalizada")
	}
}

func TestNegocioServiceCrearResuelveSlugDuplicado(t *testing.T) {
	repository := &negocioRepositoryStub{detalle: domain.NegocioDetalle{ID: uuid.New()}, slugExiste: true}
	service := NewNegocioService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), domain.CrearNegocioInput{NombreComercial: "Mi Tienda"})
	if err != nil {
		t.Fatalf("Crear() error = %v", err)
	}
	if !strings.HasPrefix(repository.createdSlug, "mi-tienda-") || len(repository.createdSlug) != len("mi-tienda-")+8 {
		t.Fatalf("slug duplicado = %q", repository.createdSlug)
	}
}

func TestNegocioServiceRechazaEdicionDeMiembro(t *testing.T) {
	repository := &negocioRepositoryStub{detalle: domain.NegocioDetalle{ID: uuid.New(), TipoMiembro: "miembro", Estado: "activo"}}
	service := NewNegocioService(repository)
	name := "Nuevo nombre"

	_, err := service.Actualizar(context.Background(), uuid.New(), repository.detalle.ID, domain.ActualizarNegocioInput{
		NombreComercial: domain.Optional[string]{Set: true, Value: &name},
	})
	if !errors.Is(err, ErrNegocioProhibido) {
		t.Fatalf("error = %v, se esperaba ErrNegocioProhibido", err)
	}
	if repository.updateCalled {
		t.Fatal("repository no debía actualizar")
	}
}

func TestNegocioServiceArchivaPropietario(t *testing.T) {
	repository := &negocioRepositoryStub{detalle: domain.NegocioDetalle{ID: uuid.New(), TipoMiembro: "propietario", Estado: "activo"}}
	service := NewNegocioService(repository)

	if err := service.Archivar(context.Background(), uuid.New(), repository.detalle.ID); err != nil {
		t.Fatalf("Archivar() error = %v", err)
	}
	if !repository.archiveCalled {
		t.Fatal("repository debía archivar")
	}
}

func TestActualizarNegocioDistingueNullDeCampoOmitido(t *testing.T) {
	var input domain.ActualizarNegocioInput
	if err := json.Unmarshal([]byte(`{"razon_social":null}`), &input); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !input.RazonSocial.Set || input.RazonSocial.Value != nil {
		t.Fatalf("null no fue preservado: %#v", input.RazonSocial)
	}
	if input.Telefono.Set {
		t.Fatal("campo omitido fue marcado como presente")
	}
}

func TestNegocioServiceValidaEstadoDeListado(t *testing.T) {
	service := NewNegocioService(&negocioRepositoryStub{})
	_, err := service.Listar(context.Background(), uuid.New(), "eliminado")
	var validation *ErrorValidacion
	if !errors.As(err, &validation) || validation.Campos["estado"] == "" {
		t.Fatalf("error = %#v, se esperaba validación de estado", err)
	}
}
