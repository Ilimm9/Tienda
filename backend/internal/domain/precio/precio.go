package precio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EstadoPendiente = "PENDIENTE"
const EstadoAutorizada = "AUTORIZADA"

type Propuesta struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID              uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	SucursalID             uuid.UUID  `gorm:"type:uuid;not null;index" json:"sucursal_id"`
	CompraDetalleID        uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"compra_detalle_id"`
	ProductoNegocioID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"producto_negocio_id"`
	PiezasBuenas           int        `gorm:"not null;default:1" json:"piezas_buenas"`
	CostoCapturado         float64    `gorm:"type:numeric(14,4);not null" json:"costo_capturado"`
	CostoAnterior          *float64   `gorm:"type:numeric(14,4)" json:"costo_anterior,omitempty"`
	PrecioAnterior         float64    `gorm:"type:numeric(14,2);not null" json:"precio_anterior"`
	MargenSugerido         float64    `gorm:"type:numeric(8,4);not null" json:"margen_sugerido"`
	PrecioSugerido         float64    `gorm:"type:numeric(14,2);not null" json:"precio_sugerido"`
	Estado                 string     `gorm:"type:varchar(20);not null;index" json:"estado"`
	CostoAutorizado        *float64   `gorm:"type:numeric(14,4)" json:"costo_autorizado,omitempty"`
	PrecioAutorizado       *float64   `gorm:"type:numeric(14,2)" json:"precio_autorizado,omitempty"`
	MargenAutorizado       *float64   `gorm:"type:numeric(8,4)" json:"margen_autorizado,omitempty"`
	AutorizadoPorUsuarioID *uuid.UUID `gorm:"type:uuid;index" json:"autorizado_por_usuario_id,omitempty"`
	AutorizadoEn           *time.Time `json:"autorizado_en,omitempty"`
	CreadoEn               time.Time  `gorm:"autoCreateTime" json:"creado_en"`
}

func (Propuesta) TableName() string { return "propuestas_costo_precio" }
func (p *Propuesta) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type PrecioSucursal struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID         uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	SucursalID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_precio_sucursal_producto" json:"sucursal_id"`
	ProductoNegocioID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_precio_sucursal_producto" json:"producto_negocio_id"`
	CostoVigente      float64   `gorm:"type:numeric(14,4);not null;default:0" json:"costo_vigente"`
	PrecioVenta       float64   `gorm:"type:numeric(14,2);not null;default:0" json:"precio_venta"`
	Margen            *float64  `gorm:"type:numeric(8,4)" json:"margen,omitempty"`
	ActualizadoEn     time.Time `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (PrecioSucursal) TableName() string { return "precios_sucursal" }
func (p *PrecioSucursal) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ConfiguracionSucursal struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SucursalID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"sucursal_id"`
	MargenPredeterminado float64   `gorm:"type:numeric(8,4);not null;default:0.30" json:"margen_predeterminado"`
}

func (ConfiguracionSucursal) TableName() string { return "configuracion_precios_sucursal" }
func (p *ConfiguracionSucursal) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
