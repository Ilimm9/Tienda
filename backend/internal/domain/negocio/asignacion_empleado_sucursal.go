package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AsignacionEmpleadoSucursal sigue la tabla `asignaciones_empleado_sucursal` de base.MD.
type AsignacionEmpleadoSucursal struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	EmpleadoID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_asignacion_empleado_sucursal" json:"empleado_id"`
	SucursalID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_asignacion_empleado_sucursal" json:"sucursal_id"`
	EsPrincipal bool       `gorm:"not null;default:false" json:"es_principal"`
	Activo      bool       `gorm:"not null;default:true;index" json:"activo"`
	AsignadoEn  time.Time  `gorm:"autoCreateTime" json:"asignado_en"`
	FinalizadoEn *time.Time `json:"finalizado_en,omitempty"`
}

func (AsignacionEmpleadoSucursal) TableName() string { return "asignaciones_empleado_sucursal" }

func (a *AsignacionEmpleadoSucursal) BeforeCreate(*gorm.DB) error {
	setID(&a.ID)
	return nil
}

type AsignacionResumen struct {
	ID             uuid.UUID  `json:"id"`
	NegocioID      uuid.UUID  `json:"negocio_id"`
	EmpleadoID     uuid.UUID  `json:"empleado_id"`
	SucursalID     uuid.UUID  `json:"sucursal_id"`
	CodigoSucursal string     `json:"codigo_sucursal"`
	NombreSucursal string     `json:"nombre_sucursal"`
	EsPrincipal    bool       `json:"es_principal"`
	Activo         bool       `json:"activo"`
	AsignadoEn     time.Time  `json:"asignado_en"`
	FinalizadoEn   *time.Time `json:"finalizado_en"`
}

type AsignarSucursalInput struct {
	SucursalID  uuid.UUID `json:"sucursal_id"`
	EsPrincipal bool      `json:"es_principal"`
}
