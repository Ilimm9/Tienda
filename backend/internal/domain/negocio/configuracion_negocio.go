package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfiguracionNegocio struct {
	ID                                        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID                                 uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"negocio_id"`
	MetodoCosteo                              string     `gorm:"type:varchar(40);not null;default:'promedio_ponderado'" json:"metodo_costeo"`
	PermitirInventarioNegativo                bool       `gorm:"not null;default:false" json:"permitir_inventario_negativo"`
	RequerirAutorizacionCambioPrecio          bool       `gorm:"not null;default:true" json:"requerir_autorizacion_cambio_precio"`
	RequerirAutorizacionEliminarPartidaVenta  bool       `gorm:"not null;default:false" json:"requerir_autorizacion_eliminar_partida_venta"`
	PermitirMultiplesCajasAbiertasPorEmpleado bool       `gorm:"not null;default:false" json:"permitir_multiples_cajas_abiertas_por_empleado"`
	PermitirCambioCajaConAutorizacion         bool       `gorm:"not null;default:true" json:"permitir_cambio_caja_con_autorizacion"`
	HabilitarCortesParcialesCaja              bool       `gorm:"not null;default:false" json:"habilitar_cortes_parciales_caja"`
	MontoCorteParcialCaja                     *float64   `gorm:"type:numeric(18,2)" json:"monto_corte_parcial_caja,omitempty"`
	HoraCorteParcialCaja                      *time.Time `gorm:"type:time" json:"hora_corte_parcial_caja,omitempty"`
	PreciosMayoreoHabilitados                 bool       `gorm:"not null;default:true" json:"precios_mayoreo_habilitados"`
	DiasDevolucionEnvasePredeterminados       *int       `json:"dias_devolucion_envase_predeterminados,omitempty"`
	CreadoEn                                  time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn                             time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (ConfiguracionNegocio) TableName() string { return "configuraciones_negocio" }

func (c *ConfiguracionNegocio) BeforeCreate(*gorm.DB) error {
	setID(&c.ID)
	return nil
}
