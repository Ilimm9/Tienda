package negocio

import (
	"context"
	"errors"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

var ErrVistasSoloPropietario = errors.New("Solo el propietario puede cambiar las vistas.")

type VistaRepository interface {
	SeccionesOcultas(ctx context.Context, negocioID uuid.UUID) (string, error)
	GuardarSeccionesOcultas(ctx context.Context, negocioID uuid.UUID, valor string) error
}

// VistaService administra Configuración > Vistas: qué secciones ve el equipo del negocio.
type VistaService struct {
	vistas   VistaRepository
	contexto ContextoRepository
}

func NewVistaService(vistas VistaRepository, contexto ContextoRepository) *VistaService {
	return &VistaService{vistas: vistas, contexto: contexto}
}

func (s *VistaService) Obtener(ctx context.Context, negocioID uuid.UUID) (domain.VistasNegocio, error) {
	valor, err := s.vistas.SeccionesOcultas(ctx, negocioID)
	if err != nil {
		return domain.VistasNegocio{}, err
	}
	return domain.VistasNegocio{SeccionesOcultas: domain.LeerSeccionesOcultas(valor), SeccionesOcultables: domain.SeccionesOcultables()}, nil
}

// Guardar solo lo permite al propietario del negocio; el cambio aplica a todo el equipo.
func (s *VistaService) Guardar(ctx context.Context, usuarioID, negocioID uuid.UUID, secciones []string) (domain.VistasNegocio, error) {
	propietario, err := s.esPropietario(ctx, usuarioID, negocioID)
	if err != nil {
		return domain.VistasNegocio{}, err
	}
	if !propietario {
		return domain.VistasNegocio{}, ErrVistasSoloPropietario
	}
	normalizadas, err := domain.NormalizarSeccionesOcultas(secciones)
	if err != nil {
		return domain.VistasNegocio{}, err
	}
	if err := s.vistas.GuardarSeccionesOcultas(ctx, negocioID, domain.EscribirSeccionesOcultas(normalizadas)); err != nil {
		return domain.VistasNegocio{}, err
	}
	return domain.VistasNegocio{SeccionesOcultas: normalizadas, SeccionesOcultables: domain.SeccionesOcultables()}, nil
}

func (s *VistaService) esPropietario(ctx context.Context, usuarioID, negocioID uuid.UUID) (bool, error) {
	opciones, err := s.contexto.ListarOpcionesContexto(ctx, usuarioID)
	if err != nil {
		return false, err
	}
	for _, opcion := range opciones {
		if opcion.ID == negocioID {
			return opcion.TipoMiembro == "propietario", nil
		}
	}
	return false, nil
}
