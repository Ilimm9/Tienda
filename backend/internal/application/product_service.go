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
	ErrInvalidBarcode           = errors.New("el código de barras debe contener entre 8 y 14 dígitos")
	ErrProductNotFound          = errors.New("no se encontró el producto en los catálogos externos")
	ErrLookupRateLimited        = errors.New("se alcanzó el límite de consultas de los catálogos externos")
	ErrLookupUnavailable        = errors.New("los catálogos externos no están disponibles")
	ErrNewImportCatalogsPending = errors.New("autoriza los catálogos nuevos detectados antes de importar")
)

var barcodePattern = regexp.MustCompile(`^\d{8,14}$`)

type ProductLookupProvider interface {
	LookupProduct(barcode string) (domain.ProductLookup, error)
}

type ProductRepository interface {
	ListByBusiness(businessID, branchID uuid.UUID) ([]domain.ProductRow, error)
	ListCategories(businessID uuid.UUID) ([]domain.CatalogOption, error)
	ListBrands(businessID uuid.UUID) ([]domain.CatalogOption, error)
	ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error)
	Create(businessID uuid.UUID, input domain.CreateProductInput) error
	Update(businessID, productID uuid.UUID, input domain.UpdateProductInput) error
	AdjustInventory(businessID, productID, userID uuid.UUID, input domain.AdjustInventoryInput) error
	Deactivate(businessID, productID uuid.UUID) error
	ListBrandsAdmin(businessID uuid.UUID) ([]domain.Marca, error)
	CreateBrand(businessID uuid.UUID, input domain.CreateMarcaInput) error
	UpdateBrand(businessID, brandID uuid.UUID, input domain.UpdateMarcaInput) error
	CheckBrandDeletion(businessID, brandID uuid.UUID) (domain.CatalogDeletionCheck, error)
	DeleteBrand(businessID, brandID uuid.UUID) error
	ImportBrands(businessID uuid.UUID, rows []domain.CatalogImportBrandRow) (domain.CatalogImportResult, error)
	ListCategoriesAdmin(businessID uuid.UUID) ([]domain.Categoria, error)
	CreateCategory(businessID uuid.UUID, input domain.CreateCategoriaInput) error
	UpdateCategory(businessID, categoryID uuid.UUID, input domain.UpdateCategoriaInput) error
	CheckCategoryDeletion(businessID, categoryID uuid.UUID) (domain.CatalogDeletionCheck, error)
	DeleteCategory(businessID, categoryID uuid.UUID) error
	ImportCategories(businessID uuid.UUID, rows []domain.CatalogImportCategoryRow) (domain.CatalogImportResult, error)
	ImportProviders(businessID uuid.UUID, rows []domain.CatalogImportProviderRow) (domain.CatalogImportResult, error)
	ImportUnits(businessID uuid.UUID, rows []domain.CatalogImportUnitRow) (domain.CatalogImportResult, error)
	ValidateProductImport(uuid.UUID, uuid.UUID, []domain.ProductImportRow) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, []domain.ProductImportCatalogProposal, error)
	CreateImportedCatalogs(uuid.UUID, []domain.ProductImportCatalogProposal) error
	CreateImportedProducts(uuid.UUID, []domain.ValidatedProductImportRow) (domain.CatalogImportResult, error)
	CreateProductImportJob(uuid.UUID, uuid.UUID) (domain.ProductImportJob, error)
	GetProductImportJob(uuid.UUID, uuid.UUID) (domain.ProductImportJob, error)
	UpdateProductImportJob(uuid.UUID, string, string, int, *domain.CatalogImportResult, *string) error
	ListProviders(uuid.UUID) ([]domain.Proveedor, error)
	CreateProvider(uuid.UUID, domain.CreateProveedorInput) error
	UpdateProvider(uuid.UUID, uuid.UUID, domain.UpdateProveedorInput) error
	CheckProviderDeletion(businessID, providerID uuid.UUID) (domain.CatalogDeletionCheck, error)
	DeleteProvider(businessID, providerID uuid.UUID) error
	ListUnits(businessID uuid.UUID) ([]domain.UnidadMedida, error)
	CreateUnit(businessID uuid.UUID, input domain.CreateUnidadMedidaInput) error
	UpdateUnit(businessID, unitID uuid.UUID, input domain.UpdateUnidadMedidaInput) error
	CheckUnitDeletion(businessID, unitID uuid.UUID) (domain.CatalogDeletionCheck, error)
	DeleteUnit(businessID, unitID uuid.UUID) error
}

func (s *ProductService) ImportProducts(businessID, branchID uuid.UUID, file io.Reader, authorizeNewCatalogs bool) (domain.CatalogImportResult, error) {
	return s.importProducts(businessID, branchID, file, authorizeNewCatalogs, nil)
}

func (s *ProductService) importProducts(businessID, branchID uuid.UUID, file io.Reader, authorizeNewCatalogs bool, progress func(string, int)) (domain.CatalogImportResult, error) {
	prepared, result, proposals, err := s.prepareProductImport(businessID, branchID, file, authorizeNewCatalogs)
	if err != nil {
		return result, err
	}
	if len(proposals) > 0 && !authorizeNewCatalogs {
		return result, ErrNewImportCatalogsPending
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
	return mergeImportResults(result, imported), nil
}

func (s *ProductService) StartProductImport(businessID, branchID uuid.UUID, content []byte, authorizeNewCatalogs bool) (domain.ProductImportJob, error) {
	job, err := s.products.CreateProductImportJob(businessID, branchID)
	if err != nil {
		return domain.ProductImportJob{}, err
	}
	go func() {
		_ = s.products.UpdateProductImportJob(job.ID, "procesando", "Validando archivo", 10, nil, nil)
		result, importErr := s.importProducts(businessID, branchID, bytes.NewReader(content), authorizeNewCatalogs, func(stage string, progress int) {
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
	prepared, result, proposals, err := s.prepareProductImport(businessID, branchID, file, false)
	result.SKUsGenerados += len(prepared)
	return domain.ProductImportPreview{CatalogImportResult: result, Insertables: len(prepared), CatalogosNuevos: proposals}, err
}

func (s *ProductService) prepareProductImport(businessID, branchID uuid.UUID, file io.Reader, authorizeNewCatalogs bool) ([]domain.ValidatedProductImportRow, domain.CatalogImportResult, []domain.ProductImportCatalogProposal, error) {
	rows, result, err := parseProductImport(file)
	if err != nil {
		return nil, result, nil, err
	}
	prepared, validation, proposals, err := s.products.ValidateProductImport(businessID, branchID, rows)
	if err != nil {
		return nil, result, nil, err
	}
	if len(proposals) > 0 && authorizeNewCatalogs {
		if err := s.products.CreateImportedCatalogs(businessID, proposals); err != nil {
			return nil, result, proposals, err
		}
		prepared, validation, _, err = s.products.ValidateProductImport(businessID, branchID, rows)
		if err != nil {
			return nil, result, proposals, err
		}
	}
	result.Omitidas += validation.Omitidas
	result.Invalidas += validation.Invalidas
	result.Errores = append(result.Errores, validation.Errores...)
	if len(proposals) == 0 {
		return prepared, result, nil, nil
	}
	return prepared, result, proposals, nil
}

func validImportImageURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func (s *ProductService) ListBrandsAdmin(businessID uuid.UUID) ([]domain.Marca, error) {
	return s.products.ListBrandsAdmin(businessID)
}
func (s *ProductService) CreateBrand(businessID uuid.UUID, i domain.CreateMarcaInput) error {
	return s.products.CreateBrand(businessID, i)
}
func (s *ProductService) UpdateBrand(businessID, id uuid.UUID, i domain.UpdateMarcaInput) error {
	return s.products.UpdateBrand(businessID, id, i)
}
func (s *ProductService) DeleteBrand(businessID, id uuid.UUID) error {
	return s.products.DeleteBrand(businessID, id)
}
func (s *ProductService) CheckBrandDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return s.products.CheckBrandDeletion(businessID, id)
}
func (s *ProductService) ListCategoriesAdmin(businessID uuid.UUID) ([]domain.Categoria, error) {
	return s.products.ListCategoriesAdmin(businessID)
}
func (s *ProductService) CreateCategory(businessID uuid.UUID, i domain.CreateCategoriaInput) error {
	return s.products.CreateCategory(businessID, i)
}
func (s *ProductService) UpdateCategory(businessID, id uuid.UUID, i domain.UpdateCategoriaInput) error {
	return s.products.UpdateCategory(businessID, id, i)
}
func (s *ProductService) DeleteCategory(businessID, id uuid.UUID) error {
	return s.products.DeleteCategory(businessID, id)
}
func (s *ProductService) CheckCategoryDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return s.products.CheckCategoryDeletion(businessID, id)
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
func (s *ProductService) DeleteProvider(businessID, id uuid.UUID) error {
	return s.products.DeleteProvider(businessID, id)
}
func (s *ProductService) CheckProviderDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return s.products.CheckProviderDeletion(businessID, id)
}
func (s *ProductService) ListUnits(businessID uuid.UUID) ([]domain.UnidadMedida, error) {
	return s.products.ListUnits(businessID)
}
func (s *ProductService) CreateUnit(businessID uuid.UUID, i domain.CreateUnidadMedidaInput) error {
	return s.products.CreateUnit(businessID, i)
}
func (s *ProductService) UpdateUnit(businessID, id uuid.UUID, i domain.UpdateUnidadMedidaInput) error {
	return s.products.UpdateUnit(businessID, id, i)
}
func (s *ProductService) DeleteUnit(businessID, id uuid.UUID) error {
	return s.products.DeleteUnit(businessID, id)
}
func (s *ProductService) CheckUnitDeletion(businessID, id uuid.UUID) (domain.CatalogDeletionCheck, error) {
	return s.products.CheckUnitDeletion(businessID, id)
}

func (s *ProductService) ListCategories(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	return s.products.ListCategories(businessID)
}

func (s *ProductService) ListBrands(businessID uuid.UUID) ([]domain.CatalogOption, error) {
	return s.products.ListBrands(businessID)
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

func (s *ProductService) ListByBusiness(businessID, branchID uuid.UUID) ([]domain.ProductRow, error) {
	return s.products.ListByBusiness(businessID, branchID)
}

func (s *ProductService) AdjustInventory(businessID, productID, userID uuid.UUID, input domain.AdjustInventoryInput) error {
	return s.products.AdjustInventory(businessID, productID, userID, input)
}
