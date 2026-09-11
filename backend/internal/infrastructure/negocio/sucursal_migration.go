package negocio

import "gorm.io/gorm"

// MigratePhaseTwo upgrades sucursales without removing legacy columns or data.
// It runs before and after AutoMigrate so both existing and fresh databases work.
func MigratePhaseTwo(db *gorm.DB) error {
	if !db.Migrator().HasTable("sucursales") {
		return nil
	}

	statements := []string{
		`ALTER TABLE sucursales ADD COLUMN IF NOT EXISTS codigo varchar(40)`,
		`ALTER TABLE sucursales ADD COLUMN IF NOT EXISTS direccion_id uuid`,
		`ALTER TABLE sucursales ADD COLUMN IF NOT EXISTS es_principal boolean NOT NULL DEFAULT false`,
		`ALTER TABLE sucursales ADD COLUMN IF NOT EXISTS eliminado_en timestamptz`,
		`DO $$
		DECLARE
			negocio_actual uuid;
			sucursal_actual uuid;
			contador integer;
			candidato varchar(40);
		BEGIN
			FOR negocio_actual IN SELECT DISTINCT negocio_id FROM sucursales ORDER BY negocio_id LOOP
				contador := 1;
				FOR sucursal_actual IN
					SELECT id FROM sucursales
					WHERE negocio_id = negocio_actual AND (codigo IS NULL OR btrim(codigo) = '')
					ORDER BY creado_en NULLS LAST, id
				LOOP
					candidato := 'SUC-' || lpad(contador::text, 3, '0');
					WHILE EXISTS (
						SELECT 1 FROM sucursales
						WHERE negocio_id = negocio_actual AND lower(codigo) = lower(candidato)
					) LOOP
						contador := contador + 1;
						candidato := 'SUC-' || lpad(contador::text, 3, '0');
					END LOOP;
					UPDATE sucursales SET codigo = candidato WHERE id = sucursal_actual;
					contador := contador + 1;
				END LOOP;
			END LOOP;
		END $$`,
		`DO $$
		DECLARE
			fila record;
			nueva_direccion_id uuid;
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'sucursales' AND column_name = 'direccion'
			) THEN
				FOR fila IN EXECUTE $query$
					SELECT id, direccion FROM sucursales
					WHERE direccion_id IS NULL AND direccion IS NOT NULL AND btrim(direccion) <> ''
					ORDER BY id
				$query$ LOOP
					nueva_direccion_id := gen_random_uuid();
					INSERT INTO direcciones (id, codigo_pais, referencias, creado_en, actualizado_en)
					VALUES (nueva_direccion_id, 'MX', fila.direccion, now(), now());
					UPDATE sucursales SET direccion_id = nueva_direccion_id
					WHERE id = fila.id AND direccion_id IS NULL;
				END LOOP;
			END IF;
		END $$`,
		`UPDATE sucursales SET eliminado_en = COALESCE(eliminado_en, actualizado_en, now()) WHERE activo = false`,
		`UPDATE sucursales SET es_principal = false WHERE activo = false AND es_principal = true`,
		`WITH duplicadas AS (
			SELECT id, row_number() OVER (PARTITION BY negocio_id ORDER BY creado_en NULLS LAST, id) AS posicion
			FROM sucursales WHERE activo = true AND es_principal = true
		)
		UPDATE sucursales SET es_principal = false
		WHERE id IN (SELECT id FROM duplicadas WHERE posicion > 1)`,
		`WITH candidatas AS (
			SELECT id, row_number() OVER (PARTITION BY negocio_id ORDER BY creado_en NULLS LAST, id) AS posicion
			FROM sucursales s
			WHERE activo = true AND NOT EXISTS (
				SELECT 1 FROM sucursales principal
				WHERE principal.negocio_id = s.negocio_id AND principal.activo = true AND principal.es_principal = true
			)
		)
		UPDATE sucursales SET es_principal = true
		WHERE id IN (SELECT id FROM candidatas WHERE posicion = 1)`,
		`ALTER TABLE sucursales ALTER COLUMN codigo SET NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sucursales_negocio_codigo_ci ON sucursales (negocio_id, lower(codigo))`,
		`CREATE INDEX IF NOT EXISTS idx_sucursales_negocio_nombre ON sucursales (negocio_id, nombre)`,
		`CREATE INDEX IF NOT EXISTS idx_sucursales_negocio_activo ON sucursales (negocio_id, activo)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sucursales_principal_activa ON sucursales (negocio_id) WHERE es_principal = true AND activo = true`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_sucursales_negocio') THEN
				ALTER TABLE sucursales ADD CONSTRAINT fk_sucursales_negocio
				FOREIGN KEY (negocio_id) REFERENCES negocios(id);
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_sucursales_direccion') THEN
				ALTER TABLE sucursales ADD CONSTRAINT fk_sucursales_direccion
				FOREIGN KEY (direccion_id) REFERENCES direcciones(id);
			END IF;
		END $$`,
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
