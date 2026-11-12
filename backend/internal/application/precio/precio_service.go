package precio

import (
	"time"

	"github.com/google/uuid"
	"tienda/backend/internal/domain/precio"
)

type PropuestaResumen struct {
	precio.Propuesta
	Producto string `json:"producto"`
	// Datos de la compra para agrupar Validar costos por ticket.
	CompraID       uuid.UUID `json:"compra_id"`
	TipoDocumento  string    `json:"tipo_documento"`
	FolioDocumento string    `json:"folio_documento"`
	ProveedorID    uuid.UUID `json:"proveedor_id"`
	Proveedor      string    `json:"proveedor"`
	FechaDocumento time.Time `json:"fecha_documento"`
	FechaRecepcion time.Time `json:"fecha_recepcion"`
	TotalCompra    float64   `json:"total_compra"`
	// Datos del renglón para la nota.
	CodigoBarras        *string `json:"codigo_barras"`
	Unidad              string  `json:"unidad"`
	CantidadEmpaques    int     `json:"cantidad_empaques"`
	PiezasDanadas       int     `json:"piezas_danadas"`
	PiezasFaltantes     int     `json:"piezas_faltantes"`
	Descuento           float64 `json:"descuento"`
	ImpuestosPorRenglon bool    `json:"impuestos_por_renglon"`
	TasaIEPS            float64 `json:"tasa_ieps"`
	TasaIVA             float64 `json:"tasa_iva"`
	CompraIEPS          float64 `json:"compra_ieps"`
	CompraIVA           float64 `json:"compra_iva"`
	ProductoNuevo       bool    `json:"producto_nuevo"`
	Nota                string  `json:"nota"`
}

// AutorizarRenglonInput es un renglón de «Validar ticket».
type AutorizarRenglonInput struct {
	PropuestaID uuid.UUID `json:"propuesta_id"`
	AutorizarInput
}

// AutorizarCompraInput valida varios renglones de una compra en una sola transacción.
type AutorizarCompraInput struct {
	Renglones []AutorizarRenglonInput `json:"renglones"`
}

// FiltroPendientes acota Validar costos por proveedor, fecha de recepción y texto (folio, proveedor o producto).
type FiltroPendientes struct {
	ProveedorID *uuid.UUID
	Desde       *time.Time
	Hasta       *time.Time
	Texto       string
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
	ListPending(uuid.UUID, uuid.UUID, FiltroPendientes) ([]PropuestaResumen, error)
	Authorize(uuid.UUID, uuid.UUID, uuid.UUID, AutorizarInput) (precio.Propuesta, error)
	BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error)
	BranchOfPurchase(businessID, purchaseID uuid.UUID) (uuid.UUID, error)
	AuthorizePurchase(businessID, userID, purchaseID uuid.UUID, input AutorizarCompraInput) ([]precio.Propuesta, error)
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
func (s *Service) ListPending(businessID, branchID uuid.UUID, filtro FiltroPendientes) ([]PropuestaResumen, error) {
	return s.repo.ListPending(businessID, branchID, filtro)
}
func (s *Service) Authorize(businessID, userID, proposalID uuid.UUID, input AutorizarInput) (precio.Propuesta, error) {
	return s.repo.Authorize(businessID, userID, proposalID, input)
}

// BranchOfProposal permite comprobar el acceso a la sucursal antes de autorizar una propuesta.
func (s *Service) BranchOfProposal(businessID, proposalID uuid.UUID) (uuid.UUID, error) {
	return s.repo.BranchOfProposal(businessID, proposalID)
}

func (s *Service) BranchOfPurchase(businessID, purchaseID uuid.UUID) (uuid.UUID, error) {
	return s.repo.BranchOfPurchase(businessID, purchaseID)
}

// AuthorizePurchase valida el ticket entero: todos los renglones o ninguno.
func (s *Service) AuthorizePurchase(businessID, userID, purchaseID uuid.UUID, input AutorizarCompraInput) ([]precio.Propuesta, error) {
	return s.repo.AuthorizePurchase(businessID, userID, purchaseID, input)
}
