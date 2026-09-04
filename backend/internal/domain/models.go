package domain

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
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type PerfilUsuario struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID     uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"usuario_id"`
	Nombres       string    `gorm:"type:varchar(120);not null" json:"nombres"`
	Apellidos     string    `gorm:"type:varchar(120);not null" json:"apellidos"`
	Telefono      *string   `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Empleado struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PerfilID      uuid.UUID `gorm:"type:uuid;not null" json:"perfil_id"`
	Numero        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"numero"`
	Puesto        string    `gorm:"type:varchar(120)" json:"puesto"`
	Estado        string    `gorm:"type:varchar(30);not null;default:'activo'" json:"estado"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Negocio struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre        string    `gorm:"type:varchar(180);not null" json:"nombre"`
	RazonSocial   *string   `gorm:"type:varchar(180)" json:"razon_social,omitempty"`
	RFC           *string   `gorm:"type:varchar(30)" json:"rfc,omitempty"`
	Telefono      *string   `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	Email         *string   `gorm:"type:varchar(255)" json:"email,omitempty"`
	Estado        string    `gorm:"type:varchar(30);not null;default:'activo'" json:"estado"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Rol struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Nombre        string    `gorm:"type:varchar(100);not null" json:"nombre"`
	Descripcion   *string   `gorm:"type:text" json:"descripcion,omitempty"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type Membresia struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_usuario_negocio" json:"usuario_id"`
	NegocioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_usuario_negocio" json:"negocio_id"`
	EmpleadoID    *uuid.UUID `gorm:"type:uuid" json:"empleado_id,omitempty"`
	RolID         uuid.UUID  `gorm:"type:uuid;not null" json:"rol_id"`
	Estado        string     `gorm:"type:varchar(30);not null;default:'activo'" json:"estado"`
	CreadoEn      time.Time  `json:"creado_en"`
	ActualizadoEn time.Time  `json:"actualizado_en"`
}

func setID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
func (p *PerfilUsuario) BeforeCreate(*gorm.DB) error { setID(&p.ID); return nil }
func (e *Empleado) BeforeCreate(*gorm.DB) error      { setID(&e.ID); return nil }
func (n *Negocio) BeforeCreate(*gorm.DB) error       { setID(&n.ID); return nil }
func (r *Rol) BeforeCreate(*gorm.DB) error           { setID(&r.ID); return nil }
func (m *Membresia) BeforeCreate(*gorm.DB) error     { setID(&m.ID); return nil }
