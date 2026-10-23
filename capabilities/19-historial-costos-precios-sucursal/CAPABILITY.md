# Capability 19: Precios, costos e historial por sucursal

## Estado

En implementación.

## Aprobación

Aprobada inicialmente por el usuario el 2026-10-20 mediante la instrucción: “Implement the plan”.

El 2026-10-20 el usuario solicitó esta revisión de alcance: eliminar los montos de `producto_negocio`; mantener el valor vigente únicamente en `precios_sucursal`; y conservar el historial en una tabla independiente. La revisión fue aprobada explícitamente el 2026-10-20 mediante “Implement the plan”.

El 2026-10-21 el usuario ajustó la regla de importación: `Costo` es opcional. Una celda vacía se importa como cero y se cuenta en la alerta de vista previa; cuando se capture, debe ser un número finito no negativo, con hasta cuatro decimales y dentro del rango almacenado. No se impone relación entre costo y precio de venta.

## Alcance y reglas

- `producto_negocio` queda como catálogo compartido y no contiene montos: se eliminan `precio_venta`, `costo_referencia` y `precio_mayoreo`.
- Los valores vigentes pertenecen exclusivamente a `precios_sucursal` por producto y sucursal.
- `historial_costo_precio_sucursal` es una bitácora independiente e inmutable con los orígenes `ALTA_INICIAL`, `COMPRA_AUTORIZADA` y `AJUSTE_PRECIO_MANUAL`.
- El alta inicial y la autorización de una compra actualizan el valor vigente y agregan su evento dentro de la misma transacción.
- El ajuste manual modifica únicamente precio de venta, exige motivo y permiso `precios.autorizar`; nunca cambia costo.
- Las propuestas de compra permanecen separadas: son solicitudes pendientes y solo generan historial al autorizarse.
- El historial se consulta por sucursal y producto mediante `precios.ver`. No existen operaciones de edición ni eliminación.
- No se migran datos ni se generan eventos retrospectivos: no existen datos reales que conservar.
- Sin fila en `precios_sucursal`, un producto está disponible como catálogo pero aparece sin precio ni costo configurado para esa sucursal.
- La plantilla de importación de productos incluye `Costo` como columna opcional. Una celda vacía se importa como cero y aumenta `costos_faltantes`; un negativo, texto, valor no finito, precisión mayor de cuatro decimales o valor fuera de rango invalida la fila.
- La vista previa expone `costos_faltantes` y muestra el total como alerta resumida, sin errores por una celda vacía.

## Contratos

- `GET /negocios/:negocioId/precios/historial?sucursal_id=&producto_negocio_id=` lista eventos en orden descendente y paginados.
- `PATCH /negocios/:negocioId/precios/productos/:productoNegocioId` recibe `{ precio_venta, motivo, sucursal_id }` y crea un ajuste manual.
- Los eventos exponen origen, valores anterior/nuevo, usuario, fecha, motivo y propuesta vinculada cuando aplica.
- Los contratos de crear, editar, listar e importar producto reciben/retornan precios y costos en el contexto de una sucursal; ningún contrato entrega un precio global desde `producto_negocio`.
- La respuesta de vista previa de importación incorpora `costos_faltantes`.

## Pruebas y aceptación

- El esquema no contiene columnas de montos en `producto_negocio`.
- El alta inicial crea precio vigente y evento con valores previos nulos.
- La autorización de compra actualiza precio vigente y agrega un único evento vinculado a la propuesta.
- El ajuste manual exige permiso y motivo, no altera el costo y queda auditado.
- La consulta no cruza datos de otras sucursales y no hay rutas para alterar eventos.
- Crear, editar, listar e importar en la sucursal A no modifica ni muestra los montos vigentes de B.
- Un producto sin fila de `precios_sucursal` se muestra sin precio; las compras toman ese costo anterior como ausente, no como cero ni como fallback global.
- La importación acepta costo ausente o cero y rechaza costo negativo, no numérico, no finito, con más de cuatro decimales o fuera de rango; también permite un costo mayor al precio de venta.
- La vista previa muestra el contador de filas sin costo sin invalidar esas filas.
- Las pruebas de precios, compras y compilación frontend pasan; resultados se registran aquí.

## Tareas y decisiones

- [x] Acordar una bitácora independiente de las propuestas de compra.
- [x] Acordar registrar el alta inicial y no migrar datos históricos existentes.
- [x] Crear modelo, migración y repositorio de historial.
- [x] Integrar altas iniciales, autorizaciones de compra, ajuste manual y consulta.
- [x] Adaptar la pantalla de precios y ejecutar verificaciones focalizadas.
- [x] Eliminar columnas de montos de `producto_negocio`, incluidos modelos y migración explícita de esquema.
- [ ] Cambiar creación y edición de productos y variantes para escribir exclusivamente `precios_sucursal` de la sucursal activa; exigir `precios.autorizar` cuando envíen montos. Parte de **edición trasladada** el 2026-10-22 a la [capability 29](../historial/29-guardar-precio-al-editar.md) por decisión del usuario; la parte de creación quedó **trasladada** el 2026-10-22 a la [capability 31](../historial/31-crear-productos-solo-propietario.md) (solo el propietario crea e importa productos).
- [x] Cambiar listado y consultas de compra para leer `precios_sucursal` de la sucursal activa y eliminar fallbacks a `producto_negocio`.
- [ ] Cambiar importación para crear precio/costo/stock locales y activar productos existentes en la sucursal sin duplicar catálogo. **Trasladada** el 2026-10-22 a la [capability 28](../historial/28-surtido-por-sucursal.md) por decisión del usuario; se da seguimiento allá.
- [x] Validar el costo opcional de la plantilla XLSX, propagarlo al precio vigente de sucursal y mostrar `costos_faltantes` en la vista previa.
- [ ] Añadir selector de productos existentes por sucursal y proteger contra altas locales duplicadas. **Trasladada** el 2026-10-22 a la [capability 28](../historial/28-surtido-por-sucursal.md) por decisión del usuario; se da seguimiento allá.
- [ ] Ejecutar regresión completa de productos, variantes, importación, compras, precios y permisos.

## Verificación

- `go test ./internal/application/precio ./internal/infrastructure/precio ./internal/interfaces/http/precio ./internal/infrastructure -run Test`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.app.json`: correcto.
- `git diff --check`: correcto.
- `go test ./internal/database ./internal/infrastructure ./internal/infrastructure/compra ./internal/infrastructure/precio ./internal/interfaces/http/precio`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.app.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2950 nodos y 8368 relaciones.
- `go test ./internal/application ./internal/infrastructure ./internal/database -run 'Test(ParseProductImport|ProductImportTemplate|MergeImportResults|ProductImport|CatalogVariantsTenancy)'`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.app.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2952 nodos y 8375 relaciones.
- `go test ./internal/application ./internal/infrastructure ./internal/database -run 'Test(ParseProductImport|ProductImportTemplate|MergeImportResults|ProductImport|CatalogVariantsTenancy)'`: correcto tras hacer opcional el costo.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.app.json`: correcto tras hacer opcional el costo.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado sin cambios estructurales adicionales.
