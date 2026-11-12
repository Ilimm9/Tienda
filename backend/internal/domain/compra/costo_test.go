package compra

import "testing"

func TestCostoFinalPorRenglonMixto(t *testing.T) {
	detalles := []CompraProveedorDetalle{
		{Subtotal: 240, PiezasBuenas: 12, TasaIEPS: 0.08, TasaIVA: 0.16},
		{Subtotal: 264, PiezasBuenas: 12, TasaIEPS: 0, TasaIVA: 0},
	}
	if err := CalcularImpuestosRenglon(detalles, nil); err != nil {
		t.Fatal(err)
	}
	if detalles[0].IEPS != 19.2 || detalles[0].IVA != 41.47 {
		t.Fatalf("impuestos gomilocas: %+v", detalles[0])
	}
	compra := CompraProveedor{Subtotal: 504, Impuestos: 60.67}
	if got := CostoFinal(compra, detalles[0]); got != 25.0558 {
		t.Fatalf("costo gomilocas = %v", got)
	}
	if got := CostoFinal(compra, detalles[1]); got != 22 {
		t.Fatalf("costo leche = %v", got)
	}
}

func TestCostoFinalCompraAnteriorProrratea(t *testing.T) {
	compra := CompraProveedor{Subtotal: 504, Impuestos: 60.67}
	leche := CompraProveedorDetalle{Subtotal: 264, PiezasBuenas: 12}
	// (264 + 60.67 × 264 / 504) / 12 = 24.6483
	if got := CostoFinal(compra, leche); got != 24.6483 {
		t.Fatalf("costo prorrateado = %v", got)
	}
}

func TestAjusteIEPSSoloRenglonesConIEPSYCuadra(t *testing.T) {
	detalles := []CompraProveedorDetalle{
		{Subtotal: 100, PiezasBuenas: 1, TasaIEPS: 0.08, TasaIVA: 0.16},
		{Subtotal: 500, PiezasBuenas: 1, TasaIVA: 0.16},
		{Subtotal: 200, PiezasBuenas: 1, TasaIEPS: 0.08},
	}
	ajustado := 25.01
	if err := CalcularImpuestosRenglon(detalles, &ajustado); err != nil {
		t.Fatal(err)
	}
	if detalles[1].IEPS != 0 {
		t.Fatalf("el renglón sin IEPS recibió ajuste: %v", detalles[1].IEPS)
	}
	suma := Money(detalles[0].IEPS + detalles[1].IEPS + detalles[2].IEPS)
	if suma != 25.01 {
		t.Fatalf("la suma no cuadra: %v (%+v)", suma, detalles)
	}
	if detalles[0].IEPS != 8.34 || detalles[2].IEPS != 16.67 {
		t.Fatalf("reparto inesperado: %v %v", detalles[0].IEPS, detalles[2].IEPS)
	}
}

func TestAjusteIEPSSinRenglonesConIEPSFalla(t *testing.T) {
	detalles := []CompraProveedorDetalle{{Subtotal: 100, PiezasBuenas: 1}}
	ajustado := 5.0
	if err := CalcularImpuestosRenglon(detalles, &ajustado); err == nil {
		t.Fatal("se esperaba error")
	}
	negativo := -1.0
	if err := CalcularImpuestosRenglon(detalles, &negativo); err == nil {
		t.Fatal("se esperaba error por negativo")
	}
}

func TestTasaValida(t *testing.T) {
	if !TasaValida(ImpuestoIEPS, 0.265) || TasaValida(ImpuestoIVA, 0.08) || !TasaValida(ImpuestoIVA, 0) {
		t.Fatal("catálogo de tasas inesperado")
	}
}
