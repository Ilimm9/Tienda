package application

import (
	"bytes"
	"errors"
	"testing"

	"tienda/backend/internal/domain"

	"github.com/google/uuid"
)

type lookupStub struct {
	calledWith string
}

func (s *lookupStub) LookupProduct(barcode string) (domain.ProductLookup, error) {
	s.calledWith = barcode
	return domain.ProductLookup{CodigoBarras: barcode}, nil
}

func TestLookupProductValidatesBarcode(t *testing.T) {
	lookup := &lookupStub{}
	service := NewProductService(nil, lookup)

	if _, err := service.LookupProduct("abc"); !errors.Is(err, ErrInvalidBarcode) {
		t.Fatalf("error = %v, se esperaba ErrInvalidBarcode", err)
	}
	if lookup.calledWith != "" {
		t.Fatal("no se debía consultar al proveedor con un código inválido")
	}

	if _, err := service.LookupProduct(" 7501055303038 "); err != nil {
		t.Fatalf("código válido rechazado: %v", err)
	}
	if lookup.calledWith != "7501055303038" {
		t.Fatalf("código enviado = %q", lookup.calledWith)
	}
}

// importRepositoryStub rechaza la fila 3 por duplicado y acepta las demás.
type importRepositoryStub struct {
	ProductRepository
}

func (importRepositoryStub) ValidateProductImport(_ uuid.UUID, _ uuid.UUID, rows []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, []domain.ProductImportCatalogProposal, error) {
	result := domain.CatalogImportResult{}
	prepared := make([]domain.ValidatedProductImportRow, 0, len(rows))
	for _, row := range rows {
		if row.Fila == 3 {
			result.Omitidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: row.Fila, Campo: "Código de barras", Motivo: "ya está registrado en el negocio"})
			continue
		}
		prepared = append(prepared, domain.ValidatedProductImportRow{Fila: row.Fila})
	}
	return prepared, result, nil, nil
}

func TestPreviewProductImportListsValidRowsAndNamesErrors(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{
		{"Agua 1.5 L", "Bebidas", "", "", "", "", "", "", "12", "8.380000000000001", "4", "", "", ""},
		{"Duplicado", "Bebidas", "", "", "", "", "", "", "10", "5", "1", "7501000000001", "", ""},
		{"Negativo", "Bebidas", "", "", "", "", "", "", "10", "5", "-2", "", "", ""},
	})
	service := NewProductService(importRepositoryStub{}, nil)

	preview, err := service.PreviewProductImport(uuid.New(), uuid.New(), bytes.NewReader(file))
	if err != nil {
		t.Fatalf("PreviewProductImport() error = %v", err)
	}
	want := domain.ProductImportValidRow{Fila: 2, Nombre: "Agua 1.5 L", Categoria: "Bebidas", PrecioVenta: 12, Costo: 8.38, StockInicial: 4}
	if len(preview.FilasCorrectas) != 1 || preview.FilasCorrectas[0] != want {
		t.Fatalf("filas_correctas = %#v, want %#v", preview.FilasCorrectas, want)
	}
	names := map[int]string{}
	for _, issue := range preview.Errores {
		names[issue.Fila] = issue.Nombre
	}
	if names[3] != "Duplicado" || names[4] != "Negativo" {
		t.Fatalf("errores = %#v, want names for parse and repository errors", preview.Errores)
	}

	content, err := service.ProductImportErrors(uuid.New(), uuid.New(), file)
	if err != nil || len(content) == 0 {
		t.Fatalf("ProductImportErrors() = %d bytes, %v", len(content), err)
	}
	clean := workbook(t, headers, [][]string{{"Agua 1.5 L", "Bebidas", "", "", "", "", "", "", "12", "8", "4", "", "", ""}})
	if _, err := service.ProductImportErrors(uuid.New(), uuid.New(), clean); !errors.Is(err, ErrNoImportErrors) {
		t.Fatalf("ProductImportErrors() error = %v, want ErrNoImportErrors", err)
	}
}
