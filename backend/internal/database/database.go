package database

import (
	"tienda/backend/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(url string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(url), &gorm.Config{})
}

func Init(db *gorm.DB) error {
	models := []interface{}{
		&domain.Usuario{}, &domain.PerfilUsuario{}, &domain.Empleado{},
		&domain.Negocio{}, &domain.Rol{}, &domain.Membresia{},
		&domain.Sucursal{}, &domain.Marca{}, &domain.Producto{}, &domain.Categoria{},
		&domain.ProductoCategoria{}, &domain.ProductoCodigo{}, &domain.ProductoImagen{},
		&domain.ProductoUnidad{}, &domain.ProductoNegocio{}, &domain.Impuesto{},
		&domain.ProductoImpuesto{}, &domain.InventarioSucursal{}, &domain.Proveedor{},
		&domain.ProductoProveedor{}, &domain.Lote{}, &domain.MovimientoInventario{},
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

	// Compatibilidad con bases creadas antes de agregar el SKU interno al vínculo negocio-producto.
	// IF NOT EXISTS hace que esta migración sea segura al reiniciar la API.
	if err := db.Exec(`ALTER TABLE IF EXISTS producto_negocio ADD COLUMN IF NOT EXISTS sku_interno varchar(120)`).Error; err != nil {
		return err
	}
	return nil
}
