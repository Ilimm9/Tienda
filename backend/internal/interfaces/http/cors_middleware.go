package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == frontendURL {
			c.Header("Access-Control-Allow-Origin", frontendURL)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-XSRF-TOKEN")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "600")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
