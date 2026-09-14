package negocio

import (
	"context"
	"errors"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type rolRepositoryStub struct {
	access             domain.ContextoNegocioSucursal
	permisos           []string
	detalle            domain.RolDetalle
	existeCodigo       bool
	miembrosConRol     int
	perteneceMembresia bool
	quedaPropietario   bool
	creado             domain.CrearRolInput
	actualizado        domain.ActualizarRolInput
	rolesAsignados     []uuid.UUID
	eliminado          bool
	err                error
}

func (r *rolRepositoryStub) ObtenerContextoNegocio(context.Context, uuid.UUID, uuid.UUID) (domain.ContextoNegocioSucursal, error) {
	return r.access, r.err
}

func (r *rolRepositoryStub) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return r.permisos, nil
}

func (r *rolRepositoryStub) ListarPermisos(context.Context) ([]domain.Permiso, error) {
	return domain.CatalogoPermisos, nil
}

func (r *rolRepositoryStub) ListarRoles(context.Context, uuid.UUID, bool) ([]domain.RolResumen, error) {
	return []domain.RolResumen{}, nil
}

func (r *rolRepositoryStub) ObtenerRol(context.Context, uuid.UUID, uuid.UUID) (domain.RolDetalle, error) {
	if r.detalle.ID == uuid.Nil {
		return domain.RolDetalle{}, ErrRolNoEncontrado
	}
	return r.detalle, nil
}

func (r *rolRepositoryStub) ExisteCodigoRol(context.Context, uuid.UUID, string) (bool, error) {
	return r.existeCodigo, nil
}

func (r *rolRepositoryStub) CrearRol(_ context.Context, _ uuid.UUID, _ uuid.UUID, input domain.CrearRolInput) (uuid.UUID, error) {
	r.creado = input
	return r.detalle.ID, nil
}

func (r *rolRepositoryStub) ActualizarRol(_ context.Context, _, _ uuid.UUID, input domain.ActualizarRolInput) error {
	r.actualizado = input
	return nil
}

func (r *rolRepositoryStub) EliminarRol(context.Context, uuid.UUID, uuid.UUID) error {
	r.eliminado = true
	return nil
}

func (r *rolRepositoryStub) ContarMiembrosConRol(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return r.miembrosConRol, nil
}

func (r *rolRepositoryStub) ListarMiembros(context.Context, uuid.UUID) ([]domain.MiembroRoles, error) {
	return []domain.MiembroRoles{}, nil
}

func (r *rolRepositoryStub) ReemplazarRolesDeMembresia(_ context.Context, _, _ uuid.UUID, roles []uuid.UUID, _ uuid.UUID) error {
	r.rolesAsignados = roles
	return nil
}

func (r *rolRepositoryStub) MembresiaPerteneceANegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.perteneceMembresia, nil
}

func (r *rolRepositoryStub) QuedaPropietarioConRolSistema(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) (bool, error) {
	return r.quedaPropietario, nil
}

func negocioActivoCon(permisos ...string) *rolRepositoryStub {
	return &rolRepositoryStub{
		access:   domain.ContextoNegocioSucursal{EstadoNegocio: "activo", TipoMiembro: "miembro"},
		permisos: permisos,
	}
}

func TestRolServiceAutorizarAceptaPermisoConcedido(t *testing.T) {
	service := NewRolService(negocioActivoCon(domain.PermisoRolVer))

	if err := service.Autorizar(context.Background(), uuid.New(), uuid.New(), domain.PermisoRolVer); err != nil {
		t.Fatalf("el permiso concedido no debe fallar: %v", err)
	}
}

func TestRolServiceAutorizarRechazaPermisoAusente(t *testing.T) {
	service := NewRolService(negocioActivoCon(domain.PermisoRolVer))

	err := service.Autorizar(context.Background(), uuid.New(), uuid.New(), domain.PermisoRolGestionar)

	if !errors.Is(err, ErrRolProhibido) {
		t.Fatalf("un permiso ausente debe rechazarse: %v", err)
	}
}

func TestRolServiceListarExigePermisoDeLectura(t *testing.T) {
	service := NewRolService(negocioActivoCon())

	_, err := service.Listar(context.Background(), uuid.New(), uuid.New(), false)

	if !errors.Is(err, ErrRolProhibido) {
		t.Fatalf("sin permiso de lectura debe rechazarse: %v", err)
	}
}

func TestRolServiceMiembroConPermisoPuedeGestionar(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	service := NewRolService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearRolInput{
		Codigo: "cajero", Nombre: "Cajero",
	})

	if err != nil {
		t.Fatalf("un miembro con permiso ya no depende de tipo_miembro: %v", err)
	}
	if repository.creado.Codigo != "CAJERO" {
		t.Fatalf("el código debe normalizarse en mayúsculas: %q", repository.creado.Codigo)
	}
}

func TestRolServiceCrearValidaCodigoYNombre(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	service := NewRolService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearRolInput{
		Codigo: "@@", Nombre: "X",
	})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) || validacion.Campos["codigo"] == "" || validacion.Campos["nombre"] == "" {
		t.Fatalf("se esperaban errores de código y nombre: %v", err)
	}
}

func TestRolServiceCrearRechazaCodigoReservado(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	service := NewRolService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearRolInput{
		Codigo: "propietario", Nombre: "Copia del sistema",
	})

	var validacion *ErrorValidacion
	if !errors.As(err, &validacion) {
		t.Fatalf("el código del rol de sistema está reservado: %v", err)
	}
}

func TestRolServiceCrearRechazaCodigoDuplicado(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	repository.existeCodigo = true
	service := NewRolService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearRolInput{
		Codigo: "CAJERO", Nombre: "Cajero",
	})

	if !errors.Is(err, ErrRolConflicto) {
		t.Fatalf("el código es único por negocio: %v", err)
	}
}

func TestRolServiceNoPermiteEditarRolDeSistema(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New(), EsRolSistema: true}
	service := NewRolService(repository)
	nombre := "Otro nombre"

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ActualizarRolInput{
		Nombre: domain.Optional[string]{Set: true, Value: &nombre},
	})

	if !errors.Is(err, ErrRolSistemaInmutable) {
		t.Fatalf("el rol de sistema es inmutable: %v", err)
	}
}

func TestRolServiceNoPermiteEliminarRolDeSistema(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New(), EsRolSistema: true}
	service := NewRolService(repository)

	err := service.Eliminar(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrRolSistemaInmutable) || repository.eliminado {
		t.Fatalf("el rol de sistema no puede eliminarse: %v", err)
	}
}

func TestRolServiceNoEliminaRolConMiembros(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	repository.miembrosConRol = 2
	service := NewRolService(repository)

	err := service.Eliminar(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrRolEnUso) || repository.eliminado {
		t.Fatalf("un rol con miembros no puede eliminarse: %v", err)
	}
}

func TestRolServiceActualizarSinCamposFalla(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	service := NewRolService(repository)

	_, err := service.Actualizar(context.Background(), uuid.New(), uuid.New(), uuid.New(), domain.ActualizarRolInput{})

	if !errors.Is(err, ErrRolSinCambios) {
		t.Fatalf("se esperaba ErrRolSinCambios: %v", err)
	}
}

func TestRolServiceNegocioArchivadoBloqueaEscritura(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolGestionar)
	repository.access.EstadoNegocio = "archivado"
	repository.detalle = domain.RolDetalle{ID: uuid.New()}
	service := NewRolService(repository)

	_, err := service.Crear(context.Background(), uuid.New(), uuid.New(), domain.CrearRolInput{
		Codigo: "CAJERO", Nombre: "Cajero",
	})

	if !errors.Is(err, ErrEstadoNegocio) {
		t.Fatalf("un negocio archivado no acepta mutaciones: %v", err)
	}
}

func TestRolServiceAsignarRolesRechazaMembresiaAjena(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolAsignar)
	repository.perteneceMembresia = false
	service := NewRolService(repository)

	err := service.AsignarRoles(context.Background(), uuid.New(), uuid.New(), uuid.New(), []uuid.UUID{uuid.New()})

	if !errors.Is(err, ErrMembresiaNoEncontrada) {
		t.Fatalf("una membresía de otro negocio no debe encontrarse: %v", err)
	}
}

func TestRolServiceAsignarRolesProtegeAlUltimoPropietario(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolAsignar)
	repository.perteneceMembresia = true
	repository.quedaPropietario = false
	service := NewRolService(repository)

	err := service.AsignarRoles(context.Background(), uuid.New(), uuid.New(), uuid.New(), []uuid.UUID{})

	if !errors.Is(err, ErrPropietarioSinRol) {
		t.Fatalf("el negocio no puede quedarse sin propietario efectivo: %v", err)
	}
}

func TestRolServiceAsignarRolesDeduplica(t *testing.T) {
	repository := negocioActivoCon(domain.PermisoRolAsignar)
	repository.perteneceMembresia = true
	repository.quedaPropietario = true
	service := NewRolService(repository)
	rolID := uuid.New()

	err := service.AsignarRoles(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		[]uuid.UUID{rolID, rolID, uuid.Nil})

	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(repository.rolesAsignados) != 1 || repository.rolesAsignados[0] != rolID {
		t.Fatalf("los roles deben deduplicarse y descartar UUID nulos: %v", repository.rolesAsignados)
	}
}
