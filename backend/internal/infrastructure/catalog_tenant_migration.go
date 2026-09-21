package infrastructure

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateCatalogTenancy convierte el catálogo legacy global en datos propiedad de un negocio.
//
// La migración se niega a adivinar cuando un producto está compartido o cuando existen filas
// huérfanas y más de un negocio. De ese modo un despliegue nunca asigna datos al tenant incorrecto.
func MigrateCatalogTenancy(db *gorm.DB) error {
	for _, table := range []string{"negocios", "marcas", "categorias", "unidades_medida", "productos", "producto_negocio", "producto_categorias", "producto_codigos"} {
		if !db.Migrator().HasTable(table) {
			return nil
		}
	}

	statements := []string{
		`ALTER TABLE marcas ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE categorias ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE unidades_medida ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE productos ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE producto_categorias ADD COLUMN IF NOT EXISTS negocio_id uuid`,
		`ALTER TABLE producto_codigos ADD COLUMN IF NOT EXISTS negocio_id uuid`,

		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM producto_negocio
				WHERE producto_id IS NOT NULL
				GROUP BY producto_id HAVING count(DISTINCT negocio_id) > 1
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: existen productos asociados a varios negocios';
			END IF;
		END $$`,

		`UPDATE productos p SET negocio_id = origen.negocio_id
		FROM (
			SELECT producto_id, min(negocio_id::text)::uuid AS negocio_id
			FROM producto_negocio GROUP BY producto_id
		) origen
		WHERE p.id = origen.producto_id AND p.negocio_id IS NULL`,

		// Los productos base de variantes no tienen fila en producto_negocio; su dueño es la familia.
		`DO $$ BEGIN
			IF to_regclass('familias_producto') IS NOT NULL AND EXISTS (
				SELECT 1 FROM information_schema.columns WHERE table_name = 'productos' AND column_name = 'familia_producto_id'
			) THEN
				UPDATE productos p SET negocio_id = f.negocio_id
				FROM familias_producto f
				WHERE f.id = p.familia_producto_id AND p.negocio_id IS NULL;
			END IF;
		END $$`,

		`DO $$ DECLARE total_negocios integer; BEGIN
			SELECT count(*) INTO total_negocios FROM negocios;
			IF total_negocios = 1 THEN
				UPDATE productos SET negocio_id = (SELECT id FROM negocios LIMIT 1) WHERE negocio_id IS NULL;
			ELSIF EXISTS (SELECT 1 FROM productos WHERE negocio_id IS NULL) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: productos sin negocio y múltiples negocios disponibles';
			END IF;
		END $$`,

		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM productos WHERE marca_id IS NOT NULL
				GROUP BY marca_id HAVING count(DISTINCT negocio_id) > 1
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: una marca legacy es usada por varios negocios';
			END IF;
			IF EXISTS (
				SELECT 1 FROM productos WHERE unidad_medida_id IS NOT NULL
				GROUP BY unidad_medida_id HAVING count(DISTINCT negocio_id) > 1
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: una unidad legacy es usada por varios negocios';
			END IF;
			IF EXISTS (
				SELECT 1 FROM producto_categorias pc
				JOIN productos p ON p.id = pc.producto_id
				GROUP BY pc.categoria_id HAVING count(DISTINCT p.negocio_id) > 1
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: una categoría legacy es usada por varios negocios';
			END IF;
		END $$`,

		`UPDATE marcas m SET negocio_id = origen.negocio_id
		FROM (SELECT marca_id, min(negocio_id::text)::uuid AS negocio_id FROM productos WHERE marca_id IS NOT NULL GROUP BY marca_id) origen
		WHERE m.id = origen.marca_id AND m.negocio_id IS NULL`,
		`UPDATE unidades_medida u SET negocio_id = origen.negocio_id
		FROM (SELECT unidad_medida_id, min(negocio_id::text)::uuid AS negocio_id FROM productos WHERE unidad_medida_id IS NOT NULL GROUP BY unidad_medida_id) origen
		WHERE u.id = origen.unidad_medida_id AND u.negocio_id IS NULL`,
		`UPDATE categorias c SET negocio_id = origen.negocio_id
		FROM (
			SELECT pc.categoria_id, min(p.negocio_id::text)::uuid AS negocio_id
			FROM producto_categorias pc JOIN productos p ON p.id = pc.producto_id
			GROUP BY pc.categoria_id
		) origen
		WHERE c.id = origen.categoria_id AND c.negocio_id IS NULL`,

		`DO $$ DECLARE total_negocios integer; BEGIN
			SELECT count(*) INTO total_negocios FROM negocios;
			IF total_negocios = 1 THEN
				UPDATE marcas SET negocio_id = (SELECT id FROM negocios LIMIT 1) WHERE negocio_id IS NULL;
				UPDATE categorias SET negocio_id = (SELECT id FROM negocios LIMIT 1) WHERE negocio_id IS NULL;
				UPDATE unidades_medida SET negocio_id = (SELECT id FROM negocios LIMIT 1) WHERE negocio_id IS NULL;
			ELSIF EXISTS (SELECT 1 FROM marcas WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM categorias WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM unidades_medida WHERE negocio_id IS NULL) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: catálogos sin dueño y múltiples negocios disponibles';
			END IF;
		END $$`,

		`UPDATE producto_categorias pc SET negocio_id = p.negocio_id
		FROM productos p WHERE p.id = pc.producto_id AND pc.negocio_id IS NULL`,
		`UPDATE producto_codigos codigo SET negocio_id = p.negocio_id
		FROM productos p WHERE p.id = codigo.producto_id AND codigo.negocio_id IS NULL`,
		`DO $$ BEGIN
			IF to_regclass('producto_variantes') IS NOT NULL AND EXISTS (
				SELECT 1 FROM information_schema.columns WHERE table_name = 'producto_codigos' AND column_name = 'producto_variante_id'
			) THEN
				UPDATE producto_codigos codigo SET negocio_id = p.negocio_id
				FROM producto_variantes v JOIN productos p ON p.id = v.producto_id
				WHERE v.id = codigo.producto_variante_id AND codigo.negocio_id IS NULL;
			END IF;
		END $$`,

		`DO $$ BEGIN
			IF EXISTS (SELECT 1 FROM marcas WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM categorias WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM unidades_medida WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM productos WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM producto_categorias WHERE negocio_id IS NULL)
				OR EXISTS (SELECT 1 FROM producto_codigos WHERE negocio_id IS NULL) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: no fue posible resolver todos los negocios propietarios';
			END IF;
			IF EXISTS (
				SELECT 1 FROM producto_negocio pn JOIN productos p ON p.id = pn.producto_id
				WHERE pn.negocio_id <> p.negocio_id
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: producto_negocio cruza negocios';
			END IF;
			IF EXISTS (
				SELECT 1 FROM producto_categorias pc
				JOIN categorias c ON c.id = pc.categoria_id
				WHERE pc.negocio_id <> c.negocio_id
			) THEN
				RAISE EXCEPTION 'migración de catálogo detenida: producto y categoría pertenecen a negocios distintos';
			END IF;
		END $$`,

		`ALTER TABLE marcas ALTER COLUMN negocio_id SET NOT NULL`,
		`ALTER TABLE categorias ALTER COLUMN negocio_id SET NOT NULL`,
		`ALTER TABLE unidades_medida ALTER COLUMN negocio_id SET NOT NULL`,
		`ALTER TABLE productos ALTER COLUMN negocio_id SET NOT NULL`,
		`ALTER TABLE producto_categorias ALTER COLUMN negocio_id SET NOT NULL`,
		`ALTER TABLE producto_codigos ALTER COLUMN negocio_id SET NOT NULL`,

		`DROP INDEX IF EXISTS idx_unidades_medida_codigo`,
		`DROP INDEX IF EXISTS idx_unidades_medida_nombre`,
		`DROP INDEX IF EXISTS idx_unidades_medida_simbolo`,
		`DROP INDEX IF EXISTS idx_producto_codigos_codigo`,
		`DROP INDEX IF EXISTS idx_familia_producto_codigo`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_marcas_negocio_nombre_ci ON marcas (negocio_id, lower(btrim(nombre)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_categorias_negocio_nombre_ci ON categorias (negocio_id, lower(btrim(nombre)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_unidades_negocio_codigo_ci ON unidades_medida (negocio_id, lower(btrim(codigo)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_unidades_negocio_nombre_ci ON unidades_medida (negocio_id, lower(btrim(nombre)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_unidades_negocio_simbolo_ci ON unidades_medida (negocio_id, lower(btrim(simbolo)))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_productos_negocio_id ON productos (negocio_id, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_marcas_negocio_id ON marcas (negocio_id, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_categorias_negocio_id ON categorias (negocio_id, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_unidades_negocio_id ON unidades_medida (negocio_id, id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_producto_codigos_negocio_codigo ON producto_codigos (negocio_id, codigo)`,

		catalogTenantConstraints,
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("migrar catálogo por negocio: %w", err)
			}
		}
		return nil
	})
}

const catalogTenantConstraints = `DO $$ BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_marcas_negocio') THEN
		ALTER TABLE marcas ADD CONSTRAINT fk_marcas_negocio FOREIGN KEY (negocio_id) REFERENCES negocios(id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_categorias_negocio') THEN
		ALTER TABLE categorias ADD CONSTRAINT fk_categorias_negocio FOREIGN KEY (negocio_id) REFERENCES negocios(id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_unidades_negocio') THEN
		ALTER TABLE unidades_medida ADD CONSTRAINT fk_unidades_negocio FOREIGN KEY (negocio_id) REFERENCES negocios(id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_productos_negocio_propietario') THEN
		ALTER TABLE productos ADD CONSTRAINT fk_productos_negocio_propietario FOREIGN KEY (negocio_id) REFERENCES negocios(id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_categorias_padre_mismo_negocio') THEN
		ALTER TABLE categorias ADD CONSTRAINT fk_categorias_padre_mismo_negocio
		FOREIGN KEY (negocio_id, categoria_padre_id) REFERENCES categorias(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_unidades_base_mismo_negocio') THEN
		ALTER TABLE unidades_medida ADD CONSTRAINT fk_unidades_base_mismo_negocio
		FOREIGN KEY (negocio_id, unidad_base_id) REFERENCES unidades_medida(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_productos_marca_mismo_negocio') THEN
		ALTER TABLE productos ADD CONSTRAINT fk_productos_marca_mismo_negocio
		FOREIGN KEY (negocio_id, marca_id) REFERENCES marcas(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_productos_unidad_mismo_negocio') THEN
		ALTER TABLE productos ADD CONSTRAINT fk_productos_unidad_mismo_negocio
		FOREIGN KEY (negocio_id, unidad_medida_id) REFERENCES unidades_medida(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_producto_categoria_producto_mismo_negocio') THEN
		ALTER TABLE producto_categorias ADD CONSTRAINT fk_producto_categoria_producto_mismo_negocio
		FOREIGN KEY (negocio_id, producto_id) REFERENCES productos(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_producto_categoria_categoria_mismo_negocio') THEN
		ALTER TABLE producto_categorias ADD CONSTRAINT fk_producto_categoria_categoria_mismo_negocio
		FOREIGN KEY (negocio_id, categoria_id) REFERENCES categorias(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_producto_codigo_producto_mismo_negocio') THEN
		ALTER TABLE producto_codigos ADD CONSTRAINT fk_producto_codigo_producto_mismo_negocio
		FOREIGN KEY (negocio_id, producto_id) REFERENCES productos(negocio_id, id);
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_producto_negocio_mismo_negocio') THEN
		ALTER TABLE producto_negocio ADD CONSTRAINT fk_producto_negocio_mismo_negocio
		FOREIGN KEY (negocio_id, producto_id) REFERENCES productos(negocio_id, id);
	END IF;
END $$`
