package http

import (
	"net/http"
	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	products *application.ProductService
}

func NewProductHandler(products *application.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

func (h *ProductHandler) List(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocioId no es válido"})
		return
	}

	// validar que la sesión tenga membresía y permiso sobre este negocio.
	products, err := h.products.ListByBusiness(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar los productos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": products, "total": len(products)})
}

func (h *ProductHandler) Categories(c *gin.Context) {
	items, err := h.products.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las categorías"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) Brands(c *gin.Context) {
	items, err := h.products.ListBrands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las marcas"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) Branches(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocioId no es válido"})
		return
	}
	items, err := h.products.ListBranches(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las sucursales"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) Create(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocioId no es válido"})
		return
	}
	var input domain.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los campos obligatorios del producto"})
		return
	}

	// validar que la sesión tenga membresía y permiso sobre este negocio.
	if err := h.products.Create(businessID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"mensaje": "Producto creado correctamente"})
}
