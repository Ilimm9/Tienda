package database

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	app "tienda/backend/internal/application/negocio"
	catalogo "tienda/backend/internal/domain"
	domain "tienda/backend/internal/domain/negocio"
	"tienda/backend/internal/infrastructure"
	infra "tienda/backend/internal/infrastructure/negocio"

	"github.com/google/uuid"
)

func TestInvitacionesPermisosSucursal(t *testing.T) {
	url := os.Getenv("CAPABILITY20_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("CAPABILITY20_TEST_DATABASE_URL no configurada")
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	var nombre string
	if err := db.Raw("SELECT current_database()").Scan(&nombre).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(nombre, "_test") {
		t.Fatal("se requiere una base aislada con sufijo _test")
	}
	if err := Init(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	propietario := crearUsuario(t, db, uuid.NewString()+"@tienda.local")
	usuario := crearUsuario(t, db, uuid.NewString()+"@tienda.local")
	negocio, err := app.NewNegocioService(infra.NewNegocioRepository(db)).Crear(ctx, propietario, domain.CrearNegocioInput{NombreComercial: "QA invitaciones " + uuid.NewString()[:8]})
	if err != nil {
		t.Fatal(err)
	}
	sucursales := app.NewSucursalService(infra.NewSucursalRepository(db))
	a, err := sucursales.Crear(ctx, propietario, negocio.ID, domain.CrearSucursalInput{Codigo: "AA", Nombre: "Centro"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := sucursales.Crear(ctx, propietario, negocio.ID, domain.CrearSucursalInput{Codigo: "BB", Nombre: "Norte"})
	if err != nil {
		t.Fatal(err)
	}
	roles := app.NewRolService(infra.NewRolRepository(db))
	var permiso domain.Permiso
	if err := db.First(&permiso, "codigo = ?", domain.PermisoPrecioAutorizar).Error; err != nil {
		t.Fatal(err)
	}
	rol, err := roles.Crear(ctx, propietario, negocio.ID, domain.CrearRolInput{Codigo: "AUTORIZADOR", Nombre: "Autorizar precios", Permisos: []uuid.UUID{permiso.ID}})
	if err != nil {
		t.Fatal(err)
	}
	var lectura domain.Permiso
	if err := db.First(&lectura, "codigo = ?", domain.PermisoPrecioVer).Error; err != nil {
		t.Fatal(err)
	}
	if len(rol.Permisos) != 2 {
		t.Fatalf("faltan lecturas: %+v", rol.Permisos)
	}
	empleados := app.NewEmpleadoService(infra.NewEmpleadoRepository(db))
	var correo string
	if err := db.Table("usuarios").Where("id = ?", usuario).Pluck("correo", &correo).Error; err != nil {
		t.Fatal(err)
	}
	empleado, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Invitado", PrimerApellido: "QA", Correo: &correo})
	if err != nil {
		t.Fatal(err)
	}
	invitaciones := app.NewInvitacionService(infra.NewInvitacionRepository(db), nil, nil)
	creada, err := invitaciones.Crear(ctx, propietario, negocio.ID, domain.CrearInvitacionInput{EmpleadoID: empleado.ID, SucursalID: a.ID, RolPredeterminadoID: rol.ID})
	if err != nil {
		t.Fatal(err)
	}
	// Dos aceptaciones simultáneas deben producir un solo vínculo y un solo éxito.
	var wg sync.WaitGroup
	resultados := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := invitaciones.Aceptar(ctx, usuario, creada.Token); resultados <- err }()
	}
	wg.Wait()
	close(resultados)
	exitos := 0
	for err := range resultados {
		if err == nil {
			exitos++
		} else if !errors.Is(err, app.ErrInvitacionYaAceptada) {
			t.Fatal(err)
		}
	}
	if exitos != 1 {
		t.Fatalf("aceptaciones exitosas=%d", exitos)
	}
	contexto := infra.NewContextoRepository(db)
	opciones, err := contexto.ListarOpcionesContexto(ctx, usuario)
	if err != nil || len(opciones) != 1 || len(opciones[0].Sucursales) != 1 || len(opciones[0].Permisos) != 2 {
		t.Fatalf("contexto incorrecto: %+v %v", opciones, err)
	}
	estado := domain.EstadoEmpleadoSuspendido
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, empleado.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err != nil {
		t.Fatal(err)
	}
	opciones, err = contexto.ListarOpcionesContexto(ctx, usuario)
	if err != nil || len(opciones) != 0 {
		t.Fatalf("suspendido conserva negocio: %+v %v", opciones, err)
	}
	permisos, err := roles.PermisosEfectivos(ctx, usuario, negocio.ID)
	if err != nil || len(permisos) != 0 {
		t.Fatalf("suspendido conserva permisos: %v %v", permisos, err)
	}
	estado = domain.EstadoEmpleadoActivo
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, empleado.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err != nil {
		t.Fatal(err)
	}
	opciones, _ = contexto.ListarOpcionesContexto(ctx, usuario)
	if len(opciones) != 1 || len(opciones[0].Sucursales) != 1 {
		t.Fatal("suspensión no recuperó asignación vigente")
	}
	estado = domain.EstadoEmpleadoTerminado
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, empleado.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err != nil {
		t.Fatal(err)
	}
	estado = domain.EstadoEmpleadoActivo
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, empleado.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err == nil {
		t.Fatal("baja se reactivó sin reasignación")
	}
	if _, err := empleados.Reactivar(ctx, propietario, negocio.ID, empleado.ID, domain.ReactivarEmpleadoInput{Roles: []uuid.UUID{rol.ID}, Sucursales: []uuid.UUID{b.ID}}); err != nil {
		t.Fatal(err)
	}
	opciones, _ = contexto.ListarOpcionesContexto(ctx, usuario)
	if len(opciones) != 1 || len(opciones[0].Sucursales) != 1 || opciones[0].Sucursales[0].ID != b.ID {
		t.Fatalf("reactivación restauró sucursal anterior: %+v", opciones)
	}
	// Una baja laboral no debe recuperar una membresía revocada por otra causa.
	estado = domain.EstadoEmpleadoTerminado
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, empleado.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err != nil {
		t.Fatal(err)
	}
	if err := db.Table("membresias_negocio").Where("negocio_id = ? AND usuario_id = ?", negocio.ID, usuario).Update("estado", "suspendido").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := empleados.Reactivar(ctx, propietario, negocio.ID, empleado.ID, domain.ReactivarEmpleadoInput{Roles: []uuid.UUID{rol.ID}, Sucursales: []uuid.UUID{b.ID}}); !errors.Is(err, app.ErrInvitacionMembresiaInactiva) {
		t.Fatalf("restauró membresía revocada: %v", err)
	}
	otroNegocio, err := app.NewNegocioService(infra.NewNegocioRepository(db)).Crear(ctx, usuario, domain.CrearNegocioInput{NombreComercial: "Otro negocio " + uuid.NewString()[:8]})
	if err != nil {
		t.Fatal(err)
	}
	opciones, err = contexto.ListarOpcionesContexto(ctx, usuario)
	if err != nil || len(opciones) != 1 || opciones[0].ID != otroNegocio.ID {
		t.Fatalf("baja afectó otro negocio: %+v %v", opciones, err)
	}
	// Una invitación pendiente no puede elevarse reenviando con un emisor sin ese rol.
	bajo := crearUsuario(t, db, uuid.NewString()+"@tienda.local")
	membresia := domain.MembresiaNegocio{NegocioID: negocio.ID, UsuarioID: bajo, TipoMiembro: "miembro", Estado: "activo"}
	if err := db.Create(&membresia).Error; err != nil {
		t.Fatal(err)
	}
	var permisosEnvio []uuid.UUID
	if err := db.Model(&domain.Permiso{}).Where("codigo IN ?", []string{domain.PermisoInvitacionEnviar, domain.PermisoEmpleadoGestionar}).Pluck("id", &permisosEnvio).Error; err != nil {
		t.Fatal(err)
	}
	emisorRol, err := roles.Crear(ctx, propietario, negocio.ID, domain.CrearRolInput{Codigo: "EMISOR", Nombre: "Emisor", Permisos: permisosEnvio})
	if err != nil {
		t.Fatal(err)
	}
	if err := roles.AsignarRoles(ctx, propietario, negocio.ID, membresia.ID, []uuid.UUID{emisorRol.ID}); err != nil {
		t.Fatal(err)
	}
	correo2 := uuid.NewString() + "@tienda.local"
	otro, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Otro", PrimerApellido: "QA", Correo: &correo2})
	if err != nil {
		t.Fatal(err)
	}
	pendiente, err := invitaciones.Crear(ctx, propietario, negocio.ID, domain.CrearInvitacionInput{EmpleadoID: otro.ID, SucursalID: a.ID, RolPredeterminadoID: rol.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invitaciones.Reenviar(ctx, bajo, negocio.ID, pendiente.Invitacion.ID, domain.ReenviarInvitacionInput{Correo: &correo2}); !errors.Is(err, app.ErrRolNoDelegable) {
		t.Fatalf("reemisión no delegable: %v", err)
	}
	var original domain.InvitacionNegocio
	if err := db.First(&original, "id = ?", pendiente.Invitacion.ID).Error; err != nil {
		t.Fatal(err)
	}
	if original.Estado != domain.EstadoInvitacionPendiente {
		t.Fatal("rechazo modificó el enlace anterior")
	}
	// El diagnóstico de roles antiguos es de solo lectura.
	if err := db.Where("rol_id = ? AND permiso_id = ?", rol.ID, lectura.ID).Delete(&domain.PermisoRol{}).Error; err != nil {
		t.Fatal(err)
	}
	listado, err := roles.Listar(ctx, propietario, negocio.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	hallado := false
	for _, r := range listado {
		if r.ID == rol.ID {
			hallado = len(r.PermisosFaltantes) == 1 && r.PermisosFaltantes[0] == domain.PermisoPrecioVer
		}
	}
	if !hallado {
		t.Fatal("no se reportó la dependencia faltante")
	}
	seleccion := []uuid.UUID{permiso.ID, lectura.ID}
	cambio := domain.ActualizarRolInput{Permisos: domain.Optional[[]uuid.UUID]{Set: true, Value: &seleccion}}
	if _, err := roles.Actualizar(ctx, propietario, negocio.ID, rol.ID, cambio); err == nil {
		t.Fatal("amplió un rol antiguo sin autorización específica")
	}
	cambio.AutorizarDependenciasExistentes = true
	if _, err := roles.Actualizar(ctx, propietario, negocio.ID, rol.ID, cambio); err != nil {
		t.Fatal(err)
	}
	// El rol puede superar al emisor después de enviar: aceptar revalida la delegación.
	var permisoBasico domain.Permiso
	if err := db.First(&permisoBasico, "codigo = ?", domain.PermisoInvitacionVer).Error; err != nil {
		t.Fatal(err)
	}
	rolCambiante, err := roles.Crear(ctx, propietario, negocio.ID, domain.CrearRolInput{Codigo: "CAMBIANTE", Nombre: "Rol cambiante", Permisos: []uuid.UUID{permisoBasico.ID}})
	if err != nil {
		t.Fatal(err)
	}
	correoCambio := uuid.NewString() + "@tienda.local"
	usuarioCambio := crearUsuario(t, db, correoCambio)
	empleadoCambio, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Cambio", PrimerApellido: "QA", Correo: &correoCambio})
	if err != nil {
		t.Fatal(err)
	}
	invitacionCambio, err := invitaciones.Crear(ctx, bajo, negocio.ID, domain.CrearInvitacionInput{EmpleadoID: empleadoCambio.ID, SucursalID: a.ID, RolPredeterminadoID: rolCambiante.ID})
	if err != nil {
		t.Fatal(err)
	}
	seleccion = []uuid.UUID{permiso.ID}
	if _, err := roles.Actualizar(ctx, propietario, negocio.ID, rolCambiante.ID, domain.ActualizarRolInput{Permisos: domain.Optional[[]uuid.UUID]{Set: true, Value: &seleccion}}); err != nil {
		t.Fatal(err)
	}
	if _, err := invitaciones.Aceptar(ctx, usuarioCambio, invitacionCambio.Token); !errors.Is(err, app.ErrRolNoDelegable) {
		t.Fatalf("aceptó rol ampliado: %v", err)
	}
	// Aceptar contra reenviar: una sola operación gana sobre el enlace original.
	correoCarrera := uuid.NewString() + "@tienda.local"
	usuarioCarrera := crearUsuario(t, db, correoCarrera)
	empleadoCarrera, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Carrera", PrimerApellido: "QA", Correo: &correoCarrera})
	if err != nil {
		t.Fatal(err)
	}
	invitacionCarrera, err := invitaciones.Crear(ctx, propietario, negocio.ID, domain.CrearInvitacionInput{EmpleadoID: empleadoCarrera.ID, SucursalID: a.ID, RolPredeterminadoID: rol.ID})
	if err != nil {
		t.Fatal(err)
	}
	resultados = make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := invitaciones.Aceptar(ctx, usuarioCarrera, invitacionCarrera.Token)
		resultados <- err
	}()
	go func() {
		defer wg.Done()
		_, err := invitaciones.Reenviar(ctx, propietario, negocio.ID, invitacionCarrera.Invitacion.ID, domain.ReenviarInvitacionInput{})
		resultados <- err
	}()
	wg.Wait()
	close(resultados)
	exitos = 0
	for err := range resultados {
		if err == nil {
			exitos++
		} else if !errors.Is(err, app.ErrInvitacionYaAceptada) && !errors.Is(err, app.ErrInvitacionNoVigente) {
			t.Fatalf("carrera reenvío: %v", err)
		}
	}
	if exitos != 1 {
		t.Fatalf("carrera produjo %d éxitos", exitos)
	}
	// Stock y variantes: ni la suma ni el detalle de otra sucursal pueden filtrarse.
	productos := infrastructure.NewProductRepository(db)
	if err := productos.Create(negocio.ID, catalogo.CreateProductInput{Nombre: "Producto QA", SucursalID: a.ID, PrecioVenta: 10, StockInicial: 3}); err != nil {
		t.Fatal(err)
	}
	lista, err := productos.ListByBusiness(negocio.ID, a.ID)
	if err != nil || len(lista) != 1 || lista[0].Stock != 3 {
		t.Fatalf("stock A: %+v %v", lista, err)
	}
	lista, err = productos.ListByBusiness(negocio.ID, b.ID)
	if err != nil || len(lista) != 1 || lista[0].Stock != 0 || len(lista[0].Inventario) != 0 {
		t.Fatalf("B filtró stock A: %+v %v", lista, err)
	}
	if err := productos.Create(negocio.ID, catalogo.CreateProductInput{Nombre: "Variante QA", SucursalID: a.ID, Variantes: []catalogo.CreateProductVariantInput{{Atributos: []catalogo.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "M"}}, GenerarSKUInterno: true, PrecioVenta: 20, StockInicial: 8}, {Atributos: []catalogo.ProductVariantAttributeInput{{Nombre: "Talla", Valor: "G"}}, GenerarSKUInterno: true, PrecioVenta: 20, StockInicial: 5}}}); err != nil {
		t.Fatal(err)
	}
	for _, sucursal := range []domain.SucursalDetalle{a, b} {
		lista, err := productos.ListByBusiness(negocio.ID, sucursal.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, producto := range lista {
			if producto.Nombre != "Variante QA" {
				continue
			}
			esperado := float64(0)
			if sucursal.ID == a.ID {
				esperado = 13
			}
			if len(producto.Variantes) != 2 || producto.Stock != esperado || producto.Variantes[0].Stock+producto.Variantes[1].Stock != esperado {
				t.Fatalf("stock variante: %+v", producto)
			}
			if esperado == 0 && (producto.Estado != "Agotado" || len(producto.Inventario) != 0) {
				t.Fatalf("estado sin stock: %+v", producto)
			}
			if esperado > 0 && (producto.Estado != "En stock" || len(producto.Inventario) != 1 || producto.Inventario[0].Stock != esperado) {
				t.Fatalf("total por sucursal: %+v", producto)
			}
			for _, inventario := range producto.Variantes[0].Inventario {
				if inventario.SucursalID != sucursal.ID.String() {
					t.Fatal("filtró inventario de otra sucursal")
				}
			}
		}
	}
	// Una ficha sin cuenta puede prepararse para una nueva invitación sin otorgar acceso.
	sinCuenta, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Sin cuenta", PrimerApellido: "QA"})
	if err != nil {
		t.Fatal(err)
	}
	estado = domain.EstadoEmpleadoTerminado
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, sinCuenta.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); err != nil {
		t.Fatal(err)
	}
	estado = domain.EstadoEmpleadoPendiente
	preparado, err := empleados.Actualizar(ctx, propietario, negocio.ID, sinCuenta.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}})
	if err != nil || preparado.MembresiaID != nil {
		t.Fatalf("preparar sin cuenta: %+v %v", preparado, err)
	}
	// El último propietario efectivo tampoco puede ser suspendido desde su ficha laboral.
	fichaPropietario, err := empleados.Crear(ctx, propietario, negocio.ID, domain.CrearEmpleadoInput{Nombre: "Propietario", PrimerApellido: "QA"})
	if err != nil {
		t.Fatal(err)
	}
	var membresiaPropietario domain.MembresiaNegocio
	if err := db.Where("negocio_id = ? AND usuario_id = ?", negocio.ID, propietario).Take(&membresiaPropietario).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("empleados").Where("id = ?", fichaPropietario.ID).Updates(map[string]any{"membresia_id": membresiaPropietario.ID, "estado": domain.EstadoEmpleadoActivo}).Error; err != nil {
		t.Fatal(err)
	}
	estado = domain.EstadoEmpleadoSuspendido
	if _, err := empleados.Actualizar(ctx, propietario, negocio.ID, fichaPropietario.ID, domain.ActualizarEmpleadoInput{Estado: domain.Optional[string]{Set: true, Value: &estado}}); !errors.Is(err, app.ErrPropietarioSinRol) {
		t.Fatalf("permitió suspender al último propietario: %v", err)
	}
	// Revocar al emisor después de emitir también bloquea la aceptación.
	usuario2 := crearUsuario(t, db, correo2)
	if err := db.Table("membresias_negocio").Where("negocio_id = ? AND usuario_id = ?", negocio.ID, propietario).Update("estado", "suspendido").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := invitaciones.Aceptar(ctx, usuario2, pendiente.Token); !errors.Is(err, app.ErrInvitacionProhibida) {
		t.Fatalf("se aceptó con emisor suspendido: %v", err)
	}
	var cuenta int64
	if err := db.Table("membresias_negocio").Where("negocio_id = ? AND usuario_id = ?", negocio.ID, usuario2).Count(&cuenta).Error; err != nil {
		t.Fatal(err)
	}
	if cuenta != 0 {
		t.Fatal("rechazo dejó membresía parcial")
	}
}
