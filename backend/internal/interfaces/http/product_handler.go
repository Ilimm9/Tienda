package http

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"tienda/backend/internal/application"
	negocioapplication "tienda/backend/internal/application/negocio"
	"tienda/backend/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	products *application.ProductService
	contexto *negocioapplication.ContextoService
}

func (h *ProductHandler) LookupProduct(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("negocioId")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocioId no es válido"})
		return
	}

	result, err := h.products.LookupProduct(c.Param("codigoBarras"))
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}

	switch {
	case errors.Is(err, application.ErrInvalidBarcode):
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
	case errors.Is(err, application.ErrProductNotFound):
		c.JSON(http.StatusNotFound, gin.H{"mensaje": err.Error()})
	case errors.Is(err, application.ErrLookupRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"mensaje": err.Error()})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"mensaje": "No fue posible consultar los catálogos externos"})
	}
}

func NewProductHandler(products *application.ProductService, contexto *negocioapplication.ContextoService) *ProductHandler {
	return &ProductHandler{products: products, contexto: contexto}
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
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	items, err := h.products.ListCategories(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las categorías"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) Brands(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	items, err := h.products.ListBrands(businessID)
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
	if !h.sucursalActiva(c, businessID, input.SucursalID) {
		return
	}

	// validar que la sesión tenga membresía y permiso sobre este negocio.
	if err := h.products.Create(businessID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"mensaje": "Producto creado correctamente"})
}

func (h *ProductHandler) Update(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	productID, ok := parseID(c, "productoId")
	if !ok {
		return
	}
	var input domain.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los campos obligatorios del producto"})
		return
	}
	if err := h.products.Update(businessID, productID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductHandler) Deactivate(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	productID, ok := parseID(c, "productoId")
	if !ok {
		return
	}
	if err := h.products.Deactivate(businessID, productID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return uuid.Nil, false
	}
	return id, true
}

func (h *ProductHandler) sucursalActiva(c *gin.Context, negocioID, sucursalID uuid.UUID) bool {
	if err := h.contexto.ValidarSucursalActiva(c.Request.Context(), negocioID, sucursalID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return false
	}
	return true
}
func (h *ProductHandler) ListBrandsAdmin(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	v, e := h.products.ListBrandsAdmin(businessID)
	if e != nil {
		c.JSON(500, gin.H{"mensaje": "No fue posible cargar las marcas"})
		return
	}
	c.JSON(200, v)
}
func (h *ProductHandler) CreateBrand(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	var i domain.CreateMarcaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "El nombre de la marca es obligatorio"})
		return
	}
	if e := h.products.CreateBrand(businessID, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(201)
}
func (h *ProductHandler) UpdateBrand(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var i domain.UpdateMarcaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "Los datos no son válidos"})
		return
	}
	if e := h.products.UpdateBrand(businessID, id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(204)
}
func (h *ProductHandler) ListCategoriesAdmin(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	v, e := h.products.ListCategoriesAdmin(businessID)
	if e != nil {
		c.JSON(500, gin.H{"mensaje": "No fue posible cargar las categorías"})
		return
	}
	c.JSON(200, v)
}
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	var i domain.CreateCategoriaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "El nombre de la categoría es obligatorio"})
		return
	}
	if e := h.products.CreateCategory(businessID, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(201)
}
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var i domain.UpdateCategoriaInput
	if c.ShouldBindJSON(&i) != nil {
		c.JSON(400, gin.H{"mensaje": "Los datos no son válidos"})
		return
	}
	if e := h.products.UpdateCategory(businessID, id, i); e != nil {
		c.JSON(400, gin.H{"mensaje": e.Error()})
		return
	}
	c.Status(204)
}

const maxCatalogImportSize = 5 << 20

func (h *ProductHandler) BrandImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "marcas")
}

func (h *ProductHandler) CategoryImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "categorias")
}

func (h *ProductHandler) ProductImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "productos")
}

func (h *ProductHandler) catalogImportTemplate(c *gin.Context, section string) {
	content, err := application.CatalogImportTemplate(section)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible generar la plantilla"})
		return
	}
	filename := "plantilla-marcas.xlsx"
	if section == "unidades" {
		filename = "plantilla-unidades-medida.xlsx"
	}
	if section == "categorias" {
		filename = "plantilla-categorias.xlsx"
	}
	if section == "productos" {
		filename = "plantilla-productos.xlsx"
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func (h *ProductHandler) ImportBrands(c *gin.Context) {
	h.importCatalog(c, "marcas")
}

func (h *ProductHandler) ImportCategories(c *gin.Context) {
	h.importCatalog(c, "categorias")
}

func (h *ProductHandler) ImportProducts(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCatalogImportSize)
	branchID, err := uuid.Parse(c.PostForm("sucursal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona una sucursal válida"})
		return
	}
	if !h.sucursalActiva(c, businessID, branchID) {
		return
	}
	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona un archivo XLSX de hasta 5 MB"})
		return
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El archivo debe tener extensión .xlsx"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "No fue posible leer el archivo"})
		return
	}
	defer file.Close()
	result, err := h.products.ImportProducts(businessID, branchID, file)
	if err != nil {
		if errors.Is(err, application.ErrInvalidImportFile) {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ProductHandler) PreviewProductImport(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCatalogImportSize)
	branchID, err := uuid.Parse(c.PostForm("sucursal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona una sucursal válida"})
		return
	}
	if !h.sucursalActiva(c, businessID, branchID) {
		return
	}
	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona un archivo XLSX de hasta 5 MB"})
		return
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El archivo debe tener extensión .xlsx"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "No fue posible leer el archivo"})
		return
	}
	defer file.Close()
	preview, err := h.products.PreviewProductImport(businessID, branchID, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *ProductHandler) UnitImportTemplate(c *gin.Context) { h.catalogImportTemplate(c, "unidades") }
func (h *ProductHandler) ImportUnits(c *gin.Context)        { h.importCatalog(c, "unidades") }

func (h *ProductHandler) importCatalog(c *gin.Context, section string) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCatalogImportSize)
	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona un archivo XLSX de hasta 5 MB"})
		return
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El archivo debe tener extensión .xlsx"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "No fue posible leer el archivo"})
		return
	}
	defer file.Close()

	var result domain.CatalogImportResult
	if section == "marcas" {
		result, err = h.products.ImportBrands(businessID, file)
	} else if section == "categorias" {
		result, err = h.products.ImportCategories(businessID, file)
	} else {
		result, err = h.products.ImportUnits(businessID, file)
	}
	if err != nil {
		if errors.Is(err, application.ErrInvalidImportFile) {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible importar el archivo"})
		return
	}
	c.JSON(http.StatusOK, result)
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

func (h *ProductHandler) ListUnits(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	items, err := h.products.ListUnits(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las unidades de medida"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ProductHandler) CreateUnit(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	var input domain.CreateUnidadMedidaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los campos obligatorios de la unidad"})
		return
	}
	if err := h.products.CreateUnit(businessID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *ProductHandler) UpdateUnit(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var input domain.UpdateUnidadMedidaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Los datos de la unidad no son válidos"})
		return
	}
	if err := h.products.UpdateUnit(businessID, id, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
