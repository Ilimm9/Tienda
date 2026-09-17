package negocio

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestPhaseFiveMigration valida la convergencia de `empleados` con base.MD sobre una base desechable.
func TestPhaseFiveMigration(t *testing.T) {
	url := os.Getenv("PHASE5_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PHASE5_TEST_DATABASE_URL no configurada")
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

	empleadoID, negocioID, membresiaID := prepararEsquemaLegacyEmpleados(t, db)

	if err := MigratePhaseFive(db); err != nil {
		t.Fatalf("primera migración: %v", err)
	}
	primera := snapshotEmpleados(t, db)
	if err := MigratePhaseFive(db); err != nil {
		t.Fatalf("segunda migración: %v", err)
	}
	if segunda := snapshotEmpleados(t, db); primera != segunda {
		t.Fatalf("la migración no es idempotente:\nprimera=%#v\nsegunda=%#v", primera, segunda)
	}

	// La fila legacy se conserva y toma su identidad desde el perfil.
	var fila struct {
		NegocioID      string
		MembresiaID    string
		NumeroEmpleado string
		Nombre         string
		PrimerApellido string
		Estado         string
	}
	err = db.Raw(`SELECT negocio_id::text, COALESCE(membresia_id::text, '') AS membresia_id,
		COALESCE(numero_empleado, '') AS numero_empleado, nombre, primer_apellido, estado
		FROM empleados WHERE id = ?`, empleadoID).Scan(&fila).Error
	if err != nil {
		t.Fatal(err)
	}
	if fila.NegocioID != negocioID.String() {
		t.Fatalf("negocio_id no derivado de la membresía: %q", fila.NegocioID)
	}
	if fila.MembresiaID != membresiaID.String() {
		t.Fatalf("el vínculo debió invertirse hacia empleados.membresia_id: %q", fila.MembresiaID)
	}
	if fila.Nombre != "Ana" || fila.PrimerApellido != "López" {
		t.Fatalf("identidad no derivada del perfil: %#v", fila)
	}
	if fila.NumeroEmpleado != "EMP-001" {
		t.Fatalf("numero no renombrado a numero_empleado: %q", fila.NumeroEmpleado)
	}

	// base.MD no declara estas columnas.
	for tabla, columna := range map[string]string{"membresias_negocio": "empleado_id", "empleados": "perfil_id"} {
		var total int
		err := db.Raw(`SELECT count(*) FROM information_schema.columns
			WHERE table_name = ? AND column_name = ?`, tabla, columna).Scan(&total).Error
		if err != nil || total != 0 {
			t.Fatalf("%s.%s debía eliminarse: %d %v", tabla, columna, total, err)
		}
	}

	// El número es único por negocio, no global.
	otroNegocio := uuid.New()
	if err := db.Exec(`INSERT INTO negocios (id, nombre, estado) VALUES (?, 'Otro', 'activo')`, otroNegocio).Error; err != nil {
		t.Fatal(err)
	}
	insertar := `INSERT INTO empleados (id, negocio_id, numero_empleado, nombre, primer_apellido, estado, creado_por_usuario_id, creado_en, actualizado_en)
		VALUES (?, ?, 'EMP-001', 'Luis', 'Pérez', 'pendiente', ?, now(), now())`
	if err := db.Exec(insertar, uuid.New(), negocioID, uuid.New()).Error; err == nil {
		t.Fatal("el número debe ser único dentro del negocio")
	}
	if err := db.Exec(insertar, uuid.New(), otroNegocio, uuid.New()).Error; err != nil {
		t.Fatalf("el mismo número debe poder existir en otro negocio: %v", err)
	}

	// Varios empleados sin número conviven: el único es parcial.
	sinNumero := `INSERT INTO empleados (id, negocio_id, nombre, primer_apellido, estado, creado_por_usuario_id, creado_en, actualizado_en)
		VALUES (?, ?, 'Sin', 'Numero', 'pendiente', ?, now(), now())`
	for range 2 {
		if err := db.Exec(sinNumero, uuid.New(), negocioID, uuid.New()).Error; err != nil {
			t.Fatalf("el único por número debe ser parcial: %v", err)
		}
	}
}

type snapshotEmpleadosResultado struct {
	Empleados      int
	SinNegocio     int
	SinNombre      int
	ConMembresia   int
}

func snapshotEmpleados(t *testing.T, db *gorm.DB) snapshotEmpleadosResultado {
	t.Helper()
	var resultado snapshotEmpleadosResultado
	consultas := map[string]*int{
		`SELECT count(*) FROM empleados`:                                        &resultado.Empleados,
		`SELECT count(*) FROM empleados WHERE negocio_id IS NULL`:               &resultado.SinNegocio,
		`SELECT count(*) FROM empleados WHERE nombre IS NULL OR btrim(nombre) = ''`: &resultado.SinNombre,
		`SELECT count(*) FROM empleados WHERE membresia_id IS NOT NULL`:         &resultado.ConMembresia,
	}
	for consulta, destino := range consultas {
		if err := db.Raw(consulta).Scan(destino).Error; err != nil {
			t.Fatalf("%s: %v", consulta, err)
		}
	}
	return resultado
}

// prepararEsquemaLegacyEmpleados reproduce la forma previa: perfil_id y número único global.
func prepararEsquemaLegacyEmpleados(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	preparacion := []string{
		`DROP TABLE IF EXISTS empleados, membresias_negocio, perfiles_usuario, negocios CASCADE`,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE negocios (
			id uuid PRIMARY KEY, nombre varchar(180) NOT NULL,
			estado varchar(30) NOT NULL DEFAULT 'activo'
		)`,
		`CREATE TABLE perfiles_usuario (
			id uuid PRIMARY KEY, usuario_id uuid NOT NULL,
			nombre varchar(100), segundo_nombre varchar(100),
			primer_apellido varchar(100), segundo_apellido varchar(100), telefono varchar(30)
		)`,
		`CREATE TABLE membresias_negocio (
			id uuid PRIMARY KEY, negocio_id uuid NOT NULL REFERENCES negocios(id),
			usuario_id uuid NOT NULL, empleado_id uuid,
			tipo_miembro varchar(30) NOT NULL DEFAULT 'miembro',
			estado varchar(30) NOT NULL DEFAULT 'activo'
		)`,
		`CREATE TABLE empleados (
			id uuid PRIMARY KEY, perfil_id uuid NOT NULL,
			numero varchar(50) NOT NULL, puesto varchar(120),
			estado varchar(30) NOT NULL DEFAULT 'activo',
			creado_en timestamptz NOT NULL DEFAULT now(),
			actualizado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX uni_empleados_numero ON empleados (numero)`,
	}
	for _, statement := range preparacion {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}

	negocioID, perfilID, empleadoID, membresiaID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	usuarioID := uuid.New()
	datos := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO negocios (id, nombre) VALUES (?, 'Tienda Centro')`, []any{negocioID}},
		{`INSERT INTO perfiles_usuario (id, usuario_id, nombre, primer_apellido, telefono)
			VALUES (?, ?, 'Ana', 'López', '5500000000')`, []any{perfilID, usuarioID}},
		{`INSERT INTO empleados (id, perfil_id, numero, puesto) VALUES (?, ?, 'EMP-001', 'Cajera')`,
			[]any{empleadoID, perfilID}},
		{`INSERT INTO membresias_negocio (id, negocio_id, usuario_id, empleado_id)
			VALUES (?, ?, ?, ?)`, []any{membresiaID, negocioID, usuarioID, empleadoID}},
	}
	for _, dato := range datos {
		if err := db.Exec(dato.sql, dato.args...).Error; err != nil {
			t.Fatalf("%s: %v", dato.sql, err)
		}
	}
	return empleadoID, negocioID, membresiaID
}
