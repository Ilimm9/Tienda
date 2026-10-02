package negocio

import (
	"log"

	"gorm.io/gorm"
)

// MigrateInvitacionSucursal liga cada invitación nueva a una sucursal del mismo negocio.
//
// Es aditiva e idempotente. Las pendientes anteriores sin sucursal se cancelan y siguen visibles:
// el administrador emite un enlace nuevo eligiendo sucursal. Las aceptadas conservan sucursal nula.
func MigrateInvitacionSucursal(db *gorm.DB) error {
	if !db.Migrator().HasTable("invitaciones_negocio") || !db.Migrator().HasTable("sucursales") {
		return nil
	}
	if err := db.Exec(`ALTER TABLE invitaciones_negocio ADD COLUMN IF NOT EXISTS sucursal_id uuid`).Error; err != nil {
		return err
	}
	resultado := db.Exec(`UPDATE invitaciones_negocio SET estado = 'cancelada'
		WHERE estado = 'pendiente' AND sucursal_id IS NULL`)
	if resultado.Error != nil {
		return resultado.Error
	}
	if resultado.RowsAffected > 0 {
		log.Printf("invitaciones: %d pendientes sin sucursal fueron canceladas; requieren un enlace nuevo", resultado.RowsAffected)
	}
	return ejecutar(db, []string{
		// Conserva solo la pendiente más reciente por empleado antes de exigir unicidad.
		`UPDATE invitaciones_negocio i SET estado = 'cancelada'
			WHERE i.estado = 'pendiente' AND i.empleado_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM invitaciones_negocio o
				WHERE o.negocio_id = i.negocio_id AND o.empleado_id = i.empleado_id AND o.estado = 'pendiente'
					AND (o.creado_en > i.creado_en OR (o.creado_en = i.creado_en AND o.id > i.id)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_sucursales_negocio_sucursal ON sucursales (negocio_id, id)`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_invitacion_sucursal_negocio') THEN
				ALTER TABLE invitaciones_negocio ADD CONSTRAINT fk_invitacion_sucursal_negocio
					FOREIGN KEY (negocio_id, sucursal_id) REFERENCES sucursales (negocio_id, id);
			END IF;
		END $$`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_invitacion_pendiente_empleado
			ON invitaciones_negocio (negocio_id, empleado_id)
			WHERE estado = 'pendiente' AND empleado_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_invitacion_negocio_sucursal_estado
			ON invitaciones_negocio (negocio_id, sucursal_id, estado)`,
	})
}
