package negocio

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
