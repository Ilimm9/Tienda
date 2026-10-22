package http

import (
	"bytes"
	"errors"
	"io"
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
	branchID, err := uuid.Parse(c.Query("sucursal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La sucursal activa es obligatoria"})
		return
	}
	if !h.sucursalActiva(c, businessID, branchID) {
		return
	}
	products, err := h.products.ListByBusiness(businessID, branchID)
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
	usuarioID, ok := AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	items, err := h.products.ListBranches(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las sucursales"})
		return
	}
	// Un miembro solo recibe las sucursales donde tiene asignación vigente.
	accesibles, err := h.contexto.SucursalesAccesibles(c.Request.Context(), usuarioID, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las sucursales"})
		return
	}
	permitidas := make(map[uuid.UUID]struct{}, len(accesibles))
	for _, id := range accesibles {
		permitidas[id] = struct{}{}
	}
	visibles := make([]domain.CatalogOption, 0, len(items))
	for _, item := range items {
		if _, ok := permitidas[item.ID]; ok {
			visibles = append(visibles, item)
		}
	}
	c.JSON(http.StatusOK, visibles)
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
	userID, ok := AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	input.UsuarioID = userID

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

func (h *ProductHandler) AdjustInventory(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	productID, ok := parseID(c, "productoId")
	if !ok {
		return
	}
	userID, ok := AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	var input domain.AdjustInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El stock y la sucursal son obligatorios y válidos"})
		return
	}
	if !h.sucursalActiva(c, businessID, input.SucursalID) {
		return
	}
	if err := h.products.AdjustInventory(businessID, productID, userID, input); err != nil {
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
	usuarioID, ok := AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return false
	}
	if err := h.contexto.ValidarSucursalActiva(c.Request.Context(), usuarioID, negocioID, sucursalID); err != nil {
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
func (h *ProductHandler) DeleteBrand(c *gin.Context) {
	h.deleteCatalog(c, "marca", h.products.DeleteBrand)
}
func (h *ProductHandler) CheckBrandDeletion(c *gin.Context) {
	h.checkCatalogDeletion(c, h.products.CheckBrandDeletion)
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
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	h.deleteCatalog(c, "categoría", h.products.DeleteCategory)
}
func (h *ProductHandler) CheckCategoryDeletion(c *gin.Context) {
	h.checkCatalogDeletion(c, h.products.CheckCategoryDeletion)
}

func (h *ProductHandler) deleteCatalog(c *gin.Context, label string, deleteFn func(uuid.UUID, uuid.UUID) error) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := deleteFn(businessID, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrCatalogItemInUse):
			c.JSON(http.StatusConflict, gin.H{"mensaje": err.Error()})
		case errors.Is(err, domain.ErrCatalogItemNotFound):
			c.JSON(http.StatusNotFound, gin.H{"mensaje": "No se encontró el " + label + " solicitado"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible eliminar el " + label})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProductHandler) checkCatalogDeletion(c *gin.Context, checkFn func(uuid.UUID, uuid.UUID) (domain.CatalogDeletionCheck, error)) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	result, err := checkFn(businessID, id)
	if err != nil {
		if errors.Is(err, domain.ErrCatalogItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"mensaje": "No se encontró el registro solicitado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible comprobar las vinculaciones"})
		return
	}
	c.JSON(http.StatusOK, result)
}

const maxCatalogImportSize = 5 << 20

func (h *ProductHandler) BrandImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "marcas")
}

func (h *ProductHandler) CategoryImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "categorias")
}

func (h *ProductHandler) ProviderImportTemplate(c *gin.Context) {
	h.catalogImportTemplate(c, "proveedores")
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
	if section == "proveedores" {
		filename = "plantilla-proveedores.xlsx"
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

func (h *ProductHandler) ImportProviders(c *gin.Context) {
	h.importCatalog(c, "proveedores")
}

func (h *ProductHandler) ImportProducts(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	branchID, content, _, ok := h.productImportUpload(c, businessID)
	if !ok {
		return
	}
	authorizeNewCatalogs := strings.EqualFold(strings.TrimSpace(c.PostForm("autorizar_catalogos_nuevos")), "true")
	job, err := h.products.StartProductImport(businessID, branchID, content, authorizeNewCatalogs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible iniciar la importación"})
		return
	}
	c.JSON(http.StatusAccepted, job)
}

func (h *ProductHandler) ProductImportStatus(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	jobID, ok := parseID(c, "importacionId")
	if !ok {
		return
	}
	job, err := h.products.GetProductImportJob(businessID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No se encontró la importación"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *ProductHandler) PreviewProductImport(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	branchID, content, _, ok := h.productImportUpload(c, businessID)
	if !ok {
		return
	}
	preview, err := h.products.PreviewProductImport(businessID, branchID, bytes.NewReader(content))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *ProductHandler) ProductImportErrors(c *gin.Context) {
	businessID, ok := parseID(c, "negocioId")
	if !ok {
		return
	}
	branchID, content, filename, ok := h.productImportUpload(c, businessID)
	if !ok {
		return
	}
	workbook, err := h.products.ProductImportErrors(businessID, branchID, content)
	if errors.Is(err, application.ErrNoImportErrors) {
		c.JSON(http.StatusConflict, gin.H{"mensaje": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="errores-`+strings.ReplaceAll(filename, `"`, "")+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", workbook)
}

// productImportUpload valida la sucursal y lee el XLSX de una petición multipart.
func (h *ProductHandler) productImportUpload(c *gin.Context, businessID uuid.UUID) (uuid.UUID, []byte, string, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCatalogImportSize)
	branchID, err := uuid.Parse(c.PostForm("sucursal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona una sucursal válida"})
		return uuid.Nil, nil, "", false
	}
	if !h.sucursalActiva(c, businessID, branchID) {
		return uuid.Nil, nil, "", false
	}
	fileHeader, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Selecciona un archivo XLSX de hasta 5 MB"})
		return uuid.Nil, nil, "", false
	}
	if strings.ToLower(filepath.Ext(fileHeader.Filename)) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El archivo debe tener extensión .xlsx"})
		return uuid.Nil, nil, "", false
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "No fue posible leer el archivo"})
		return uuid.Nil, nil, "", false
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "No fue posible leer el archivo"})
		return uuid.Nil, nil, "", false
	}
	return branchID, content, filepath.Base(fileHeader.Filename), true
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
	} else if section == "proveedores" {
		result, err = h.products.ImportProviders(businessID, file)
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
func (h *ProductHandler) DeleteProvider(c *gin.Context) {
	h.deleteCatalog(c, "proveedor", h.products.DeleteProvider)
}
func (h *ProductHandler) CheckProviderDeletion(c *gin.Context) {
	h.checkCatalogDeletion(c, h.products.CheckProviderDeletion)
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

func (h *ProductHandler) DeleteUnit(c *gin.Context) {
	h.deleteCatalog(c, "unidad de medida", h.products.DeleteUnit)
}
func (h *ProductHandler) CheckUnitDeletion(c *gin.Context) {
	h.checkCatalogDeletion(c, h.products.CheckUnitDeletion)
}
