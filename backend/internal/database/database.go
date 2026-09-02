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
	}
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			if err := db.Migrator().CreateTable(model); err != nil {
				return err
			}
		}
	}
	return db.AutoMigrate(models...)
}
