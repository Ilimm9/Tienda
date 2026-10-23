package infrastructure

import (
	"context"
	"os"
	"strings"
	"testing"

	"tienda/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requiere una base `_test` ya migrada por la API (database.Init) con al menos un negocio sin semilla.
func TestCatalogSeederSiembraUnidadesYProveedor(t *testing.T) {
	url := os.Getenv("SEMILLA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SEMILLA_TEST_DATABASE_URL no configurada")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		t.Fatalf("abrir base temporal: %v", err)
	}
	var databaseName string
	if err := db.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil || !strings.HasSuffix(databaseName, "_test") {
		t.Fatalf("se requiere una base con sufijo _test; actual=%q err=%v", databaseName, err)
	}
	var ids []string
	if err := db.Raw("SELECT id::text FROM negocios ORDER BY creado_en DESC LIMIT 1").Scan(&ids).Error; err != nil || len(ids) == 0 {
		t.Fatalf("se requiere un negocio: %v", err)
	}
	negocioID := uuid.MustParse(ids[0])

	seeder := NewCatalogSeeder(db)
	if err := seeder.SembrarCatalogoInicial(context.Background(), negocioID); err != nil {
		t.Fatalf("sembrar: %v", err)
	}
	// Las unidades se omiten si ya existen: una segunda corrida no falla ni duplica.
	if err := seeder.SembrarCatalogoInicial(context.Background(), negocioID); err != nil {
		t.Fatalf("segunda semilla: %v", err)
	}
	var unidades []domain.UnidadMedida
	db.Where("negocio_id = ? AND tipo = ?", negocioID, "EMPAQUE").Order("factor_a_base").Find(&unidades)
	if len(unidades) != 3 || unidades[0].Nombre != "Pieza" || unidades[0].FactorABase != 1 || unidades[1].FactorABase != 12 || unidades[2].FactorABase != 24 {
		t.Fatalf("unidades sembradas: %#v", unidades)
	}
	var proveedores int64
	db.Model(&domain.Proveedor{}).Where("negocio_id = ? AND nombre = ? AND activo", negocioID, NombreProveedorGeneral).Count(&proveedores)
	if proveedores == 0 {
		t.Fatal("falta el proveedor general")
	}
}
