package negocio

import (
	"context"
	"errors"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrAsignacionNoEncontrada = errors.New("asignación no encontrada")
	ErrAsignacionProhibida    = errors.New("no tienes permiso para realizar esta acción")
	ErrAsignacionDuplicada    = errors.New("el empleado ya está asignado a esa sucursal")
	ErrAsignacionSucursalAjena = errors.New("la sucursal no pertenece a este negocio")
)

type AsignacionRepository interface {
	ObtenerContextoNegocio(ctx context.Context, usuarioID, negocioID uuid.UUID) (domain.ContextoNegocioSucursal, error)
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
	ObtenerEmpleado(ctx context.Context, negocioID, empleadoID uuid.UUID) (domain.EmpleadoDetalle, error)
	SucursalActivaDelNegocio(ctx context.Context, negocioID, sucursalID uuid.UUID) (bool, error)
	Listar(ctx context.Context, negocioID, empleadoID uuid.UUID, incluirFinalizadas bool) ([]domain.AsignacionResumen, error)
	ExisteActiva(ctx context.Context, negocioID, empleadoID, sucursalID uuid.UUID) (bool, error)
	Asignar(ctx context.Context, negocioID, empleadoID uuid.UUID, input domain.AsignarSucursalInput) (uuid.UUID, error)
	EstablecerPrincipal(ctx context.Context, negocioID, empleadoID, asignacionID uuid.UUID) error
	Finalizar(ctx context.Context, negocioID, empleadoID, asignacionID uuid.UUID) error
}

type AsignacionService struct {
	asignaciones AsignacionRepository
}

func NewAsignacionService(asignaciones AsignacionRepository) *AsignacionService {
	return &AsignacionService{asignaciones: asignaciones}
}

func (s *AsignacionService) Listar(ctx context.Context, usuarioID, negocioID, empleadoID uuid.UUID, incluirFinalizadas bool) ([]domain.AsignacionResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoAsignacionVer, false); err != nil {
		return nil, err
	}
	// Confirma que el empleado pertenece al negocio antes de exponer sus sucursales.
	if _, err := s.asignaciones.ObtenerEmpleado(ctx, negocioID, empleadoID); err != nil {
		return nil, err
	}
	return s.asignaciones.Listar(ctx, negocioID, empleadoID, incluirFinalizadas)
}

func (s *AsignacionService) Asignar(ctx context.Context, usuarioID, negocioID, empleadoID uuid.UUID, input domain.AsignarSucursalInput) ([]domain.AsignacionResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoAsignacionEditar, true); err != nil {
		return nil, err
	}
	if _, err := s.asignaciones.ObtenerEmpleado(ctx, negocioID, empleadoID); err != nil {
		return nil, err
	}
	if input.SucursalID == uuid.Nil {
		return nil, &ErrorValidacion{Campos: map[string]string{"sucursal_id": "es obligatorio"}}
	}
	// Una sucursal de otro negocio nunca puede asignarse, aunque el UUID exista.
	pertenece, err := s.asignaciones.SucursalActivaDelNegocio(ctx, negocioID, input.SucursalID)
	if err != nil {
		return nil, err
	}
	if !pertenece {
		return nil, ErrAsignacionSucursalAjena
	}
	existe, err := s.asignaciones.ExisteActiva(ctx, negocioID, empleadoID, input.SucursalID)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, ErrAsignacionDuplicada
	}
	if _, err := s.asignaciones.Asignar(ctx, negocioID, empleadoID, input); err != nil {
		return nil, err
	}
	return s.asignaciones.Listar(ctx, negocioID, empleadoID, false)
}

func (s *AsignacionService) EstablecerPrincipal(ctx context.Context, usuarioID, negocioID, empleadoID, asignacionID uuid.UUID) ([]domain.AsignacionResumen, error) {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoAsignacionEditar, true); err != nil {
		return nil, err
	}
	if err := s.asignaciones.EstablecerPrincipal(ctx, negocioID, empleadoID, asignacionID); err != nil {
		return nil, err
	}
	return s.asignaciones.Listar(ctx, negocioID, empleadoID, false)
}

func (s *AsignacionService) Finalizar(ctx context.Context, usuarioID, negocioID, empleadoID, asignacionID uuid.UUID) error {
	if err := s.autorizar(ctx, usuarioID, negocioID, domain.PermisoAsignacionEditar, true); err != nil {
		return err
	}
	return s.asignaciones.Finalizar(ctx, negocioID, empleadoID, asignacionID)
}

func (s *AsignacionService) autorizar(ctx context.Context, usuarioID, negocioID uuid.UUID, permiso string, escritura bool) error {
	access, err := s.asignaciones.ObtenerContextoNegocio(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if escritura && access.EstadoNegocio != "activo" {
		return ErrEstadoNegocio
	}
	permisos, err := s.asignaciones.PermisosEfectivos(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	for _, actual := range permisos {
		if actual == permiso {
			return nil
		}
	}
	return ErrAsignacionProhibida
}
