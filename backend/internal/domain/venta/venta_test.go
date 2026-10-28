package venta

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDesgloseIVA(t *testing.T) {
	for _, caso := range []struct{ total, subtotal, iva Centavos }{
		{6400, 5517, 883},
		{1160, 1000, 160},
		{1, 1, 0},
		{0, 0, 0},
		{1850, 1595, 255},
	} {
		subtotal, iva := DesgloseIVA(caso.total)
		if subtotal != caso.subtotal || iva != caso.iva || subtotal+iva != caso.total {
			t.Fatalf("total %d: subtotal %d iva %d; esperado %d y %d", caso.total, subtotal, iva, caso.subtotal, caso.iva)
		}
	}
}

func TestCobro(t *testing.T) {
	recibido := Centavos(10000)
	if r, cambio, err := Cobro(PagoEfectivo, 6400, &recibido); err != nil || r != 10000 || cambio != 3600 {
		t.Fatalf("efectivo: %d %d %v", r, cambio, err)
	}
	exacto := Centavos(6400)
	if _, cambio, err := Cobro(PagoEfectivo, 6400, &exacto); err != nil || cambio != 0 {
		t.Fatalf("efectivo exacto: %d %v", cambio, err)
	}
	poco := Centavos(6399)
	if _, _, err := Cobro(PagoEfectivo, 6400, &poco); !errors.Is(err, ErrPagoInsuficiente) {
		t.Fatalf("recibido menor al total debe fallar: %v", err)
	}
	if _, _, err := Cobro(PagoEfectivo, 6400, nil); !errors.Is(err, ErrPagoInsuficiente) {
		t.Fatalf("efectivo sin recibido debe fallar: %v", err)
	}
	for _, forma := range []string{PagoTarjeta, PagoTransferencia} {
		if r, cambio, err := Cobro(forma, 6400, &poco); err != nil || r != 6400 || cambio != 0 {
			t.Fatalf("%s cobra el total exacto: %d %d %v", forma, r, cambio, err)
		}
	}
	if _, _, err := Cobro("CREDITO", 6400, nil); !errors.Is(err, ErrFormaPagoInvalida) {
		t.Fatalf("crédito queda fuera de alcance: %v", err)
	}
}

func TestMargenYNivel(t *testing.T) {
	if margen, ok := Margen(2000, 1500); !ok || margen != 0.25 {
		t.Fatalf("margen 25 %%: %v %v", margen, ok)
	}
	if margen, ok := Margen(1000, 1200); !ok || margen != -0.2 {
		t.Fatalf("margen negativo: %v", margen)
	}
	if margen, _ := Margen(300, 100); margen != 0.6667 {
		t.Fatalf("margen redondeado a 4 decimales: %v", margen)
	}
	if _, ok := Margen(0, 100); ok {
		t.Fatal("sin precio no hay margen")
	}
	for _, caso := range []struct {
		precio, costo Centavos
		nivel         string
	}{
		{1000, 1200, NivelMargenNegativo},
		{1000, 900, NivelMargenBajo},
		{1000, 801, NivelMargenBajo},
		{1000, 800, NivelMargenOK},
		{1000, 1000, NivelMargenBajo},
	} {
		if nivel := NivelMargen(caso.precio, caso.costo); nivel != caso.nivel {
			t.Fatalf("precio %d costo %d: %s, esperado %s", caso.precio, caso.costo, nivel, caso.nivel)
		}
	}
}

func TestCentavosJSON(t *testing.T) {
	for texto, esperado := range map[string]Centavos{`64`: 6400, `64.5`: 6450, `64.50`: 6450, `"0.07"`: 7, `0`: 0, `1000000`: 100000000} {
		var valor Centavos
		if err := json.Unmarshal([]byte(texto), &valor); err != nil || valor != esperado {
			t.Fatalf("%s: %d %v", texto, valor, err)
		}
	}
	for _, texto := range []string{`64.505`, `1e3`, `"abc"`, `.5`, `64.`, `"--1"`} {
		var valor Centavos
		if err := json.Unmarshal([]byte(texto), &valor); err == nil {
			t.Fatalf("%s debe rechazarse; quedó %d", texto, valor)
		}
	}
	salida, _ := json.Marshal(struct{ Total, Cambio Centavos }{6400, -50})
	if string(salida) != `{"Total":64.00,"Cambio":-0.50}` {
		t.Fatalf("JSON: %s", salida)
	}
	if FormatoFolio(232) != "V-000232" {
		t.Fatalf("folio: %s", FormatoFolio(232))
	}
}
