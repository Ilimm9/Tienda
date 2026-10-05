package precio

import (
	"github.com/google/uuid"
	"tienda/backend/internal/domain/precio"
)

type PropuestaResumen struct {
	precio.Propuesta
	Producto string `json:"producto"`
}
type AutorizarInput struct {
	Costo       float64 `json:"costo"`
	Margen      float64 `json:"margen"`
	PrecioVenta float64 `json:"precio_venta"`
}
type Repository interface {
	ListPending(uuid.UUID, uuid.UUID) ([]PropuestaResumen, error)
	Authorize(uuid.UUID, uuid.UUID, uuid.UUID, AutorizarInput) (precio.Propuesta, error)
	BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error)
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) ListPending(businessID, branchID uuid.UUID) ([]PropuestaResumen, error) {
	return s.repo.ListPending(businessID, branchID)
}
func (s *Service) Authorize(businessID, userID, proposalID uuid.UUID, input AutorizarInput) (precio.Propuesta, error) {
	return s.repo.Authorize(businessID, userID, proposalID, input)
}

// BranchOfProposal permite comprobar el acceso a la sucursal antes de autorizar una propuesta.
func (s *Service) BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error) {
	return s.repo.BranchOfProposal(businessID, proposalID)
}
