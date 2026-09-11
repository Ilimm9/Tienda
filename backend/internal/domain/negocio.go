package domain

import (
	"bytes"
	"encoding/json"
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

type Negocio struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Slug               string     `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	NombreComercial    string     `gorm:"type:varchar(180);not null" json:"nombre_comercial"`
	Nombre             string     `gorm:"type:varchar(180);not null" json:"-"`
	RazonSocial        *string    `gorm:"type:varchar(220)" json:"razon_social,omitempty"`
	RFC                *string    `gorm:"type:varchar(20);index" json:"rfc,omitempty"`
	Telefono           *string    `gorm:"type:varchar(30)" json:"telefono,omitempty"`
	Correo             *string    `gorm:"type:varchar(254)" json:"correo,omitempty"`
	EmailLegacy        *string    `gorm:"column:email;type:varchar(255)" json:"-"`
	DireccionID        *uuid.UUID `gorm:"type:uuid;index" json:"direccion_id,omitempty"`
	ArchivoLogoID      *uuid.UUID `gorm:"type:uuid" json:"archivo_logo_id,omitempty"`
	CodigoMoneda       string     `gorm:"type:varchar(3);not null;default:'MXN'" json:"codigo_moneda"`
	ZonaHoraria        string     `gorm:"type:varchar(80);not null;default:'America/Mexico_City'" json:"zona_horaria"`
	Estado             string     `gorm:"type:varchar(30);not null;default:'activo';index" json:"estado"`
	CreadoPorUsuarioID *uuid.UUID `gorm:"type:uuid;index" json:"creado_por_usuario_id,omitempty"`
	CreadoEn           time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn      time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
	ArchivadoEn        *time.Time `json:"archivado_en,omitempty"`
	Direccion          *Direccion `gorm:"foreignKey:DireccionID" json:"direccion,omitempty"`
}

func (Negocio) TableName() string { return "negocios" }
func (n *Negocio) BeforeCreate(*gorm.DB) error {
	setID(&n.ID)
	n.Nombre = n.NombreComercial
	n.EmailLegacy = n.Correo
	return nil
}
func (n *Negocio) BeforeUpdate(*gorm.DB) error {
	n.Nombre = n.NombreComercial
	n.EmailLegacy = n.Correo
	return nil
}

type ConfiguracionNegocio struct {
	ID                                        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID                                 uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"negocio_id"`
	MetodoCosteo                              string     `gorm:"type:varchar(40);not null;default:'promedio_ponderado'" json:"metodo_costeo"`
	PermitirInventarioNegativo                bool       `gorm:"not null;default:false" json:"permitir_inventario_negativo"`
	RequerirAutorizacionCambioPrecio          bool       `gorm:"not null;default:true" json:"requerir_autorizacion_cambio_precio"`
	RequerirAutorizacionEliminarPartidaVenta  bool       `gorm:"not null;default:false" json:"requerir_autorizacion_eliminar_partida_venta"`
	PermitirMultiplesCajasAbiertasPorEmpleado bool       `gorm:"not null;default:false" json:"permitir_multiples_cajas_abiertas_por_empleado"`
	PermitirCambioCajaConAutorizacion         bool       `gorm:"not null;default:true" json:"permitir_cambio_caja_con_autorizacion"`
	HabilitarCortesParcialesCaja              bool       `gorm:"not null;default:false" json:"habilitar_cortes_parciales_caja"`
	MontoCorteParcialCaja                     *float64   `gorm:"type:numeric(18,2)" json:"monto_corte_parcial_caja,omitempty"`
	HoraCorteParcialCaja                      *time.Time `gorm:"type:time" json:"hora_corte_parcial_caja,omitempty"`
	PreciosMayoreoHabilitados                 bool       `gorm:"not null;default:true" json:"precios_mayoreo_habilitados"`
	DiasDevolucionEnvasePredeterminados       *int       `json:"dias_devolucion_envase_predeterminados,omitempty"`
	CreadoEn                                  time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn                             time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (ConfiguracionNegocio) TableName() string { return "configuraciones_negocio" }
func (c *ConfiguracionNegocio) BeforeCreate(*gorm.DB) error {
	setID(&c.ID)
	return nil
}

type MembresiaNegocio struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NegocioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_negocio_usuario" json:"negocio_id"`
	UsuarioID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_membresia_negocio_usuario" json:"usuario_id"`
	EmpleadoID    *uuid.UUID `gorm:"type:uuid" json:"empleado_id,omitempty"`
	RolID         *uuid.UUID `gorm:"type:uuid" json:"rol_id,omitempty"`
	TipoMiembro   string     `gorm:"type:varchar(30);not null;default:'miembro';index" json:"tipo_miembro"`
	Estado        string     `gorm:"type:varchar(30);not null;default:'activo';index" json:"estado"`
	SeUnioEn      *time.Time `json:"se_unio_en,omitempty"`
	SuspendidoEn  *time.Time `json:"suspendido_en,omitempty"`
	RevocadoEn    *time.Time `json:"revocado_en,omitempty"`
	CreadoEn      time.Time  `gorm:"autoCreateTime" json:"creado_en"`
	ActualizadoEn time.Time  `gorm:"autoUpdateTime" json:"actualizado_en"`
}

func (MembresiaNegocio) TableName() string { return "membresias_negocio" }
func (m *MembresiaNegocio) BeforeCreate(*gorm.DB) error {
	setID(&m.ID)
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

type CrearNegocioInput struct {
	NombreComercial string          `json:"nombre_comercial" binding:"required"`
	RazonSocial     *string         `json:"razon_social"`
	RFC             *string         `json:"rfc"`
	Telefono        *string         `json:"telefono"`
	Correo          *string         `json:"correo"`
	CodigoMoneda    string          `json:"codigo_moneda"`
	ZonaHoraria     string          `json:"zona_horaria"`
	Direccion       *DireccionInput `json:"direccion"`
}

// Optional preserves the difference between an omitted JSON field and explicit null.
type Optional[T any] struct {
	Set   bool
	Value *T
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
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

type ActualizarNegocioInput struct {
	NombreComercial Optional[string]                   `json:"nombre_comercial"`
	RazonSocial     Optional[string]                   `json:"razon_social"`
	RFC             Optional[string]                   `json:"rfc"`
	Telefono        Optional[string]                   `json:"telefono"`
	Correo          Optional[string]                   `json:"correo"`
	CodigoMoneda    Optional[string]                   `json:"codigo_moneda"`
	ZonaHoraria     Optional[string]                   `json:"zona_horaria"`
	Direccion       Optional[ActualizarDireccionInput] `json:"direccion"`
}

type NegocioResumen struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	NombreComercial string    `json:"nombre_comercial"`
	RFC             *string   `json:"rfc"`
	TipoMiembro     string    `json:"tipo_miembro"`
	Estado          string    `json:"estado"`
	TieneSucursales bool      `json:"tiene_sucursales"`
	TotalSucursales int64     `json:"total_sucursales"`
	CreadoEn        time.Time `json:"creado_en"`
}

type NegocioDetalle struct {
	ID              uuid.UUID  `json:"id"`
	Slug            string     `json:"slug"`
	NombreComercial string     `json:"nombre_comercial"`
	RazonSocial     *string    `json:"razon_social"`
	RFC             *string    `json:"rfc"`
	Telefono        *string    `json:"telefono"`
	Correo          *string    `json:"correo"`
	CodigoMoneda    string     `json:"codigo_moneda"`
	ZonaHoraria     string     `json:"zona_horaria"`
	Estado          string     `json:"estado"`
	TipoMiembro     string     `json:"tipo_miembro"`
	TieneSucursales bool       `json:"tiene_sucursales"`
	TotalSucursales int64      `json:"total_sucursales"`
	CreadoEn        time.Time  `json:"creado_en"`
	ActualizadoEn   time.Time  `json:"actualizado_en"`
	ArchivadoEn     *time.Time `json:"archivado_en"`
	Direccion       *Direccion `json:"direccion"`
}
