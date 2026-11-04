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

// VistaHandler expone Configuración > Vistas del negocio activo.
type VistaHandler struct {
	vistas *application.VistaService
}

func NewVistaHandler(vistas *application.VistaService) *VistaHandler {
	return &VistaHandler{vistas: vistas}
}

type guardarVistasRequest struct {
	SeccionesOcultas []string `json:"secciones_ocultas"`
}

func (h *VistaHandler) Obtener(c *gin.Context) {
	negocioID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.", "campos": gin.H{}})
		return
	}
	vistas, err := h.vistas.Obtener(c.Request.Context(), negocioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible cargar las vistas.", "campos": gin.H{}})
		return
	}
	c.JSON(http.StatusOK, vistas)
}

func (h *VistaHandler) Guardar(c *gin.Context) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
		return
	}
	negocioID, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.", "campos": gin.H{}})
		return
	}
	var request guardarVistasRequest
	if err := c.ShouldBindJSON(&request); err != nil || request.SeccionesOcultas == nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "SOLICITUD_INVALIDA", "mensaje": "Indica las secciones ocultas.", "campos": gin.H{"secciones_ocultas": "es obligatorio"}})
		return
	}
	vistas, err := h.vistas.Guardar(c.Request.Context(), usuarioID, negocioID, request.SeccionesOcultas)
	switch {
	case errors.Is(err, application.ErrVistasSoloPropietario):
		c.JSON(http.StatusForbidden, gin.H{"codigo": "SOLO_PROPIETARIOS", "mensaje": err.Error(), "campos": gin.H{}})
	case errors.Is(err, domain.ErrSeccionNoOcultable):
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "SECCION_NO_VALIDA", "mensaje": "Solo se pueden ocultar Categorías y Marcas.", "campos": gin.H{"secciones_ocultas": "valor no permitido"}})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible guardar las vistas.", "campos": gin.H{}})
	default:
		c.JSON(http.StatusOK, vistas)
	}
}
