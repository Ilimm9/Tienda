package negocio

import (
	"context"
	"os"
	"sort"
	"strings"
	"testing"

	domain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Requiere una base `_test` ya migrada por la API (database.Init) con al menos un usuario.
func TestCrearNegocioSiembraRolCajero(t *testing.T) {
	url := os.Getenv("CAJERO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("CAJERO_TEST_DATABASE_URL no configurada")
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
	if err := db.Raw("SELECT id::text FROM usuarios LIMIT 1").Scan(&ids).Error; err != nil || len(ids) == 0 {
		t.Fatalf("se requiere un usuario: %v", err)
	}
	usuarioID := uuid.MustParse(ids[0])

	negocio, err := NewNegocioRepository(db).Crear(context.Background(), usuarioID, "cajero-"+uuid.NewString()[:8],
		domain.CrearNegocioInput{NombreComercial: "Prueba Cajero", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City"})
	if err != nil {
		t.Fatalf("crear negocio: %v", err)
	}
	var rol domain.Rol
	if err := db.Where("negocio_id = ? AND codigo = ?", negocio.ID, domain.CodigoRolCajero).Take(&rol).Error; err != nil {
		t.Fatalf("el negocio nuevo debe tener rol Cajero: %v", err)
	}
	if rol.EsRolSistema || !rol.Activo || rol.Nombre != "Cajero" {
		t.Fatalf("Cajero debe ser editable y activo: %#v", rol)
	}
	var permisos []string
	db.Raw(`SELECT p.codigo FROM permisos_rol pr JOIN permisos p ON p.id = pr.permiso_id WHERE pr.rol_id = ?`, rol.ID).Scan(&permisos)
	sort.Strings(permisos)
	if strings.Join(permisos, ",") != "ventas.registrar" {
		t.Fatalf("permisos del Cajero: %v", permisos)
	}
	var puesto domain.PuestoNegocio
	if err := db.Where("negocio_id = ? AND nombre = ?", negocio.ID, "Cajero").Take(&puesto).Error; err != nil {
		t.Fatalf("el negocio nuevo debe tener puesto Cajero: %v", err)
	}
	if !puesto.Activo || puesto.RolSugeridoID == nil || *puesto.RolSugeridoID != rol.ID {
		t.Fatalf("el puesto Cajero debe sugerir el rol Cajero: %#v", puesto)
	}
}
