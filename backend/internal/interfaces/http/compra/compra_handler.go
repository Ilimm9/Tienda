package compra

import (
	"context"
	"net/http"
	"time"

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
	branchID, err := uuid.Parse(c.Query("sucursal_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La sucursal no es válida"})
		return
	}
	permitidas, ok := h.sucursalesPermitidas(c, businessID)
	if !ok {
		return
	}
	// El listado es de la sucursal activa, no de todas las que el usuario puede operar.
	if _, ok := permitidas[branchID]; !ok {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return
	}
	items, err := h.service.List(businessID, branchID)
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
	permitidas, ok := h.sucursalesPermitidas(c, businessID)
	if !ok {
		return
	}
	item, err := h.service.View(businessID, purchaseID)
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

// Export responde el .xlsx de compras (hoja 1 y, si se pide, COMPROBACIÓN).
func (h *Handler) Export(c *gin.Context) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return
	}
	userID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return
	}
	branchID, err := uuid.Parse(c.Query("sucursal_id"))
	if err != nil || h.acceso.ValidarSucursalActiva(c.Request.Context(), userID, businessID, branchID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return
	}
	filtro := compraapp.FiltroExportacion{SucursalID: branchID, IncluirSuspendidas: c.Query("incluir_suspendidas") == "true", IncluirComprobacion: c.Query("comprobacion") == "true"}
	if value := c.Query("proveedor_id"); value != "" {
		providerID, err := uuid.Parse(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El proveedor no es válido"})
			return
		}
		filtro.ProveedorID = &providerID
	}
	for key, target := range map[string]**time.Time{"desde": &filtro.Desde, "hasta": &filtro.Hasta} {
		if value := c.Query(key); value != "" {
			date, err := compraapp.ParseDate(value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La fecha no es válida"})
				return
			}
			*target = &date
		}
	}
	for _, value := range c.QueryArray("compra_id") {
		purchaseID, err := uuid.Parse(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La compra no es válida"})
			return
		}
		filtro.CompraIDs = append(filtro.CompraIDs, purchaseID)
	}
	archivo, _, err := h.service.Exportar(businessID, filtro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible generar el Excel"})
		return
	}
	nombre := "compras-" + time.Now().Format("2006-01-02") + ".xlsx"
	c.Header("Content-Disposition", `attachment; filename="`+nombre+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", archivo)
}

// compraPermitida resuelve la compra y exige acceso a su sucursal, como Get.
func (h *Handler) compraPermitida(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	businessID, ok := id(c, "negocioId")
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	purchaseID, ok := id(c, "compraId")
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	userID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": "Sesión requerida"})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	item, err := h.service.Get(businessID, purchaseID)
	if err != nil || h.acceso.ValidarSucursalActiva(c.Request.Context(), userID, businessID, item.SucursalID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "Compra no encontrada"})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return businessID, purchaseID, userID, true
}

// Edit cambia el documento y, si ningún costo está validado, los productos.
func (h *Handler) Edit(c *gin.Context) {
	businessID, purchaseID, userID, ok := h.compraPermitida(c)
	if !ok {
		return
	}
	var input compraapp.EditarCompraInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los campos de la compra"})
		return
	}
	item, err := h.service.Edit(businessID, userID, purchaseID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Suspend saca las piezas y cancela los costos pendientes; el motivo es obligatorio.
func (h *Handler) Suspend(c *gin.Context) {
	businessID, purchaseID, userID, ok := h.compraPermitida(c)
	if !ok {
		return
	}
	var input compraapp.MotivoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El motivo es obligatorio"})
		return
	}
	item, err := h.service.Suspend(businessID, userID, purchaseID, input.Motivo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Reactivate vuelve a sumar las piezas; el motivo es opcional.
func (h *Handler) Reactivate(c *gin.Context) {
	businessID, purchaseID, userID, ok := h.compraPermitida(c)
	if !ok {
		return
	}
	var input compraapp.MotivoInput
	_ = c.ShouldBindJSON(&input)
	item, err := h.service.Reactivate(businessID, userID, purchaseID, input.Motivo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
