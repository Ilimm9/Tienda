package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Sucursal struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Nombre        string    `gorm:"type:varchar(180);not null" json:"nombre"`
	Direccion     *string   `gorm:"type:text" json:"direccion,omitempty"`
	Telefono      *string   `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
	Negocio       Negocio   `gorm:"foreignKey:NegocioID" json:"-"`
}

type Marca struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre        string    `gorm:"type:varchar(180);not null" json:"nombre"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Producto struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre            string     `gorm:"type:varchar(255);not null" json:"nombre"`
	Descripcion       *string    `gorm:"type:text" json:"descripcion,omitempty"`
	MarcaID           *uuid.UUID `gorm:"type:uuid;index" json:"marca_id,omitempty"`
	Contenido         *float64   `json:"contenido,omitempty"`
	UnidadContenido   *string    `gorm:"type:varchar(30)" json:"unidad_contenido,omitempty"`
	Presentacion      *string    `gorm:"type:varchar(100)" json:"presentacion,omitempty"`
	RequiereCaducidad bool       `gorm:"not null;default:false" json:"requiere_caducidad"`
	EsPerecedero      bool       `gorm:"not null;default:false" json:"es_perecedero"`
	Activo            bool       `gorm:"not null;default:true" json:"activo"`
	CreadoEn          time.Time  `json:"creado_en"`
	ActualizadoEn     time.Time  `json:"actualizado_en"`
	Marca             *Marca     `gorm:"foreignKey:MarcaID" json:"-"`
}

type Categoria struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre           string     `gorm:"type:varchar(180);not null" json:"nombre"`
	CategoriaPadreID *uuid.UUID `gorm:"type:uuid;index" json:"categoria_padre_id,omitempty"`
	Descripcion      *string    `gorm:"type:text" json:"descripcion,omitempty"`
	Activo           bool       `gorm:"not null;default:true" json:"activo"`
	CreadoEn         time.Time  `json:"creado_en"`
	ActualizadoEn    time.Time  `json:"actualizado_en"`
	CategoriaPadre   *Categoria `gorm:"foreignKey:CategoriaPadreID" json:"-"`
}

type ProductoCategoria struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_producto_categoria" json:"producto_id"`
	CategoriaID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_producto_categoria" json:"categoria_id"`
	EsPrincipal bool      `gorm:"not null;default:false" json:"es_principal"`
	CreadoEn    time.Time `json:"creado_en"`
	Producto    Producto  `gorm:"foreignKey:ProductoID" json:"-"`
	Categoria   Categoria `gorm:"foreignKey:CategoriaID" json:"-"`
}

type ProductoCodigo struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoID    uuid.UUID `gorm:"type:uuid;not null;index" json:"producto_id"`
	Tipo          string    `gorm:"type:varchar(20);not null" json:"tipo"`
	Codigo        string    `gorm:"type:varchar(120);uniqueIndex;not null" json:"codigo"`
	EsPrincipal   bool      `gorm:"not null;default:false" json:"es_principal"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
	Producto      Producto  `gorm:"foreignKey:ProductoID" json:"-"`
}

type ProductoImagen struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoID    uuid.UUID `gorm:"type:uuid;not null;index" json:"producto_id"`
	URL           string    `gorm:"type:text;not null" json:"url"`
	EsPrincipal   bool      `gorm:"not null;default:false" json:"es_principal"`
	Orden         int       `gorm:"not null;default:0" json:"orden"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
	Producto      Producto  `gorm:"foreignKey:ProductoID" json:"-"`
}

type ProductoUnidad struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoID    uuid.UUID `gorm:"type:uuid;not null;index" json:"producto_id"`
	Nombre        string    `gorm:"type:varchar(80);not null" json:"nombre"`
	Factor        float64   `gorm:"not null" json:"factor"`
	CodigoBarras  *string   `gorm:"type:varchar(120)" json:"codigo_barras,omitempty"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
	Producto      Producto  `gorm:"foreignKey:ProductoID" json:"-"`
}

type ProductoNegocio struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID              uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_producto_negocio" json:"negocio_id"`
	ProductoID             uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_producto_negocio" json:"producto_id"`
	SKUInterno             *string   `gorm:"type:varchar(120)" json:"sku_interno,omitempty"`
	PrecioVenta            float64   `gorm:"type:numeric(14,2);not null" json:"precio_venta"`
	PrecioMayoreo          *float64  `gorm:"type:numeric(14,2)" json:"precio_mayoreo,omitempty"`
	CostoReferencia        *float64  `gorm:"type:numeric(14,2)" json:"costo_referencia,omitempty"`
	PrecioIncluyeImpuestos bool      `gorm:"not null;default:true" json:"precio_incluye_impuestos"`
	Activo                 bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn               time.Time `json:"creado_en"`
	ActualizadoEn          time.Time `json:"actualizado_en"`
	Negocio                Negocio   `gorm:"foreignKey:NegocioID" json:"-"`
	Producto               Producto  `gorm:"foreignKey:ProductoID" json:"-"`
}

type Impuesto struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre        string    `gorm:"type:varchar(120);not null" json:"nombre"`
	Tipo          string    `gorm:"type:varchar(30);not null" json:"tipo"`
	Tasa          float64   `gorm:"type:numeric(7,4);not null" json:"tasa"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type ProductoImpuesto struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoNegocioID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_producto_impuesto" json:"producto_negocio_id"`
	ImpuestoID        uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_producto_impuesto" json:"impuesto_id"`
	CreadoEn          time.Time       `json:"creado_en"`
	ProductoNegocio   ProductoNegocio `gorm:"foreignKey:ProductoNegocioID" json:"-"`
	Impuesto          Impuesto        `gorm:"foreignKey:ImpuestoID" json:"-"`
}

type InventarioSucursal struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	SucursalID        uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_inventario_sucursal_producto" json:"sucursal_id"`
	ProductoNegocioID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_inventario_sucursal_producto" json:"producto_negocio_id"`
	StockActual       float64         `gorm:"not null;default:0" json:"stock_actual"`
	StockMinimo       float64         `gorm:"not null;default:0" json:"stock_minimo"`
	StockMaximo       *float64        `json:"stock_maximo,omitempty"`
	Ubicacion         *string         `gorm:"type:varchar(120)" json:"ubicacion,omitempty"`
	CreadoEn          time.Time       `json:"creado_en"`
	ActualizadoEn     time.Time       `json:"actualizado_en"`
	Sucursal          Sucursal        `gorm:"foreignKey:SucursalID" json:"-"`
	ProductoNegocio   ProductoNegocio `gorm:"foreignKey:ProductoNegocioID" json:"-"`
}

type Proveedor struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Nombre        string    `gorm:"type:varchar(180);not null" json:"nombre"`
	RazonSocial   *string   `gorm:"type:varchar(180)" json:"razon_social,omitempty"`
	RFC           *string   `gorm:"type:varchar(30)" json:"rfc,omitempty"`
	Telefono      *string   `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	Email         *string   `gorm:"type:varchar(255)" json:"email,omitempty"`
	Direccion     *string   `gorm:"type:text" json:"direccion,omitempty"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
	Negocio       Negocio   `gorm:"foreignKey:NegocioID" json:"-"`
}

type ProductoProveedor struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoNegocioID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_producto_proveedor" json:"producto_negocio_id"`
	ProveedorID       uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_producto_proveedor" json:"proveedor_id"`
	CodigoProveedor   *string         `gorm:"type:varchar(120)" json:"codigo_proveedor,omitempty"`
	Costo             *float64        `gorm:"type:numeric(14,2)" json:"costo,omitempty"`
	EsPrincipal       bool            `gorm:"not null;default:false" json:"es_principal"`
	CreadoEn          time.Time       `json:"creado_en"`
	ActualizadoEn     time.Time       `json:"actualizado_en"`
	ProductoNegocio   ProductoNegocio `gorm:"foreignKey:ProductoNegocioID" json:"-"`
	Proveedor         Proveedor       `gorm:"foreignKey:ProveedorID" json:"-"`
}

type Lote struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	SucursalID        uuid.UUID       `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	ProductoNegocioID uuid.UUID       `gorm:"type:uuid;not null;index" json:"producto_negocio_id"`
	ProveedorID       *uuid.UUID      `gorm:"type:uuid;index" json:"proveedor_id,omitempty"`
	NumeroLote        *string         `gorm:"type:varchar(120)" json:"numero_lote,omitempty"`
	CantidadInicial   float64         `gorm:"not null" json:"cantidad_inicial"`
	CantidadActual    float64         `gorm:"not null" json:"cantidad_actual"`
	CostoUnitario     *float64        `gorm:"type:numeric(14,2)" json:"costo_unitario,omitempty"`
	FechaFabricacion  *time.Time      `json:"fecha_fabricacion,omitempty"`
	FechaCaducidad    *time.Time      `json:"fecha_caducidad,omitempty"`
	FechaRecepcion    time.Time       `gorm:"not null" json:"fecha_recepcion"`
	Activo            bool            `gorm:"not null;default:true" json:"activo"`
	CreadoEn          time.Time       `json:"creado_en"`
	ActualizadoEn     time.Time       `json:"actualizado_en"`
	Sucursal          Sucursal        `gorm:"foreignKey:SucursalID" json:"-"`
	ProductoNegocio   ProductoNegocio `gorm:"foreignKey:ProductoNegocioID" json:"-"`
	Proveedor         *Proveedor      `gorm:"foreignKey:ProveedorID" json:"-"`
}

type MovimientoInventario struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	SucursalID        uuid.UUID       `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	ProductoNegocioID uuid.UUID       `gorm:"type:uuid;not null;index" json:"producto_negocio_id"`
	LoteID            *uuid.UUID      `gorm:"type:uuid;index" json:"lote_id,omitempty"`
	Tipo              string          `gorm:"type:varchar(40);not null" json:"tipo"`
	Cantidad          float64         `gorm:"not null" json:"cantidad"`
	StockAnterior     float64         `gorm:"not null" json:"stock_anterior"`
	StockNuevo        float64         `gorm:"not null" json:"stock_nuevo"`
	CostoUnitario     *float64        `gorm:"type:numeric(14,2)" json:"costo_unitario,omitempty"`
	Motivo            *string         `gorm:"type:text" json:"motivo,omitempty"`
	Referencia        *string         `gorm:"type:varchar(180)" json:"referencia,omitempty"`
	UsuarioID         *uuid.UUID      `gorm:"type:uuid;index" json:"usuario_id,omitempty"`
	CreadoEn          time.Time       `json:"creado_en"`
	Sucursal          Sucursal        `gorm:"foreignKey:SucursalID" json:"-"`
	ProductoNegocio   ProductoNegocio `gorm:"foreignKey:ProductoNegocioID" json:"-"`
	Lote              *Lote           `gorm:"foreignKey:LoteID" json:"-"`
	Usuario           *Usuario        `gorm:"foreignKey:UsuarioID" json:"-"`
}

func (Sucursal) TableName() string             { return "sucursales" }
func (Marca) TableName() string                { return "marcas" }
func (Producto) TableName() string             { return "productos" }
func (Categoria) TableName() string            { return "categorias" }
func (ProductoCategoria) TableName() string    { return "producto_categorias" }
func (ProductoCodigo) TableName() string       { return "producto_codigos" }
func (ProductoImagen) TableName() string       { return "producto_imagenes" }
func (ProductoUnidad) TableName() string       { return "producto_unidades" }
func (ProductoNegocio) TableName() string      { return "producto_negocio" }
func (Impuesto) TableName() string             { return "impuestos" }
func (ProductoImpuesto) TableName() string     { return "producto_impuestos" }
func (InventarioSucursal) TableName() string   { return "inventario_sucursal" }
func (Proveedor) TableName() string            { return "proveedores" }
func (ProductoProveedor) TableName() string    { return "producto_proveedor" }
func (Lote) TableName() string                 { return "lotes" }
func (MovimientoInventario) TableName() string { return "movimientos_inventario" }

func (s *Sucursal) BeforeCreate(*gorm.DB) error             { setID(&s.ID); return nil }
func (m *Marca) BeforeCreate(*gorm.DB) error                { setID(&m.ID); return nil }
func (p *Producto) BeforeCreate(*gorm.DB) error             { setID(&p.ID); return nil }
func (c *Categoria) BeforeCreate(*gorm.DB) error            { setID(&c.ID); return nil }
func (p *ProductoCategoria) BeforeCreate(*gorm.DB) error    { setID(&p.ID); return nil }
func (p *ProductoCodigo) BeforeCreate(*gorm.DB) error       { setID(&p.ID); return nil }
func (p *ProductoImagen) BeforeCreate(*gorm.DB) error       { setID(&p.ID); return nil }
func (p *ProductoUnidad) BeforeCreate(*gorm.DB) error       { setID(&p.ID); return nil }
func (p *ProductoNegocio) BeforeCreate(*gorm.DB) error      { setID(&p.ID); return nil }
func (i *Impuesto) BeforeCreate(*gorm.DB) error             { setID(&i.ID); return nil }
func (p *ProductoImpuesto) BeforeCreate(*gorm.DB) error     { setID(&p.ID); return nil }
func (i *InventarioSucursal) BeforeCreate(*gorm.DB) error   { setID(&i.ID); return nil }
func (p *Proveedor) BeforeCreate(*gorm.DB) error            { setID(&p.ID); return nil }
func (p *ProductoProveedor) BeforeCreate(*gorm.DB) error    { setID(&p.ID); return nil }
func (l *Lote) BeforeCreate(*gorm.DB) error                 { setID(&l.ID); return nil }
func (m *MovimientoInventario) BeforeCreate(*gorm.DB) error { setID(&m.ID); return nil }
