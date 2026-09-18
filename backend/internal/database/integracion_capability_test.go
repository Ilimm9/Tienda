package database

import (
	"context"
	"os"
	"strings"
	"testing"

	negocioapplication "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"
	negocioinfra "tienda/backend/internal/infrastructure/negocio"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestCapabilidadCompletaAislamiento es la verificación integral de fase 8.
//
// Arranca el esquema real con Init, crea dos negocios de cuentas distintas y comprueba que
// ninguna operación de la capability cruza la frontera del negocio.
func TestCapabilidadCompletaAislamiento(t *testing.T) {
	url := os.Getenv("PHASE8_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("PHASE8_TEST_DATABASE_URL no configurada")
	}
	db, err := Open(url)
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

	// El esquema completo debe poder construirse desde cero y volver a ejecutarse sin cambios.
	if err := Init(db); err != nil {
		t.Fatalf("primer Init: %v", err)
	}
	if err := Init(db); err != nil {
		t.Fatalf("segundo Init: la migración no es idempotente: %v", err)
	}

	ctx := context.Background()
	negocioRepo := negocioinfra.NewNegocioRepository(db)
	negocioService := negocioapplication.NewNegocioService(negocioRepo)
	rolService := negocioapplication.NewRolService(negocioinfra.NewRolRepository(db))
	empleadoService := negocioapplication.NewEmpleadoService(negocioinfra.NewEmpleadoRepository(db))
	sucursalService := negocioapplication.NewSucursalService(negocioinfra.NewSucursalRepository(db))
	asignacionService := negocioapplication.NewAsignacionService(negocioinfra.NewAsignacionRepository(db))

	// Correos únicos por ejecución: la prueba debe poder repetirse sobre la misma base temporal.
	sufijo := uuid.NewString()[:8]
	ana := crearUsuario(t, db, "ana-"+sufijo+"@tienda.mx")
	beto := crearUsuario(t, db, "beto-"+sufijo+"@tienda.mx")

	negocioAna, err := negocioService.Crear(ctx, ana, domain.CrearNegocioInput{NombreComercial: "Tienda Ana"})
	if err != nil {
		t.Fatalf("crear negocio de Ana: %v", err)
	}
	negocioBeto, err := negocioService.Crear(ctx, beto, domain.CrearNegocioInput{NombreComercial: "Tienda Beto"})
	if err != nil {
		t.Fatalf("crear negocio de Beto: %v", err)
	}

	// Un negocio creado en runtime debe dejar a su propietario con permisos completos de inmediato.
	permisos, err := rolService.PermisosEfectivos(ctx, ana, negocioAna.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(permisos) != len(domain.CatalogoPermisos) {
		t.Fatalf("el propietario recibió %d permisos, se esperaban %d", len(permisos), len(domain.CatalogoPermisos))
	}

	// Ana opera su negocio con normalidad.
	sucursalAna, err := sucursalService.Crear(ctx, ana, negocioAna.ID, domain.CrearSucursalInput{
		Codigo: "SUC-001", Nombre: "Matriz Ana",
	})
	if err != nil {
		t.Fatalf("crear sucursal: %v", err)
	}
	correo := "empleada-" + sufijo + "@tienda.mx"
	empleadaAna, err := empleadoService.Crear(ctx, ana, negocioAna.ID, domain.CrearEmpleadoInput{
		Nombre: "Carla", PrimerApellido: "Ruiz", Correo: &correo,
	})
	if err != nil {
		t.Fatalf("crear empleado: %v", err)
	}
	if _, err := asignacionService.Asignar(ctx, ana, negocioAna.ID, empleadaAna.ID,
		domain.AsignarSucursalInput{SucursalID: sucursalAna.ID}); err != nil {
		t.Fatalf("asignar sucursal: %v", err)
	}

	// Beto no debe poder ver ni tocar nada del negocio de Ana.
	t.Run("negocio ajeno es indistinguible de inexistente", func(t *testing.T) {
		if _, err := negocioService.Obtener(ctx, beto, negocioAna.ID); err == nil {
			t.Fatal("Beto no debe poder leer el negocio de Ana")
		}
		if _, err := sucursalService.Listar(ctx, beto, negocioAna.ID, "activo", ""); err == nil {
			t.Fatal("Beto no debe poder listar sucursales de Ana")
		}
		if _, err := empleadoService.Listar(ctx, beto, negocioAna.ID, "", ""); err == nil {
			t.Fatal("Beto no debe poder listar empleados de Ana")
		}
		if _, err := rolService.Listar(ctx, beto, negocioAna.ID, false); err == nil {
			t.Fatal("Beto no debe poder listar roles de Ana")
		}
	})

	t.Run("no se cruzan entidades entre negocios", func(t *testing.T) {
		// La sucursal de Ana no existe dentro del negocio de Beto.
		if _, err := sucursalService.Obtener(ctx, beto, negocioBeto.ID, sucursalAna.ID); err == nil {
			t.Fatal("una sucursal de otro negocio no debe encontrarse")
		}
		// El empleado de Ana tampoco.
		if _, err := empleadoService.Obtener(ctx, beto, negocioBeto.ID, empleadaAna.ID); err == nil {
			t.Fatal("un empleado de otro negocio no debe encontrarse")
		}
		// Y una sucursal ajena no puede asignarse a un empleado propio.
		empleadoBeto, err := empleadoService.Crear(ctx, beto, negocioBeto.ID, domain.CrearEmpleadoInput{
			Nombre: "Dario", PrimerApellido: "Sosa",
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = asignacionService.Asignar(ctx, beto, negocioBeto.ID, empleadoBeto.ID,
			domain.AsignarSucursalInput{SucursalID: sucursalAna.ID})
		if err == nil {
			t.Fatal("no debe poder asignarse una sucursal de otro negocio")
		}
	})

	t.Run("un miembro sin roles no puede mutar", func(t *testing.T) {
		// Beto entra al negocio de Ana como miembro, pero sin ningún rol asignado.
		membresia := domain.MembresiaNegocio{
			NegocioID: negocioAna.ID, UsuarioID: beto, TipoMiembro: "miembro", Estado: "activo",
		}
		if err := db.Create(&membresia).Error; err != nil {
			t.Fatal(err)
		}

		// Ahora sí puede leer, porque la lectura solo exige membresía activa.
		if _, err := sucursalService.Listar(ctx, beto, negocioAna.ID, "activo", ""); err != nil {
			t.Fatalf("un miembro activo debe poder leer: %v", err)
		}
		// Pero no puede escribir en ningún módulo de la capability.
		if _, err := sucursalService.Crear(ctx, beto, negocioAna.ID, domain.CrearSucursalInput{
			Codigo: "SUC-999", Nombre: "Intrusa",
		}); err == nil {
			t.Fatal("un miembro sin permisos no debe crear sucursales")
		}
		if _, err := empleadoService.Crear(ctx, beto, negocioAna.ID, domain.CrearEmpleadoInput{
			Nombre: "Eva", PrimerApellido: "Mora",
		}); err == nil {
			t.Fatal("un miembro sin permisos no debe crear empleados")
		}
		if _, err := negocioService.Actualizar(ctx, beto, negocioAna.ID, domain.ActualizarNegocioInput{}); err == nil {
			t.Fatal("un miembro sin permisos no debe editar el negocio")
		}

		// Con un rol que concede el permiso, la misma operación sí procede.
		rol, err := rolService.Crear(ctx, ana, negocioAna.ID, domain.CrearRolInput{
			Codigo: "SUPERVISOR", Nombre: "Supervisor",
			Permisos: idsDePermisos(t, db, domain.PermisoSucursalCrear, domain.PermisoSucursalVer),
		})
		if err != nil {
			t.Fatalf("crear rol: %v", err)
		}
		if err := rolService.AsignarRoles(ctx, ana, negocioAna.ID, membresia.ID, []uuid.UUID{rol.ID}); err != nil {
			t.Fatalf("asignar rol: %v", err)
		}
		if _, err := sucursalService.Crear(ctx, beto, negocioAna.ID, domain.CrearSucursalInput{
			Codigo: "SUC-002", Nombre: "Norte Ana",
		}); err != nil {
			t.Fatalf("con el permiso concedido debe poder crear: %v", err)
		}
		// El permiso concedido no alcanza para otros módulos.
		if _, err := empleadoService.Crear(ctx, beto, negocioAna.ID, domain.CrearEmpleadoInput{
			Nombre: "Eva", PrimerApellido: "Mora",
		}); err == nil {
			t.Fatal("el permiso de sucursales no debe habilitar el módulo de equipo")
		}
	})

	t.Run("el negocio conserva un propietario efectivo", func(t *testing.T) {
		var membresiaAna uuid.UUID
		var texto string
		err := db.Raw(`SELECT id::text FROM membresias_negocio WHERE negocio_id = ? AND usuario_id = ?`,
			negocioAna.ID, ana).Scan(&texto).Error
		if err != nil {
			t.Fatal(err)
		}
		membresiaAna = uuid.MustParse(texto)

		// Quitarle el rol de sistema al único propietario debe rechazarse.
		if err := rolService.AsignarRoles(ctx, ana, negocioAna.ID, membresiaAna, []uuid.UUID{}); err == nil {
			t.Fatal("el negocio no puede quedarse sin propietario efectivo")
		}
	})

	t.Run("base.MD: columnas retiradas no reaparecen", func(t *testing.T) {
		retiradas := []struct{ tabla, columna string }{
			{"membresias_negocio", "rol_id"},
			{"membresias_negocio", "empleado_id"},
			{"empleados", "perfil_id"},
		}
		for _, retirada := range retiradas {
			var total int
			err := db.Raw(`SELECT count(*) FROM information_schema.columns
				WHERE table_name = ? AND column_name = ?`, retirada.tabla, retirada.columna).Scan(&total).Error
			if err != nil || total != 0 {
				t.Fatalf("%s.%s no debía existir: %d %v", retirada.tabla, retirada.columna, total, err)
			}
		}
	})
}

func crearUsuario(t *testing.T, db *gorm.DB, correo string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.Exec(`INSERT INTO usuarios (id, correo, hash_contrasena, estado, creado_en, actualizado_en)
		VALUES (?, ?, 'hash-de-prueba', 'activo', now(), now())`, id, correo).Error
	if err != nil {
		t.Fatalf("crear usuario %s: %v", correo, err)
	}
	return id
}

func idsDePermisos(t *testing.T, db *gorm.DB, codigos ...string) []uuid.UUID {
	t.Helper()
	textos := make([]string, 0, len(codigos))
	if err := db.Raw(`SELECT id::text FROM permisos WHERE codigo IN ?`, codigos).Scan(&textos).Error; err != nil {
		t.Fatal(err)
	}
	ids := make([]uuid.UUID, 0, len(textos))
	for _, texto := range textos {
		ids = append(ids, uuid.MustParse(texto))
	}
	return ids
}
