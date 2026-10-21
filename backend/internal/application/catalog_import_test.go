package application

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"tienda/backend/internal/domain"

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

func TestParseProviderImportAcceptsCurrentFormFields(t *testing.T) {
	file := workbook(t, []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"}, [][]string{{"  Comercial Norte  ", "Norte SA", "NOR010101AA1", "5551234567", "ventas@norte.test", "Av. Central 1"}})

	rows, result, err := parseProviderImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProviderImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Nombre != "Comercial Norte" || rows[0].Email != "ventas@norte.test" || result.Invalidas != 0 {
		t.Fatalf("rows = %#v, result = %#v", rows, result)
	}
}

func TestParseProviderImportRejectsMissingNameAndInvalidEmail(t *testing.T) {
	file := workbook(t, []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"}, [][]string{{"", "", "RFC010101AA1", "", "", ""}, {"Proveedor", "", "", "", "correo-invalido", ""}})

	rows, result, err := parseProviderImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProviderImport() error = %v", err)
	}
	if len(rows) != 0 || result.Invalidas != 2 || len(result.Errores) != 2 {
		t.Fatalf("rows = %#v, result = %#v", rows, result)
	}
}

func TestParseUnitImportUsesSymbolWithoutCode(t *testing.T) {
	file := workbook(t, []string{"Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}, [][]string{{"Kilogramo", "kg", "PESO", "1", "3"}})

	rows, result, err := parseUnitImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseUnitImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Nombre != "Kilogramo" || rows[0].Simbolo != "kg" || rows[0].Tipo != "PESO" {
		t.Fatalf("rows = %#v, want one unit with symbol", rows)
	}
	if result.Invalidas != 0 || len(result.Errores) != 0 {
		t.Fatalf("result = %#v, want valid import", result)
	}
}

func TestParseUnitImportRejectsLegacyCodeHeader(t *testing.T) {
	file := workbook(t, []string{"Código", "Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}, [][]string{{"kg", "Kilogramo", "kg", "PESO", "1", "3"}})

	_, _, err := parseUnitImport(bytes.NewReader(file))
	if !errors.Is(err, ErrInvalidImportFile) {
		t.Fatalf("parseUnitImport() error = %v, want invalid legacy template", err)
	}
}

func TestUnitImportTemplateExcludesCode(t *testing.T) {
	content, err := CatalogImportTemplate("unidades")
	if err != nil {
		t.Fatalf("CatalogImportTemplate() error = %v", err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatalf("could not open template: %v", err)
	}
	defer book.Close()
	rows, err := book.GetRows("Unidades")
	if err != nil || len(rows) == 0 || !reflect.DeepEqual(rows[0], []string{"Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}) {
		t.Fatalf("unit headers = %#v, error = %v", rows, err)
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

func TestProviderImportTemplateHasExpectedHeaders(t *testing.T) {
	content, err := CatalogImportTemplate("proveedores")
	if err != nil {
		t.Fatalf("CatalogImportTemplate() error = %v", err)
	}
	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	rows, err := book.GetRows("Proveedores")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rows[0], []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("headers = %#v, want %#v", got, want)
	}
}

func TestParseProductImportGeneratesSKUAndReportsFieldErrors(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{
		{"Refresco", "", "Acme", "", "Botella", "600", "ml", "Mililitro", "18.50", "11.125", "4", "7501055303038", "Sabor: Cola", "Distribuidora Norte"},
		{"", "", "", "", "", "-2", "", "", "-10", "0", "texto", "123", "", ""},
	})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].Proveedor != "Distribuidora Norte" || rows[0].Categoria != "" || rows[0].Contenido == nil || *rows[0].Contenido != 600 {
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
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Producto corto", "", "", "", "", "", "", "", "10", "2", "2", "1234567", "", ""}})

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

func TestParseProductImportPreservesNumericBarcodeWithoutScientificNotation(t *testing.T) {
	file := productWorkbookWithNumericBarcode(t, 7501064191909)

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].CodigoBarras != "7501064191909" {
		t.Fatalf("rows = %#v, want the original 13-digit barcode", rows)
	}
	if result.Invalidas != 0 || len(result.Errores) != 0 {
		t.Fatalf("result = %#v, want a valid numeric barcode", result)
	}
}

func TestParseProductImportTreatsVisuallyEmptyBarcodeAsAbsent(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Producto sin código", "", "", "", "", "", "", "", "10", "2", "2", "\u200B\u200C\u200D\u2060\uFEFF", "", ""}})

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 1 || rows[0].CodigoBarras != "" {
		t.Fatalf("rows = %#v, want one product without barcode", rows)
	}
	if result.Invalidas != 0 || len(result.Errores) != 0 || len(result.Advertencias) != 0 {
		t.Fatalf("result = %#v, want valid import without barcode issues", result)
	}
}

func TestParseProductImportRejectsLegacySKUHeader(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Refresco", "", "", "", "", "Botella", "", "", "", "18.50", "2", "4", "", "", ""}})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err == nil || !isImportFormatError(err) || len(rows) != 0 || result.Invalidas != 0 {
		t.Fatalf("rows = %#v, result = %#v, error = %v; want rejected legacy headers", rows, result, err)
	}
}

func TestParseProductImportParsesFlexibleVariantAttributes(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Huevo Kinder", "", "Kinder", "", "Unidad", "", "", "", "28", "12.5", "0", "", "Colección: Dinosaurios; Color: Azul: intenso", ""}})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if result.Invalidas != 0 || len(rows) != 1 || len(rows[0].Variantes) != 2 {
		t.Fatalf("rows = %#v, result = %#v, want one row with two variant attributes", rows, result)
	}
	if rows[0].Variantes[0].Nombre != "Colección" || rows[0].Variantes[1].Valor != "Azul: intenso" {
		t.Fatalf("attributes = %#v, want parsed attributes", rows[0].Variantes)
	}
}

func TestParseProductImportRejectsLegacyEqualsVariantFormat(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Refresco", "", "", "", "Botella", "", "", "", "18.50", "2", "4", "", "Sabor=Cola", ""}})

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 0 || result.Invalidas != 1 || len(result.Errores) != 1 || result.Errores[0].Campo != "Variantes" || result.Errores[0].Motivo != "usa el formato Atributo: valor separado por punto y coma" {
		t.Fatalf("rows = %#v, result = %#v; want rejected legacy variant format", rows, result)
	}
}

func TestParseProductImportRequiresValidCost(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{
		{"Sin costo", "", "", "", "", "", "", "", "10", "", "1", "", "", ""},
		{"Cero", "", "", "", "", "", "", "", "10", "0", "1", "", "", ""},
		{"Negativo", "", "", "", "", "", "", "", "10", "-1", "1", "", "", ""},
		{"Precisión", "", "", "", "", "", "", "", "10", "1.12345", "1", "", "", ""},
		{"Mayor al precio", "", "", "", "", "", "", "", "10", "25.1234", "1", "", "", ""},
	})

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 3 || rows[0].CostoProveedor != 0 || rows[1].CostoProveedor != 0 || rows[2].CostoProveedor != 25.1234 {
		t.Fatalf("rows = %#v, want the empty, zero, and positive cost rows", rows)
	}
	if result.Invalidas != 2 || result.CostosFaltantes != 1 {
		t.Fatalf("result = %#v, want two invalid rows and one missing cost", result)
	}
}

func TestMergeImportResultsPreservesProductSummary(t *testing.T) {
	parsed := domain.CatalogImportResult{
		Procesadas:      5,
		Omitidas:        1,
		Invalidas:       1,
		CostosFaltantes: 1,
		Errores:         []domain.CatalogImportIssue{{Fila: 3, Campo: "Nombre", Motivo: "es obligatorio"}},
		Advertencias:    []domain.CatalogImportIssue{{Fila: 4, Campo: "Código de barras", Motivo: "es corto"}},
	}
	created := domain.CatalogImportResult{
		Creadas:                   3,
		Omitidas:                  1,
		SKUsGenerados:             3,
		ProductosSimplesCreados:   1,
		ProductosBaseCreados:      1,
		ProductosBaseReutilizados: 1,
		VariantesCreadas:          2,
		Errores:                   []domain.CatalogImportIssue{{Fila: 5, Campo: "Código de barras", Motivo: "ya existe"}},
	}

	got := mergeImportResults(parsed, created)
	if got.Procesadas != 5 || got.Creadas != 3 || got.Omitidas != 2 || got.Invalidas != 1 || got.CostosFaltantes != 1 || got.SKUsGenerados != 3 || got.ProductosSimplesCreados != 1 || got.ProductosBaseCreados != 1 || got.ProductosBaseReutilizados != 1 || got.VariantesCreadas != 2 || len(got.Errores) != 2 || len(got.Advertencias) != 1 {
		t.Fatalf("mergeImportResults() = %#v, want complete product import summary", got)
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
	productRows, err := book.GetRows("Productos")
	if err != nil || len(productRows) == 0 || !reflect.DeepEqual(productRows[0], []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}) {
		t.Fatalf("product headers = %#v, error = %v", productRows, err)
	}
	rows, err := book.GetRows("Instrucciones")
	if err != nil || len(rows) < 3 || rows[1][0] != "Campos obligatorios" {
		t.Fatalf("instructions = %#v, error = %v", rows, err)
	}
	instructions := make(map[string]string, len(rows))
	for _, row := range rows {
		if len(row) >= 2 {
			instructions[row[0]] = row[1]
		}
	}
	if instructions["Contenido y unidad de contenido"] != "Indican cuánto trae el empaque. Ejemplo: para una botella de 600 ml, escribe Contenido: 600 y Unidad de contenido: ml." {
		t.Fatalf("content unit instruction = %q", instructions["Contenido y unidad de contenido"])
	}
	if instructions["Unidad de medida"] != "Es la unidad estandarizada del catálogo. Escribe su nombre completo, por ejemplo Mililitro. Si no existe, aparecerá en la vista previa para autorizar su creación." {
		t.Fatalf("measurement unit instruction = %q", instructions["Unidad de medida"])
	}
	if instructions["Variantes"] != "Opcional. Usa Atributo: valor; por ejemplo Colección: Dinosaurios o Sabor: Fresa; Color: Rojo." {
		t.Fatalf("variant instruction = %q, want colon format", instructions["Variantes"])
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

func productWorkbookWithNumericBarcode(t *testing.T, barcode int64) []byte {
	t.Helper()
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Costo", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	book := excelize.NewFile()
	defer book.Close()
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		if err := book.SetCellValue("Sheet1", cell, header); err != nil {
			t.Fatal(err)
		}
	}
	values := []string{"Producto numérico", "", "", "", "", "", "", "", "10", "", "1", "", "", ""}
	for index, value := range values {
		cell, _ := excelize.CoordinatesToCellName(index+1, 2)
		if err := book.SetCellValue("Sheet1", cell, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := book.SetCellValue("Sheet1", "L2", barcode); err != nil {
		t.Fatal(err)
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
