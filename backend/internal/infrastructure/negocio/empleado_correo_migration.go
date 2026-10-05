package negocio

import (
	"log"

	"gorm.io/gorm"
)

const indiceCorreoEmpleado = "uq_empleados_negocio_correo"

// MigrateEmpleadoCorreoUnico impide repetir un correo entre empleados del mismo negocio.
//
// Si ya existen duplicados no crea el índice: los reporta para que un administrador los
// resuelva, y lo intenta de nuevo en el siguiente arranque. La validación de aplicación
// sigue rechazando duplicados nuevos mientras tanto.
func MigrateEmpleadoCorreoUnico(db *gorm.DB) error {
	if !db.Migrator().HasTable("empleados") {
		return nil
	}
	var duplicados int64
	if err := db.Raw(`SELECT count(*) FROM (
			SELECT 1 FROM empleados WHERE correo IS NOT NULL
			GROUP BY negocio_id, lower(correo) HAVING count(*) > 1
		) repetidos`).Scan(&duplicados).Error; err != nil {
		return err
	}
	if duplicados > 0 {
		log.Printf("evento=migracion_omitida indice=%s correos_duplicados=%d accion=corregir_empleados", indiceCorreoEmpleado, duplicados)
		return nil
	}
	return db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ` + indiceCorreoEmpleado + `
		ON empleados (negocio_id, lower(correo)) WHERE correo IS NOT NULL`).Error
}
