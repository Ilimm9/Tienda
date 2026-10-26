# Capability 35: Control general de cuentas y tiendas en la administración de plataforma

- Estado: **En implementación**.
- Aprobación explícita: 2026-10-26, mediante la aprobación del plan en Claude Code. Alcance pedido por el usuario: «solo el listado global de cuentas, fechas de registro, las tiendas de igual manera su fecha de creación con cuántos productos y empleados cuenta cada una, de manera general».
- Fecha: 2026-10-26.
- Base: [15 · Administración de plataforma](historial/15-administracion-plataforma-metricas.md). Dominio `plataforma` en `domain`, `application`, `infrastructure` e `interfaces/http`; frontend `features/plataforma`.

## Problema

El administrador de plataforma (`ADMIN_LECTURA`) solo puede ver los usuarios dentro de cada negocio. No existe un listado global de cuentas con su fecha de registro. El listado de tiendas tampoco muestra la fecha de creación ni el número de productos.

## Alcance

1. **Listado global de cuentas** (solo lectura). Muestra nombre, correo, estado, si el correo está verificado, fecha de registro, último inicio de sesión y las tiendas a las que pertenece la cuenta (membresías activas).
2. **Tiendas**: agregar el número de productos activos y mostrar la fecha de creación y el número de empleados activos (`total_personal`, que ya existe) en el listado y en el detalle.

Quedan fuera: acciones sobre cuentas o tiendas, exportaciones, métricas nuevas de actividad y datos sensibles (hash, intentos fallidos, bloqueos).

## Reglas

- El acceso exige `ADMIN_LECTURA` mediante el grupo `/api/v1/plataforma` existente (`RequireAdminLectura`).
- El listado de cuentas excluye las cuentas con rol de plataforma.
- Filtro de estado de cuenta:
  - `activo`: estado activo, correo verificado y no deshabilitada;
  - `sin_verificar`: sin `correo_verificado_en`;
  - `deshabilitado`: `deshabilitado_en` presente o estado distinto de activo.
  - Cualquier otro valor se ignora.
- La búsqueda no distingue mayúsculas y compara contra el correo y el nombre del perfil.
- Paginación igual que en negocios: página mínima 1; tamaño por omisión 20, máximo 100. Orden: registro más reciente primero.
- Productos: `productos.activo = TRUE` del negocio.

## Contratos

- `GET /api/v1/plataforma/cuentas?q=&estado=&page=&page_size=`, que responde:
  `{ items: [{ id, nombre, correo, estado, correo_verificado, creado_en, ultimo_inicio_sesion_en, total_negocios, negocios: [string] }], total, page, page_size }`.
- `GET /api/v1/plataforma/negocios` y `GET /api/v1/plataforma/negocios/:id` agregan `total_productos`.

## Pantallas

- `/administracion-plataforma/cuentas`: búsqueda, filtro de estado, tabla (Cuenta, Estado, Verificada, Registro, Último acceso, Tiendas) y paginación. Se agrega el enlace «Cuentas» en la barra del shell.
- `/administracion-plataforma/negocios`: nuevas columnas «Creada» y «Productos»; «Personal» se renombra a «Empleados».
- Detalle de tienda: nueva tarjeta «Productos activos».

## Pruebas y criterios de aceptación

- Go: normalización de filtros y paginación en `ListarCuentas`; compilación y pruebas del dominio `plataforma`.
- Angular: `PlataformaService.cuentas` envía filtros y paginación a `/plataforma/cuentas`.
- Manual: un administrador ve todas las cuentas con su fecha de registro, y las tiendas con fecha de creación, productos y empleados. Una cuenta sin rol de plataforma recibe 403.

## Tareas

- [x] Capability e índices.
- [x] Backend: dominio, repositorio, servicio, handler y ruta. Además, `normalizarPagina` se comparte entre negocios y cuentas.
- [x] Frontend: modelos, servicio, pantalla de cuentas, columnas de tiendas y detalle.
- [x] Pruebas y verificación automatizada.
- [ ] Prueba manual en navegador y aceptación final del usuario.

## Decisiones

- La tabla de perfiles es `perfil_usuarios` (nombre por omisión de GORM).
- No se agregó una prueba de middleware específica para `/cuentas`: la ruta está en el mismo grupo `RequireAdminLectura` que ya cubre `middleware_test.go`.

## Resultados de verificación (2026-10-26)

- `go build ./...` y `go vet` de `plataforma` y `cmd`: correctos.
- `go test ./internal/application/plataforma/... ./internal/interfaces/http/plataforma/...`: correctos. Incluye la prueba nueva `TestListarCuentasNormalizaFiltrosYPaginacion`.
- Consulta de solo lectura contra el PostgreSQL local (programa temporal, ya eliminado):
  - `ListarCuentas` con los 4 filtros y con búsqueda: correcto (2 cuentas verificadas, 0 sin verificar, 0 deshabilitadas).
  - `ListarNegocios`: correcto, con `total_productos=1653` y `total_personal=1`.
- Angular (Node 22.23.2): `ng test --include=src/app/features/plataforma/**` pasa 6 de 6 pruebas. `ng build --configuration development` compila correctamente.

## Correcciones autorizadas a la capability 15 (2026-10-26)

El usuario autorizó corregirlas («si, corrigelas»). Contra la base local, el detalle de tienda del panel de plataforma fallaba en dos consultas existentes:

1. `ObtenerNegocio` fallaba con `invalid field found for struct NegocioDetalle's field Sucursales`. **Corrección:** `gorm:"-"` en `Sucursales` y `Actividad` de `NegocioDetalle`, porque se cargan con consultas aparte.
2. `ListarUsuarios` fallaba con `relation "perfiles_usuario" does not exist`. **Corrección:** se usa la tabla `perfil_usuarios`.

Verificación contra PostgreSQL local (solo lectura): el detalle devuelve 1653 productos, 1 empleado, 2 sucursales y 4 días de actividad; la lista de usuarios devuelve 2 usuarios. No hubo errores. Las pruebas de `plataforma` y `domain` pasan.
