package plataforma

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const RolAdminLectura = "ADMIN_LECTURA"

type RolPlataforma struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Codigo      string    `gorm:"type:varchar(60);not null;uniqueIndex" json:"codigo"`
	Nombre      string    `gorm:"type:varchar(120);not null" json:"nombre"`
	Descripcion *string   `gorm:"type:text" json:"descripcion,omitempty"`
	Activo      bool      `gorm:"not null;default:true;index" json:"activo"`
	CreadoEn    time.Time `gorm:"autoCreateTime" json:"creado_en"`
}

func (RolPlataforma) TableName() string { return "roles_plataforma" }
func (r *RolPlataforma) BeforeCreate(*gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type RolUsuarioPlataforma struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_rol_usuario_plataforma" json:"usuario_id"`
	RolID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_rol_usuario_plataforma" json:"rol_id"`
	Origen     string    `gorm:"type:varchar(30);not null" json:"origen"`
	AsignadoEn time.Time `gorm:"autoCreateTime" json:"asignado_en"`
}

func (RolUsuarioPlataforma) TableName() string { return "roles_usuario_plataforma" }
func (r *RolUsuarioPlataforma) BeforeCreate(*gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type AuditoriaPlataforma struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UsuarioID uuid.UUID  `gorm:"type:uuid;not null;index" json:"usuario_id"`
	RolID     *uuid.UUID `gorm:"type:uuid" json:"rol_id,omitempty"`
	Accion    string     `gorm:"type:varchar(60);not null;index" json:"accion"`
	Motivo    string     `gorm:"type:varchar(500);not null" json:"motivo"`
	Origen    string     `gorm:"type:varchar(30);not null" json:"origen"`
	CreadoEn  time.Time  `gorm:"autoCreateTime;index" json:"creado_en"`
}

func (AuditoriaPlataforma) TableName() string { return "auditoria_plataforma" }
func (a *AuditoriaPlataforma) BeforeCreate(*gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type ActividadUsuarioDiaria struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Fecha              time.Time `gorm:"type:date;not null;uniqueIndex:idx_actividad_usuario_negocio_dia" json:"fecha"`
	UsuarioID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_actividad_usuario_negocio_dia;index" json:"usuario_id"`
	NegocioID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_actividad_usuario_negocio_dia;index" json:"negocio_id"`
	PrimeraActividadEn time.Time `gorm:"not null" json:"primera_actividad_en"`
	UltimaActividadEn  time.Time `gorm:"not null;index" json:"ultima_actividad_en"`
	TotalSolicitudes   int64     `gorm:"not null;default:1" json:"total_solicitudes"`
}

func (ActividadUsuarioDiaria) TableName() string { return "actividad_usuario_diaria" }
func (a *ActividadUsuarioDiaria) BeforeCreate(*gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type Resumen struct {
	NegociosTotal      int64             `json:"negocios_total"`
	NegociosActivos    int64             `json:"negocios_activos"`
	NegociosArchivados int64             `json:"negocios_archivados"`
	SucursalesActivas  int64             `json:"sucursales_activas"`
	CuentasRegistradas int64             `json:"cuentas_registradas"`
	UsuariosActivos7d  int64             `json:"usuarios_activos_7d"`
	UsuariosActivos30d int64             `json:"usuarios_activos_30d"`
	Actividad          []ActividadDiaria `json:"actividad"`
}

type ActividadDiaria struct {
	Fecha       time.Time `json:"fecha"`
	Usuarios    int64     `json:"usuarios"`
	Solicitudes int64     `json:"solicitudes"`
}

type NegocioResumen struct {
	ID                uuid.UUID  `json:"id"`
	Slug              string     `json:"slug"`
	NombreComercial   string     `json:"nombre_comercial"`
	RFC               *string    `json:"rfc"`
	Estado            string     `json:"estado"`
	CreadoEn          time.Time  `json:"creado_en"`
	TotalSucursales   int64      `json:"total_sucursales"`
	TotalPersonal     int64      `json:"total_personal"`
	TotalProductos    int64      `json:"total_productos"`
	CuentasVinculadas int64      `json:"cuentas_vinculadas"`
	UltimaActividadEn *time.Time `json:"ultima_actividad_en"`
}

type SucursalResumen struct {
	ID               uuid.UUID `json:"id"`
	Codigo           string    `json:"codigo"`
	Nombre           string    `json:"nombre"`
	EsPrincipal      bool      `json:"es_principal"`
	Estado           string    `json:"estado"`
	PersonalAsignado int64     `json:"personal_asignado"`
	ConCuentaActiva  int64     `json:"con_cuenta_activa"`
}

type NegocioDetalle struct {
	NegocioResumen
	RazonSocial       *string           `json:"razon_social"`
	Correo            *string           `json:"correo"`
	ZonaHoraria       string            `json:"zona_horaria"`
	TotalPropietarios int64             `json:"total_propietarios"`
	Sucursales        []SucursalResumen `gorm:"-" json:"sucursales"`
	Actividad         []ActividadDiaria `gorm:"-" json:"actividad"`
}

type UsuarioResumen struct {
	ID                   uuid.UUID  `json:"id"`
	Nombre               string     `json:"nombre"`
	Correo               string     `json:"correo"`
	Estado               string     `json:"estado"`
	TipoMiembro          string     `json:"tipo_miembro"`
	Roles                []string   `json:"roles"`
	Sucursales           []string   `json:"sucursales"`
	UltimoInicioSesionEn *time.Time `json:"ultimo_inicio_sesion_en"`
}

type PaginaNegocios struct {
	Items    []NegocioResumen `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type CuentaResumen struct {
	ID                   uuid.UUID  `json:"id"`
	Nombre               string     `json:"nombre"`
	Correo               string     `json:"correo"`
	Estado               string     `json:"estado"`
	CorreoVerificado     bool       `json:"correo_verificado"`
	CreadoEn             time.Time  `json:"creado_en"`
	UltimoInicioSesionEn *time.Time `json:"ultimo_inicio_sesion_en"`
	TotalNegocios        int64      `json:"total_negocios"`
	Negocios             []string   `json:"negocios"`
}

type PaginaCuentas struct {
	Items    []CuentaResumen `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}
