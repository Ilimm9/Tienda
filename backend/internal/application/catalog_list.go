package application

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
)

// MaxCatalogListRows limita cuántos renglones acepta una lista pegada.
const MaxCatalogListRows = 1000

// maxCategoryDepth evita rutas absurdas («A > B > … ») en listas y XLSX.
const maxCategoryDepth = 10

var (
	ErrCatalogListEmpty    = errors.New("la lista no tiene renglones")
	ErrCatalogListTooLarge = fmt.Errorf("la lista pasa de %d renglones; divídela o usa el archivo XLSX", MaxCatalogListRows)
)

// Las listas pegadas pasan por las mismas validaciones y el mismo importador
// transaccional que el XLSX: se omiten los renglones inválidos, repetidos o
// existentes y el resto se crea en una sola transacción.

func (s *ProductService) ImportBrandList(businessID uuid.UUID, rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error) {
	result, err := newListResult(len(rows))
	if err != nil {
		return result, err
	}
	valid := make([]domain.CatalogImportBrandRow, 0, len(rows))
	for _, row := range rows {
		if brand, issue := validateBrandRow(row); issue != nil {
			addInvalid(&result, issue)
		} else {
			valid = append(valid, brand)
		}
	}
	imported, err := s.products.ImportBrands(businessID, valid)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportCategoryList(businessID uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	result, err := newListResult(len(rows))
	if err != nil {
		return result, err
	}
	valid := make([]domain.CatalogImportCategoryRow, 0, len(rows))
	for _, row := range rows {
		if category, issue := validateCategoryPathRow(row); issue != nil {
			addInvalid(&result, issue)
		} else {
			valid = append(valid, category)
		}
	}
	imported, err := s.products.ImportCategoryPaths(businessID, valid)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportUnitList(businessID uuid.UUID, rows []domain.CatalogImportUnitRow) (domain.CatalogImportResult, error) {
	result, err := newListResult(len(rows))
	if err != nil {
		return result, err
	}
	valid := make([]domain.CatalogImportUnitRow, 0, len(rows))
	for _, row := range rows {
		if unit, issue := validateUnitRow(row); issue != nil {
			addInvalid(&result, issue)
		} else {
			valid = append(valid, unit)
		}
	}
	imported, err := s.products.ImportUnits(businessID, valid)
	return mergeImportResults(result, imported), err
}

func (s *ProductService) ImportProviderList(businessID uuid.UUID, rows []domain.CatalogImportProviderRow) (domain.CatalogImportResult, error) {
	result, err := newListResult(len(rows))
	if err != nil {
		return result, err
	}
	valid := make([]domain.CatalogImportProviderRow, 0, len(rows))
	for _, row := range rows {
		if provider, issue := validateProviderRow(row); issue != nil {
			addInvalid(&result, issue)
		} else {
			valid = append(valid, provider)
		}
	}
	imported, err := s.products.ImportProviders(businessID, valid)
	return mergeImportResults(result, imported), err
}

func newListResult(total int) (domain.CatalogImportResult, error) {
	result := domain.CatalogImportResult{Procesadas: total, Errores: make([]domain.CatalogImportIssue, 0), Advertencias: make([]domain.CatalogImportIssue, 0)}
	if total == 0 {
		return result, ErrCatalogListEmpty
	}
	if total > MaxCatalogListRows {
		return result, ErrCatalogListTooLarge
	}
	return result, nil
}

func addInvalid(result *domain.CatalogImportResult, issue *domain.CatalogImportIssue) {
	result.Invalidas++
	result.Errores = append(result.Errores, *issue)
}

func rowIssue(fila int, campo, motivo string) *domain.CatalogImportIssue {
	return &domain.CatalogImportIssue{Fila: fila, Campo: campo, Motivo: motivo}
}

func validateName(fila int, nombre string, maximo int) *domain.CatalogImportIssue {
	if nombre == "" {
		return rowIssue(fila, "Nombre", "es obligatorio")
	}
	if domain.ExcedeLongitud(nombre, maximo) {
		return rowIssue(fila, "Nombre", fmt.Sprintf("no debe pasar de %d caracteres", maximo))
	}
	return nil
}

func validateBrandRow(row domain.CatalogImportBrandRow) (domain.CatalogImportBrandRow, *domain.CatalogImportIssue) {
	row.Nombre = strings.TrimSpace(row.Nombre)
	return row, validateName(row.Fila, row.Nombre, domain.MaxNombreCatalogo)
}

// validateCategoryRow valida un renglón del XLSX; la categoría padre puede ser
// un nombre o una ruta «A > B».
func validateCategoryRow(row domain.CatalogImportCategoryRow) (domain.CatalogImportCategoryRow, *domain.CatalogImportIssue) {
	row.Nombre = strings.TrimSpace(row.Nombre)
	row.Descripcion = strings.TrimSpace(row.Descripcion)
	row.CategoriaPadre = strings.TrimSpace(row.CategoriaPadre)
	if issue := validateName(row.Fila, row.Nombre, domain.MaxNombreCatalogo); issue != nil {
		return row, issue
	}
	if row.CategoriaPadre != "" {
		padre, issue := cleanCategoryPath(row.Fila, "Categoría padre", strings.Split(row.CategoriaPadre, ">"))
		if issue != nil {
			return row, issue
		}
		row.CategoriaPadre = strings.Join(padre, " > ")
	}
	return row, nil
}

// validateCategoryPathRow valida un renglón de la lista: ruta completa desde la raíz.
func validateCategoryPathRow(row domain.CatalogImportCategoryRow) (domain.CatalogImportCategoryRow, *domain.CatalogImportIssue) {
	ruta, issue := cleanCategoryPath(row.Fila, "Nombre", row.Ruta)
	if issue != nil {
		return row, issue
	}
	row.Ruta = ruta
	row.Nombre = ruta[len(ruta)-1]
	return row, nil
}

func cleanCategoryPath(fila int, campo string, partes []string) ([]string, *domain.CatalogImportIssue) {
	if len(partes) == 0 {
		return nil, rowIssue(fila, campo, "es obligatorio")
	}
	if len(partes) > maxCategoryDepth {
		return nil, rowIssue(fila, campo, fmt.Sprintf("admite hasta %d niveles", maxCategoryDepth))
	}
	limpias := make([]string, 0, len(partes))
	for _, parte := range partes {
		parte = strings.TrimSpace(parte)
		if parte == "" {
			return nil, rowIssue(fila, campo, "falta un nombre antes o después de «>»")
		}
		if domain.ExcedeLongitud(parte, domain.MaxNombreCatalogo) {
			return nil, rowIssue(fila, campo, fmt.Sprintf("cada nombre debe tener máximo %d caracteres", domain.MaxNombreCatalogo))
		}
		limpias = append(limpias, parte)
	}
	return limpias, nil
}

func validateUnitRow(row domain.CatalogImportUnitRow) (domain.CatalogImportUnitRow, *domain.CatalogImportIssue) {
	row.Nombre = strings.TrimSpace(row.Nombre)
	row.Simbolo = strings.TrimSpace(row.Simbolo)
	if issue := validateName(row.Fila, row.Nombre, domain.MaxNombreUnidad); issue != nil {
		return row, issue
	}
	if row.Simbolo == "" {
		return row, rowIssue(row.Fila, "Símbolo", "es obligatorio")
	}
	if domain.ExcedeLongitud(row.Simbolo, domain.MaxSimboloUnidad) {
		return row, rowIssue(row.Fila, "Símbolo", fmt.Sprintf("no debe pasar de %d caracteres", domain.MaxSimboloUnidad))
	}
	tipo, err := domain.NormalizarTipoUnidad(row.Tipo)
	if err != nil {
		return row, rowIssue(row.Fila, "Tipo", "debe ser peso, volumen, longitud, área, cantidad o empaque")
	}
	row.Tipo = tipo
	if !(row.FactorABase > 0) || row.FactorABase > 999999999999 {
		return row, rowIssue(row.Fila, "Factor a base", "debe ser un número mayor a 0")
	}
	if row.Decimales < 0 || row.Decimales > 6 {
		return row, rowIssue(row.Fila, "Decimales", "debe ser un entero de 0 a 6")
	}
	return row, nil
}

func validateProviderRow(row domain.CatalogImportProviderRow) (domain.CatalogImportProviderRow, *domain.CatalogImportIssue) {
	row.Nombre = strings.TrimSpace(row.Nombre)
	row.RazonSocial = strings.TrimSpace(row.RazonSocial)
	row.Telefono = strings.TrimSpace(row.Telefono)
	row.Email = strings.TrimSpace(row.Email)
	row.Direccion = strings.TrimSpace(row.Direccion)
	if issue := validateName(row.Fila, row.Nombre, domain.MaxNombreCatalogo); issue != nil {
		return row, issue
	}
	if domain.ExcedeLongitud(row.RazonSocial, domain.MaxRazonSocial) {
		return row, rowIssue(row.Fila, "Razón social", fmt.Sprintf("no debe pasar de %d caracteres", domain.MaxRazonSocial))
	}
	rfc, err := domain.NormalizarRFC(row.RFC)
	if err != nil {
		return row, rowIssue(row.Fila, "RFC", "no tiene un formato válido")
	}
	row.RFC = rfc
	if domain.ExcedeLongitud(row.Telefono, domain.MaxTelefonoProveedor) {
		return row, rowIssue(row.Fila, "Teléfono", fmt.Sprintf("no debe pasar de %d caracteres", domain.MaxTelefonoProveedor))
	}
	if row.Email != "" {
		address, parseErr := mail.ParseAddress(row.Email)
		if parseErr != nil || address.Address != row.Email || domain.ExcedeLongitud(row.Email, domain.MaxEmailProveedor) {
			return row, rowIssue(row.Fila, "Correo", "debe ser una dirección válida")
		}
	}
	return row, nil
}
