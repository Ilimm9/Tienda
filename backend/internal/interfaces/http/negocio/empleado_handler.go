package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmpleadoHandler struct {
	empleados *application.EmpleadoService
}

func NewEmpleadoHandler(empleados *application.EmpleadoService) *EmpleadoHandler {
	return &EmpleadoHandler{empleados: empleados}
}

func (h *EmpleadoHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.empleados.Listar(c.Request.Context(), usuarioID, negocioID, c.Query("estado"), c.Query("buscar"))
	if err != nil {
		responderErrorEmpleado(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *EmpleadoHandler) Crear(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.CrearEmpleadoInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.empleados.Crear(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorEmpleado(c, err)
		return
	}
	c.JSON(http.StatusCreated, detalle)
}

func (h *EmpleadoHandler) Obtener(c *gin.Context) {
	usuarioID, negocioID, empleadoID, ok := idsEmpleado(c)
	if !ok {
		return
	}
	detalle, err := h.empleados.Obtener(c.Request.Context(), usuarioID, negocioID, empleadoID)
	if err != nil {
		responderErrorEmpleado(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *EmpleadoHandler) Actualizar(c *gin.Context) {
	usuarioID, negocioID, empleadoID, ok := idsEmpleado(c)
	if !ok {
		return
	}
	var input domain.ActualizarEmpleadoInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.empleados.Actualizar(c.Request.Context(), usuarioID, negocioID, empleadoID, input)
	if err != nil {
		responderErrorEmpleado(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func idsEmpleado(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	empleadoID, err := uuid.Parse(c.Param("empleadoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del empleado no es válido.",
			"campos": gin.H{"empleadoId": "debe ser UUID"},
		})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return usuarioID, negocioID, empleadoID, true
}

func responderErrorEmpleado(c *gin.Context, err error) {
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
	case errors.Is(err, application.ErrEmpleadoProhibido):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrEmpleadoSinCambios):
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "SIN_CAMBIOS", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrEmpleadoConflicto), errors.Is(err, application.ErrEstadoNegocio):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CONFLICTO_EMPLEADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
		})
	}
}
