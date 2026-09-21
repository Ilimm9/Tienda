package database

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	cuentaapplication "tienda/backend/internal/application/cuenta"
	"tienda/backend/internal/domain"
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"
	"tienda/backend/internal/infrastructure"
	cuentainfra "tienda/backend/internal/infrastructure/cuenta"

	"github.com/google/uuid"
)

func TestSecurityPhaseCatalogTenancyMigrationAndIsolation(t *testing.T) {
	url := os.Getenv("SECURITY_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SECURITY_TEST_DATABASE_URL no está configurada")
	}
	if !strings.Contains(url, "tienda_security_test") {
		t.Fatal("la prueba destructiva sólo puede usar una base llamada tienda_security_test")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`DROP SCHEMA public CASCADE`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE SCHEMA public`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Init(db); err != nil {
		t.Fatalf("primera migración: %v", err)
	}
	if err := Init(db); err != nil {
		t.Fatalf("migración idempotente: %v", err)
	}

	first := negociodomain.Negocio{ID: uuid.New(), Slug: "negocio-uno", NombreComercial: "Negocio Uno", Nombre: "Negocio Uno", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	second := negociodomain.Negocio{ID: uuid.New(), Slug: "negocio-dos", NombreComercial: "Negocio Dos", Nombre: "Negocio Dos", CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
	if err := db.Create(&first).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}

	repository := infrastructure.NewProductRepository(db)
	if err := repository.CreateBrand(first.ID, domain.CreateMarcaInput{Nombre: "Compartida"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.CreateBrand(second.ID, domain.CreateMarcaInput{Nombre: "Compartida"}); err != nil {
		t.Fatalf("el mismo nombre debe permitirse en otro negocio: %v", err)
	}
	firstBrands, err := repository.ListBrandsAdmin(first.ID)
	if err != nil || len(firstBrands) != 1 || firstBrands[0].NegocioID != first.ID {
		t.Fatalf("marcas primer negocio=%+v error=%v", firstBrands, err)
	}
	secondBrands, err := repository.ListBrandsAdmin(second.ID)
	if err != nil || len(secondBrands) != 1 || secondBrands[0].NegocioID != second.ID {
		t.Fatalf("marcas segundo negocio=%+v error=%v", secondBrands, err)
	}
	newName := "Intrusión"
	if err := repository.UpdateBrand(first.ID, secondBrands[0].ID, domain.UpdateMarcaInput{Nombre: &newName}); err == nil {
		t.Fatal("un negocio pudo modificar la marca de otro")
	}
	var unchanged domain.Marca
	if err := db.First(&unchanged, "id = ?", secondBrands[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.Nombre != "Compartida" {
		t.Fatalf("la marca ajena cambió a %q", unchanged.Nombre)
	}

	verifiedAt := time.Now().UTC()
	user := cuentadomain.Usuario{ID: uuid.New(), Correo: "sesiones@example.com", HashContrasena: "no-usada-en-prueba", Estado: "activo", CorreoVerificadoEn: &verifiedAt}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	sessions := cuentaapplication.NewSessionService(cuentainfra.NewSessionRepository(db), cuentaapplication.SessionConfig{
		Duration: 24 * time.Hour, RememberDuration: 30 * 24 * time.Hour,
		RememberIdleDuration: 7 * 24 * time.Hour, ActivityTouchInterval: 5 * time.Minute,
	})
	firstSession, err := sessions.Create(user.ID, false, "127.0.0.1", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	secondSession, err := sessions.Create(user.ID, false, "127.0.0.1", "integration-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Revoke(firstSession.Token, "integration-test"); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.Authenticate(firstSession.Token); !errors.Is(err, cuentaapplication.ErrInvalidSession) {
		t.Fatalf("la sesión revocada se reutilizó: %v", err)
	}
	if _, err := sessions.Authenticate(secondSession.Token); err != nil {
		t.Fatalf("revocar una sesión afectó una sesión independiente: %v", err)
	}
}
