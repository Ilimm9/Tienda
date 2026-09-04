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

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return uuid.Nil, false
	}
	return id, true
}
func (h *ProductHandler) ListBrandsAdmin(c *gin.Context) {
	v, e := h.products.ListBrandsAdmin()
	if e != nil {
		c.JSON(500, gin.H{"mensaje": "No fue posible cargar las marcas"})
		return
	}
	c.JSON(200, v)
}
func (h *ProductHandler) CreateBrand(c *gin.Context) {
	var i domain.CreateMarcaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "El nombre de la marca es obligatorio"})
		return
	}
	if e := h.products.CreateBrand(i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(201)
}
func (h *ProductHandler) UpdateBrand(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var i domain.UpdateMarcaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "Los datos no son válidos"})
		return
	}
	if e := h.products.UpdateBrand(id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(204)
}
func (h *ProductHandler) ListCategoriesAdmin(c *gin.Context) {
	v, e := h.products.ListCategoriesAdmin()
	if e != nil {
		c.JSON(500, gin.H{"mensaje": "No fue posible cargar las categorías"})
		return
	}
	c.JSON(200, v)
}
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var i domain.CreateCategoriaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "El nombre de la categoría es obligatorio"})
		return
	}
	if e := h.products.CreateCategory(i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(201)
}
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var i domain.UpdateCategoriaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "Los datos no son válidos"})
		return
	}
	if e := h.products.UpdateCategory(id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(204)
}
func (h *ProductHandler) ListProviders(c *gin.Context) {
	id, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	v, e := h.products.ListProviders(id)
	if e != nil {
		c.JSON(500, gin.H{"mensaje": "No fue posible cargar los proveedores"})
		return
	}
	c.JSON(200, v)
}
func (h *ProductHandler) CreateProvider(c *gin.Context) {
	id, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	var i domain.CreateProveedorInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "El nombre del proveedor es obligatorio"})
		return
	}
	if e := h.products.CreateProvider(id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(201)
}
func (h *ProductHandler) UpdateProvider(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var i domain.UpdateProveedorInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "Los datos no son válidos"})
		return
	}
	if e := h.products.UpdateProvider(businessID, id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(204)
}
