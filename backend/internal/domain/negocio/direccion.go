package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Direccion struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CodigoPais     string    `gorm:"type:varchar(2);not null;default:'MX'" json:"codigo_pais"`
	Estado         *string   `gorm:"type:varchar(120)" json:"estado,omitempty"`
	Municipio      *string   `gorm:"type:varchar(120)" json:"municipio,omitempty"`
	Ciudad         *string   `gorm:"type:varchar(120)" json:"ciudad,omitempty"`
	Colonia        *string   `gorm:"type:varchar(150)" json:"colonia,omitempty"`
	CodigoPostal   *string   `gorm:"type:varchar(12)" json:"codigo_postal,omitempty"`
	Calle          *string   `gorm:"type:varchar(180)" json:"calle,omitempty"`
	NumeroExterior *string   `gorm:"type:varchar(30)" json:"numero_exterior,omitempty"`
	NumeroInterior *string   `gorm:"type:varchar(30)" json:"numero_interior,omitempty"`
	Referencias    *string   `gorm:"type:text" json:"referencias,omitempty"`
	CreadoEn       time.Time `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn  time.Time `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (Direccion) TableName() string { return "direcciones" }

func (d *Direccion) BeforeCreate(*gorm.DB) error {
	setID(&d.ID)
	return nil
}

type DireccionInput struct {
	CodigoPais     string  `json:"codigo_pais"`
	Estado         *string `json:"estado"`
	Municipio      *string `json:"municipio"`
	Ciudad         *string `json:"ciudad"`
	Colonia        *string `json:"colonia"`
	CodigoPostal   *string `json:"codigo_postal"`
	Calle          *string `json:"calle"`
	NumeroExterior *string `json:"numero_exterior"`
	NumeroInterior *string `json:"numero_interior"`
	Referencias    *string `json:"referencias"`
}

type ActualizarDireccionInput struct {
	CodigoPais     Optional[string] `json:"codigo_pais"`
	Estado         Optional[string] `json:"estado"`
	Municipio      Optional[string] `json:"municipio"`
	Ciudad         Optional[string] `json:"ciudad"`
	Colonia        Optional[string] `json:"colonia"`
	CodigoPostal   Optional[string] `json:"codigo_postal"`
	Calle          Optional[string] `json:"calle"`
	NumeroExterior Optional[string] `json:"numero_exterior"`
	NumeroInterior Optional[string] `json:"numero_interior"`
	Referencias    Optional[string] `json:"referencias"`
}
