package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MembresiaNegocio struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_negocio_usuario" json:"negocio_id"`
	UsuarioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_negocio_usuario" json:"usuario_id"`
	EmpleadoID    *uuid.UUID `gorm:"type:uuid" json:"empleado_id,omitempty"`
	RolID         *uuid.UUID `gorm:"type:uuid" json:"rol_id,omitempty"`
	TipoMiembro   string     `gorm:"type:varchar(30);not null;default:'miembro';index" json:"tipo_miembro"`
	Estado        string     `gorm:"type:varchar(30);not null;default:'activo';index" json:"estado"`
	SeUnioEn      *time.Time `json:"se_unio_en,omitempty"`
	SuspendidoEn  *time.Time `json:"suspendido_en,omitempty"`
	RevocadoEn    *time.Time `json:"revocado_en,omitempty"`
	CreadoEn      time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (MembresiaNegocio) TableName() string { return "membresias_negocio" }

func (m *MembresiaNegocio) BeforeCreate(*gorm.DB) error {
	setID(&m.ID)
	return nil
}
