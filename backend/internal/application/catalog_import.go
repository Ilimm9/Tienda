package application

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"tienda/backend/internal/domain"
	"unicode"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

var ErrInvalidImportFile = errors.New("el archivo no tiene el formato de plantilla esperado")
var ErrNoImportErrors = errors.New("el archivo no tiene filas con errores")

// productImportHeaders es el formato único de la plantilla de productos
// (adecuación 1): plantilla, lectura y archivo de errores usan este orden.
var productImportHeaders = []string{"Código de barras", "Nombre", "Variantes", "Precio de venta", "Costo", "Stock inicial", "Unidad de medida", "Categoría", "Proveedor"}

// legacyProductImportHeaders identifica la plantilla anterior para explicar el rechazo.
var legacyProductImportHeaders = []string{"Nombre", "Categoría", "Marca", "Descripción", "Presentación"}

// Posición de cada columna de productImportHeaders.
const (
	productColCodigo = iota
	productColNombre
	productColVariantes
	productColPrecio
	productColCosto
	productColStock
	productColUnidad
	productColCategoria
	productColProveedor
)

var shortBarcodePattern = regexp.MustCompile(`^\d{1,7}$`)
var costPattern = regexp.MustCompile(`^\d+(?:\.\d+)?$`)

func (s *ProductService) ImportBrands(businessID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseBrandImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportBrands(businessID, rows)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportCategories(businessID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseCategoryImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportCategories(businessID, rows)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportUnits(businessID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseUnitImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportUnits(businessID, rows)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportProviders(businessID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseProviderImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportProviders(businessID, rows)
	return mergeImportResults(result, imported), err
}

func productImportErrorsWorkbook(content []byte, issues []domain.CatalogImportIssue) ([]byte, error) {
	headers := productImportHeaders
	rows, _, err := readProductImportRows(bytes.NewReader(content), headers)
	if err != nil {
		return nil, err
	}
	reasons := make(map[int][]string)
	for _, issue := range issues {
		reason := issue.Motivo
		if issue.Campo != "" {
			reason = issue.Campo + ": " + reason
		}
		reasons[issue.Fila] = append(reasons[issue.Fila], reason)
	}
	book := excelize.NewFile()
	defer book.Close()
	sheet := "Productos"
	book.SetSheetName(book.GetSheetName(0), sheet)
	errorHeaders := append(append([]string{}, headers...), "Errores")
	if err := book.SetSheetRow(sheet, "A1", &errorHeaders); err != nil {
		return nil, err
	}
	next := 2
	for index, row := range rows {
		rowReasons, failed := reasons[index+2]
		if !failed {
			continue
		}
		values := make([]string, len(headers)+1)
		for column := range headers {
			values[column] = cell(row, column)
		}
		values[len(headers)] = strings.Join(rowReasons, "; ")
		start, _ := excelize.CoordinatesToCellName(1, next)
		if err := book.SetSheetRow(sheet, start, &values); err != nil {
			return nil, err
		}
		next++
	}
	style, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"2563EB"}, Pattern: 1}})
	if err != nil {
		return nil, err
	}
	lastData, _ := excelize.ColumnNumberToName(len(headers))
	errorColumn, _ := excelize.ColumnNumberToName(len(errorHeaders))
	book.SetCellStyle(sheet, "A1", errorColumn+"1", style)
	book.SetColWidth(sheet, "A", lastData, 22)
	book.SetColWidth(sheet, errorColumn, errorColumn, 70)
	book.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	var output bytes.Buffer
	if err := book.Write(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func CatalogImportTemplate(section string) ([]byte, error) {
	book := excelize.NewFile()
	defer book.Close()
	sheet := "Marcas"
	headers := []string{"Nombre"}
	if section == "categorias" {
		sheet = "Categorías"
		headers = []string{"Nombre", "Descripción", "Categoría padre"}
	}
	if section == "unidades" {
		sheet = "Unidades"
		headers = []string{"Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}
	}
	if section == "proveedores" {
		sheet = "Proveedores"
		headers = []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"}
	}
	if section == "productos" {
		sheet = "Productos"
		headers = productImportHeaders
	}
	book.SetSheetName(book.GetSheetName(0), sheet)
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		book.SetCellValue(sheet, cell, header)
	}
	style, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"2563EB"}, Pattern: 1}})
	if err != nil {
		return nil, err
	}
	lastColumn, _ := excelize.ColumnNumberToName(len(headers))
	book.SetCellStyle(sheet, "A1", lastColumn+"1", style)
	book.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	for index := range headers {
		column, _ := excelize.ColumnNumberToName(index + 1)
		book.SetColWidth(sheet, column, column, 28)
	}
	if section == "productos" {
		instructions, err := book.NewSheet("Instrucciones")
		if err != nil {
			return nil, err
		}
		instructionRows := [][]string{
			{"Carga masiva de productos", "Completa una fila por producto. La hoja Productos debe conservar sus encabezados, en el mismo orden y sin columnas extra."},
			{"Campos obligatorios", "Nombre, Precio de venta y Stock inicial."},
			{"Campos opcionales", "Código de barras, Variantes, Costo, Unidad de medida, Categoría y Proveedor."},
			{"Nombre", "Incluye el tamaño o presentación en el nombre. Ejemplo: Coca-Cola 600 ml, Coca-Cola 2 L, Corona Mega."},
			{"Unidad de medida", "Es la unidad estandarizada del catálogo. Escribe su nombre completo, por ejemplo Pieza. Si no existe, aparecerá en la vista previa para autorizar su creación."},
			{"Números", "Precio, stock y costo deben ser números mayores o iguales a cero. El costo se redondea a 2 decimales."},
			{"Catálogos", "Unidad de medida, Categoría y Proveedor son opcionales. Si indicas uno nuevo, revísalo y autoriza su creación en la vista previa."},
			{"Código de barras", "De 8 a 14 dígitos se usa para buscar una imagen automáticamente. De 1 a 7 dígitos se acepta con advertencia y sin búsqueda de imagen."},
			{"SKU interno", "Se genera automáticamente al importar y no debe incluirse como columna en el archivo."},
			{"Variantes", "Opcional. Usa Atributo: valor; por ejemplo Colección: Dinosaurios o Sabor: Fresa; Color: Rojo."},
			{"Productos existentes", "El código de barras manda: si ya está registrado, la fila corresponde a ese producto. Sin código, el producto se busca por su nombre. Un nombre existente con otro código debe indicar la variante con el formato Atributo: valor."},
			{"Resultado", "Las filas inválidas se reportan con fila, campo y motivo; las válidas se crean."},
		}
		for rowIndex, row := range instructionRows {
			for columnIndex, value := range row {
				cell, _ := excelize.CoordinatesToCellName(columnIndex+1, rowIndex+1)
				book.SetCellValue("Instrucciones", cell, value)
			}
		}
		book.SetColWidth("Instrucciones", "A", "A", 24)
		book.SetColWidth("Instrucciones", "B", "B", 95)
		book.SetCellStyle("Instrucciones", "A1", "B1", style)
		book.SetActiveSheet(instructions)
		book.SetActiveSheet(0)
	}
	var output bytes.Buffer
	if err := book.Write(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func parseProductImport(file io.Reader) ([]domain.ProductImportRow, domain.CatalogImportResult, error) {
	headers := productImportHeaders
	rows, result, err := readProductImportRows(file, headers)
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.ProductImportRow, 0, len(rows))
	for index, row := range rows {
		rowNumber := index + 2
		hasError := false
		addError := func(field, reason string) {
			hasError = true
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: rowNumber, Nombre: cell(row, productColNombre), Campo: field, Motivo: reason})
		}
		addWarning := func(field, reason string) {
			result.Advertencias = append(result.Advertencias, domain.CatalogImportIssue{Fila: rowNumber, Nombre: cell(row, productColNombre), Campo: field, Motivo: reason})
		}
		if cell(row, productColNombre) == "" {
			addError("Nombre", "es obligatorio")
		}
		precioText, costText, stockText := cell(row, productColPrecio), cell(row, productColCosto), cell(row, productColStock)
		precio, precioErr := strconv.ParseFloat(precioText, 64)
		stock, stockErr := strconv.ParseFloat(stockText, 64)
		if precioText == "" {
			addError("Precio de venta", "es obligatorio")
		} else if precioErr != nil || precio < 0 {
			addError("Precio de venta", "debe ser un número mayor o igual a cero")
		}
		cost := 0.0
		costCaptured := costText != ""
		if costText == "" {
			result.CostosFaltantes++
		} else {
			parsedCost, costErr := parseImportedCost(costText)
			if costErr != nil {
				addError("Costo", costErr.Error())
			} else {
				cost = parsedCost
			}
		}
		if stockText == "" {
			addError("Stock inicial", "es obligatorio")
		} else if stockErr != nil || stock < 0 {
			addError("Stock inicial", "debe ser un número mayor o igual a cero")
		}
		barcode := optionalBarcodeCell(row, productColCodigo)
		if barcode != "" && !barcodePattern.MatchString(barcode) && !shortBarcodePattern.MatchString(barcode) {
			addError("Código de barras", "debe contener solo de 1 a 14 dígitos")
		} else if shortBarcodePattern.MatchString(barcode) {
			addWarning("Código de barras", "tiene menos de 8 dígitos; se importará sin búsqueda automática de imagen")
		}
		variantes, variantErr := parseVariantAttributes(cell(row, productColVariantes))
		if variantErr != nil {
			addError("Variantes", variantErr.Error())
		}
		if hasError {
			result.Invalidas++
			continue
		}
		valid = append(valid, domain.ProductImportRow{Fila: rowNumber, Nombre: cell(row, productColNombre), Categoria: cell(row, productColCategoria), UnidadMedida: cell(row, productColUnidad), PrecioVenta: precio, CostoProveedor: cost, CostoCapturado: costCaptured, StockInicial: stock, CodigoBarras: barcode, Variantes: variantes, Proveedor: cell(row, productColProveedor)})
	}
	return valid, result, nil
}

// parseImportedCost redondea a 2 decimales: Excel guarda residuos de flotante
// como 8.380000000000001 aunque muestre 8.38.
func parseImportedCost(value string) (float64, error) {
	if !costPattern.MatchString(value) {
		return 0, errors.New("debe ser un número mayor o igual a cero")
	}
	integerPart := strings.SplitN(value, ".", 2)[0]
	if len(strings.TrimLeft(integerPart, "0")) > 10 {
		return 0, errors.New("debe ser un número mayor o igual a cero dentro del rango permitido")
	}
	cost, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 {
		return 0, errors.New("debe ser un número mayor o igual a cero dentro del rango permitido")
	}
	return math.Round(cost*100) / 100, nil
}

func parseVariantAttributes(value string) ([]domain.ProductVariantAttributeInput, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	attributes := make([]domain.ProductVariantAttributeInput, 0)
	seen := make(map[string]struct{})
	for _, item := range strings.Split(value, ";") {
		parts := strings.SplitN(item, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, errors.New("usa el formato Atributo: valor separado por punto y coma")
		}
		name, attributeValue := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		key := strings.ToLower(strings.TrimSpace(name))
		if _, exists := seen[key]; exists {
			return nil, errors.New("no puede repetir el mismo atributo")
		}
		seen[key] = struct{}{}
		attributes = append(attributes, domain.ProductVariantAttributeInput{Nombre: name, Valor: attributeValue})
	}
	return attributes, nil
}

func parseUnitImport(file io.Reader) ([]domain.CatalogImportUnitRow, domain.CatalogImportResult, error) {
	rows, result, err := readImportRows(file, []string{"Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"})
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.CatalogImportUnitRow, 0, len(rows))
	for index, row := range rows {
		factor, factorErr := strconv.ParseFloat(cell(row, 3), 64)
		decimals, decimalsErr := strconv.Atoi(cell(row, 4))
		if cell(row, 0) == "" || cell(row, 1) == "" || cell(row, 2) == "" || factorErr != nil || decimalsErr != nil || factor <= 0 || decimals < 0 || decimals > 6 {
			result.Invalidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: index + 2, Motivo: "La unidad requiere nombre, símbolo, tipo, factor válido y decimales entre 0 y 6"})
			continue
		}
		valid = append(valid, domain.CatalogImportUnitRow{Fila: index + 2, Nombre: cell(row, 0), Simbolo: cell(row, 1), Tipo: cell(row, 2), FactorABase: factor, Decimales: decimals})
	}
	return valid, result, nil
}

func parseBrandImport(file io.Reader) ([]domain.CatalogImportBrandRow, domain.CatalogImportResult, error) {
	rows, result, err := readImportRows(file, []string{"Nombre"})
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.CatalogImportBrandRow, 0, len(rows))
	for index, row := range rows {
		nombre := cell(row, 0)
		if nombre == "" {
			result.Invalidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: index + 2, Motivo: "El nombre es obligatorio"})
			continue
		}
		valid = append(valid, domain.CatalogImportBrandRow{Fila: index + 2, Nombre: nombre})
	}
	return valid, result, nil
}

func parseCategoryImport(file io.Reader) ([]domain.CatalogImportCategoryRow, domain.CatalogImportResult, error) {
	rows, result, err := readImportRows(file, []string{"Nombre", "Descripción", "Categoría padre"})
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.CatalogImportCategoryRow, 0, len(rows))
	for index, row := range rows {
		nombre := cell(row, 0)
		if nombre == "" {
			result.Invalidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: index + 2, Motivo: "El nombre es obligatorio"})
			continue
		}
		valid = append(valid, domain.CatalogImportCategoryRow{Fila: index + 2, Nombre: nombre, Descripcion: cell(row, 1), CategoriaPadre: cell(row, 2)})
	}
	return valid, result, nil
}

func parseProviderImport(file io.Reader) ([]domain.CatalogImportProviderRow, domain.CatalogImportResult, error) {
	rows, result, err := readImportRows(file, []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"})
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.CatalogImportProviderRow, 0, len(rows))
	for index, row := range rows {
		rowNumber := index + 2
		nombre, email := cell(row, 0), cell(row, 4)
		if nombre == "" {
			result.Invalidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: rowNumber, Campo: "Nombre", Motivo: "es obligatorio"})
			continue
		}
		if email != "" {
			address, parseErr := mail.ParseAddress(email)
			if parseErr != nil || address.Address != email {
				result.Invalidas++
				result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: rowNumber, Campo: "Correo", Motivo: "debe ser una dirección válida"})
				continue
			}
		}
		valid = append(valid, domain.CatalogImportProviderRow{Fila: rowNumber, Nombre: nombre, RazonSocial: cell(row, 1), RFC: cell(row, 2), Telefono: cell(row, 3), Email: email, Direccion: cell(row, 5)})
	}
	return valid, result, nil
}

// readProductImportRows exige exactamente las columnas de la plantilla vigente:
// mismo orden y ninguna columna extra (adecuación 1).
func readProductImportRows(file io.Reader, headers []string) ([][]string, domain.CatalogImportResult, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, domain.CatalogImportResult{}, fmt.Errorf("%w: no se pudo leer el archivo XLSX", ErrInvalidImportFile)
	}
	if err := checkProductImportHeaders(content, headers); err != nil {
		return nil, domain.CatalogImportResult{}, err
	}
	return readImportRowsWithOptions(bytes.NewReader(content), headers, excelize.Options{RawCellValue: true})
}

func checkProductImportHeaders(content []byte, headers []string) error {
	book, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("%w: no se pudo leer el archivo XLSX", ErrInvalidImportFile)
	}
	defer book.Close()
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return ErrInvalidImportFile
	}
	rows, err := book.GetRows(sheets[0])
	if err != nil || len(rows) == 0 {
		return ErrInvalidImportFile
	}
	first := rows[0]
	legacy := len(first) >= len(legacyProductImportHeaders)
	for index, expected := range legacyProductImportHeaders {
		if legacy && cell(first, index) != expected {
			legacy = false
		}
	}
	if legacy {
		return fmt.Errorf("%w: es la plantilla anterior; descarga la plantilla nueva", ErrInvalidImportFile)
	}
	for index, expected := range headers {
		if cell(first, index) != expected {
			return fmt.Errorf("%w: se esperaba la columna %q en la posición %d", ErrInvalidImportFile, expected, index+1)
		}
	}
	for index := len(headers); index < len(first); index++ {
		// El archivo de errores descargado se puede volver a subir con su columna «Errores».
		if extra := cell(first, index); extra != "" && !(index == len(headers) && extra == "Errores") {
			return fmt.Errorf("%w: la columna %q no pertenece a la plantilla", ErrInvalidImportFile, extra)
		}
	}
	return nil
}

func readImportRows(file io.Reader, headers []string) ([][]string, domain.CatalogImportResult, error) {
	return readImportRowsWithOptions(file, headers)
}

func readImportRowsWithOptions(file io.Reader, headers []string, options ...excelize.Options) ([][]string, domain.CatalogImportResult, error) {
	book, err := excelize.OpenReader(file)
	if err != nil {
		return nil, domain.CatalogImportResult{}, fmt.Errorf("%w: no se pudo leer el archivo XLSX", ErrInvalidImportFile)
	}
	defer book.Close()
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, domain.CatalogImportResult{}, ErrInvalidImportFile
	}
	rows, err := book.GetRows(sheets[0], options...)
	if err != nil || len(rows) == 0 {
		return nil, domain.CatalogImportResult{}, ErrInvalidImportFile
	}
	for index, expected := range headers {
		if cell(rows[0], index) != expected {
			return nil, domain.CatalogImportResult{}, fmt.Errorf("%w: se esperaba la columna %q", ErrInvalidImportFile, expected)
		}
	}
	data := make([][]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if !isEmptyRow(row) {
			data = append(data, row)
		}
	}
	return data, domain.CatalogImportResult{Procesadas: len(data), Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}, nil
}

func cell(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

// optionalBarcodeCell treats cells that only contain whitespace or invisible
// Unicode marks as empty. Spreadsheet exports can leave these marks behind
// even when a barcode cell looks blank to the user.
func optionalBarcodeCell(row []string, index int) string {
	value := cell(row, index)
	if value == "" || !containsVisibleRune(value) {
		return ""
	}
	return value
}

func containsVisibleRune(value string) bool {
	for _, runeValue := range value {
		if unicode.IsSpace(runeValue) {
			continue
		}
		switch runeValue {
		case '\u200B', '\u200C', '\u200D', '\u2060', '\uFEFF':
			continue
		default:
			return true
		}
	}
	return false
}

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func mergeImportResults(first, second domain.CatalogImportResult) domain.CatalogImportResult {
	return domain.CatalogImportResult{
		Procesadas:                first.Procesadas,
		Creadas:                   first.Creadas + second.Creadas,
		Omitidas:                  first.Omitidas + second.Omitidas,
		Invalidas:                 first.Invalidas + second.Invalidas,
		SKUsGenerados:             first.SKUsGenerados + second.SKUsGenerados,
		ProductosSimplesCreados:   first.ProductosSimplesCreados + second.ProductosSimplesCreados,
		ProductosBaseCreados:      first.ProductosBaseCreados + second.ProductosBaseCreados,
		ProductosBaseReutilizados: first.ProductosBaseReutilizados + second.ProductosBaseReutilizados,
		VariantesCreadas:          first.VariantesCreadas + second.VariantesCreadas,
		CostosFaltantes:           first.CostosFaltantes + second.CostosFaltantes,
		AgregadasSucursal:         first.AgregadasSucursal + second.AgregadasSucursal,
		Actualizadas:              first.Actualizadas + second.Actualizadas,
		Errores:                   append(first.Errores, second.Errores...),
		Advertencias:              append(first.Advertencias, second.Advertencias...),
	}
}
