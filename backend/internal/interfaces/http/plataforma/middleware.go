package plataforma

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/plataforma"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireAdminLectura(service *application.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := transporthttp.AuthenticatedUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
			return
		}
		if err := service.Autorizar(c.Request.Context(), userID); err != nil {
			if errors.Is(err, application.ErrAccesoDenegado) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"codigo": "ACCESO_PLATAFORMA_DENEGADO", "mensaje": "No tienes acceso a la administración de plataforma.", "campos": gin.H{}})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible validar el acceso.", "campos": gin.H{}})
			}
			return
		}
		c.Next()
	}
}

// RejectPlatformAdmin conserva la separación de funciones: una cuenta operativa global no puede crear, aceptar invitaciones ni operar como miembro de un tenant.
func RejectPlatformAdmin(service *application.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := transporthttp.AuthenticatedUserID(c)
		if !ok {
			c.Next()
			return
		}
		roles, err := service.RolesUsuario(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible validar el acceso.", "campos": gin.H{}})
			return
		}
		for _, role := range roles {
			if role == "ADMIN_LECTURA" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"codigo": "SEPARACION_PLATAFORMA", "mensaje": "Una cuenta de plataforma no puede operar tiendas.", "campos": gin.H{}})
				return
			}
		}
		c.Next()
	}
}

// RegistrarActividadNegocio agrega únicamente solicitudes autenticadas exitosas. Nunca persiste rutas, parámetros, cuerpos, IP o agente de usuario.
func RegistrarActividadNegocio(service *application.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Writer.Status() < 200 || c.Writer.Status() >= 400 {
			return
		}
		userID, ok := transporthttp.AuthenticatedUserID(c)
		if !ok {
			return
		}
		businessID, err := uuid.Parse(c.Param("negocioId"))
		if err != nil {
			return
		}
		_ = service.RegistrarActividad(c.Request.Context(), userID, businessID)
	}
}
