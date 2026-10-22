# Observaciones a los cambios de Ilian

- Fecha: 2026-10-22.
- Revisados: commits de `ilimm9` del 2026-10-21 (`85cb7c2` «ajustes tabla productos», merge `c8030a1` «cambios rama ilu» y `d384c52` «correccion compras»), contrastados con sus capabilities 19, 21, 22 y 23, más hallazgos de la prueba con el archivo de la clienta (`thrs/docs-prueba/06-productos-lista-clienta.xlsx`).
- Método: lectura de código. Salvo donde se indica, no se ejecutó la app.
- Documento de seguimiento: **no autoriza implementación**. Cada corrección requiere su capability aprobada.

## Resumen

| # | Observación | Gravedad | Estado |
|---|---|---|---|
| 1 | Editar producto no guarda el precio | Alta | Pendiente |
| 2 | Compras no muestra productos sin proveedor | Alta | Pendiente (capability por crear, ID 27) |
| 3 | Crear y editar con montos no exige `precios.autorizar` | Media | Pendiente (tarea abierta en la 19) |
| 4 | Variantes nuevas al editar quedan sin precio en la sucursal | Media | Pendiente |
| 5 | La importación no activa productos existentes en otra sucursal | Media | Pendiente (tarea abierta en la 19) |
| 6 | El merge `c8030a1` volvió lento el listado de productos | Media | Corregido en `d384c52`; falta medir |
| 7 | Costos de Excel con residuo de flotante rechazados | Media | Resuelto en la capability 24 |
| 8 | Pruebas manuales y aceptación pendientes (21, 22, 23) | Baja | Pendiente |

## Detalle

### 1. Editar producto no guarda el precio
- `ProductRepository.Update` (`backend/internal/infrastructure/product_repository.go`) solo revisa que `PrecioVenta` no sea negativo y nunca lo escribe en `precios_sucursal`.
- El formulario de edición sí deja cambiar el precio (`productos.component.ts`, `precio_venta: product.precio`).
- Efecto: el usuario cambia el precio, guarda, ve que todo salió bien y el precio queda igual. Falla sin mostrar error.
- La 19 tiene la tarea abierta («Cambiar creación y edición… para escribir exclusivamente `precios_sucursal`»), pero no advierte este comportamiento.
- Sugerencia: mientras se implementa, deshabilitar el campo precio al editar o mandar al usuario a «Validar costos y precios».

### 2. Registrar compra no muestra productos sin proveedor
- `compra.Repository.ListProductos` (`backend/internal/infrastructure/compra/compra_repository.go:40`, introducido en `a827745` «inicio proveedores», 2026-09-30) solo lista productos con fila en `producto_proveedor` para el proveedor elegido.
- El vínculo solo se crea al importar con la columna Proveedor llena, al editar el producto o al crear el producto desde la compra.
- El export de Odoo de la clienta no trae proveedor, así que ninguno de sus 948 productos aparece en Compras. Se confirmó en local.
- Propuesta (capability por crear, ID 27): listar todos los productos del negocio, con los del proveedor elegido primero, y crear el vínculo al guardar la compra.

### 3. Montos sin `precios.autorizar`
- `POST /catalogo/productos` y `PATCH /catalogo/productos/:id` (`cmd/api/main.go`) solo exigen el permiso de gestión de catálogo.
- La regla de la 19 pide `precios.autorizar` cuando se envían montos. Hoy cualquier usuario que puede crear productos fija el precio y el costo iniciales de la sucursal.

### 4. Variantes nuevas al editar quedan sin precio
- `createFamilyVariantOnUpdate` (también la usa `convertSimpleProductToFamily`) crea `producto_negocio`, pero no llama a `createInitialBranchPrice`.
- Efecto: la variante aparece «sin precio» en la sucursal aunque se capturó uno.

### 5. Importación y productos existentes
- Una fila cuyo producto ya existe en el negocio se omite como duplicada. No se activa en la sucursal actual ni se le crea precio, costo o stock.
- Es una tarea abierta de la 19 («activar productos existentes en la sucursal sin duplicar catálogo»).

### 6. Rendimiento del listado
- El merge `c8030a1` agregó un `LEFT JOIN LATERAL` de existencias con `OR` que impedía usar índices (costo cuadrático con más de 1500 productos).
- Ilian lo corrigió en `d384c52` con una subconsulta agregada.
- Pendiente, según su propio registro en `PLAN-precios-por-sucursal.md`: `EXPLAIN ANALYZE` contra una base con más de 1500 productos y las pruebas de integración (se omitieron por falta de `SECURITY_TEST_DATABASE_URL`).

### 7. Costos con residuo de flotante
- La regla de la 19 (hasta 4 decimales) rechazaba valores como `8.380000000000001`, que Excel guarda aunque muestre `8.38`: 20 filas del archivo de la clienta.
- Resuelto en la [capability 24](24-correccion-masiva-producto.md): el costo importado se redondea a 2 decimales.

### 8. Pendientes de cierre
- 21 (lista en memoria): cerrada.
- 22 (código opcional desde Compras) y 23 (tarjetas compactables): verificadas, falta prueba manual y aceptación final. La 22 no tiene prueba automática de backend para el alta sin código.
- `TestProductionAcceptsStrongOTPSecret` (`internal/config`) sigue fallando; está registrada como falla previa.
- `alta-inicial.spec.ts`: 6 pruebas fallan con `window.matchMedia is not a function`, a raíz del cambio a `p-select` en el alta inicial (`85cb7c2`, capability 14).

## Lo que sí coincide con sus documentos

- 19: historial independiente, ajuste manual con motivo y permiso, montos fuera de `producto_negocio`, listado y compras leyendo `precios_sucursal`, y costo opcional en la plantilla.
- 21, 22 y 23: el código corresponde a lo descrito.
- Las casillas marcadas como hechas en sus `.md` están implementadas; lo pendiente está marcado como pendiente. El único hueco de documentación es la observación 1.

## Siguientes pasos sugeridos

1. Atender la 1 con prioridad (falla sin error, visible para el usuario).
2. Redactar y aprobar la capability de Compras (ID 27, observación 2).
3. Cerrar las tareas abiertas de la 19 (observaciones 3, 4 y 5) con una regresión completa de productos, variantes, importación, compras y precios.
4. Corregir las pruebas de `alta-inicial` (simular `matchMedia` en el entorno de pruebas).
