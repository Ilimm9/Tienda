package cuenta

import "gorm.io/gorm"

// MigrateSessionBaseline marca una sola vez las cuentas activas preexistentes como
// verificadas. Antes de la fase OTP el sistema no exigía correo verificado y bloquear
// esas cuentas durante el cambio de tipo de sesión sería una regresión.
func MigrateSessionBaseline(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE TABLE IF NOT EXISTS migraciones_aplicacion (
			clave varchar(160) PRIMARY KEY,
			ejecutada_en timestamptz NOT NULL DEFAULT now()
		)`).Error; err != nil {
			return err
		}
		result := tx.Exec(`INSERT INTO migraciones_aplicacion (clave)
			VALUES ('fase_2_correo_legacy_verificado') ON CONFLICT DO NOTHING`)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Exec(`UPDATE usuarios
			SET correo_verificado_en = COALESCE(correo_verificado_en, creado_en, now())
			WHERE estado = 'activo' AND correo_verificado_en IS NULL`).Error
	})
}
