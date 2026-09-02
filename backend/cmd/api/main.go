package main

import (
	"log"
	"net/http"

	"tienda/backend/internal/application"
	"tienda/backend/internal/config"
	"tienda/backend/internal/database"
	"tienda/backend/internal/infrastructure"
	authhttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Init(db); err != nil {
		log.Fatal(err)
	}
	repo := infrastructure.NewUserRepository(db)
	handler := authhttp.NewAuthHandler(application.NewAuthService(repo), cfg)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.FrontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	router.GET("/api/v1/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"estado": "ok"}) })
	auth := router.Group("/api/v1/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/logout", handler.Logout)
	auth.GET("/me", handler.Me)
	log.Printf("API escuchando en http://localhost:%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
