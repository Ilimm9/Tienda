package precio

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
	precioapp "tienda/backend/internal/application/precio"
	transport "tienda/backend/internal/interfaces/http"
	"time"
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
	filtro, ok := filtroPendientes(c)
	if !ok {
		return
	}
	items, err := h.service.ListPending(businessID, branchID, filtro)
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

func (h *Handler) ListHistory(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	branchID, branchErr := uuid.Parse(c.Query("sucursal_id"))
	productID, productErr := uuid.Parse(c.Query("producto_negocio_id"))
	if err != nil || branchErr != nil || productErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El negocio, sucursal y producto son obligatorios"})
		return
	}
	if !h.sucursalPermitida(c, businessID, branchID) {
		return
	}
	items, err := h.service.ListHistory(businessID, branchID, productID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar el historial"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) AdjustPrice(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	productID, productErr := uuid.Parse(c.Param("productoNegocioId"))
	userID, ok := transport.AuthenticatedUserID(c)
	if err != nil || productErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	var input precioapp.AjustarPrecioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa el precio y motivo"})
		return
	}
	if !h.sucursalPermitida(c, businessID, input.SucursalID) {
		return
	}
	item, err := h.service.AdjustPrice(businessID, userID, productID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// filtroPendientes lee proveedor_id, desde, hasta (fecha de recepción) y q.
func filtroPendientes(c *gin.Context) (precioapp.FiltroPendientes, bool) {
	filtro := precioapp.FiltroPendientes{Texto: strings.TrimSpace(c.Query("q"))}
	if value := c.Query("proveedor_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El proveedor no es válido"})
			return filtro, false
		}
		filtro.ProveedorID = &id
	}
	for key, target := range map[string]**time.Time{"desde": &filtro.Desde, "hasta": &filtro.Hasta} {
		if value := c.Query(key); value != "" {
			date, err := time.Parse("2006-01-02", value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La fecha no es válida"})
				return filtro, false
			}
			*target = &date
		}
	}
	return filtro, true
}

// AuthorizePurchase valida el ticket entero en una transacción, con el precio de cada renglón.
func (h *Handler) AuthorizePurchase(c *gin.Context) {
	businessID, err := uuid.Parse(c.Param("negocioId"))
	purchaseID, purchaseErr := uuid.Parse(c.Param("compraId"))
	userID, ok := transport.AuthenticatedUserID(c)
	if err != nil || purchaseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	branchID, err := h.service.BranchOfPurchase(businessID, purchaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "Compra no encontrada"})
		return
	}
	if !h.sucursalPermitida(c, businessID, branchID) {
		return
	}
	var input precioapp.AutorizarCompraInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los importes"})
		return
	}
	items, err := h.service.AuthorizePurchase(businessID, userID, purchaseID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}
