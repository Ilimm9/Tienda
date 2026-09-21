package cuenta

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const PropositoVerificacionCorreo = "verificacion_correo"

type DesafioAutenticacion struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"usuario_id"`
	Proposito        string     `gorm:"type:varchar(50);not null;index" json:"proposito"`
	HashOTP          string     `gorm:"type:char(64);not null" json:"-"`
	IntentosFallidos int        `gorm:"not null;default:0" json:"-"`
	DireccionIP      string     `gorm:"type:varchar(45);not null;index" json:"-"`
	UltimoEnvioEn    time.Time  `gorm:"not null;index" json:"-"`
	ExpiraEn         time.Time  `gorm:"not null;index" json:"expira_en"`
	UsadoEn          *time.Time `json:"usado_en,omitempty"`
	CreadoEn         time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	Usuario          Usuario    `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (d *DesafioAutenticacion) BeforeCreate(*gorm.DB) error {
	setID(&d.ID)
	return nil
}
