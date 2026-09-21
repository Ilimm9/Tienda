package application

import (
	"bytes"
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
	Update(businessID, productID uuid.UUID, input domain.UpdateProductInput) error
	Deactivate(businessID, productID uuid.UUID) error
	ListBrandsAdmin() ([]domain.Marca, error)
	CreateBrand(domain.CreateMarcaInput) error
	UpdateBrand(uuid.UUID, domain.UpdateMarcaInput) error
	ImportBrands([]domain.CatalogImportBrandRow) (domain.CatalogImportResult, error)
	ListCategoriesAdmin() ([]domain.Categoria, error)
	CreateCategory(domain.CreateCategoriaInput) error
	UpdateCategory(uuid.UUID, domain.UpdateCategoriaInput) error
	ImportCategories([]domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error)
	ImportUnits([]domain.CatalogImportUnitRow) (domain.CatalogImportResult, error)
	ValidateProductImport(uuid.UUID, uuid.UUID, []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, error)
	CreateImportedProducts(uuid.UUID, []domain.ValidatedProductImportRow) (domain.CatalogImportResult, error)
	CreateProductImportJob(uuid.UUID, uuid.UUID) (domain.ProductImportJob, error)
	GetProductImportJob(uuid.UUID, uuid.UUID) (domain.ProductImportJob, error)
	UpdateProductImportJob(uuid.UUID, string, string, int, *domain.CatalogImportResult, *string) error
	ListProviders(uuid.UUID) ([]domain.Proveedor, error)
	CreateProvider(uuid.UUID, domain.CreateProveedorInput) error
	UpdateProvider(uuid.UUID, uuid.UUID, domain.UpdateProveedorInput) error
	ListUnits() ([]domain.UnidadMedida, error)
	CreateUnit(domain.CreateUnidadMedidaInput) error
	UpdateUnit(uuid.UUID, domain.UpdateUnidadMedidaInput) error
}

func (s *ProductService) ImportProducts(businessID, branchID uuid.UUID, file io.Reader) (domain.CatalogImportResult, error) {
	return s.importProducts(businessID, branchID, file, nil)
}

func (s *ProductService) importProducts(businessID, branchID uuid.UUID, file io.Reader, progress func(string, int)) (domain.CatalogImportResult, error) {
	prepared, result, err := s.prepareProductImport(businessID, branchID, file)
	if err != nil {
		return result, err
	}
	if progress != nil {
		progress("Archivo validado", 20)
	}
	if progress != nil {
		progress("Creando productos", 30)
	}
	imported, err := s.products.CreateImportedProducts(businessID, prepared)
	if err != nil {
		return result, err
	}
	result.Creadas += imported.Creadas
	result.Omitidas += imported.Omitidas
	result.Invalidas += imported.Invalidas
	result.SKUsGenerados += imported.SKUsGenerados
	result.Errores = append(result.Errores, imported.Errores...)
	return result, nil
}

func (s *ProductService) StartProductImport(businessID, branchID uuid.UUID, content []byte) (domain.ProductImportJob, error) {
	job, err := s.products.CreateProductImportJob(businessID, branchID)
	if err != nil {
		return domain.ProductImportJob{}, err
	}
	go func() {
		_ = s.products.UpdateProductImportJob(job.ID, "procesando", "Validando archivo", 10, nil, nil)
		result, importErr := s.importProducts(businessID, branchID, bytes.NewReader(content), func(stage string, progress int) {
			_ = s.products.UpdateProductImportJob(job.ID, "procesando", stage, progress, nil, nil)
		})
		if importErr != nil {
			message := importErr.Error()
			_ = s.products.UpdateProductImportJob(job.ID, "fallida", "No fue posible importar", 100, nil, &message)
			return
		}
		_ = s.products.UpdateProductImportJob(job.ID, "completada", "Importación terminada", 100, &result, nil)
	}()
	return job, nil
}

func (s *ProductService) GetProductImportJob(businessID, jobID uuid.UUID) (domain.ProductImportJob, error) {
	return s.products.GetProductImportJob(businessID, jobID)
}

func (s *ProductService) PreviewProductImport(businessID, branchID uuid.UUID, file io.Reader) (domain.ProductImportPreview, error) {
	prepared, result, err := s.prepareProductImport(businessID, branchID, file)
	for _, row := range prepared {
		if row.Input.SKUInterno == nil || strings.TrimSpace(*row.Input.SKUInterno) == "" {
			result.SKUsGenerados++
		}
	}
	return domain.ProductImportPreview{CatalogImportResult: result, Insertables: len(prepared)}, err
}

func (s *ProductService) prepareProductImport(businessID, branchID uuid.UUID, file io.Reader) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, error) {
	rows, result, err := parseProductImport(file)
	if err != nil {
		return nil, result, err
	}
	prepared, validation, err := s.products.ValidateProductImport(businessID, branchID, rows)
	if err != nil {
		return nil, result, err
	}
	result.Omitidas += validation.Omitidas
	result.Invalidas += validation.Invalidas
	result.Errores = append(result.Errores, validation.Errores...)
	return prepared, result, nil
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

func (s *ProductService) Update(businessID, productID uuid.UUID, input domain.UpdateProductInput) error {
	return s.products.Update(businessID, productID, input)
}

func (s *ProductService) Deactivate(businessID, productID uuid.UUID) error {
	return s.products.Deactivate(businessID, productID)
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
