package database

import (
	"os"
	"strings"
	"testing"

	compraapp "tienda/backend/internal/application/compra"
	"tienda/backend/internal/domain"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	"tienda/backend/internal/infrastructure"
	comprainfra "tienda/backend/internal/infrastructure/compra"

	"github.com/google/uuid"
)

// TestCompraProveedoresVarios covers capability 27: Registrar compra lists every
// product of the business, links the purchased product to the provider and a
// product may have several providers with a single principal.
func TestCompraProveedoresVarios(t *testing.T) {
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
	products := infrastructure.NewProductRepository(db)
	compras := comprainfra.NewRepository(db)

	negocio := negociodomain.Negocio{ID: uuid.New(), Slug: "compras", NombreComercial: "Compras", Nombre: "Compras", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	if err := db.Create(&negocio).Error; err != nil {
		t.Fatal(err)
	}
	usuario := cuentadomain.Usuario{ID: uuid.New(), Correo: "compras@example.com", HashContrasena: "no-usada-en-prueba", Estado: "activo"}
	if err := db.Create(&usuario).Error; err != nil {
		t.Fatal(err)
	}
	sucursal := negociodomain.Sucursal{ID: uuid.New(), NegocioID: negocio.ID, Codigo: "SUC-001", Nombre: "A", EsPrincipal: true, Activo: true}
	if err := db.Create(&sucursal).Error; err != nil {
		t.Fatal(err)
	}
	if err := products.CreateCategory(negocio.ID, domain.CreateCategoriaInput{Nombre: "General"}); err != nil {
		t.Fatal(err)
	}
	categorias, _ := products.ListCategoriesAdmin(negocio.ID)
	categoria := categorias[0].ID
	caja := domain.UnidadMedida{ID: uuid.New(), NegocioID: negocio.ID, Nombre: "Caja 12", Simbolo: "CJ12", Tipo: "EMPAQUE", FactorABase: 12, Activo: true}
	if err := db.Create(&caja).Error; err != nil {
		t.Fatal(err)
	}
	proveedorA := domain.Proveedor{ID: uuid.New(), NegocioID: negocio.ID, Nombre: "Proveedor A", Activo: true}
	proveedorB := domain.Proveedor{ID: uuid.New(), NegocioID: negocio.ID, Nombre: "Proveedor B", Activo: true}
	for _, proveedor := range []*domain.Proveedor{&proveedorA, &proveedorB} {
		if err := db.Create(proveedor).Error; err != nil {
			t.Fatal(err)
		}
	}

	rows := []domain.ValidatedProductImportRow{
		{Fila: 2, Input: domain.CreateImportedProductInput{Nombre: "Arroz", CategoriaID: &categoria, SucursalID: sucursal.ID, PrecioVenta: 30, CostoProveedor: 20}},
		{Fila: 3, Input: domain.CreateImportedProductInput{Nombre: "Zanahoria", CategoriaID: &categoria, SucursalID: sucursal.ID, PrecioVenta: 15, CostoProveedor: 8, ProveedorID: &proveedorA.ID}},
	}
	if res, err := products.CreateImportedProducts(negocio.ID, rows); err != nil || res.Creadas != 2 {
		t.Fatalf("importación inicial: %+v %v", res, err)
	}
	oferta := func(nombre string) (uuid.UUID, uuid.UUID) {
		var row struct{ OfertaID, ProductoID uuid.UUID }
		if err := db.Raw(`SELECT pn.id AS oferta_id, p.id AS producto_id FROM producto_negocio pn JOIN productos p ON p.id = pn.producto_id WHERE pn.negocio_id = ? AND p.nombre = ?`, negocio.ID, nombre).Scan(&row).Error; err != nil || row.OfertaID == uuid.Nil {
			t.Fatalf("oferta de %s: %v", nombre, err)
		}
		return row.OfertaID, row.ProductoID
	}
	arroz, arrozProducto := oferta("Arroz")
	zanahoria, _ := oferta("Zanahoria")

	vinculos := func(ofertaID uuid.UUID) map[uuid.UUID]bool {
		var items []domain.ProductoProveedor
		if err := db.Where("producto_negocio_id = ?", ofertaID).Find(&items).Error; err != nil {
			t.Fatal(err)
		}
		result := map[uuid.UUID]bool{}
		for _, item := range items {
			result[item.ProveedorID] = item.EsPrincipal
		}
		return result
	}
	listado := func(proveedorID uuid.UUID) []compraapp.ProductoComprable {
		items, err := compras.ListProductos(negocio.ID, proveedorID)
		if err != nil {
			t.Fatal(err)
		}
		return items
	}
	folio := 0
	comprar := func(proveedorID, ofertaID uuid.UUID) error {
		folio++
		_, err := compras.Create(negocio.ID, usuario.ID, compraapp.CrearCompraInput{
			ProveedorID: proveedorID.String(), SucursalID: sucursal.ID.String(), TipoDocumento: "TICKET",
			FolioDocumento: "F-" + string(rune('0'+folio)), FechaDocumento: "2026-10-22", FechaRecepcion: "2026-10-22", FormaPago: "EFECTIVO",
			Detalles: []compraapp.CrearDetalleInput{{ProductoNegocioID: ofertaID.String(), UnidadMedidaID: caja.ID.String(), CantidadEmpaques: 1, CostoEmpaque: 120}},
		})
		return err
	}

	t.Run("lista todos los productos, primero los del proveedor", func(t *testing.T) {
		items := listado(proveedorA.ID)
		if len(items) != 2 || items[0].ID != zanahoria || !items[0].DelProveedor || items[1].ID != arroz || items[1].DelProveedor {
			t.Fatalf("listado de A = %+v", items)
		}
		items = listado(proveedorB.ID)
		if len(items) != 2 || items[0].DelProveedor || items[1].DelProveedor {
			t.Fatalf("listado de B = %+v", items)
		}
	})

	t.Run("comprar un producto sin proveedor lo liga como principal", func(t *testing.T) {
		if err := comprar(proveedorA.ID, arroz); err != nil {
			t.Fatal(err)
		}
		if got := vinculos(arroz); len(got) != 1 || !got[proveedorA.ID] {
			t.Fatalf("vínculos del arroz = %+v", got)
		}
	})

	t.Run("comprar a un segundo proveedor agrega un vínculo no principal", func(t *testing.T) {
		if err := comprar(proveedorB.ID, arroz); err != nil {
			t.Fatal(err)
		}
		got := vinculos(arroz)
		if len(got) != 2 || !got[proveedorA.ID] || got[proveedorB.ID] {
			t.Fatalf("vínculos del arroz = %+v", got)
		}
		for _, proveedorID := range []uuid.UUID{proveedorA.ID, proveedorB.ID} {
			items := listado(proveedorID)
			if items[0].ID != arroz && items[1].ID != arroz {
				t.Fatalf("el arroz no aparece en %v", items)
			}
			for _, item := range items {
				if item.ID == arroz && !item.DelProveedor {
					t.Fatalf("el arroz debe ser del proveedor %s", proveedorID)
				}
			}
		}
		if err := comprar(proveedorB.ID, arroz); err != nil {
			t.Fatalf("recompra al mismo proveedor: %v", err)
		}
		if got := vinculos(arroz); len(got) != 2 {
			t.Fatalf("la recompra duplicó el vínculo: %+v", got)
		}
	})

	t.Run("editar el principal conserva los demás proveedores", func(t *testing.T) {
		input := domain.UpdateProductInput{SucursalID: sucursal.ID, Nombre: "Arroz", CategoriaID: &categoria, PrecioVenta: 30, ProveedorID: &proveedorB.ID}
		if err := products.Update(negocio.ID, arrozProducto, input); err != nil {
			t.Fatal(err)
		}
		got := vinculos(arroz)
		if len(got) != 2 || got[proveedorA.ID] || !got[proveedorB.ID] {
			t.Fatalf("vínculos tras cambiar el principal = %+v", got)
		}
		input.ProveedorID = nil
		if err := products.Update(negocio.ID, arrozProducto, input); err != nil {
			t.Fatal(err)
		}
		got = vinculos(arroz)
		if len(got) != 2 || got[proveedorA.ID] || got[proveedorB.ID] {
			t.Fatalf("vínculos sin principal = %+v", got)
		}
		if err := comprar(proveedorA.ID, arroz); err != nil {
			t.Fatal(err)
		}
		if got := vinculos(arroz); len(got) != 2 || got[proveedorA.ID] {
			t.Fatalf("una compra a un proveedor ya ligado no debe volverlo principal: %+v", got)
		}
	})

	t.Run("la base admite varios proveedores y un solo principal", func(t *testing.T) {
		tercero := domain.Proveedor{ID: uuid.New(), NegocioID: negocio.ID, Nombre: "Proveedor C", Activo: true}
		if err := db.Create(&tercero).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&domain.ProductoProveedor{ProductoNegocioID: zanahoria, ProveedorID: tercero.ID}).Error; err != nil {
			t.Fatalf("segundo proveedor rechazado: %v", err)
		}
		if err := db.Create(&domain.ProductoProveedor{ProductoNegocioID: zanahoria, ProveedorID: proveedorB.ID, EsPrincipal: true}).Error; err == nil {
			t.Fatal("se aceptaron dos proveedores principales")
		}
	})
}
