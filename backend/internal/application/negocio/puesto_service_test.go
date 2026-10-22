package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type puestoRepositoryStub struct {
	puesto      domain.PuestoNegocio
	duplicado   bool
	rolExiste   bool
	guardado    *domain.PuestoNegocio
	excluidoDup *uuid.UUID
}

func (r *puestoRepositoryStub) Listar(context.Context, uuid.UUID) ([]domain.PuestoNegocio, error) {
	return []domain.PuestoNegocio{r.puesto}, nil
}
func (r *puestoRepositoryStub) Obtener(context.Context, uuid.UUID, uuid.UUID) (domain.PuestoNegocio, error) {
	if r.puesto.ID == uuid.Nil {
		return domain.PuestoNegocio{}, ErrPuestoNoEncontrado
	}
	return r.puesto, nil
}
func (r *puestoRepositoryStub) ExisteNombre(_ context.Context, _ uuid.UUID, _ string, excluir *uuid.UUID) (bool, error) {
	r.excluidoDup = excluir
	return r.duplicado, nil
}
func (r *puestoRepositoryStub) RolDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.rolExiste, nil
}
func (r *puestoRepositoryStub) Crear(_ context.Context, puesto *domain.PuestoNegocio) error {
	r.guardado = puesto
	return nil
}
func (r *puestoRepositoryStub) Guardar(_ context.Context, puesto *domain.PuestoNegocio) error {
	r.guardado = puesto
	return nil
}

func puestoService(repo *puestoRepositoryStub, permisos ...string) *PuestoService {
	return NewPuestoService(repo, NewEmpleadoService(empleadoStub(permisos...)))
}

func TestPuestoServiceCrearNormalizaYExigePermiso(t *testing.T) {
	repo := &puestoRepositoryStub{}
	if _, err := puestoService(repo, domain.PermisoEmpleadoVer).Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearPuestoInput{Nombre: "Cajero"}); !errors.Is(err, ErrEmpleadoProhibido) {
		t.Fatalf("sin permiso de gestión debe rechazarse: %v", err)
	}

	puesto, err := puestoService(repo, domain.PermisoEmpleadoGestionar).Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearPuestoInput{Nombre: "  Encargado   de  sucursal "})
	if err != nil || puesto.Nombre != "Encargado de sucursal" || !puesto.Activo || repo.guardado == nil {
		t.Fatalf("puesto no normalizado: %#v %v", puesto, err)
	}
}

func TestPuestoServiceCrearRechazaDuplicadoCortoYRolAjeno(t *testing.T) {
	service := puestoService(&puestoRepositoryStub{duplicado: true}, domain.PermisoEmpleadoGestionar)
	if _, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearPuestoInput{Nombre: "cajero"}); !errors.Is(err, ErrPuestoDuplicado) {
		t.Fatalf("duplicado debe rechazarse: %v", err)
	}

	var validacion *ErrorValidacion
	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearPuestoInput{Nombre: "C"})
	if !errors.As(err, &validacion) || validacion.Campos["nombre"] == "" {
		t.Fatalf("nombre corto debe rechazarse: %v", err)
	}

	rol := uuid.New()
	service = puestoService(&puestoRepositoryStub{}, domain.PermisoEmpleadoGestionar)
	_, err = service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearPuestoInput{Nombre: "Cajero", RolSugeridoID: &rol})
	if !errors.As(err, &validacion) || validacion.Campos["rol_sugerido_id"] == "" {
		t.Fatalf("rol ajeno debe rechazarse: %v", err)
	}
}

func TestPuestoServiceActualizarDesactivaExcluyendoseDelDuplicado(t *testing.T) {
	repo := &puestoRepositoryStub{puesto: domain.PuestoNegocio{ID: uuid.New(), Nombre: "Cajero", Activo: true}}
	activo := false

	puesto, err := puestoService(repo, domain.PermisoEmpleadoGestionar).Actualizar(context.Background(), uuid.New(), uuid.New(), repo.puesto.ID, domain.ActualizarPuestoInput{Activo: &activo})

	if err != nil || puesto.Activo || repo.excluidoDup == nil || *repo.excluidoDup != repo.puesto.ID {
		t.Fatalf("debe desactivar sin chocar consigo mismo: %#v %v", puesto, err)
	}
}

func TestEmpleadoServiceTelefonoDiezDigitos(t *testing.T) {
	service := NewEmpleadoService(empleadoStub(domain.PermisoEmpleadoGestionar))
	var validacion *ErrorValidacion
	for _, telefono := range []string{"551234567", "55123456789", "+52 55 1234 5678", "55 1234 5678"} {
		_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{Nombre: "Ana", PrimerApellido: "Ruiz", Telefono: texto(telefono)})
		if !errors.As(err, &validacion) || validacion.Campos["telefono"] == "" {
			t.Fatalf("%q debe rechazarse: %v", telefono, err)
		}
	}
	if _, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{Nombre: "Ana", PrimerApellido: "Ruiz", Telefono: texto("5512345678")}); err != nil {
		t.Fatalf("10 dígitos debe aceptarse: %v", err)
	}
	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ActualizarEmpleadoInput{Telefono: domain.Optional[string]{Set: true, Value: texto("123")}})
	if !errors.As(err, &validacion) || validacion.Campos["telefono"] == "" {
		t.Fatalf("actualizar con teléfono inválido debe rechazarse: %v", err)
	}
}
