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
type AjustarPrecioInput struct {
	SucursalID  uuid.UUID `json:"sucursal_id"`
	PrecioVenta float64   `json:"precio_venta"`
	Motivo      string    `json:"motivo"`
}
type Repository interface {
	ListPending(uuid.UUID, uuid.UUID) ([]PropuestaResumen, error)
	Authorize(uuid.UUID, uuid.UUID, uuid.UUID, AutorizarInput) (precio.Propuesta, error)
	BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error)
	ListHistory(uuid.UUID, uuid.UUID, uuid.UUID, int) ([]precio.HistorialCostoPrecioSucursal, error)
	AdjustPrice(uuid.UUID, uuid.UUID, uuid.UUID, AjustarPrecioInput) (precio.HistorialCostoPrecioSucursal, error)
}

func (s *Service) ListHistory(businessID, branchID, productID uuid.UUID, limit int) ([]precio.HistorialCostoPrecioSucursal, error) {
	return s.repo.ListHistory(businessID, branchID, productID, limit)
}
func (s *Service) AdjustPrice(businessID, userID, productID uuid.UUID, input AjustarPrecioInput) (precio.HistorialCostoPrecioSucursal, error) {
	return s.repo.AdjustPrice(businessID, userID, productID, input)
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
