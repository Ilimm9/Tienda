# Capability 31: Solo el propietario crea productos

- Estado: **Cerrada** el 2026-10-23 con aceptación final del usuario.
- Aprobación explícita: 2026-10-22, el usuario eligió atender la observación 3 de la 26, decidió el alcance y aprobó el plan.
- Fecha: 2026-10-22.
- Origen: observación 3 de la [26](26-observaciones-ilian.md): crear productos con precio y costo no exige `precios.autorizar`.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Backend en `interfaces/http` de productos y `cmd/api`; frontend `features/productos`. Sin cambios de esquema ni del dominio `negocio`.
- ID: se usa el 31 porque otra sesión usó el 30 («acceso por persona») aunque su archivo ya no está en el árbol.

## Causa (verificada en código)

- `POST /catalogo/productos` y las rutas de carga masiva (`cmd/api/main.go`) solo exigen `catalogo.gestionar`.
- Crear e importar escriben el precio y el costo iniciales en `precios_sucursal` (`createInitialBranchPrice`). Cualquier colaborador con permiso de catálogo los fija sin `precios.autorizar`.
- La edición ya quedó cubierta por la [29](29-guardar-precio-al-editar.md).

## Alcance y reglas

- **Agregar nuevo producto** y **Carga masiva** son solo del **propietario** del negocio (`membresias_negocio.tipo_miembro = 'propietario'`), sin importar los permisos de su rol.
- Backend: crear, validar importación, descargar errores de importación, importar y consultar el estado de la importación responden `403` con código `SOLO_PROPIETARIOS` («Solo el propietario puede crear productos.») a quien no es propietario.
- Frontend: los botones solo se muestran al propietario; las rutas `/productos/nuevo` y `/productos/importar` llevan a «Sin acceso» a los demás.
- Sin cambios: editar productos (reglas de la 29), descargar la plantilla, ajustar inventario y «Crear nuevo…» en Registrar compra, que crea el producto sin precio.
- Como el propietario puede todo, los campos de precio y costo al crear no se bloquean.

## Fuera de alcance
- Un permiso nuevo «Crear productos» (descartado por el usuario).
- Exigir `inventario.ajustar` para el stock inicial.
- «Crear nuevo…» en Compras.

## Relación con la capability 19
- Cierra la parte de **creación** de la tarea abierta de la 19 («exigir `precios.autorizar` cuando envíen montos»): solo el propietario crea productos con montos. Queda anotado en la 19.

## Contratos
- `POST …/catalogo/productos`, `POST …/catalogo/productos/validar-importacion`, `POST …/catalogo/productos/importacion/errores`, `POST …/catalogo/productos/importar` y `GET …/catalogo/productos/importaciones/:importacionId`: `403 {"codigo":"SOLO_PROPIETARIOS"}` si quien llama no es propietario.

## Pruebas y criterios de aceptación
- Un miembro con `catalogo.gestionar` recibe `403 SOLO_PROPIETARIOS` al crear y al importar; el propietario crea normalmente; el miembro sigue editando.
- Los botones se ocultan para un miembro; las rutas de crear e importar lo llevan a «Sin acceso».
- `go test ./internal/...`, `ng test` y `ng build` en verde (salvo fallas previas registradas).

## Tareas
- [x] Aprobación del usuario (2026-10-22).
- [x] Backend: middleware `SoloPropietario` en las rutas de crear e importar.
- [x] Frontend: botones y guard de ruta.
- [x] Pruebas, registro de resultados y base `registro-productos.md`.
- [x] Prueba manual y aceptación final del usuario (2026-10-23).

## Decisiones
- 2026-10-22: el usuario decidió que el botón de nuevo producto sea solo del propietario, reconocido por su membresía y no por un permiso.
- 2026-10-22: la carga masiva también queda solo para el propietario; «Crear nuevo…» en Compras no cambia.
- 2026-10-23: el usuario confirmó que crear, importar y «Agregar del catálogo» ([28](28-surtido-por-sucursal.md)) sigan siendo solo del propietario, no de roles de administrador.
- 2026-10-22 (implementación): `ProductHandler.SoloPropietario` lee el tipo de miembro con `ContextoService.ListarOpciones`; no se modificó el dominio `negocio`. En el frontend, `soloPropietarioGuard` vive en `features/productos` y usa `ContextoService.negocio().tipo_miembro`.

## Pruebas ejecutadas (2026-10-22)

- `go build ./...` y `go vet ./internal/... ./cmd/...`: correctos.
- `TestSoloPropietarioCreaProductos` (`internal/interfaces/http/product_owner_test.go`): el propietario pasa; un miembro con `catalogo.gestionar` y `precios.autorizar`, el propietario de otro negocio y quien no tiene membresía reciben `403 SOLO_PROPIETARIOS`; sin sesión, `401`.
- `go test ./internal/...` con la base de integración: correcta salvo `TestProductionAcceptsStrongOTPSecret`, falla previa registrada.
- `ng test --include src/app/features/productos`: 47/47. Pruebas nuevas: el guard deja pasar al propietario y lleva a «Sin acceso» a un miembro o sin negocio; los botones se ven para el propietario y se ocultan para un miembro; con la lista vacía, el miembro ve «Pide al propietario que agregue productos.».
- `ng test` completo: 247/253; las 6 fallas son las previas de `alta-inicial.spec.ts` (`window.matchMedia`).
- `ng build`: correcto. `graphify update .`: correcto.
- La regla de rutas en `main.go` no tiene prueba automática de extremo a extremo; se cubre con la prueba manual.

## Pendientes

- Reiniciar el backend activo.
- Prueba manual:
  - como propietario, ver ambos botones, crear e importar;
  - como colaborador con permiso de catálogo: no ve los botones; `/productos/nuevo` y `/productos/importar` llevan a «Sin acceso»; editar productos y «Crear nuevo…» en Compras siguen funcionando.
- Aceptación final del usuario.

## Cierre (2026-10-23)

- Aceptación final del usuario: «acepto la 22, 23, 27, 28, 29 y 31». Cerrada y archivada en `historial/`.
