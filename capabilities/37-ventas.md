# Capability 37: Ventas

- Estado: **Verificada** (Fases 1–5 verificadas; Fase 6 cerrada por decisión del usuario; pendientes pruebas manuales, aceptación final y cobro con tarjeta/transferencia).
- Aprobación explícita: 2026-10-28, mediante aprobación del plan de Fase 0 sobre [PLAN-VENTAS.md](../PLAN-VENTAS.md). Autoriza **solo la Fase 1** (datos y lógica, sin interfaz). Cada fase siguiente requiere aprobación propia y las capturas del diseño del usuario (tableros 04b, 04, 01, 01b y menú lateral); no se diseñan pantallas nuevas.
- Fecha: 2026-10-28.
- Fase 2 aprobada explícitamente el 2026-10-28 («Sí, apruebo la Fase 2»), sobre la captura del tablero 04b · Nueva venta compartida por el usuario.
- Fase 3 aprobada explícitamente el 2026-10-28 («puedes continuar con la fase tres, deja pendientes cobros por transferencia»), aclarado como **solo efectivo**: tarjeta y transferencia quedan pendientes. La interfaz espera las capturas del usuario (ventana de cobro y venta terminada).
- Fase 4 aprobada explícitamente el 2026-10-28 («Sí, apruebo la Fase 4»), sobre las capturas del tablero 04 (lista y ticket) y con la política de cancelación propuesta («Mi propuesta completa»).
- Fase 5 aprobada explícitamente el 2026-10-28 («continua con la siguiente fase»). Botón del menú según las capturas compartidas; para Inicio no hay captura del tablero 01 y se reutilizan la bienvenida y las tarjetas «Ir directo a» existentes.
- Fase 6 marcada como completada por decisión explícita del usuario el 2026-10-28 («la fase 6 menciona el acceso a la ruta desde el inicio y bloqueo de rutas, pero eso ya está, puedes marcar como completada esa fase»). Sin cambios de código: lo existente cubre el bloqueo por URL y el menú filtrado por permisos.
- Organización conforme a [AGENTS.md](../AGENTS.md). Dominio nuevo `venta` en `domain`, `application`, `infrastructure` e `interfaces/http`. Toca además `negocio` (permisos y rol Cajero) y `database` (migraciones). Base de conocimiento nueva: [modulo-ventas.md](modulo-ventas.md); se actualiza [gestion-colaboradores.md](gestion-colaboradores.md).

## Alcance

Entra: buscar productos por código de barras o nombre, registrar una venta con descuento de inventario y folio consecutivo, cancelar una venta regresando piezas, consultar ventas, permisos de ventas y rol Cajero. Fuera: facturación, crédito, clientes, corte de caja, cajas, impresora física, granel, descuento de lotes y búsqueda por SKU.

## Reglas

- **Montos:** las tablas de ventas guardan centavos en `bigint`; el dominio calcula en `int64` (`venta.Centavos`). La API recibe y responde pesos con hasta 2 decimales, convertidos de forma exacta desde/hacia el texto JSON (sin flotantes). Precios y costos se leen de `precios_sucursal` con `round(x*100)::bigint`.
- **IVA:** 16 % incluido sobre el total del ticket: `subtotal = round(total ÷ 1.16)`, `iva = total − subtotal` ($64.00 → $55.17 + $8.83).
- **Búsqueda:** primero código exacto en `producto_codigos` (cualquier longitud). Si el texto son solo dígitos y tiene 6 o más, no se busca por nombre. Si no, por nombre sin acentos ni mayúsculas (todas las palabras deben aparecer). Solo productos activos con precio mayor a 0 en la sucursal; agotados aparecen con `agotado = true`. Máximo 20 resultados.
- **Registrar (una transacción):** las líneas repetidas se suman; se bloquean filas de `inventario_sucursal` ordenadas por `producto_negocio_id`; se revalida existencia (error `EXISTENCIA_INSUFICIENTE` con el nombre del producto, sin guardar nada); se toma foto de nombre, código, precio y costo; se descuenta stock con movimiento `VENTA_SALIDA` (cantidad negativa, `referencia` = folio); al final se reserva el folio `V-000001` consecutivo por negocio.
- **Cobro:** `EFECTIVO` exige recibido ≥ total y calcula cambio; `TARJETA` y `TRANSFERENCIA` aceptan referencia opcional (recibido = total, cambio 0).
- **Cancelar:** motivo obligatorio; solo ventas `PAGADA`; estado `CANCELADA`, usuario y fecha; piezas de vuelta con movimiento `VENTA_CANCELACION` positivo. Cancelar dos veces se rechaza.
- **Margen (Verificar costos):** `(precio − costo) ÷ precio`; `negativo` si < 0, `bajo` si < 20 %, `ok` en otro caso. Solo se muestra; `precios_sucursal.margen` no cambia.
- **Permisos:** `ventas.registrar` cobra, busca y ve solo sus ventas; `ventas.ver` ve todas las ventas de sus sucursales; `ventas.cancelar` (requiere `ventas.ver`); `ventas.ver_costos` (requiere `ventas.registrar`) recibe costo y margen. Sin `ventas.ver_costos` la API no envía costos. Propietario los recibe automáticamente.
- **Rol Cajero:** por defecto solo `ventas.registrar`. Migración única (`ventas_rol_cajero_v1`) en roles `CAJERO` existentes: quita `catalogo.ver`, `compras.ver` y `compras.registrar` y agrega `ventas.registrar`; otros permisos agregados por el usuario se conservan. Roles personalizados (p. ej. «vendedor») no se tocan.

## Contratos (prefijo `/api/v1/negocios/:negocioId`)

| Método y ruta | Permiso | Descripción |
| --- | --- | --- |
| `GET /ventas/productos?sucursal_id&q` | `ventas.registrar` | Coincidencias `{items:[{producto_negocio_id, nombre, codigo_barras, precio, existencia, agotado, costo?, margen?, nivel_margen?}], por_codigo}` |
| `POST /ventas` | `ventas.registrar` | `{sucursal_id, forma_pago, recibido?, referencia?, lineas:[{producto_negocio_id, cantidad}]}` → 201 venta con líneas. 409 `EXISTENCIA_INSUFICIENTE` / 400 `PAGO_INSUFICIENTE`, `PRODUCTO_NO_DISPONIBLE`, `VENTA_INVALIDA` |
| `GET /ventas?sucursal_id&folio&usuario_id&forma_pago&estado&desde&hasta` | `ventas.ver` o `ventas.registrar` (solo propias) | `{items, total}` sin líneas |
| `GET /ventas/:ventaId` | `ventas.ver` o `ventas.registrar` (solo propias) | Venta con líneas |
| `POST /ventas/:ventaId/cancelar` | `ventas.cancelar` | `{motivo}` → venta cancelada. 409 `VENTA_YA_CANCELADA` |

Ventas de sucursales no accesibles o de otro cajero responden 404, igual que una inexistente.

## Tablas

| Tabla | Uso |
| --- | --- |
| `ventas` | Negocio, sucursal, número y folio (único por negocio), usuario que cobró, subtotal/IVA/total/recibido/cambio en centavos, forma de pago, referencia, estado, datos de cancelación. |
| `venta_lineas` | Producto comercial, código y nombre al momento, cantidad, precio y costo unitario, importe (centavos). |
| `venta_folio_consecutivos` | Siguiente número de folio por negocio. |
| `movimientos_inventario` (existente) | Tipos nuevos `VENTA_SALIDA` y `VENTA_CANCELACION`. |

## Pruebas y criterios de aceptación (Fase 1)

- Unitarias: IVA, cambio y pago insuficiente, conversión exacta de pesos, margen y nivel, validaciones de entrada.
- Integración (`VENTAS_TEST_DATABASE_URL`, base `tienda_ventas_test`): código exacto y corto, nombre sin acentos, producto sin precio excluido, agotado marcado, registrar descuenta stock + movimiento + folio +1, líneas repetidas se suman, existencia insuficiente sin guardar nada, ventas concurrentes con folios distintos y consecutivos, cancelar regresa piezas y no se repite, otra sucursal/negocio rechazado, filtro de ventas propias, migración de rol Cajero.
- `go build ./...`, `go vet ./...`, `go test ./...`, `ng test` y `ng build`.

## Fase 2 · Pantalla de captura (tablero 04b)

Ruta `/ventas/nueva` (`ventas.registrar`). Se respeta la captura 04b; las vistas no compartidas (desplegable, «Verificar costos» activo, lista vacía, deshacer) siguen los componentes existentes del kit.

- Encabezado «VENTAS · {SUCURSAL}» (sin caja ni nombre de cajero), «Nueva venta», texto de ayuda y «Volver a ventas» (`/ventas`). Sin fichas «Prueba el lector».
- Campo con foco automático: Enter agrega (código exacto o la coincidencia resaltada), Escape limpia, flechas recorren el desplegable; el lector funciona como teclado. Al escribir texto se muestra el desplegable con precio y existencia; agotados aparecen deshabilitados.
- Código inexistente: aviso «No encontramos el código …» y la lista no cambia. Agotado: aviso de sin existencias, no se agrega. Escanear de nuevo suma 1 sin pasar la existencia.
- Tabla: #, producto con código, existencia (ámbar si `existencia_baja`: existencia ≤ mínimo de la sucursal), precio, cantidad − / + (+ deshabilitado en el tope), importe y eliminar con «Deshacer». La última fila agregada o modificada se resalta.
- «Verificar costos» solo con `ventas.ver_costos`: agrega costo, margen y utilidad; margen negativo en rojo con aviso y enlace a «Validar costos y precios», menor a 20 % en ámbar.
- Resumen: productos, piezas, subtotal, IVA (16 %), total (centavos enteros, misma fórmula que el backend). «Terminar venta» visible y deshabilitado hasta la Fase 3; «Cancelar y vaciar lista» pide confirmación.
- La lista se guarda en `localStorage` por negocio y sucursal y se recupera al recargar.
- Backend: la búsqueda agrega `existencia_baja` (existencia > 0 y ≤ `inventario_sucursal.stock_minimo`).
- Fuera: botón F2 y menú (Fase 5), cobro (Fase 3), historial (Fase 4).

Criterio de aceptación: capturar 10 productos solo con el lector, sin mouse. Pruebas: `ng test` del módulo (escaneo repetido, código inexistente, agotado, tope de +, eliminar y deshacer, totales $64.00 → $55.17/$8.83, costos ocultos sin permiso, borrador restaurado) y `ng build`.

## Fase 3 · Cobro y venta terminada (solo efectivo)

- «Terminar venta» abre la ventana de cobro en efectivo: recibido, billetes rápidos, cambio y «Falta» cuando no alcanza; «Confirmar cobro» bloqueado si recibido < total.
- Confirmar llama a `POST /ventas`; un `EXISTENCIA_INSUFICIENTE` o `PRODUCTO_NO_DISPONIBLE` se muestra nombrando el producto y la lista se conserva.
- Pantalla final con folio, total, recibido y cambio, «Imprimir ticket» (impresión del navegador) y «Nueva venta» (lista vacía y borrador borrado).
- Tarjeta y transferencia: pendientes (el backend ya las acepta; la interfaz no las ofrece en esta fase).
- Diseño: según las capturas que comparta el usuario; no se diseña sin ellas. Ventana de cobro recibida el 2026-10-28: «Total a cobrar · N piezas», total, forma de pago (Efectivo activo; Tarjeta y Transferencia visibles y deshabilitadas por ser pendientes), «Recibido», rápidos «Exacto» y los billetes de $20 a $1,000 que cubren el total (máximo tres), «Falta» en rojo o «Cambio», «Seguir capturando» y «Confirmar cobro».
- Venta terminada recibida el 2026-10-28: ícono de confirmación, «Venta {folio} cobrada», total, recuadro con forma de pago, piezas, recibido y «Cambio a entregar», nota «Las existencias de la sucursal ya se descontaron.», «Imprimir ticket», «Nueva venta» (recibe el foco para seguir con Enter) y «Ver ventas del día» (`/ventas`, marcador hasta la Fase 4).
- Ticket impreso (sin captura de diseño): solo al imprimir se muestra un ticket de 72 mm con negocio, sucursal, folio, fecha, líneas, subtotal, IVA, total, efectivo y cambio; una regla global en `styles.css` oculta el resto de la página al imprimir.
- Si la confirmación falla por existencias, la ventana muestra el mensaje con el producto, la lista se conserva y la línea toma la existencia disponible como tope (bajar la cantidad siempre se permite).

## Fase 4 · Historial y ticket (tablero 04)

- `/ventas`: rótulo con la sucursal activa, «Ventas», «Nueva venta» (con `ventas.registrar`), periodo Hoy/Semana/Mes (calendario del negocio, semana desde lunes) con fecha, tarjetas Vendido, Tickets, Ticket promedio y Efectivo en caja (solo ventas pagadas; efectivo = total cobrado en efectivo), búsqueda por folio o cajero, filtro por forma de pago y tabla Folio, Hora, Atendió, Artículos, Pago, Total (tachado si está cancelada), Estado y «Ver ticket». Pie «N tickets · M cancelados». Un cajero solo ve sus ventas.
- `/ventas/:ventaId`: «Ticket V-…», fecha y hora, «Atendió {nombre}» (sin caja), Volver, Reimprimir y Cancelar venta; líneas y resumen con estado, forma de pago, recibido, cambio y total. Una venta cancelada muestra motivo, quién y cuándo.
- Acceso: `ventas.ver` o `ventas.registrar`.
- **Cancelación:** solo con `ventas.cancelar`, solo ventas pagadas del mismo día (zona horaria del negocio), motivo de lista `ERROR_CAPTURA` «Error de captura», `COBRO_DUPLICADO` «Cobro duplicado», `DEVOLUCION_INMEDIATA` «Devolución inmediata del cliente» u `OTRO` «Otro» (con explicación obligatoria); confirmación que dice cuántas piezas regresan y que no se puede deshacer. Fuera de plazo responde 409 `VENTA_FUERA_DE_PLAZO`. Días anteriores se atenderán como devolución en el futuro.
- Backend: `GET /ventas` acepta `periodo=hoy|semana|mes` y devuelve `resumen` (vendido, tickets, ticket promedio, efectivo, cancelados) y `desde`/`hasta` del periodo; cada venta trae `atendio` (nombre del perfil o correo) y `cancelable`; la cancelada trae `cancelada_por`. Columna nueva `ventas.motivo_cancelacion_codigo`.
- El ticket impreso pasa a un componente compartido por venta terminada y reimpresión.

## Fase 5 · Accesos rápidos

- Menú lateral: botón «Nueva venta» con distintivo F2 arriba de «Administración», en todas las pantallas, solo con `ventas.registrar` (contraído muestra solo el ícono).
- Inicio: botón «Nueva venta» en la bienvenida y primera tarjeta de «Ir directo a», con `ventas.registrar`.
- Tras cobrar, F2 en la venta terminada empieza la siguiente; con la ventana de cobro abierta no hace nada.
- Archivos: `layout/sidebar/sidebar.component.*` (botón), `layout/shell/app-shell.component.ts` (escucha F2), `features/ventas/atajo-nueva-venta.service.ts` (reglas), `features/home/home.component.html` (bienvenida y acceso) y `.atajo-tecla` en `styles/componentes.css`.
- F2 abre `/ventas/nueva` desde cualquier pantalla con `ventas.registrar`. No navega si hay cambios pendientes registrados (`CambiosPendientesService`) ni si el foco está en un formulario con datos capturados; en esos casos avisa «Guarda o descarta los cambios antes de abrir una venta». En la captura, F2 enfoca el campo.
- Fuera: menú reducido e inicio de vendedor (Fase 6).

## Fase 6 · Inicio para vendedor (cerrada por decisión del usuario)

Revisión del 2026-10-28:
- Cubierto por lo existente: todas las secciones tienen guard de permisos (`permisoGuard`/`permisoChildGuard`; Sucursales redirige a una ruta protegida) y la API autoriza; el menú se filtra por permisos, así que un Cajero con solo `ventas.registrar` ve Inicio y Ventas; Ventas muestra solo sus ventas; el acceso desde Inicio se hizo en la Fase 5.
- No implementado (pendiente, sin aprobar): Inicio de vendedor del tablero 01b (números del turno, últimas ventas, consulta de precio, aviso de productos por agotarse), redirección por permisos y menú con «Mis ventas» y «Consultar precios».
- Detalle conocido: el elemento «Ventas» del menú no declara permiso y lo ve cualquier miembro; sin permisos de ventas, la ruta muestra «Sin acceso».

## Decisiones

- 2026-10-28: dinero en centavos `bigint` (el resto del proyecto usa `float64`); IVA 16 % fijo incluido (tabla `impuestos` vacía); búsqueda por código y nombre, no por SKU (el usuario indicó que el código de barras es lo visible); Cajero = solo `ventas.registrar` con migración de roles existentes; sin campo «caja»; cantidades enteras; lotes no se descuentan.
- 2026-10-28: el texto «solo dígitos, 6 o más» del plan se ajusta a «código exacto primero, cualquier longitud», porque la base local tiene 53 códigos de menos de 6 dígitos.
- 2026-10-28: el costo se guarda en centavos redondeados (`costo_vigente` tiene 4 decimales); sin costo capturado (0) la línea guarda costo nulo.
- 2026-10-28: el folio se reserva al final de la transacción para no consumir números en ventas fallidas y mantener la serie sin huecos.

- 2026-10-28 (Fase 2): el encabezado muestra solo la sucursal; las fichas de prueba del lector eran del diseño y se omiten; existencia baja = mínimo de la sucursal.

## Anomalías registradas (fuera de alcance)

- SKU: GORM escribe en `sk_uinterno` y el SQL lee `sku_interno`; la validación de SKU duplicado no funciona. Requiere capability propia.
- Sin `CHECK stock_actual >= 0`; `lotes.cantidad_actual` no se descuenta; 149 precios menores al costo y 5 en 0 en la base local.

## Tareas

### Fase 1 · Datos y lógica
- [x] Dominio `venta`: modelos, centavos, cálculos.
- [x] Aplicación: servicio, validaciones y contratos.
- [x] Infraestructura: búsqueda, registrar y cancelar transaccionales, listar/obtener.
- [x] HTTP y rutas.
- [x] Permisos, rol Cajero y migración.
- [x] Frontend: códigos de permiso y nombre de módulo.
- [x] Pruebas unitarias, de handler e integración.
- [x] Verificación y documentación.

### Fase 2 · Captura (04b)
- [x] Backend: `existencia_baja` en la búsqueda.
- [x] Rutas, servicio, cálculos y borrador.
- [x] Pantalla de captura.
- [x] Pruebas y verificación.
- [ ] Prueba manual: 10 productos solo con el lector.

### Fase 3 · Cobro (solo efectivo)
- [x] Servicio `registrar` y modelos de la venta.
- [x] Ventana de cobro en efectivo (según captura del usuario).
- [x] Venta terminada e impresión (según captura del usuario).
- [x] Pruebas y verificación.
- [ ] Prueba manual: venta guardada, inventario baja, folio sube en 1, ticket impreso.

### Fase 4 · Historial (04)
- [x] Backend: periodo, resumen, nombres, cancelable y motivo de lista.
- [x] Ticket impreso compartido.
- [x] Lista de ventas.
- [x] Detalle del ticket y cancelación.
- [x] Pruebas y verificación.
- [ ] Prueba manual: cancelar un ticket lo tacha y regresa sus piezas; reimpresión.

### Fase 5 · Accesos rápidos
- [x] Botón del menú y F2 global.
- [x] Accesos en Inicio.
- [x] Pruebas y verificación.
- [ ] Prueba manual: llegar a la venta con F2 desde varias pantallas.

### Fase 6 · Inicio para vendedor
- [x] Bloqueo por URL (ya existente, revisado).
- [x] Menú filtrado por permisos (ya existente, revisado).
- Pendiente sin aprobar: Inicio 01b, redirección por permisos, «Mis ventas»/«Consultar precios», permiso del elemento «Ventas» del menú.

### Pendientes sin aprobar
- Cobro con tarjeta y transferencia.
- Protección de cambios sin guardar en Compras para F2.
- Regla opcional: un encargado no cancela sus propias ventas.

## Verificación

- 2026-10-28 (Fase 1): `go build ./...` y `go vet ./...` correctos.
- `go test ./...`: todo pasa salvo `TestProductionAcceptsStrongOTPSecret` (`internal/config`), falla previa ajena.
- Unitarias: `internal/domain/venta` (IVA $64.00 → $55.17 + $8.83, cobro, margen y nivel, conversión exacta de pesos, folio), `internal/application/venta` (código primero, escaneo inexistente sin búsqueda por nombre, normalización, costos ocultos, líneas sumadas, validaciones, motivo) e `internal/interfaces/http/venta` (403 sin permisos, cajero limitado a sus ventas, sucursal ajena 404, ticket ajeno 404, costos solo con `ventas.ver_costos`).
- Integración `TestVentas` en base temporal `tienda_ventas_test` (Postgres local 5432, creada con autorización y borrada al terminar): 10 de 10 subpruebas — código exacto y corto, nombre sin acentos, sin precio/precio 0/otra sucursal excluidos, comodines escapados, margen negativo; registrar descuenta stock, foto de precio estable ante cambio posterior, movimientos y folio V-000001; tarjeta con referencia y folio +1; errores sin rastro ni folio consumido; 8 ventas simultáneas con folios consecutivos; dos cajeros por la última pieza (uno gana, el otro recibe el nombre); cancelación única con piezas de vuelta; filtros de listado; migración del Cajero idempotente y sin tocar roles personalizados. `TestCrearNegocioSiembraRolCajero` pasa con el nuevo permiso.
- Búsqueda por nombre contra la base local `tienda2` en transacción de solo lectura: ~25 ms con 2,400 productos.
- Frontend: `ng test` de `contexto`, `invitaciones`, `roles-permisos` y `layout` (Node 22.23.2): 71 de 71. `ng build` correcto con las advertencias de presupuesto ya existentes.
- Pendientes conocidos: la migración del Cajero y los permisos nuevos se aplican al reiniciar la API; un Cajero migrado verá el marcador de Ventas hasta la Fase 2/6. Sin prueba manual (no hay interfaz en esta fase).
- 2026-10-28 (Fase 2): `ng test --include src/app/features/ventas` (Node 22.23.2): 15 de 15 — encabezado con sucursal y foco, escaneo repetido en una línea, código inexistente, agotado, tope de + y escaneo de más, totales $64.00 → $55.17/$8.83, quitar con deshacer en su lugar, coincidencias con flechas/Enter, Escape, «Verificar costos» solo con permiso y alerta bajo costo, borrador restaurado y guardado.
- Suite Angular completa: 294 de 300; las 6 fallas son las conocidas de `alta-inicial.spec.ts`. `ng build` correcto; `nueva-venta.component.css` rebasa el presupuesto de 4 kB por 2.58 kB (advertencia, como otras pantallas).
- Backend tras `existencia_baja`: `TestVentas` pasa en `tienda_ventas_test` (creada y borrada); `go test ./internal/...` solo con la falla previa de configuración OTP.
- Archivos: `features/ventas/` (`nueva-venta.component.*`, `venta-calculos.ts`, `ventas.service.ts`, `ventas.models.ts`, specs, rutas), `app.routes.ts` (contexto y migas de Ventas) y `FeedbackService.conAccion` (aviso con «Deshacer»).
- Pendientes: prueba manual con lector y aceptación. Hasta la Fase 5 no hay botón hacia `/ventas/nueva`; se abre por URL. «Terminar venta» queda deshabilitado hasta la Fase 3.
- 2026-10-28 (Fase 3, parcial): `ng test --include src/app/features/ventas`: 21 de 21 (incluye conversión exacta de «Recibido», «Terminar venta» habilitado solo con productos, «Falta»/cambio, billetes rápidos, tarjeta y transferencia deshabilitadas, envío `POST /ventas` con `recibido` «100.00», lista y borrador vaciados tras cobrar, error de existencia con el nombre del producto y tope ajustado; `VentasService.registrar` y borrador). `ng build` correcto. Pendiente: pantalla de venta terminada con «Imprimir ticket», según captura.
- 2026-10-28 (Fase 3, cierre): pantalla de venta terminada e impresión. `ng test --include src/app/features/ventas`: 21 de 21 (además: pantalla final con folio, total, piezas y cambio, ticket imprimible con líneas, «Imprimir ticket» llama a la impresión, «Nueva venta» regresa a la captura vacía). Suite completa: 300 de 306, solo las 6 fallas conocidas de `alta-inicial.spec.ts`. `ng build` correcto; `nueva-venta.component.css` 10.96 kB contra presupuesto de 4 kB (advertencia; conviene separar cobro y venta terminada en componentes propios en una fase posterior). Guardado, descuento de inventario y folio +1 cubiertos por `TestVentas` (Fase 1).
- 2026-10-28 (Fase 4): backend — `go test ./internal/...` solo con la falla previa de configuración OTP; `TestVentas` 11 de 11 en `tienda_ventas_test` (creada y borrada), con resumen por hoy/semana/mes, forma de pago que solo acota la tabla, venta de ayer fuera de «hoy», no cancelable y rechazada con `VENTA_FUERA_DE_PLAZO`, «Atendió» con perfil o correo, motivo de lista con explicación y quién canceló. Unitarias de motivos («Otro» exige explicación), periodo inválido y ticket promedio; handler valida `periodo`.
- Frontend — `ng test --include src/app/features/ventas`: 33 de 33 (estable en dos corridas): lista con sucursal, fecha, métricas, pie, hora/fecha, tachado, filtros por folio/cajero sin acentos y forma de pago, cambio de periodo; ticket con encabezado sin caja, botón de cancelar solo con permiso y deshabilitado fuera de plazo, motivo obligatorio y explicación para «Otro», ticket tachado con quién y motivo, error de plazo, reimpresión y ticket inexistente; fechas en Ciudad de México. Suite completa: 312 de 318, solo las 6 conocidas de `alta-inicial.spec.ts` (una corrida previa mostró una séptima falla intermitente fuera de ventas que no se repitió). `ng build` correcto; `nueva-venta.component.css` 10.3 kB (advertencia).
- Durante la fase, un `gofmt -w internal` reformateó por error 4 archivos ajenos (asignaciones y pruebas de migración de `negocio`); se revirtieron con el parche inverso y quedaron como estaban.
- Pendientes: prueba manual; listado limitado a 1000 tickets por periodo (el resumen sí cuenta todos); el menú aún muestra Ventas sin permiso (Fase 6).
- 2026-10-28 (Fase 5): `ng test` de `features/ventas`, `features/home` y `layout`: 70 de 70 (botón del menú con F2 solo con permiso, «Nueva venta» en bienvenida y primer acceso, F2 navega, respeta permiso, tecla sostenida y captura abierta, no navega con cambios pendientes ni con un formulario con datos, en la captura enfoca el campo). Suite completa: 321 de 327, solo las 6 conocidas de `alta-inicial.spec.ts`. `ng build` correcto. Prettier había convertido una expresión `\u0300-\u036f` en caracteres literales; se restauró la forma escapada.
- Pendientes: prueba manual; formularios de otros módulos que no registran cambios pendientes (por ejemplo Compras) se protegen solo por la regla «foco en formulario con datos»; si el foco está fuera del formulario, F2 navega (el borrador de Compras no se guarda).
- 2026-10-28 (ajuste móvil, reportado por el usuario): las ventanas de cobro y de cancelar no se veían bien en teléfono. Reproducido con Chrome sin interfaz dentro de un marco de 375 px: con el teclado abierto (≈420 px de alto) «Confirmar cobro» quedaba fuera de la vista, las opciones de pago quedaban apretadas e iOS acercaba la pantalla en campos con letra menor a 16 px. Corrección solo de estilos (`nueva-venta.component.css`, `ticket-venta.component.css`, `ventas.component.css`): ambas ventanas limitan su alto y se desplazan con el pie fijo abajo; en ≤ 600 px suben desde abajo a todo lo ancho con botones apilados y margen para el área segura; campos a 16 px en teléfono. Capturas antes/después a 375×667 y 375×420 revisadas; tablet sin cambios. `ng test` de ventas 39 de 39 y `ng build` correcto. Pendiente: confirmar en un teléfono real.
