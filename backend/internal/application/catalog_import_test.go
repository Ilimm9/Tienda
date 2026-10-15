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
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{
		{"Refresco", "", "Acme", "", "Botella", "600", "ml", "Mililitro", "18.50", "4", "7501055303038", "Sabor: Cola", "Distribuidora Norte"},
		{"", "", "", "", "", "-2", "", "", "-10", "texto", "123", "", ""},
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
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Producto corto", "", "", "", "", "", "", "", "10", "2", "1234567", "", ""}})

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

func TestParseProductImportRejectsLegacySKUHeader(t *testing.T) {
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Refresco", "", "", "", "", "Botella", "", "", "", "18.50", "4", "", "", ""}})
	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err == nil || !isImportFormatError(err) || len(rows) != 0 || result.Invalidas != 0 {
		t.Fatalf("rows = %#v, result = %#v, error = %v; want rejected legacy headers", rows, result, err)
	}
}

func TestParseProductImportParsesFlexibleVariantAttributes(t *testing.T) {
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Huevo Kinder", "", "Kinder", "", "Unidad", "", "", "", "28", "0", "", "Colección: Dinosaurios; Color: Azul: intenso", ""}})
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
	headers := []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}
	file := workbook(t, headers, [][]string{{"Refresco", "", "", "", "Botella", "", "", "", "18.50", "4", "", "Sabor=Cola", ""}})

	rows, result, err := parseProductImport(bytes.NewReader(file))
	if err != nil {
		t.Fatalf("parseProductImport() error = %v", err)
	}
	if len(rows) != 0 || result.Invalidas != 1 || len(result.Errores) != 1 || result.Errores[0].Campo != "Variantes" || result.Errores[0].Motivo != "usa el formato Atributo: valor separado por punto y coma" {
		t.Fatalf("rows = %#v, result = %#v; want rejected legacy variant format", rows, result)
	}
}

func TestMergeImportResultsPreservesProductSummary(t *testing.T) {
	parsed := domain.CatalogImportResult{
		Procesadas:   5,
		Omitidas:     1,
		Invalidas:    1,
		Errores:      []domain.CatalogImportIssue{{Fila: 3, Campo: "Nombre", Motivo: "es obligatorio"}},
		Advertencias: []domain.CatalogImportIssue{{Fila: 4, Campo: "Código de barras", Motivo: "es corto"}},
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
	if got.Procesadas != 5 || got.Creadas != 3 || got.Omitidas != 2 || got.Invalidas != 1 || got.SKUsGenerados != 3 || got.ProductosSimplesCreados != 1 || got.ProductosBaseCreados != 1 || got.ProductosBaseReutilizados != 1 || got.VariantesCreadas != 2 || len(got.Errores) != 2 || len(got.Advertencias) != 1 {
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
	if err != nil || len(productRows) == 0 || !reflect.DeepEqual(productRows[0], []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras", "Variantes", "Proveedor"}) {
		t.Fatalf("product headers = %#v, error = %v", productRows, err)
	}
	rows, err := book.GetRows("Instrucciones")
	if err != nil || len(rows) < 3 || rows[1][0] != "Campos obligatorios" {
		t.Fatalf("instructions = %#v, error = %v", rows, err)
	}
	if rows[7][1] != "Opcional. Usa Atributo: valor; por ejemplo Colección: Dinosaurios o Sabor: Fresa; Color: Rojo." {
		t.Fatalf("variant instruction = %q, want colon format", rows[7][1])
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
