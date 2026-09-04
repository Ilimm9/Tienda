package domain

import "github.com/google/uuid"

type ProductRow struct {
	ID        uuid.UUID `json:"id"`
	Nombre    string    `json:"nombre"`
	ImagenURL *string   `json:"imagen_url"`
	SKU       *string   `json:"sku"`
	Precio    float64   `json:"precio"`
	Stock     float64   `json:"stock"`
	Categoria *string   `json:"categoria"`
	Estado    string    `json:"estado"`
}

type CatalogOption struct {
	ID     uuid.UUID `json:"id"`
	Nombre string    `json:"nombre"`
}

type CreateProductInput struct {
	Nombre          string     `json:"nombre" binding:"required"`
	SKUInterno      string     `json:"sku_interno" binding:"required"`
	MarcaID         *uuid.UUID `json:"marca_id"`
	CategoriaID     uuid.UUID  `json:"categoria_id" binding:"required"`
	SucursalID      uuid.UUID  `json:"sucursal_id" binding:"required"`
	Descripcion     *string    `json:"descripcion"`
	Contenido       *float64   `json:"contenido"`
	UnidadContenido *string    `json:"unidad_contenido"`
	Presentacion    *string    `json:"presentacion"`
	PrecioVenta     float64    `json:"precio_venta"`
	StockInicial    float64    `json:"stock_inicial"`
}

type CreateMarcaInput struct {
	Nombre string `json:"nombre" binding:"required"`
}
type UpdateMarcaInput struct {
	Nombre *string `json:"nombre"`
	Activo *bool   `json:"activo"`
}
type CreateCategoriaInput struct {
	Nombre           string     `json:"nombre" binding:"required"`
	CategoriaPadreID *uuid.UUID `json:"categoria_padre_id"`
	Descripcion      *string    `json:"descripcion"`
}
type UpdateCategoriaInput struct {
	Nombre           *string    `json:"nombre"`
	CategoriaPadreID *uuid.UUID `json:"categoria_padre_id"`
	Descripcion      *string    `json:"descripcion"`
	Activo           *bool      `json:"activo"`
}
type CreateProveedorInput struct {
	Nombre      string  `json:"nombre" binding:"required"`
	RazonSocial *string `json:"razon_social"`
	RFC         *string `json:"rfc"`
	Telefono    *string `json:"telefono"`
	Email       *string `json:"email"`
	Direccion   *string `json:"direccion"`
}
type UpdateProveedorInput struct {
	Nombre      *string `json:"nombre"`
	RazonSocial *string `json:"razon_social"`
	RFC         *string `json:"rfc"`
	Telefono    *string `json:"telefono"`
	Email       *string `json:"email"`
	Direccion   *string `json:"direccion"`
	Activo      *bool   `json:"activo"`
}
