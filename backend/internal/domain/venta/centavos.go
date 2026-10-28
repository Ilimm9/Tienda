package venta

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Centavos guarda importes exactos como enteros. En JSON viaja como pesos con dos
// decimales (64.00) y se convierte desde el texto, nunca mediante flotantes.
type Centavos int64

var ErrImporteInvalido = errors.New("el importe no es válido")

// Pesos devuelve el importe con dos decimales, por ejemplo "64.00" o "-0.50".
func (c Centavos) Pesos() string {
	valor := int64(c)
	signo := ""
	if valor < 0 {
		signo = "-"
		valor = -valor
	}
	return fmt.Sprintf("%s%d.%02d", signo, valor/100, valor%100)
}

func (c Centavos) MarshalJSON() ([]byte, error) { return []byte(c.Pesos()), nil }

func (c *Centavos) UnmarshalJSON(data []byte) error {
	texto := strings.Trim(strings.TrimSpace(string(data)), `"`)
	if texto == "null" {
		return nil
	}
	valor, err := ParsePesos(texto)
	if err != nil {
		return err
	}
	*c = valor
	return nil
}

// ParsePesos convierte "64", "64.5" o "64.50" a centavos; rechaza más de dos decimales y notación exponencial.
func ParsePesos(texto string) (Centavos, error) {
	texto = strings.TrimSpace(texto)
	negativo := strings.HasPrefix(texto, "-")
	texto = strings.TrimPrefix(texto, "-")
	entero, fraccion, conPunto := strings.Cut(texto, ".")
	if entero == "" || !soloDigitos(entero) || (conPunto && (fraccion == "" || len(fraccion) > 2 || !soloDigitos(fraccion))) {
		return 0, ErrImporteInvalido
	}
	for len(fraccion) < 2 {
		fraccion += "0"
	}
	pesos, err := strconv.ParseInt(entero, 10, 64)
	if err != nil || pesos > 1_000_000_000 {
		return 0, ErrImporteInvalido
	}
	centavos, _ := strconv.ParseInt(fraccion, 10, 64)
	total := Centavos(pesos*100 + centavos)
	if negativo {
		total = -total
	}
	return total, nil
}

func soloDigitos(texto string) bool {
	for _, r := range texto {
		if r < '0' || r > '9' {
			return false
		}
	}
	return texto != ""
}
