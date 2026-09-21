package application

import (
	"bytes"
	"errors"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestParseBrandImportNormalizesNames(t *testing.T) {
	file := workbook(t, []string{"Nombre"}, [][]string{{"Acme"}, {""}, {"  Norte  "}})
	rows, result, err := parseBrandImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseBrandImport() error = %v", err)
	}
	if len(rows) != 2 || rows[1].Nombre != "Norte" {
		t.Fatalf("rows = %#v, want two normalized rows", rows)
	}
	if result.Procesadas != 2 || result.Invalidas != 0 || len(result.Errores) != 0 {
		t.Fatalf("result = %#v, want two valid rows", result)
	}
}

func TestParseCategoryImportRejectsUnexpectedHeaders(t *testing.T) {
	file := workbook(t, []string{"Nombre", "Padre"}, nil)
	_, _, err := parseCategoryImport(bytes.NewReader(file))
	if err == nil || !isImportFormatError(err) {
		t.Fatalf("parseCategoryImport() error = %v, want format error", err)
	}
}

func TestCatalogImportTemplateHasExpectedHeaders(t *testing.T) {
	content, err := CatalogImportTemplate("categorias")
	if err != nil {
		t.Fatalf("CatalogImportTemplate() error = %v", err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("could not open template: %v", err)
	}
	defer book.Close()
	rows, err := book.GetRows("Categorías")
	if err != nil || len(rows) == 0 || len(rows[0]) != 3 || rows[0][2] != "Categoría padre" {
		t.Fatalf("template rows = %#v, error = %v", rows, err)
	}
}

func TestParseProductImportKeepsOptionalSKUAndReportsFieldErrors(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes"}
	file := workbook(t, headers, [][]string{
		{"Refresco", "REF-600", "", "Acme", "", "Botella", "600", "ml", "Mililitro", "18.50", "4", "7501055303038", "Sabor=Cola"},
		{"", "", "", "", "", "", "-2", "", "", "-10", "texto", "123", ""},
	})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].SKUInterno != "REF-600" || rows[0].Categoria != "" || rows[0].Contenido == nil || *rows[0].Contenido != 600 {
		t.Fatalf("rows = %#v, want one parsed product", rows)
	}
	if result.Procesadas != 2 || result.Invalidas != 1 || len(result.Errores) != 4 || len(result.Advertencias) != 1 {
		t.Fatalf("result = %#v, want one invalid row with four field errors and one warning", result)
	}
	if result.Errores[0].Campo != "Nombre" || result.Errores[1].Campo != "Contenido" {
		t.Fatalf("errors = %#v, want field-specific errors", result.Errores)
	}
}

func TestParseProductImportWarnsForShortNumericBarcode(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes"}
	file := workbook(t, headers, [][]string{{"Producto corto", "", "", "", "", "", "", "", "", "10", "2", "1234567", ""}})

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || result.Invalidas != 0 || len(result.Errores) != 0 || len(result.Advertencias) != 1 {
		t.Fatalf("rows = %#v, result = %#v, want one insertable row with one warning", rows, result)
	}
	if result.Advertencias[0].Fila != 2 || result.Advertencias[0].Campo != "Código de barras" {
		t.Fatalf("warnings = %#v, want barcode warning for row 2", result.Advertencias)
	}
}

func TestParseProductImportAllowsMissingSKU(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes"}
	file := workbook(t, headers, [][]string{{"Refresco", "", "", "", "", "Botella", "", "", "", "18.50", "4", "", ""}})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].SKUInterno != "" || result.Invalidas != 0 || len(result.Errores) != 0 {
		t.Fatalf("rows = %#v, result = %#v, want one valid row without SKU", rows, result)
	}
}

func TestParseProductImportParsesFlexibleVariantAttributes(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes"}
	file := workbook(t, headers, [][]string{{"Huevo Kinder", "", "", "Kinder", "", "Unidad", "", "", "", "28", "0", "", "Colección=Dinosaurios; Color=Azul"}})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if result.Invalidas != 0 || len(rows) != 1 || len(rows[0].Variantes) != 2 {
		t.Fatalf("rows = %#v, result = %#v, want one row with two variant attributes", rows, result)
	}
	if rows[0].Variantes[0].Nombre != "Colección" || rows[0].Variantes[1].Valor != "Azul" {
		t.Fatalf("attributes = %#v, want parsed attributes", rows[0].Variantes)
	}
}

func TestProductImportTemplateIncludesInstructions(t *testing.T) {
	content, err := CatalogImportTemplate("productos")
	if err != nil {
		t.Fatalf("CatalogImportTemplate() error = %v", err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("could not open template: %v", err)
	}
	defer book.Close()
	if book.GetSheetName(0) != "Productos" {
		t.Fatalf("first sheet = %q, want Productos", book.GetSheetName(0))
	}
	rows, err := book.GetRows("Instrucciones")
	if err != nil || len(rows) < 3 || rows[1][0] != "Campos obligatorios" {
		t.Fatalf("instructions = %#v, error = %v", rows, err)
	}
}

func workbook(t *testing.T, headers []string, data [][]string) []byte {
	t.Helper()
	book := excelize.NewFile()
	defer book.Close()
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		if err := book.SetCellValue("Sheet1", cell, header); err != nil {
			t.Fatal(err)
		}
	}
	for rowIndex, row := range data {
		for columnIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+2)
			if err := book.SetCellValue("Sheet1", cell, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	var buffer bytes.Buffer
	if err := book.Write(&buffer); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func isImportFormatError(err error) bool {
	return errors.Is(err, ErrInvalidImportFile)
}
