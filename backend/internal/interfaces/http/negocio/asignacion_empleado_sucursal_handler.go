package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AsignacionHandler struct {
	asignaciones *application.AsignacionService
}

func NewAsignacionHandler(asignaciones *application.AsignacionService) *AsignacionHandler {
	return &AsignacionHandler{asignaciones: asignaciones}
}

func (h *AsignacionHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, empleadoID, ok := idsEmpleado(c)
	if !ok {
		return
	}
	items, err := h.asignaciones.Listar(c.Request.Context(), usuarioID, negocioID, empleadoID,
		c.Query("estado") == "todos")
	if err != nil {
		responderErrorAsignacion(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *AsignacionHandler) Asignar(c *gin.Context) {
	usuarioID, negocioID, empleadoID, ok := idsEmpleado(c)
	if !ok {
		return
	}
	var input domain.AsignarSucursalInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	items, err := h.asignaciones.Asignar(c.Request.Context(), usuarioID, negocioID, empleadoID, input)
	if err != nil {
		responderErrorAsignacion(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"items": items, "total": len(items)})
}

func (h *AsignacionHandler) EstablecerPrincipal(c *gin.Context) {
	usuarioID, negocioID, empleadoID, asignacionID, ok := idsAsignacion(c)
	if !ok {
		return
	}
	items, err := h.asignaciones.EstablecerPrincipal(c.Request.Context(), usuarioID, negocioID, empleadoID, asignacionID)
	if err != nil {
		responderErrorAsignacion(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *AsignacionHandler) Finalizar(c *gin.Context) {
	usuarioID, negocioID, empleadoID, asignacionID, ok := idsAsignacion(c)
	if !ok {
		return
	}
	if err := h.asignaciones.Finalizar(c.Request.Context(), usuarioID, negocioID, empleadoID, asignacionID); err != nil {
		responderErrorAsignacion(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func idsAsignacion(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	usuarioID, negocioID, empleadoID, ok := idsEmpleado(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	asignacionID, err := uuid.Parse(c.Param("asignacionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador de la asignación no es válido.",
			"campos": gin.H{"asignacionId": "debe ser UUID"},
		})
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return usuarioID, negocioID, empleadoID, asignacionID, true
}

func responderErrorAsignacion(c *gin.Context, err error) {
	var validation *application.ErrorValidacion
	switch {
	case errors.As(err, &validation):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"codigo": "DATOS_INVALIDOS", "mensaje": validation.Error(), "campos": validation.Campos,
		})
	case errors.Is(err, application.ErrNegocioNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "NEGOCIO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el negocio solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrEmpleadoNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "EMPLEADO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el empleado solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrAsignacionNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "ASIGNACION_NO_ENCONTRADA", "mensaje": "No fue posible encontrar la asignación solicitada.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrAsignacionSucursalAjena):
		// Una sucursal de otro negocio es indistinguible de una inexistente.
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "SUCURSAL_NO_ENCONTRADA", "mensaje": "No fue posible encontrar la sucursal solicitada.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrAsignacionProhibida):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrAsignacionDuplicada), errors.Is(err, application.ErrEstadoNegocio):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CONFLICTO_ASIGNACION", "mensaje": err.Error(), "campos": gin.H{},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
		})
	}
}
