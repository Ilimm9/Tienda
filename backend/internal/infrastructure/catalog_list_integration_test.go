package infrastructure

import (
	"errors"
	"os"
	"sort"
	"strings"
	"testing"

	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requiere una base con sufijo `_test` ya migrada por la API (database.Init).
// Cada prueba crea su propio negocio.
func listaTestDB(t *testing.T) (*gorm.DB, uuid.UUID) {
	t.Helper()
	url := os.Getenv("LISTA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LISTA_TEST_DATABASE_URL no configurada")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir base temporal: %v", err)
	}
	var name string
	if err := db.Raw("SELECT current_database()").Scan(&name).Error; err != nil || !strings.HasSuffix(name, "_test") {
		t.Fatalf("se requiere una base con sufijo _test; actual=%q err=%v", name, err)
	}
	id := uuid.New()
	if err := db.Exec("INSERT INTO negocios (id, slug, nombre_comercial, nombre) VALUES (?, ?, 'Prueba', 'Prueba')", id, "prueba-"+id.String()).Error; err != nil {
		t.Fatalf("crear negocio: %v", err)
	}
	return db, id
}

// rutas devuelve «A > B» de cada categoría del negocio, ordenadas.
func rutas(t *testing.T, db *gorm.DB, negocioID uuid.UUID) []string {
	t.Helper()
	var all []domain.Categoria
	if err := db.Where("negocio_id = ?", negocioID).Find(&all).Error; err != nil {
		t.Fatal(err)
	}
	byID := map[uuid.UUID]domain.Categoria{}
	for _, c := range all {
		byID[c.ID] = c
	}
	out := make([]string, 0, len(all))
	for _, c := range all {
		path := []string{c.Nombre}
		for p := c.CategoriaPadreID; p != nil; p = byID[*p].CategoriaPadreID {
			path = append([]string{byID[*p].Nombre}, path...)
		}
		out = append(out, strings.Join(path, " > "))
	}
	sort.Strings(out)
	return out
}

func motivos(result domain.CatalogImportResult) map[int]string {
	out := map[int]string{}
	for _, issue := range result.Errores {
		out[issue.Fila] = strings.TrimSpace(issue.Campo + " " + issue.Motivo)
	}
	return out
}

func TestImportCategoryPathsCreaAncestrosYOmiteExistentes(t *testing.T) {
	db, negocio := listaTestDB(t)
	repo := NewProductRepository(db)
	if err := repo.CreateCategory(negocio, domain.CreateCategoriaInput{Nombre: "Bebidas"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateCategory(negocio, domain.CreateCategoriaInput{Nombre: "Lácteos"}); err != nil {
		t.Fatal(err)
	}

	result, err := repo.ImportCategoryPaths(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 1, Ruta: []string{"Bebidas", "Frías", "Agua"}},           // crea Frías (ancestro) y Agua
		{Fila: 2, Ruta: []string{"bebidas", "FRIAS"}},                   // creada como ancestro en el renglón 1
		{Fila: 3, Ruta: []string{"Bebidas"}},                            // ya existía
		{Fila: 4, Ruta: []string{"Lácteos", "Frías"}},                   // Frías ya se creó en el renglón 1: se omite
		{Fila: 5, Ruta: []string{"Lacteos", "Frías", "Leche"}},          // Frías está en otro lado: inválido
		{Fila: 6, Ruta: []string{"Bebidas", "Frías", "Agua"}},           // repetida
		{Fila: 7, Ruta: []string{"Bebidas", "Frías  ", "Agua mineral"}}, // espacios dobles: misma Frías
		{Fila: 8, Ruta: []string{"Comida", "Postres"}},                  // ancestro nuevo en la raíz
		{Fila: 9, Ruta: []string{"Frías", "Hielo"}},                     // el primer nombre puede estar en cualquier nivel
		{Fila: 10, Ruta: []string{"Uno", "Dos", "Uno"}},                 // nombre repetido en la ruta
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Bebidas", "Bebidas > Frías", "Bebidas > Frías > Agua", "Bebidas > Frías > Agua mineral", "Bebidas > Frías > Hielo",
		"Comida", "Comida > Postres", "Lácteos",
	}
	if got := rutas(t, db, negocio); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("rutas = %v\nwant  = %v", got, want)
	}
	if result.Creadas != 6 || result.Omitidas != 4 || result.Invalidas != 2 {
		t.Fatalf("result = %#v", result)
	}
	m := motivos(result)
	if m[2] != "La categoría está repetida en la lista" || m[3] != "La categoría ya existe como «Bebidas»" ||
		m[4] != "La categoría está repetida en la lista" || m[5] != "Nombre «Frías» ya existe como «Bebidas › Frías»" ||
		m[6] != "La categoría está repetida en la lista" || !strings.Contains(m[10], "dos veces") {
		t.Fatalf("motivos = %#v", m)
	}
	// El renglón inválido no dejó ancestros sueltos.
	var leche int64
	db.Model(&domain.Categoria{}).Where("negocio_id = ? AND nombre = 'Leche'", negocio).Count(&leche)
	if leche != 0 {
		t.Fatal("no debió crear Leche")
	}
}

func TestImportCategoriesXLSXResuelvePadresPorNombreORuta(t *testing.T) {
	db, negocio := listaTestDB(t)
	repo := NewProductRepository(db)
	if _, err := repo.ImportCategoryPaths(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 1, Ruta: []string{"Bebidas", "Frías"}},
	}); err != nil {
		t.Fatal(err)
	}

	result, err := repo.ImportCategories(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 2, Nombre: "Hielo", CategoriaPadre: "Frías"},          // por nombre
		{Fila: 3, Nombre: "Agua", CategoriaPadre: "Bebidas > Frías"}, // por ruta
		{Fila: 4, Nombre: "Postres", CategoriaPadre: "Dulces"},       // padre más abajo en el archivo
		{Fila: 5, Nombre: "Dulces"},                                  // raíz
		{Fila: 6, Nombre: "frias", CategoriaPadre: "Dulces"},         // ya existe (en Bebidas)
		{Fila: 7, Nombre: "Huérfana", CategoriaPadre: "No existe"},   // padre inexistente
		{Fila: 8, Nombre: "A", CategoriaPadre: "B"},                  // circular
		{Fila: 9, Nombre: "B", CategoriaPadre: "A"},                  // circular
		{Fila: 10, Nombre: "Postres", CategoriaPadre: "Dulces"},      // repetida en el archivo
		{Fila: 11, Nombre: "Sal", CategoriaPadre: "Dulces > Frías"},  // ruta que no coincide
		{Fila: 12, Nombre: "Eco", CategoriaPadre: "Eco"},             // ella misma
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Bebidas", "Bebidas > Frías", "Bebidas > Frías > Agua", "Bebidas > Frías > Hielo",
		"Dulces", "Dulces > Postres",
	}
	if got := rutas(t, db, negocio); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("rutas = %v\nwant  = %v", got, want)
	}
	m := motivos(result)
	if m[6] != "La categoría ya existe como «Bebidas › Frías»" || m[10] != "La categoría está repetida en el archivo" ||
		!strings.Contains(m[7], "no existe") || !strings.Contains(m[8], "circular") || !strings.Contains(m[9], "circular") ||
		!strings.Contains(m[11], "no existe") || !strings.Contains(m[12], "misma") {
		t.Fatalf("motivos = %#v", m)
	}
	if result.Creadas != 4 || result.Omitidas != 2 || result.Invalidas != 5 {
		t.Fatalf("result = %#v", result)
	}
}

// El primer renglón falla por padre inexistente; el de abajo con el mismo
// nombre y padre válido sí se crea.
func TestImportCategoriesXLSXNoArrastraAlRepetidoValido(t *testing.T) {
	db, negocio := listaTestDB(t)
	repo := NewProductRepository(db)
	result, err := repo.ImportCategories(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 2, Nombre: "Postres", CategoriaPadre: "No existe"},
		{Fila: 3, Nombre: "Postres", CategoriaPadre: "Dulces"},
		{Fila: 4, Nombre: "Dulces"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := rutas(t, db, negocio); strings.Join(got, "|") != "Dulces|Dulces > Postres" {
		t.Fatalf("rutas = %v; result = %#v", got, result)
	}
	if m := motivos(result); len(m) != 1 || !strings.Contains(m[2], "no existe") {
		t.Fatalf("motivos = %#v", m)
	}
}

func TestCategoriaIndividualNombreUnico(t *testing.T) {
	db, negocio := listaTestDB(t)
	repo := NewProductRepository(db)
	if _, err := repo.ImportCategoryPaths(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 1, Ruta: []string{"Bebidas", "Frías"}},
		{Fila: 2, Ruta: []string{"Lácteos"}},
	}); err != nil {
		t.Fatal(err)
	}
	var bebidas, lacteos, frias domain.Categoria
	db.Where("negocio_id = ? AND nombre = 'Bebidas'", negocio).First(&bebidas)
	db.Where("negocio_id = ? AND nombre = 'Lácteos'", negocio).First(&lacteos)
	db.Where("negocio_id = ? AND nombre = 'Frías'", negocio).First(&frias)

	for _, nombre := range []string{" frias ", "FRÍAS", "Fr\u00edas", "Frías  "} {
		if err := repo.CreateCategory(negocio, domain.CreateCategoriaInput{Nombre: nombre, CategoriaPadreID: &lacteos.ID}); !errors.Is(err, domain.ErrCategoriaDuplicada) {
			t.Fatalf("%q: err = %v", nombre, err)
		}
	}
	if err := repo.CreateCategory(negocio, domain.CreateCategoriaInput{Nombre: "Yogur", CategoriaPadreID: &lacteos.ID}); err != nil {
		t.Fatalf("nombre nuevo: err = %v", err)
	}
	// Mover sin renombrar no choca consigo misma.
	if err := repo.UpdateCategory(negocio, frias.ID, domain.UpdateCategoriaInput{CategoriaPadreID: &lacteos.ID}); err != nil {
		t.Fatalf("mover: err = %v", err)
	}
	nombre := "FRÍAS"
	if err := repo.UpdateCategory(negocio, frias.ID, domain.UpdateCategoriaInput{Nombre: &nombre}); err != nil {
		t.Fatalf("renombrar a sí misma: err = %v", err)
	}
	otro := "lacteos"
	if err := repo.UpdateCategory(negocio, bebidas.ID, domain.UpdateCategoriaInput{Nombre: &otro}); !errors.Is(err, domain.ErrCategoriaDuplicada) {
		t.Fatalf("renombrar a otra existente: err = %v", err)
	}
}

func TestImportListaOmiteYSigueEnUnaTransaccion(t *testing.T) {
	db, negocio := listaTestDB(t)
	repo := NewProductRepository(db)
	if err := repo.CreateUnit(negocio, domain.CreateUnidadMedidaInput{Nombre: "Kilogramo", Simbolo: "kg", Tipo: "PESO", FactorABase: 1, Decimales: 3}); err != nil {
		t.Fatal(err)
	}
	result, err := repo.ImportUnits(negocio, []domain.CatalogImportUnitRow{
		{Fila: 1, Nombre: "Gramo", Simbolo: "g", Tipo: "PESO", FactorABase: 0.001, Decimales: 0},
		{Fila: 2, Nombre: "Kilo", Simbolo: "KG", Tipo: "PESO", FactorABase: 1},       // símbolo existente
		{Fila: 3, Nombre: "kilogramo", Simbolo: "kgr", Tipo: "PESO", FactorABase: 1}, // nombre existente
		{Fila: 4, Nombre: "Litro", Simbolo: "l", Tipo: "VOLUMEN", FactorABase: 1},
	})
	if err != nil || result.Creadas != 2 || result.Omitidas != 2 {
		t.Fatalf("result = %#v err = %v", result, err)
	}

	// Una falla de base de datos a la mitad no deja nada guardado.
	var antes int64
	db.Model(&domain.Marca{}).Where("negocio_id = ?", negocio).Count(&antes)
	_, err = repo.ImportBrands(negocio, []domain.CatalogImportBrandRow{
		{Fila: 1, Nombre: "Bimbo"},
		{Fila: 2, Nombre: strings.Repeat("x", 200)}, // excede varchar(180): la base rechaza el lote
	})
	if err == nil {
		t.Fatal("se esperaba error de base de datos")
	}
	var despues int64
	db.Model(&domain.Marca{}).Where("negocio_id = ?", negocio).Count(&despues)
	if antes != despues {
		t.Fatalf("rollback fallido: antes=%d después=%d", antes, despues)
	}

	// Lo mismo con categorías: el ancestro creado se revierte.
	_, err = repo.ImportCategoryPaths(negocio, []domain.CatalogImportCategoryRow{
		{Fila: 1, Ruta: []string{"Nueva", "Hija"}},
		{Fila: 2, Ruta: []string{"Nueva", strings.Repeat("y", 200)}},
	})
	if err == nil {
		t.Fatal("se esperaba error de base de datos en categorías")
	}
	if got := rutas(t, db, negocio); len(got) != 0 {
		t.Fatalf("rollback de categorías fallido: %v", got)
	}
}
