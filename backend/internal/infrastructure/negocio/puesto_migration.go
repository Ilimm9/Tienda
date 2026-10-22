package negocio

import "gorm.io/gorm"

// MigratePuestosNegocio crea el índice único por nombre y la llave de `empleados.puesto_id`; las tablas las crea AutoMigrate.
// El puesto como texto libre se retira: aún no hay datos en producción que conservar.
func MigratePuestosNegocio(db *gorm.DB) error {
	for _, statement := range []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + indicePuestoNombre + ` ON puestos_negocio (negocio_id, lower(nombre))`,
		`ALTER TABLE empleados DROP COLUMN IF EXISTS puesto`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_empleados_puesto') THEN
				ALTER TABLE empleados ADD CONSTRAINT fk_empleados_puesto
					FOREIGN KEY (puesto_id) REFERENCES puestos_negocio(id) ON DELETE RESTRICT;
			END IF;
		END $$`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
