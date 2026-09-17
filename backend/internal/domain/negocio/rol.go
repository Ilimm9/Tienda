package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Rol sigue la tabla `roles` de base.MD: pertenece a un negocio y agrupa permisos globales.
type Rol struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"negocio_id"`
	Codigo             string     `gorm:"type:varchar(60);not null" json:"codigo"`
	Nombre             string     `gorm:"type:varchar(120);not null" json:"nombre"`
	Descripcion        *string    `gorm:"type:text" json:"descripcion,omitempty"`
	EsRolSistema       bool       `gorm:"not null;default:false" json:"es_rol_sistema"`
	Activo             bool       `gorm:"not null;default:true;index" json:"activo"`
	CreadoPorUsuarioID *uuid.UUID `gorm:"type:uuid;index" json:"creado_por_usuario_id,omitempty"`
	CreadoEn           time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn      time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (Rol) TableName() string { return "roles" }

func (r *Rol) BeforeCreate(*gorm.DB) error {
	setID(&r.ID)
	return nil
}

// CodigoRolPropietario identifica al rol de sistema que conserva todos los permisos del negocio.
const CodigoRolPropietario = "PROPIETARIO"

type PermisoRol struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RolID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_permisos_rol_unico" json:"rol_id"`
	PermisoID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_permisos_rol_unico" json:"permiso_id"`
	CreadoEn  time.Time `gorm:"autoCreateTime" json:"creado_en"`
}

func (PermisoRol) TableName() string { return "permisos_rol" }

func (p *PermisoRol) BeforeCreate(*gorm.DB) error {
	setID(&p.ID)
	return nil
}

type RolMembresia struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	MembresiaNegocioID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_roles_membresia_unico" json:"membresia_negocio_id"`
	RolID               uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_roles_membresia_unico" json:"rol_id"`
	AsignadoPorUsuarioID *uuid.UUID `gorm:"type:uuid" json:"asignado_por_usuario_id,omitempty"`
	AsignadoEn          time.Time  `gorm:"autoCreateTime" json:"asignado_en"`
}

func (RolMembresia) TableName() string { return "roles_membresia" }

func (r *RolMembresia) BeforeCreate(*gorm.DB) error {
	setID(&r.ID)
	return nil
}

type CrearRolInput struct {
	Codigo      string      `json:"codigo"`
	Nombre      string      `json:"nombre"`
	Descripcion *string     `json:"descripcion"`
	Permisos    []uuid.UUID `json:"permisos"`
}

type ActualizarRolInput struct {
	Nombre      Optional[string]      `json:"nombre"`
	Descripcion Optional[string]      `json:"descripcion"`
	Activo      Optional[bool]        `json:"activo"`
	Permisos    Optional[[]uuid.UUID] `json:"permisos"`
}

type RolResumen struct {
	ID            uuid.UUID `json:"id"`
	NegocioID     uuid.UUID `json:"negocio_id"`
	Codigo        string    `json:"codigo"`
	Nombre        string    `json:"nombre"`
	Descripcion   *string   `json:"descripcion"`
	EsRolSistema  bool      `json:"es_rol_sistema"`
	Activo        bool      `json:"activo"`
	TotalPermisos int       `json:"total_permisos"`
	TotalMiembros int       `json:"total_miembros"`
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}

type RolDetalle struct {
	ID            uuid.UUID   `json:"id"`
	NegocioID     uuid.UUID   `json:"negocio_id"`
	Codigo        string      `json:"codigo"`
	Nombre        string      `json:"nombre"`
	Descripcion   *string     `json:"descripcion"`
	EsRolSistema  bool        `json:"es_rol_sistema"`
	Activo        bool        `json:"activo"`
	Permisos      []uuid.UUID `json:"permisos"`
	TotalMiembros int         `json:"total_miembros"`
	CreadoEn      time.Time   `json:"creado_en"`
	ActualizadoEn time.Time   `json:"actualizado_en"`
}

// MiembroRoles describe los roles asignados a una membresía del negocio.
type MiembroRoles struct {
	MembresiaID uuid.UUID   `json:"membresia_id"`
	UsuarioID   uuid.UUID   `json:"usuario_id"`
	Correo      string      `json:"correo"`
	TipoMiembro string      `json:"tipo_miembro"`
	Estado      string      `json:"estado"`
	Roles       []uuid.UUID `json:"roles"`
}
