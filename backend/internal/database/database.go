package database

import (
	"tienda/backend/internal/domain"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	DevelopmentBusinessID = "11111111-1111-4111-8111-111111111111"
	developmentBranchID   = "22222222-2222-4222-8222-222222222222"
)

func Open(url string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(url), &gorm.Config{})
}

func Init(db *gorm.DB) error {
	if err := negocioinfra.MigratePhaseOne(db); err != nil {
		return err
	}
	if err := negocioinfra.MigratePhaseTwo(db); err != nil {
		return err
	}

	models := []interface{}{
		&cuentadomain.Usuario{}, &cuentadomain.PerfilUsuario{}, &negociodomain.Empleado{},
		&negociodomain.Direccion{}, &negociodomain.Negocio{}, &negociodomain.ConfiguracionNegocio{},
		&negociodomain.Rol{}, &negociodomain.MembresiaNegocio{},
		&negociodomain.Sucursal{}, &domain.Marca{}, &domain.Producto{}, &domain.Categoria{},
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
	if err := negocioinfra.MigratePhaseTwo(db); err != nil {
		return err
	}

	// Compatibilidad con bases creadas antes de agregar el SKU interno al vínculo negocio-producto.
	// IF NOT EXISTS hace que esta migración sea segura al reiniciar la API.
	if err := db.Exec(`ALTER TABLE IF EXISTS producto_negocio ADD COLUMN IF NOT EXISTS sku_interno varchar(120)`).Error; err != nil {
		return err
	}
	return nil
}

// SeedDevelopment provides the minimum organization data needed to exercise the
// catalog and inventory flows locally. It must never run in production.
// TODO(sucursales): remove this seed when businesses and branches are created by
// their complete onboarding flow.
func SeedDevelopment(db *gorm.DB) error {
	businessID := uuid.MustParse(DevelopmentBusinessID)
	branchID := uuid.MustParse(developmentBranchID)

	return db.Transaction(func(tx *gorm.DB) error {
		var business negociodomain.Negocio
		if err := tx.Where("id = ?", businessID).Attrs(negociodomain.Negocio{
			ID: businessID, Slug: "negocio-prueba", NombreComercial: "Negocio de prueba",
			Nombre: "Negocio de prueba", CodigoMoneda: "MXN",
			ZonaHoraria: "America/Mexico_City", Estado: "activo",
		}).FirstOrCreate(&business).Error; err != nil {
			return err
		}

		var branch negociodomain.Sucursal
		return tx.Where("id = ?", branchID).Attrs(negociodomain.Sucursal{
			ID: branchID, NegocioID: businessID, Codigo: "SUC-001", Nombre: "Tienda prueba",
			EsPrincipal: true, Activo: true,
		}).FirstOrCreate(&branch).Error
	})
}
