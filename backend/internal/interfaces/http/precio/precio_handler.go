package precio

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	precioapp "tienda/backend/internal/application/precio"
	transport "tienda/backend/internal/interfaces/http"
)

// AccesoSucursal es el contrato hacia `negocio`: si el usuario puede operar la sucursal.
type AccesoSucursal interface {
	ValidarSucursalActiva(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) error
}

type Handler struct {
	service *precioapp.Service
	acceso  AccesoSucursal
}

func NewHandler(service *precioapp.Service, acceso AccesoSucursal) *Handler {
	return &Handler{service: service, acceso: acceso}
}

// sucursalPermitida responde 404 sin distinguir sucursal inexistente, ajena o no asignada.
func (h *Handler) sucursalPermitida(c *gin.Context, businessID, branchID uuid.UUID) bool {
	userID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return false
	}
	if h.acceso.ValidarSucursalActiva(c.Request.Context(), userID, businessID, branchID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return false
	}
	return true
}
func (h *Handler) ListPending(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	branchID, branchErr := uuid.Parse(c.Query("sucursal_id"))
	if err != nil || branchErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocio y la sucursal son obligatorios"})
		return
	}
	if !h.sucursalPermitida(c, businessID, branchID) {
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
	branchID, err := h.service.BranchOfProposal(businessID, proposalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "Propuesta no encontrada"})
		return
	}
	if !h.sucursalPermitida(c, businessID, branchID) {
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
