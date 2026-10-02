package negocio

import (
	"strings"
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

// DuracionInvitacion es la vigencia del enlace, contada desde cada emisión o reenvío.
const DuracionInvitacion = 72 * time.Hour

// InvitacionNegocio sigue la tabla `invitaciones_negocio` de base.MD.
//
// Solo se almacena el hash del token: el valor en claro existe una vez, dentro del enlace copiable.
// `SucursalID` es nulo únicamente en invitaciones históricas previas a la asignación por sucursal.
type InvitacionNegocio struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	EmpleadoID           *uuid.UUID `gorm:"type:uuid;index" json:"empleado_id,omitempty"`
	SucursalID           *uuid.UUID `gorm:"type:uuid" json:"sucursal_id,omitempty"`
	Correo               string     `gorm:"type:varchar(254);not null;index" json:"correo"`
	RolPredeterminadoID  *uuid.UUID `gorm:"type:uuid" json:"rol_predeterminado_id,omitempty"`
	HashToken            string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	Estado               string     `gorm:"type:varchar(30);not null;default:'pendiente';index" json:"estado"`
	InvitadoPorUsuarioID uuid.UUID  `gorm:"type:uuid;not null" json:"invitado_por_usuario_id"`
	ExpiraEn             time.Time  `gorm:"not null;index" json:"expira_en"`
	AceptadoPorUsuarioID *uuid.UUID `gorm:"type:uuid" json:"aceptado_por_usuario_id,omitempty"`
	AceptadoEn           *time.Time `json:"aceptado_en,omitempty"`
	CreadoEn             time.Time  `gorm:"autoCreateTime" json:"creado_en"`
}

func (InvitacionNegocio) TableName() string { return "invitaciones_negocio" }

func (i *InvitacionNegocio) BeforeCreate(*gorm.DB) error {
	setID(&i.ID)
	return nil
}

type CrearInvitacionInput struct {
	EmpleadoID          uuid.UUID `json:"empleado_id"`
	SucursalID          uuid.UUID `json:"sucursal_id"`
	RolPredeterminadoID uuid.UUID `json:"rol_predeterminado_id"`
}

// ReenviarInvitacionInput conserva destinatario cuando `Correo` es nulo y lo corrige cuando llega.
type ReenviarInvitacionInput struct {
	Correo *string `json:"correo"`
}

// RegistroInvitacionInput no incluye correo: el servidor lo toma de la invitación.
type RegistroInvitacionInput struct {
	Nombres    string `json:"nombres"`
	Apellidos  string `json:"apellidos"`
	Telefono   string `json:"telefono"`
	Contrasena string `json:"contrasena"`
}

type FiltroInvitaciones struct {
	Estado     string
	SucursalID *uuid.UUID
	// SinAceptar agrupa pendientes y expiradas; las canceladas quedan en el historial.
	SinAceptar bool
}

type InvitacionResumen struct {
	ID                  uuid.UUID  `json:"id"`
	NegocioID           uuid.UUID  `json:"negocio_id"`
	EmpleadoID          *uuid.UUID `json:"empleado_id"`
	NombreEmpleado      string     `json:"nombre_empleado"`
	Correo              string     `json:"correo"`
	SucursalID          *uuid.UUID `json:"sucursal_id"`
	NombreSucursal      string     `json:"nombre_sucursal"`
	RolPredeterminadoID *uuid.UUID `json:"rol_predeterminado_id"`
	NombreRol           string     `json:"nombre_rol"`
	Estado              string     `json:"estado"`
	ExpiraEn            time.Time  `json:"expira_en"`
	AceptadoEn          *time.Time `json:"aceptado_en"`
	CreadoEn            time.Time  `json:"creado_en"`
}

// InvitacionCreada acompaña al resumen con el token en claro, que no vuelve a estar disponible.
type InvitacionCreada struct {
	Invitacion    InvitacionResumen `json:"invitacion"`
	Token         string            `json:"token"`
	CorreoEnviado bool              `json:"correo_enviado"`
}

// InvitacionPublica es lo que ve quien abre el enlace: sin correo completo ni nombre del empleado.
type InvitacionPublica struct {
	Correo            string    `json:"-"`
	NombreEmpleado    string    `json:"-"`
	CorreoEnmascarado string    `json:"correo_enmascarado"`
	NombreNegocio     string    `json:"nombre_negocio"`
	NombreSucursal    string    `json:"nombre_sucursal"`
	NombreRol         string    `json:"nombre_rol"`
	ExpiraEn          time.Time `json:"expira_en"`
	RequiereCuenta    bool      `json:"requiere_cuenta"`
}

type InvitacionAceptada struct {
	Aceptada   bool       `json:"aceptada"`
	NegocioID  uuid.UUID  `json:"negocio_id"`
	SucursalID *uuid.UUID `json:"sucursal_id"`
}

// EnmascararCorreo deja visibles primera y última letra del usuario y el dominio completo.
func EnmascararCorreo(correo string) string {
	partes := strings.SplitN(correo, "@", 2)
	if len(partes) != 2 || partes[0] == "" {
		return "•••@•••"
	}
	usuario := []rune(partes[0])
	if len(usuario) < 3 {
		return string(usuario[:1]) + "•••@" + partes[1]
	}
	return string(usuario[:1]) + "•••" + string(usuario[len(usuario)-1:]) + "@" + partes[1]
}
