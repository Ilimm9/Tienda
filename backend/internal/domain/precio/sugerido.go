package precio

import "math"

// PrecioSugerido redondea hacia arriba al peso el costo con su margen.
// Antes de subir se redondea a 4 decimales para que el error de punto flotante
// (10 × 1.1 = 11.000000000000002) no suba un peso de más.
func PrecioSugerido(costo, margen float64) float64 {
	return math.Ceil(math.Round(costo*(1+margen)*10000) / 10000)
}
