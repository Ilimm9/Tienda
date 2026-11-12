package compra

import (
	"fmt"
	"strconv"
	"strings"

	compraapp "tienda/backend/internal/application/compra"
	compradomain "tienda/backend/internal/domain/compra"

	"github.com/xuri/excelize/v2"
)

// Encabezados de «hoja 1», los mismos de compras_valle.xlsx (tabla Tabla2) más los de la actualización.
var encabezadosHoja1 = []string{
	"NO. DE TIKET", "FECHA", "CODIGO", "NOMBRE DEL PRODUCTO", "PROVEEDOR", "SE LE APLICA IEPS?", "COSTO TOTAL SIN IMPUESTO",
	"COSTO TOTAL CON IMPUESTO", "CANTIDAD COMPRADA", "UNIDAD DE MEDIDA", "CANTIDAD", "TOTAL DE PIEZAS", "COSTO POR CAJA",
	"COSTO POR PIEZA", "COSTO UNITARIO EN SISTEMA", "DIFERENCIA DE COSTO", "OBSERVACIONES", "P.V EN SISTEMA", "NUEVO P.V",
	"PRODUCTO NUEVO", "NOMBRE DEL NUEVO PRODUCTO", "GRAMAJE O NO. DE PIEZAS", "PRECIO DE VENTA", "PIEZAS DAÑADAS",
	"PIEZAS FALTANTES", "PIEZAS BUENAS", "DESCUENTO", "IVA", "ESTADO DE VALIDACIÓN", "ESTADO DE LA COMPRA",
}

var encabezadosComprobacion = []string{
	"No. de ticket", "Productos registrados", "Importe sin impuestos", "IEPS", "IVA", "Importe con impuestos", "Total del documento", "Diferencia", "Nota",
}

const (
	hoja1        = "hoja 1"
	comprobacion = "COMPROBACIÓN"
)

// Excel genera el .xlsx de compras con excelize.
type Excel struct{}

func NewExcel() *Excel { return &Excel{} }

// Generar arma «hoja 1» (un renglón por producto) y, si se pide, «COMPROBACIÓN» (uno por ticket).
func (Excel) Generar(compras []compraapp.CompraExportacion, conComprobacion bool) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", hoja1); err != nil {
		return nil, err
	}
	estilos, err := nuevosEstilos(f)
	if err != nil {
		return nil, err
	}
	tickets := nombresTicket(compras)
	for i, encabezado := range encabezadosHoja1 {
		if err := f.SetCellValue(hoja1, celda(i+1, 1), encabezado); err != nil {
			return nil, err
		}
	}
	fila := 1
	for _, compra := range compras {
		for _, r := range compra.Renglones {
			fila++
			if err := escribirRenglon(f, estilos, fila, tickets[compra.ID.String()], compra, r); err != nil {
				return nil, err
			}
		}
	}
	ultima := fila
	if ultima > 1 {
		if err := f.AddTable(hoja1, &excelize.Table{Range: "A1:" + celda(len(encabezadosHoja1), ultima), Name: "Tabla2", StyleName: "TableStyleMedium15"}); err != nil {
			return nil, err
		}
	}
	_ = f.SetColWidth(hoja1, "A", "A", 22)
	_ = f.SetColWidth(hoja1, "D", "D", 36)
	_ = f.SetColWidth(hoja1, "E", "E", 22)
	_ = f.SetColWidth(hoja1, "AC", "AC", 30)
	_ = f.SetPanes(hoja1, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	if conComprobacion {
		if err := escribirComprobacion(f, estilos, compras, tickets, ultima); err != nil {
			return nil, err
		}
	}
	full := true
	_ = f.SetCalcProps(&excelize.CalcPropsOptions{FullCalcOnLoad: &full})
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type estilos struct{ fecha, dinero, dinero4 int }

func nuevosEstilos(f *excelize.File) (estilos, error) {
	fecha := "dd/mm/yyyy"
	dinero := `"$"#,##0.00`
	dinero4 := `"$"#,##0.0000`
	var e estilos
	var err error
	if e.fecha, err = f.NewStyle(&excelize.Style{CustomNumFmt: &fecha}); err != nil {
		return e, err
	}
	if e.dinero, err = f.NewStyle(&excelize.Style{CustomNumFmt: &dinero}); err != nil {
		return e, err
	}
	e.dinero4, err = f.NewStyle(&excelize.Style{CustomNumFmt: &dinero4})
	return e, err
}

// nombresTicket usa «tipo folio»; si dos compras comparten folio con distinto proveedor, agrega el proveedor
// para que SUMIF de COMPROBACIÓN no las junte.
func nombresTicket(compras []compraapp.CompraExportacion) map[string]string {
	cuenta := map[string]int{}
	for _, compra := range compras {
		cuenta[compra.Ticket]++
	}
	nombres := map[string]string{}
	for _, compra := range compras {
		nombre := compra.Ticket
		if cuenta[compra.Ticket] > 1 {
			nombre += " · " + compra.Proveedor
		}
		nombres[compra.ID.String()] = nombre
	}
	return nombres
}

func escribirRenglon(f *excelize.File, e estilos, fila int, ticket string, compra compraapp.CompraExportacion, r compraapp.RenglonExportacion) error {
	n := strconv.Itoa(fila)
	ieps := "NO"
	if r.ImpuestosPorRenglon && r.TasaIEPS > 0 {
		ieps = "SI " + porcentaje(r.TasaIEPS)
	} else if !r.ImpuestosPorRenglon && compra.IEPS > 0 {
		ieps = "SI"
	}
	nuevo, nombreNuevo := "NO", ""
	if r.ProductoNuevo {
		nuevo, nombreNuevo = "SI", r.Producto
	}
	estado := "Activa"
	if compra.Estado == compradomain.EstadoSuspendida {
		estado = "Suspendida"
	}
	valores := map[string]any{
		"A": ticket, "B": compra.FechaRecepcion, "C": r.Codigo, "D": r.Producto, "E": compra.Proveedor, "F": ieps,
		"G": r.Subtotal, "H": compradomain.Money(r.Subtotal + r.IEPS), "I": r.CantidadEmpaques, "J": r.Unidad, "K": r.PiezasPorEmpaque,
		"L": r.TotalPiezas, "N": r.CostoFinal, "T": nuevo, "U": nombreNuevo, "V": r.Contenido, "X": r.PiezasDanadas,
		"Y": r.PiezasFaltantes, "Z": r.PiezasBuenas, "AA": r.Descuento, "AB": r.IVA, "AC": r.EstadoValidacion, "AD": estado,
	}
	opcionales := map[string]*float64{"O": r.CostoAnterior, "R": r.PrecioAnterior, "S": r.PrecioNuevo, "W": r.PrecioAutorizado}
	for col, value := range opcionales {
		if value != nil {
			valores[col] = *value
		}
	}
	for col, value := range valores {
		if err := f.SetCellValue(hoja1, col+n, value); err != nil {
			return err
		}
	}
	formulas := map[string]string{
		"M": fmt.Sprintf(`IF(I%[1]s=0,"",H%[1]s/I%[1]s)`, n),
		"P": fmt.Sprintf(`IF(O%[1]s="","",N%[1]s-O%[1]s)`, n),
		"Q": fmt.Sprintf(`IF(P%[1]s="","PRIMER COSTO",IF(P%[1]s>0.01,"SUBIR EL P.V",IF(P%[1]s<-0.01,"BAJÓ EL P.C","SIN CAMBIO")))`, n),
	}
	for col, formula := range formulas {
		if err := f.SetCellFormula(hoja1, col+n, formula); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(hoja1, "B"+n, "B"+n, e.fecha); err != nil {
		return err
	}
	for _, rango := range [][2]string{{"G", "H"}, {"M", "M"}, {"R", "S"}, {"W", "W"}, {"AA", "AA"}} {
		if err := f.SetCellStyle(hoja1, rango[0]+n, rango[1]+n, e.dinero); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(hoja1, "N"+n, "P"+n, e.dinero4); err != nil {
		return err
	}
	// AB lleva el importe del IVA; el formato muestra la tasa sin volverlo texto, para que SUMIF lo sume.
	formato := `"prorrateado · $"#,##0.00`
	if r.ImpuestosPorRenglon {
		formato = `"` + porcentaje(r.TasaIVA) + ` · $"#,##0.00`
	}
	estiloIVA, err := f.NewStyle(&excelize.Style{CustomNumFmt: &formato})
	if err != nil {
		return err
	}
	return f.SetCellStyle(hoja1, "AB"+n, "AB"+n, estiloIVA)
}

func escribirComprobacion(f *excelize.File, e estilos, compras []compraapp.CompraExportacion, tickets map[string]string, ultima int) error {
	if _, err := f.NewSheet(comprobacion); err != nil {
		return err
	}
	for i, encabezado := range encabezadosComprobacion {
		if err := f.SetCellValue(comprobacion, celda(i+1, 1), encabezado); err != nil {
			return err
		}
	}
	if ultima < 2 {
		ultima = 2
	}
	rango := func(col string) string { return fmt.Sprintf("'%s'!$%s$2:$%s$%d", hoja1, col, col, ultima) }
	for i, compra := range compras {
		n := strconv.Itoa(i + 2)
		nota := []string{}
		if compra.Observaciones != "" {
			nota = append(nota, compra.Observaciones)
		}
		if compra.Descuento > 0 {
			nota = append(nota, fmt.Sprintf("Descuento $%.2f", compra.Descuento))
		}
		if compra.Estado == compradomain.EstadoSuspendida {
			nota = append(nota, "Compra suspendida")
		}
		valores := map[string]any{"A": tickets[compra.ID.String()], "G": compra.Total, "I": strings.Join(nota, " · ")}
		for col, value := range valores {
			if err := f.SetCellValue(comprobacion, col+n, value); err != nil {
				return err
			}
		}
		formulas := map[string]string{
			"B": fmt.Sprintf("COUNTIF(%s,A%s)", rango("A"), n),
			"C": fmt.Sprintf("SUMIF(%s,A%s,%s)", rango("A"), n, rango("G")),
			"D": fmt.Sprintf("SUMIF(%s,A%s,%s)-C%s", rango("A"), n, rango("H"), n),
			"E": fmt.Sprintf("SUMIF(%s,A%s,%s)", rango("A"), n, rango("AB")),
			"F": fmt.Sprintf("C%[1]s+D%[1]s+E%[1]s", n),
			"H": fmt.Sprintf("ROUND(G%[1]s-F%[1]s,2)", n),
		}
		for col, formula := range formulas {
			if err := f.SetCellFormula(comprobacion, col+n, formula); err != nil {
				return err
			}
		}
		if err := f.SetCellStyle(comprobacion, "C"+n, "H"+n, e.dinero); err != nil {
			return err
		}
	}
	if len(compras) > 0 {
		if err := f.AddTable(comprobacion, &excelize.Table{Range: "A1:" + celda(len(encabezadosComprobacion), len(compras)+1), Name: "Tabla3", StyleName: "TableStyleMedium15"}); err != nil {
			return err
		}
	}
	_ = f.SetColWidth(comprobacion, "A", "A", 26)
	_ = f.SetColWidth(comprobacion, "B", "H", 18)
	_ = f.SetColWidth(comprobacion, "I", "I", 40)
	return nil
}

func celda(col, fila int) string {
	nombre, _ := excelize.CoordinatesToCellName(col, fila)
	return nombre
}

func porcentaje(tasa float64) string {
	return strings.TrimSuffix(strings.TrimRight(strconv.FormatFloat(tasa*100, 'f', 2, 64), "0"), ".") + " %"
}
