package database

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"math"
	"strconv"
	"os"
	"strings"
	"testing"

	compraapp "tienda/backend/internal/application/compra"
	negocioapp "tienda/backend/internal/application/negocio"
	precioapp "tienda/backend/internal/application/precio"
	"tienda/backend/internal/domain"
	compradomain "tienda/backend/internal/domain/compra"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	preciodomain "tienda/backend/internal/domain/precio"
	"tienda/backend/internal/infrastructure"
	comprainfra "tienda/backend/internal/infrastructure/compra"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"
	precioinfra "tienda/backend/internal/infrastructure/precio"
	negociohttp "tienda/backend/internal/interfaces/http/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// comprasFixture prepara un negocio con sucursal, proveedor, caja de 12 y tres productos
// (Gomilocas, Leche y Arroz) para las pruebas de la actualización de compras.
type comprasFixture struct {
	db        *gorm.DB
	compras   *comprainfra.Repository
	precios   *precioinfra.Repository
	negocio   negociodomain.Negocio
	usuario   cuentadomain.Usuario
	sucursal  negociodomain.Sucursal
	proveedor domain.Proveedor
	caja      domain.UnidadMedida
	productos map[string]uuid.UUID
	folio     int
}

func nuevoComprasFixture(t *testing.T) *comprasFixture {
	t.Helper()
	url := os.Getenv("SECURITY_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SECURITY_TEST_DATABASE_URL no está configurada")
	}
	if !strings.Contains(url, "tienda_security_test") {
		t.Fatal("la prueba destructiva sólo puede usar una base llamada tienda_security_test")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{`DROP SCHEMA public CASCADE`, `CREATE SCHEMA public`} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := Init(db); err != nil {
		t.Fatal(err)
	}
	f := &comprasFixture{db: db, compras: comprainfra.NewRepository(db), precios: precioinfra.NewRepository(db), productos: map[string]uuid.UUID{}}
	products := infrastructure.NewProductRepository(db)
	f.negocio = negociodomain.Negocio{ID: uuid.New(), Slug: "valle", NombreComercial: "Valle", Nombre: "Valle", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	f.usuario = cuentadomain.Usuario{ID: uuid.New(), Correo: "valle@example.com", HashContrasena: "no-usada-en-prueba", Estado: "activo"}
	f.sucursal = negociodomain.Sucursal{ID: uuid.New(), NegocioID: f.negocio.ID, Codigo: "SUC-001", Nombre: "Oaxaca", EsPrincipal: true, Activo: true}
	f.caja = domain.UnidadMedida{ID: uuid.New(), NegocioID: f.negocio.ID, Nombre: "Caja 12", Simbolo: "CJ12", Tipo: "EMPAQUE", FactorABase: 12, Activo: true}
	f.proveedor = domain.Proveedor{ID: uuid.New(), NegocioID: f.negocio.ID, Nombre: "Depósito del Valle", Activo: true}
	for _, item := range []any{&f.negocio, &f.usuario, &f.sucursal, &f.caja, &f.proveedor} {
		if err := db.Create(item).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := products.CreateCategory(f.negocio.ID, domain.CreateCategoriaInput{Nombre: "General"}); err != nil {
		t.Fatal(err)
	}
	categorias, _ := products.ListCategoriesAdmin(f.negocio.ID)
	categoria := categorias[0].ID
	rows := []domain.ValidatedProductImportRow{
		{Fila: 2, Input: domain.CreateImportedProductInput{Nombre: "Gomilocas", CategoriaID: &categoria, SucursalID: f.sucursal.ID, PrecioVenta: 28, CostoProveedor: 19}},
		{Fila: 3, Input: domain.CreateImportedProductInput{Nombre: "Leche", CategoriaID: &categoria, SucursalID: f.sucursal.ID, PrecioVenta: 27, CostoProveedor: 21}},
		{Fila: 4, Input: domain.CreateImportedProductInput{Nombre: "Arroz", CategoriaID: &categoria, SucursalID: f.sucursal.ID, PrecioVenta: 30, CostoProveedor: 20}},
	}
	if res, err := products.CreateImportedProducts(f.negocio.ID, rows); err != nil || res.Creadas != 3 {
		t.Fatalf("importación inicial: %+v %v", res, err)
	}
	for _, nombre := range []string{"Gomilocas", "Leche", "Arroz"} {
		var row struct{ ID uuid.UUID }
		if err := db.Raw(`SELECT pn.id FROM producto_negocio pn JOIN productos p ON p.id = pn.producto_id WHERE pn.negocio_id = ? AND p.nombre = ?`, f.negocio.ID, nombre).Scan(&row).Error; err != nil || row.ID == uuid.Nil {
			t.Fatalf("producto %s: %v", nombre, err)
		}
		f.productos[nombre] = row.ID
	}
	return f
}

func tasa(v float64) *float64 { return &v }

func (f *comprasFixture) renglon(nombre string, costoCaja float64, ieps, iva *float64) compraapp.CrearDetalleInput {
	return compraapp.CrearDetalleInput{ProductoNegocioID: f.productos[nombre].String(), UnidadMedidaID: f.caja.ID.String(), CantidadEmpaques: 1, CostoEmpaque: costoCaja, TasaIEPS: ieps, TasaIVA: iva}
}

func (f *comprasFixture) entrada(detalles ...compraapp.CrearDetalleInput) compraapp.CrearCompraInput {
	f.folio++
	return compraapp.CrearCompraInput{ProveedorID: f.proveedor.ID.String(), SucursalID: f.sucursal.ID.String(), TipoDocumento: "FACTURA",
		FolioDocumento: "PRUEBA-" + string(rune('A'+f.folio)), FechaDocumento: "2026-11-01", FechaRecepcion: "2026-11-03", FormaPago: "EFECTIVO", Detalles: detalles}
}

func (f *comprasFixture) comprar(t *testing.T, input compraapp.CrearCompraInput) compradomain.CompraProveedor {
	t.Helper()
	compra, err := f.compras.Create(f.negocio.ID, f.usuario.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	return compra
}

func (f *comprasFixture) propuesta(t *testing.T, detalleID uuid.UUID) preciodomain.Propuesta {
	t.Helper()
	var p preciodomain.Propuesta
	if err := f.db.Where("compra_detalle_id = ?", detalleID).First(&p).Error; err != nil {
		t.Fatal(err)
	}
	return p
}

func (f *comprasFixture) stock(t *testing.T, nombre string) float64 {
	t.Helper()
	var stock float64
	if err := f.db.Raw(`SELECT COALESCE(SUM(stock_actual), 0) FROM inventario_sucursal WHERE sucursal_id = ? AND producto_negocio_id = ?`, f.sucursal.ID, f.productos[nombre]).Scan(&stock).Error; err != nil {
		t.Fatal(err)
	}
	return stock
}

func (f *comprasFixture) pendientes(t *testing.T, filtro precioapp.FiltroPendientes) []precioapp.PropuestaResumen {
	t.Helper()
	items, err := f.precios.ListPending(f.negocio.ID, f.sucursal.ID, filtro)
	if err != nil {
		t.Fatal(err)
	}
	return items
}

func cerca(a, b float64) bool { return math.Abs(a-b) < 0.00005 }

func TestComprasImpuestosPorRenglon(t *testing.T) {
	f := nuevoComprasFixture(t)

	t.Run("IEPS e IVA mixtos dan el costo de cada producto", func(t *testing.T) {
		compra := f.comprar(t, f.entrada(f.renglon("Gomilocas", 240, tasa(0.08), tasa(0.16)), f.renglon("Leche", 264, tasa(0), tasa(0))))
		if compra.IEPS != 19.2 || compra.IVA != 41.47 || compra.Total != 564.67 {
			t.Fatalf("totales = ieps %v iva %v total %v", compra.IEPS, compra.IVA, compra.Total)
		}
		gomilocas := f.propuesta(t, compra.Detalles[0].ID)
		leche := f.propuesta(t, compra.Detalles[1].ID)
		if !cerca(gomilocas.CostoCapturado, 25.0558) || !cerca(leche.CostoCapturado, 22) {
			t.Fatalf("costos = %v / %v", gomilocas.CostoCapturado, leche.CostoCapturado)
		}
		if gomilocas.PrecioSugerido != preciodomain.PrecioSugerido(25.0558, gomilocas.MargenSugerido) {
			t.Fatalf("precio sugerido = %v", gomilocas.PrecioSugerido)
		}
		var movimiento domain.MovimientoInventario
		f.db.Where("producto_negocio_id = ? AND tipo = 'COMPRA_ENTRADA'", f.productos["Leche"]).First(&movimiento)
		if movimiento.CostoUnitario == nil || !cerca(*movimiento.CostoUnitario, 22) {
			t.Fatalf("costo del movimiento = %v", movimiento.CostoUnitario)
		}
		for _, item := range f.pendientes(t, precioapp.FiltroPendientes{}) {
			if item.CompraDetalleID == compra.Detalles[1].ID && !cerca(item.CostoCapturado, 22) {
				t.Fatalf("validar costos muestra %v para la leche", item.CostoCapturado)
			}
		}
	})

	t.Run("las tasas usadas se recuerdan y se actualizan", func(t *testing.T) {
		items, err := f.compras.ListProductos(f.negocio.ID, f.proveedor.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, item := range items {
			if item.ID == f.productos["Gomilocas"] && (item.TasaIEPS == nil || *item.TasaIEPS != 0.08 || item.TasaIVA == nil || *item.TasaIVA != 0.16) {
				t.Fatalf("tasas de gomilocas = %v %v", item.TasaIEPS, item.TasaIVA)
			}
			if item.ID == f.productos["Arroz"] && (item.TasaIEPS != nil || item.TasaIVA != nil) {
				t.Fatal("el arroz no tiene tasas guardadas")
			}
		}
		f.comprar(t, f.entrada(f.renglon("Gomilocas", 240, tasa(0.25), tasa(0.16))))
		var count int64
		f.db.Model(&domain.ProductoImpuesto{}).Where("producto_negocio_id = ?", f.productos["Gomilocas"]).Count(&count)
		if count != 2 {
			t.Fatalf("producto_impuestos de gomilocas = %d", count)
		}
		items, _ = f.compras.ListProductos(f.negocio.ID, f.proveedor.ID)
		for _, item := range items {
			if item.ID == f.productos["Gomilocas"] && *item.TasaIEPS != 0.25 {
				t.Fatalf("tasa recordada = %v", *item.TasaIEPS)
			}
		}
	})

	t.Run("el ajuste de IEPS solo va a renglones con IEPS y cuadra", func(t *testing.T) {
		input := f.entrada(f.renglon("Gomilocas", 100, tasa(0.08), tasa(0.16)), f.renglon("Leche", 500, tasa(0), tasa(0.16)), f.renglon("Arroz", 200, tasa(0.08), tasa(0)))
		input.IEPSAjustado = tasa(25.01)
		compra := f.comprar(t, input)
		if compra.IEPS != 25.01 || compra.Detalles[1].IEPS != 0 {
			t.Fatalf("ieps = %v, leche %v", compra.IEPS, compra.Detalles[1].IEPS)
		}
	})

	t.Run("una tasa fuera del catálogo se rechaza", func(t *testing.T) {
		if _, err := f.compras.Create(f.negocio.ID, f.usuario.ID, f.entrada(f.renglon("Arroz", 100, tasa(0.07), nil))); err == nil {
			t.Fatal("se aceptó IEPS 7 %")
		}
	})

	t.Run("cliente anterior sin tasas prorratea como antes", func(t *testing.T) {
		input := f.entrada(f.renglon("Gomilocas", 240, nil, nil), f.renglon("Leche", 264, nil, nil))
		input.IncluyeIVA = true
		compra := f.comprar(t, input)
		if compra.Detalles[0].ImpuestosPorRenglon || compra.IVA != 80.64 {
			t.Fatalf("compra anterior = %+v", compra)
		}
		leche := f.propuesta(t, compra.Detalles[1].ID)
		// (264 + 80.64 × 264 / 504) / 12 = 25.52
		if !cerca(leche.CostoCapturado, 25.52) {
			t.Fatalf("costo prorrateado = %v", leche.CostoCapturado)
		}
		svc := compraapp.NewService(f.compras)
		leida, err := svc.Get(f.negocio.ID, compra.ID)
		if err != nil {
			t.Fatal(err)
		}
		if leida.Detalles[0].IVA+leida.Detalles[1].IVA != 80.64 {
			t.Fatalf("IVA mostrado en renglones = %v + %v", leida.Detalles[0].IVA, leida.Detalles[1].IVA)
		}
	})
}

func TestValidarCostosPorCompra(t *testing.T) {
	f := nuevoComprasFixture(t)
	otro := domain.Proveedor{ID: uuid.New(), NegocioID: f.negocio.ID, Nombre: "Ricolino", Activo: true}
	if err := f.db.Create(&otro).Error; err != nil {
		t.Fatal(err)
	}
	primera := f.comprar(t, f.entrada(f.renglon("Gomilocas", 240, tasa(0.08), tasa(0.16)), f.renglon("Leche", 264, tasa(0), tasa(0))))
	segunda := f.entrada(f.renglon("Arroz", 120, tasa(0), tasa(0.16)))
	segunda.ProveedorID = otro.ID.String()
	segunda.FechaRecepcion = "2026-11-08"
	ricolino := f.comprar(t, segunda)

	t.Run("cada propuesta trae su compra, proveedor y nota", func(t *testing.T) {
		items := f.pendientes(t, precioapp.FiltroPendientes{})
		if len(items) != 3 {
			t.Fatalf("pendientes = %d", len(items))
		}
		if items[0].CompraID != ricolino.ID || items[0].Proveedor != "Ricolino" {
			t.Fatalf("la compra más reciente va primero: %+v", items[0])
		}
		for _, item := range items {
			if item.CompraID == primera.ID && item.Producto == "Gomilocas" {
				if item.FolioDocumento != primera.FolioDocumento || item.TotalCompra != primera.Total || !strings.Contains(item.Nota, "IEPS 8 %") || !strings.Contains(item.Nota, "1 × Caja 12") {
					t.Fatalf("renglón de gomilocas = %+v", item)
				}
				if item.PrecioSugerido != preciodomain.PrecioSugerido(item.CostoCapturado, item.MargenSugerido) {
					t.Fatalf("precio sugerido inconsistente: %+v", item)
				}
			}
		}
	})

	t.Run("filtra por proveedor, fechas y texto", func(t *testing.T) {
		if items := f.pendientes(t, precioapp.FiltroPendientes{ProveedorID: &otro.ID}); len(items) != 1 {
			t.Fatalf("por proveedor = %d", len(items))
		}
		desde := ricolino.FechaRecepcion
		if items := f.pendientes(t, precioapp.FiltroPendientes{Desde: &desde}); len(items) != 1 {
			t.Fatalf("desde = %d", len(items))
		}
		hasta := primera.FechaRecepcion
		if items := f.pendientes(t, precioapp.FiltroPendientes{Hasta: &hasta}); len(items) != 2 {
			t.Fatalf("hasta = %d", len(items))
		}
		if items := f.pendientes(t, precioapp.FiltroPendientes{Texto: "lech"}); len(items) != 1 {
			t.Fatalf("texto producto = %d", len(items))
		}
		if items := f.pendientes(t, precioapp.FiltroPendientes{Texto: primera.FolioDocumento}); len(items) != 2 {
			t.Fatalf("texto folio = %d", len(items))
		}
	})

	t.Run("validar ticket autoriza todo o nada", func(t *testing.T) {
		items := f.pendientes(t, precioapp.FiltroPendientes{})
		var renglones []precioapp.AutorizarRenglonInput
		for _, item := range items {
			if item.CompraID == primera.ID {
				renglones = append(renglones, precioapp.AutorizarRenglonInput{PropuestaID: item.ID, AutorizarInput: precioapp.AutorizarInput{Costo: item.CostoCapturado, Margen: item.MargenSugerido, PrecioVenta: item.PrecioSugerido}})
			}
		}
		ajeno := append([]precioapp.AutorizarRenglonInput{}, renglones...)
		for _, item := range items {
			if item.CompraID == ricolino.ID {
				ajeno = append(ajeno, precioapp.AutorizarRenglonInput{PropuestaID: item.ID, AutorizarInput: precioapp.AutorizarInput{Costo: 1, PrecioVenta: 2}})
			}
		}
		if _, err := f.precios.AuthorizePurchase(f.negocio.ID, f.usuario.ID, primera.ID, precioapp.AutorizarCompraInput{Renglones: ajeno}); err == nil {
			t.Fatal("se aceptó un renglón de otra compra")
		}
		if len(f.pendientes(t, precioapp.FiltroPendientes{})) != 3 {
			t.Fatal("el rechazo dejó renglones autorizados")
		}
		autorizadas, err := f.precios.AuthorizePurchase(f.negocio.ID, f.usuario.ID, primera.ID, precioapp.AutorizarCompraInput{Renglones: renglones})
		if err != nil || len(autorizadas) != 2 {
			t.Fatalf("validar ticket: %v %d", err, len(autorizadas))
		}
		var precio preciodomain.PrecioSucursal
		f.db.Where("sucursal_id = ? AND producto_negocio_id = ?", f.sucursal.ID, f.productos["Leche"]).First(&precio)
		if !cerca(precio.CostoVigente, 22) || precio.PrecioVenta != math.Ceil(precio.PrecioVenta) {
			t.Fatalf("precio vigente de la leche = %+v", precio)
		}
		if len(f.pendientes(t, precioapp.FiltroPendientes{})) != 1 {
			t.Fatal("quedaron pendientes del ticket validado")
		}
	})
}

func leerExcel(t *testing.T, archivo []byte) *excelize.File {
	t.Helper()
	f, err := excelize.OpenReader(bytes.NewReader(archivo))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestComprasExcel(t *testing.T) {
	f := nuevoComprasFixture(t)
	servicio := compraapp.NewService(f.compras).ConExportador(comprainfra.NewExcel())
	otro := domain.Proveedor{ID: uuid.New(), NegocioID: f.negocio.ID, Nombre: "Ricolino", Activo: true}
	if err := f.db.Create(&otro).Error; err != nil {
		t.Fatal(err)
	}
	mixta := f.comprar(t, f.entrada(f.renglon("Gomilocas", 240, tasa(0.08), tasa(0.16)), f.renglon("Leche", 264, tasa(0), tasa(0))))
	anterior := f.entrada(f.renglon("Arroz", 120, nil, nil))
	anterior.IncluyeIVA = true
	anterior.ProveedorID = otro.ID.String()
	anterior.FechaRecepcion = "2026-11-09"
	vieja := f.comprar(t, anterior)

	exportar := func(filtro compraapp.FiltroExportacion) *excelize.File {
		t.Helper()
		filtro.SucursalID = f.sucursal.ID
		archivo, _, err := servicio.Exportar(f.negocio.ID, filtro)
		if err != nil {
			t.Fatal(err)
		}
		return leerExcel(t, archivo)
	}
	filas := func(x *excelize.File) [][]string {
		rows, err := x.GetRows("hoja 1")
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}

	t.Run("hoja 1 con encabezados, fecha de recepción e impuestos", func(t *testing.T) {
		x := exportar(compraapp.FiltroExportacion{IncluirComprobacion: true})
		rows := filas(x)
		if len(rows) != 4 || rows[0][0] != "NO. DE TIKET" || rows[0][29] != "ESTADO DE LA COMPRA" {
			t.Fatalf("encabezados/filas = %d %v", len(rows), rows[0])
		}
		if rows[1][0] != "FACTURA "+mixta.FolioDocumento || rows[1][1] != "03/11/2026" || rows[1][5] != "SI 8 %" || rows[2][5] != "NO" {
			t.Fatalf("renglón de gomilocas = %v", rows[1])
		}
		if rows[3][5] != "NO" || rows[3][29] != "Activa" {
			t.Fatalf("renglón anterior = %v", rows[3])
		}
		for fila := 2; fila <= 3; fila++ {
			diferencia, err := x.CalcCellValue("COMPROBACIÓN", "H"+strconv.Itoa(fila), excelize.Options{RawCellValue: true})
			if err != nil || (diferencia != "0" && diferencia != "-0") {
				t.Fatalf("COMPROBACIÓN fila %d: diferencia %q %v", fila, diferencia, err)
			}
		}
		if productos, _ := x.CalcCellValue("COMPROBACIÓN", "B2"); productos != "2" {
			t.Fatalf("productos registrados = %q", productos)
		}
	})

	t.Run("filtra por proveedor, fechas y compra", func(t *testing.T) {
		if rows := filas(exportar(compraapp.FiltroExportacion{ProveedorID: &otro.ID})); len(rows) != 2 {
			t.Fatalf("por proveedor = %d", len(rows))
		}
		desde := vieja.FechaRecepcion
		if rows := filas(exportar(compraapp.FiltroExportacion{Desde: &desde})); len(rows) != 2 {
			t.Fatalf("desde = %d", len(rows))
		}
		if rows := filas(exportar(compraapp.FiltroExportacion{CompraIDs: []uuid.UUID{mixta.ID}})); len(rows) != 3 {
			t.Fatalf("por compra = %d", len(rows))
		}
		if hojas := exportar(compraapp.FiltroExportacion{}).GetSheetList(); len(hojas) != 1 {
			t.Fatalf("sin COMPROBACIÓN = %v", hojas)
		}
	})

	t.Run("O y R son la foto de la propuesta y no cambian al validar", func(t *testing.T) {
		antes := filas(exportar(compraapp.FiltroExportacion{CompraIDs: []uuid.UUID{mixta.ID}}))
		p := f.propuesta(t, mixta.Detalles[0].ID)
		if _, err := f.precios.Authorize(f.negocio.ID, f.usuario.ID, p.ID, precioapp.AutorizarInput{Costo: p.CostoCapturado, Margen: 0.3, PrecioVenta: 40}); err != nil {
			t.Fatal(err)
		}
		despues := filas(exportar(compraapp.FiltroExportacion{CompraIDs: []uuid.UUID{mixta.ID}}))
		if antes[1][14] != despues[1][14] || antes[1][17] != despues[1][17] {
			t.Fatalf("O/R cambiaron: %v/%v → %v/%v", antes[1][14], antes[1][17], despues[1][14], despues[1][17])
		}
		if despues[1][18] != "$40.00" || !strings.HasPrefix(despues[1][28], "Validado el") || antes[1][28] != "Pendiente" {
			t.Fatalf("validación en Excel = %v / %v", despues[1][18], despues[1][28])
		}
	})
}

func (f *comprasFixture) fijarStock(t *testing.T, nombre string, stock float64) {
	t.Helper()
	if err := f.db.Exec(`UPDATE inventario_sucursal SET stock_actual = ? WHERE sucursal_id = ? AND producto_negocio_id = ?`, stock, f.sucursal.ID, f.productos[nombre]).Error; err != nil {
		t.Fatal(err)
	}
}

func edicion(compra compradomain.CompraProveedor) compraapp.EditarCompraInput {
	return compraapp.EditarCompraInput{TipoDocumento: compra.TipoDocumento, FolioDocumento: compra.FolioDocumento, FechaDocumento: compra.FechaDocumento.Format("2006-01-02"),
		FechaRecepcion: compra.FechaRecepcion.Format("2006-01-02"), FormaPago: compra.FormaPago}
}

func renglonEditado(d compradomain.CompraProveedorDetalle, empaques int, costo float64) compraapp.EditarDetalleInput {
	id := d.ID.String()
	ieps, iva := d.TasaIEPS, d.TasaIVA
	return compraapp.EditarDetalleInput{ID: &id, CrearDetalleInput: compraapp.CrearDetalleInput{ProductoNegocioID: d.ProductoNegocioID.String(), UnidadMedidaID: d.UnidadMedidaID.String(),
		CantidadEmpaques: empaques, CostoEmpaque: costo, NumeroLote: d.NumeroLote, TasaIEPS: &ieps, TasaIVA: &iva}}
}

func TestComprasEditarSuspenderReactivar(t *testing.T) {
	f := nuevoComprasFixture(t)
	lote := "L-1"

	t.Run("editar el encabezado respeta el folio único", func(t *testing.T) {
		primera := f.comprar(t, f.entrada(f.renglon("Arroz", 120, tasa(0), tasa(0))))
		segunda := f.comprar(t, f.entrada(f.renglon("Arroz", 120, tasa(0), tasa(0))))
		input := edicion(segunda)
		input.FolioDocumento = primera.FolioDocumento
		if _, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, segunda.ID, input); err == nil || !strings.Contains(err.Error(), "ya está registrado") {
			t.Fatalf("folio duplicado: %v", err)
		}
		input.FolioDocumento = "CORREGIDO-1"
		input.Motivo = tasa2("El folio se capturó mal")
		editada, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, segunda.ID, input)
		if err != nil || editada.FolioDocumento != "CORREGIDO-1" || editada.EditadoPorUsuarioID == nil {
			t.Fatalf("editar encabezado: %v %+v", err, editada)
		}
		var eventos int64
		f.db.Model(&compradomain.CompraProveedorEvento{}).Where("compra_id = ? AND tipo = ?", segunda.ID, compradomain.EventoEdicion).Count(&eventos)
		if eventos != 1 {
			t.Fatalf("eventos de edición = %d", eventos)
		}
	})

	t.Run("cambiar el costo actualiza propuesta, precio sugerido, lote y stock", func(t *testing.T) {
		renglon := f.renglon("Gomilocas", 240, tasa(0.08), tasa(0.16))
		renglon.NumeroLote = &lote
		compra := f.comprar(t, f.entrada(renglon, f.renglon("Leche", 264, tasa(0), tasa(0))))
		stockAntes := f.stock(t, "Gomilocas")
		input := edicion(compra)
		input.Detalles = []compraapp.EditarDetalleInput{renglonEditado(compra.Detalles[0], 2, 300), renglonEditado(compra.Detalles[1], 1, 264)}
		editada, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, compra.ID, input)
		if err != nil {
			t.Fatal(err)
		}
		if f.stock(t, "Gomilocas") != stockAntes+12 {
			t.Fatalf("stock = %v, antes %v", f.stock(t, "Gomilocas"), stockAntes)
		}
		p := f.propuesta(t, compra.Detalles[0].ID)
		// 600 + 48 IEPS + 103.68 IVA = 751.68 / 24 = 31.32
		if !cerca(p.CostoCapturado, 31.32) || p.PiezasBuenas != 24 || p.PrecioSugerido != preciodomain.PrecioSugerido(31.32, p.MargenSugerido) {
			t.Fatalf("propuesta = %+v", p)
		}
		var l domain.Lote
		f.db.Where("id = ?", editada.Detalles[0].LoteID).First(&l)
		if l.CantidadInicial != 24 || l.CantidadActual != 24 || l.CostoUnitario == nil || !cerca(*l.CostoUnitario, 31.32) {
			t.Fatalf("lote = %+v", l)
		}
		if editada.Total != 1015.68 {
			t.Fatalf("total = %v", editada.Total)
		}
		var ajustes int64
		f.db.Model(&domain.MovimientoInventario{}).Where("producto_negocio_id = ? AND tipo = ?", f.productos["Gomilocas"], compradomain.MovimientoCompraAjuste).Count(&ajustes)
		if ajustes != 1 {
			t.Fatalf("movimientos de ajuste = %d", ajustes)
		}
	})

	t.Run("quitar y agregar renglones", func(t *testing.T) {
		compra := f.comprar(t, f.entrada(f.renglon("Leche", 264, tasa(0), tasa(0)), f.renglon("Arroz", 120, tasa(0), tasa(0.16))))
		arrozAntes := f.stock(t, "Arroz")
		input := edicion(compra)
		quitar := renglonEditado(compra.Detalles[1], 1, 120)
		quitar.Quitar = true
		nuevo := compraapp.EditarDetalleInput{CrearDetalleInput: f.renglon("Gomilocas", 100, tasa(0), tasa(0))}
		input.Detalles = []compraapp.EditarDetalleInput{renglonEditado(compra.Detalles[0], 1, 264), quitar, nuevo}
		editada, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, compra.ID, input)
		if err != nil {
			t.Fatal(err)
		}
		if len(editada.Detalles) != 2 || f.stock(t, "Arroz") != arrozAntes-12 || f.propuesta(t, compra.Detalles[1].ID).Estado != preciodomain.EstadoCancelada {
			t.Fatalf("quitar renglón: %d renglones, stock %v", len(editada.Detalles), f.stock(t, "Arroz"))
		}
		if editada.Total != 364 {
			t.Fatalf("total = %v", editada.Total)
		}
	})

	t.Run("bajar piezas ya vendidas se rechaza", func(t *testing.T) {
		compra := f.comprar(t, f.entrada(f.renglon("Leche", 264, tasa(0), tasa(0))))
		f.fijarStock(t, "Leche", 3)
		input := edicion(compra)
		input.Detalles = []compraapp.EditarDetalleInput{renglonEditado(compra.Detalles[0], 0, 264)}
		input.Detalles[0].CantidadEmpaques = 1
		input.Detalles[0].PiezasFaltantes = 11
		if _, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, compra.ID, input); err == nil || !strings.Contains(err.Error(), "Leche") {
			t.Fatalf("se aceptó dejar stock negativo: %v", err)
		}
		if f.stock(t, "Leche") != 3 {
			t.Fatal("el rechazo cambió el stock")
		}
		f.fijarStock(t, "Leche", 500)
	})

	t.Run("con un costo validado solo se edita el documento", func(t *testing.T) {
		compra := f.comprar(t, f.entrada(f.renglon("Arroz", 120, tasa(0), tasa(0))))
		p := f.propuesta(t, compra.Detalles[0].ID)
		if _, err := f.precios.Authorize(f.negocio.ID, f.usuario.ID, p.ID, precioapp.AutorizarInput{Costo: p.CostoCapturado, Margen: 0.3, PrecioVenta: 14}); err != nil {
			t.Fatal(err)
		}
		input := edicion(compra)
		input.Detalles = []compraapp.EditarDetalleInput{renglonEditado(compra.Detalles[0], 2, 120)}
		if _, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, compra.ID, input); err == nil || !strings.Contains(err.Error(), "validados") {
			t.Fatalf("se editó con costo validado: %v", err)
		}
		input.Detalles = nil
		input.FormaPago = "TARJETA"
		if _, err := f.compras.Edit(f.negocio.ID, f.usuario.ID, compra.ID, input); err != nil {
			t.Fatalf("encabezado con costo validado: %v", err)
		}
	})

	t.Run("suspender, reactivar y suspender otra vez", func(t *testing.T) {
		renglon := f.renglon("Gomilocas", 240, tasa(0.08), tasa(0.16))
		renglon.NumeroLote = &lote
		compra := f.comprar(t, f.entrada(renglon, f.renglon("Leche", 264, tasa(0), tasa(0))))
		validada := f.propuesta(t, compra.Detalles[1].ID)
		if _, err := f.precios.Authorize(f.negocio.ID, f.usuario.ID, validada.ID, precioapp.AutorizarInput{Costo: validada.CostoCapturado, Margen: 0.3, PrecioVenta: 29}); err != nil {
			t.Fatal(err)
		}
		gomilocas, leche := f.stock(t, "Gomilocas"), f.stock(t, "Leche")
		if _, err := f.compras.Suspend(f.negocio.ID, f.usuario.ID, compra.ID, " "); err == nil {
			t.Fatal("se suspendió sin motivo")
		}
		suspendida, err := f.compras.Suspend(f.negocio.ID, f.usuario.ID, compra.ID, "El proveedor recogió la mercancía")
		if err != nil {
			t.Fatal(err)
		}
		if suspendida.Estado != compradomain.EstadoSuspendida || f.stock(t, "Gomilocas") != gomilocas-12 || f.stock(t, "Leche") != leche-12 {
			t.Fatalf("suspender: estado %s stock %v/%v", suspendida.Estado, f.stock(t, "Gomilocas"), f.stock(t, "Leche"))
		}
		pendiente := f.propuesta(t, compra.Detalles[0].ID)
		if pendiente.Estado != preciodomain.EstadoCancelada || f.propuesta(t, compra.Detalles[1].ID).Estado != preciodomain.EstadoAutorizada {
			t.Fatalf("propuestas tras suspender: %s / validada %s", pendiente.Estado, f.propuesta(t, compra.Detalles[1].ID).Estado)
		}
		for _, item := range f.pendientes(t, precioapp.FiltroPendientes{}) {
			if item.ID == pendiente.ID {
				t.Fatal("la propuesta cancelada sigue en Validar costos")
			}
		}
		var l domain.Lote
		f.db.Where("id = ?", suspendida.Detalles[0].LoteID).First(&l)
		if l.CantidadActual != 0 {
			t.Fatalf("lote tras suspender = %v", l.CantidadActual)
		}
		if _, err := f.compras.Suspend(f.negocio.ID, f.usuario.ID, compra.ID, "otra vez"); err == nil {
			t.Fatal("se suspendió una compra suspendida")
		}
		// Una cancelada por otra causa no debe regresar al reactivar.
		otra := f.comprar(t, f.entrada(f.renglon("Arroz", 120, tasa(0), tasa(0))))
		f.db.Model(&preciodomain.Propuesta{}).Where("compra_detalle_id = ?", otra.Detalles[0].ID).Update("estado", preciodomain.EstadoCancelada)

		activa, err := f.compras.Reactivate(f.negocio.ID, f.usuario.ID, compra.ID, "")
		if err != nil {
			t.Fatal(err)
		}
		if activa.Estado != compradomain.EstadoActiva || activa.ReactivadoEn == nil || f.stock(t, "Gomilocas") != gomilocas || f.stock(t, "Leche") != leche {
			t.Fatalf("reactivar: %s stock %v/%v", activa.Estado, f.stock(t, "Gomilocas"), f.stock(t, "Leche"))
		}
		restaurada := f.propuesta(t, compra.Detalles[0].ID)
		if restaurada.Estado != preciodomain.EstadoPendiente || restaurada.CostoAnterior == nil || *restaurada.CostoAnterior != *pendiente.CostoAnterior || restaurada.PrecioAnterior != pendiente.PrecioAnterior {
			t.Fatalf("propuesta restaurada = %+v", restaurada)
		}
		if f.propuesta(t, otra.Detalles[0].ID).Estado != preciodomain.EstadoCancelada {
			t.Fatal("regresó una propuesta que no canceló la suspensión")
		}
		if _, err := f.compras.Reactivate(f.negocio.ID, f.usuario.ID, compra.ID, ""); err == nil {
			t.Fatal("se reactivó una compra activa")
		}
		if _, err := f.compras.Suspend(f.negocio.ID, f.usuario.ID, compra.ID, "Segunda devolución"); err != nil {
			t.Fatalf("suspender otra vez: %v", err)
		}
		var movimientos []domain.MovimientoInventario
		f.db.Where("producto_negocio_id = ? AND referencia = ?", f.productos["Gomilocas"], compra.FolioDocumento).Order("creado_en").Find(&movimientos)
		tipos := []string{}
		for _, m := range movimientos {
			tipos = append(tipos, m.Tipo)
		}
		if strings.Join(tipos, ",") != "COMPRA_ENTRADA,COMPRA_SUSPENSION,COMPRA_REACTIVACION,COMPRA_SUSPENSION" {
			t.Fatalf("kardex = %v", tipos)
		}
		lista, _ := f.compras.List(f.negocio.ID, f.sucursal.ID)
		for _, item := range lista {
			if item.ID == compra.ID && (item.Estado != compradomain.EstadoSuspendida || item.CostosCancelados != 1 || item.CostosValidados != 1) {
				t.Fatalf("lista = %+v", item)
			}
		}
		excel := compraapp.NewService(f.compras).ConExportador(comprainfra.NewExcel())
		archivo, n, _ := excel.Exportar(f.negocio.ID, compraapp.FiltroExportacion{SucursalID: f.sucursal.ID, CompraIDs: []uuid.UUID{compra.ID}})
		if n != 0 || len(archivo) == 0 {
			t.Fatalf("el Excel por omisión incluyó la suspendida: %d", n)
		}
	})

	t.Run("suspender con piezas vendidas se rechaza y lista los productos", func(t *testing.T) {
		compra := f.comprar(t, f.entrada(f.renglon("Leche", 264, tasa(0), tasa(0)), f.renglon("Arroz", 120, tasa(0), tasa(0))))
		f.fijarStock(t, "Leche", 2)
		f.fijarStock(t, "Arroz", 1)
		_, err := f.compras.Suspend(f.negocio.ID, f.usuario.ID, compra.ID, "Devolución")
		if err == nil || !strings.Contains(err.Error(), "Leche") || !strings.Contains(err.Error(), "Arroz") {
			t.Fatalf("suspender sin stock: %v", err)
		}
	})
}

func tasa2(s string) *string { return &s }

func TestComprasPermisosSuspender(t *testing.T) {
	f := nuevoComprasFixture(t)
	ctx := context.Background()
	roles := negocioapp.NewRolService(negocioinfra.NewRolRepository(f.db))
	duena := crearUsuario(t, f.db, "duena-"+uuid.NewString()[:8]+"@tienda.mx")
	cajero := crearUsuario(t, f.db, "cajero-"+uuid.NewString()[:8]+"@tienda.mx")
	negocio, err := negocioapp.NewNegocioService(negocioinfra.NewNegocioRepository(f.db)).Crear(ctx, duena, negociodomain.CrearNegocioInput{NombreComercial: "Tienda Dueña"})
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		id, _ := uuid.Parse(c.GetHeader("X-Usuario"))
		c.Set("authenticated_user_id", id)
	})
	ok := func(c *gin.Context) { c.Status(http.StatusNoContent) }
	router.POST("/negocios/:negocioId/compras/:compraId/suspender", negociohttp.RequierePermiso(roles, negociodomain.PermisoCompraSuspender), ok)
	router.PATCH("/negocios/:negocioId/compras/:compraId", negociohttp.RequierePermiso(roles, negociodomain.PermisoCompraEditar), ok)
	llamar := func(metodo, ruta string, usuario uuid.UUID) int {
		req := httptest.NewRequest(metodo, "/negocios/"+negocio.ID.String()+ruta, nil)
		req.Header.Set("X-Usuario", usuario.String())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Code
	}
	compra := "/compras/" + uuid.NewString()
	if code := llamar(http.MethodPost, compra+"/suspender", duena); code != http.StatusNoContent {
		t.Fatalf("propietario sin asignación explícita = %d", code)
	}
	if code := llamar(http.MethodPatch, compra, duena); code != http.StatusNoContent {
		t.Fatalf("propietario editar = %d", code)
	}
	if code := llamar(http.MethodPost, compra+"/suspender", cajero); code != http.StatusForbidden {
		t.Fatalf("sin permiso = %d", code)
	}
	if deps := negociodomain.DependenciasPermiso(negociodomain.PermisoCompraEditar); strings.Join(deps, ",") != "compras.ver,catalogo.ver" {
		t.Fatalf("dependencias de editar = %v", deps)
	}
}
