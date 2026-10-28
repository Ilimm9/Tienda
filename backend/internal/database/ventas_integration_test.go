package database

import (
	"errors"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	ventaapp "tienda/backend/internal/application/venta"
	"tienda/backend/internal/domain"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	preciodomain "tienda/backend/internal/domain/precio"
	ventadomain "tienda/backend/internal/domain/venta"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"
	ventainfra "tienda/backend/internal/infrastructure/venta"

	"github.com/google/uuid"
)

// TestVentas cubre la Fase 1 de la capability 37: búsqueda, registro transaccional con folio consecutivo, existencias, cancelación, ventas propias y migración del rol Cajero.
func TestVentas(t *testing.T) {
	url := os.Getenv("VENTAS_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("VENTAS_TEST_DATABASE_URL no está configurada")
	}
	if !strings.Contains(url, "tienda_ventas_test") {
		t.Fatal("la prueba destructiva sólo puede usar una base llamada tienda_ventas_test")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`DROP SCHEMA public CASCADE`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE SCHEMA public`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Init(db); err != nil {
		t.Fatal(err)
	}
	repo := ventainfra.NewRepository(db)
	service := ventaapp.NewService(repo)

	crearNegocio := func(slug string) negociodomain.Negocio {
		negocio := negociodomain.Negocio{ID: uuid.New(), Slug: slug, NombreComercial: slug, Nombre: slug, CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
		if err := db.Create(&negocio).Error; err != nil {
			t.Fatal(err)
		}
		return negocio
	}
	crearSucursal := func(negocioID uuid.UUID, codigo string) negociodomain.Sucursal {
		sucursal := negociodomain.Sucursal{ID: uuid.New(), NegocioID: negocioID, Codigo: codigo, Nombre: codigo, EsPrincipal: codigo == "SUC-001", Activo: true}
		if err := db.Create(&sucursal).Error; err != nil {
			t.Fatal(err)
		}
		return sucursal
	}
	crearUsuario := func(correo string) uuid.UUID {
		usuario := cuentadomain.Usuario{ID: uuid.New(), Correo: correo, HashContrasena: "no-usada-en-prueba", Estado: "activo"}
		if err := db.Create(&usuario).Error; err != nil {
			t.Fatal(err)
		}
		return usuario.ID
	}
	// crearProducto registra producto, código opcional, precio/costo en la sucursal (si precio >= 0) y existencia.
	crearProducto := func(negocioID, sucursalID uuid.UUID, nombre, codigo string, precio, costo float64, stock float64) uuid.UUID {
		producto := domain.Producto{NegocioID: negocioID, Nombre: nombre, Activo: true}
		if err := db.Create(&producto).Error; err != nil {
			t.Fatal(err)
		}
		oferta := domain.ProductoNegocio{NegocioID: negocioID, ProductoID: &producto.ID, PrecioIncluyeImpuestos: true, Activo: true}
		if err := db.Create(&oferta).Error; err != nil {
			t.Fatal(err)
		}
		if codigo != "" {
			if err := db.Create(&domain.ProductoCodigo{NegocioID: negocioID, ProductoID: &producto.ID, Tipo: "GTIN", Codigo: codigo, EsPrincipal: true}).Error; err != nil {
				t.Fatal(err)
			}
		}
		if precio >= 0 {
			if err := db.Create(&preciodomain.PrecioSucursal{NegocioID: negocioID, SucursalID: sucursalID, ProductoNegocioID: oferta.ID, PrecioVenta: precio, CostoVigente: costo}).Error; err != nil {
				t.Fatal(err)
			}
		}
		if err := db.Create(&domain.InventarioSucursal{SucursalID: sucursalID, ProductoNegocioID: oferta.ID, StockActual: stock}).Error; err != nil {
			t.Fatal(err)
		}
		return oferta.ID
	}
	stock := func(sucursalID, productoID uuid.UUID) float64 {
		var inventario domain.InventarioSucursal
		if err := db.Where("sucursal_id = ? AND producto_negocio_id = ?", sucursalID, productoID).First(&inventario).Error; err != nil {
			t.Fatal(err)
		}
		return inventario.StockActual
	}
	contar := func(tabla string) int64 {
		var total int64
		if err := db.Table(tabla).Count(&total).Error; err != nil {
			t.Fatal(err)
		}
		return total
	}

	negocio := crearNegocio("tienda-ventas")
	sucursal := crearSucursal(negocio.ID, "SUC-001")
	otraSucursal := crearSucursal(negocio.ID, "SUC-002")
	cajeroA := crearUsuario("cajero-a@example.com")
	cajeroB := crearUsuario("cajero-b@example.com")
	if err := db.Exec(`INSERT INTO perfil_usuarios (id, usuario_id, nombres, primer_apellido, creado_en, actualizado_en) VALUES (gen_random_uuid(), ?, 'Beatriz', 'Cajera', now(), now())`, cajeroB).Error; err != nil {
		t.Fatal(err)
	}
	coca := crearProducto(negocio.ID, sucursal.ID, "Coca-Cola 600 ml", "7501055300075", 18.5, 12, 50)
	leche := crearProducto(negocio.ID, sucursal.ID, "Leche Lala 1 L", "7501020515343", 28, 30.1234, 5)
	galletas := crearProducto(negocio.ID, sucursal.ID, "Galletas Marías", "123", 15, 9, 0)
	crearProducto(negocio.ID, sucursal.ID, "Jabón Zote", "7501026005688", -1, 0, 8)
	crearProducto(negocio.ID, sucursal.ID, "Producto en cero", "", 0, 0, 8)
	ultimo := crearProducto(negocio.ID, sucursal.ID, "Último pan", "", 10, 5, 1)
	ajeno := crearNegocio("otra-tienda")
	ajenaSucursal := crearSucursal(ajeno.ID, "SUC-001")
	crearProducto(ajeno.ID, ajenaSucursal.ID, "Coca ajena", "7501055300075", 99, 1, 9)

	efectivo := func(pesos string) *ventadomain.Centavos {
		valor, err := ventadomain.ParsePesos(pesos)
		if err != nil {
			t.Fatal(err)
		}
		return &valor
	}
	vender := func(usuarioID uuid.UUID, forma string, recibido *ventadomain.Centavos, lineas ...ventaapp.LineaInput) (ventadomain.Venta, error) {
		return service.Registrar(negocio.ID, usuarioID, ventaapp.RegistrarInput{SucursalID: sucursal.ID.String(), FormaPago: forma, Recibido: recibido, Lineas: lineas})
	}
	linea := func(id uuid.UUID, cantidad int) ventaapp.LineaInput {
		return ventaapp.LineaInput{ProductoNegocioID: id.String(), Cantidad: cantidad}
	}

	t.Run("busca por código exacto, también corto, sin mezclar negocios", func(t *testing.T) {
		resultado, err := service.Buscar(negocio.ID, sucursal.ID, "7501055300075", true)
		if err != nil || !resultado.PorCodigo || len(resultado.Items) != 1 || resultado.Items[0].ProductoNegocioID != coca {
			t.Fatalf("código exacto: %+v %v", resultado, err)
		}
		item := resultado.Items[0]
		if item.Precio != 1850 || item.Existencia != 50 || item.Agotado || item.Costo == nil || *item.Costo != 1200 {
			t.Fatalf("precio, existencia y costo en centavos: %+v", item)
		}
		if item.ExistenciaBaja {
			t.Fatalf("50 piezas sin mínimo no es existencia baja: %+v", item)
		}
		db.Model(&domain.InventarioSucursal{}).Where("producto_negocio_id = ?", leche).Update("stock_minimo", 5)
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "7501020515343", false)
		if len(resultado.Items) != 1 || !resultado.Items[0].ExistenciaBaja {
			t.Fatalf("existencia igual al mínimo es baja: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "123", false)
		if len(resultado.Items) != 1 || resultado.Items[0].ProductoNegocioID != galletas || !resultado.Items[0].Agotado || resultado.Items[0].Costo != nil {
			t.Fatalf("código corto agotado y sin costos: %+v", resultado)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "7500000000000", false)
		if !resultado.PorCodigo || len(resultado.Items) != 0 {
			t.Fatalf("código inexistente: %+v", resultado)
		}
	})

	t.Run("busca por nombre sin acentos ni mayúsculas y excluye productos sin precio", func(t *testing.T) {
		resultado, _ := service.Buscar(negocio.ID, sucursal.ID, "GALLETAS marias", false)
		if len(resultado.Items) != 1 || resultado.Items[0].Nombre != "Galletas Marías" {
			t.Fatalf("nombre sin acentos: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "jabon", false)
		if len(resultado.Items) != 0 {
			t.Fatalf("producto sin precio en la sucursal no aparece: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "cero", false)
		if len(resultado.Items) != 0 {
			t.Fatalf("producto con precio 0 no aparece: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, otraSucursal.ID, "coca", false)
		if len(resultado.Items) != 0 {
			t.Fatalf("otra sucursal sin precio no ofrece el producto: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "100%_", false)
		if len(resultado.Items) != 0 {
			t.Fatalf("los comodines LIKE se escapan: %+v", resultado.Items)
		}
		resultado, _ = service.Buscar(negocio.ID, sucursal.ID, "leche", true)
		if len(resultado.Items) != 1 || resultado.Items[0].NivelMargen == nil || *resultado.Items[0].NivelMargen != ventadomain.NivelMargenNegativo {
			t.Fatalf("precio menor al costo marca margen negativo: %+v", resultado.Items)
		}
	})

	var primera ventadomain.Venta
	t.Run("registrar descuenta inventario, guarda foto y asigna folio", func(t *testing.T) {
		movimientos := contar("movimientos_inventario")
		primera, err = vender(cajeroA, ventadomain.PagoEfectivo, efectivo("100"), linea(coca, 2), linea(leche, 1), linea(coca, 1))
		if err != nil {
			t.Fatal(err)
		}
		if primera.Folio != "V-000001" || primera.Numero != 1 || primera.Estado != ventadomain.EstadoPagada {
			t.Fatalf("folio inicial: %+v", primera)
		}
		// 3 × 18.50 + 28.00 = 83.50; recibido 100 → cambio 16.50.
		if primera.Total != 8350 || primera.Subtotal+primera.IVA != primera.Total || primera.Cambio != 1650 || primera.Piezas != 4 {
			t.Fatalf("importes: %+v", primera)
		}
		if len(primera.Lineas) != 2 || primera.Lineas[0].ProductoNegocioID != coca || primera.Lineas[0].Cantidad != 3 || primera.Lineas[0].Importe != 5550 {
			t.Fatalf("las líneas repetidas se suman en una: %+v", primera.Lineas)
		}
		if primera.Lineas[1].CostoUnitario == nil || *primera.Lineas[1].CostoUnitario != 3012 || primera.Lineas[1].CodigoBarras == nil {
			t.Fatalf("foto de costo y código: %+v", primera.Lineas[1])
		}
		if stock(sucursal.ID, coca) != 47 || stock(sucursal.ID, leche) != 4 {
			t.Fatalf("existencias tras vender: coca %v leche %v", stock(sucursal.ID, coca), stock(sucursal.ID, leche))
		}
		var salidas []domain.MovimientoInventario
		db.Where("tipo = ? AND referencia = ?", ventadomain.MovimientoVentaSalida, "V-000001").Find(&salidas)
		if contar("movimientos_inventario") != movimientos+2 || len(salidas) != 2 || salidas[0].Cantidad >= 0 {
			t.Fatalf("movimientos de salida: %+v", salidas)
		}
		// Un cambio de precio posterior no altera el ticket.
		db.Model(&preciodomain.PrecioSucursal{}).Where("producto_negocio_id = ?", coca).Update("precio_venta", 20)
		guardada, _ := service.Obtener(negocio.ID, primera.ID)
		if guardada.Lineas[0].PrecioUnitario != 1850 || guardada.Total != 8350 {
			t.Fatalf("el ticket conserva su precio: %+v", guardada)
		}
		db.Model(&preciodomain.PrecioSucursal{}).Where("producto_negocio_id = ?", coca).Update("precio_venta", 18.5)
	})

	t.Run("tarjeta cobra el total y guarda la referencia; el folio sube en 1", func(t *testing.T) {
		referencia := "AUT-778"
		venta, err := service.Registrar(negocio.ID, cajeroB, ventaapp.RegistrarInput{SucursalID: sucursal.ID.String(), FormaPago: ventadomain.PagoTarjeta, Referencia: &referencia, Lineas: []ventaapp.LineaInput{linea(coca, 1)}})
		if err != nil || venta.Numero != primera.Numero+1 || venta.Recibido != venta.Total || venta.Cambio != 0 || venta.Referencia == nil || *venta.Referencia != referencia {
			t.Fatalf("venta con tarjeta: %+v %v", venta, err)
		}
	})

	t.Run("errores no guardan nada", func(t *testing.T) {
		ventas, movimientos, antes := contar("ventas"), contar("movimientos_inventario"), stock(sucursal.ID, coca)
		_, err := vender(cajeroA, ventadomain.PagoEfectivo, efectivo("1000"), linea(coca, 1), linea(galletas, 1))
		var existencia *ventadomain.ExistenciaInsuficienteError
		if !errors.As(err, &existencia) || existencia.Producto != "Galletas Marías" {
			t.Fatalf("agotado nombra el producto: %v", err)
		}
		_, err = vender(cajeroA, ventadomain.PagoEfectivo, efectivo("10000"), linea(leche, 5))
		if !errors.As(err, &existencia) || existencia.Disponible != 4 {
			t.Fatalf("más de lo disponible: %v", err)
		}
		if _, err := vender(cajeroA, ventadomain.PagoEfectivo, efectivo("18.49"), linea(coca, 1)); !errors.Is(err, ventadomain.ErrPagoInsuficiente) {
			t.Fatalf("recibido menor al total: %v", err)
		}
		var jabon struct{ ID uuid.UUID }
		if err := db.Raw(`SELECT pn.id FROM producto_negocio pn JOIN productos p ON p.id = pn.producto_id WHERE p.nombre = 'Jabón Zote'`).Scan(&jabon).Error; err != nil || jabon.ID == uuid.Nil {
			t.Fatalf("jabón: %v", err)
		}
		if _, err := vender(cajeroA, ventadomain.PagoTarjeta, nil, linea(jabon.ID, 1)); !errors.Is(err, ventaapp.ErrProductoNoDisponible) {
			t.Fatalf("sin precio no se vende: %v", err)
		}
		if _, err := service.Registrar(ajeno.ID, cajeroA, ventaapp.RegistrarInput{SucursalID: sucursal.ID.String(), FormaPago: ventadomain.PagoTarjeta, Lineas: []ventaapp.LineaInput{linea(coca, 1)}}); err == nil {
			t.Fatal("una sucursal de otro negocio se rechaza")
		}
		if contar("ventas") != ventas || contar("movimientos_inventario") != movimientos || stock(sucursal.ID, coca) != antes {
			t.Fatal("una venta fallida no deja rastro")
		}
		siguiente, err := vender(cajeroA, ventadomain.PagoTarjeta, nil, linea(coca, 1))
		if err != nil || siguiente.Numero != primera.Numero+2 {
			t.Fatalf("las ventas fallidas no consumen folio: %+v %v", siguiente, err)
		}
	})

	t.Run("ventas simultáneas reciben folios distintos y consecutivos", func(t *testing.T) {
		var maximo int64
		db.Raw(`SELECT COALESCE(MAX(numero), 0) FROM ventas WHERE negocio_id = ?`, negocio.ID).Scan(&maximo)
		const total = 8
		numeros := make([]int64, total)
		errores := make([]error, total)
		var grupo sync.WaitGroup
		for i := 0; i < total; i++ {
			grupo.Add(1)
			go func(i int) {
				defer grupo.Done()
				venta, err := vender(cajeroA, ventadomain.PagoTarjeta, nil, linea(coca, 1))
				numeros[i], errores[i] = venta.Numero, err
			}(i)
		}
		grupo.Wait()
		for _, err := range errores {
			if err != nil {
				t.Fatal(err)
			}
		}
		sort.Slice(numeros, func(i, j int) bool { return numeros[i] < numeros[j] })
		for i, numero := range numeros {
			if numero != maximo+int64(i)+1 {
				t.Fatalf("folios esperados consecutivos desde %d: %v", maximo+1, numeros)
			}
		}
	})

	t.Run("dos cajeros por la última pieza: solo uno la vende", func(t *testing.T) {
		resultados := make([]error, 2)
		var grupo sync.WaitGroup
		for i := range resultados {
			grupo.Add(1)
			go func(i int) {
				defer grupo.Done()
				_, resultados[i] = vender(cajeroA, ventadomain.PagoEfectivo, efectivo("10"), linea(ultimo, 1))
			}(i)
		}
		grupo.Wait()
		exitos := 0
		for _, err := range resultados {
			var existencia *ventadomain.ExistenciaInsuficienteError
			switch {
			case err == nil:
				exitos++
			case !errors.As(err, &existencia) || existencia.Producto != "Último pan":
				t.Fatalf("el perdedor debe recibir el nombre del producto: %v", err)
			}
		}
		if exitos != 1 || stock(sucursal.ID, ultimo) != 0 {
			t.Fatalf("una sola venta y stock en 0: %d %v", exitos, stock(sucursal.ID, ultimo))
		}
	})

	t.Run("cancelar regresa piezas una sola vez", func(t *testing.T) {
		antesCoca, antesLeche := stock(sucursal.ID, coca), stock(sucursal.ID, leche)
		if _, err := service.Cancelar(ventaapp.CancelarComando{NegocioID: negocio.ID, VentaID: primera.ID, UsuarioID: cajeroB, MotivoCodigo: ventadomain.MotivoErrorCaptura, Sucursales: []uuid.UUID{otraSucursal.ID}}); !errors.Is(err, ventaapp.ErrVentaNoEncontrada) {
			t.Fatalf("sin acceso a la sucursal responde como inexistente: %v", err)
		}
		cancelada, err := service.Cancelar(ventaapp.CancelarComando{NegocioID: negocio.ID, VentaID: primera.ID, UsuarioID: cajeroB, MotivoCodigo: ventadomain.MotivoDevolucionInmediata, Detalle: " Cliente devolvió ", Sucursales: []uuid.UUID{sucursal.ID}})
		if err != nil || cancelada.Estado != ventadomain.EstadoCancelada || cancelada.MotivoCancelacion == nil || *cancelada.MotivoCancelacion != "Devolución inmediata del cliente: Cliente devolvió" ||
			cancelada.MotivoCancelacionCodigo == nil || *cancelada.MotivoCancelacionCodigo != ventadomain.MotivoDevolucionInmediata ||
			cancelada.CanceladaPor == nil || *cancelada.CanceladaPor != "Beatriz Cajera" || cancelada.Cancelable ||
			cancelada.CanceladaPorUsuarioID == nil || *cancelada.CanceladaPorUsuarioID != cajeroB || cancelada.CanceladaEn == nil || len(cancelada.Lineas) != 2 {
			t.Fatalf("venta cancelada: %+v %v", cancelada, err)
		}
		if stock(sucursal.ID, coca) != antesCoca+3 || stock(sucursal.ID, leche) != antesLeche+1 {
			t.Fatal("las piezas regresan al inventario")
		}
		var entradas []domain.MovimientoInventario
		db.Where("tipo = ? AND referencia = ?", ventadomain.MovimientoVentaCancelacion, "V-000001").Find(&entradas)
		if len(entradas) != 2 || entradas[0].Cantidad <= 0 || entradas[0].Motivo == nil {
			t.Fatalf("movimientos de cancelación: %+v", entradas)
		}
		if _, err := service.Cancelar(ventaapp.CancelarComando{NegocioID: negocio.ID, VentaID: primera.ID, UsuarioID: cajeroB, MotivoCodigo: ventadomain.MotivoCobroDuplicado, Sucursales: []uuid.UUID{sucursal.ID}}); !errors.Is(err, ventaapp.ErrVentaYaCancelada) {
			t.Fatalf("no se cancela dos veces: %v", err)
		}
		if stock(sucursal.ID, coca) != antesCoca+3 {
			t.Fatal("una segunda cancelación no regresa piezas")
		}
	})

	t.Run("listar filtra por cajero, folio, forma de pago y sucursal", func(t *testing.T) {
		listar := func(filtro ventaapp.Filtro) []ventadomain.Venta {
			listado, err := service.Listar(negocio.ID, filtro)
			if err != nil {
				t.Fatal(err)
			}
			return listado.Items
		}
		todas := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}})
		propias := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, UsuarioID: &cajeroB})
		if len(propias) != 1 || propias[0].UsuarioID != cajeroB || len(todas) <= len(propias) {
			t.Fatalf("ventas propias: %d de %d", len(propias), len(todas))
		}
		if propias[0].Atendio != "Beatriz Cajera" || todas[len(todas)-1].Atendio != "cajero-a@example.com" {
			t.Fatalf("atendió usa el perfil o el correo: %q %q", propias[0].Atendio, todas[len(todas)-1].Atendio)
		}
		for _, folio := range []string{"V-000001", "1", "000001"} {
			if items := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, Folio: folio}); len(items) != 1 || items[0].ID != primera.ID {
				t.Fatalf("folio %s: %d", folio, len(items))
			}
		}
		if items := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, Folio: "abc"}); len(items) != 0 {
			t.Fatal("un folio inválido no lista todo")
		}
		if items := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, FormaPago: ventadomain.PagoEfectivo, Estado: ventadomain.EstadoCancelada}); len(items) != 1 {
			t.Fatalf("forma de pago y estado: %d", len(items))
		}
		if items := listar(ventaapp.Filtro{Sucursales: []uuid.UUID{otraSucursal.ID}}); len(items) != 0 {
			t.Fatal("otra sucursal no ve estas ventas")
		}
		if listado, _ := service.Listar(ajeno.ID, ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}}); len(listado.Items) != 0 {
			t.Fatal("otro negocio no ve estas ventas")
		}
	})

	t.Run("resumen del periodo: vendido, tickets, promedio, efectivo y cancelados", func(t *testing.T) {
		var esperado struct{ Vendido, Tickets, Efectivo, Cancelados int64 }
		db.Raw(`SELECT COALESCE(SUM(total_centavos) FILTER (WHERE estado = 'PAGADA'), 0) AS vendido, COUNT(*) FILTER (WHERE estado = 'PAGADA') AS tickets,
			COALESCE(SUM(total_centavos) FILTER (WHERE estado = 'PAGADA' AND forma_pago = 'EFECTIVO'), 0) AS efectivo, COUNT(*) FILTER (WHERE estado = 'CANCELADA') AS cancelados
			FROM ventas WHERE negocio_id = ?`, negocio.ID).Scan(&esperado)
		for _, periodo := range []string{"hoy", "semana", "mes"} {
			listado, err := service.Listar(negocio.ID, ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, Periodo: periodo, FormaPago: ventadomain.PagoTarjeta})
			if err != nil {
				t.Fatal(err)
			}
			r := listado.Resumen
			if int64(r.Vendido) != esperado.Vendido || int64(r.Tickets) != esperado.Tickets || int64(r.Efectivo) != esperado.Efectivo || int64(r.Cancelados) != esperado.Cancelados || esperado.Cancelados != 1 {
				t.Fatalf("%s: resumen %+v, esperado %+v", periodo, r, esperado)
			}
			if r.TicketPromedio != ventaapp.Promedio(r.Vendido, r.Tickets) || listado.Desde == nil || listado.Hasta == nil || !listado.Desde.Before(*listado.Hasta) {
				t.Fatalf("%s: promedio y límites %+v", periodo, listado)
			}
			for _, item := range listado.Items {
				if item.FormaPago != ventadomain.PagoTarjeta {
					t.Fatal("la forma de pago solo acota la tabla")
				}
			}
		}
		// Una venta de ayer queda fuera de «hoy» y ya no se puede cancelar.
		ayer, err := vender(cajeroA, ventadomain.PagoTarjeta, nil, linea(coca, 1))
		if err != nil {
			t.Fatal(err)
		}
		db.Exec(`UPDATE ventas SET creado_en = now() - interval '1 day' WHERE id = ?`, ayer.ID)
		hoy, _ := service.Listar(negocio.ID, ventaapp.Filtro{Sucursales: []uuid.UUID{sucursal.ID}, Periodo: "hoy"})
		for _, item := range hoy.Items {
			if item.ID == ayer.ID {
				t.Fatal("la venta de ayer no es de hoy")
			}
			if item.Estado == ventadomain.EstadoPagada && !item.Cancelable {
				t.Fatalf("una venta pagada de hoy es cancelable: %+v", item)
			}
		}
		if detalle, _ := service.Obtener(negocio.ID, ayer.ID); detalle.Cancelable {
			t.Fatal("una venta de ayer no es cancelable")
		}
		if _, err := service.Cancelar(ventaapp.CancelarComando{NegocioID: negocio.ID, VentaID: ayer.ID, UsuarioID: cajeroB, MotivoCodigo: ventadomain.MotivoErrorCaptura, Sucursales: []uuid.UUID{sucursal.ID}}); !errors.Is(err, ventaapp.ErrVentaFueraDePlazo) {
			t.Fatalf("fuera de plazo: %v", err)
		}
	})

	t.Run("migración del rol Cajero conserva permisos agregados y no toca roles personalizados", func(t *testing.T) {
		permisosDe := func(rolID uuid.UUID) string {
			var codigos []string
			db.Raw(`SELECT p.codigo FROM permisos_rol pr JOIN permisos p ON p.id = pr.permiso_id WHERE pr.rol_id = ? ORDER BY p.codigo`, rolID).Scan(&codigos)
			return strings.Join(codigos, ",")
		}
		crearRol := func(codigo string, permisos ...string) uuid.UUID {
			rol := negociodomain.Rol{NegocioID: negocio.ID, Codigo: codigo, Nombre: codigo, Activo: true}
			if err := db.Create(&rol).Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec(`INSERT INTO permisos_rol (id, rol_id, permiso_id, creado_en) SELECT gen_random_uuid(), ?, id, now() FROM permisos WHERE codigo IN ?`, rol.ID, permisos).Error; err != nil {
				t.Fatal(err)
			}
			return rol.ID
		}
		cajero := crearRol("cajero", "catalogo.ver", "compras.ver", "compras.registrar", "sucursales.ver")
		vendedor := crearRol("VENDEDOR", "catalogo.ver", "compras.ver")
		if err := db.Exec(`DELETE FROM migraciones_aplicacion WHERE clave = 'ventas_rol_cajero_v1'`).Error; err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if err := negocioinfra.MigrateRolCajeroVentas(db); err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				// La segunda corrida no debe tocar cambios hechos después de migrar.
				db.Exec(`INSERT INTO permisos_rol (id, rol_id, permiso_id, creado_en) SELECT gen_random_uuid(), ?, id, now() FROM permisos WHERE codigo = 'catalogo.ver'`, cajero)
			}
		}
		if got := permisosDe(cajero); got != "catalogo.ver,sucursales.ver,ventas.registrar" {
			t.Fatalf("permisos del Cajero migrado: %s", got)
		}
		if got := permisosDe(vendedor); got != "catalogo.ver,compras.ver" {
			t.Fatalf("roles personalizados no cambian: %s", got)
		}
		var catalogo int64
		db.Model(&negociodomain.Permiso{}).Where("codigo_modulo = ?", "ventas").Count(&catalogo)
		if catalogo != 4 {
			t.Fatalf("catálogo de permisos de ventas: %d", catalogo)
		}
	})
}
