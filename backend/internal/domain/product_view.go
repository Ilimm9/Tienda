package domain

import "github.com/google/uuid"

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
	Descripcion     *string              `json:"descripcion"`
	Presentacion    *string              `json:"presentacion"`
	Contenido       *float64             `json:"contenido"`
	UnidadContenido *string              `json:"unidad_contenido"`
	UnidadMedida    *string              `json:"unidad_medida"`
	UnidadMedidaID  *uuid.UUID           `json:"unidad_medida_id"`
	CodigoBarras    *string              `json:"codigo_barras"`
	Inventario      []ProductBranchStock `json:"inventario" gorm:"-"`
	Estado          string               `json:"estado"`
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
	Nombre          string     `json:"nombre" binding:"required"`
	SKUInterno      string     `json:"sku_interno" binding:"required"`
	MarcaID         *uuid.UUID `json:"marca_id"`
	CategoriaID     uuid.UUID  `json:"categoria_id" binding:"required"`
	SucursalID      uuid.UUID  `json:"sucursal_id" binding:"required"`
	Descripcion     *string    `json:"descripcion"`
	Contenido       *float64   `json:"contenido"`
	UnidadContenido *string    `json:"unidad_contenido"`
	UnidadMedidaID  *uuid.UUID `json:"unidad_medida_id"`
	Presentacion    *string    `json:"presentacion"`
	PrecioVenta     float64    `json:"precio_venta"`
	StockInicial    float64    `json:"stock_inicial"`
	CodigoBarras    *string    `json:"codigo_barras"`
	ImagenURL       *string    `json:"imagen_url"`
}

// UpdateProductInput intentionally excludes stock and branch assignment: inventory
// movements are managed separately from commercial product data.
type UpdateProductInput struct {
	Nombre          string     `json:"nombre" binding:"required"`
	SKUInterno      string     `json:"sku_interno" binding:"required"`
	MarcaID         *uuid.UUID `json:"marca_id"`
	CategoriaID     uuid.UUID  `json:"categoria_id" binding:"required"`
	Descripcion     *string    `json:"descripcion"`
	Contenido       *float64   `json:"contenido"`
	UnidadContenido *string    `json:"unidad_contenido"`
	UnidadMedidaID  *uuid.UUID `json:"unidad_medida_id"`
	Presentacion    *string    `json:"presentacion"`
	PrecioVenta     float64    `json:"precio_venta"`
	CodigoBarras    *string    `json:"codigo_barras"`
	ImagenURL       *string    `json:"imagen_url"`
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
	Campo  string `json:"campo,omitempty"`
	Motivo string `json:"motivo"`
}

// CatalogImportResult is returned after a bulk catalog import.
type CatalogImportResult struct {
	Procesadas   int                  `json:"procesadas"`
	Creadas      int                  `json:"creadas"`
	Omitidas     int                  `json:"omitidas"`
	Invalidas    int                  `json:"invalidas"`
	Errores      []CatalogImportIssue `json:"errores"`
	Advertencias []CatalogImportIssue `json:"advertencias"`
}

// ProductImportPreview is the non-mutating result shown before a bulk import
// is confirmed. Insertables have passed all local validation.
type ProductImportPreview struct {
	CatalogImportResult
	Insertables int `json:"insertables"`
}

// ProductImportRow is the spreadsheet representation of a product.
type ProductImportRow struct {
	Fila            int
	Nombre          string
	SKUInterno      string
	Categoria       string
	Marca           string
	Descripcion     string
	Presentacion    string
	Contenido       *float64
	UnidadContenido string
	UnidadMedida    string
	PrecioVenta     float64
	StockInicial    float64
	CodigoBarras    string
	ImagenURL       string
}

type ValidatedProductImportRow struct {
	Fila  int
	Input CreateImportedProductInput
}

// CreateImportedProductInput permits the optional fields supported
type CreateImportedProductInput struct {
	Nombre          string
	SKUInterno      *string
	MarcaID         *uuid.UUID
	CategoriaID     *uuid.UUID
	SucursalID      uuid.UUID
	Descripcion     *string
	Contenido       *float64
	UnidadContenido *string
	UnidadMedidaID  *uuid.UUID
	Presentacion    *string
	PrecioVenta     float64
	StockInicial    float64
	CodigoBarras    *string
	ImagenURL       *string
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
type CatalogImportUnitRow struct {
	Fila        int
	Codigo      string
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
	Codigo          string     `json:"codigo" binding:"required"`
	Nombre          string     `json:"nombre" binding:"required"`
	Simbolo         string     `json:"simbolo" binding:"required"`
	Tipo            string     `json:"tipo" binding:"required"`
	UnidadBaseID    *uuid.UUID `json:"unidad_base_id"`
	FactorABase     float64    `json:"factor_a_base"`
	PermiteFraccion *bool      `json:"permite_fraccion"`
	Decimales       int        `json:"decimales"`
}
type UpdateUnidadMedidaInput struct {
	Codigo          *string    `json:"codigo"`
	Nombre          *string    `json:"nombre"`
	Simbolo         *string    `json:"simbolo"`
	Tipo            *string    `json:"tipo"`
	UnidadBaseID    *uuid.UUID `json:"unidad_base_id"`
	FactorABase     *float64   `json:"factor_a_base"`
	PermiteFraccion *bool      `json:"permite_fraccion"`
	Decimales       *int       `json:"decimales"`
	Activo          *bool      `json:"activo"`
}
