package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

type ContextoHandler struct {
	contexto *application.ContextoService
}

func NewContextoHandler(contexto *application.ContextoService) *ContextoHandler {
	return &ContextoHandler{contexto: contexto}
}

func (h *ContextoHandler) Opciones(c *gin.Context) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
		return
	}
	items, err := h.contexto.ListarOpciones(c.Request.Context(), usuarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible cargar opciones de contexto.", "campos": gin.H{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func responderErrorContexto(c *gin.Context, err error) {
	if errors.Is(err, application.ErrContextoNegocioNoDisponible) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"codigo": "NEGOCIO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el negocio solicitado.", "campos": gin.H{}})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{}})
}
