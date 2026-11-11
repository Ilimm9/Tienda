package application

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
)

// listRepositoryStub registra lo que llega al importador; el resto de la
// interfaz no se usa.
type listRepositoryStub struct {
	ProductRepository
	units      []domain.CatalogImportUnitRow
	providers  []domain.CatalogImportProviderRow
	categories []domain.CatalogImportCategoryRow
	brands     []domain.CatalogImportBrandRow
}

func (s *listRepositoryStub) ImportBrands(_ uuid.UUID, rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error) {
	s.brands = rows
	return domain.CatalogImportResult{Creadas: len(rows)}, nil
}
func (s *listRepositoryStub) ImportUnits(_ uuid.UUID, rows []domain.CatalogImportUnitRow) (domain.CatalogImportResult, error) {
	s.units = rows
	return domain.CatalogImportResult{Creadas: len(rows)}, nil
}
func (s *listRepositoryStub) ImportProviders(_ uuid.UUID, rows []domain.CatalogImportProviderRow) (domain.CatalogImportResult, error) {
	s.providers = rows
	return domain.CatalogImportResult{Creadas: len(rows)}, nil
}
func (s *listRepositoryStub) ImportCategoryPaths(_ uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	s.categories = rows
	return domain.CatalogImportResult{Creadas: len(rows)}, nil
}

func TestImportUnitListValidatesAndNormalizes(t *testing.T) {
	repo := &listRepositoryStub{}
	service := NewProductService(repo, nil)
	result, err := service.ImportUnitList(uuid.New(), []domain.CatalogImportUnitRow{
		{Fila: 1, Nombre: " Kilogramo ", Simbolo: "kg", Tipo: "peso", FactorABase: 1000, Decimales: 3},
		{Fila: 2, Nombre: "Metro cuadrado", Simbolo: "m2", Tipo: "Área", FactorABase: 1, Decimales: 2},
		{Fila: 3, Nombre: "Pieza", Simbolo: "pz", Tipo: "UNIDAD", FactorABase: 1},
		{Fila: 4, Nombre: "Caja", Simbolo: "cja", Tipo: "EMPAQUE", FactorABase: 0},
		{Fila: 5, Nombre: "Gramo", Simbolo: "", Tipo: "PESO", FactorABase: 1},
		{Fila: 6, Nombre: "Litro", Simbolo: "l", Tipo: "VOLUMEN", FactorABase: 1, Decimales: 7},
		{Fila: 7, Nombre: "Largo", Simbolo: strings.Repeat("x", 21), Tipo: "LONGITUD", FactorABase: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.units) != 2 || repo.units[0].Tipo != "PESO" || repo.units[0].Nombre != "Kilogramo" || repo.units[1].Tipo != "AREA" {
		t.Fatalf("válidas = %#v", repo.units)
	}
	if result.Procesadas != 7 || result.Creadas != 2 || result.Invalidas != 5 || len(result.Errores) != 5 {
		t.Fatalf("result = %#v", result)
	}
	campos := []string{}
	for _, issue := range result.Errores {
		campos = append(campos, issue.Campo)
	}
	if strings.Join(campos, ",") != "Tipo,Factor a base,Símbolo,Decimales,Símbolo" {
		t.Fatalf("campos = %v", campos)
	}
}

func TestImportProviderListValidatesRFCAndEmail(t *testing.T) {
	repo := &listRepositoryStub{}
	service := NewProductService(repo, nil)
	result, err := service.ImportProviderList(uuid.New(), []domain.CatalogImportProviderRow{
		{Fila: 1, Nombre: "Sigma", RFC: "sig010101ab1", Email: "ventas@sigma.mx"},
		{Fila: 2, Nombre: "Sin RFC"},
		{Fila: 3, Nombre: "RFC malo", RFC: "XYZ"},
		{Fila: 4, Nombre: "Correo malo", Email: "no-es-correo"},
		{Fila: 5, Nombre: ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.providers) != 2 || repo.providers[0].RFC != "SIG010101AB1" || repo.providers[1].RFC != "" {
		t.Fatalf("válidos = %#v", repo.providers)
	}
	if result.Invalidas != 3 || result.Errores[0].Fila != 3 || result.Errores[0].Campo != "RFC" || result.Errores[1].Campo != "Correo" || result.Errores[2].Campo != "Nombre" {
		t.Fatalf("result = %#v", result)
	}
}

func TestImportCategoryListCleansPaths(t *testing.T) {
	repo := &listRepositoryStub{}
	service := NewProductService(repo, nil)
	result, err := service.ImportCategoryList(uuid.New(), []domain.CatalogImportCategoryRow{
		{Fila: 1, Ruta: []string{" Bebidas ", " Frías "}},
		{Fila: 2, Ruta: []string{"Bebidas", ""}},
		{Fila: 3, Ruta: nil},
		{Fila: 4, Ruta: []string{strings.Repeat("a", 181)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repo.categories) != 1 || repo.categories[0].Nombre != "Frías" || strings.Join(repo.categories[0].Ruta, "|") != "Bebidas|Frías" {
		t.Fatalf("válidas = %#v", repo.categories)
	}
	if result.Invalidas != 3 {
		t.Fatalf("result = %#v", result)
	}
}

func TestImportListRejectsEmptyAndTooLarge(t *testing.T) {
	service := NewProductService(&listRepositoryStub{}, nil)
	if _, err := service.ImportBrandList(uuid.New(), nil); !errors.Is(err, ErrCatalogListEmpty) {
		t.Fatalf("vacía: err = %v", err)
	}
	rows := make([]domain.CatalogImportBrandRow, MaxCatalogListRows+1)
	if _, err := service.ImportBrandList(uuid.New(), rows); !errors.Is(err, ErrCatalogListTooLarge) {
		t.Fatalf("grande: err = %v", err)
	}
}

func TestParseImportsApplyNewRules(t *testing.T) {
	units := workbook(t, []string{"Nombre", "Símbolo", "Tipo", "Factor a base", "Decimales"}, [][]string{
		{"Kilogramo", "kg", "peso", "1000", "3"},
		{"Bulto", "blt", "BULTO", "1", "0"},
	})
	rows, result, err := parseUnitImport(bytes.NewReader(units))
	if err != nil || len(rows) != 1 || rows[0].Tipo != "PESO" || result.Invalidas != 1 || result.Errores[0].Campo != "Tipo" {
		t.Fatalf("unidades: rows=%#v result=%#v err=%v", rows, result, err)
	}

	providers := workbook(t, []string{"Nombre", "Razón social", "RFC", "Teléfono", "Correo", "Dirección"}, [][]string{
		{"Norte", "", "nor010101aa1", "", "", ""},
		{"Sur", "", "MAL", "", "", ""},
	})
	prows, presult, err := parseProviderImport(bytes.NewReader(providers))
	if err != nil || len(prows) != 1 || prows[0].RFC != "NOR010101AA1" || presult.Invalidas != 1 || presult.Errores[0].Campo != "RFC" {
		t.Fatalf("proveedores: rows=%#v result=%#v err=%v", prows, presult, err)
	}

	categories := workbook(t, []string{"Nombre", "Descripción", "Categoría padre"}, [][]string{
		{"Agua", "", " Bebidas >  Frías "},
		{"Hielo", "", "Bebidas > "},
	})
	crows, cresult, err := parseCategoryImport(bytes.NewReader(categories))
	if err != nil || len(crows) != 1 || crows[0].CategoriaPadre != "Bebidas > Frías" || cresult.Invalidas != 1 {
		t.Fatalf("categorías: rows=%#v result=%#v err=%v", crows, cresult, err)
	}
}

func TestIndividualUnitAndProviderRules(t *testing.T) {
	service := NewProductService(&individualRepositoryStub{}, nil)
	if err := service.CreateUnit(uuid.New(), domain.CreateUnidadMedidaInput{Nombre: "X", Simbolo: "x", Tipo: "BULTO", FactorABase: 1}); !errors.Is(err, domain.ErrTipoUnidadInvalido) {
		t.Fatalf("CreateUnit tipo inválido: %v", err)
	}
	tipo := "volumen"
	if err := service.UpdateUnit(uuid.New(), uuid.New(), domain.UpdateUnidadMedidaInput{Tipo: &tipo}); err != nil {
		t.Fatalf("UpdateUnit volumen: %v", err)
	}
	malo := "abc"
	if err := service.CreateProvider(uuid.New(), domain.CreateProveedorInput{Nombre: "P", RFC: &malo}); !errors.Is(err, domain.ErrRFCInvalido) {
		t.Fatalf("CreateProvider RFC inválido: %v", err)
	}
	bueno := "sig010101ab1"
	if err := service.UpdateProvider(uuid.New(), uuid.New(), domain.UpdateProveedorInput{RFC: &bueno}); err != nil || bueno != "SIG010101AB1" {
		t.Fatalf("UpdateProvider RFC: err=%v rfc=%q", err, bueno)
	}
}

type individualRepositoryStub struct{ ProductRepository }

func (individualRepositoryStub) CreateUnit(uuid.UUID, domain.CreateUnidadMedidaInput) error {
	return nil
}
func (individualRepositoryStub) UpdateUnit(_, _ uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
	if i.Tipo == nil || *i.Tipo != "VOLUMEN" {
		return errors.New("tipo sin normalizar")
	}
	return nil
}
func (individualRepositoryStub) CreateProvider(uuid.UUID, domain.CreateProveedorInput) error {
	return nil
}
func (individualRepositoryStub) UpdateProvider(uuid.UUID, uuid.UUID, domain.UpdateProveedorInput) error {
	return nil
}
