package cuenta

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Usuario struct {
	ID                           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Correo                       string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"correo"`
	HashContrasena               string     `gorm:"not null" json:"-"`
	Estado                       string     `gorm:"type:varchar(30);not null;default:'activo'" json:"estado"`
	CorreoVerificadoEn           *time.Time `json:"correo_verificado_en,omitempty"`
	IntentosInicioSesionFallidos int        `gorm:"not null;default:0" json:"-"`
	BloqueadoHasta               *time.Time `json:"-"`
	UltimoInicioSesionEn         *time.Time `json:"ultimo_inicio_sesion_en,omitempty"`
	ContrasenaCambiadaEn         *time.Time `json:"contrasena_cambiada_en,omitempty"`
	CreadoEn                     time.Time  `json:"creado_en"`
	ActualizadoEn                time.Time  `json:"actualizado_en"`
	DeshabilitadoEn              *time.Time `json:"deshabilitado_en,omitempty"`
}

func (u *Usuario) BeforeCreate(*gorm.DB) error {
	setID(&u.ID)
	return nil
}

func setID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
