package http

import (
	"net/http"

	cuentaapplication "tienda/backend/internal/application/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	authenticatedUserKey    = "authenticated_user_id"
	authenticatedEmailKey   = "authenticated_user_email"
	authenticatedSessionKey = "authenticated_session"
)

type SessionAuthenticator interface {
	Authenticate(token string) (cuentaapplication.AuthenticatedSession, error)
}

type SessionAuthenticatorFunc func(token string) (cuentaapplication.AuthenticatedSession, error)

func (f SessionAuthenticatorFunc) Authenticate(token string) (cuentaapplication.AuthenticatedSession, error) {
	return f(token)
}

func RequireAuth(authenticator SessionAuthenticator, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, err := c.Cookie(cookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
			return
		}
		session, err := authenticator.Authenticate(value)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_INVALIDA", "mensaje": "La sesión no es válida o expiró.", "campos": gin.H{}})
			return
		}
		c.Set(authenticatedUserKey, session.UserID)
		c.Set(authenticatedEmailKey, session.Email)
		c.Set(authenticatedSessionKey, session)
		c.Next()
	}
}

func AuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(authenticatedUserKey)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}

func AuthenticatedEmail(c *gin.Context) (string, bool) {
	value, exists := c.Get(authenticatedEmailKey)
	if !exists {
		return "", false
	}
	email, ok := value.(string)
	return email, ok
}

func authenticatedSession(c *gin.Context) (cuentaapplication.AuthenticatedSession, bool) {
	value, exists := c.Get(authenticatedSessionKey)
	if !exists {
		return cuentaapplication.AuthenticatedSession{}, false
	}
	session, ok := value.(cuentaapplication.AuthenticatedSession)
	return session, ok
}
