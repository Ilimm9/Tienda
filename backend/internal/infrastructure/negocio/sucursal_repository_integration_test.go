package negocio

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPhaseTwoMigrationAndRepository(t *testing.T) {
	url := os.Getenv("PHASE2_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PHASE2_TEST_DATABASE_URL no configurada")
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

	prepareLegacyBranchSchema(t, db)
	if err := MigratePhaseTwo(db); err != nil {
		t.Fatalf("primera migración: %v", err)
	}
	before := migrationSnapshot(t, db)
	if err := MigratePhaseTwo(db); err != nil {
		t.Fatalf("segunda migración: %v", err)
	}
	after := migrationSnapshot(t, db)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("segunda migración alteró datos:\nantes=%#v\ndespués=%#v", before, after)
	}

	if before.AddressCount != 3 {
		t.Fatalf("direcciones migradas = %d, se esperaban 3", before.AddressCount)
	}
	if before.Branches[0].Codigo != "SUC-001" || before.Branches[1].Codigo != "SUC-002" {
		t.Fatalf("backfill no determinista: %#v", before.Branches)
	}
	if !before.Branches[0].EsPrincipal || before.Branches[1].EsPrincipal || before.Branches[2].EsPrincipal {
		t.Fatalf("principal legacy incorrecta: %#v", before.Branches)
	}
	if before.Branches[0].Referencias != "Calle Uno" {
		t.Fatalf("dirección legacy no preservada: %#v", before.Branches[0])
	}

	repository := NewSucursalRepository(db)
	ctx := context.Background()
	businessA := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	businessB := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	newID, err := repository.Crear(ctx, businessA, domain.CrearSucursalInput{
		Codigo: "NORTE", Nombre: "Norte", EsPrincipal: true,
	})
	if err != nil {
		t.Fatalf("crear principal: %v", err)
	}
	detail, err := repository.Obtener(ctx, businessA, newID)
	if err != nil || !detail.EsPrincipal {
		t.Fatalf("nueva principal = %#v, err=%v", detail, err)
	}
	if err := repository.Archivar(ctx, businessA, newID); !errors.Is(err, application.ErrSucursalPrincipalRequerida) {
		t.Fatalf("archivar principal con alternativas: %v", err)
	}

	legacySecond := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-000000000002")
	makePrincipal := true
	if err := repository.Actualizar(ctx, businessA, legacySecond, domain.ActualizarSucursalInput{
		EsPrincipal: domain.Optional[bool]{Set: true, Value: &makePrincipal},
	}); err != nil {
		t.Fatalf("promover secundaria: %v", err)
	}
	if err := repository.Archivar(ctx, businessA, newID); err != nil {
		t.Fatalf("archivar secundaria: %v", err)
	}
	if err := repository.Restaurar(ctx, businessA, newID); err != nil {
		t.Fatalf("restaurar secundaria: %v", err)
	}
	detail, _ = repository.Obtener(ctx, businessA, newID)
	if detail.EsPrincipal {
		t.Fatal("restauración reemplazó una principal existente")
	}

	if _, err := repository.Crear(ctx, businessA, domain.CrearSucursalInput{Codigo: "NORTE", Nombre: "Duplicada"}); !errors.Is(err, application.ErrSucursalConflicto) {
		t.Fatalf("código duplicado mismo negocio: %v", err)
	}
	if _, err := repository.Crear(ctx, businessB, domain.CrearSucursalInput{Codigo: "NORTE", Nombre: "Permitida"}); err != nil {
		t.Fatalf("mismo código en otro negocio: %v", err)
	}
}

type branchSnapshot struct {
	ID           string
	Codigo       string
	EsPrincipal  bool
	Referencias  string
	DireccionID  string
	EliminadoSet bool
}

type phaseTwoSnapshot struct {
	Branches     []branchSnapshot
	AddressCount int64
}

func migrationSnapshot(t *testing.T, db *gorm.DB) phaseTwoSnapshot {
	t.Helper()
	var rows []struct {
		ID           string
		Codigo       string
		EsPrincipal  bool
		Referencias  *string
		DireccionID  *string
		EliminadoSet bool
	}
	if err := db.Table("sucursales AS s").
		Select("s.id::text AS id, s.codigo, s.es_principal, d.referencias, s.direccion_id::text AS direccion_id, s.eliminado_en IS NOT NULL AS eliminado_set").
		Joins("LEFT JOIN direcciones d ON d.id = s.direccion_id").Order("s.negocio_id, s.creado_en, s.id").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	result := phaseTwoSnapshot{Branches: make([]branchSnapshot, 0, len(rows))}
	for _, row := range rows {
		result.Branches = append(result.Branches, branchSnapshot{
			ID: row.ID, Codigo: row.Codigo, EsPrincipal: row.EsPrincipal,
			Referencias: valueOrEmpty(row.Referencias), DireccionID: valueOrEmpty(row.DireccionID),
			EliminadoSet: row.EliminadoSet,
		})
	}
	if err := db.Table("direcciones").Count(&result.AddressCount).Error; err != nil {
		t.Fatal(err)
	}
	return result
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func prepareLegacyBranchSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`DROP SCHEMA public CASCADE`,
		`CREATE SCHEMA public`,
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE negocios (id uuid PRIMARY KEY)`,
		`CREATE TABLE direcciones (
			id uuid PRIMARY KEY, codigo_pais varchar(2) NOT NULL DEFAULT 'MX', estado varchar(120),
			municipio varchar(120), ciudad varchar(120), colonia varchar(150), codigo_postal varchar(12),
			calle varchar(180), numero_exterior varchar(30), numero_interior varchar(30), referencias text,
			creado_en timestamptz NOT NULL, actualizado_en timestamptz NOT NULL
		)`,
		`CREATE TABLE sucursales (
			id uuid PRIMARY KEY, negocio_id uuid NOT NULL, nombre varchar(180) NOT NULL, telefono varchar(30),
			direccion text, activo boolean NOT NULL DEFAULT true, creado_en timestamptz NOT NULL,
			actualizado_en timestamptz NOT NULL
		)`,
		`INSERT INTO negocios (id) VALUES
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'), ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb')`,
		`INSERT INTO sucursales (id, negocio_id, nombre, direccion, activo, creado_en, actualizado_en) VALUES
			('aaaaaaaa-aaaa-4aaa-8aaa-000000000001', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Matriz', 'Calle Uno', true, '2026-01-01', '2026-01-01'),
			('aaaaaaaa-aaaa-4aaa-8aaa-000000000002', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Norte', 'Calle Dos', true, '2026-01-02', '2026-01-02'),
			('aaaaaaaa-aaaa-4aaa-8aaa-000000000003', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'Antigua', 'Calle Tres', false, '2026-01-03', '2026-01-03')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("preparar esquema legacy: %v", err)
		}
	}
}
