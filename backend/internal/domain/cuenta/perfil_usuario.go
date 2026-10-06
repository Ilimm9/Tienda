package cuenta

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PerfilUsuario struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"usuario_id"`
	Nombres         string    `gorm:"type:varchar(120);not null" json:"nombres"`
	PrimerApellido  string    `gorm:"type:varchar(120);not null" json:"primer_apellido"`
	SegundoApellido *string   `gorm:"type:varchar(120)" json:"segundo_apellido,omitempty"`
	Telefono        *string   `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	CreadoEn        time.Time `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn   time.Time `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (p *PerfilUsuario) BeforeCreate(*gorm.DB) error {
	setID(&p.ID)
	return nil
}
