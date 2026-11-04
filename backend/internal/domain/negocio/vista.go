package negocio

import (
	"errors"
	"sort"
	"strings"
)

// Secciones que un negocio puede ocultar desde Configuración > Vistas (adecuación 1).
// Ocultar es una preferencia visual: no borra datos ni cambia permisos.
const (
	SeccionCategorias = "categorias"
	SeccionMarcas     = "marcas"
)

var ErrSeccionNoOcultable = errors.New("la sección no se puede ocultar")

// SeccionesOcultables lista, en orden estable, las secciones que admite Vistas.
func SeccionesOcultables() []string {
	return []string{SeccionCategorias, SeccionMarcas}
}

// VistasNegocio es la preferencia de vistas de un negocio.
type VistasNegocio struct {
	SeccionesOcultas    []string `json:"secciones_ocultas"`
	SeccionesOcultables []string `json:"secciones_ocultables"`
}

// NormalizarSeccionesOcultas valida, quita repetidos y ordena las secciones.
func NormalizarSeccionesOcultas(secciones []string) ([]string, error) {
	permitidas := make(map[string]struct{}, 2)
	for _, seccion := range SeccionesOcultables() {
		permitidas[seccion] = struct{}{}
	}
	vistas := make(map[string]struct{}, len(secciones))
	resultado := make([]string, 0, len(secciones))
	for _, seccion := range secciones {
		seccion = strings.ToLower(strings.TrimSpace(seccion))
		if _, ok := permitidas[seccion]; !ok {
			return nil, ErrSeccionNoOcultable
		}
		if _, repetida := vistas[seccion]; repetida {
			continue
		}
		vistas[seccion] = struct{}{}
		resultado = append(resultado, seccion)
	}
	sort.Strings(resultado)
	return resultado, nil
}

// LeerSeccionesOcultas convierte el valor guardado en lista; ignora valores que ya no son ocultables.
func LeerSeccionesOcultas(valor string) []string {
	resultado := make([]string, 0, 2)
	for _, parte := range strings.Split(valor, ",") {
		if normalizadas, err := NormalizarSeccionesOcultas([]string{parte}); err == nil && len(normalizadas) == 1 {
			resultado = append(resultado, normalizadas[0])
		}
	}
	normalizadas, _ := NormalizarSeccionesOcultas(resultado)
	return normalizadas
}

// EscribirSeccionesOcultas convierte la lista normalizada al valor guardado.
func EscribirSeccionesOcultas(secciones []string) string {
	return strings.Join(secciones, ",")
}
