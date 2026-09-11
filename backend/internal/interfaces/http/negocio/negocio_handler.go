package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NegocioHandler struct {
	negocios *application.NegocioService
}

func NewNegocioHandler(negocios *application.NegocioService) *NegocioHandler {
	return &NegocioHandler{negocios: negocios}
}

func (h *NegocioHandler) Listar(c *gin.Context) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		responderErrorNegocio(c, errors.New("sesión sin usuario"))
		return
	}
	items, err := h.negocios.Listar(c.Request.Context(), usuarioID, c.DefaultQuery("estado", "activo"))
	if err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *NegocioHandler) Crear(c *gin.Context) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		responderErrorNegocio(c, errors.New("sesión sin usuario"))
		return
	}
	var input domain.CrearNegocioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "DATOS_INVALIDOS", "mensaje": "Revisa los datos enviados.", "campos": gin.H{}})
		return
	}
	detalle, err := h.negocios.Crear(c.Request.Context(), usuarioID, input)
	if err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.JSON(http.StatusCreated, detalle)
}

func (h *NegocioHandler) Obtener(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	detalle, err := h.negocios.Obtener(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *NegocioHandler) Actualizar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.ActualizarNegocioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "DATOS_INVALIDOS", "mensaje": "Revisa los datos enviados.", "campos": gin.H{}})
		return
	}
	detalle, err := h.negocios.Actualizar(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *NegocioHandler) Archivar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	if err := h.negocios.Archivar(c.Request.Context(), usuarioID, negocioID); err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NegocioHandler) Restaurar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	detalle, err := h.negocios.Restaurar(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorNegocio(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func idsNegocio(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		responderErrorNegocio(c, errors.New("sesión sin usuario"))
		return uuid.Nil, uuid.Nil, false
	}
	negocioID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.", "campos": gin.H{"negocioId": "debe ser UUID"}})
		return uuid.Nil, uuid.Nil, false
	}
	return usuarioID, negocioID, true
}

func responderErrorNegocio(c *gin.Context, err error) {
	var validation *application.ErrorValidacion
	switch {
	case errors.As(err, &validation):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"codigo": "DATOS_INVALIDOS", "mensaje": validation.Error(), "campos": validation.Campos})
	case errors.Is(err, application.ErrNegocioNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"codigo": "NEGOCIO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el negocio solicitado.", "campos": gin.H{}})
	case errors.Is(err, application.ErrNegocioProhibido):
		c.JSON(http.StatusForbidden, gin.H{"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{}})
	case errors.Is(err, application.ErrNegocioSinCambios):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "SIN_CAMBIOS", "mensaje": err.Error(), "campos": gin.H{}})
	case errors.Is(err, application.ErrEstadoNegocio), errors.Is(err, application.ErrNegocioConflicto):
		c.JSON(http.StatusConflict, gin.H{"codigo": "CONFLICTO_NEGOCIO", "mensaje": err.Error(), "campos": gin.H{}})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{}})
	}
}
