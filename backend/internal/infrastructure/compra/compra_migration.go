package compra

import (
	compradomain "tienda/backend/internal/domain/compra"

	"gorm.io/gorm"
)

// Migrate siembra de forma idempotente las tasas de IEPS e IVA en la tabla existente `impuestos`.
// Solo inserta lo que falta; nunca borra ni cambia renglones existentes.
func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, tasa := range compradomain.CatalogoTasas {
			err := tx.Exec(`INSERT INTO impuestos (id, nombre, tipo, tasa, activo, creado_en, actualizado_en)
				SELECT gen_random_uuid(), ?, ?, ?, true, now(), now()
				WHERE NOT EXISTS (SELECT 1 FROM impuestos WHERE tipo = ? AND tasa = ?)`,
				tasa.Nombre, tasa.Tipo, tasa.Tasa, tasa.Tipo, tasa.Tasa).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
