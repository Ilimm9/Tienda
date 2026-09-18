package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type empleadoRepositoryStub struct {
	access       domain.ContextoNegocioSucursal
	permisos     []string
	detalle      domain.EmpleadoDetalle
	existeNumero bool
	creado       domain.CrearEmpleadoInput
	actualizado  domain.ActualizarEmpleadoInput
	creoLlamado  bool
}

func (r *empleadoRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return r.access, nil
}
func (r *empleadoRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return r.permisos, nil
}
func (r *empleadoRepositoryStub) Listar(context.Context, uuid.UUID, string, string) ([]domain.EmpleadoResumen, error) {
	return []domain.EmpleadoResumen{}, nil
}
func (r *empleadoRepositoryStub) Obtener(context.Context, uuid.UUID, uuid.UUID) (domain.EmpleadoDetalle, error) {
	if r.detalle.ID == uuid.Nil {
		return domain.EmpleadoDetalle{}, ErrEmpleadoNoEncontrado
	}
	return r.detalle, nil
}
func (r *empleadoRepositoryStub) ExisteNumero(context.Context, uuid.UUID, string, *uuid.UUID) (bool, error) {
	return r.existeNumero, nil
}
func (r *empleadoRepositoryStub) Crear(_ context.Context, _, _ uuid.UUID, input domain.CrearEmpleadoInput) (uuid.UUID, error) {
	r.creoLlamado = true
	r.creado = input
	return r.detalle.ID, nil
}
func (r *empleadoRepositoryStub) Actualizar(_ context.Context, _, _ uuid.UUID, input domain.ActualizarEmpleadoInput) error {
	r.actualizado = input
	return nil
}

func empleadoStub(permisos ...string) *empleadoRepositoryStub {
	return &empleadoRepositoryStub{
		access:   domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		permisos: permisos,
		detalle:  domain.EmpleadoDetalle{ID: uuid.New()},
	}
}

func texto(valor string) *string { return &valor }

func TestEmpleadoServiceListarExigePermiso(t *testing.T) {
	service := NewEmpleadoService(empleadoStub())

	_, err := service.Listar(context.Background(), uuid.New(), uuid.New(), "", "")

	if !errors.Is(err, ErrEmpleadoProhibido) {
		t.Fatalf("sin permiso de lectura debe rechazarse: %v", err)
	}
}

func TestEmpleadoServiceListarValidaEstado(t *testing.T) {
	service := NewEmpleadoService(empleadoStub(domain.PermisoEmpleadoVer))

	_, err := service.Listar(context.Background(), uuid.New(), uuid.New(), "inventado", "")

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) {
		t.Fatalf("un estado desconocido debe rechazarse: %v", err)
	}
}

func TestEmpleadoServiceCrearNormalizaYNacePendiente(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	service := NewEmpleadoService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{
		NumeroEmpleado: texto("  emp-001 "), Nombre: "  Ana ", PrimerApellido: " López ",
		Correo: texto("  ANA@Tienda.MX "), Puesto: texto("   "),
	})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if *repository.creado.NumeroEmpleado != "EMP-001" {
		t.Fatalf("el número debe normalizarse: %q", *repository.creado.NumeroEmpleado)
	}
	if repository.creado.Nombre != "Ana" || repository.creado.PrimerApellido != "López" {
		t.Fatalf("nombre y apellido deben recortarse: %#v", repository.creado)
	}
	if *repository.creado.Correo != "ana@tienda.mx" {
		t.Fatalf("el correo debe normalizarse en minúsculas: %q", *repository.creado.Correo)
	}
	if repository.creado.Puesto != nil {
		t.Fatal("un puesto vacío debe guardarse como nulo")
	}
}

func TestEmpleadoServiceCrearValidaCampos(t *testing.T) {
	service := NewEmpleadoService(empleadoStub(domain.PermisoEmpleadoGestionar))

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{
		Nombre: "A", PrimerApellido: "", Correo: texto("no-es-correo"), ContratadoEn: texto("13/05/2026"),
	})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) {
		t.Fatalf("se esperaba error de validación: %v", err)
	}
	for _, campo := range []string{"nombre", "primer_apellido", "correo", "contratado_en"} {
		if validacion.Campos[campo] == "" {
			t.Fatalf("falta el error de %s: %#v", campo, validacion.Campos)
		}
	}
}

func TestEmpleadoServiceCrearRechazaNumeroDuplicado(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	repository.existeNumero = true
	service := NewEmpleadoService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{
		NumeroEmpleado: texto("EMP-001"), Nombre: "Ana", PrimerApellido: "López",
	})

	if !errors.Is(err, ErrEmpleadoConflicto) || repository.creoLlamado {
		t.Fatalf("el número es único por negocio: %v", err)
	}
}

func TestEmpleadoServiceCrearPermiteEmpleadoSinNumeroNiCuenta(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	service := NewEmpleadoService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{
		Nombre: "Ana", PrimerApellido: "López",
	})

	if err != nil {
		t.Fatalf("un empleado puede registrarse sin número ni cuenta: %v", err)
	}
	if repository.creado.NumeroEmpleado != nil {
		t.Fatal("un número vacío debe guardarse como nulo")
	}
}

func TestEmpleadoServiceNegocioArchivadoBloqueaEscritura(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	repository.access.EstadoNegocio = "archivado"
	service := NewEmpleadoService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearEmpleadoInput{
		Nombre: "Ana", PrimerApellido: "López",
	})

	if !errors.Is(err, ErrEstadoNegocio) {
		t.Fatalf("un negocio archivado no acepta altas: %v", err)
	}
}

func TestEmpleadoServiceActualizarSinCamposFalla(t *testing.T) {
	service := NewEmpleadoService(empleadoStub(domain.PermisoEmpleadoGestionar))

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.ActualizarEmpleadoInput{})

	if !errors.Is(err, ErrEmpleadoSinCambios) {
		t.Fatalf("se esperaba ErrEmpleadoSinCambios: %v", err)
	}
}

func TestEmpleadoServiceActualizarValidaEstado(t *testing.T) {
	service := NewEmpleadoService(empleadoStub(domain.PermisoEmpleadoGestionar))
	estado := "inventado"

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || validacion.Campos["estado"] == "" {
		t.Fatalf("un estado desconocido debe rechazarse: %v", err)
	}
}

func TestEmpleadoServiceActualizarLimpiaCampoConNull(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	service := NewEmpleadoService(repository)

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.ActualizarEmpleadoInput{Telefono: domain.Optional[string]{Set: true, Value: nil}})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !repository.actualizado.Telefono.Set || repository.actualizado.Telefono.Value != nil {
		t.Fatalf("un null explícito debe limpiar el campo: %#v", repository.actualizado.Telefono)
	}
}

func TestEmpleadoServiceEmpleadoAjenoNoSeEncuentra(t *testing.T) {
	repository := empleadoStub(domain.PermisoEmpleadoGestionar)
	repository.detalle = domain.EmpleadoDetalle{}
	service := NewEmpleadoService(repository)
	nombre := "Ana"

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		domain.ActualizarEmpleadoInput{Nombre: domain.Optional[string]{Set: true, Value: &nombre}})

	if !errors.Is(err, ErrEmpleadoNoEncontrado) {
		t.Fatalf("un empleado de otro negocio no debe encontrarse: %v", err)
	}
}

func TestNombreCompletoEmpleadoOmiteOpcionalesVacios(t *testing.T) {
	vacio := ""
	completo := domain.NombreCompletoEmpleado("Ana", &vacio, "López", texto("García"))

	if completo != "Ana López García" {
		t.Fatalf("nombre completo inesperado: %q", completo)
	}
}
