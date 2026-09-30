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
}

type CrearCompraInput struct {
	ProveedorID    string              `json:"proveedor_id"`
	SucursalID     string              `json:"sucursal_id"`
	TipoDocumento  string              `json:"tipo_documento"`
	FolioDocumento string              `json:"folio_documento"`
	NumeroCuenta   *string             `json:"numero_cuenta"`
	FechaDocumento string              `json:"fecha_documento"`
	FechaRecepcion string              `json:"fecha_recepcion"`
	FormaPago      string              `json:"forma_pago"`
	IncluyeIEPS    bool                `json:"incluye_ieps"`
	IEPS           float64             `json:"ieps"`
	IncluyeIVA     bool                `json:"incluye_iva"`
	Observaciones  *string             `json:"observaciones"`
	Detalles       []CrearDetalleInput `json:"detalles"`
}

type Repository interface {
	List(uuid.UUID) ([]compra.CompraProveedor, error)
	Get(uuid.UUID, uuid.UUID) (compra.CompraProveedor, error)
	ListProductos(uuid.UUID, uuid.UUID) ([]ProductoComprable, error)
	Create(uuid.UUID, uuid.UUID, CrearCompraInput) (compra.CompraProveedor, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(businessID uuid.UUID) ([]compra.CompraProveedor, error) {
	return s.repo.List(businessID)
}
func (s *Service) Get(businessID, id uuid.UUID) (compra.CompraProveedor, error) {
	return s.repo.Get(businessID, id)
}
func (s *Service) ListProductos(businessID, providerID uuid.UUID) ([]ProductoComprable, error) {
	return s.repo.ListProductos(businessID, providerID)
}
func (s *Service) Create(businessID, userID uuid.UUID, input CrearCompraInput) (compra.CompraProveedor, error) {
	return s.repo.Create(businessID, userID, input)
}

func ParseDate(value string) (time.Time, error) { return time.Parse("2006-01-02", value) }
