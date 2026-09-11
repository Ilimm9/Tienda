package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rol struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Nombre        string    `gorm:"type:varchar(100);not null" json:"nombre"`
	Descripcion   *string   `gorm:"type:text" json:"descripcion,omitempty"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

func (r *Rol) BeforeCreate(*gorm.DB) error {
	setID(&r.ID)
	return nil
}
