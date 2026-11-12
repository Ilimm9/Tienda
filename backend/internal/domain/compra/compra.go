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

// Estados de la compra. Una compra suspendida devolvió sus piezas y se puede reactivar.
const (
	EstadoActiva     = "ACTIVA"
	EstadoSuspendida = "SUSPENDIDA"
)

// Tipos de evento de la bitácora de compras.
const (
	EventoEdicion      = "EDICION"
	EventoSuspension   = "SUSPENSION"
	EventoReactivacion = "REACTIVACION"
)

// Movimientos de inventario que agrega la edición, suspensión y reactivación.
// El movimiento original COMPRA_ENTRADA nunca se modifica.
const (
	MovimientoCompraEntrada      = "COMPRA_ENTRADA"
	MovimientoCompraAjuste       = "COMPRA_AJUSTE"
	MovimientoCompraSuspension   = "COMPRA_SUSPENSION"
	MovimientoCompraReactivacion = "COMPRA_REACTIVACION"
)

type CompraProveedor struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID             uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	ProveedorID           uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_compra_documento" json:"proveedor_id"`
	SucursalID            uuid.UUID `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	TipoDocumento         string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_compra_documento" json:"tipo_documento"`
	FolioDocumento        string    `gorm:"type:varchar(120);not null;uniqueIndex:idx_compra_documento" json:"folio_documento"`
	NumeroCuenta          *string   `gorm:"type:varchar(120)" json:"numero_cuenta,omitempty"`
	FechaDocumento        time.Time `gorm:"type:date;not null" json:"fecha_documento"`
	FechaRecepcion        time.Time `gorm:"type:date;not null" json:"fecha_recepcion"`
	FormaPago             string    `gorm:"type:varchar(30);not null" json:"forma_pago"`
	ImportePagado         float64   `gorm:"type:numeric(14,2);not null" json:"importe_pagado"`
	Subtotal              float64   `gorm:"type:numeric(14,2);not null" json:"subtotal"`
	IEPS                  float64   `gorm:"type:numeric(14,2);not null;default:0" json:"ieps"`
	IVA                   float64   `gorm:"type:numeric(14,2);not null;default:0" json:"iva"`
	Impuestos             float64   `gorm:"type:numeric(14,2);not null;default:0" json:"impuestos"`
	Total                 float64   `gorm:"type:numeric(14,2);not null" json:"total"`
	Observaciones         *string   `gorm:"type:text" json:"observaciones,omitempty"`
	CapturadoPorUsuarioID uuid.UUID `gorm:"type:uuid;not null;index" json:"capturado_por_usuario_id"`
	RecibidoPorUsuarioID  uuid.UUID `gorm:"type:uuid;not null;index" json:"recibido_por_usuario_id"`
	CreadoEn              time.Time `json:"creado_en"`
	// Columnas aditivas: estado de la compra y quién la suspendió, reactivó o editó.
	Estado                 string                   `gorm:"type:varchar(20);not null;default:'ACTIVA';index" json:"estado"`
	MotivoSuspension       *string                  `gorm:"type:text" json:"motivo_suspension,omitempty"`
	SuspendidoPorUsuarioID *uuid.UUID               `gorm:"type:uuid" json:"suspendido_por_usuario_id,omitempty"`
	SuspendidoEn           *time.Time               `json:"suspendido_en,omitempty"`
	ReactivadoPorUsuarioID *uuid.UUID               `gorm:"type:uuid" json:"reactivado_por_usuario_id,omitempty"`
	ReactivadoEn           *time.Time               `json:"reactivado_en,omitempty"`
	ActualizadoEn          *time.Time               `json:"actualizado_en,omitempty"`
	EditadoPorUsuarioID    *uuid.UUID               `gorm:"type:uuid" json:"editado_por_usuario_id,omitempty"`
	Detalles               []CompraProveedorDetalle `gorm:"foreignKey:CompraID" json:"detalles,omitempty"`
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
	// ImpuestosPorRenglon marca las compras con IEPS e IVA propios por producto.
	// Las compras anteriores quedan en false y siguen con el prorrateo de la compra.
	ImpuestosPorRenglon bool    `gorm:"not null;default:false" json:"impuestos_por_renglon"`
	TasaIEPS            float64 `gorm:"type:numeric(8,4);not null;default:0" json:"tasa_ieps"`
	IEPS                float64 `gorm:"type:numeric(14,2);not null;default:0" json:"ieps"`
	TasaIVA             float64 `gorm:"type:numeric(8,4);not null;default:0" json:"tasa_iva"`
	IVA                 float64 `gorm:"type:numeric(14,2);not null;default:0" json:"iva"`
	// Cancelado marca el renglón quitado al editar; se conserva para el historial.
	Cancelado bool `gorm:"not null;default:false" json:"cancelado"`
	// LoteID es el lote que creó el renglón al recibirse; nulo si no llevó lote o es anterior.
	LoteID   *uuid.UUID `gorm:"type:uuid" json:"lote_id,omitempty"`
	CreadoEn time.Time  `json:"creado_en"`
}

// CompraProveedorEvento es la bitácora de ediciones, suspensiones y reactivaciones.
type CompraProveedorEvento struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompraID  uuid.UUID `gorm:"type:uuid;not null;index" json:"compra_id"`
	UsuarioID uuid.UUID `gorm:"type:uuid;not null" json:"usuario_id"`
	Tipo      string    `gorm:"type:varchar(20);not null" json:"tipo"`
	Motivo    *string   `gorm:"type:text" json:"motivo,omitempty"`
	Antes     []byte    `gorm:"type:jsonb" json:"antes,omitempty"`
	Despues   []byte    `gorm:"type:jsonb" json:"despues,omitempty"`
	CreadoEn  time.Time `gorm:"autoCreateTime" json:"creado_en"`
}

func (CompraProveedorEvento) TableName() string { return "compras_proveedor_eventos" }
func (e *CompraProveedorEvento) BeforeCreate(*gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (CompraProveedor) TableName() string        { return "compras_proveedor" }
func (CompraProveedorDetalle) TableName() string { return "compras_proveedor_detalles" }
func (c *CompraProveedor) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Estado == "" {
		c.Estado = EstadoActiva
	}
	return nil
}
func (d *CompraProveedorDetalle) BeforeCreate(*gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
