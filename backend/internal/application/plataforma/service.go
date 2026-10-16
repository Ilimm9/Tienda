package plataforma

import (
	"context"
	"errors"
	"strings"
	"time"

	domain "tienda/backend/internal/domain/plataforma"

	"github.com/google/uuid"
)

var (
	ErrAccesoDenegado      = errors.New("acceso de plataforma denegado")
	ErrNegocioNoEncontrado = errors.New("negocio no encontrado")
)

type Repository interface {
	RolesUsuario(ctx context.Context, usuarioID uuid.UUID) ([]string, error)
	Resumen(ctx context.Context, desde, hasta time.Time) (domain.Resumen, error)
	ListarNegocios(ctx context.Context, query, estado string, page, pageSize int) (domain.PaginaNegocios, error)
	ObtenerNegocio(ctx context.Context, negocioID uuid.UUID, desde, hasta time.Time) (domain.NegocioDetalle, error)
	ListarUsuarios(ctx context.Context, negocioID uuid.UUID) ([]domain.UsuarioResumen, error)
	RegistrarActividad(ctx context.Context, usuarioID, negocioID uuid.UUID, at time.Time) error
	LimpiarActividad(ctx context.Context, antes time.Time) error
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) RolesUsuario(ctx context.Context, usuarioID uuid.UUID) ([]string, error) {
	return s.repository.RolesUsuario(ctx, usuarioID)
}

func (s *Service) Autorizar(ctx context.Context, usuarioID uuid.UUID) error {
	roles, err := s.repository.RolesUsuario(ctx, usuarioID)
	if err != nil {
		return err
	}
	for _, role := range roles {
		if role == domain.RolAdminLectura {
			return nil
		}
	}
	return ErrAccesoDenegado
}

func (s *Service) Resumen(ctx context.Context, desde, hasta time.Time) (domain.Resumen, error) {
	return s.repository.Resumen(ctx, normalizarDesde(desde), normalizarHasta(hasta))
}

func (s *Service) ListarNegocios(ctx context.Context, query, estado string, page, pageSize int) (domain.PaginaNegocios, error) {
	query = strings.TrimSpace(query)
	if estado != "" && estado != "activo" && estado != "archivado" {
		estado = ""
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.repository.ListarNegocios(ctx, query, estado, page, pageSize)
}

func (s *Service) ObtenerNegocio(ctx context.Context, negocioID uuid.UUID, desde, hasta time.Time) (domain.NegocioDetalle, error) {
	return s.repository.ObtenerNegocio(ctx, negocioID, normalizarDesde(desde), normalizarHasta(hasta))
}

func (s *Service) ListarUsuarios(ctx context.Context, negocioID uuid.UUID) ([]domain.UsuarioResumen, error) {
	return s.repository.ListarUsuarios(ctx, negocioID)
}

func (s *Service) RegistrarActividad(ctx context.Context, usuarioID, negocioID uuid.UUID) error {
	if usuarioID == uuid.Nil || negocioID == uuid.Nil {
		return nil
	}
	return s.repository.RegistrarActividad(ctx, usuarioID, negocioID, time.Now().UTC())
}

func (s *Service) LimpiarActividad() error {
	return s.repository.LimpiarActividad(context.Background(), time.Now().UTC().AddDate(-1, -1, 0))
}

func normalizarHasta(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now().UTC()
	}
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
}

func normalizarDesde(value time.Time) time.Time {
	if value.IsZero() {
		value = time.Now().UTC().AddDate(0, 0, -29)
	}
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}
