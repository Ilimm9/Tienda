package negocio

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permiso sigue la tabla `permisos` de base.MD: catálogo global, no pertenece a ningún negocio.
type Permiso struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Codigo       string    `gorm:"type:varchar(120);not null;uniqueIndex" json:"codigo"`
	CodigoModulo string    `gorm:"type:varchar(30);not null;index" json:"codigo_modulo"`
	Nombre       string    `gorm:"type:varchar(160);not null" json:"nombre"`
	Descripcion  *string   `gorm:"type:text" json:"descripcion,omitempty"`
	CreadoEn     time.Time `gorm:"autoCreateTime" json:"creado_en"`
}

func (Permiso) TableName() string { return "permisos" }

func (p *Permiso) BeforeCreate(*gorm.DB) error {
	setID(&p.ID)
	return nil
}

// Códigos de permiso del alcance de esta capability.
const (
	PermisoNegocioVer        = "negocios.ver"
	PermisoNegocioEditar     = "negocios.editar"
	PermisoNegocioArchivar   = "negocios.archivar"
	PermisoSucursalVer       = "sucursales.ver"
	PermisoSucursalCrear     = "sucursales.crear"
	PermisoSucursalEditar    = "sucursales.editar"
	PermisoSucursalArchivar  = "sucursales.archivar"
	PermisoRolVer            = "roles.ver"
	PermisoRolGestionar      = "roles.gestionar"
	PermisoRolAsignar        = "roles.asignar"
	PermisoEmpleadoVer       = "equipo.empleados.ver"
	PermisoEmpleadoGestionar = "equipo.empleados.gestionar"
	PermisoInvitacionVer     = "equipo.invitaciones.ver"
	PermisoInvitacionEnviar  = "equipo.invitaciones.enviar"
	PermisoAsignacionVer     = "equipo.asignaciones.ver"
	PermisoAsignacionEditar  = "equipo.asignaciones.editar"
	PermisoCatalogoVer       = "catalogo.ver"
	PermisoCatalogoGestionar = "catalogo.gestionar"
)

// CatalogoPermisos es la fuente del sembrado idempotente del catálogo global.
var CatalogoPermisos = []Permiso{
	{Codigo: PermisoNegocioVer, CodigoModulo: "negocios", Nombre: "Ver negocios"},
	{Codigo: PermisoNegocioEditar, CodigoModulo: "negocios", Nombre: "Editar negocio"},
	{Codigo: PermisoNegocioArchivar, CodigoModulo: "negocios", Nombre: "Archivar o restaurar negocio"},
	{Codigo: PermisoSucursalVer, CodigoModulo: "sucursales", Nombre: "Ver sucursales"},
	{Codigo: PermisoSucursalCrear, CodigoModulo: "sucursales", Nombre: "Registrar sucursales"},
	{Codigo: PermisoSucursalEditar, CodigoModulo: "sucursales", Nombre: "Editar sucursales"},
	{Codigo: PermisoSucursalArchivar, CodigoModulo: "sucursales", Nombre: "Archivar o restaurar sucursales"},
	{Codigo: PermisoRolVer, CodigoModulo: "roles", Nombre: "Ver roles y permisos"},
	{Codigo: PermisoRolGestionar, CodigoModulo: "roles", Nombre: "Crear y editar roles"},
	{Codigo: PermisoRolAsignar, CodigoModulo: "roles", Nombre: "Asignar roles a miembros"},
	{Codigo: PermisoEmpleadoVer, CodigoModulo: "equipo", Nombre: "Ver empleados"},
	{Codigo: PermisoEmpleadoGestionar, CodigoModulo: "equipo", Nombre: "Registrar y editar empleados"},
	{Codigo: PermisoInvitacionVer, CodigoModulo: "equipo", Nombre: "Ver invitaciones"},
	{Codigo: PermisoInvitacionEnviar, CodigoModulo: "equipo", Nombre: "Invitar y cancelar invitaciones"},
	{Codigo: PermisoAsignacionVer, CodigoModulo: "equipo", Nombre: "Ver asignaciones de sucursal"},
	{Codigo: PermisoAsignacionEditar, CodigoModulo: "equipo", Nombre: "Asignar empleados a sucursales"},
	{Codigo: PermisoCatalogoVer, CodigoModulo: "catalogo", Nombre: "Ver catálogo"},
	{Codigo: PermisoCatalogoGestionar, CodigoModulo: "catalogo", Nombre: "Gestionar catálogo"},
}
