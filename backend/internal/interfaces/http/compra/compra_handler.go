package compra

import (
	"context"
	"net/http"

	compraapp "tienda/backend/internal/application/compra"
	transport "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AccesoSucursal es el contrato hacia `negocio`: qué sucursales puede operar el usuario.
type AccesoSucursal interface {
	ValidarSucursalActiva(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) error
	SucursalesAccesibles(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]uuid.UUID, error)
}

type Handler struct {
	service *compraapp.Service
	acceso  AccesoSucursal
}

func NewHandler(service *compraapp.Service, acceso AccesoSucursal) *Handler {
	return &Handler{service: service, acceso: acceso}
}

// sucursalesPermitidas devuelve el conjunto de sucursales del usuario; responde el error y `false` si no puede resolverlo.
func (h *Handler) sucursalesPermitidas(c *gin.Context, businessID uuid.UUID) (map[uuid.UUID]struct{}, bool) {
	userID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return nil, false
	}
	ids, err := h.acceso.SucursalesAccesibles(c.Request.Context(), userID, businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible comprobar el acceso a la sucursal"})
		return nil, false
	}
	permitidas := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		permitidas[id] = struct{}{}
	}
	return permitidas, true
}

func (h *Handler) List(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	permitidas, ok := h.sucursalesPermitidas(c, businessID)
	if !ok {
		return
	}
	todas, err := h.service.List(businessID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las compras"})
		return
	}
	items := todas[:0:0]
	for _, item := range todas {
		if _, ok := permitidas[item.SucursalID]; ok {
			items = append(items, item)
		}
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
	permitidas, ok := h.sucursalesPermitidas(c, businessID)
	if !ok {
		return
	}
	item, err := h.service.Get(businessID, purchaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "Compra no encontrada"})
		return
	}
	// Una compra de otra sucursal responde igual que una inexistente.
	if _, ok := permitidas[item.SucursalID]; !ok {
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
	branchID, err := uuid.Parse(input.SucursalID)
	if err != nil || h.acceso.ValidarSucursalActiva(c.Request.Context(), userID, businessID, branchID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
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
