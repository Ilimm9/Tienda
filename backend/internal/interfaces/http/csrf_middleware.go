package http

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	cuentaapplication "tienda/backend/internal/application/cuenta"

	"github.com/gin-gonic/gin"
)

const CSRFCookieName = "XSRF-TOKEN"

func RequireTrustedOrigin(frontendURL string) gin.HandlerFunc {
	allowed, err := url.Parse(frontendURL)
	allowedOrigin := ""
	if err == nil && allowed.Scheme != "" && allowed.Host != "" {
		allowedOrigin = allowed.Scheme + "://" + allowed.Host
	}
	return func(c *gin.Context) {
		if !isUnsafeMethod(c.Request.Method) {
			c.Next()
			return
		}
		if allowedOrigin == "" || c.GetHeader("Origin") != allowedOrigin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"codigo": "ORIGEN_NO_PERMITIDO", "mensaje": "El origen de la solicitud no está permitido.", "campos": gin.H{}})
			return
		}
		c.Next()
	}
}

func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isUnsafeMethod(c.Request.Method) {
			c.Next()
			return
		}
		session, ok := authenticatedSession(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
			return
		}
		cookieToken, err := c.Cookie(CSRFCookieName)
		headerToken := strings.TrimSpace(c.GetHeader("X-XSRF-TOKEN"))
		if err != nil || cookieToken == "" || headerToken == "" ||
			subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 ||
			subtle.ConstantTimeCompare([]byte(cuentaapplication.HashSessionToken(headerToken)), []byte(session.CSRFHash)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"codigo": "CSRF_INVALIDO", "mensaje": "La validación de seguridad de la solicitud falló.", "campos": gin.H{}})
			return
		}
		c.Next()
	}
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
