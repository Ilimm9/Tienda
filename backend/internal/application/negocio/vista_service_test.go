package negocio

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type vistaRepositoryStub struct {
	valor    string
	guardado *string
}

func (r *vistaRepositoryStub) SeccionesOcultas(context.Context, uuid.UUID) (string, error) {
	return r.valor, nil
}

func (r *vistaRepositoryStub) GuardarSeccionesOcultas(_ context.Context, _ uuid.UUID, valor string) error {
	r.guardado = &valor
	return nil
}

func TestVistaServiceObtenerEmpiezaConTodoVisible(t *testing.T) {
	service := NewVistaService(&vistaRepositoryStub{}, &contextoRepositoryStub{})
	vistas, err := service.Obtener(context.Background(), uuid.New())
	if err != nil || len(vistas.SeccionesOcultas) != 0 || !reflect.DeepEqual(vistas.SeccionesOcultables, []string{"categorias", "marcas"}) {
		t.Fatalf("vistas = %#v, err = %v", vistas, err)
	}
}

func TestVistaServiceGuardarSoloPropietario(t *testing.T) {
	negocioID := uuid.New()
	for _, tipo := range []string{"empleado", "invitado"} {
		repo := &vistaRepositoryStub{}
		contexto := &contextoRepositoryStub{opciones: []domain.ContextoNegocio{{ID: negocioID, TipoMiembro: tipo}}}
		_, err := NewVistaService(repo, contexto).Guardar(context.Background(), uuid.New(), negocioID, []string{"categorias"})
		if !errors.Is(err, ErrVistasSoloPropietario) || repo.guardado != nil {
			t.Fatalf("%s: err = %v, guardado = %v", tipo, err, repo.guardado)
		}
	}
	// Ser propietario de otro negocio no basta.
	repo := &vistaRepositoryStub{}
	contexto := &contextoRepositoryStub{opciones: []domain.ContextoNegocio{{ID: uuid.New(), TipoMiembro: "propietario"}}}
	if _, err := NewVistaService(repo, contexto).Guardar(context.Background(), uuid.New(), negocioID, nil); !errors.Is(err, ErrVistasSoloPropietario) {
		t.Fatalf("otro negocio: err = %v", err)
	}
}

func TestVistaServiceGuardarNormalizaYValida(t *testing.T) {
	negocioID := uuid.New()
	contexto := &contextoRepositoryStub{opciones: []domain.ContextoNegocio{{ID: negocioID, TipoMiembro: "propietario"}}}
	repo := &vistaRepositoryStub{}
	vistas, err := NewVistaService(repo, contexto).Guardar(context.Background(), uuid.New(), negocioID, []string{"marcas", "Categorias", "marcas"})
	if err != nil || repo.guardado == nil || *repo.guardado != "categorias,marcas" || !reflect.DeepEqual(vistas.SeccionesOcultas, []string{"categorias", "marcas"}) {
		t.Fatalf("vistas = %#v, guardado = %v, err = %v", vistas, repo.guardado, err)
	}
	repo = &vistaRepositoryStub{}
	if _, err := NewVistaService(repo, contexto).Guardar(context.Background(), uuid.New(), negocioID, []string{"productos"}); !errors.Is(err, domain.ErrSeccionNoOcultable) || repo.guardado != nil {
		t.Fatalf("productos: err = %v, guardado = %v", err, repo.guardado)
	}
}
