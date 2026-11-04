package database

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	negocioapplication "tienda/backend/internal/application/negocio"
	negociodomain "tienda/backend/internal/domain/negocio"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"

	"github.com/google/uuid"
)

// TestVistasNegocio covers Configuración > Vistas (adecuación 1): the
// preference is stored per business and is created when the row is missing.
func TestVistasNegocio(t *testing.T) {
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
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := negocioinfra.NewVistaRepository(db)
	crearNegocio := func(slug string) uuid.UUID {
		negocio := negociodomain.Negocio{ID: uuid.New(), Slug: slug, NombreComercial: slug, Nombre: slug, CodigoMoneda: "MXN", ZonaHoraria: "America/Mexico_City", Estado: "activo"}
		if err := db.Create(&negocio).Error; err != nil {
			t.Fatal(err)
		}
		return negocio.ID
	}
	conConfiguracion, sinConfiguracion := crearNegocio("con-config"), crearNegocio("sin-config")
	if err := db.Create(&negociodomain.ConfiguracionNegocio{NegocioID: conConfiguracion, MetodoCosteo: "promedio_ponderado"}).Error; err != nil {
		t.Fatal(err)
	}
	service := negocioapplication.NewVistaService(repo, propietarioDe{conConfiguracion, sinConfiguracion})

	for _, negocioID := range []uuid.UUID{conConfiguracion, sinConfiguracion} {
		vistas, err := service.Obtener(ctx, negocioID)
		if err != nil || len(vistas.SeccionesOcultas) != 0 {
			t.Fatalf("inicio con todo visible: %+v %v", vistas, err)
		}
		if _, err := service.Guardar(ctx, uuid.New(), negocioID, []string{"categorias"}); err != nil {
			t.Fatal(err)
		}
	}
	var filas int64
	if err := db.Model(&negociodomain.ConfiguracionNegocio{}).Where("negocio_id = ?", sinConfiguracion).Count(&filas).Error; err != nil || filas != 1 {
		t.Fatalf("configuración creada = %d, %v", filas, err)
	}
	if _, err := service.Guardar(ctx, uuid.New(), conConfiguracion, []string{"marcas", "categorias"}); err != nil {
		t.Fatal(err)
	}
	vistas, err := service.Obtener(ctx, conConfiguracion)
	if err != nil || !reflect.DeepEqual(vistas.SeccionesOcultas, []string{"categorias", "marcas"}) {
		t.Fatalf("con configuración: %+v %v", vistas, err)
	}
	vistas, err = service.Obtener(ctx, sinConfiguracion)
	if err != nil || !reflect.DeepEqual(vistas.SeccionesOcultas, []string{"categorias"}) {
		t.Fatalf("aislamiento por negocio: %+v %v", vistas, err)
	}
	if _, err := service.Guardar(ctx, uuid.New(), conConfiguracion, []string{}); err != nil {
		t.Fatal(err)
	}
	if vistas, _ := service.Obtener(ctx, conConfiguracion); len(vistas.SeccionesOcultas) != 0 {
		t.Fatalf("mostrar todo de nuevo: %+v", vistas)
	}
}

// propietarioDe hace propietario al usuario de los negocios indicados.
type propietarioDe []uuid.UUID

func (p propietarioDe) ListarOpcionesContexto(context.Context, uuid.UUID) ([]negociodomain.ContextoNegocio, error) {
	opciones := make([]negociodomain.ContextoNegocio, 0, len(p))
	for _, id := range p {
		opciones = append(opciones, negociodomain.ContextoNegocio{ID: id, TipoMiembro: "propietario"})
	}
	return opciones, nil
}

func (propietarioDe) NegocioActivoAccesible(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func (propietarioDe) SucursalActivaAccesible(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func (propietarioDe) SucursalesAccesibles(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
