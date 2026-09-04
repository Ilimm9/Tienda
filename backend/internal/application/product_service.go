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
