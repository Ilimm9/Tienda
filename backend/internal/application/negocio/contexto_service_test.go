package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type contextoRepositoryStub struct {
	opciones      []domain.ContextoNegocio
	negocioOK     bool
	sucursalOK    bool
	errorOpciones error
	errorNegocio  error
	errorSucursal error
}

func (r *contextoRepositoryStub) ListarOpcionesContexto(context.Context, uuid.UUID) ([]domain.ContextoNegocio, error) {
	return r.opciones, r.errorOpciones
}

func (r *contextoRepositoryStub) NegocioActivoAccesible(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.negocioOK, r.errorNegocio
}

func (r *contextoRepositoryStub) SucursalActivaDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.sucursalOK, r.errorSucursal
}

func TestContextoServiceListarOpcionesDevuelveListaVacia(t *testing.T) {
	service := NewContextoService(&contextoRepositoryStub{opciones: []domain.ContextoNegocio{}})

	items, err := service.ListarOpciones(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("una cuenta sin negocios debe recibir lista vacía, recibió %d", len(items))
	}
}

func TestContextoServiceListarOpcionesConservaNegocioSinSucursales(t *testing.T) {
	negocioID := uuid.New()
	service := NewContextoService(&contextoRepositoryStub{opciones: []domain.ContextoNegocio{
		{ID: negocioID, NombreComercial: "Tienda Centro", TipoMiembro: "propietario", Sucursales: []domain.ContextoSucursal{}},
	}})

	items, err := service.ListarOpciones(context.Background(), uuid.New())

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(items) != 1 || items[0].ID != negocioID || len(items[0].Sucursales) != 0 {
		t.Fatalf("negocio sin sucursales debe aparecer con arreglo vacío: %+v", items)
	}
}

func TestContextoServiceValidarNegocioActivo(t *testing.T) {
	service := NewContextoService(&contextoRepositoryStub{negocioOK: true})

	if err := service.ValidarNegocioActivo(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("negocio accesible no debe fallar: %v", err)
	}
}

func TestContextoServiceValidarNegocioAjenoNoDisponible(t *testing.T) {
	service := NewContextoService(&contextoRepositoryStub{negocioOK: false})

	err := service.ValidarNegocioActivo(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrContextoNegocioNoDisponible) {
		t.Fatalf("negocio ajeno o inexistente debe ser indistinguible: %v", err)
	}
}

func TestContextoServiceValidarSucursalAjenaNoDisponible(t *testing.T) {
	service := NewContextoService(&contextoRepositoryStub{sucursalOK: false})

	err := service.ValidarSucursalActiva(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrContextoSucursalNoDisponible) {
		t.Fatalf("sucursal ajena o archivada debe rechazarse: %v", err)
	}
}

func TestContextoServiceValidarSucursalActiva(t *testing.T) {
	service := NewContextoService(&contextoRepositoryStub{sucursalOK: true})

	if err := service.ValidarSucursalActiva(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("sucursal activa del negocio no debe fallar: %v", err)
	}
}

func TestContextoServicePropagaErrorDeRepositorio(t *testing.T) {
	fallo := errors.New("falla de base")
	service := NewContextoService(&contextoRepositoryStub{errorNegocio: fallo})

	err := service.ValidarNegocioActivo(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, fallo) {
		t.Fatalf("el error de infraestructura no debe convertirse en negocio no disponible: %v", err)
	}
}
