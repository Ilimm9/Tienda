package negocio

import (
	"context"
	"os"
	"strings"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestPhaseSevenAsignaciones valida la invariante de una sola principal activa por empleado.
func TestPhaseSevenAsignaciones(t *testing.T) {
	url := os.Getenv("PHASE7_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PHASE7_TEST_DATABASE_URL no configurada")
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

	negocioID, empleadoID, matriz, norte := prepararEsquemaAsignaciones(t, db)
	repository := NewAsignacionRepository(db)
	ctx := context.Background()

	// La primera asignación se vuelve principal aunque no se solicite.
	if _, err := repository.Asignar(ctx, negocioID, empleadoID, domain.AsignarSucursalInput{SucursalID: matriz}); err != nil {
		t.Fatalf("primera asignación: %v", err)
	}
	if principales(t, db, empleadoID) != 1 {
		t.Fatal("la primera asignación activa debe quedar como principal")
	}

	// La segunda entra como secundaria.
	segundaID, err := repository.Asignar(ctx, negocioID, empleadoID, domain.AsignarSucursalInput{SucursalID: norte})
	if err != nil {
		t.Fatalf("segunda asignación: %v", err)
	}
	if principales(t, db, empleadoID) != 1 {
		t.Fatal("no puede haber dos principales activas")
	}

	// Promover degrada la anterior en la misma transacción.
	if err := repository.EstablecerPrincipal(ctx, negocioID, empleadoID, segundaID); err != nil {
		t.Fatalf("promover: %v", err)
	}
	if principales(t, db, empleadoID) != 1 {
		t.Fatal("promover debe degradar a la principal anterior")
	}
	var principalActual string
	err = db.Raw(`SELECT sucursal_id::text FROM asignaciones_empleado_sucursal
		WHERE empleado_id = ? AND activo = TRUE AND es_principal = TRUE`, empleadoID).Scan(&principalActual).Error
	if err != nil || principalActual != norte.String() {
		t.Fatalf("la principal no es la promovida: %q %v", principalActual, err)
	}

	// Retirar la principal promueve automáticamente a la otra activa.
	if err := repository.Finalizar(ctx, negocioID, empleadoID, segundaID); err != nil {
		t.Fatalf("finalizar: %v", err)
	}
	if principales(t, db, empleadoID) != 1 {
		t.Fatal("al retirar la principal, otra activa debe ocuparla")
	}

	// Reasignar una sucursal retirada reactiva la fila en vez de duplicarla.
	if _, err := repository.Asignar(ctx, negocioID, empleadoID, domain.AsignarSucursalInput{SucursalID: norte}); err != nil {
		t.Fatalf("reasignar: %v", err)
	}
	var totalFilas int
	err = db.Raw(`SELECT count(*) FROM asignaciones_empleado_sucursal
		WHERE empleado_id = ? AND sucursal_id = ?`, empleadoID, norte).Scan(&totalFilas).Error
	if err != nil || totalFilas != 1 {
		t.Fatalf("el único (empleado, sucursal) debe respetarse: %d %v", totalFilas, err)
	}

	// Retirar todas deja al empleado sin principal, sin violar la invariante.
	if err := db.Exec(`UPDATE asignaciones_empleado_sucursal SET activo = FALSE, es_principal = FALSE
		WHERE empleado_id = ?`, empleadoID).Error; err != nil {
		t.Fatal(err)
	}
	if principales(t, db, empleadoID) != 0 {
		t.Fatal("sin asignaciones activas no debe quedar principal")
	}
}

func principales(t *testing.T, db *gorm.DB, empleadoID uuid.UUID) int {
	t.Helper()
	var total int
	err := db.Raw(`SELECT count(*) FROM asignaciones_empleado_sucursal
		WHERE empleado_id = ? AND activo = TRUE AND es_principal = TRUE`, empleadoID).Scan(&total).Error
	if err != nil {
		t.Fatal(err)
	}
	return total
}

func prepararEsquemaAsignaciones(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	preparacion := []string{
		`DROP TABLE IF EXISTS asignaciones_empleado_sucursal, empleados, sucursales, negocios CASCADE`,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE negocios (id uuid PRIMARY KEY, nombre varchar(180) NOT NULL,
			estado varchar(30) NOT NULL DEFAULT 'activo')`,
		`CREATE TABLE sucursales (id uuid PRIMARY KEY, negocio_id uuid NOT NULL REFERENCES negocios(id),
			codigo varchar(40) NOT NULL, nombre varchar(180) NOT NULL,
			activo boolean NOT NULL DEFAULT true, eliminado_en timestamptz)`,
		`CREATE TABLE empleados (id uuid PRIMARY KEY, negocio_id uuid NOT NULL REFERENCES negocios(id),
			nombre varchar(100) NOT NULL, primer_apellido varchar(100) NOT NULL,
			estado varchar(30) NOT NULL DEFAULT 'pendiente',
			creado_por_usuario_id uuid NOT NULL,
			creado_en timestamptz NOT NULL DEFAULT now(),
			actualizado_en timestamptz NOT NULL DEFAULT now())`,
		`CREATE TABLE asignaciones_empleado_sucursal (
			id uuid PRIMARY KEY,
			negocio_id uuid NOT NULL REFERENCES negocios(id),
			empleado_id uuid NOT NULL REFERENCES empleados(id),
			sucursal_id uuid NOT NULL REFERENCES sucursales(id),
			es_principal boolean NOT NULL DEFAULT false,
			activo boolean NOT NULL DEFAULT true,
			asignado_en timestamptz NOT NULL DEFAULT now(),
			finalizado_en timestamptz)`,
		`CREATE UNIQUE INDEX idx_asignacion_empleado_sucursal
			ON asignaciones_empleado_sucursal (empleado_id, sucursal_id)`,
		`CREATE UNIQUE INDEX idx_asignacion_principal_activa
			ON asignaciones_empleado_sucursal (empleado_id)
			WHERE es_principal = true AND activo = true`,
	}
	for _, statement := range preparacion {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}

	negocioID, empleadoID := uuid.New(), uuid.New()
	matriz, norte := uuid.New(), uuid.New()
	datos := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO negocios (id, nombre) VALUES (?, 'Tienda Centro')`, []any{negocioID}},
		{`INSERT INTO sucursales (id, negocio_id, codigo, nombre) VALUES (?, ?, 'SUC-001', 'Matriz')`,
			[]any{matriz, negocioID}},
		{`INSERT INTO sucursales (id, negocio_id, codigo, nombre) VALUES (?, ?, 'SUC-002', 'Norte')`,
			[]any{norte, negocioID}},
		{`INSERT INTO empleados (id, negocio_id, nombre, primer_apellido, creado_por_usuario_id)
			VALUES (?, ?, 'Ana', 'López', ?)`, []any{empleadoID, negocioID, uuid.New()}},
	}
	for _, dato := range datos {
		if err := db.Exec(dato.sql, dato.args...).Error; err != nil {
			t.Fatalf("%s: %v", dato.sql, err)
		}
	}
	return negocioID, empleadoID, matriz, norte
}
