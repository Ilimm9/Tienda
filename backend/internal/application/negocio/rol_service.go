package negocio

import (
	"context"
	"errors"
	"regexp"
	"strings"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrRolNoEncontrado     = errors.New("rol no encontrado")
	ErrRolProhibido        = errors.New("no tienes permiso para realizar esta acción")
	ErrRolSinCambios       = errors.New("no se recibieron campos para actualizar")
	ErrRolConflicto        = errors.New("ya existe un rol con ese código")
	ErrRolSistemaInmutable = errors.New("el rol de sistema no puede modificarse ni eliminarse")
	ErrRolEnUso            = errors.New("el rol tiene miembros asignados")
	ErrMembresiaNoEncontrada = errors.New("membresía no encontrada")
	ErrPropietarioSinRol   = errors.New("el negocio debe conservar al menos un propietario con el rol de sistema")
)

var codigoRolPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{1,59}$`)

type RolRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	ListarPermisos(ctx context.Context) ([]domain.Permiso, error)
	ListarRoles(ctx context.Context, negocioID uuid.UUID, incluirInactivos bool) ([]domain.RolResumen, error)
	ObtenerRol(ctx context.Context, negocioID, rolID uuid.UUID) (domain.RolDetalle, error)
	ExisteCodigoRol(ctx context.Context, negocioID uuid.UUID, codigo string) (bool, error)
	CrearRol(ctx context.Context, negocioID uuid.UUID, creadoPor uuid.UUID, input domain.CrearRolInput) (uuid.UUID, error)
	ActualizarRol(ctx context.Context, negocioID, rolID uuid.UUID, input domain.ActualizarRolInput) error
	EliminarRol(ctx context.Context, negocioID, rolID uuid.UUID) error
	ContarMiembrosConRol(ctx context.Context, negocioID, rolID uuid.UUID) (int, error)
	ListarMiembros(ctx context.Context, negocioID uuid.UUID) ([]domain.MiembroRoles, error)
	ReemplazarRolesDeMembresia(ctx context.Context, negocioID, membresiaID uuid.UUID, roles []uuid.UUID, asignadoPor uuid.UUID) error
	MembresiaPerteneceANegocio(ctx context.Context, negocioID, membresiaID uuid.UUID) (bool, error)
	QuedaPropietarioConRolSistema(ctx context.Context, negocioID, membresiaID uuid.UUID, roles []uuid.UUID) (bool, error)
}

type RolService struct {
	roles RolRepository
}

func NewRolService(roles RolRepository) *RolService {
	return &RolService{roles: roles}
}

// PermisosEfectivos resuelve la cadena usuario -> membresía -> roles -> permisos.
func (s *RolService) PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error) {
	return s.roles.PermisosEfectivos(ctx, usuarioID, negocioID)
}

// Autorizar falla cuando la membresía activa no concentra el permiso solicitado.
func (s *RolService) Autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string) error {
	permisos, err := s.roles.PermisosEfectivos(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	for _, actual := range permisos {
		if actual == permiso {
			return nil
		}
	}
	return ErrRolProhibido
}

func (s *RolService) ListarPermisos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]domain.Permiso, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolVer, false); err != nil {
		return nil, err
	}
	return s.roles.ListarPermisos(ctx)
}

func (s *RolService) Listar(ctx context.Context, usuarioID, negocioID uuid.UUID, incluirInactivos bool) ([]domain.RolResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolVer, false); err != nil {
		return nil, err
	}
	return s.roles.ListarRoles(ctx, negocioID, incluirInactivos)
}

func (s *RolService) Obtener(ctx context.Context, usuarioID, negocioID, rolID uuid.UUID) (domain.RolDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolVer, false); err != nil {
		return domain.RolDetalle{}, err
	}
	return s.roles.ObtenerRol(ctx, negocioID, rolID)
}

func (s *RolService) Crear(ctx context.Context, usuarioID, negocioID uuid.UUID, input domain.CrearRolInput) (domain.RolDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolGestionar, true); err != nil {
		return domain.RolDetalle{}, err
	}
	normalizarCrearRol(&input)
	if err := validarCrearRol(input); err != nil {
		return domain.RolDetalle{}, err
	}
	existe, err := s.roles.ExisteCodigoRol(ctx, negocioID, input.Codigo)
	if err != nil {
		return domain.RolDetalle{}, err
	}
	if existe {
		return domain.RolDetalle{}, ErrRolConflicto
	}
	rolID, err := s.roles.CrearRol(ctx, negocioID, usuarioID, input)
	if err != nil {
		return domain.RolDetalle{}, err
	}
	return s.roles.ObtenerRol(ctx, negocioID, rolID)
}

func (s *RolService) Actualizar(ctx context.Context, usuarioID, negocioID, rolID uuid.UUID, input domain.ActualizarRolInput) (domain.RolDetalle, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolGestionar, true); err != nil {
		return domain.RolDetalle{}, err
	}
	actual, err := s.roles.ObtenerRol(ctx, negocioID, rolID)
	if err != nil {
		return domain.RolDetalle{}, err
	}
	// El rol de sistema conserva siempre todos los permisos: no se edita ni se desactiva.
	if actual.EsRolSistema {
		return domain.RolDetalle{}, ErrRolSistemaInmutable
	}
	if !input.Nombre.Set && !input.Descripcion.Set && !input.Activo.Set && !input.Permisos.Set {
		return domain.RolDetalle{}, ErrRolSinCambios
	}
	if err := normalizarYValidarActualizarRol(&input); err != nil {
		return domain.RolDetalle{}, err
	}
	if err := s.roles.ActualizarRol(ctx, negocioID, rolID, input); err != nil {
		return domain.RolDetalle{}, err
	}
	return s.roles.ObtenerRol(ctx, negocioID, rolID)
}

func (s *RolService) Eliminar(ctx context.Context, usuarioID, negocioID, rolID uuid.UUID) error {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolGestionar, true); err != nil {
		return err
	}
	actual, err := s.roles.ObtenerRol(ctx, negocioID, rolID)
	if err != nil {
		return err
	}
	if actual.EsRolSistema {
		return ErrRolSistemaInmutable
	}
	miembros, err := s.roles.ContarMiembrosConRol(ctx, negocioID, rolID)
	if err != nil {
		return err
	}
	if miembros > 0 {
		return ErrRolEnUso
	}
	return s.roles.EliminarRol(ctx, negocioID, rolID)
}

func (s *RolService) ListarMiembros(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]domain.MiembroRoles, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolVer, false); err != nil {
		return nil, err
	}
	return s.roles.ListarMiembros(ctx, negocioID)
}

func (s *RolService) AsignarRoles(ctx context.Context, usuarioID, negocioID, membresiaID uuid.UUID, roles []uuid.UUID) error {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoRolAsignar, true); err != nil {
		return err
	}
	pertenece, err := s.roles.MembresiaPerteneceANegocio(ctx, negocioID, membresiaID)
	if err != nil {
		return err
	}
	if !pertenece {
		return ErrMembresiaNoEncontrada
	}
	roles = rolesUnicos(roles)
	// Invariante heredada de fase 1: el negocio nunca queda sin propietario efectivo.
	queda, err := s.roles.QuedaPropietarioConRolSistema(ctx, negocioID, membresiaID, roles)
	if err != nil {
		return err
	}
	if !queda {
		return ErrPropietarioSinRol
	}
	return s.roles.ReemplazarRolesDeMembresia(ctx, negocioID, membresiaID, roles, usuarioID)
}

// autorizar exige membresía activa y, para escritura, negocio activo más el permiso indicado.
func (s *RolService) autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string, escritura bool) error {
	access, err := s.roles.ObtenerContextoNegocio(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if escritura && access.EstadoNegocio != "activo" {
		return ErrEstadoNegocio
	}
	return s.Autorizar(ctx, usuarioID, negocioID, permiso)
}

func normalizarCrearRol(input *domain.CrearRolInput) {
	input.Codigo = strings.ToUpper(strings.TrimSpace(input.Codigo))
	input.Nombre = strings.TrimSpace(input.Nombre)
	input.Descripcion = limpiarOpcional(input.Descripcion, false)
	input.Permisos = rolesUnicos(input.Permisos)
}

func validarCrearRol(input domain.CrearRolInput) error {
	campos := map[string]string{}
	if !codigoRolPattern.MatchString(input.Codigo) {
		campos["codigo"] = "usa entre 2 y 60 caracteres: letras, números, guion o guion bajo"
	}
	if strings.EqualFold(input.Codigo, domain.CodigoRolPropietario) {
		campos["codigo"] = "ese código está reservado para el rol de sistema"
	}
	if longitud := len([]rune(input.Nombre)); longitud < 2 || longitud > 120 {
		campos["nombre"] = "usa entre 2 y 120 caracteres"
	}
	if len(campos) > 0 {
		return &ErrorValidacion{Campos: campos}
	}
	return nil
}

func normalizarYValidarActualizarRol(input *domain.ActualizarRolInput) error {
	campos := map[string]string{}
	if input.Nombre.Set {
		if input.Nombre.Value == nil {
			campos["nombre"] = "no puede quedar vacío"
		} else {
			nombre := strings.TrimSpace(*input.Nombre.Value)
			if longitud := len([]rune(nombre)); longitud < 2 || longitud > 120 {
				campos["nombre"] = "usa entre 2 y 120 caracteres"
			}
			input.Nombre.Value = &nombre
		}
	}
	if input.Permisos.Set && input.Permisos.Value != nil {
		unicos := rolesUnicos(*input.Permisos.Value)
		input.Permisos.Value = &unicos
	}
	if len(campos) > 0 {
		return &ErrorValidacion{Campos: campos}
	}
	return nil
}

func rolesUnicos(valores []uuid.UUID) []uuid.UUID {
	vistos := make(map[uuid.UUID]struct{}, len(valores))
	unicos := make([]uuid.UUID, 0, len(valores))
	for _, valor := range valores {
		if valor == uuid.Nil {
			continue
		}
		if _, existe := vistos[valor]; existe {
			continue
		}
		vistos[valor] = struct{}{}
		unicos = append(unicos, valor)
	}
	return unicos
}
