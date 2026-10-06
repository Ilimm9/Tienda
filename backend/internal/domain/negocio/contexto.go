package negocio

import "github.com/google/uuid"

type ContextoSucursal struct {
	ID          uuid.UUID `json:"id"`
	Codigo      string    `json:"codigo"`
	Nombre      string    `json:"nombre"`
	EsPrincipal bool      `json:"es_principal"`
}

type ContextoNegocio struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	NombreComercial string    `json:"nombre_comercial"`
	TipoMiembro     string    `json:"tipo_miembro"`
	// Permisos son los códigos efectivos de la membresía; la interfaz los usa para ocultar, la API sigue autorizando.
	Permisos   []string           `json:"permisos"`
	Sucursales []ContextoSucursal `json:"sucursales"`
}
