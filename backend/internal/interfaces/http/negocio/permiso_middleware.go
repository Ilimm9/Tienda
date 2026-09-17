package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequierePermiso exige que la membresía activa concentre el permiso indicado sobre el negocio de la ruta.
//
// Sustituye la regla temporal de fases 1 a 3, donde solo `tipo_miembro = propietario` podía mutar.
func RequierePermiso(roles *application.RolService, permiso string) gin.HandlerFunc {
	return func(c *gin.Context) {
		usuarioID, ok := transporthttp.AuthenticatedUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{},
			})
			return
		}
		negocioID, err := uuid.Parse(c.Param("negocioId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.",
				"campos": gin.H{"negocioId": "debe ser UUID"},
			})
			return
		}
		if err := roles.Autorizar(c.Request.Context(), usuarioID, negocioID, permiso); err != nil {
			if errors.Is(err, application.ErrRolProhibido) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"codigo": "ACCESO_DENEGADO", "mensaje": "No tienes permiso para realizar esta acción.", "campos": gin.H{},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
			})
			return
		}
		c.Next()
	}
}
