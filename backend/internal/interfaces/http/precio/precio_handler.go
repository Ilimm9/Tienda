package precio

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	precioapp "tienda/backend/internal/application/precio"
	transport "tienda/backend/internal/interfaces/http"
)

type Handler struct{ service *precioapp.Service }

func NewHandler(service *precioapp.Service) *Handler { return &Handler{service: service} }
func (h *Handler) ListPending(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	branchID, branchErr := uuid.Parse(c.Query("sucursal_id"))
	if err != nil || branchErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocio y la sucursal son obligatorios"})
		return
	}
	items, err := h.service.ListPending(businessID, branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las propuestas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}
func (h *Handler) Authorize(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	proposalID, proposalErr := uuid.Parse(c.Param("propuestaId"))
	userID, ok := transport.AuthenticatedUserID(c)
	if err != nil || proposalErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	var input precioapp.AutorizarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los importes"})
		return
	}
	item, err := h.service.Authorize(businessID, userID, proposalID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
