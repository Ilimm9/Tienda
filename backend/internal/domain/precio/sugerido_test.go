package precio

import "testing"

func TestPrecioSugerido(t *testing.T) {
	casos := []struct{ costo, margen, esperado float64 }{
		{10, 0.10, 11},
		{25.0558, 0.30, 33},
		{20, 0.30, 26},
		{22, 0.30, 29},
		{0, 0.30, 0},
	}
	for _, caso := range casos {
		if got := PrecioSugerido(caso.costo, caso.margen); got != caso.esperado {
			t.Fatalf("PrecioSugerido(%v, %v) = %v; esperado %v", caso.costo, caso.margen, got, caso.esperado)
		}
	}
}
