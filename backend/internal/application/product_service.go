package application

import (
	"tienda/backend/internal/domain"

	"github.com/google/uuid"
)

type ProductRepository interface {
	ListByBusiness(businessID uuid.UUID) ([]domain.ProductRow, error)
	ListCategories() ([]domain.CatalogOption, error)
	ListBrands() ([]domain.CatalogOption, error)
	ListBranches(businessID uuid.UUID) ([]domain.CatalogOption, error)
	Create(businessID uuid.UUID, input domain.CreateProductInput) error
	ListBrandsAdmin() ([]domain.Marca, error)
	CreateBrand(domain.CreateMarcaInput) error
	UpdateBrand(uuid.UUID, domain.UpdateMarcaInput) error
	ListCategoriesAdmin() ([]domain.Categoria, error)
	CreateCategory(domain.CreateCategoriaInput) error
	UpdateCategory(uuid.UUID, domain.UpdateCategoriaInput) error
	ListProviders(uuid.UUID) ([]domain.Proveedor, error)
	CreateProvider(uuid.UUID, domain.CreateProveedorInput) error
	UpdateProvider(uuid.UUID, uuid.UUID, domain.UpdateProveedorInput) error
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

type ProductService struct {
	products ProductRepository
}

func NewProductService(products ProductRepository) *ProductService {
	return &ProductService{products: products}
}

func (s *ProductService) ListByBusiness(businessID uuid.UUID) ([]domain.ProductRow, error) {
	return s.products.ListByBusiness(businessID)
}
