package compra

import (
	"errors"
	"math"
)

// Tipos de impuesto del catálogo `impuestos`.
const (
	ImpuestoIVA  = "IVA"
	ImpuestoIEPS = "IEPS"
)

// TasaImpuesto es una tasa del catálogo; la tasa es fracción (0.16 = 16 %).
type TasaImpuesto struct {
	Nombre string
	Tipo   string
	Tasa   float64
}

// CatalogoTasas son las tasas que se ofrecen al capturar una compra.
var CatalogoTasas = []TasaImpuesto{
	{Nombre: "IVA 16 %", Tipo: ImpuestoIVA, Tasa: 0.16},
	{Nombre: "IVA 0 %", Tipo: ImpuestoIVA, Tasa: 0},
	{Nombre: "Sin IEPS", Tipo: ImpuestoIEPS, Tasa: 0},
	{Nombre: "IEPS 8 %", Tipo: ImpuestoIEPS, Tasa: 0.08},
	{Nombre: "IEPS 25 %", Tipo: ImpuestoIEPS, Tasa: 0.25},
	{Nombre: "IEPS 26.5 %", Tipo: ImpuestoIEPS, Tasa: 0.265},
	{Nombre: "IEPS 30 %", Tipo: ImpuestoIEPS, Tasa: 0.30},
	{Nombre: "IEPS 53 %", Tipo: ImpuestoIEPS, Tasa: 0.53},
	{Nombre: "IEPS 160 %", Tipo: ImpuestoIEPS, Tasa: 1.60},
}

// TasaValida indica si la tasa existe en el catálogo para ese tipo.
func TasaValida(tipo string, tasa float64) bool {
	for _, item := range CatalogoTasas {
		if item.Tipo == tipo && math.Abs(item.Tasa-tasa) < 1e-9 {
			return true
		}
	}
	return false
}

// CostoFinal es el costo por pieza buena con impuestos que entra al inventario y a la propuesta.
// Renglón con impuestos propios: (subtotal + IEPS + IVA) ÷ piezas buenas.
// Compra anterior: los impuestos de la compra se prorratean por subtotal, como siempre.
func CostoFinal(compra CompraProveedor, detalle CompraProveedorDetalle) float64 {
	if detalle.PiezasBuenas <= 0 {
		return 0
	}
	if detalle.ImpuestosPorRenglon {
		return Money4((detalle.Subtotal + detalle.IEPS + detalle.IVA) / float64(detalle.PiezasBuenas))
	}
	prorrateo := 0.0
	if compra.Subtotal > 0 {
		prorrateo = compra.Impuestos * detalle.Subtotal / compra.Subtotal
	}
	return Money4((detalle.Subtotal + prorrateo) / float64(detalle.PiezasBuenas))
}

// CalcularImpuestosRenglon llena IEPS e IVA de cada renglón con sus tasas. Si `iepsAjustado`
// viene y difiere de la suma calculada, la diferencia se reparte entre los renglones con IEPS,
// proporcional a su subtotal; el último absorbe los centavos para que la suma cuadre exacta.
func CalcularImpuestosRenglon(detalles []CompraProveedorDetalle, iepsAjustado *float64) error {
	calculado := 0.0
	baseIEPS := 0.0
	ultimo := -1
	for i := range detalles {
		detalles[i].ImpuestosPorRenglon = true
		detalles[i].IEPS = Money(detalles[i].Subtotal * detalles[i].TasaIEPS)
		calculado += detalles[i].IEPS
		if detalles[i].TasaIEPS > 0 {
			baseIEPS += detalles[i].Subtotal
			ultimo = i
		}
	}
	if iepsAjustado != nil {
		ajustado := *iepsAjustado
		if ajustado < 0 || math.IsNaN(ajustado) || math.IsInf(ajustado, 0) {
			return errors.New("el importe de IEPS no es válido")
		}
		diferencia := Money(Money(ajustado) - Money(calculado))
		if diferencia != 0 {
			if ultimo < 0 || baseIEPS <= 0 {
				return errors.New("para ajustar el IEPS del ticket marca al menos un producto con IEPS")
			}
			repartido := 0.0
			for i := range detalles {
				if detalles[i].TasaIEPS <= 0 {
					continue
				}
				parte := Money(diferencia * detalles[i].Subtotal / baseIEPS)
				if i == ultimo {
					parte = Money(diferencia - repartido)
				}
				repartido = Money(repartido + parte)
				detalles[i].IEPS = Money(detalles[i].IEPS + parte)
				if detalles[i].IEPS < 0 {
					return errors.New("el ajuste de IEPS deja un producto con IEPS negativo")
				}
			}
		}
	}
	for i := range detalles {
		detalles[i].IVA = Money((detalles[i].Subtotal + detalles[i].IEPS) * detalles[i].TasaIVA)
	}
	return nil
}

// CompletarImpuestosRenglon llena IEPS e IVA del renglón en compras anteriores con el prorrateo,
// solo para mostrarlos; no se guarda.
func (c *CompraProveedor) CompletarImpuestosRenglon() {
	for i := range c.Detalles {
		d := &c.Detalles[i]
		if d.ImpuestosPorRenglon || c.Subtotal <= 0 {
			continue
		}
		d.IEPS = Money(c.IEPS * d.Subtotal / c.Subtotal)
		d.IVA = Money(c.IVA * d.Subtotal / c.Subtotal)
	}
}

func Money(value float64) float64  { return math.Round(value*100) / 100 }
func Money4(value float64) float64 { return math.Round(value*10000) / 10000 }
