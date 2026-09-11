package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Sucursal struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Codigo        string     `gorm:"type:varchar(40);not null" json:"codigo"`
	Nombre        string     `gorm:"type:varchar(180);not null" json:"nombre"`
	Telefono      *string    `gorm:"type:varchar(30)" json:"telefono"`
	DireccionID   *uuid.UUID `gorm:"type:uuid;index" json:"direccion_id"`
	EsPrincipal   bool       `gorm:"not null;default:false" json:"es_principal"`
	Activo        bool       `gorm:"not null;default:true;index" json:"-"`
	CreadoEn      time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
	EliminadoEn   *time.Time `json:"eliminado_en"`
	Direccion     *Direccion `gorm:"foreignKey:DireccionID" json:"direccion"`
}

func (Sucursal) TableName() string { return "sucursales" }

func (s *Sucursal) BeforeCreate(*gorm.DB) error {
	setID(&s.ID)
	return nil
}

type CrearSucursalInput struct {
	Codigo      string          `json:"codigo"`
	Nombre      string          `json:"nombre"`
	Telefono    *string         `json:"telefono"`
	EsPrincipal bool            `json:"es_principal"`
	Direccion   *DireccionInput `json:"direccion"`
}

type ActualizarSucursalInput struct {
	Nombre      Optional[string]                   `json:"nombre"`
	Telefono    Optional[string]                   `json:"telefono"`
	EsPrincipal Optional[bool]                     `json:"es_principal"`
	Direccion   Optional[ActualizarDireccionInput] `json:"direccion"`
}

type SucursalResumen struct {
	ID                uuid.UUID `json:"id"`
	NegocioID         uuid.UUID `json:"negocio_id"`
	Codigo            string    `json:"codigo"`
	Nombre            string    `json:"nombre"`
	Telefono          *string   `json:"telefono"`
	DireccionResumida *string   `json:"direccion_resumida"`
	EsPrincipal       bool      `json:"es_principal"`
	Estado            string    `json:"estado"`
	TipoMiembro       string    `json:"tipo_miembro"`
	CreadoEn          time.Time `json:"creado_en"`
	ActualizadoEn     time.Time `json:"actualizado_en"`
}

type SucursalDetalle struct {
	ID            uuid.UUID  `json:"id"`
	NegocioID     uuid.UUID  `json:"negocio_id"`
	Codigo        string     `json:"codigo"`
	Nombre        string     `json:"nombre"`
	Telefono      *string    `json:"telefono"`
	EsPrincipal   bool       `json:"es_principal"`
	Estado        string     `json:"estado"`
	TipoMiembro   string     `json:"tipo_miembro"`
	CreadoEn      time.Time  `json:"creado_en"`
	ActualizadoEn time.Time  `json:"actualizado_en"`
	EliminadoEn   *time.Time `json:"eliminado_en"`
	Direccion     *Direccion `json:"direccion"`
}

type ContextoNegocioSucursal struct {
	EstadoNegocio string
	TipoMiembro   string
}
