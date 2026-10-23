package database

import (
	"errors"
	"os"
	"strings"
	"testing"

	"tienda/backend/internal/domain"
	negociodomain "tienda/backend/internal/domain/negocio"
	preciodomain "tienda/backend/internal/domain/precio"
	"tienda/backend/internal/infrastructure"

	"github.com/google/uuid"
)

// TestSurtidoPorSucursal covers capability 28: each branch lists only its own
// assortment, the owner adds catalog products to a branch, and bulk imports add
// or update existing products without duplicating the business catalog.
func TestSurtidoPorSucursal(t *testing.T) {
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
	if err := db.Exec(`DROP SCHEMA public CASCADE`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE SCHEMA public`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Init(db); err != nil {
		t.Fatal(err)
	}
	repo := infrastructure.NewProductRepository(db)

	negocio := negociodomain.Negocio{ID: uuid.New(), Slug: "surtido", NombreComercial: "Surtido", Nombre: "Surtido", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	if err := db.Create(&negocio).Error; err != nil {
		t.Fatal(err)
	}
	sucursalA := negociodomain.Sucursal{ID: uuid.New(), NegocioID: negocio.ID, Codigo: "SUC-001", Nombre: "A", EsPrincipal: true, Activo: true}
	sucursalB := negociodomain.Sucursal{ID: uuid.New(), NegocioID: negocio.ID, Codigo: "SUC-002", Nombre: "B", Activo: true}
	for _, sucursal := range []*negociodomain.Sucursal{&sucursalA, &sucursalB} {
		if err := db.Create(sucursal).Error; err != nil {
			t.Fatal(err)
		}
	}

	talla := func(valor string) []domain.ProductVariantAttributeInput {
		return []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: valor}}
	}
	archivo := func(precioJabon, costoJabon, stockJabon float64, costoCapturado bool) []domain.ProductImportRow {
		return []domain.ProductImportRow{
			{Fila: 2, Nombre: "Jabón", PrecioVenta: precioJabon, CostoProveedor: costoJabon, CostoCapturado: costoCapturado, StockInicial: stockJabon, CodigoBarras: "7501000000017"},
			{Fila: 3, Nombre: "Cloro", PrecioVenta: 20, CostoProveedor: 8, CostoCapturado: true, StockInicial: 3},
			{Fila: 4, Nombre: "Playera", PrecioVenta: 99, CostoProveedor: 50, CostoCapturado: true, StockInicial: 2, Variantes: talla("M")},
			{Fila: 5, Nombre: "Playera", PrecioVenta: 99, CostoProveedor: 50, CostoCapturado: true, StockInicial: 1, Variantes: talla("G")},
		}
	}
	importar := func(sucursalID uuid.UUID, rows []domain.ProductImportRow) (domain.CatalogImportResult, []domain.ValidatedProductImportRow, domain.CatalogImportResult) {
		prepared, preview, _, err := repo.ValidateProductImport(negocio.ID, sucursalID, rows)
		if err != nil {
			t.Fatal(err)
		}
		result, err := repo.CreateImportedProducts(negocio.ID, prepared)
		if err != nil {
			t.Fatal(err)
		}
		return preview, prepared, result
	}
	contar := func(tabla string) int64 {
		var total int64
		if err := db.Table(tabla).Count(&total).Error; err != nil {
			t.Fatal(err)
		}
		return total
	}
	listar := func(sucursalID uuid.UUID) map[string]domain.ProductRow {
		items, err := repo.ListByBusiness(negocio.ID, sucursalID)
		if err != nil {
			t.Fatal(err)
		}
		result := map[string]domain.ProductRow{}
		for _, item := range items {
			result[item.Nombre] = item
		}
		return result
	}

	if _, _, result := importar(sucursalA.ID, archivo(10, 5, 4, true)); result.Creadas != 4 {
		t.Fatalf("importación inicial en A: %+v", result)
	}
	productosIniciales, ofertasIniciales := contar("productos"), contar("producto_negocio")
	listaA := listar(sucursalA.ID)
	jabon, playera := listaA["Jabón"], listaA["Playera"]

	t.Run("una sucursal nueva no muestra el catálogo", func(t *testing.T) {
		if got := listar(sucursalB.ID); len(got) != 0 {
			t.Fatalf("B debe estar vacía: %+v", got)
		}
		if got := listar(sucursalA.ID); len(got) != 3 || len(got["Playera"].Variantes) != 2 {
			t.Fatalf("A cambió: %+v", got)
		}
		faltantes, err := repo.ListCatalogNotInBranch(negocio.ID, sucursalB.ID)
		if err != nil || len(faltantes) != 3 {
			t.Fatalf("fuera de B = %+v %v", faltantes, err)
		}
	})

	t.Run("agregar del catálogo sin precio ni stock", func(t *testing.T) {
		agregados, err := repo.AddToBranch(negocio.ID, domain.AddToBranchInput{SucursalID: sucursalB.ID, ProductoIDs: []uuid.UUID{playera.ID}})
		if err != nil || agregados != 2 {
			t.Fatalf("agregados = %d %v", agregados, err)
		}
		listaB := listar(sucursalB.ID)
		if len(listaB) != 1 || len(listaB["Playera"].Variantes) != 2 || listaB["Playera"].Variantes[0].Precio != 0 || listaB["Playera"].Stock != 0 {
			t.Fatalf("B = %+v", listaB)
		}
		if again, err := repo.AddToBranch(negocio.ID, domain.AddToBranchInput{SucursalID: sucursalB.ID, ProductoIDs: []uuid.UUID{playera.ID}}); err != nil || again != 0 {
			t.Fatalf("agregar otra vez = %d %v", again, err)
		}
		if contar("productos") != productosIniciales || contar("producto_negocio") != ofertasIniciales {
			t.Fatal("agregar a la sucursal duplicó el catálogo")
		}
	})

	t.Run("importar en B agrega y actualiza sin duplicar", func(t *testing.T) {
		preview, prepared, result := importar(sucursalB.ID, archivo(12, 6, 7, true))
		acciones := map[int]string{}
		for _, row := range prepared {
			acciones[row.Fila] = row.Accion
		}
		if acciones[2] != domain.ImportAccionAgregarSucursal || acciones[3] != domain.ImportAccionAgregarSucursal || acciones[4] != domain.ImportAccionActualizar || acciones[5] != domain.ImportAccionActualizar {
			t.Fatalf("acciones = %+v (errores %+v)", acciones, preview.Errores)
		}
		if result.AgregadasSucursal != 2 || result.Actualizadas != 2 || result.Creadas != 0 {
			t.Fatalf("resultado = %+v", result)
		}
		if contar("productos") != productosIniciales || contar("producto_negocio") != ofertasIniciales {
			t.Fatal("la importación duplicó el catálogo")
		}
		listaB := listar(sucursalB.ID)
		if listaB["Jabón"].Precio != 12 || listaB["Jabón"].Stock != 7 || listaB["Playera"].Stock != 3 {
			t.Fatalf("B = %+v", listaB)
		}
		if listaA := listar(sucursalA.ID); listaA["Jabón"].Precio != 10 || listaA["Jabón"].Stock != 4 {
			t.Fatalf("A cambió: %+v", listaA["Jabón"])
		}
	})

	jabonOferta := func() uuid.UUID {
		var oferta domain.ProductoNegocio
		if err := db.Where("producto_id = ?", jabon.ID).First(&oferta).Error; err != nil {
			t.Fatal(err)
		}
		return oferta.ID
	}()
	precioB := func() preciodomain.PrecioSucursal {
		var precio preciodomain.PrecioSucursal
		if err := db.Where("sucursal_id = ? AND producto_negocio_id = ?", sucursalB.ID, jabonOferta).First(&precio).Error; err != nil {
			t.Fatal(err)
		}
		return precio
	}

	t.Run("reimportar reemplaza stock y actualiza precio y costo", func(t *testing.T) {
		preview, _, result := importar(sucursalB.ID, archivo(15, 9, 2, true))
		avisos := 0
		for _, item := range preview.Advertencias {
			if item.Fila == 2 && (item.Campo == "Stock" || item.Campo == "Precio de venta" || item.Campo == "Costo") {
				avisos++
			}
		}
		if avisos != 3 {
			t.Fatalf("advertencias del jabón = %+v", preview.Advertencias)
		}
		if result.Actualizadas != 1 || preview.Omitidas != 3 {
			t.Fatalf("resultado = %+v, vista previa = %+v", result, preview)
		}
		if listaB := listar(sucursalB.ID); listaB["Jabón"].Stock != 2 || listaB["Jabón"].Precio != 15 {
			t.Fatalf("jabón en B = %+v", listaB["Jabón"])
		}
		if actual := precioB(); actual.CostoVigente != 9 {
			t.Fatalf("costo en B = %+v", actual)
		}
		var eventos []preciodomain.HistorialCostoPrecioSucursal
		db.Where("producto_negocio_id = ? AND sucursal_id = ? AND origen = ?", jabonOferta, sucursalB.ID, preciodomain.OrigenCargaMasiva).Find(&eventos)
		if len(eventos) != 1 || *eventos[0].PrecioAnterior != 12 || *eventos[0].PrecioNuevo != 15 || *eventos[0].CostoAnterior != 6 || *eventos[0].CostoNuevo != 9 {
			t.Fatalf("eventos = %+v", eventos)
		}
		var movimiento domain.MovimientoInventario
		if err := db.Where("producto_negocio_id = ? AND sucursal_id = ? AND tipo = 'AJUSTE_MANUAL'", jabonOferta, sucursalB.ID).First(&movimiento).Error; err != nil || movimiento.Cantidad != -5 {
			t.Fatalf("movimiento = %+v %v", movimiento, err)
		}
	})

	t.Run("costo vacío conserva el costo y una fila sin cambios no escribe", func(t *testing.T) {
		_, _, result := importar(sucursalB.ID, archivo(16, 0, 2, false))
		if result.Actualizadas != 1 {
			t.Fatalf("resultado = %+v", result)
		}
		if actual := precioB(); actual.CostoVigente != 9 || actual.PrecioVenta != 16 {
			t.Fatalf("precio en B = %+v", actual)
		}
		preview, prepared, _ := importar(sucursalB.ID, archivo(16, 0, 2, false))
		if len(prepared) != 0 || preview.Omitidas != 4 {
			t.Fatalf("sin cambios: %d filas, %+v", len(prepared), preview)
		}
	})

	t.Run("fila repetida en el archivo y datos compartidos distintos", func(t *testing.T) {
		filas := []domain.ProductImportRow{
			{Fila: 2, Nombre: "Cloro", PrecioVenta: 21, StockInicial: 3, Marca: "Otra marca"},
			{Fila: 3, Nombre: "Cloro", PrecioVenta: 22, StockInicial: 3},
		}
		prepared, preview, _, err := repo.ValidateProductImport(negocio.ID, sucursalB.ID, filas)
		if err != nil {
			t.Fatal(err)
		}
		if len(prepared) != 1 || prepared[0].Fila != 2 || preview.Omitidas != 1 {
			t.Fatalf("prepared = %+v, preview = %+v", prepared, preview)
		}
		conservados := false
		for _, item := range preview.Advertencias {
			conservados = conservados || strings.Contains(item.Motivo, "se conservan los datos del catálogo")
		}
		if !conservados {
			t.Fatalf("advertencias = %+v", preview.Advertencias)
		}
	})

	t.Run("crear un producto existente ofrece agregarlo", func(t *testing.T) {
		if err := db.Exec(`DELETE FROM precios_sucursal WHERE sucursal_id = ?`, sucursalB.ID).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DELETE FROM movimientos_inventario WHERE sucursal_id = ?`, sucursalB.ID).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DELETE FROM inventario_sucursal WHERE sucursal_id = ? AND producto_negocio_id = ?`, sucursalB.ID, jabonOferta).Error; err != nil {
			t.Fatal(err)
		}
		alta := domain.CreateProductInput{Nombre: " jabon ", SucursalID: sucursalB.ID, PrecioVenta: 11, StockInicial: 1}
		var otraSucursal *domain.ProductInOtherBranchError
		if err := repo.Create(negocio.ID, alta); !errors.As(err, &otraSucursal) || otraSucursal.ProductoID != jabon.ID {
			t.Fatalf("crear en B = %v", err)
		}
		if err := repo.AddProductToBranch(negocio.ID, jabon.ID, domain.AddProductToBranchInput{SucursalID: sucursalB.ID, PrecioVenta: 11, Costo: 4, StockInicial: 6}); err != nil {
			t.Fatal(err)
		}
		if listaB := listar(sucursalB.ID); listaB["Jabón"].Precio != 11 || listaB["Jabón"].Stock != 6 {
			t.Fatalf("jabón en B = %+v", listaB["Jabón"])
		}
		if err := repo.AddProductToBranch(negocio.ID, jabon.ID, domain.AddProductToBranchInput{SucursalID: sucursalB.ID}); !errors.Is(err, domain.ErrAlreadyInBranch) {
			t.Fatalf("agregar otra vez = %v", err)
		}
		alta.SucursalID = sucursalA.ID
		if err := repo.Create(negocio.ID, alta); !errors.Is(err, domain.ErrAlreadyInBranch) {
			t.Fatalf("crear en A = %v", err)
		}
		porCodigo := domain.CreateProductInput{Nombre: "Otro nombre", SucursalID: sucursalA.ID, CodigoBarras: ptr("7501000000017")}
		if err := repo.Create(negocio.ID, porCodigo); !errors.Is(err, domain.ErrAlreadyInBranch) {
			t.Fatalf("crear por código en A = %v", err)
		}
		if contar("productos") != productosIniciales || contar("producto_negocio") != ofertasIniciales {
			t.Fatal("crear o agregar duplicó el catálogo")
		}
	})

	t.Run("agregar todos", func(t *testing.T) {
		agregados, err := repo.AddToBranch(negocio.ID, domain.AddToBranchInput{SucursalID: sucursalB.ID, Todos: true})
		if err != nil || agregados != 0 {
			t.Fatalf("B ya tiene todo: %d %v", agregados, err)
		}
		if faltantes, _ := repo.ListCatalogNotInBranch(negocio.ID, sucursalB.ID); len(faltantes) != 0 {
			t.Fatalf("faltantes = %+v", faltantes)
		}
	})
}

func ptr(value string) *string { return &value }
