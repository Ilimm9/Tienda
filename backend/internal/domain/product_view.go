package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrCatalogItemInUse    = errors.New("catalog item is in use")
	ErrCatalogItemNotFound = errors.New("catalog item not found")
	// ErrPricePermissionRequired is returned when an edit changes a branch price without precios.autorizar.
	ErrPricePermissionRequired = errors.New("necesitas permiso para cambiar precios")
)

type CatalogDeletionCheck struct {
	PuedeEliminarse bool     `json:"puede_eliminarse"`
	Vinculaciones   []string `json:"vinculaciones"`
}

type ProductImportJob struct {
	ID           uuid.UUID            `json:"id"`
	Estado       string               `json:"estado"`
	Etapa        string               `json:"etapa"`
	Porcentaje   int                  `json:"porcentaje"`
	MensajeError *string              `json:"mensaje_error,omitempty"`
	Resultado    *CatalogImportResult `json:"resultado,omitempty"`
}

type ProductRow struct {
	ID              uuid.UUID            `json:"id"`
	Nombre          string               `json:"nombre"`
	ImagenURL       *string              `json:"imagen_url"`
	SKU             *string              `json:"sku"`
	Precio          float64              `json:"precio"`
	Stock           float64              `json:"stock"`
	Categoria       *string              `json:"categoria"`
	CategoriaID     *uuid.UUID           `json:"categoria_id"`
	Marca           *string              `json:"marca"`
	MarcaID         *uuid.UUID           `json:"marca_id"`
	Proveedor       *string              `json:"proveedor"`
	ProveedorID     *uuid.UUID           `json:"proveedor_id"`
	Descripcion     *string              `json:"descripcion"`
	Presentacion    *string              `json:"presentacion"`
	Contenido       *float64             `json:"contenido"`
	UnidadContenido *string              `json:"unidad_contenido"`
	UnidadMedida    *string              `json:"unidad_medida"`
	UnidadMedidaID  *uuid.UUID           `json:"unidad_medida_id"`
	CodigoBarras    *string              `json:"codigo_barras"`
	Inventario      []ProductBranchStock `json:"inventario" gorm:"-"`
	Estado          string               `json:"estado"`
	Variantes       []ProductVariantRow  `json:"variantes" gorm:"-"`
}

type ProductVariantAttributeInput struct {
	Nombre string `json:"nombre"`
	Valor  string `json:"valor"`
}

type CreateProductVariantInput struct {
	Atributos         []ProductVariantAttributeInput `json:"atributos"`
	SKUInterno        string                         `json:"sku_interno"`
	GenerarSKUInterno bool                           `json:"generar_sku_interno"`
	PrecioVenta       float64                        `json:"precio_venta"`
	StockInicial      float64                        `json:"stock_inicial"`
	CodigoBarras      *string                        `json:"codigo_barras"`
}

type ProductVariantRow struct {
	ID           uuid.UUID                      `json:"id"`
	SKU          *string                        `json:"sku"`
	Precio       float64                        `json:"precio"`
	Stock        float64                        `json:"stock"`
	CodigoBarras *string                        `json:"codigo_barras"`
	Atributos    []ProductVariantAttributeInput `json:"atributos"`
	Inventario   []ProductBranchStock           `json:"inventario"`
}

type ProductBranchStock struct {
	SucursalID string  `json:"sucursal_id"`
	Sucursal   string  `json:"sucursal"`
	Stock      float64 `json:"stock"`
}

type CatalogOption struct {
	ID     uuid.UUID `json:"id"`
	Nombre string    `json:"nombre"`
}

type CreateProductInput struct {
	Nombre            string                      `json:"nombre" binding:"required"`
	SKUInterno        string                      `json:"sku_interno"`
	GenerarSKUInterno bool                        `json:"generar_sku_interno"`
	MarcaID           *uuid.UUID                  `json:"marca_id"`
	ProveedorID       *uuid.UUID                  `json:"proveedor_id"`
	CategoriaID       *uuid.UUID                  `json:"categoria_id,omitempty"`
	SucursalID        uuid.UUID                   `json:"sucursal_id" binding:"required"`
	Descripcion       *string                     `json:"descripcion"`
	Contenido         *float64                    `json:"contenido"`
	UnidadContenido   *string                     `json:"unidad_contenido"`
	UnidadMedidaID    *uuid.UUID                  `json:"unidad_medida_id"`
	Presentacion      *string                     `json:"presentacion"`
	PrecioVenta       float64                     `json:"precio_venta"`
	CostoProveedor    float64                     `json:"costo_proveedor"`
	StockInicial      float64                     `json:"stock_inicial"`
	CodigoBarras      *string                     `json:"codigo_barras"`
	ImagenURL         *string                     `json:"imagen_url"`
	Variantes         []CreateProductVariantInput `json:"variantes"`
	UsuarioID         uuid.UUID                   `json:"-"`
}

// UpdateProductInput intentionally excludes stock: inventory movements are managed separately from commercial product data.
// SucursalID only scopes the sale price, which lives in precios_sucursal of the active branch.
type UpdateProductInput struct {
	SucursalID   uuid.UUID `json:"sucursal_id" binding:"required"`
	MotivoPrecio *string   `json:"motivo_precio"`
	UsuarioID    uuid.UUID `json:"-"`
	// PuedeCambiarPrecio is resolved by the handler from precios.autorizar.
	PuedeCambiarPrecio bool                        `json:"-"`
	Nombre             string                      `json:"nombre" binding:"required"`
	MarcaID            *uuid.UUID                  `json:"marca_id"`
	ProveedorID        *uuid.UUID                  `json:"proveedor_id"`
	CategoriaID        *uuid.UUID                  `json:"categoria_id"`
	Descripcion        *string                     `json:"descripcion"`
	Contenido          *float64                    `json:"contenido"`
	UnidadContenido    *string                     `json:"unidad_contenido"`
	UnidadMedidaID     *uuid.UUID                  `json:"unidad_medida_id"`
	Presentacion       *string                     `json:"presentacion"`
	PrecioVenta        float64                     `json:"precio_venta"`
	CodigoBarras       *string                     `json:"codigo_barras"`
	ImagenURL          *string                     `json:"imagen_url"`
	Variantes          []UpdateProductVariantInput `json:"variantes"`
}

// UpdateProductVariantInput only changes commercial variant data. SKU and inventory stay owned by their respective flows.
type UpdateProductVariantInput struct {
	ID           *uuid.UUID                     `json:"id"`
	Atributos    []ProductVariantAttributeInput `json:"atributos" binding:"required"`
	PrecioVenta  float64                        `json:"precio_venta"`
	CodigoBarras *string                        `json:"codigo_barras"`
}

type AdjustInventoryInput struct {
	SucursalID uuid.UUID  `json:"sucursal_id" binding:"required"`
	VarianteID *uuid.UUID `json:"variante_id"`
	Stock      float64    `json:"stock" binding:"gte=0"`
}

type ProductLookup struct {
	CodigoBarras    string   `json:"codigo_barras"`
	Nombre          *string  `json:"nombre"`
	Descripcion     *string  `json:"descripcion"`
	Marca           *string  `json:"marca"`
	Categoria       *string  `json:"categoria"`
	PrecioSugerido  *float64 `json:"precio_sugerido"`
	Contenido       *float64 `json:"contenido"`
	UnidadContenido *string  `json:"unidad_contenido"`
	ImagenURL       *string  `json:"imagen_url"`
	Fuentes         []string `json:"fuentes"`
}

type CreateMarcaInput struct {
	Nombre string `json:"nombre" binding:"required"`
}
type UpdateMarcaInput struct {
	Nombre *string `json:"nombre"`
	Activo *bool   `json:"activo"`
}

// CatalogImportIssue identifies a row that could not be imported.
type CatalogImportIssue struct {
	Fila   int    `json:"fila"`
	Nombre string `json:"nombre,omitempty"`
	Campo  string `json:"campo,omitempty"`
	Motivo string `json:"motivo"`
}

// CatalogImportResult is returned after a bulk catalog import.
type CatalogImportResult struct {
	Procesadas                int                  `json:"procesadas"`
	Creadas                   int                  `json:"creadas"`
	Omitidas                  int                  `json:"omitidas"`
	Invalidas                 int                  `json:"invalidas"`
	SKUsGenerados             int                  `json:"skus_generados"`
	ProductosSimplesCreados   int                  `json:"productos_simples_creados"`
	ProductosBaseCreados      int                  `json:"productos_base_creados"`
	ProductosBaseReutilizados int                  `json:"productos_base_reutilizados"`
	VariantesCreadas          int                  `json:"variantes_creadas"`
	CostosFaltantes           int                  `json:"costos_faltantes"`
	AgregadasSucursal         int                  `json:"agregadas_sucursal"`
	Actualizadas              int                  `json:"actualizadas"`
	Errores                   []CatalogImportIssue `json:"errores"`
	Advertencias              []CatalogImportIssue `json:"advertencias"`
}

// ProductImportPreview is the non-mutating result shown before a bulk import
// is confirmed. Insertables have passed all local validation.
type ProductImportPreview struct {
	CatalogImportResult
	Insertables     int                            `json:"insertables"`
	CatalogosNuevos []ProductImportCatalogProposal `json:"catalogos_nuevos"`
	FilasCorrectas  []ProductImportValidRow        `json:"filas_correctas"`
	// Surtido por sucursal (28): filas de productos que ya existen en el negocio.
	FilasAgregarSucursal []ProductImportValidRow `json:"filas_agregar_sucursal"`
	FilasActualizar      []ProductImportValidRow `json:"filas_actualizar"`
}

// ProductImportValidRow summarizes an insertable row in the import preview.
type ProductImportValidRow struct {
	Fila         int     `json:"fila"`
	Nombre       string  `json:"nombre"`
	Categoria    string  `json:"categoria"`
	PrecioVenta  float64 `json:"precio_venta"`
	Costo        float64 `json:"costo"`
	StockInicial float64 `json:"stock_inicial"`
}

// ProductImportCatalogProposal is a new optional catalog value found in a
// product spreadsheet. It must be explicitly approved before import.
type ProductImportCatalogProposal struct {
	Tipo      string `json:"tipo"`
	Nombre    string `json:"nombre"`
	Productos int    `json:"productos"`
}

// ProductImportRow is the spreadsheet representation of a product.
type ProductImportRow struct {
	Fila            int
	Nombre          string
	Categoria       string
	Marca           string
	Proveedor       string
	Descripcion     string
	Presentacion    string
	Contenido       *float64
	UnidadContenido string
	UnidadMedida    string
	PrecioVenta     float64
	CostoProveedor  float64
	// CostoCapturado distingue una celda de costo vacía de un costo 0.
	CostoCapturado bool
	StockInicial   float64
	CodigoBarras   string
	ImagenURL      string
	Variantes      []ProductVariantAttributeInput
}

// Acciones de una fila de importación (capability 28).
const (
	ImportAccionCrear           = "crear"
	ImportAccionAgregarSucursal = "agregar_sucursal"
	ImportAccionActualizar      = "actualizar"
)

type ValidatedProductImportRow struct {
	Fila  int
	Input CreateImportedProductInput
	// Accion vacía equivale a crear. Para agregar o actualizar, ProductoNegocioID
	// es la oferta existente del catálogo.
	Accion            string
	ProductoNegocioID uuid.UUID
}

// CreateImportedProductInput permits the optional fields supported
type CreateImportedProductInput struct {
	Nombre          string
	MarcaID         *uuid.UUID
	ProveedorID     *uuid.UUID
	CategoriaID     *uuid.UUID
	SucursalID      uuid.UUID
	Descripcion     *string
	Contenido       *float64
	UnidadContenido *string
	UnidadMedidaID  *uuid.UUID
	Presentacion    *string
	PrecioVenta     float64
	CostoProveedor  float64
	StockInicial    float64
	CodigoBarras    *string
	ImagenURL       *string
	Variantes       []ProductVariantAttributeInput
	CostoCapturado  bool
	UsuarioID       uuid.UUID
}

type CatalogImportBrandRow struct {
	Fila   int
	Nombre string
}

type CatalogImportCategoryRow struct {
	Fila           int
	Nombre         string
	Descripcion    string
	CategoriaPadre string
}
type CatalogImportProviderRow struct {
	Fila        int
	Nombre      string
	RazonSocial string
	RFC         string
	Telefono    string
	Email       string
	Direccion   string
}
type CatalogImportUnitRow struct {
	Fila        int
	Nombre      string
	Simbolo     string
	Tipo        string
	FactorABase float64
	Decimales   int
}
type CreateCategoriaInput struct {
	Nombre           string     `json:"nombre" binding:"required"`
	CategoriaPadreID *uuid.UUID `json:"categoria_padre_id"`
	Descripcion      *string    `json:"descripcion"`
}
type UpdateCategoriaInput struct {
	Nombre           *string    `json:"nombre"`
	CategoriaPadreID *uuid.UUID `json:"categoria_padre_id"`
	Descripcion      *string    `json:"descripcion"`
	Activo           *bool      `json:"activo"`
}
type CreateProveedorInput struct {
	Nombre      string  `json:"nombre" binding:"required"`
	RazonSocial *string `json:"razon_social"`
	RFC         *string `json:"rfc"`
	Telefono    *string `json:"telefono"`
	Email       *string `json:"email"`
	Direccion   *string `json:"direccion"`
}
type UpdateProveedorInput struct {
	Nombre      *string `json:"nombre"`
	RazonSocial *string `json:"razon_social"`
	RFC         *string `json:"rfc"`
	Telefono    *string `json:"telefono"`
	Email       *string `json:"email"`
	Direccion   *string `json:"direccion"`
	Activo      *bool   `json:"activo"`
}

type CreateUnidadMedidaInput struct {
	Nombre          string     `json:"nombre" binding:"required"`
	Simbolo         string     `json:"simbolo" binding:"required"`
	Tipo            string     `json:"tipo" binding:"required"`
	UnidadBaseID    *uuid.UUID `json:"unidad_base_id"`
	FactorABase     float64    `json:"factor_a_base"`
	PermiteFraccion *bool      `json:"permite_fraccion"`
	Decimales       int        `json:"decimales"`
}
type UpdateUnidadMedidaInput struct {
	Nombre          *string    `json:"nombre"`
	Simbolo         *string    `json:"simbolo"`
	Tipo            *string    `json:"tipo"`
	UnidadBaseID    *uuid.UUID `json:"unidad_base_id"`
	FactorABase     *float64   `json:"factor_a_base"`
	PermiteFraccion *bool      `json:"permite_fraccion"`
	Decimales       *int       `json:"decimales"`
	Activo          *bool      `json:"activo"`
}

// ErrAlreadyInBranch: the product already belongs to the branch assortment (capability 28).
var ErrAlreadyInBranch = errors.New("este producto ya existe en esta sucursal")

// ProductInOtherBranchError: creating a simple product that already exists in the
// business catalog but not in the active branch; the client may add it instead.
type ProductInOtherBranchError struct {
	ProductoID uuid.UUID
	Nombre     string
}

func (e *ProductInOtherBranchError) Error() string {
	return "este producto ya existe en otra sucursal"
}

// BranchCatalogProduct is a business product that is not yet in a branch.
type BranchCatalogProduct struct {
	ID                 uuid.UUID `json:"id"`
	Nombre             string    `json:"nombre"`
	Presentacion       *string   `json:"presentacion"`
	SKU                *string   `json:"sku"`
	CodigoBarras       *string   `json:"codigo_barras"`
	Categoria          *string   `json:"categoria"`
	VariantesFaltantes int       `json:"variantes_faltantes"`
}

// AddToBranchInput adds catalog products to a branch without price or stock.
type AddToBranchInput struct {
	SucursalID  uuid.UUID   `json:"sucursal_id" binding:"required"`
	ProductoIDs []uuid.UUID `json:"producto_ids"`
	Todos       bool        `json:"todos"`
}

// AddProductToBranchInput adds one existing simple product to a branch with the
// amounts captured in the product form.
type AddProductToBranchInput struct {
	SucursalID   uuid.UUID `json:"sucursal_id" binding:"required"`
	PrecioVenta  float64   `json:"precio_venta"`
	Costo        float64   `json:"costo"`
	StockInicial float64   `json:"stock_inicial"`
	UsuarioID    uuid.UUID `json:"-"`
}
