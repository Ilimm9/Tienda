package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PuestoNegocio es el catálogo de puestos del negocio; lo comparten todas sus sucursales.
// El empleado conserva el nombre del puesto como texto; el catálogo solo alimenta la elección.
type PuestoNegocio struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Nombre        string     `gorm:"type:varchar(120);not null" json:"nombre"`
	RolSugeridoID *uuid.UUID `gorm:"type:uuid" json:"rol_sugerido_id"`
	Activo        bool       `gorm:"not null;default:true" json:"activo"`
	CreadoEn      time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (PuestoNegocio) TableName() string { return "puestos_negocio" }

func (p *PuestoNegocio) BeforeCreate(*gorm.DB) error {
	setID(&p.ID)
	return nil
}

type CrearPuestoInput struct {
	Nombre        string     `json:"nombre"`
	RolSugeridoID *uuid.UUID `json:"rol_sugerido_id"`
}

type ActualizarPuestoInput struct {
	Nombre        Optional[string]    `json:"nombre"`
	RolSugeridoID Optional[uuid.UUID] `json:"rol_sugerido_id"`
	Activo        *bool               `json:"activo"`
}
