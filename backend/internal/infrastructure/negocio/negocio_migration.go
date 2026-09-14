package negocio

import "gorm.io/gorm"

// migrateNegociosPhaseOne upgrades the legacy development schema without
// dropping data. Every statement is safe to run more than once.
func MigratePhaseOne(db *gorm.DB) error {
	businessStatements := []string{
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS slug varchar(120)`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS nombre_comercial varchar(180)`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS correo varchar(254)`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS direccion_id uuid`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS archivo_logo_id uuid`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS codigo_moneda varchar(3) DEFAULT 'MXN'`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS zona_horaria varchar(80) DEFAULT 'America/Mexico_City'`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS creado_por_usuario_id uuid`,
		`ALTER TABLE negocios ADD COLUMN IF NOT EXISTS archivado_en timestamptz`,
		`UPDATE negocios SET nombre_comercial = nombre WHERE nombre_comercial IS NULL OR btrim(nombre_comercial) = ''`,
		`UPDATE negocios SET correo = email WHERE correo IS NULL AND email IS NOT NULL`,
		`UPDATE negocios SET slug = 'negocio-' || substr(replace(id::text, '-', ''), 1, 8) WHERE slug IS NULL OR btrim(slug) = ''`,
		`UPDATE negocios SET codigo_moneda = 'MXN' WHERE codigo_moneda IS NULL OR btrim(codigo_moneda) = ''`,
		`UPDATE negocios SET zona_horaria = 'America/Mexico_City' WHERE zona_horaria IS NULL OR btrim(zona_horaria) = ''`,
		`ALTER TABLE negocios ALTER COLUMN slug SET NOT NULL`,
		`ALTER TABLE negocios ALTER COLUMN nombre_comercial SET NOT NULL`,
		`ALTER TABLE negocios ALTER COLUMN codigo_moneda SET NOT NULL`,
		`ALTER TABLE negocios ALTER COLUMN zona_horaria SET NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_negocios_slug ON negocios (lower(slug))`,
		`CREATE INDEX IF NOT EXISTS idx_negocios_creado_por_usuario_id ON negocios (creado_por_usuario_id)`,
	}
	membershipStatements := []string{
		`ALTER TABLE membresias_negocio ADD COLUMN IF NOT EXISTS tipo_miembro varchar(30) DEFAULT 'miembro'`,
		`ALTER TABLE membresias_negocio ADD COLUMN IF NOT EXISTS se_unio_en timestamptz`,
		`ALTER TABLE membresias_negocio ADD COLUMN IF NOT EXISTS suspendido_en timestamptz`,
		`ALTER TABLE membresias_negocio ADD COLUMN IF NOT EXISTS revocado_en timestamptz`,
		`UPDATE membresias_negocio SET tipo_miembro = 'miembro' WHERE tipo_miembro IS NULL OR btrim(tipo_miembro) = ''`,
		`UPDATE membresias_negocio SET se_unio_en = COALESCE(creado_en, now()) WHERE se_unio_en IS NULL`,
		`ALTER TABLE membresias_negocio ALTER COLUMN tipo_miembro SET NOT NULL`,
		// Desde fase 4 la columna ya no existe: base.MD no la declara y su dato vive en roles_membresia.
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_name = 'membresias_negocio' AND column_name = 'rol_id') THEN
				ALTER TABLE membresias_negocio ALTER COLUMN rol_id DROP NOT NULL;
			END IF;
		END $$`,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if tx.Migrator().HasTable("membresia") && !tx.Migrator().HasTable("membresias_negocio") {
			if err := tx.Migrator().RenameTable("membresia", "membresias_negocio"); err != nil {
				return err
			}
		}
		if tx.Migrator().HasTable("negocios") {
			for _, statement := range businessStatements {
				if err := tx.Exec(statement).Error; err != nil {
					return err
				}
			}
		}
		if tx.Migrator().HasTable("membresias_negocio") {
			for _, statement := range membershipStatements {
				if err := tx.Exec(statement).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
