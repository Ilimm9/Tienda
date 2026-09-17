package negocio

import (
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MigratePhaseFour lleva roles y membresías a la forma de base.MD y siembra el catálogo de permisos.
//
// La tabla `roles` previa era andamiaje sin consumidores, por lo que esta migración converge su forma
// en lugar de conservar un puente legacy. Aun así respeta filas existentes: rellena antes de restringir.
func MigratePhaseFour(db *gorm.DB) error {
	if !db.Migrator().HasTable("roles") {
		return nil
	}

	statements := []string{
		// 1. Ampliar `roles` hasta la forma de base.MD. Ampliar varchar no es destructivo.
		`ALTER TABLE roles ALTER COLUMN nombre TYPE varchar(120)`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS codigo varchar(60)`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS es_rol_sistema boolean NOT NULL DEFAULT false`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS activo boolean NOT NULL DEFAULT true`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS creado_por_usuario_id uuid`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS creado_en timestamptz NOT NULL DEFAULT now()`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS actualizado_en timestamptz NOT NULL DEFAULT now()`,

		// 2. Backfill determinista de `codigo` desde `nombre`, resolviendo colisiones por negocio.
		`DO $$
		DECLARE
			fila record;
			candidato varchar(60);
			base_codigo varchar(60);
			contador integer;
		BEGIN
			FOR fila IN
				SELECT id, negocio_id, nombre FROM roles
				WHERE codigo IS NULL OR btrim(codigo) = ''
				ORDER BY negocio_id, creado_en NULLS LAST, id
			LOOP
				base_codigo := upper(regexp_replace(translate(COALESCE(fila.nombre, 'ROL'),
					'áéíóúÁÉÍÓÚñÑüÜ', 'aeiouAEIOUnNuU'), '[^A-Za-z0-9]+', '_', 'g'));
				base_codigo := btrim(base_codigo, '_');
				IF base_codigo IS NULL OR base_codigo = '' THEN
					base_codigo := 'ROL';
				END IF;
				base_codigo := left(base_codigo, 55);
				candidato := base_codigo;
				contador := 1;
				WHILE EXISTS (
					SELECT 1 FROM roles
					WHERE negocio_id = fila.negocio_id AND lower(codigo) = lower(candidato)
				) LOOP
					contador := contador + 1;
					candidato := base_codigo || '_' || contador::text;
				END LOOP;
				UPDATE roles SET codigo = candidato WHERE id = fila.id;
			END LOOP;
		END $$`,
		`ALTER TABLE roles ALTER COLUMN codigo SET NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_negocio_codigo_ci ON roles (negocio_id, lower(codigo))`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_roles_negocio') THEN
				ALTER TABLE roles ADD CONSTRAINT fk_roles_negocio
				FOREIGN KEY (negocio_id) REFERENCES negocios(id);
			END IF;
		END $$`,

		// 3. Tablas nuevas del modelo de base.MD.
		`CREATE TABLE IF NOT EXISTS permisos (
			id uuid PRIMARY KEY,
			codigo varchar(120) NOT NULL,
			codigo_modulo varchar(30) NOT NULL,
			nombre varchar(160) NOT NULL,
			descripcion text,
			creado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_permisos_codigo ON permisos (codigo)`,
		`CREATE INDEX IF NOT EXISTS idx_permisos_modulo ON permisos (codigo_modulo)`,
		`CREATE TABLE IF NOT EXISTS permisos_rol (
			id uuid PRIMARY KEY,
			rol_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			permiso_id uuid NOT NULL REFERENCES permisos(id),
			creado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_permisos_rol_unico ON permisos_rol (rol_id, permiso_id)`,
		`CREATE TABLE IF NOT EXISTS roles_membresia (
			id uuid PRIMARY KEY,
			membresia_negocio_id uuid NOT NULL REFERENCES membresias_negocio(id) ON DELETE CASCADE,
			rol_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			asignado_por_usuario_id uuid,
			asignado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_membresia_unico ON roles_membresia (membresia_negocio_id, rol_id)`,
	}

	if err := ejecutar(db, statements); err != nil {
		return err
	}
	if err := sembrarPermisos(db); err != nil {
		return err
	}
	return sembrarRolPropietario(db)
}

// sembrarPermisos inserta el catálogo global por código sin borrar códigos desconocidos ya presentes.
func sembrarPermisos(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, permiso := range domain.CatalogoPermisos {
			err := tx.Exec(`INSERT INTO permisos (id, codigo, codigo_modulo, nombre, creado_en)
				VALUES (gen_random_uuid(), ?, ?, ?, now())
				ON CONFLICT (codigo) DO UPDATE SET codigo_modulo = EXCLUDED.codigo_modulo, nombre = EXCLUDED.nombre`,
				permiso.Codigo, permiso.CodigoModulo, permiso.Nombre).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// sembrarRolPropietario garantiza el rol de sistema con todos los permisos y lo asigna a los propietarios.
func sembrarRolPropietario(db *gorm.DB) error {
	statements := []string{
		// Un rol de sistema PROPIETARIO por negocio.
		`INSERT INTO roles (id, negocio_id, codigo, nombre, descripcion, es_rol_sistema, activo, creado_en, actualizado_en)
		SELECT gen_random_uuid(), n.id, '` + domain.CodigoRolPropietario + `', 'Propietario',
			'Rol de sistema con todos los permisos del negocio.', true, true, now(), now()
		FROM negocios n
		WHERE NOT EXISTS (
			SELECT 1 FROM roles r WHERE r.negocio_id = n.id AND lower(r.codigo) = lower('` + domain.CodigoRolPropietario + `')
		)`,
		// Marcarlo como rol de sistema aunque existiera de antes con ese código.
		`UPDATE roles SET es_rol_sistema = true, activo = true
		WHERE lower(codigo) = lower('` + domain.CodigoRolPropietario + `') AND es_rol_sistema = false`,
		// Todos los permisos vigentes pertenecen al rol de sistema.
		`INSERT INTO permisos_rol (id, rol_id, permiso_id, creado_en)
		SELECT gen_random_uuid(), r.id, p.id, now()
		FROM roles r CROSS JOIN permisos p
		WHERE r.es_rol_sistema = true
		ON CONFLICT (rol_id, permiso_id) DO NOTHING`,
		// Backfill del rol único legacy hacia la relación muchos-a-muchos.
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'membresias_negocio' AND column_name = 'rol_id'
			) THEN
				INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_en)
				SELECT gen_random_uuid(), m.id, m.rol_id, now()
				FROM membresias_negocio m
				WHERE m.rol_id IS NOT NULL
					AND EXISTS (SELECT 1 FROM roles r WHERE r.id = m.rol_id)
				ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING;
			END IF;
		END $$`,
		// Toda membresía propietaria activa conserva el rol de sistema.
		`INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_en)
		SELECT gen_random_uuid(), m.id, r.id, now()
		FROM membresias_negocio m
		JOIN roles r ON r.negocio_id = m.negocio_id AND r.es_rol_sistema = true
		WHERE m.tipo_miembro = 'propietario' AND m.estado = 'activo'
		ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING`,
		// Columnas que base.MD no declara: se retiran una vez backfilleadas.
		`ALTER TABLE membresias_negocio DROP COLUMN IF EXISTS rol_id`,
	}
	return ejecutar(db, statements)
}

// sembrarRolPropietarioDeNegocio crea el rol de sistema de un negocio y se lo asigna a su propietario.
// Se usa tanto al migrar negocios existentes como al crear uno nuevo en runtime.
func sembrarRolPropietarioDeNegocio(tx *gorm.DB, negocioID, membresiaID, creadoPor uuid.UUID) error {
	err := tx.Exec(`INSERT INTO roles (id, negocio_id, codigo, nombre, descripcion, es_rol_sistema, activo, creado_por_usuario_id, creado_en, actualizado_en)
		SELECT gen_random_uuid(), ?, ?, 'Propietario',
			'Rol de sistema con todos los permisos del negocio.', true, true, ?, now(), now()
		WHERE NOT EXISTS (
			SELECT 1 FROM roles r WHERE r.negocio_id = ? AND lower(r.codigo) = lower(?)
		)`, negocioID, domain.CodigoRolPropietario, creadoPor, negocioID, domain.CodigoRolPropietario).Error
	if err != nil {
		return err
	}
	err = tx.Exec(`INSERT INTO permisos_rol (id, rol_id, permiso_id, creado_en)
		SELECT gen_random_uuid(), r.id, p.id, now()
		FROM roles r CROSS JOIN permisos p
		WHERE r.negocio_id = ? AND r.es_rol_sistema = true
		ON CONFLICT (rol_id, permiso_id) DO NOTHING`, negocioID).Error
	if err != nil {
		return err
	}
	return tx.Exec(`INSERT INTO roles_membresia (id, membresia_negocio_id, rol_id, asignado_por_usuario_id, asignado_en)
		SELECT gen_random_uuid(), ?, r.id, ?, now()
		FROM roles r WHERE r.negocio_id = ? AND r.es_rol_sistema = true
		ON CONFLICT (membresia_negocio_id, rol_id) DO NOTHING`, membresiaID, creadoPor, negocioID).Error
}

func ejecutar(db *gorm.DB, statements []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
