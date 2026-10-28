package database

import (
	"time"

	"tienda/backend/internal/domain"
	compradomain "tienda/backend/internal/domain/compra"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	plataformadomain "tienda/backend/internal/domain/plataforma"
	preciodomain "tienda/backend/internal/domain/precio"
	ventadomain "tienda/backend/internal/domain/venta"
	"tienda/backend/internal/infrastructure"
	cuentainfra "tienda/backend/internal/infrastructure/cuenta"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"
	plataformainfra "tienda/backend/internal/infrastructure/plataforma"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(url string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(url), &gorm.Config{NowFunc: nowUTC})
}

func nowUTC() time.Time { return time.Now().UTC() }

func Init(db *gorm.DB) error {
	if err := infrastructure.MigrateCatalogTenancy(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseOne(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseTwo(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseFour(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseFive(db); err != nil {
		return err
	}
	if err := cuentainfra.MigratePerfilApellidos(db); err != nil {
		return err
	}

	models := []interface{}{
		&cuentadomain.Usuario{}, &cuentadomain.PerfilUsuario{}, &cuentadomain.SesionUsuario{}, &cuentadomain.DesafioAutenticacion{},
		&negociodomain.Direccion{}, &negociodomain.Negocio{}, &negociodomain.ConfiguracionNegocio{},
		&negociodomain.Rol{}, &negociodomain.MembresiaNegocio{},
		&negociodomain.Permiso{}, &negociodomain.PermisoRol{}, &negociodomain.RolMembresia{},
		&negociodomain.Empleado{}, &negociodomain.InvitacionNegocio{},
		&negociodomain.AsignacionEmpleadoSucursal{}, &negociodomain.PuestoNegocio{},
		&negociodomain.Sucursal{}, &domain.Marca{}, &domain.UnidadMedida{}, &domain.Producto{}, &domain.ImportacionProducto{}, &domain.FamiliaProducto{}, &domain.FamiliaProductoConsecutivo{}, &domain.ProductoVariante{}, &domain.ProductoVarianteAtributo{}, &domain.Categoria{},
		&domain.ProductoCategoria{}, &domain.ProductoCodigo{}, &domain.ProductoImagen{},
		&domain.ProductoUnidad{}, &domain.ProductoNegocio{}, &domain.ProductoSKUConsecutivo{}, &domain.Impuesto{},
		&domain.ProductoImpuesto{}, &domain.InventarioSucursal{}, &domain.Proveedor{},
		&domain.ProductoProveedor{}, &domain.Lote{}, &domain.MovimientoInventario{},
		&compradomain.CompraProveedor{}, &compradomain.CompraProveedorDetalle{},
		&preciodomain.Propuesta{}, &preciodomain.PrecioSucursal{}, &preciodomain.HistorialCostoPrecioSucursal{}, &preciodomain.ConfiguracionSucursal{},
		&ventadomain.Venta{}, &ventadomain.VentaLinea{}, &ventadomain.VentaFolioConsecutivo{},
		&plataformadomain.RolPlataforma{}, &plataformadomain.RolUsuarioPlataforma{},
		&plataformadomain.AuditoriaPlataforma{}, &plataformadomain.ActividadUsuarioDiaria{},
	}
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			if err := db.Migrator().CreateTable(model); err != nil {
				return err
			}
		}
	}
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}
	// Los importes vigentes pertenecen a precios_sucursal. borrar en caso de bd. creada, eliminar posterior
	for _, statement := range []string{
		`ALTER TABLE IF EXISTS producto_negocio DROP COLUMN IF EXISTS precio_venta`,
		`ALTER TABLE IF EXISTS producto_negocio DROP COLUMN IF EXISTS costo_referencia`,
		`ALTER TABLE IF EXISTS producto_negocio DROP COLUMN IF EXISTS precio_mayoreo`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	if err := cuentainfra.MigrateSessionBaseline(db); err != nil {
		return err
	}
	if err := plataformainfra.Migrate(db); err != nil {
		return err
	}
	if err := infrastructure.MigrateCatalogTenancy(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseTwo(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseFour(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseFive(db); err != nil {
		return err
	}
	if err := negocioinfra.MigrateRolCajeroVentas(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseSeven(db); err != nil {
		return err
	}
	if err := negocioinfra.MigrateInvitacionSucursal(db); err != nil {
		return err
	}
	if err := negocioinfra.MigrateEmpleadoCorreoUnico(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePuestosNegocio(db); err != nil {
		return err
	}

	// Compatibilidad con bases creadas antes de agregar el SKU interno al vínculo negocio-producto.
	// IF NOT EXISTS hace que esta migración sea segura al reiniciar la API.
	if err := db.Exec(`ALTER TABLE IF EXISTS producto_negocio ADD COLUMN IF NOT EXISTS sku_interno varchar(120)`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE IF EXISTS productos ADD COLUMN IF NOT EXISTS unidad_medida_id uuid`).Error; err != nil {
		return err
	}
	// Una venta pertenece a un producto simple o a una variante, nunca a ambos.
	// La aplicación se inicializará sobre una base nueva; los DROP NOT NULL hacen
	// explícita la nulabilidad requerida por los modelos antes de instalar las
	// restricciones de integridad.
	for _, statement := range []string{
		`ALTER TABLE producto_negocio ALTER COLUMN producto_id DROP NOT NULL`,
		`ALTER TABLE producto_codigos ALTER COLUMN producto_id DROP NOT NULL`,
		`ALTER TABLE producto_negocio DROP CONSTRAINT IF EXISTS chk_producto_negocio_propietario`,
		`ALTER TABLE producto_negocio ADD CONSTRAINT chk_producto_negocio_propietario CHECK ((producto_id IS NOT NULL AND producto_variante_id IS NULL) OR (producto_id IS NULL AND producto_variante_id IS NOT NULL))`,
		`ALTER TABLE producto_codigos DROP CONSTRAINT IF EXISTS chk_producto_codigo_propietario`,
		`ALTER TABLE producto_codigos ADD CONSTRAINT chk_producto_codigo_propietario CHECK ((producto_id IS NOT NULL AND producto_variante_id IS NULL) OR (producto_id IS NULL AND producto_variante_id IS NOT NULL))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_producto_negocio_variante_unica ON producto_negocio (negocio_id, producto_variante_id) WHERE producto_variante_id IS NOT NULL`,
		// Ventas (capability 37): importes en centavos no negativos, cantidades positivas y estados conocidos.
		`ALTER TABLE ventas DROP CONSTRAINT IF EXISTS chk_ventas_importes`,
		`ALTER TABLE ventas ADD CONSTRAINT chk_ventas_importes CHECK (total_centavos >= 0 AND subtotal_centavos >= 0 AND iva_centavos >= 0 AND recibido_centavos >= total_centavos AND cambio_centavos = recibido_centavos - total_centavos)`,
		`ALTER TABLE ventas DROP CONSTRAINT IF EXISTS chk_ventas_estado`,
		`ALTER TABLE ventas ADD CONSTRAINT chk_ventas_estado CHECK (estado IN ('PAGADA', 'CANCELADA') AND forma_pago IN ('EFECTIVO', 'TARJETA', 'TRANSFERENCIA'))`,
		`ALTER TABLE venta_lineas DROP CONSTRAINT IF EXISTS chk_venta_lineas_cantidad`,
		`ALTER TABLE venta_lineas ADD CONSTRAINT chk_venta_lineas_cantidad CHECK (cantidad > 0 AND precio_unitario_centavos >= 0 AND importe_centavos = precio_unitario_centavos * cantidad)`,
		`CREATE INDEX IF NOT EXISTS idx_ventas_negocio_creado ON ventas (negocio_id, creado_en DESC)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
