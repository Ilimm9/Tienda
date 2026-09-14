package negocio

import "gorm.io/gorm"

// MigratePhaseSeven agrega las restricciones que GORM no puede declarar desde los tags del struct.
//
// El único parcial es la última defensa de concurrencia contra dos sucursales principales activas
// para el mismo empleado, igual que el de `sucursales` en fase 2.
func MigratePhaseSeven(db *gorm.DB) error {
	if !db.Migrator().HasTable("asignaciones_empleado_sucursal") {
		return nil
	}
	return ejecutar(db, []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_asignacion_principal_activa
			ON asignaciones_empleado_sucursal (empleado_id)
			WHERE es_principal = true AND activo = true`,
		`CREATE INDEX IF NOT EXISTS idx_asignacion_negocio_sucursal
			ON asignaciones_empleado_sucursal (negocio_id, sucursal_id)`,
	})
}
