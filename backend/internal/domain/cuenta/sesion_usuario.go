package cuenta

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SesionUsuario conserva únicamente hashes de credenciales; los tokens en claro
// se entregan una vez al navegador y nunca se persisten.
type SesionUsuario struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"usuario_id"`
	HashToken         string     `gorm:"type:char(64);not null;uniqueIndex" json:"-"`
	HashCSRF          string     `gorm:"type:char(64);not null" json:"-"`
	DireccionIP       string     `gorm:"type:varchar(45);not null" json:"-"`
	AgenteUsuario     string     `gorm:"type:varchar(512);not null" json:"-"`
	Recordarme        bool       `gorm:"not null;default:false" json:"recordarme"`
	EmitidoEn         time.Time  `gorm:"not null" json:"emitido_en"`
	ExpiraEn          time.Time  `gorm:"not null;index" json:"expira_en"`
	UltimaActividadEn time.Time  `gorm:"not null" json:"ultima_actividad_en"`
	RevocadoEn        *time.Time `json:"revocado_en,omitempty"`
	MotivoRevocacion  *string    `gorm:"type:varchar(120)" json:"-"`
	Usuario           Usuario    `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (s *SesionUsuario) BeforeCreate(*gorm.DB) error {
	setID(&s.ID)
	return nil
}
