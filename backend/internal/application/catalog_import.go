package application

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/xuri/excelize/v2"
)

var ErrInvalidImportFile = errors.New("el archivo no tiene el formato de plantilla esperado")

func (s *ProductService) ImportBrands(file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseBrandImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportBrands(rows)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportCategories(file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseCategoryImport(file)
	if err != nil {
		return result, err
	}
	imported, err := s.products.ImportCategories(rows)
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
	var output bytes.Buffer
	if err := book.Write(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
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
	return data, domain.CatalogImportResult{Procesadas: len(data), Errores: make([]domain.CatalogImportIssue, 0)}, nil
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
	return domain.CatalogImportResult{Procesadas: first.Procesadas, Creadas: second.Creadas, Omitidas: first.Omitidas + second.Omitidas, Invalidas: first.Invalidas + second.Invalidas, Errores: append(first.Errores, second.Errores...)}
}
