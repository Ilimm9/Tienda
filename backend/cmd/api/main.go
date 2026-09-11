package main

import (
	"log"
	"net/http"

	"tienda/backend/internal/application"
	cuentaapplication "tienda/backend/internal/application/cuenta"
	negocioapplication "tienda/backend/internal/application/negocio"
	"tienda/backend/internal/config"
	"tienda/backend/internal/database"
	"tienda/backend/internal/infrastructure"
	cuentainfra "tienda/backend/internal/infrastructure/cuenta"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"
	cuentahttp "tienda/backend/internal/interfaces/http/cuenta"
	negociohttp "tienda/backend/internal/interfaces/http/negocio"

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
	if cfg.AppEnv == "development" {
		if err := database.SeedDevelopment(db); err != nil {
			log.Fatal(err)
		}
	}
	cuentaRepo := cuentainfra.NewUserRepository(db)
	cuentaHandler := cuentahttp.NewAuthHandler(cuentaapplication.NewAuthService(cuentaRepo), cfg)
	productRepo := infrastructure.NewProductRepository(db)
	precioCheckClient := infrastructure.NewPrecioCheckClient(cfg.PrecioCheckBaseURL, cfg.PrecioCheckAPIKey)
	upcItemDBClient := infrastructure.NewUPCItemDBClient(cfg.UPCItemDBBaseURL)
	productLookup := infrastructure.NewFallbackProductLookup(precioCheckClient, upcItemDBClient)
	productHandler := transporthttp.NewProductHandler(application.NewProductService(productRepo, productLookup))
	negocioRepo := negocioinfra.NewNegocioRepository(db)
	negocioHandler := negociohttp.NewNegocioHandler(negocioapplication.NewNegocioService(negocioRepo))
	sucursalRepo := negocioinfra.NewSucursalRepository(db)
	sucursalHandler := negociohttp.NewSucursalHandler(negocioapplication.NewSucursalService(sucursalRepo))
	router := gin.Default()
	router.Use(transporthttp.CORSMiddleware(cfg.FrontendURL))
	router.GET("/api/v1/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"estado": "ok"}) })
	auth := router.Group("/api/v1/auth")
	auth.POST("/register", cuentaHandler.Register)
	auth.POST("/login", cuentaHandler.Login)
	auth.POST("/logout", cuentaHandler.Logout)
	auth.GET("/me", cuentaHandler.Me)
	negocios := router.Group("/api/v1/negocios")
	negocios.Use(transporthttp.RequireAuth(cfg))
	negocios.GET("", negocioHandler.Listar)
	negocios.POST("", negocioHandler.Crear)
	negocios.GET("/:negocioId", negocioHandler.Obtener)
	negocios.PATCH("/:negocioId", negocioHandler.Actualizar)
	negocios.DELETE("/:negocioId", negocioHandler.Archivar)
	negocios.POST("/:negocioId/restaurar", negocioHandler.Restaurar)
	negocios.GET("/:negocioId/administracion/sucursales", sucursalHandler.Listar)
	negocios.POST("/:negocioId/administracion/sucursales", sucursalHandler.Crear)
	negocios.GET("/:negocioId/administracion/sucursales/:sucursalId", sucursalHandler.Obtener)
	negocios.PATCH("/:negocioId/administracion/sucursales/:sucursalId", sucursalHandler.Actualizar)
	negocios.DELETE("/:negocioId/administracion/sucursales/:sucursalId", sucursalHandler.Archivar)
	negocios.POST("/:negocioId/administracion/sucursales/:sucursalId/restaurar", sucursalHandler.Restaurar)
	router.GET("/api/v1/negocios/:negocioId/catalogo/productos", productHandler.List)
	router.GET("/api/v1/negocios/:negocioId/catalogo/productos/consulta-codigo/:codigoBarras", productHandler.LookupProduct)
	router.GET("/api/v1/negocios/:negocioId/catalogo/productos/consulta-preciocheck/:codigoBarras", productHandler.LookupProduct)
	router.GET("/api/v1/negocios/:negocioId/catalogo/categorias", productHandler.Categories)
	router.GET("/api/v1/negocios/:negocioId/catalogo/marcas", productHandler.Brands)
	router.GET("/api/v1/negocios/:negocioId/sucursales", productHandler.Branches)
	router.POST("/api/v1/negocios/:negocioId/catalogo/productos", productHandler.Create)
	router.GET("/api/v1/catalogo/marcas", productHandler.ListBrandsAdmin)
	router.POST("/api/v1/catalogo/marcas", productHandler.CreateBrand)
	router.PATCH("/api/v1/catalogo/marcas/:id", productHandler.UpdateBrand)
	router.GET("/api/v1/catalogo/categorias", productHandler.ListCategoriesAdmin)
	router.POST("/api/v1/catalogo/categorias", productHandler.CreateCategory)
	router.PATCH("/api/v1/catalogo/categorias/:id", productHandler.UpdateCategory)
	router.GET("/api/v1/negocios/:negocioId/catalogo/proveedores", productHandler.ListProviders)
	router.POST("/api/v1/negocios/:negocioId/catalogo/proveedores", productHandler.CreateProvider)
	router.PATCH("/api/v1/negocios/:negocioId/catalogo/proveedores/:id", productHandler.UpdateProvider)
	log.Printf("API escuchando en http://localhost:%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
