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
	SucursalActivaDelNegocio(context.Context, uuid.UUID, uuid.UUID) (bool, error)
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

func (s *ContextoService) ValidarSucursalActiva(ctx context.Context, negocioID, sucursalID uuid.UUID) error {
	ok, err := s.contexto.SucursalActivaDelNegocio(ctx, negocioID, sucursalID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrContextoSucursalNoDisponible
	}
	return nil
}
