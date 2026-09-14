package negocio

import "gorm.io/gorm"

// MigratePhaseFive lleva `empleados` a la forma de base.MD e invierte el vínculo con la membresía.
//
// La tabla previa era andamiaje sin consumidores: `perfil_id`, `numero` único global y sin negocio.
// La migración es defensiva por si alguna base ya tiene filas: rellena antes de restringir y nunca borra.
func MigratePhaseFive(db *gorm.DB) error {
	if !db.Migrator().HasTable("empleados") {
		return nil
	}

	statements := []string{
		// 1. Columnas de base.MD.
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS membresia_id uuid`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS nombre varchar(100)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS segundo_nombre varchar(100)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS primer_apellido varchar(100)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS segundo_apellido varchar(100)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS correo varchar(254)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS telefono varchar(30)`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS contratado_en date`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS terminado_en date`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS creado_por_usuario_id uuid`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS creado_en timestamptz NOT NULL DEFAULT now()`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS actualizado_en timestamptz NOT NULL DEFAULT now()`,

		// 2. `numero` global pasa a `numero_empleado` único por negocio.
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_name = 'empleados' AND column_name = 'numero')
			AND NOT EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_name = 'empleados' AND column_name = 'numero_empleado') THEN
				ALTER TABLE empleados RENAME COLUMN numero TO numero_empleado;
			END IF;
		END $$`,
		`ALTER TABLE empleados ADD COLUMN IF NOT EXISTS numero_empleado varchar(40)`,
		`ALTER TABLE empleados ALTER COLUMN numero_empleado TYPE varchar(40)`,
		`ALTER TABLE empleados ALTER COLUMN numero_empleado DROP NOT NULL`,
		`DROP INDEX IF EXISTS idx_empleados_numero`,
		`DROP INDEX IF EXISTS uni_empleados_numero`,

		// 3. Derivar identidad de filas preexistentes desde el perfil, cuando pueda resolverse.
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_name = 'empleados' AND column_name = 'perfil_id')
			AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'perfiles_usuario') THEN
				UPDATE empleados e SET
					nombre = COALESCE(NULLIF(btrim(e.nombre), ''), NULLIF(btrim(p.nombre), ''), 'Sin nombre'),
					primer_apellido = COALESCE(NULLIF(btrim(e.primer_apellido), ''), NULLIF(btrim(p.primer_apellido), ''), 'Sin apellido'),
					segundo_nombre = COALESCE(e.segundo_nombre, p.segundo_nombre),
					segundo_apellido = COALESCE(e.segundo_apellido, p.segundo_apellido),
					telefono = COALESCE(e.telefono, p.telefono)
				FROM perfiles_usuario p
				WHERE p.id = e.perfil_id;
			END IF;
		END $$`,
		// 4. El vínculo correcto es empleados.membresia_id, no membresias_negocio.empleado_id.
		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
				WHERE table_name = 'membresias_negocio' AND column_name = 'empleado_id') THEN
				UPDATE empleados e SET membresia_id = m.id
				FROM membresias_negocio m
				WHERE m.empleado_id = e.id AND e.membresia_id IS NULL;
				UPDATE empleados e SET negocio_id = m.negocio_id
				FROM membresias_negocio m
				WHERE m.id = e.membresia_id AND e.negocio_id IS NULL;
			END IF;
		END $$`,
		`ALTER TABLE membresias_negocio DROP COLUMN IF EXISTS empleado_id`,
		`ALTER TABLE empleados DROP COLUMN IF EXISTS perfil_id`,

		// 5. Restricciones finales: solo si todas las filas quedaron resueltas.
		`UPDATE empleados SET estado = 'pendiente' WHERE estado IS NULL OR btrim(estado) = ''`,
		`ALTER TABLE empleados ALTER COLUMN estado SET DEFAULT 'pendiente'`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM empleados WHERE negocio_id IS NULL) THEN
				ALTER TABLE empleados ALTER COLUMN negocio_id SET NOT NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM empleados WHERE nombre IS NULL OR btrim(nombre) = '') THEN
				ALTER TABLE empleados ALTER COLUMN nombre SET NOT NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM empleados WHERE primer_apellido IS NULL OR btrim(primer_apellido) = '') THEN
				ALTER TABLE empleados ALTER COLUMN primer_apellido SET NOT NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM empleados WHERE creado_por_usuario_id IS NULL) THEN
				ALTER TABLE empleados ALTER COLUMN creado_por_usuario_id SET NOT NULL;
			END IF;
		END $$`,

		// 6. Índices y claves de base.MD.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empleados_negocio_numero ON empleados (negocio_id, numero_empleado)
			WHERE numero_empleado IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_empleados_negocio_correo ON empleados (negocio_id, correo)`,
		`CREATE INDEX IF NOT EXISTS idx_empleados_negocio_estado ON empleados (negocio_id, estado)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empleados_membresia ON empleados (membresia_id)
			WHERE membresia_id IS NOT NULL`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_empleados_negocio')
			AND NOT EXISTS (SELECT 1 FROM empleados WHERE negocio_id IS NULL) THEN
				ALTER TABLE empleados ADD CONSTRAINT fk_empleados_negocio
				FOREIGN KEY (negocio_id) REFERENCES negocios(id);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_empleados_membresia') THEN
				ALTER TABLE empleados ADD CONSTRAINT fk_empleados_membresia
				FOREIGN KEY (membresia_id) REFERENCES membresias_negocio(id);
			END IF;
		END $$`,
	}

	return ejecutar(db, statements)
}
