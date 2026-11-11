package domain

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ErrCategoriaDuplicada se devuelve al crear o renombrar una categoría con un
// nombre que ya usa otra del negocio (sin distinguir mayúsculas, acentos ni espacios dobles).
var ErrCategoriaDuplicada = errors.New("ya existe una categoría con ese nombre")

var (
	ErrRFCInvalido        = errors.New("el RFC no tiene un formato válido")
	ErrTipoUnidadInvalido = errors.New("el tipo de unidad debe ser peso, volumen, longitud, área, cantidad o empaque")
)

// Longitudes máximas de las columnas de catálogos; se validan antes de insertar
// para que un renglón largo se omita en lugar de abortar la importación completa.
const (
	MaxNombreCatalogo    = 180
	MaxNombreUnidad      = 100
	MaxSimboloUnidad     = 20
	MaxTelefonoProveedor = 30
	MaxEmailProveedor    = 255
	MaxRazonSocial       = 180
)

var rfcPattern = regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`)

var tiposUnidad = map[string]struct{}{
	"PESO": {}, "VOLUMEN": {}, "LONGITUD": {}, "AREA": {}, "CANTIDAD": {}, "EMPAQUE": {},
}

var sinAcentos = strings.NewReplacer("Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U")

// NormalizarRFC recorta y pasa a mayúsculas. Vacío es válido (el RFC es opcional).
func NormalizarRFC(valor string) (string, error) {
	rfc := strings.ToUpper(strings.TrimSpace(valor))
	if rfc == "" || rfcPattern.MatchString(rfc) {
		return rfc, nil
	}
	return rfc, ErrRFCInvalido
}

// NormalizarTipoUnidad acepta «peso», «Área», «AREA»… y devuelve la clave guardada.
func NormalizarTipoUnidad(valor string) (string, error) {
	tipo := sinAcentos.Replace(strings.ToUpper(strings.TrimSpace(valor)))
	if _, ok := tiposUnidad[tipo]; ok {
		return tipo, nil
	}
	return tipo, ErrTipoUnidadInvalido
}

// ExcedeLongitud cuenta caracteres, no bytes, igual que varchar de PostgreSQL.
func ExcedeLongitud(valor string, maximo int) bool {
	return utf8.RuneCountInString(valor) > maximo
}
