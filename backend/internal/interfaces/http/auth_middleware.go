package http

import (
	"errors"
	"net/http"

	"tienda/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const authenticatedUserKey = "authenticated_user_id"

func RequireAuth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, err := c.Cookie("tienda_session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{}})
			return
		}
		userID, _, err := parseSession(value, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"codigo": "SESION_INVALIDA", "mensaje": "La sesión no es válida o expiró.", "campos": gin.H{}})
			return
		}
		c.Set(authenticatedUserKey, userID)
		c.Next()
	}
}

func parseSession(value, secret string) (uuid.UUID, jwt.MapClaims, error) {
	token, err := jwt.Parse(value, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("algoritmo de sesión inválido")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return uuid.Nil, nil, errors.New("sesión inválida")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, nil, errors.New("claims inválidos")
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, nil, err
	}
	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return userID, claims, nil
}

func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(authenticatedUserKey)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}
