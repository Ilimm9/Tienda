package cuenta

import "gorm.io/gorm"

// MigratePerfilApellidos separa los apellidos del perfil en primer y segundo apellido.
//
// Corre antes de AutoMigrate: una columna obligatoria no puede agregarse a una tabla con filas.
// No hay forma segura de partir un apellido compuesto, así que el valor completo pasa a
// `primer_apellido` y la persona lo corrige después. `apellidos` se conserva como respaldo,
// ya sin obligatoriedad, hasta retirarla en una migración posterior.
func MigratePerfilApellidos(db *gorm.DB) error {
	if !db.Migrator().HasTable("perfil_usuarios") {
		return nil
	}
	statements := []string{
		`ALTER TABLE perfil_usuarios ADD COLUMN IF NOT EXISTS primer_apellido varchar(120)`,
		`ALTER TABLE perfil_usuarios ADD COLUMN IF NOT EXISTS segundo_apellido varchar(120)`,
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'perfil_usuarios' AND column_name = 'apellidos') THEN
				UPDATE perfil_usuarios SET primer_apellido = btrim(apellidos) WHERE primer_apellido IS NULL;
				ALTER TABLE perfil_usuarios ALTER COLUMN apellidos DROP NOT NULL;
			END IF;
		END $$`,
		`UPDATE perfil_usuarios SET primer_apellido = '' WHERE primer_apellido IS NULL`,
		// El valor por omisión mantiene funcionando a una instancia anterior durante el despliegue.
		`ALTER TABLE perfil_usuarios ALTER COLUMN primer_apellido SET DEFAULT ''`,
		`ALTER TABLE perfil_usuarios ALTER COLUMN primer_apellido SET NOT NULL`,
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
