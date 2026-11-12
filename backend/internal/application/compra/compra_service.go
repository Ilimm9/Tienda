package compra

import (
	"tienda/backend/internal/domain/compra"
	"time"

	"github.com/google/uuid"
)

type ProductoComprable struct {
	ID           uuid.UUID `json:"id"`
	Nombre       string    `json:"nombre"`
	SKU          *string   `json:"sku"`
	CodigoBarras *string   `json:"codigo_barras"`
	DelProveedor bool      `json:"del_proveedor"`
	// Tasas recordadas del producto; nil si nunca se capturaron.
	TasaIEPS *float64 `json:"tasa_ieps"`
	TasaIVA  *float64 `json:"tasa_iva"`
}

// CompraResumen es un renglón de la lista: la compra con proveedor, piezas y estado de sus costos.
type CompraResumen struct {
	compra.CompraProveedor
	Proveedor        string `json:"proveedor"`
	Productos        int    `json:"productos"`
	Piezas           int    `json:"piezas"`
	CostosPendientes int    `json:"costos_pendientes"`
	CostosValidados  int    `json:"costos_validados"`
	CostosCancelados int    `json:"costos_cancelados"`
}

type CrearDetalleInput struct {
	ProductoNegocioID   string  `json:"producto_negocio_id"`
	UnidadMedidaID      string  `json:"unidad_medida_id"`
	CantidadEmpaques    int     `json:"cantidad_empaques"`
	CostoEmpaque        float64 `json:"costo_empaque"`
	Descuento           float64 `json:"descuento"`
	PiezasDanadas       int     `json:"piezas_danadas"`
	PiezasFaltantes     int     `json:"piezas_faltantes"`
	NumeroLote          *string `json:"numero_lote"`
	FechaCaducidad      *string `json:"fecha_caducidad"`
	ProductoNuevo       bool    `json:"producto_nuevo"`
	NombreProductoNuevo string  `json:"nombre_producto_nuevo"`
	CodigoBarrasNuevo   string  `json:"codigo_barras_nuevo"`
	// Tasas por renglón (fracción). Si algún renglón las trae, la compra calcula impuestos por renglón.
	TasaIEPS *float64 `json:"tasa_ieps"`
	TasaIVA  *float64 `json:"tasa_iva"`
}

type CrearCompraInput struct {
	ProveedorID    string  `json:"proveedor_id"`
	SucursalID     string  `json:"sucursal_id"`
	TipoDocumento  string  `json:"tipo_documento"`
	FolioDocumento string  `json:"folio_documento"`
	NumeroCuenta   *string `json:"numero_cuenta"`
	FechaDocumento string  `json:"fecha_documento"`
	FechaRecepcion string  `json:"fecha_recepcion"`
	FormaPago      string  `json:"forma_pago"`
	IncluyeIEPS    bool    `json:"incluye_ieps"`
	IEPS           float64 `json:"ieps"`
	IncluyeIVA     bool    `json:"incluye_iva"`
	// IEPSAjustado es el importe de IEPS del ticket; la diferencia se reparte entre los renglones con IEPS.
	IEPSAjustado  *float64            `json:"ieps_ajustado"`
	Observaciones *string             `json:"observaciones"`
	Detalles      []CrearDetalleInput `json:"detalles"`
}

type Repository interface {
	List(uuid.UUID, uuid.UUID) ([]CompraResumen, error)
	Get(uuid.UUID, uuid.UUID) (compra.CompraProveedor, error)
	ListProductos(uuid.UUID, uuid.UUID) ([]ProductoComprable, error)
	Create(uuid.UUID, uuid.UUID, CrearCompraInput) (compra.CompraProveedor, error)
	ListExport(uuid.UUID, FiltroExportacion) ([]CompraExportacion, error)
	View(uuid.UUID, uuid.UUID) (CompraVista, error)
	Gestor
}

type Service struct {
	repo       Repository
	exportador Exportador
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(businessID, branchID uuid.UUID) ([]CompraResumen, error) {
	return s.repo.List(businessID, branchID)
}
func (s *Service) Get(businessID, id uuid.UUID) (compra.CompraProveedor, error) {
	item, err := s.repo.Get(businessID, id)
	if err == nil {
		item.CompletarImpuestosRenglon()
	}
	return item, err
}

// PorRenglon indica si la compra trae tasas por renglón (cliente nuevo).
func (input CrearCompraInput) PorRenglon() bool {
	for _, detalle := range input.Detalles {
		if detalle.TasaIEPS != nil || detalle.TasaIVA != nil {
			return true
		}
	}
	return false
}
func (s *Service) ListProductos(businessID, providerID uuid.UUID) ([]ProductoComprable, error) {
	return s.repo.ListProductos(businessID, providerID)
}
func (s *Service) Create(businessID, userID uuid.UUID, input CrearCompraInput) (compra.CompraProveedor, error) {
	return s.repo.Create(businessID, userID, input)
}

func ParseDate(value string) (time.Time, error) { return time.Parse("2006-01-02", value) }
