package application

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

var ErrInvalidImportFile = errors.New("el archivo no tiene el formato de plantilla esperado")

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
		headers = []string{"Código", "Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}
	}
	if section == "productos" {
		sheet = "Productos"
		headers = []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras"}
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
			{"Carga masiva de productos", "Completa una fila por producto. La hoja Productos debe conservar sus encabezados."},
			{"Campos obligatorios", "Nombre, Precio de venta y Stock inicial."},
			{"Campos opcionales", "SKU interno, Categoría, Marca, Descripción, Presentación, Contenido, Unidad de contenido, Unidad de medida y Código de barras."},
			{"Números", "Precio, stock y contenido deben ser números mayores o iguales a cero."},
			{"Catálogos", "Categoría, Marca y Unidad de medida solo se validan si se indican y deben existir activas."},
			{"Código de barras", "Si se indica, debe tener de 8 a 14 dígitos y se usará para buscar una imagen automáticamente."},
			{"Duplicados", "SKU interno y código de barras, cuando se indiquen, no pueden repetirse."},
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
	headers := []string{"Nombre", "SKU interno", "Categoría", "Marca", "Descripción", "Presentación", "Contenido", "Unidad de contenido", "Unidad de medida", "Precio de venta", "Stock inicial", "Código de barras"}
	rows, result, err := readImportRows(file, headers)
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.ProductImportRow, 0, len(rows))
	for index, row := range rows {
		rowNumber := index + 2
		hasError := false
		addError := func(field, reason string) {
			hasError = true
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: rowNumber, Campo: field, Motivo: reason})
		}
		if cell(row, 0) == "" {
			addError("Nombre", "es obligatorio")
		}
		contenidoText := cell(row, 6)
		var contenido *float64
		if contenidoText != "" {
			value, parseErr := strconv.ParseFloat(contenidoText, 64)
			if parseErr != nil || value < 0 {
				addError("Contenido", "debe ser un número mayor o igual a cero")
			} else {
				contenido = &value
			}
		}
		precioText, stockText := cell(row, 9), cell(row, 10)
		precio, precioErr := strconv.ParseFloat(precioText, 64)
		stock, stockErr := strconv.ParseFloat(stockText, 64)
		if precioText == "" {
			addError("Precio de venta", "es obligatorio")
		} else if precioErr != nil || precio < 0 {
			addError("Precio de venta", "debe ser un número mayor o igual a cero")
		}
		if stockText == "" {
			addError("Stock inicial", "es obligatorio")
		} else if stockErr != nil || stock < 0 {
			addError("Stock inicial", "debe ser un número mayor o igual a cero")
		}
		barcode := cell(row, 11)
		if barcode != "" && !barcodePattern.MatchString(barcode) {
			addError("Código de barras", "debe contener entre 8 y 14 dígitos")
		}
		if hasError {
			result.Invalidas++
			continue
		}
		valid = append(valid, domain.ProductImportRow{Fila: rowNumber, Nombre: cell(row, 0), SKUInterno: cell(row, 1), Categoria: cell(row, 2), Marca: cell(row, 3), Descripcion: cell(row, 4), Presentacion: cell(row, 5), Contenido: contenido, UnidadContenido: cell(row, 7), UnidadMedida: cell(row, 8), PrecioVenta: precio, StockInicial: stock, CodigoBarras: barcode})
	}
	return valid, result, nil
}

func parseUnitImport(file io.Reader) ([]domain.CatalogImportUnitRow, domain.CatalogImportResult, error) {
	rows, result, err := readImportRows(file, []string{"Código", "Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"})
	if err != nil {
		return nil, result, err
	}
	valid := make([]domain.CatalogImportUnitRow, 0, len(rows))
	for index, row := range rows {
		factor, factorErr := strconv.ParseFloat(cell(row, 4), 64)
		decimals, decimalsErr := strconv.Atoi(cell(row, 5))
		if cell(row, 0) == "" || cell(row, 1) == "" || cell(row, 2) == "" || cell(row, 3) == "" || factorErr != nil || decimalsErr != nil || factor <= 0 || decimals < 0 || decimals > 6 {
			result.Invalidas++
			result.Errores = append(result.Errores, domain.CatalogImportIssue{Fila: index + 2, Motivo: "La unidad requiere código, nombre, símbolo, tipo, factor válido y decimales entre 0 y 6"})
			continue
		}
		valid = append(valid, domain.CatalogImportUnitRow{Fila: index + 2, Codigo: cell(row, 0), Nombre: cell(row, 1), Simbolo: cell(row, 2), Tipo: cell(row, 3), FactorABase: factor, Decimales: decimals})
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

func readImportRows(file io.Reader, headers []string) ([][]string, domain.CatalogImportResult, error) {
	book, err := excelize.OpenReader(file)
	if err != nil {
		return nil, domain.CatalogImportResult{}, fmt.Errorf("%w: no se pudo leer el archivo XLSX", ErrInvalidImportFile)
	}
	defer book.Close()
	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, domain.CatalogImportResult{}, ErrInvalidImportFile
	}
	rows, err := book.GetRows(sheets[0])
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

func isEmptyRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func mergeImportResults(first, second domain.CatalogImportResult) domain.CatalogImportResult {
	return domain.CatalogImportResult{Procesadas: first.Procesadas, Creadas: second.Creadas, Omitidas: first.Omitidas + second.Omitidas, Invalidas: first.Invalidas + second.Invalidas, Errores: append(first.Errores, second.Errores...), Advertencias: append(first.Advertencias, second.Advertencias...)}
}
