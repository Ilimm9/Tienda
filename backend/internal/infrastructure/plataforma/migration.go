package plataforma

import (
	domain "tienda/backend/internal/domain/plataforma"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&domain.RolPlataforma{},
		&domain.RolUsuarioPlataforma{},
		&domain.AuditoriaPlataforma{},
		&domain.ActividadUsuarioDiaria{},
	); err != nil {
		return err
	}

	descripcion := "Consulta global de negocios, sucursales, usuarios y métricas de actividad"
	return db.Where("codigo = ?", domain.RolAdminLectura).
		Attrs(domain.RolPlataforma{
			ID: uuid.New(), Codigo: domain.RolAdminLectura, Nombre: "Administrador de lectura",
			Descripcion: &descripcion, Activo: true,
		}).FirstOrCreate(&domain.RolPlataforma{}).Error
}
