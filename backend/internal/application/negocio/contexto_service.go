package negocio

import (
	"context"
	"errors"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var (
	ErrContextoNegocioNoDisponible  = errors.New("negocio no disponible")
	ErrContextoSucursalNoDisponible = errors.New("sucursal no disponible")
)

type ContextoRepository interface {
	ListarOpcionesContexto(context.Context, uuid.UUID) ([]domain.ContextoNegocio, error)
	NegocioActivoAccesible(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	// SucursalActivaAccesible exige además asignación vigente cuando el usuario no es propietario.
	SucursalActivaAccesible(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) (bool, error)
	SucursalesAccesibles(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]uuid.UUID, error)
}

type ContextoService struct {
	contexto ContextoRepository
}

func NewContextoService(contexto ContextoRepository) *ContextoService {
	return &ContextoService{contexto: contexto}
}

func (s *ContextoService) ListarOpciones(ctx context.Context, usuarioID uuid.UUID) ([]domain.ContextoNegocio, error) {
	return s.contexto.ListarOpcionesContexto(ctx, usuarioID)
}

func (s *ContextoService) ValidarNegocioActivo(ctx context.Context, usuarioID, negocioID uuid.UUID) error {
	ok, err := s.contexto.NegocioActivoAccesible(ctx, usuarioID, negocioID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrContextoNegocioNoDisponible
	}
	return nil
}

// ValidarSucursalActiva responde igual para una sucursal inexistente, ajena o no asignada: no revela cuál es el caso.
func (s *ContextoService) ValidarSucursalActiva(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) error {
	ok, err := s.contexto.SucursalActivaAccesible(ctx, usuarioID, negocioID, sucursalID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrContextoSucursalNoDisponible
	}
	return nil
}

// SucursalesAccesibles devuelve las sucursales activas donde el usuario puede operar en el negocio.
func (s *ContextoService) SucursalesAccesibles(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]uuid.UUID, error) {
	return s.contexto.SucursalesAccesibles(ctx, usuarioID, negocioID)
}
