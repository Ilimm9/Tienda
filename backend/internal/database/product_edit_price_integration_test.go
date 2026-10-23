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
	"gorm.io/gorm"
)

// TestProductEditWritesBranchPrice covers capability 29: editing a product writes
// its sale price into precios_sucursal of the active branch, with history and
// the precios.autorizar rule.
func TestProductEditWritesBranchPrice(t *testing.T) {
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

	negocio := negociodomain.Negocio{ID: uuid.New(), Slug: "precios", NombreComercial: "Precios", Nombre: "Precios", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
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
	if err := repo.CreateCategory(negocio.ID, domain.CreateCategoriaInput{Nombre: "General"}); err != nil {
		t.Fatal(err)
	}
	categorias, _ := repo.ListCategoriesAdmin(negocio.ID)
	categoria := categorias[0].ID

	talla := func(valor string) []domain.ProductVariantAttributeInput {
		return []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: valor}}
	}
	rows := []domain.ValidatedProductImportRow{
		{Fila: 2, Input: domain.CreateImportedProductInput{Nombre: "Jabón", CategoriaID: &categoria, SucursalID: sucursalA.ID, PrecioVenta: 10, CostoProveedor: 5}},
		{Fila: 3, Input: domain.CreateImportedProductInput{Nombre: "Cloro", CategoriaID: &categoria, SucursalID: sucursalA.ID, PrecioVenta: 20, CostoProveedor: 8}},
		{Fila: 4, Input: domain.CreateImportedProductInput{Nombre: "Playera", CategoriaID: &categoria, SucursalID: sucursalA.ID, PrecioVenta: 99, CostoProveedor: 50, Variantes: talla("M")}},
		{Fila: 5, Input: domain.CreateImportedProductInput{Nombre: "Playera", CategoriaID: &categoria, SucursalID: sucursalA.ID, PrecioVenta: 99, CostoProveedor: 50, Variantes: talla("G")}},
	}
	if res, err := repo.CreateImportedProducts(negocio.ID, rows); err != nil || res.Creadas != 4 {
		t.Fatalf("importación inicial: %+v %v", res, err)
	}
	productos, err := repo.ListByBusiness(negocio.ID, sucursalA.ID)
	if err != nil {
		t.Fatal(err)
	}
	porNombre := map[string]domain.ProductRow{}
	for _, producto := range productos {
		porNombre[producto.Nombre] = producto
	}
	jabon, cloro, playera := porNombre["Jabón"], porNombre["Cloro"], porNombre["Playera"]

	ofertaSimple := func(productoID uuid.UUID) uuid.UUID {
		var oferta domain.ProductoNegocio
		if err := db.Where("negocio_id = ? AND producto_id = ? AND activo = TRUE", negocio.ID, productoID).First(&oferta).Error; err != nil {
			t.Fatalf("oferta del producto: %v", err)
		}
		return oferta.ID
	}
	ofertaVariante := func(varianteID uuid.UUID) uuid.UUID {
		var oferta domain.ProductoNegocio
		if err := db.Where("negocio_id = ? AND producto_variante_id = ? AND activo = TRUE", negocio.ID, varianteID).First(&oferta).Error; err != nil {
			t.Fatalf("oferta de la variante: %v", err)
		}
		return oferta.ID
	}
	precio := func(sucursalID, ofertaID uuid.UUID) (preciodomain.PrecioSucursal, bool) {
		var actual preciodomain.PrecioSucursal
		err := db.Where("sucursal_id = ? AND producto_negocio_id = ?", sucursalID, ofertaID).First(&actual).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return actual, false
		}
		if err != nil {
			t.Fatal(err)
		}
		return actual, true
	}
	eventos := func(ofertaID uuid.UUID, origen string) []preciodomain.HistorialCostoPrecioSucursal {
		var items []preciodomain.HistorialCostoPrecioSucursal
		if err := db.Where("producto_negocio_id = ? AND origen = ?", ofertaID, origen).Order("creado_en ASC").Find(&items).Error; err != nil {
			t.Fatal(err)
		}
		return items
	}
	usuario := uuid.New()
	edicion := func(sucursalID uuid.UUID, nombre string, precioVenta float64, puede bool) domain.UpdateProductInput {
		return domain.UpdateProductInput{SucursalID: sucursalID, UsuarioID: usuario, PuedeCambiarPrecio: puede, Nombre: nombre, CategoriaID: &categoria, PrecioVenta: precioVenta}
	}

	// La sucursal B tiene su propio precio del jabón; editar en A no debe tocarlo.
	jabonOferta := ofertaSimple(jabon.ID)
	if err := db.Create(&preciodomain.PrecioSucursal{NegocioID: negocio.ID, SucursalID: sucursalB.ID, ProductoNegocioID: jabonOferta, CostoVigente: 6, PrecioVenta: 25}).Error; err != nil {
		t.Fatal(err)
	}

	t.Run("cambio de precio con permiso", func(t *testing.T) {
		if err := repo.Update(negocio.ID, jabon.ID, edicion(sucursalA.ID, "Jabón", 12.345, true)); err != nil {
			t.Fatal(err)
		}
		actual, _ := precio(sucursalA.ID, jabonOferta)
		if actual.PrecioVenta != 12.35 || actual.Margen == nil || *actual.Margen < 1.46 || *actual.Margen > 1.48 {
			t.Fatalf("precio en A = %+v", actual)
		}
		ajustes := eventos(jabonOferta, preciodomain.OrigenAjustePrecioManual)
		if len(ajustes) != 1 || *ajustes[0].PrecioAnterior != 10 || *ajustes[0].PrecioNuevo != 12.35 || *ajustes[0].Motivo != "Edición del producto" || ajustes[0].UsuarioID == nil || *ajustes[0].UsuarioID != usuario {
			t.Fatalf("evento de ajuste = %+v", ajustes)
		}
		otra, _ := precio(sucursalB.ID, jabonOferta)
		if otra.PrecioVenta != 25 {
			t.Fatalf("la sucursal B cambió: %+v", otra)
		}
	})

	t.Run("mismo precio no crea evento", func(t *testing.T) {
		input := edicion(sucursalA.ID, "Jabón", 12.35, true)
		motivo := "No debe guardarse"
		input.MotivoPrecio = &motivo
		if err := repo.Update(negocio.ID, jabon.ID, input); err != nil {
			t.Fatal(err)
		}
		if ajustes := eventos(jabonOferta, preciodomain.OrigenAjustePrecioManual); len(ajustes) != 1 {
			t.Fatalf("eventos = %d, se esperaba 1", len(ajustes))
		}
	})

	t.Run("motivo capturado", func(t *testing.T) {
		input := edicion(sucursalA.ID, "Jabón", 13, true)
		motivo := "  Promoción  "
		input.MotivoPrecio = &motivo
		if err := repo.Update(negocio.ID, jabon.ID, input); err != nil {
			t.Fatal(err)
		}
		ajustes := eventos(jabonOferta, preciodomain.OrigenAjustePrecioManual)
		if len(ajustes) != 2 || *ajustes[1].Motivo != "Promoción" {
			t.Fatalf("eventos = %+v", ajustes)
		}
	})

	t.Run("sin permiso no cambia el precio", func(t *testing.T) {
		err := repo.Update(negocio.ID, jabon.ID, edicion(sucursalA.ID, "Jabón azul", 15, false))
		if !errors.Is(err, domain.ErrPricePermissionRequired) {
			t.Fatalf("err = %v", err)
		}
		actual, _ := precio(sucursalA.ID, jabonOferta)
		var producto domain.Producto
		db.First(&producto, "id = ?", jabon.ID)
		if actual.PrecioVenta != 13 || producto.Nombre != "Jabón" {
			t.Fatalf("la edición rechazada dejó cambios: precio %v, nombre %q", actual.PrecioVenta, producto.Nombre)
		}
		if err := repo.Update(negocio.ID, jabon.ID, edicion(sucursalA.ID, "Jabón azul", 13, false)); err != nil {
			t.Fatalf("sin permiso y sin cambio de precio: %v", err)
		}
	})

	t.Run("producto sin precio en la sucursal", func(t *testing.T) {
		cloroOferta := ofertaSimple(cloro.ID)
		if err := repo.Update(negocio.ID, cloro.ID, edicion(sucursalB.ID, "Cloro", 0, false)); err != nil {
			t.Fatal(err)
		}
		if _, existe := precio(sucursalB.ID, cloroOferta); existe {
			t.Fatal("se creó un precio en cero que nadie capturó")
		}
		if err := repo.Update(negocio.ID, cloro.ID, edicion(sucursalB.ID, "Cloro", 30, false)); !errors.Is(err, domain.ErrPricePermissionRequired) {
			t.Fatalf("sin permiso: %v", err)
		}
		if err := repo.Update(negocio.ID, cloro.ID, edicion(sucursalB.ID, "Cloro", 30, true)); err != nil {
			t.Fatal(err)
		}
		actual, existe := precio(sucursalB.ID, cloroOferta)
		if !existe || actual.PrecioVenta != 30 || actual.CostoVigente != 0 {
			t.Fatalf("precio creado = %+v", actual)
		}
		if altas := eventos(cloroOferta, preciodomain.OrigenAltaInicial); len(altas) != 2 {
			t.Fatalf("altas iniciales = %d, se esperaban 2 (A por importación y B por edición)", len(altas))
		}
		original, _ := precio(sucursalA.ID, cloroOferta)
		if original.PrecioVenta != 20 {
			t.Fatalf("la sucursal A cambió: %+v", original)
		}
	})

	t.Run("variantes existentes y nuevas", func(t *testing.T) {
		var m, g uuid.UUID
		for _, variante := range playera.Variantes {
			switch variante.Atributos[0].Valor {
			case "M":
				m = variante.ID
			case "G":
				g = variante.ID
			}
		}
		input := edicion(sucursalA.ID, "Playera", 0, true)
		input.Variantes = []domain.UpdateProductVariantInput{
			{ID: &m, Atributos: talla("M"), PrecioVenta: 105},
			{ID: &g, Atributos: talla("G"), PrecioVenta: 99},
			{Atributos: talla("CH"), PrecioVenta: 110},
			{Atributos: talla("XL"), PrecioVenta: 0},
		}
		sinPermiso := input
		sinPermiso.PuedeCambiarPrecio = false
		sinPermiso.Variantes = []domain.UpdateProductVariantInput{input.Variantes[0], input.Variantes[1], input.Variantes[2]}
		sinPermiso.Variantes[0].PrecioVenta = 99
		if err := repo.Update(negocio.ID, playera.ID, sinPermiso); !errors.Is(err, domain.ErrPricePermissionRequired) {
			t.Fatalf("variante nueva con precio sin permiso: %v", err)
		}
		if err := repo.Update(negocio.ID, playera.ID, input); err != nil {
			t.Fatal(err)
		}
		mOferta, gOferta := ofertaVariante(m), ofertaVariante(g)
		if actual, _ := precio(sucursalA.ID, mOferta); actual.PrecioVenta != 105 {
			t.Fatalf("precio M = %v", actual.PrecioVenta)
		}
		if len(eventos(mOferta, preciodomain.OrigenAjustePrecioManual)) != 1 || len(eventos(gOferta, preciodomain.OrigenAjustePrecioManual)) != 0 {
			t.Fatal("eventos de variantes existentes incorrectos")
		}
		var nuevas []struct {
			Valor    string
			OfertaID uuid.UUID
		}
		if err := db.Table("producto_variantes AS pv").
			Select("pva.valor, pn.id AS oferta_id").
			Joins("JOIN producto_variante_atributos pva ON pva.producto_variante_id = pv.id").
			Joins("JOIN producto_negocio pn ON pn.producto_variante_id = pv.id AND pn.activo = TRUE").
			Where("pv.producto_id = ? AND pva.valor IN ?", playera.ID, []string{"CH", "XL"}).
			Scan(&nuevas).Error; err != nil || len(nuevas) != 2 {
			t.Fatalf("variantes nuevas: %+v %v", nuevas, err)
		}
		for _, nueva := range nuevas {
			actual, existe := precio(sucursalA.ID, nueva.OfertaID)
			switch nueva.Valor {
			case "CH":
				if !existe || actual.PrecioVenta != 110 || len(eventos(nueva.OfertaID, preciodomain.OrigenAltaInicial)) != 1 {
					t.Fatalf("variante CH = %+v existe=%v", actual, existe)
				}
			case "XL":
				if existe {
					t.Fatal("la variante XL sin precio capturado obtuvo precio")
				}
			}
		}
	})

	t.Run("sucursal ajena o motivo largo", func(t *testing.T) {
		if err := repo.Update(negocio.ID, jabon.ID, edicion(uuid.New(), "Jabón", 13, true)); err == nil {
			t.Fatal("aceptó una sucursal que no es del negocio")
		}
		input := edicion(sucursalA.ID, "Jabón", 14, true)
		motivo := strings.Repeat("a", 201)
		input.MotivoPrecio = &motivo
		if err := repo.Update(negocio.ID, jabon.ID, input); err == nil {
			t.Fatal("aceptó un motivo de más de 200 caracteres")
		}
	})
}
