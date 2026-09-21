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
	productService := application.NewProductService(productRepo, productLookup)
	negocioRepo := negocioinfra.NewNegocioRepository(db)
	negocioHandler := negociohttp.NewNegocioHandler(negocioapplication.NewNegocioService(negocioRepo))
	sucursalRepo := negocioinfra.NewSucursalRepository(db)
	sucursalHandler := negociohttp.NewSucursalHandler(negocioapplication.NewSucursalService(sucursalRepo))
	contextoRepo := negocioinfra.NewContextoRepository(db)
	contextoService := negocioapplication.NewContextoService(contextoRepo)
	contextoHandler := negociohttp.NewContextoHandler(contextoService)
	rolRepo := negocioinfra.NewRolRepository(db)
	rolService := negocioapplication.NewRolService(rolRepo)
	rolHandler := negociohttp.NewRolHandler(rolService)
	empleadoRepo := negocioinfra.NewEmpleadoRepository(db)
	empleadoHandler := negociohttp.NewEmpleadoHandler(negocioapplication.NewEmpleadoService(empleadoRepo))
	invitacionRepo := negocioinfra.NewInvitacionRepository(db)
	invitacionHandler := negociohttp.NewInvitacionHandler(negocioapplication.NewInvitacionService(invitacionRepo))
	asignacionRepo := negocioinfra.NewAsignacionRepository(db)
	asignacionHandler := negociohttp.NewAsignacionHandler(negocioapplication.NewAsignacionService(asignacionRepo))
	productHandler := transporthttp.NewProductHandler(productService, contextoService)
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
	negocios.GET("/:negocioId/administracion/permisos", rolHandler.Permisos)
	negocios.GET("/:negocioId/administracion/mis-permisos", rolHandler.MisPermisos)
	negocios.GET("/:negocioId/administracion/roles", rolHandler.Listar)
	negocios.POST("/:negocioId/administracion/roles", rolHandler.Crear)
	negocios.GET("/:negocioId/administracion/roles/:rolId", rolHandler.Obtener)
	negocios.PATCH("/:negocioId/administracion/roles/:rolId", rolHandler.Actualizar)
	negocios.DELETE("/:negocioId/administracion/roles/:rolId", rolHandler.Eliminar)
	negocios.GET("/:negocioId/administracion/miembros", rolHandler.ListarMiembros)
	negocios.PUT("/:negocioId/administracion/miembros/:membresiaId/roles", rolHandler.AsignarRoles)
	negocios.GET("/:negocioId/administracion/empleados", empleadoHandler.Listar)
	negocios.POST("/:negocioId/administracion/empleados", empleadoHandler.Crear)
	negocios.GET("/:negocioId/administracion/empleados/:empleadoId", empleadoHandler.Obtener)
	negocios.PATCH("/:negocioId/administracion/empleados/:empleadoId", empleadoHandler.Actualizar)
	negocios.GET("/:negocioId/administracion/invitaciones", invitacionHandler.Listar)
	negocios.POST("/:negocioId/administracion/invitaciones", invitacionHandler.Crear)
	negocios.DELETE("/:negocioId/administracion/invitaciones/:invitacionId", invitacionHandler.Cancelar)
	negocios.GET("/:negocioId/administracion/empleados/:empleadoId/sucursales", asignacionHandler.Listar)
	negocios.POST("/:negocioId/administracion/empleados/:empleadoId/sucursales", asignacionHandler.Asignar)
	negocios.POST("/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId/principal", asignacionHandler.EstablecerPrincipal)
	negocios.DELETE("/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId", asignacionHandler.Finalizar)
	// La consulta del enlace es pública: quien lo abre todavía puede no tener cuenta.
	router.GET("/api/v1/invitaciones/:token", invitacionHandler.Consultar)
	// Aceptar sí exige sesión iniciada con el correo invitado.
	invitaciones := router.Group("/api/v1/invitaciones")
	invitaciones.Use(transporthttp.RequireAuth(cfg))
	invitaciones.POST("/:token/aceptar", invitacionHandler.Aceptar)

	contexto := router.Group("/api/v1/contexto")
	contexto.Use(transporthttp.RequireAuth(cfg))
	contexto.GET("/opciones", contextoHandler.Opciones)
	negocioActual := router.Group("/api/v1/negocios/:negocioId")
	negocioActual.Use(transporthttp.RequireAuth(cfg), negociohttp.RequireNegocioActivo(contextoService))
	negocioActual.GET("/catalogo/productos", productHandler.List)
	negocioActual.GET("/catalogo/productos/consulta-codigo/:codigoBarras", productHandler.LookupProduct)
	negocioActual.GET("/catalogo/productos/consulta-preciocheck/:codigoBarras", productHandler.LookupProduct)
	negocioActual.GET("/catalogo/categorias", productHandler.Categories)
	negocioActual.GET("/catalogo/marcas", productHandler.Brands)
	negocioActual.GET("/sucursales", productHandler.Branches)
	negocioActual.POST("/catalogo/productos", productHandler.Create)
	negocioActual.PATCH("/catalogo/productos/:productoId", productHandler.Update)
	negocioActual.DELETE("/catalogo/productos/:productoId", productHandler.Deactivate)
	negocioActual.GET("/catalogo/productos/importacion/plantilla", productHandler.ProductImportTemplate)
	negocioActual.POST("/catalogo/productos/validar-importacion", productHandler.PreviewProductImport)
	negocioActual.POST("/catalogo/productos/importar", productHandler.ImportProducts)
	negocioActual.GET("/catalogo/productos/importaciones/:importacionId", productHandler.ProductImportStatus)
	router.GET("/api/v1/catalogo/marcas", productHandler.ListBrandsAdmin)
	router.POST("/api/v1/catalogo/marcas", productHandler.CreateBrand)
	router.PATCH("/api/v1/catalogo/marcas/:id", productHandler.UpdateBrand)
	router.GET("/api/v1/catalogo/marcas/importacion/plantilla", productHandler.BrandImportTemplate)
	router.POST("/api/v1/catalogo/marcas/importar", productHandler.ImportBrands)
	router.GET("/api/v1/catalogo/categorias", productHandler.ListCategoriesAdmin)
	router.POST("/api/v1/catalogo/categorias", productHandler.CreateCategory)
	router.PATCH("/api/v1/catalogo/categorias/:id", productHandler.UpdateCategory)
	router.GET("/api/v1/catalogo/unidades-medida", productHandler.ListUnits)
	router.POST("/api/v1/catalogo/unidades-medida", productHandler.CreateUnit)
	router.PATCH("/api/v1/catalogo/unidades-medida/:id", productHandler.UpdateUnit)
	router.GET("/api/v1/catalogo/categorias/importacion/plantilla", productHandler.CategoryImportTemplate)
	router.POST("/api/v1/catalogo/categorias/importar", productHandler.ImportCategories)
	router.GET("/api/v1/catalogo/unidades/importacion/plantilla", productHandler.UnitImportTemplate)
	router.POST("/api/v1/catalogo/unidades/importar", productHandler.ImportUnits)
	negocioActual.GET("/catalogo/proveedores", productHandler.ListProviders)
	negocioActual.POST("/catalogo/proveedores", productHandler.CreateProvider)
	negocioActual.PATCH("/catalogo/proveedores/:id", productHandler.UpdateProvider)
	log.Printf("API escuchando en http://localhost:%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
