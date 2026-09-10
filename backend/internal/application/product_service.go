package application

import (
	"errors"
	"io"
	"net/url"
	"regexp"
	"strings"
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrInvalidBarcode    = errors.New("el código de barras debe contener entre 8 y 14 dígitos")
	ErrProductNotFound   = errors.New("no se encontró el producto en los catálogos externos")
	ErrLookupRateLimited = errors.New("se alcanzó el límite de consultas de los catálogos externos")
	ErrLookupUnavailable = errors.New("los catálogos externos no están disponibles")
)

var barcodePattern = regexp.MustCompile(`^\d{8,14}$`)

type ProductLookupProvider interface {
	LookupProduct(barcode string) (domain.ProductLookup, error)
}

type ProductRepository interface {
	ListByBusiness(businessID uuid.UUID) ([]domain.ProductRow, error)
	ListCategories() ([]domain.CatalogOption, error)
	ListBrands() ([]domain.CatalogOption, error)
	ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error)
	Create(businessID uuid.UUID, input domain.CreateProductInput) error
	ListBrandsAdmin() ([]domain.Marca, error)
	CreateBrand(domain.CreateMarcaInput) error
	UpdateBrand(uuid.UUID, domain.UpdateMarcaInput) error
	ImportBrands([]domain.CatalogImportBrandRow) (domain.CatalogImportResult, error)
	ListCategoriesAdmin() ([]domain.Categoria, error)
	CreateCategory(domain.CreateCategoriaInput) error
	UpdateCategory(uuid.UUID, domain.UpdateCategoriaInput) error
	ImportCategories([]domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error)
	ImportUnits([]domain.CatalogImportUnitRow) (domain.CatalogImportResult, error)
	ImportProducts(uuid.UUID, uuid.UUID, []domain.ProductImportRow) (domain.CatalogImportResult, error)
	ListProviders(uuid.UUID) ([]domain.Proveedor, error)
	CreateProvider(uuid.UUID, domain.CreateProveedorInput) error
	UpdateProvider(uuid.UUID, uuid.UUID, domain.UpdateProveedorInput) error
	ListUnits() ([]domain.UnidadMedida, error)
	CreateUnit(domain.CreateUnidadMedidaInput) error
	UpdateUnit(uuid.UUID, domain.UpdateUnidadMedidaInput) error
}

func (s *ProductService) ImportProducts(businessID, branchID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	rows, result, err := parseProductImport(file)
	if err != nil {
		return result, err
	}
	for index := range rows {
		if rows[index].CodigoBarras == "" {
			continue
		}
		lookup, lookupErr := s.LookupProduct(rows[index].CodigoBarras)
		if lookupErr != nil {
			result.Advertencias = append(result.Advertencias, domain.CatalogImportIssue{Fila: rows[index].Fila, Motivo: "No fue posible buscar la imagen; se creará sin imagen"})
			continue
		}
		if lookup.ImagenURL == nil || !validImportImageURL(strings.TrimSpace(*lookup.ImagenURL)) {
			result.Advertencias = append(result.Advertencias, domain.CatalogImportIssue{Fila: rows[index].Fila, Motivo: "No se encontró una imagen para el código de barras"})
			continue
		}
		rows[index].ImagenURL = strings.TrimSpace(*lookup.ImagenURL)
	}
	imported, err := s.products.ImportProducts(businessID, branchID, rows)
	return mergeImportResults(result, imported), err
}

func validImportImageURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func (s *ProductService) ListBrandsAdmin() ([]domain.Marca, error) {
	return s.products.ListBrandsAdmin()
}
func (s *ProductService) CreateBrand(i domain.CreateMarcaInput) error {
	return s.products.CreateBrand(i)
}
func (s *ProductService) UpdateBrand(id uuid.UUID, i domain.UpdateMarcaInput) error {
	return s.products.UpdateBrand(id, i)
}
func (s *ProductService) ListCategoriesAdmin() ([]domain.Categoria, error) {
	return s.products.ListCategoriesAdmin()
}
func (s *ProductService) CreateCategory(i domain.CreateCategoriaInput) error {
	return s.products.CreateCategory(i)
}
func (s *ProductService) UpdateCategory(id uuid.UUID, i domain.UpdateCategoriaInput) error {
	return s.products.UpdateCategory(id, i)
}
func (s *ProductService) ListProviders(id uuid.UUID) ([]domain.Proveedor, error) {
	return s.products.ListProviders(id)
}
func (s *ProductService) CreateProvider(id uuid.UUID, i domain.CreateProveedorInput) error {
	return s.products.CreateProvider(id, i)
}
func (s *ProductService) UpdateProvider(businessID, id uuid.UUID, i domain.UpdateProveedorInput) error {
	return s.products.UpdateProvider(businessID, id, i)
}
func (s *ProductService) ListUnits() ([]domain.UnidadMedida, error) { return s.products.ListUnits() }
func (s *ProductService) CreateUnit(i domain.CreateUnidadMedidaInput) error {
	return s.products.CreateUnit(i)
}
func (s *ProductService) UpdateUnit(id uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
	return s.products.UpdateUnit(id, i)
}

func (s *ProductService) ListCategories() ([]domain.CatalogOption, error) {
	return s.products.ListCategories()
}

func (s *ProductService) ListBrands() ([]domain.CatalogOption, error) {
	return s.products.ListBrands()
}

func (s *ProductService) ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	return s.products.ListBranches(businessID)
}

func (s *ProductService) Create(businessID uuid.UUID, input domain.CreateProductInput) error {
	return s.products.Create(businessID, input)
}

func (s *ProductService) LookupProduct(barcode string) (domain.ProductLookup, error) {
	barcode = strings.TrimSpace(barcode)
	if !barcodePattern.MatchString(barcode) {
		return domain.ProductLookup{}, ErrInvalidBarcode
	}
	return s.productLookup.LookupProduct(barcode)
}

type ProductService struct {
	products      ProductRepository
	productLookup ProductLookupProvider
}

func NewProductService(products ProductRepository, productLookup ProductLookupProvider) *ProductService {
	return &ProductService{products: products, productLookup: productLookup}
}

func (s *ProductService) ListByBusiness(businessID uuid.UUID) ([]domain.ProductRow, error) {
	return s.products.ListByBusiness(businessID)
}
