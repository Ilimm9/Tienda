package negocio

import (
	"net/http"

	application "tienda/backend/internal/application/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireNegocioActivo(contexto *application.ContextoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		usuarioID, ok := transporthttp.AuthenticatedUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
			return
		}
		negocioID, err := uuid.Parse(c.Param("negocioId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.", "campos": gin.H{"negocioId": "debe ser UUID"}})
			return
		}
		if err := contexto.ValidarNegocioActivo(c.Request.Context(), usuarioID, negocioID); err != nil {
			responderErrorContexto(c, err)
			return
		}
		c.Next()
	}
}
