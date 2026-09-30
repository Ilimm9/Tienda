package compra

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DocumentoTicket  = "TICKET"
	DocumentoFactura = "FACTURA"
)

type CompraProveedor struct {
	ID                    uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID             uuid.UUID                `gorm:"type:uuid;not null;index" json:"negocio_id"`
	ProveedorID           uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_compra_documento" json:"proveedor_id"`
	SucursalID            uuid.UUID                `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	TipoDocumento         string                   `gorm:"type:varchar(20);not null;uniqueIndex:idx_compra_documento" json:"tipo_documento"`
	FolioDocumento        string                   `gorm:"type:varchar(120);not null;uniqueIndex:idx_compra_documento" json:"folio_documento"`
	NumeroCuenta          *string                  `gorm:"type:varchar(120)" json:"numero_cuenta,omitempty"`
	FechaDocumento        time.Time                `gorm:"type:date;not null" json:"fecha_documento"`
	FechaRecepcion        time.Time                `gorm:"type:date;not null" json:"fecha_recepcion"`
	FormaPago             string                   `gorm:"type:varchar(30);not null" json:"forma_pago"`
	ImportePagado         float64                  `gorm:"type:numeric(14,2);not null" json:"importe_pagado"`
	Subtotal              float64                  `gorm:"type:numeric(14,2);not null" json:"subtotal"`
	IEPS                  float64                  `gorm:"type:numeric(14,2);not null;default:0" json:"ieps"`
	IVA                   float64                  `gorm:"type:numeric(14,2);not null;default:0" json:"iva"`
	Impuestos             float64                  `gorm:"type:numeric(14,2);not null;default:0" json:"impuestos"`
	Total                 float64                  `gorm:"type:numeric(14,2);not null" json:"total"`
	Observaciones         *string                  `gorm:"type:text" json:"observaciones,omitempty"`
	CapturadoPorUsuarioID uuid.UUID                `gorm:"type:uuid;not null;index" json:"capturado_por_usuario_id"`
	RecibidoPorUsuarioID  uuid.UUID                `gorm:"type:uuid;not null;index" json:"recibido_por_usuario_id"`
	CreadoEn              time.Time                `json:"creado_en"`
	Detalles              []CompraProveedorDetalle `gorm:"foreignKey:CompraID" json:"detalles,omitempty"`
}

type CompraProveedorDetalle struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompraID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"compra_id"`
	ProductoNegocioID uuid.UUID  `gorm:"type:uuid;not null;index" json:"producto_negocio_id"`
	UnidadMedidaID    uuid.UUID  `gorm:"type:uuid;not null" json:"unidad_medida_id"`
	CodigoBarras      *string    `gorm:"type:varchar(120)" json:"codigo_barras,omitempty"`
	CantidadEmpaques  int        `gorm:"not null" json:"cantidad_empaques"`
	PiezasPorEmpaque  int        `gorm:"not null" json:"piezas_por_empaque"`
	TotalPiezas       int        `gorm:"not null" json:"total_piezas"`
	PiezasDanadas     int        `gorm:"not null;default:0" json:"piezas_danadas"`
	PiezasFaltantes   int        `gorm:"not null;default:0" json:"piezas_faltantes"`
	PiezasBuenas      int        `gorm:"not null" json:"piezas_buenas"`
	CostoEmpaque      float64    `gorm:"type:numeric(14,2);not null" json:"costo_empaque"`
	CostoPieza        float64    `gorm:"type:numeric(14,4);not null" json:"costo_pieza"`
	Descuento         float64    `gorm:"type:numeric(14,2);not null;default:0" json:"descuento"`
	Subtotal          float64    `gorm:"type:numeric(14,2);not null" json:"subtotal"`
	NumeroLote        *string    `gorm:"type:varchar(120)" json:"numero_lote,omitempty"`
	FechaCaducidad    *time.Time `gorm:"type:date" json:"fecha_caducidad,omitempty"`
	ProductoNuevo     bool       `gorm:"not null;default:false" json:"producto_nuevo"`
	CreadoEn          time.Time  `json:"creado_en"`
}

func (CompraProveedor) TableName() string        { return "compras_proveedor" }
func (CompraProveedorDetalle) TableName() string { return "compras_proveedor_detalles" }
func (c *CompraProveedor) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
func (d *CompraProveedorDetalle) BeforeCreate(*gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
