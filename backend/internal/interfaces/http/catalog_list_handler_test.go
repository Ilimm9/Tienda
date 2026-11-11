package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type catalogListRepositoryStub struct {
	application.ProductRepository
	brands []domain.CatalogImportBrandRow
}

func (s *catalogListRepositoryStub) ImportBrands(_ uuid.UUID, rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error) {
	s.brands = rows
	return domain.CatalogImportResult{Creadas: len(rows), Errores: []domain.CatalogImportIssue{}}, nil
}
func (s *catalogListRepositoryStub) ImportCategoryPaths(uuid.UUID, []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error) {
	return domain.CatalogImportResult{}, fmt.Errorf("base caída")
}
func (s *catalogListRepositoryStub) CreateCategory(uuid.UUID, domain.CreateCategoriaInput) error {
	return domain.ErrCategoriaDuplicada
}

func catalogListRouter(repo *catalogListRepositoryStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := &ProductHandler{products: application.NewProductService(repo, nil)}
	router := gin.New()
	router.POST("/negocios/:negocioId/catalogo/marcas/lista", handler.ImportBrandList)
	router.POST("/negocios/:negocioId/catalogo/categorias/lista", handler.ImportCategoryList)
	router.POST("/negocios/:negocioId/catalogo/categorias", handler.CreateCategory)
	return router
}

func postJSON(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(res, req)
	return res
}

func TestImportCatalogListResponses(t *testing.T) {
	repo := &catalogListRepositoryStub{}
	router := catalogListRouter(repo)
	base := "/negocios/" + uuid.NewString() + "/catalogo/"

	res := postJSON(router, base+"marcas/lista", `{"renglones":[{"fila":1,"nombre":"Bimbo"},{"fila":2,"nombre":""}]}`)
	if res.Code != http.StatusOK {
		t.Fatalf("200 esperado, llegó %d: %s", res.Code, res.Body)
	}
	var result domain.CatalogImportResult
	if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(repo.brands) != 1 || result.Creadas != 1 || result.Invalidas != 1 || result.Procesadas != 2 || result.Errores[0].Fila != 2 {
		t.Fatalf("result = %#v, brands = %#v", result, repo.brands)
	}

	casos := map[string]string{
		"vacía":      `{"renglones":[]}`,
		"sin campo":  `{}`,
		"json malo":  `{"renglones":`,
		"tipo malo":  `{"renglones":"Bimbo"}`,
		"muy grande": `{"renglones":[` + strings.TrimSuffix(strings.Repeat(`{"fila":1,"nombre":"x"},`, application.MaxCatalogListRows+1), ",") + `]}`,
	}
	for nombre, body := range casos {
		if res := postJSON(router, base+"marcas/lista", body); res.Code != http.StatusBadRequest {
			t.Errorf("%s: 400 esperado, llegó %d", nombre, res.Code)
		}
	}

	big := `{"renglones":[{"fila":1,"nombre":"` + strings.Repeat("x", 1<<20) + `"}]}`
	if res := postJSON(router, base+"marcas/lista", big); res.Code != http.StatusBadRequest {
		t.Errorf("cuerpo > 1 MB: 400 esperado, llegó %d", res.Code)
	}

	if res := postJSON(router, base+"categorias/lista", `{"renglones":[{"fila":1,"ruta":["A"]}]}`); res.Code != http.StatusInternalServerError || !strings.Contains(res.Body.String(), "ningún registro") {
		t.Fatalf("falla de base: %d %s", res.Code, res.Body)
	}

	if res := postJSON(router, base+"categorias", `{"nombre":"Frías"}`); res.Code != http.StatusConflict {
		t.Fatalf("categoría duplicada: 409 esperado, llegó %d", res.Code)
	}
}
