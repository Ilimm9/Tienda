package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Estados de empleado. base.MD los declara como enum `estado_empleado`; el proyecto los
// implementa como varchar validado en service, igual que el resto de estados desde fase 1.
const (
	EstadoEmpleadoPendiente  = "pendiente"
	EstadoEmpleadoActivo     = "activo"
	EstadoEmpleadoSuspendido = "suspendido"
	EstadoEmpleadoTerminado  = "terminado"
)

// Empleado sigue la tabla `empleados` de base.MD: pertenece al negocio y existe sin cuenta.
type Empleado struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	MembresiaID        *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"membresia_id,omitempty"`
	NumeroEmpleado     *string    `gorm:"type:varchar(40)" json:"numero_empleado"`
	Nombre             string     `gorm:"type:varchar(100);not null" json:"nombre"`
	SegundoNombre      *string    `gorm:"type:varchar(100)" json:"segundo_nombre"`
	PrimerApellido     string     `gorm:"type:varchar(100);not null" json:"primer_apellido"`
	SegundoApellido    *string    `gorm:"type:varchar(100)" json:"segundo_apellido"`
	Correo             *string    `gorm:"type:varchar(254)" json:"correo"`
	Telefono           *string    `gorm:"type:varchar(30)" json:"telefono"`
	Puesto             *string    `gorm:"type:varchar(120)" json:"puesto"`
	Estado             string     `gorm:"type:varchar(30);not null;default:'pendiente';index" json:"estado"`
	ContratadoEn       *time.Time `gorm:"type:date" json:"contratado_en"`
	TerminadoEn        *time.Time `gorm:"type:date" json:"terminado_en"`
	CreadoPorUsuarioID uuid.UUID  `gorm:"type:uuid;not null;index" json:"creado_por_usuario_id"`
	CreadoEn           time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn      time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (Empleado) TableName() string { return "empleados" }

func (e *Empleado) BeforeCreate(*gorm.DB) error {
	setID(&e.ID)
	return nil
}

type CrearEmpleadoInput struct {
	NumeroEmpleado  *string `json:"numero_empleado"`
	Nombre          string  `json:"nombre"`
	SegundoNombre   *string `json:"segundo_nombre"`
	PrimerApellido  string  `json:"primer_apellido"`
	SegundoApellido *string `json:"segundo_apellido"`
	Correo          *string `json:"correo"`
	Telefono        *string `json:"telefono"`
	Puesto          *string `json:"puesto"`
	ContratadoEn    *string `json:"contratado_en"`
}

type ActualizarEmpleadoInput struct {
	NumeroEmpleado  Optional[string] `json:"numero_empleado"`
	Nombre          Optional[string] `json:"nombre"`
	SegundoNombre   Optional[string] `json:"segundo_nombre"`
	PrimerApellido  Optional[string] `json:"primer_apellido"`
	SegundoApellido Optional[string] `json:"segundo_apellido"`
	Correo          Optional[string] `json:"correo"`
	Telefono        Optional[string] `json:"telefono"`
	Puesto          Optional[string] `json:"puesto"`
	Estado          Optional[string] `json:"estado"`
	ContratadoEn    Optional[string] `json:"contratado_en"`
	TerminadoEn     Optional[string] `json:"terminado_en"`
}

type EmpleadoResumen struct {
	ID             uuid.UUID  `json:"id"`
	NegocioID      uuid.UUID  `json:"negocio_id"`
	NumeroEmpleado *string    `json:"numero_empleado"`
	NombreCompleto string     `json:"nombre_completo"`
	Correo         *string    `json:"correo"`
	Telefono       *string    `json:"telefono"`
	Puesto         *string    `json:"puesto"`
	Estado         string     `json:"estado"`
	TieneCuenta    bool       `json:"tiene_cuenta"`
	CreadoEn       time.Time  `json:"creado_en"`
	ActualizadoEn  time.Time  `json:"actualizado_en"`
}

type EmpleadoDetalle struct {
	ID              uuid.UUID  `json:"id"`
	NegocioID       uuid.UUID  `json:"negocio_id"`
	MembresiaID     *uuid.UUID `json:"membresia_id"`
	NumeroEmpleado  *string    `json:"numero_empleado"`
	Nombre          string     `json:"nombre"`
	SegundoNombre   *string    `json:"segundo_nombre"`
	PrimerApellido  string     `json:"primer_apellido"`
	SegundoApellido *string    `json:"segundo_apellido"`
	NombreCompleto  string     `json:"nombre_completo"`
	Correo          *string    `json:"correo"`
	Telefono        *string    `json:"telefono"`
	Puesto          *string    `json:"puesto"`
	Estado          string     `json:"estado"`
	ContratadoEn    *time.Time `json:"contratado_en"`
	TerminadoEn     *time.Time `json:"terminado_en"`
	TieneCuenta     bool       `json:"tiene_cuenta"`
	CreadoEn        time.Time  `json:"creado_en"`
	ActualizadoEn   time.Time  `json:"actualizado_en"`
}

// NombreCompletoEmpleado arma el nombre visible sin depender de campos opcionales vacíos.
func NombreCompletoEmpleado(nombre string, segundoNombre *string, primerApellido string, segundoApellido *string) string {
	partes := []string{nombre}
	if segundoNombre != nil && *segundoNombre != "" {
		partes = append(partes, *segundoNombre)
	}
	partes = append(partes, primerApellido)
	if segundoApellido != nil && *segundoApellido != "" {
		partes = append(partes, *segundoApellido)
	}
	resultado := ""
	for indice, parte := range partes {
		if parte == "" {
			continue
		}
		if indice > 0 && resultado != "" {
			resultado += " "
		}
		resultado += parte
	}
	return resultado
}
