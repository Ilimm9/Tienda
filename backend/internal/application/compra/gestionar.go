package compra

import (
	"tienda/backend/internal/domain/compra"

	"github.com/google/uuid"
)

// EditarDetalleInput es un renglón al editar: con `ID` cambia uno existente, sin él agrega uno nuevo.
type EditarDetalleInput struct {
	ID *string `json:"id"`
	// Quitar cancela el renglón: salen sus piezas y su costo por validar se cancela.
	Quitar bool `json:"quitar"`
	CrearDetalleInput
}

// EditarCompraInput cambia el documento y, si ningún costo está validado, los productos.
// Sin `Detalles` solo cambia el encabezado. El proveedor no se cambia.
type EditarCompraInput struct {
	TipoDocumento  string               `json:"tipo_documento"`
	FolioDocumento string               `json:"folio_documento"`
	NumeroCuenta   *string              `json:"numero_cuenta"`
	FechaDocumento string               `json:"fecha_documento"`
	FechaRecepcion string               `json:"fecha_recepcion"`
	FormaPago      string               `json:"forma_pago"`
	Observaciones  *string              `json:"observaciones"`
	IEPSAjustado   *float64             `json:"ieps_ajustado"`
	Detalles       []EditarDetalleInput `json:"detalles"`
	Motivo         *string              `json:"motivo"`
}

// MotivoInput es el cuerpo de suspender (obligatorio) y reactivar (opcional).
type MotivoInput struct {
	Motivo string `json:"motivo"`
}

// Gestor edita, suspende y reactiva compras; lo implementa infraestructura en una transacción.
type Gestor interface {
	Edit(businessID, userID, purchaseID uuid.UUID, input EditarCompraInput) (compra.CompraProveedor, error)
	Suspend(businessID, userID, purchaseID uuid.UUID, motivo string) (compra.CompraProveedor, error)
	Reactivate(businessID, userID, purchaseID uuid.UUID, motivo string) (compra.CompraProveedor, error)
}

func (s *Service) Edit(businessID, userID, purchaseID uuid.UUID, input EditarCompraInput) (compra.CompraProveedor, error) {
	return s.repo.Edit(businessID, userID, purchaseID, input)
}
func (s *Service) Suspend(businessID, userID, purchaseID uuid.UUID, motivo string) (compra.CompraProveedor, error) {
	return s.repo.Suspend(businessID, userID, purchaseID, motivo)
}
func (s *Service) Reactivate(businessID, userID, purchaseID uuid.UUID, motivo string) (compra.CompraProveedor, error) {
	return s.repo.Reactivate(businessID, userID, purchaseID, motivo)
}

// RenglonVista es un renglón del detalle con lo que la ventana muestra.
type RenglonVista struct {
	compra.CompraProveedorDetalle
	Producto    string  `json:"producto"`
	Unidad      string  `json:"unidad"`
	CostoFinal  float64 `json:"costo_final"`
	EstadoCosto string  `json:"estado_costo"`
}

// CompraVista es la compra con proveedor y renglones vigentes para Detalle y Editar.
type CompraVista struct {
	compra.CompraProveedor
	Proveedor string         `json:"proveedor"`
	Renglones []RenglonVista `json:"renglones"`
}

func (s *Service) View(businessID, id uuid.UUID) (CompraVista, error) {
	return s.repo.View(businessID, id)
}
