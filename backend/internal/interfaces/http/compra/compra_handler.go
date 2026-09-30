package compra

import (
	"net/http"

	compraapp "tienda/backend/internal/application/compra"
	transport "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ service *compraapp.Service }

func NewHandler(service *compraapp.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	items, err := h.service.List(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las compras"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}
func (h *Handler) Get(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	purchaseID, ok := id(c, "compraId")
	if !ok {
		return
	}
	item, err := h.service.Get(businessID, purchaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "Compra no encontrada"})
		return
	}
	c.JSON(http.StatusOK, item)
}
func (h *Handler) Products(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	providerID, err := uuid.Parse(c.Query("proveedor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El proveedor no es válido"})
		return
	}
	items, err := h.service.ListProductos(businessID, providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar los productos"})
		return
	}
	c.JSON(http.StatusOK, items)
}
func (h *Handler) Create(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	userID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	var input compraapp.CrearCompraInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los campos de la compra"})
		return
	}
	item, err := h.service.Create(businessID, userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}
func id(c *gin.Context, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(c.Param(key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return uuid.Nil, false
	}
	return parsed, true
}
