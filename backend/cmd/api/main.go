package main

import (
	"log"
	"net/http"
	"time"

	"tienda/backend/internal/application"
	compraapplication "tienda/backend/internal/application/compra"
	correoapplication "tienda/backend/internal/application/correo"
	cuentaapplication "tienda/backend/internal/application/cuenta"
	negocioapplication "tienda/backend/internal/application/negocio"
	precioapplication "tienda/backend/internal/application/precio"
	"tienda/backend/internal/config"
	"tienda/backend/internal/database"
	negociodomain "tienda/backend/internal/domain/negocio"
	"tienda/backend/internal/infrastructure"
	comprainfra "tienda/backend/internal/infrastructure/compra"
	correoinfra "tienda/backend/internal/infrastructure/correo"
	cuentainfra "tienda/backend/internal/infrastructure/cuenta"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"
	precioinfra "tienda/backend/internal/infrastructure/precio"
	transporthttp "tienda/backend/internal/interfaces/http"
	comprahttp "tienda/backend/internal/interfaces/http/compra"
	cuentahttp "tienda/backend/internal/interfaces/http/cuenta"
	negociohttp "tienda/backend/internal/interfaces/http/negocio"
	preciohttp "tienda/backend/internal/interfaces/http/precio"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
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
	sessionRepo := cuentainfra.NewSessionRepository(db)
	verificationRepo := cuentainfra.NewVerificationRepository(db)
	sessionService := cuentaapplication.NewSessionService(sessionRepo, cuentaapplication.SessionConfig{
		Duration:              cfg.SessionDuration,
		RememberDuration:      cfg.RememberDuration,
		RememberIdleDuration:  cfg.RememberIdle,
		ActivityTouchInterval: cfg.SessionTouchInterval,
	})
	go func() {
		ticker := time.NewTicker(cfg.SessionCleanup)
		defer ticker.Stop()
		for range ticker.C {
			if err := sessionService.CleanupInactive(cfg.SessionRetention); err != nil {
				log.Printf("no fue posible depurar sesiones inactivas: %v", err)
			}
		}
	}()
	mailer, err := newMailer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	verificationService := cuentaapplication.NewVerificationService(cuentaRepo, verificationRepo, mailer, cuentaapplication.VerificationConfig{
		HMACSecret: cfg.OTPHMACSecret, TTL: cfg.OTPTTL, MaxAttempts: cfg.OTPMaxAttempts,
		ResendWait: cfg.OTPResendWait, HourlySendMax: cfg.OTPHourlySendMax,
	})
	passwordResetHandler := cuentahttp.NewPasswordResetHandler(cuentaapplication.NewPasswordResetService(cuentaRepo, verificationRepo, mailer, cuentaapplication.PasswordResetConfig{
		HMACSecret: cfg.OTPHMACSecret, TTL: cfg.PasswordResetTTL, MaxAttempts: cfg.OTPMaxAttempts,
		ResendWait: cfg.OTPResendWait, HourlySendMax: cfg.OTPHourlySendMax,
	}))
	cuentaHandler := cuentahttp.NewAuthHandler(cuentaapplication.NewAuthService(cuentaRepo), sessionService, verificationService, cfg)
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
	invitacionHandler := negociohttp.NewInvitacionHandler(negocioapplication.NewInvitacionService(invitacionRepo, mailer, registroInvitacion{verificacion: verificationService}))
	asignacionRepo := negocioinfra.NewAsignacionRepository(db)
	asignacionHandler := negociohttp.NewAsignacionHandler(negocioapplication.NewAsignacionService(asignacionRepo))
	productHandler := transporthttp.NewProductHandler(productService, contextoService)
	compraHandler := comprahttp.NewHandler(compraapplication.NewService(comprainfra.NewRepository(db)), contextoService)
	precioHandler := preciohttp.NewHandler(precioapplication.NewService(precioinfra.NewRepository(db)), contextoService)
	router := gin.Default()
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}); err != nil {
		log.Fatal(err)
	}
	router.Use(transporthttp.CORSMiddleware(cfg.FrontendURL))
	router.Use(transporthttp.RequireTrustedOrigin(cfg.FrontendURL))
	router.GET("/api/v1/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"estado": "ok"}) })
	auth := router.Group("/api/v1/auth")
	// El registro informa si un correo ya tiene cuenta; el límite por IP frena el sondeo masivo.
	auth.POST("/register", transporthttp.LimitarPorIP(10, time.Minute), cuentaHandler.Register)
	auth.POST("/login", cuentaHandler.Login)
	auth.POST("/verificar-correo", cuentaHandler.VerifyEmail)
	auth.POST("/reenviar-verificacion", cuentaHandler.ResendVerification)
	auth.POST("/solicitar-recuperacion", passwordResetHandler.Request)
	auth.POST("/restablecer-contrasena", passwordResetHandler.Reset)
	authProtected := auth.Group("")
	authProtected.Use(transporthttp.RequireAuth(sessionService, cfg.SessionCookieName()), transporthttp.RequireCSRF())
	authProtected.POST("/logout", cuentaHandler.Logout)
	authProtected.GET("/me", cuentaHandler.Me)
	negocios := router.Group("/api/v1/negocios")
	negocios.Use(transporthttp.RequireAuth(sessionService, cfg.SessionCookieName()), transporthttp.RequireCSRF())
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
	negocios.POST("/:negocioId/administracion/invitaciones/:invitacionId/reenviar", invitacionHandler.Reenviar)
	negocios.GET("/:negocioId/administracion/empleados/:empleadoId/sucursales", asignacionHandler.Listar)
	negocios.POST("/:negocioId/administracion/empleados/:empleadoId/sucursales", asignacionHandler.Asignar)
	negocios.POST("/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId/principal", asignacionHandler.EstablecerPrincipal)
	negocios.DELETE("/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId", asignacionHandler.Finalizar)
	// Consulta y registro son públicos: quien abre el enlace todavía puede no tener cuenta.
	invitacionesPublicas := router.Group("/api/v1/invitaciones")
	invitacionesPublicas.GET("/:token", transporthttp.LimitarPorIP(60, time.Minute), invitacionHandler.Consultar)
	invitacionesPublicas.POST("/:token/registro", transporthttp.LimitarPorIP(10, time.Minute), invitacionHandler.Registrar)
	// Aceptar sí exige sesión verificada con el correo invitado.
	invitaciones := router.Group("/api/v1/invitaciones")
	invitaciones.Use(transporthttp.RequireAuth(sessionService, cfg.SessionCookieName()), transporthttp.RequireCSRF())
	invitaciones.POST("/:token/aceptar", invitacionHandler.Aceptar)

	contexto := router.Group("/api/v1/contexto")
	contexto.Use(transporthttp.RequireAuth(sessionService, cfg.SessionCookieName()))
	contexto.GET("/opciones", contextoHandler.Opciones)
	negocioActual := router.Group("/api/v1/negocios/:negocioId")
	negocioActual.Use(transporthttp.RequireAuth(sessionService, cfg.SessionCookieName()), transporthttp.RequireCSRF(), negociohttp.RequireNegocioActivo(contextoService))
	catalogoLectura := negocioActual.Group("")
	catalogoLectura.Use(negociohttp.RequierePermiso(rolService, negociodomain.PermisoCatalogoVer))
	catalogoLectura.GET("/catalogo/productos", productHandler.List)
	catalogoLectura.GET("/catalogo/productos/consulta-codigo/:codigoBarras", productHandler.LookupProduct)
	catalogoLectura.GET("/catalogo/productos/consulta-preciocheck/:codigoBarras", productHandler.LookupProduct)
	catalogoLectura.GET("/catalogo/categorias", productHandler.ListCategoriesAdmin)
	catalogoLectura.GET("/catalogo/marcas", productHandler.ListBrandsAdmin)
	catalogoLectura.GET("/catalogo/unidades-medida", productHandler.ListUnits)
	catalogoLectura.GET("/catalogo/proveedores", productHandler.ListProviders)
	catalogoLectura.GET("/sucursales", productHandler.Branches)
	comprasLectura := negocioActual.Group("")
	comprasLectura.Use(negociohttp.RequierePermiso(rolService, negociodomain.PermisoCompraVer))
	comprasLectura.GET("/compras", compraHandler.List)
	comprasLectura.GET("/compras/:compraId", compraHandler.Get)
	comprasLectura.GET("/compras/productos", compraHandler.Products)

	catalogoGestion := negocioActual.Group("")
	catalogoGestion.Use(negociohttp.RequierePermiso(rolService, negociodomain.PermisoCatalogoGestionar))
	catalogoGestion.POST("/catalogo/productos", productHandler.Create)
	comprasGestion := negocioActual.Group("")
	comprasGestion.Use(negociohttp.RequierePermiso(rolService, negociodomain.PermisoCompraRegistrar))
	comprasGestion.POST("/compras", compraHandler.Create)
	comprasGestion.POST("/compras/unidades", productHandler.CreateUnit)
	negocioActual.GET("/precios/propuestas", negociohttp.RequierePermiso(rolService, negociodomain.PermisoPrecioVer), precioHandler.ListPending)
	negocioActual.PATCH("/precios/propuestas/:propuestaId/autorizar", negociohttp.RequierePermiso(rolService, negociodomain.PermisoPrecioAutorizar), precioHandler.Authorize)
	catalogoGestion.PATCH("/catalogo/productos/:productoId", productHandler.Update)
	catalogoGestion.DELETE("/catalogo/productos/:productoId", productHandler.Deactivate)
	catalogoGestion.GET("/catalogo/productos/importacion/plantilla", productHandler.ProductImportTemplate)
	catalogoGestion.POST("/catalogo/productos/validar-importacion", productHandler.PreviewProductImport)
	catalogoGestion.POST("/catalogo/productos/importar", productHandler.ImportProducts)
	catalogoGestion.GET("/catalogo/productos/importaciones/:importacionId", productHandler.ProductImportStatus)
	catalogoGestion.POST("/catalogo/marcas", productHandler.CreateBrand)
	catalogoGestion.PATCH("/catalogo/marcas/:id", productHandler.UpdateBrand)
	catalogoGestion.GET("/catalogo/marcas/importacion/plantilla", productHandler.BrandImportTemplate)
	catalogoGestion.POST("/catalogo/marcas/importar", productHandler.ImportBrands)
	catalogoGestion.POST("/catalogo/categorias", productHandler.CreateCategory)
	catalogoGestion.PATCH("/catalogo/categorias/:id", productHandler.UpdateCategory)
	catalogoGestion.GET("/catalogo/categorias/importacion/plantilla", productHandler.CategoryImportTemplate)
	catalogoGestion.POST("/catalogo/categorias/importar", productHandler.ImportCategories)
	catalogoGestion.POST("/catalogo/unidades-medida", productHandler.CreateUnit)
	catalogoGestion.PATCH("/catalogo/unidades-medida/:id", productHandler.UpdateUnit)
	catalogoGestion.GET("/catalogo/unidades/importacion/plantilla", productHandler.UnitImportTemplate)
	catalogoGestion.POST("/catalogo/unidades/importar", productHandler.ImportUnits)
	catalogoGestion.POST("/catalogo/proveedores", productHandler.CreateProvider)
	catalogoGestion.PATCH("/catalogo/proveedores/:id", productHandler.UpdateProvider)
	catalogoGestion.GET("/catalogo/proveedores/importacion/plantilla", productHandler.ProviderImportTemplate)
	catalogoGestion.POST("/catalogo/proveedores/importar", productHandler.ImportProviders)
	log.Printf("API escuchando en http://localhost:%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}

// newMailer arma el servicio de correo: Mailpit en desarrollo y Amazon SES por SMTP en producción.
func newMailer(cfg config.Config) (*correoapplication.Servicio, error) {
	plantillas, err := correoinfra.NewRenderizador()
	if err != nil {
		return nil, err
	}
	logo, err := correoinfra.Logo()
	if err != nil {
		return nil, err
	}
	transporte, err := correoinfra.NewSMTPTransporte(correoinfra.SMTPConfig{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, TLS: cfg.SMTPTLS,
		Usuario: cfg.SMTPUsername, Contrasena: cfg.SMTPPassword, Timeout: cfg.SMTPTimeout,
		Remitente: cfg.MailFrom, NombreRemitente: cfg.MailFromName, ResponderA: cfg.MailReplyTo,
		ConfigurationSetSES: cfg.SESConfigurationSet,
	}, logo)
	if err != nil {
		return nil, err
	}
	return correoapplication.NewServicio(transporte, plantillas, correoapplication.Config{
		NombreApp: cfg.MailFromName, FrontendURL: cfg.FrontendURL,
	}), nil
}
