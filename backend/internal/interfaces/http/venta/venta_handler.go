package venta

import (
	"context"
	"errors"
	"net/http"
	"time"

	ventaapp "tienda/backend/internal/application/venta"
	negociodomain "tienda/backend/internal/domain/negocio"
	ventadomain "tienda/backend/internal/domain/venta"
	transport "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AccesoSucursal es el contrato hacia `negocio`: qué sucursales puede operar el usuario.
type AccesoSucursal interface {
	ValidarSucursalActiva(ctx context.Context, usuarioID, negocioID, sucursalID uuid.UUID) error
	SucursalesAccesibles(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]uuid.UUID, error)
}

// Permisos resuelve los permisos efectivos para decidir costos y ventas propias.
type Permisos interface {
	PermisosEfectivos(ctx context.Context, usuarioID, negocioID uuid.UUID) ([]string, error)
}

type Handler struct {
	service  *ventaapp.Service
	acceso   AccesoSucursal
	permisos Permisos
}

func NewHandler(service *ventaapp.Service, acceso AccesoSucursal, permisos Permisos) *Handler {
	return &Handler{service: service, acceso: acceso, permisos: permisos}
}

// sesion reúne negocio, usuario y permisos; responde el error y `false` si no puede resolverlos.
func (h *Handler) sesion(c *gin.Context) (uuid.UUID, uuid.UUID, map[string]bool, bool) {
	negocioID, ok := id(c, "negocioId")
	if !ok {
		return uuid.Nil, uuid.Nil, nil, false
	}
	usuarioID, ok := transport.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Sesión requerida"})
		return uuid.Nil, uuid.Nil, nil, false
	}
	codigos, err := h.permisos.PermisosEfectivos(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible comprobar tus permisos"})
		return uuid.Nil, uuid.Nil, nil, false
	}
	permisos := make(map[string]bool, len(codigos))
	for _, codigo := range codigos {
		permisos[codigo] = true
	}
	return negocioID, usuarioID, permisos, true
}

// alcance resuelve sucursales visibles y, sin `ventas.ver`, limita a las ventas propias.
func (h *Handler) alcance(c *gin.Context, negocioID, usuarioID uuid.UUID, permisos map[string]bool) ([]uuid.UUID, *uuid.UUID, bool) {
	if !permisos[negociodomain.PermisoVentaVer] && !permisos[negociodomain.PermisoVentaRegistrar] {
		c.JSON(http.StatusForbidden, gin.H{"codigo": "ACCESO_DENEGADO", "mensaje": "No tienes permiso para realizar esta acción."})
		return nil, nil, false
	}
	sucursales, err := h.acceso.SucursalesAccesibles(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible comprobar el acceso a la sucursal"})
		return nil, nil, false
	}
	if permisos[negociodomain.PermisoVentaVer] {
		return sucursales, nil, true
	}
	return sucursales, &usuarioID, true
}

func (h *Handler) Productos(c *gin.Context) {
	negocioID, usuarioID, permisos, ok := h.sesion(c)
	if !ok {
		return
	}
	sucursalID, err := uuid.Parse(c.Query("sucursal_id"))
	if err != nil || h.acceso.ValidarSucursalActiva(c.Request.Context(), usuarioID, negocioID, sucursalID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return
	}
	resultado, err := h.service.Buscar(negocioID, sucursalID, c.Query("q"), permisos[negociodomain.PermisoVentaVerCostos])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible buscar productos"})
		return
	}
	c.JSON(http.StatusOK, resultado)
}

func (h *Handler) Registrar(c *gin.Context) {
	negocioID, usuarioID, permisos, ok := h.sesion(c)
	if !ok {
		return
	}
	var input ventaapp.RegistrarInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "VENTA_INVALIDA", "mensaje": "Revisa los datos de la venta"})
		return
	}
	sucursalID, err := uuid.Parse(input.SucursalID)
	if err != nil || h.acceso.ValidarSucursalActiva(c.Request.Context(), usuarioID, negocioID, sucursalID) != nil {
		c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
		return
	}
	creada, err := h.service.Registrar(negocioID, usuarioID, input)
	if err != nil {
		responderError(c, err)
		return
	}
	if !permisos[negociodomain.PermisoVentaVerCostos] {
		ventaapp.OcultarCostos(&creada)
	}
	c.JSON(http.StatusCreated, creada)
}

func (h *Handler) List(c *gin.Context) {
	negocioID, usuarioID, permisos, ok := h.sesion(c)
	if !ok {
		return
	}
	sucursales, propias, ok := h.alcance(c, negocioID, usuarioID, permisos)
	if !ok {
		return
	}
	filtro := ventaapp.Filtro{Sucursales: sucursales, UsuarioID: propias, Folio: c.Query("folio"), FormaPago: c.Query("forma_pago"), Estado: c.Query("estado"), Periodo: c.Query("periodo")}
	if texto := c.Query("sucursal_id"); texto != "" {
		sucursalID, err := uuid.Parse(texto)
		if err != nil || !contiene(sucursales, sucursalID) {
			c.JSON(http.StatusNotFound, gin.H{"mensaje": "No fue posible encontrar la sucursal solicitada"})
			return
		}
		filtro.Sucursales = []uuid.UUID{sucursalID}
	}
	if texto := c.Query("usuario_id"); texto != "" && propias == nil {
		cajero, err := uuid.Parse(texto)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El cajero no es válido"})
			return
		}
		filtro.UsuarioID = &cajero
	}
	for clave, destino := range map[string]**time.Time{"desde": &filtro.Desde, "hasta": &filtro.Hasta} {
		if texto := c.Query(clave); texto != "" {
			fecha, err := time.Parse(time.RFC3339, texto)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"mensaje": "La fecha no es válida"})
				return
			}
			*destino = &fecha
		}
	}
	listado, err := h.service.Listar(negocioID, filtro)
	if errors.Is(err, ventaapp.ErrPeriodoInvalido) {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El periodo no es válido"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar las ventas"})
		return
	}
	c.JSON(http.StatusOK, listado)
}

func (h *Handler) Get(c *gin.Context) {
	negocioID, usuarioID, permisos, ok := h.sesion(c)
	if !ok {
		return
	}
	ventaID, ok := id(c, "ventaId")
	if !ok {
		return
	}
	sucursales, propias, ok := h.alcance(c, negocioID, usuarioID, permisos)
	if !ok {
		return
	}
	item, err := h.service.Obtener(negocioID, ventaID)
	// Una venta de otra sucursal o de otro cajero responde igual que una inexistente.
	if err != nil || !contiene(sucursales, item.SucursalID) || (propias != nil && item.UsuarioID != *propias) {
		if err != nil && !errors.Is(err, ventaapp.ErrVentaNoEncontrada) {
			c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cargar la venta"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"codigo": "VENTA_NO_ENCONTRADA", "mensaje": "Venta no encontrada"})
		return
	}
	if !permisos[negociodomain.PermisoVentaVerCostos] {
		ventaapp.OcultarCostos(&item)
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) Cancelar(c *gin.Context) {
	negocioID, usuarioID, permisos, ok := h.sesion(c)
	if !ok {
		return
	}
	ventaID, ok := id(c, "ventaId")
	if !ok {
		return
	}
	var input struct {
		Motivo  string `json:"motivo"`
		Detalle string `json:"detalle"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "MOTIVO_REQUERIDO", "mensaje": ventaapp.ErrMotivoRequerido.Error()})
		return
	}
	sucursales, err := h.acceso.SucursalesAccesibles(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible comprobar el acceso a la sucursal"})
		return
	}
	cancelada, err := h.service.Cancelar(ventaapp.CancelarComando{NegocioID: negocioID, VentaID: ventaID, UsuarioID: usuarioID, MotivoCodigo: input.Motivo, Detalle: input.Detalle, Sucursales: sucursales})
	if err != nil {
		responderError(c, err)
		return
	}
	if !permisos[negociodomain.PermisoVentaVerCostos] {
		ventaapp.OcultarCostos(&cancelada)
	}
	c.JSON(http.StatusOK, cancelada)
}

func responderError(c *gin.Context, err error) {
	var existencia *ventadomain.ExistenciaInsuficienteError
	switch {
	case errors.As(err, &existencia):
		c.JSON(http.StatusConflict, gin.H{"codigo": "EXISTENCIA_INSUFICIENTE", "mensaje": existencia.Error(),
			"producto": existencia.Producto, "disponible": existencia.Disponible})
	case errors.Is(err, ventadomain.ErrPagoInsuficiente):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "PAGO_INSUFICIENTE", "mensaje": "El importe recibido es menor al total"})
	case errors.Is(err, ventaapp.ErrProductoNoDisponible):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "PRODUCTO_NO_DISPONIBLE", "mensaje": "Uno de los productos ya no está disponible para venta en esta sucursal"})
	case errors.Is(err, ventaapp.ErrVentaNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{"codigo": "VENTA_NO_ENCONTRADA", "mensaje": "Venta no encontrada"})
	case errors.Is(err, ventaapp.ErrVentaYaCancelada):
		c.JSON(http.StatusConflict, gin.H{"codigo": "VENTA_YA_CANCELADA", "mensaje": "La venta ya está cancelada"})
	case errors.Is(err, ventaapp.ErrVentaFueraDePlazo):
		c.JSON(http.StatusConflict, gin.H{"codigo": "VENTA_FUERA_DE_PLAZO", "mensaje": "Solo se pueden cancelar ventas del mismo día"})
	case errors.Is(err, ventadomain.ErrMotivoInvalido), errors.Is(err, ventadomain.ErrDetalleRequerido), errors.Is(err, ventaapp.ErrMotivoRequerido):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "MOTIVO_REQUERIDO", "mensaje": err.Error()})
	case errors.Is(err, ventaapp.ErrVentaInvalida), errors.Is(err, ventaapp.ErrSucursalInvalida), errors.Is(err, ventaapp.ErrReferenciaLarga), errors.Is(err, ventadomain.ErrFormaPagoInvalida), errors.Is(err, ventadomain.ErrImporteInvalido):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "VENTA_INVALIDA", "mensaje": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la venta"})
	}
}

func contiene(ids []uuid.UUID, buscado uuid.UUID) bool {
	for _, item := range ids {
		if item == buscado {
			return true
		}
	}
	return false
}

func id(c *gin.Context, key string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(c.Param(key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "El identificador no es válido"})
		return uuid.Nil, false
	}
	return parsed, true
}
