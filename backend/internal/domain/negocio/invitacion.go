package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Estados de invitación conforme a `estado_invitacion` de base.MD.
const (
	EstadoInvitacionPendiente = "pendiente"
	EstadoInvitacionAceptada  = "aceptada"
	EstadoInvitacionExpirada  = "expirada"
	EstadoInvitacionCancelada = "cancelada"
)

// DuracionInvitacion es la vigencia del enlace copiable entregado al invitar.
const DuracionInvitacion = 7 * 24 * time.Hour

// InvitacionNegocio sigue la tabla `invitaciones_negocio` de base.MD.
//
// Solo se almacena el hash del token: el valor en claro existe una vez, dentro del enlace copiable.
type InvitacionNegocio struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	EmpleadoID          *uuid.UUID `gorm:"type:uuid;index" json:"empleado_id,omitempty"`
	Correo              string     `gorm:"type:varchar(254);not null;index" json:"correo"`
	RolPredeterminadoID *uuid.UUID `gorm:"type:uuid" json:"rol_predeterminado_id,omitempty"`
	HashToken           string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	Estado              string     `gorm:"type:varchar(30);not null;default:'pendiente';index" json:"estado"`
	InvitadoPorUsuarioID uuid.UUID `gorm:"type:uuid;not null" json:"invitado_por_usuario_id"`
	ExpiraEn            time.Time  `gorm:"not null;index" json:"expira_en"`
	AceptadoPorUsuarioID *uuid.UUID `gorm:"type:uuid" json:"aceptado_por_usuario_id,omitempty"`
	AceptadoEn          *time.Time `json:"aceptado_en,omitempty"`
	CreadoEn            time.Time  `gorm:"autoCreateTime" json:"creado_en"`
}

func (InvitacionNegocio) TableName() string { return "invitaciones_negocio" }

func (i *InvitacionNegocio) BeforeCreate(*gorm.DB) error {
	setID(&i.ID)
	return nil
}

type CrearInvitacionInput struct {
	EmpleadoID          uuid.UUID  `json:"empleado_id"`
	RolPredeterminadoID *uuid.UUID `json:"rol_predeterminado_id"`
}

type InvitacionResumen struct {
	ID              uuid.UUID  `json:"id"`
	NegocioID       uuid.UUID  `json:"negocio_id"`
	EmpleadoID      *uuid.UUID `json:"empleado_id"`
	NombreEmpleado  string     `json:"nombre_empleado"`
	Correo          string     `json:"correo"`
	RolPredeterminadoID *uuid.UUID `json:"rol_predeterminado_id"`
	Estado          string     `json:"estado"`
	ExpiraEn        time.Time  `json:"expira_en"`
	AceptadoEn      *time.Time `json:"aceptado_en"`
	CreadoEn        time.Time  `json:"creado_en"`
}

// InvitacionCreada acompaña al resumen con el token en claro, que no vuelve a estar disponible.
type InvitacionCreada struct {
	Invitacion InvitacionResumen `json:"invitacion"`
	Token      string            `json:"token"`
}

// InvitacionPublica es lo que ve quien abre el enlace, sin revelar datos internos del negocio.
type InvitacionPublica struct {
	Correo          string    `json:"correo"`
	NombreNegocio   string    `json:"nombre_negocio"`
	NombreEmpleado  string    `json:"nombre_empleado"`
	ExpiraEn        time.Time `json:"expira_en"`
	RequiereCuenta  bool      `json:"requiere_cuenta"`
}
