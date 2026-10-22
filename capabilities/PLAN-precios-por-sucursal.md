# Plan: catálogo de negocio sin precios; costo y precio solo por sucursal

Encargado: Ilian.

Borrador de trabajo. No es una capability aprobada (AGENTS.md exige `capabilities/NN-nombre.md` en estado `Aprobada` antes de implementar). Si se acepta, se convierte en la capability 19. La actualización documental del 2026-10-19 no autoriza implementación.

## Situación actual (verificada en código)

1. La importación (cap. 16-18) no trae costo: en la primera compra `CostoAnterior` queda nulo y no hay variación que calcular (`compra_repository.go:190-226`).
2. Hay dos precios sin relación: `ProductoNegocio.precio_venta` (lo que muestra el catálogo, `product_repository.go:2167`, `:2235`) y `precios_sucursal.precio_venta` (lo que autoriza la cap. 12, `precio_repository.go:36-48`). Autorizar un precio no actualiza el catálogo.
3. La edición manual de precio escribe en `ProductoNegocio.precio_venta` (`product_repository.go:1130`, `:1411`) y sobrescribe el valor sin historial.
4. `precios_sucursal` también se sobrescribe en cada autorización; el único historial es `propuestas_costo_precio`, y solo para compras.
5. Los permisos ya se exigen por sección y operación: Catálogo usa `catalogo.ver`/`catalogo.gestionar`, Compras usa `compras.ver`/`compras.registrar` y Precios usa `precios.ver`/`precios.autorizar`. La ruta de autorización ya exige `precios.autorizar`; el handler comprueba además el acceso a la sucursal. La interfaz también aplica permisos en rutas y acciones.
6. No existe módulo de ventas (frontend `features/ventas/` solo tiene `ventas.routes.ts`; backend sin tablas de ventas).

## Acordado

### Modelo
- `producto_negocio` es **solo catálogo a nivel negocio**: se **retiran sus campos de montos** (`precio_venta`, `costo_referencia`, `precio_mayoreo`; verificado que `precio_mayoreo` no se usa en backend ni frontend).
- Costo proveedor y precio de venta viven **solo** en `precios_sucursal`, por sucursal.
- Stock en `InventarioSucursal`, por sucursal (sin cambios).
- Sin tablas nuevas.
- No hay datos reales: no se migra nada (decidido 2026-10-16).

| Dato | Dónde |
|---|---|
| Producto (nombre, marca, categoría, variantes, códigos, SKU) | `Producto` + `ProductoNegocio`, sin montos |
| Stock | `InventarioSucursal` |
| Costo proveedor vigente | `precios_sucursal.costo_vigente` |
| Precio de venta vigente | `precios_sucursal.precio_venta` |
| Historial de costo y precio | `propuestas_costo_precio` |

### Formulario individual de producto
- Se mantiene como está y se le agrega el campo **Costo proveedor**.
- Precio de venta y costo proveedor → `precios_sucursal` de la **sucursal activa**.
- Stock inicial → `InventarioSucursal` de la sucursal activa.
- Las demás sucursales no reciben inventario ni precios; el producto ya existe en `producto_negocio` para que lo den de alta sin volver a registrarlo.
- **Selector de producto existente:** al crear, el formulario ofrece un select/buscador con los productos ya cargados en `producto_negocio`.
  - Si se elige uno: se toma como base, se reutiliza su `ProductoNegocio` y solo se capturan costo, precio y stock de la sucursal activa.
  - Si no aparece: se crea un registro nuevo.
  - Incluye variantes (cada una tiene su propio `ProductoNegocio`).
  - Un producto ya activo en la sucursal (con `InventarioSucursal`) aparece marcado y no se puede volver a elegir.
  - Los datos de catálogo del producto elegido son **solo lectura**: un `ProductoNegocio` es compartido por todas las sucursales y modificarlo las afectaría a todas. Si se necesita algo distinto, se crea otro producto.
- **Opción de registrar solo el producto:** se permite crear el producto sin stock ni precios en la sucursal; la primera compra los carga (y así esa primera entrada queda en el historial).

### Carga masiva
- Columna nueva **Costo** en la plantilla de productos. Probablemente obligatoria; lo define Ilian.
- Costo 0 o costo mayor al precio: **advertencia** en la vista previa; la fila se puede importar.
- Todo (precio, costo, stock) va a la sucursal elegida en la importación.
- Una fila que coincide con un producto ya existente en el negocio lo **activa en esa sucursal** en vez de rechazarlo como duplicado (hoy se rechaza con "ya está registrado en el negocio").

### Quién modifica precios
- **Sin aprobación del dueño.** Al no haber precio base en `producto_negocio`, no hay nada que aprobar "sobre el precio del negocio".
- Modifica costo y precio de sucursal quien tenga el permiso `precios.autorizar` en esa sucursal. El rol `PROPIETARIO` lo tiene por defecto y puede asignarlo a otros roles.
- Se conserva el control de permisos existente por sección y operación. Autorizar propuestas ya exige `precios.autorizar`; el nuevo endpoint de cambio manual deberá exigirlo también, además del acceso a la sucursal.

### Dos formas de cambiar un precio
1. **Por compra (flujo actual de cap. 12):** la recepción crea una propuesta pendiente con costo capturado, costo anterior, precio anterior, margen y precio sugerido. Quien tenga `precios.autorizar` la ajusta y autoriza; eso actualiza `precios_sucursal`.
2. **Manual:** desde la pantalla de precios (o la edición del producto) se modifica **solo el precio de venta** de la sucursal activa, sin compra. Se sigue permitiendo como hoy, pero escribe en `precios_sucursal`, no en `producto_negocio`. El **costo proveedor no se cambia manualmente**: solo entra por compra (o por el alta inicial).

### Historial (opción A acordada)
- Todo cambio de costo o precio queda en **`propuestas_costo_precio`**, reutilizando la tabla.
- Nueva columna **`origen`**: `COMPRA` o `MANUAL`.
- Un cambio manual se registra como propuesta **autorizada en el mismo acto** (quién y cuándo), con su valor anterior y el nuevo.
- `compra_detalle_id` pasa a ser **opcional** (hoy es `not null` + `uniqueIndex`): obligatorio solo cuando `origen = COMPRA`.
- `precios_sucursal` sigue guardando solo el valor vigente; el historial es la tabla de propuestas.
- **El alta inicial no se registra en el historial** (formulario o Excel). Si se cargan stock y precios en el alta, esa carga inicial no tendrá historial; el historial empieza con el primer cambio (compra o manual).

## Regla para el futuro módulo de ventas (no se implementa ahora)

- El detalle de cada venta debe guardar **precio unitario, costo unitario y ganancia del momento** de la venta.
- Los reportes y dashboards históricos suman esos valores guardados, nunca el precio o costo vigente.
- Así un cambio de precio o costo (ej.: margen de cajetilla de 60 a 50) no altera ventas ya registradas.

## Fases

### Fase 1: retirar montos de `producto_negocio`
- Eliminar `precio_venta`, `costo_referencia` y `precio_mayoreo` del modelo.
- Backend: crear y editar producto y variantes sin montos en `ProductoNegocio`; escribir precio y costo en `precios_sucursal` de la sucursal activa (8 archivos Go de producción y 2 de pruebas usan `precio_venta`/`PrecioVenta`).
- Listado y detalle: precio y costo salen de `precios_sucursal` de la sucursal activa; sin fila = sin precio en esa sucursal.
- Frontend: el formulario conserva el precio, agrega **Costo proveedor** y envía ambos a la sucursal activa (`product.models.ts`, `productos.component.*`, `precios.component.ts`, `precios.service.ts`).

### Fase 2: historial y edición manual
- `propuestas_costo_precio`: columna `origen` (`COMPRA`/`MANUAL`) y `compra_detalle_id` opcional.
- Endpoint de cambio manual del **precio de venta** por sucursal (no del costo): actualiza `precios_sucursal` y registra la propuesta autorizada en el acto.
- Conservar `precios.autorizar` en la autorización de propuestas y exigirlo también en el nuevo cambio manual, sin retirar los controles de sección ni de acceso a sucursal.
- Pantalla de precios: ver historial por producto y sucursal (origen, valores anterior/nuevo, quién, cuándo).

### Fase 3: selector de producto existente
- Endpoint de búsqueda en `producto_negocio` (nombre, presentación, código de barras, SKU), incluyendo variantes y marcando las ya activas en la sucursal.
- Alta en sucursal de un `ProductoNegocio` existente: crea `InventarioSucursal` y `precios_sucursal`, sin duplicar `Producto` ni `ProductoNegocio`.

### Fase 4: carga masiva con costo y activación por sucursal
- Columna **Costo** en plantilla, parser y vista previa.
- Precio y costo → `precios_sucursal`; stock → `InventarioSucursal` de la sucursal elegida.
- Producto existente en el negocio: se activa en la sucursal; si ya está activo ahí, es duplicado. La vista previa distingue "nuevo", "se activa en esta sucursal" y "duplicado".
- Resultado: la primera compra ya trae `CostoAnterior`.

### Fase 5: variación de costo de compra (objetivo de la clienta)
- En la propuesta de compra: costo anterior, costo nuevo y % de variación.

## Pendientes por decidir

1. **Criterio para reconocer un producto existente en la importación (pensarlo bien).** Primera opción: nombre + presentación; segunda opción: código de barras. Falta definir:
   - Variantes: ¿la clave incluye la combinación de atributos (nombre + presentación + `Sabor: Fresa`)?
   - Si coincide nombre + presentación pero el código de barras es otro (o al revés), ¿error, advertencia o producto nuevo?
2. Columna Costo obligatoria u opcional, y si se siguen aceptando archivos de 13 columnas: lo define Ilian (probablemente obligatoria).
3. Alerta de variación de costo (Fase 5): ¿avisar cuando el costo de una compra suba o baje más de cierto porcentaje respecto al anterior (ej.: aviso si sube más de 10 %)? ¿Se quiere y con qué porcentaje?
4. Precio de mayoreo: hoy solo existe la columna `producto_negocio.precio_mayoreo` (sin uso) y la bandera `precios_mayoreo_habilitados` en la configuración del negocio. Recomendado: retirarla ahora y, cuando se diseñe ventas, agregar a `precios_sucursal` el precio de mayoreo junto con la cantidad mínima que lo activa.

## Pruebas previstas

- `ProductoNegocio` no tiene columnas de montos.
- Crear producto en la sucursal A: montos en `precios_sucursal` de A y stock en `InventarioSucursal` de A; la sucursal B no tiene filas.
- Selector en la sucursal B: reutiliza el `ProductoNegocio` existente y crea filas solo de B; un producto ya activo en B no se puede elegir.
- Cambio manual de precio: actualiza `precios_sucursal` y crea propuesta `MANUAL` autorizada con valor anterior y nuevo; el costo no cambia.
- Alta de producto sin stock ni precios: existe en `producto_negocio`; la primera compra crea `InventarioSucursal`, la propuesta y, al autorizar, `precios_sucursal`.
- Compra: crea propuesta `COMPRA` pendiente; al autorizar actualiza `precios_sucursal`.
- Usuario sin `precios.autorizar` no autoriza ni cambia precios manualmente.
- Importar con costo: se crean `precios_sucursal` con costo y precio.
- Importar en B productos que ya existen en A: se activan en B, no se rechazan.
- Primera compra con costo cargado: `CostoAnterior` presente.
- Aislamiento: los precios de una sucursal no afectan a otra.
- Regresión: plantillas, vista previa, variantes y duplicados de cap. 16-18; flujo de compras de cap. 10 y 12.

## Riesgos

- Retirar columnas de `producto_negocio` toca todo el código que lee `precio_venta` (backend y frontend).
- `compra_detalle_id` deja de ser obligatorio: revisar consultas de cap. 12 que asumen que toda propuesta viene de una compra.
- Cambia la plantilla de importación: los archivos de 13 columnas dejan de ser válidos si no se acepta el formato anterior.
- El nuevo cambio manual debe exigir `precios.autorizar` y acceso a la sucursal, aunque se inicie desde la edición del producto; `catalogo.gestionar` por sí solo no debe permitir cambiar el precio vigente de sucursal. La autorización de propuestas ya está protegida.

## Base verificada

Leído en código: modelos `PrecioSucursal`, `Propuesta`, `ProductoNegocio`; `createPriceProposal` (`compra_repository.go:190-226`); autorización (`precio_repository.go:36-48`); middleware de permisos por sección y operación en `backend/cmd/api/main.go`, complementado por `sucursalPermitida` en el handler de precios; rol `PROPIETARIO` (`domain/negocio/rol.go:32`); consultas de listado; conteo de archivos con `precio_venta`. No ejecutado contra base de datos.

Actualización documental 2026-10-19: permisos contrastados con rutas HTTP, rutas de Catálogo y acciones de la interfaz; convención de capabilities contrastada con `AGENTS.md`. Sin cambios de código ni ejecución de pruebas de aplicación.

## Corrección de rendimiento del listado de productos (2026-10-21)

Aprobada por el usuario en conversación el 2026-10-21 (solo corrección de consulta, sin cambio de contrato).

- **Síntoma:** con más de 1500 productos, el listado tarda notablemente cada vez que se entra a la vista. La recarga al volver ya existía (`effect` en `productos.component.ts`); lo nuevo es el tiempo de respuesta.
- **Causa:** el merge `c8030a1` introdujo en `ProductRepository.ListByBusiness` un `LEFT JOIN LATERAL existencias` que, por cada producto, recorría todo el inventario de la sucursal filtrando con `oferta.producto_id = p.id OR variante.producto_id = p.id`; el `OR` impedía usar índices (costo cuadrático).
- **Cambio:** `existencias` pasa a ser una subconsulta agregada una sola vez por `COALESCE(oferta.producto_id, variante.producto_id)` y unida por `existencias.producto_id = p.id`. Mismas condiciones: oferta activa del negocio, variante activa, solo la sucursal seleccionada. Contrato HTTP sin cambios.
- **Índices:** no se agregan; ya existen `idx_inventario_sucursal_producto (sucursal_id, producto_negocio_id)` y `idx_precio_sucursal_producto (sucursal_id, producto_negocio_id)`.
- **Pruebas ejecutadas:** `go build ./...` y `go vet ./internal/infrastructure/` correctos; `go test ./...` correcto salvo `internal/config` `TestProductionAcceptsStrongOTPSecret`, falla ajena a este cambio (configuración OTP/entorno).
- **Pendientes:** las pruebas de integración (`catalog_variants_tenancy_integration_test.go`) se omitieron por falta de `SECURITY_TEST_DATABASE_URL`; falta medir tiempo y `EXPLAIN ANALYZE` contra la base con más de 1500 productos, y comprobar en navegador que stock y estado coinciden con los de antes (incluyendo variantes).
- **Fuera de alcance:** caché del listado en frontend, paginación en servidor y endpoint de detalle para edición (posible capability 21).
