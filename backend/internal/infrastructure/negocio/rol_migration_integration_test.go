package negocio

import (
	"os"
	"strings"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestPhaseFourMigration valida la convergencia con base.MD sobre una base desechable con sufijo _test.
func TestPhaseFourMigration(t *testing.T) {
	url := os.Getenv("PHASE4_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PHASE4_TEST_DATABASE_URL no configurada")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir base temporal: %v", err)
	}
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("la prueba destructiva requiere una base con sufijo _test; actual=%q", databaseName)
	}

	negocioID, membresiaPropietario, membresiaMiembro, rolLegacyID := prepararEsquemaLegacyRoles(t, db)

	if err := MigratePhaseFour(db); err != nil {
		t.Fatalf("primera migración: %v", err)
	}
	primera := snapshotRoles(t, db)
	if err := MigratePhaseFour(db); err != nil {
		t.Fatalf("segunda migración: %v", err)
	}
	segunda := snapshotRoles(t, db)

	if primera != segunda {
		t.Fatalf("la migración no es idempotente:\nprimera=%#v\nsegunda=%#v", primera, segunda)
	}

	// El catálogo global queda sembrado completo.
	if primera.Permisos != len(domain.CatalogoPermisos) {
		t.Fatalf("permisos sembrados = %d, se esperaban %d", primera.Permisos, len(domain.CatalogoPermisos))
	}

	// El rol legacy conserva su fila y recibe un código derivado del nombre.
	var codigoLegacy string
	if err := db.Raw(`SELECT codigo FROM roles WHERE id = ?`, rolLegacyID).Scan(&codigoLegacy).Error; err != nil {
		t.Fatal(err)
	}
	if codigoLegacy != "ENCARGADO_DE_PISO" {
		t.Fatalf("backfill de código inesperado: %q", codigoLegacy)
	}

	// El rol de sistema existe una sola vez por negocio y concentra todos los permisos.
	var rolSistemaID string
	err = db.Raw(`SELECT id::text FROM roles WHERE negocio_id = ? AND es_rol_sistema = true`, negocioID).Scan(&rolSistemaID).Error
	if err != nil || rolSistemaID == "" {
		t.Fatalf("rol de sistema ausente: %v", err)
	}
	var permisosDelRol int
	if err := db.Raw(`SELECT count(*) FROM permisos_rol WHERE rol_id = ?`, rolSistemaID).Scan(&permisosDelRol).Error; err != nil {
		t.Fatal(err)
	}
	if permisosDelRol != len(domain.CatalogoPermisos) {
		t.Fatalf("el rol de sistema tiene %d permisos, se esperaban %d", permisosDelRol, len(domain.CatalogoPermisos))
	}

	// El propietario activo conserva el rol de sistema.
	var propietarioTieneSistema int
	err = db.Raw(`SELECT count(*) FROM roles_membresia WHERE membresia_negocio_id = ? AND rol_id = ?`,
		membresiaPropietario, rolSistemaID).Scan(&propietarioTieneSistema).Error
	if err != nil || propietarioTieneSistema != 1 {
		t.Fatalf("propietario sin rol de sistema: %d %v", propietarioTieneSistema, err)
	}

	// El rol único legacy quedó backfilleado hacia la relación muchos-a-muchos.
	var miembroTieneLegacy int
	err = db.Raw(`SELECT count(*) FROM roles_membresia WHERE membresia_negocio_id = ? AND rol_id = ?`,
		membresiaMiembro, rolLegacyID).Scan(&miembroTieneLegacy).Error
	if err != nil || miembroTieneLegacy != 1 {
		t.Fatalf("rol legacy no backfilleado: %d %v", miembroTieneLegacy, err)
	}

	// base.MD no declara `rol_id` en membresias_negocio.
	var columnaLegacy int
	err = db.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'membresias_negocio' AND column_name = 'rol_id'`).Scan(&columnaLegacy).Error
	if err != nil || columnaLegacy != 0 {
		t.Fatalf("la columna rol_id debía eliminarse tras el backfill: %d %v", columnaLegacy, err)
	}

	// El único por negocio no distingue mayúsculas y no cruza negocios.
	otroNegocio := uuid.New()
	if err := db.Exec(`INSERT INTO negocios (id, nombre, estado) VALUES (?, 'Otro', 'activo')`, otroNegocio).Error; err != nil {
		t.Fatal(err)
	}
	err = db.Exec(`INSERT INTO roles (id, negocio_id, codigo, nombre, creado_en, actualizado_en)
		VALUES (?, ?, 'encargado_de_piso', 'Duplicado', now(), now())`, uuid.New(), negocioID).Error
	if err == nil {
		t.Fatal("el código de rol debe ser único por negocio sin distinguir mayúsculas")
	}
	err = db.Exec(`INSERT INTO roles (id, negocio_id, codigo, nombre, creado_en, actualizado_en)
		VALUES (?, ?, 'ENCARGADO_DE_PISO', 'Permitido', now(), now())`, uuid.New(), otroNegocio).Error
	if err != nil {
		t.Fatalf("el mismo código debe poder existir en otro negocio: %v", err)
	}
}

type snapshotRolesResultado struct {
	Roles           int
	RolesSistema    int
	Permisos        int
	PermisosRol     int
	RolesMembresia  int
	CodigosVacios   int
}

func snapshotRoles(t *testing.T, db *gorm.DB) snapshotRolesResultado {
	t.Helper()
	var resultado snapshotRolesResultado
	consultas := map[string]*int{
		`SELECT count(*) FROM roles`:                                            &resultado.Roles,
		`SELECT count(*) FROM roles WHERE es_rol_sistema = true`:                &resultado.RolesSistema,
		`SELECT count(*) FROM permisos`:                                         &resultado.Permisos,
		`SELECT count(*) FROM permisos_rol`:                                     &resultado.PermisosRol,
		`SELECT count(*) FROM roles_membresia`:                                  &resultado.RolesMembresia,
		`SELECT count(*) FROM roles WHERE codigo IS NULL OR btrim(codigo) = ''`: &resultado.CodigosVacios,
	}
	for consulta, destino := range consultas {
		if err := db.Raw(consulta).Scan(destino).Error; err != nil {
			t.Fatalf("%s: %v", consulta, err)
		}
	}
	return resultado
}

// prepararEsquemaLegacyRoles reproduce la forma previa: roles sin código y membresías con rol único.
func prepararEsquemaLegacyRoles(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	preparacion := []string{
		`DROP TABLE IF EXISTS roles_membresia, permisos_rol, permisos, roles, membresias_negocio, negocios CASCADE`,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE negocios (
			id uuid PRIMARY KEY,
			nombre varchar(180) NOT NULL,
			estado varchar(30) NOT NULL DEFAULT 'activo',
			creado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE membresias_negocio (
			id uuid PRIMARY KEY,
			negocio_id uuid NOT NULL REFERENCES negocios(id),
			usuario_id uuid NOT NULL,
			rol_id uuid,
			tipo_miembro varchar(30) NOT NULL DEFAULT 'miembro',
			estado varchar(30) NOT NULL DEFAULT 'activo',
			creado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE roles (
			id uuid PRIMARY KEY,
			negocio_id uuid NOT NULL,
			nombre varchar(100) NOT NULL,
			descripcion text,
			creado_en timestamptz NOT NULL DEFAULT now(),
			actualizado_en timestamptz NOT NULL DEFAULT now()
		)`,
	}
	for _, statement := range preparacion {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}

	negocioID := uuid.New()
	rolLegacyID := uuid.New()
	membresiaPropietario := uuid.New()
	membresiaMiembro := uuid.New()
	datos := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO negocios (id, nombre, estado) VALUES (?, 'Tienda Centro', 'activo')`, []any{negocioID}},
		{`INSERT INTO roles (id, negocio_id, nombre) VALUES (?, ?, 'Encargado de piso')`, []any{rolLegacyID, negocioID}},
		{`INSERT INTO membresias_negocio (id, negocio_id, usuario_id, tipo_miembro, estado)
			VALUES (?, ?, ?, 'propietario', 'activo')`, []any{membresiaPropietario, negocioID, uuid.New()}},
		{`INSERT INTO membresias_negocio (id, negocio_id, usuario_id, rol_id, tipo_miembro, estado)
			VALUES (?, ?, ?, ?, 'miembro', 'activo')`, []any{membresiaMiembro, negocioID, uuid.New(), rolLegacyID}},
	}
	for _, dato := range datos {
		if err := db.Exec(dato.sql, dato.args...).Error; err != nil {
			t.Fatalf("%s: %v", dato.sql, err)
		}
	}
	return negocioID, membresiaPropietario, membresiaMiembro, rolLegacyID
}
