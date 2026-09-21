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

func TestCatalogVariantsTenancyIsolation(t *testing.T) {
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
	type ctx struct {
		negocio, sucursal, categoria uuid.UUID
	}
	var ctxs []ctx
	for i, slug := range []string{"uno", "dos"} {
		n := negociodomain.Negocio{ID: uuid.New(), Slug: slug, NombreComercial: slug, Nombre: slug, CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
		if err := db.Create(&n).Error; err != nil {
			t.Fatal(err)
		}
		s := negociodomain.Sucursal{ID: uuid.New(), NegocioID: n.ID, Codigo: "SUC-001", Nombre: "P", EsPrincipal: true, Activo: true}
		if err := db.Create(&s).Error; err != nil {
			t.Fatal(err)
		}
		if err := repo.CreateCategory(n.ID, domain.CreateCategoriaInput{Nombre: "Ropa"}); err != nil {
			t.Fatal(i, err)
		}
		cats, _ := repo.ListCategoriesAdmin(n.ID)
		ctxs = append(ctxs, ctx{n.ID, s.ID, cats[0].ID})
	}
	code := "7500000000017"
	for i, c := range ctxs {
		cat := c.categoria
		rows := []domain.ValidatedProductImportRow{
			{Fila: 2, Input: domain.CreateImportedProductInput{Nombre: "Jabón", CategoriaID: &cat, SucursalID: c.sucursal, PrecioVenta: 10, StockInicial: 3, CodigoBarras: &code}},
			{Fila: 3, Input: domain.CreateImportedProductInput{Nombre: "Playera", CategoriaID: &cat, SucursalID: c.sucursal, PrecioVenta: 99, Variantes: []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "M"}}}},
			{Fila: 4, Input: domain.CreateImportedProductInput{Nombre: "Playera", CategoriaID: &cat, SucursalID: c.sucursal, PrecioVenta: 99, Variantes: []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "G"}}}},
		}
		res, err := repo.CreateImportedProducts(c.negocio, rows)
		if err != nil || res.Creadas != 3 || len(res.Errores) != 0 {
			t.Fatalf("import negocio %d: %+v %v", i, res, err)
		}
		list, err := repo.ListByBusiness(c.negocio)
		if err != nil || len(list) != 2 {
			t.Fatalf("list %d: %+v %v", i, list, err)
		}
	}
	// Update y Deactivate no deben cruzar negocios.
	list1, _ := repo.ListByBusiness(ctxs[0].negocio)
	var jabon uuid.UUID
	for _, p := range list1 {
		if p.Nombre == "Jabón" {
			jabon = p.ID
		}
	}
	upd := domain.UpdateProductInput{Nombre: "Jabón 2", SKUInterno: "SKU-X", CategoriaID: ctxs[0].categoria, PrecioVenta: 12, CodigoBarras: &code}
	if err := repo.Update(ctxs[0].negocio, jabon, upd); err != nil {
		t.Fatalf("update propio: %v", err)
	}
	upd.CategoriaID = ctxs[1].categoria
	if err := repo.Update(ctxs[0].negocio, jabon, upd); err == nil {
		t.Fatal("aceptó categoría de otro negocio")
	}
	if err := repo.Update(ctxs[1].negocio, jabon, upd); err == nil {
		t.Fatal("otro negocio editó producto ajeno")
	}
	if err := repo.Deactivate(ctxs[1].negocio, jabon); err == nil {
		t.Fatal("otro negocio desactivó producto ajeno")
	}
	job, err := repo.CreateProductImportJob(ctxs[0].negocio, ctxs[0].sucursal)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetProductImportJob(ctxs[1].negocio, job.ID); err == nil {
		t.Fatal("otro negocio leyó job ajeno")
	}
	if _, err := repo.GetProductImportJob(ctxs[0].negocio, job.ID); err != nil {
		t.Fatal(err)
	}
	// Una variante manual reutiliza la base importada del mismo negocio.
	if err := repo.Create(ctxs[0].negocio, domain.CreateProductInput{Nombre: "Playera", CategoriaID: &ctxs[0].categoria, SucursalID: ctxs[0].sucursal,
		Variantes: []domain.CreateProductVariantInput{{Atributos: []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "CH"}}, GenerarSKUInterno: true, PrecioVenta: 99}}}); err != nil {
		t.Fatalf("variante manual sobre base importada: %v", err)
	}
	if err := repo.Create(ctxs[0].negocio, domain.CreateProductInput{Nombre: "Otro", CategoriaID: &ctxs[1].categoria, SucursalID: ctxs[0].sucursal,
		Variantes: []domain.CreateProductVariantInput{{Atributos: []domain.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "CH"}}, GenerarSKUInterno: true}}}); err == nil {
		t.Fatal("variante con categoría ajena aceptada")
	}
}
