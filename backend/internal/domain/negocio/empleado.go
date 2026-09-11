package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Empleado struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PerfilID      uuid.UUID `gorm:"type:uuid;not null" json:"perfil_id"`
	Numero        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"numero"`
	Puesto        string    `gorm:"type:varchar(120)" json:"puesto"`
	Estado        string    `gorm:"type:varchar(30);not null;default:'activo'" json:"estado"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

func (e *Empleado) BeforeCreate(*gorm.DB) error {
	setID(&e.ID)
	return nil
}
