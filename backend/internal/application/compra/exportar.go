package compra

import (
	"time"

	"github.com/google/uuid"
)

// FiltroExportacion elige las compras del Excel. Sin `CompraIDs` se usan proveedor y fechas de recepción.
type FiltroExportacion struct {
	SucursalID          uuid.UUID
	ProveedorID         *uuid.UUID
	Desde               *time.Time
	Hasta               *time.Time
	CompraIDs           []uuid.UUID
	IncluirSuspendidas  bool
	IncluirComprobacion bool
}

// RenglonExportacion es una fila de «hoja 1»: un producto comprado.
type RenglonExportacion struct {
	CompraID            uuid.UUID
	Codigo              string
	Producto            string
	TasaIEPS            float64
	ImpuestosPorRenglon bool
	Subtotal            float64
	IEPS                float64
	IVA                 float64
	TasaIVA             float64
	CantidadEmpaques    int
	Unidad              string
	PiezasPorEmpaque    int
	TotalPiezas         int
	CostoFinal          float64
	CostoAnterior       *float64
	PrecioAnterior      *float64
	PrecioNuevo         *float64
	ProductoNuevo       bool
	Contenido           string
	PrecioAutorizado    *float64
	PiezasDanadas       int
	PiezasFaltantes     int
	PiezasBuenas        int
	Descuento           float64
	EstadoValidacion    string
}

// CompraExportacion agrupa los renglones de un ticket; alimenta la hoja COMPROBACIÓN.
type CompraExportacion struct {
	ID             uuid.UUID
	Ticket         string
	FechaRecepcion time.Time
	Proveedor      string
	Estado         string
	IEPS           float64
	Total          float64
	Observaciones  string
	Descuento      float64
	Renglones      []RenglonExportacion
}

// Exportador arma el archivo; lo implementa infraestructura con excelize.
type Exportador interface {
	Generar(compras []CompraExportacion, comprobacion bool) ([]byte, error)
}

// ConExportador agrega el generador de Excel al servicio.
func (s *Service) ConExportador(exportador Exportador) *Service {
	s.exportador = exportador
	return s
}

// Exportar arma el Excel de las compras del filtro.
func (s *Service) Exportar(businessID uuid.UUID, filtro FiltroExportacion) ([]byte, int, error) {
	compras, err := s.repo.ListExport(businessID, filtro)
	if err != nil {
		return nil, 0, err
	}
	archivo, err := s.exportador.Generar(compras, filtro.IncluirComprobacion)
	return archivo, len(compras), err
}
