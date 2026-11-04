package database

import (
	"os"
	"strings"
	"testing"

	"tienda/backend/internal/domain"
	negociodomain "tienda/backend/internal/domain/negocio"
	"tienda/backend/internal/infrastructure"

	"github.com/google/uuid"
)

// TestImportacionSinPresentacionNiMarca covers adecuación 1: the product
// spreadsheet no longer carries Presentación or Marca, so the barcode decides
// which product a row is and, without a barcode, the name does.
func TestImportacionSinPresentacionNiMarca(t *testing.T) {
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
	negocio := negociodomain.Negocio{ID: uuid.New(), Slug: "adecuacion", NombreComercial: "Adecuación", Nombre: "Adecuación", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	if err := db.Create(&negocio).Error; err != nil {
		t.Fatal(err)
	}
	sucursal := negociodomain.Sucursal{ID: uuid.New(), NegocioID: negocio.ID, Codigo: "SUC-001", Nombre: "Centro", EsPrincipal: true, Activo: true}
	if err := db.Create(&sucursal).Error; err != nil {
		t.Fatal(err)
	}
	marca := domain.Marca{NegocioID: negocio.ID, Nombre: "Acme", Activo: true}
	if err := db.Create(&marca).Error; err != nil {
		t.Fatal(err)
	}
	texto := func(value string) *string { return &value }
	crear := func(input domain.CreateProductInput) {
		t.Helper()
		input.SucursalID = sucursal.ID
		if err := repo.Create(negocio.ID, input); err != nil {
			t.Fatalf("crear %s: %v", input.Nombre, err)
		}
	}
	// Productos registrados desde el formulario, con presentación y marca.
	crear(domain.CreateProductInput{Nombre: "Coca-Cola", Presentacion: texto("600 ml"), MarcaID: &marca.ID, PrecioVenta: 18, StockInicial: 5, CodigoBarras: texto("7501000000024"), GenerarSKUInterno: true})
	crear(domain.CreateProductInput{Nombre: "Gansito", PrecioVenta: 15, StockInicial: 4, GenerarSKUInterno: true})
	crear(domain.CreateProductInput{Nombre: "Pan", Presentacion: texto("Chico"), PrecioVenta: 5, StockInicial: 1, GenerarSKUInterno: true})
	crear(domain.CreateProductInput{Nombre: "Pan", Presentacion: texto("Grande"), PrecioVenta: 9, StockInicial: 1, GenerarSKUInterno: true})
	crear(domain.CreateProductInput{Nombre: "Playera", MarcaID: &marca.ID, Descripcion: texto("Algodón"), Presentacion: texto("Pieza"), PrecioVenta: 99, GenerarSKUInterno: true,
		Variantes: []domain.CreateProductVariantInput{{Atributos: []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "M"}}, PrecioVenta: 99, StockInicial: 2, GenerarSKUInterno: true}}})

	contar := func(tabla string) int64 {
		var total int64
		if err := db.Table(tabla).Count(&total).Error; err != nil {
			t.Fatal(err)
		}
		return total
	}
	talla := func(valor string) []domain.ProductVariantAttributeInput {
		return []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: valor}}
	}
	validar := func(rows []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult) {
		t.Helper()
		prepared, result, _, err := repo.ValidateProductImport(negocio.ID, sucursal.ID, rows)
		if err != nil {
			t.Fatal(err)
		}
		return prepared, result
	}
	porFila := func(prepared []domain.ValidatedProductImportRow) map[int]domain.ValidatedProductImportRow {
		result := map[int]domain.ValidatedProductImportRow{}
		for _, item := range prepared {
			result[item.Fila] = item
		}
		return result
	}
	errorDe := func(result domain.CatalogImportResult, fila int) string {
		for _, item := range result.Errores {
			if item.Fila == fila {
				return item.Campo + ": " + item.Motivo
			}
		}
		return ""
	}

	t.Run("el código manda y sin código se busca por nombre", func(t *testing.T) {
		prepared, result := validar([]domain.ProductImportRow{
			{Fila: 2, Nombre: "Refresco cola", CodigoBarras: "7501000000024", PrecioVenta: 19, StockInicial: 5},
			{Fila: 3, Nombre: "gansito", PrecioVenta: 16, StockInicial: 4},
			{Fila: 4, Nombre: "Pan", PrecioVenta: 6, StockInicial: 1},
			{Fila: 5, Nombre: "Pan Grande", PrecioVenta: 10, StockInicial: 1},
		})
		filas := porFila(prepared)
		if filas[2].Accion != domain.ImportAccionActualizar || filas[3].Accion != domain.ImportAccionActualizar || filas[5].Accion != domain.ImportAccionActualizar {
			t.Fatalf("prepared = %+v, errores = %+v", prepared, result.Errores)
		}
		if _, ok := filas[4]; ok || !strings.Contains(errorDe(result, 4), "varios productos con ese nombre") {
			t.Fatalf("Pan ambiguo: prepared = %+v, errores = %+v", prepared, result.Errores)
		}
		aviso := false
		for _, item := range result.Advertencias {
			aviso = aviso || (item.Fila == 2 && strings.Contains(item.Motivo, "«Coca-Cola»"))
		}
		if !aviso {
			t.Fatalf("advertencias = %+v, want aviso de nombre distinto", result.Advertencias)
		}
	})

	t.Run("nombre existente con otro código pide variante", func(t *testing.T) {
		_, result := validar([]domain.ProductImportRow{{Fila: 2, Nombre: "Gansito", CodigoBarras: "7501000000031", PrecioVenta: 15, StockInicial: 1}})
		if !strings.HasPrefix(errorDe(result, 2), "Variantes:") {
			t.Fatalf("errores = %+v", result.Errores)
		}
	})

	t.Run("una variante nueva se agrega a la familia existente aunque tenga marca y presentación", func(t *testing.T) {
		familias, productos := contar("familias_producto"), contar("productos")
		prepared, result := validar([]domain.ProductImportRow{
			{Fila: 2, Nombre: "Playera", PrecioVenta: 99, StockInicial: 3, Variantes: talla("G")},
			{Fila: 3, Nombre: "Playera", PrecioVenta: 120, StockInicial: 2, Variantes: talla("M")},
		})
		filas := porFila(prepared)
		if len(result.Errores) != 0 || filas[2].Accion != "" || filas[2].Input.FamiliaClave == "" || filas[3].Accion != domain.ImportAccionActualizar {
			t.Fatalf("prepared = %+v, errores = %+v", prepared, result.Errores)
		}
		created, err := repo.CreateImportedProducts(negocio.ID, prepared)
		if err != nil || created.Creadas != 1 || created.ProductosBaseReutilizados != 1 || len(created.Errores) != 0 {
			t.Fatalf("resultado = %+v, err = %v", created, err)
		}
		if contar("familias_producto") != familias || contar("productos") != productos {
			t.Fatalf("se duplicó la familia: familias %d->%d, productos %d->%d", familias, contar("familias_producto"), productos, contar("productos"))
		}
		var variantes int64
		if err := db.Table("producto_variantes").Where("activo = TRUE").Count(&variantes).Error; err != nil || variantes != 2 {
			t.Fatalf("variantes = %d, err = %v", variantes, err)
		}
	})

	t.Run("un producto nuevo no guarda presentación ni marca", func(t *testing.T) {
		prepared, result := validar([]domain.ProductImportRow{{Fila: 2, Nombre: "Corona Mega", CodigoBarras: "7501000000048", PrecioVenta: 45, StockInicial: 12}})
		if len(prepared) != 1 || prepared[0].Accion != "" || len(result.Errores) != 0 {
			t.Fatalf("prepared = %+v, errores = %+v", prepared, result.Errores)
		}
		if _, err := repo.CreateImportedProducts(negocio.ID, prepared); err != nil {
			t.Fatal(err)
		}
		var producto domain.Producto
		if err := db.Where("negocio_id = ? AND nombre = ?", negocio.ID, "Corona Mega").First(&producto).Error; err != nil {
			t.Fatal(err)
		}
		if producto.Presentacion != nil || producto.MarcaID != nil || producto.Descripcion != nil || producto.Contenido != nil {
			t.Fatalf("producto = %+v", producto)
		}
	})
}
