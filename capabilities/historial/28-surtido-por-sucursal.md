# Capability 28: Surtido por sucursal

- Estado: **Cerrada** el 2026-10-23 con aceptación final del usuario.
- Aprobación explícita: 2026-10-22. El usuario ajustó el alcance («al inicio de crear otra sucursal, aunque haya productos en el negocio no los muestres todavía, hasta que el propietario lo indique con un botón… las cargas masivas actualizarían costo o stock mas no duplicarían los productos») y aprobó el plan.
- Fecha: 2026-10-22.
- Origen: prueba manual de la [capability 24](24-correccion-masiva-producto.md) y observación 5 de la [26](26-observaciones-ilian.md). Al cambiar a otra sucursal se ve todo el catálogo del negocio (agotado y sin precio), y la importación omite como duplicado lo que ya existe en el negocio.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Backend en `domain`, `application`, `infrastructure` e `interfaces/http` de productos; frontend `features/productos` y `features/precios`. Sin cambios de esquema; el historial de precios agrega el valor de origen `CARGA_MASIVA`.

## Situación anterior

- El catálogo es del negocio: `productos` y `producto_negocio` existen una sola vez. Lo propio de cada sucursal es `inventario_sucursal` (existencias) y `precios_sucursal` (precio y costo vigentes).
- `ListByBusiness` listaba todo el catálogo activo y solo unía existencias y precio de la sucursal activa.
- La importación omitía como duplicada una fila cuyo producto ya existía en el negocio, aunque no estuviera en la sucursal activa.

## Objetivo

Que cada sucursal tenga su propio surtido dentro del catálogo único del negocio, que el propietario decida qué productos del catálogo entran a una sucursal, y que la carga masiva agregue o actualice sin duplicar el catálogo.

## Alcance y reglas

### Regla base
- Un producto simple o una variante **pertenece a la sucursal** cuando su `producto_negocio` tiene fila en `inventario_sucursal` para esa sucursal, aunque el stock sea 0.
- La fila ya se crea al crear, importar o recibir una compra. Se verificó en la base local `tienda2`: todo producto activo la tiene.

### 1. Listado por sucursal
- La lista de Productos muestra solo productos con inventario en la sucursal activa. Una familia aparece si al menos una variante pertenece a la sucursal y muestra solo esas variantes.
- Una sucursal nueva empieza vacía, aunque el negocio tenga productos.

### 2. Agregar del catálogo (solo propietario)
- Botón **«Agregar del catálogo»** para el propietario. Abre los productos del negocio que no están en la sucursal, con buscador, casillas y «Seleccionar todos».
- Agregar crea `inventario_sucursal` con stock 0 para el producto (o todas sus variantes activas), sin precio: quedan «sin precio» para Validar costos y precios o para la carga masiva.
- No modifica `productos`, `producto_negocio` ni otras sucursales. Es idempotente.

### 3. Crear un producto que ya existe en otra sucursal
- Al crear un producto simple con el mismo código de barras, o el mismo nombre y presentación, que uno del negocio:
  - si ya está en la sucursal: «Este producto ya existe en esta sucursal»;
  - si no: `409 PRODUCTO_EN_OTRA_SUCURSAL`, y el formulario ofrece **«Agregar a esta sucursal»** con el precio, costo y stock capturados.
- Con variantes se mantiene el rechazo y se indica usar «Agregar del catálogo».

### 4. Carga masiva sin duplicar
- Una fila cuyo producto o variante ya existe en el negocio:
  - **no está en la sucursal**: «Se agregará a esta sucursal»; se crean inventario (con movimiento de entrada si hay stock) y precio con el costo y precio de la fila;
  - **ya está en la sucursal**: «Se actualizará»; el **stock se reemplaza** por el del Excel (movimiento de ajuste por la diferencia) y se actualizan precio y costo, con un evento `CARGA_MASIVA` en el historial. La vista previa muestra los cambios de stock, precio y costo como **advertencias**. Una celda de costo vacía conserva el costo vigente. Si nada cambia, no se escribe nada.
- Los datos compartidos del catálogo (nombre, marca, categoría, unidad, proveedor, descripción, imagen) no se modifican; si la fila trae otros, advertencia «Se conservan los datos del catálogo».
- Una fila repetida dentro del mismo archivo se sigue omitiendo.
- El reporte agrega las secciones compactables «Se agregarán a esta sucursal» y «Se actualizarán».

## Fuera de alcance
- Quitar un producto de una sucursal.
- El interruptor «Ver catálogo completo» y el formulario por producto de la versión anterior de esta capability.
- Consulta de precio desde otras sucursales y transferencias entre sucursales.
- Cambios de esquema.

## Relación con otras capabilities
- Absorbe dos tareas abiertas de la [19](../19-historial-costos-precios-sucursal.md), trasladadas el 2026-10-22 por decisión del usuario.
- Usa `SoloPropietario` de la [31](31-crear-productos-solo-propietario.md).
- Compras ([27](27-compras-cualquier-producto.md)) lista todo el catálogo; recibir una compra agrega el producto al surtido de la sucursal.

## Contratos
- `GET …/catalogo/productos?sucursal_id=…` → solo productos de la sucursal.
- `GET …/catalogo/productos/fuera-de-sucursal?sucursal_id=…` (propietario) → productos del negocio que no están en la sucursal.
- `POST …/catalogo/productos/agregar-a-sucursal` (propietario) `{sucursal_id, producto_ids[] | todos}` → `{agregados}`.
- `POST …/catalogo/productos/:productoId/agregar-a-sucursal` (propietario) `{sucursal_id, precio_venta, costo, stock_inicial}` → `201`; `409` si ya está.
- `POST …/catalogo/productos` → `409 {codigo: PRODUCTO_EN_OTRA_SUCURSAL, producto_id, nombre}`.
- Vista previa de importación: `filas_agregar_sucursal[]` y `filas_actualizar[]`; resultado: `agregadas_sucursal` y `actualizadas`.
- Historial de precios: origen `CARGA_MASIVA`.

## Pruebas y criterios de aceptación
- Sucursal B sin productos: lista vacía; A sin cambios.
- Agregar del catálogo en B: aparecen solo los elegidos, sin precio; no se crean filas en `productos` ni `producto_negocio`.
- Importar en B un Excel con productos de A: «Se agregarán a esta sucursal»; ningún producto se duplica. Reimportar con otros valores: «Se actualizarán», stock reemplazado y precio y costo con evento `CARGA_MASIVA`.
- Crear en B un producto de A ofrece agregarlo; en A responde que ya existe en la sucursal.
- Solo el propietario ve y usa «Agregar del catálogo».
- `go test ./internal/...`, `ng test` y `ng build` en verde (salvo fallas previas registradas).

## Tareas
- [x] Aprobación del usuario (2026-10-22, con alcance ajustado).
- [x] Decidir la relación con la 19 (2026-10-22: tareas trasladadas a la 28).
- [x] Backend: listado por sucursal y productos fuera de la sucursal.
- [x] Backend: agregar del catálogo y agregar desde el formulario.
- [x] Backend: importación que agrega o actualiza sin duplicar.
- [x] Frontend: botón y diálogo «Agregar del catálogo»; aviso al crear un producto existente.
- [x] Frontend: secciones nuevas del reporte e historial «Carga masiva».
- [x] Pruebas, registro de resultados y bases `registro-productos.md` y `modulo-compras.md`.
- [x] Prueba manual y aceptación final del usuario (2026-10-23).

## Decisiones
- 2026-10-22: el usuario pidió que su otra sucursal no muestre ni reciba todos los productos.
- 2026-10-22: el usuario decidió trasladar a la 28 las dos tareas abiertas de la 19 sobre productos existentes por sucursal.
- 2026-10-22: una sucursal nueva no muestra productos hasta que el propietario los agrega con «Agregar del catálogo» (elegir con casillas o todos; entran sin precio ni stock).
- 2026-10-22: la carga masiva actualiza precio, costo y stock de productos que ya están en la sucursal; el stock se reemplaza y se avisa como advertencia.
- 2026-10-22: al crear un producto que existe en otra sucursal, el formulario ofrece «Agregar a esta sucursal».
- 2026-10-23 (implementación): «Agregar del catálogo» es una página (`/catalogo/productos/agregar-del-catalogo`), como crear e importar, porque la app no usa diálogos modales. Las rutas nuevas usan `SoloPropietario` (31).
- 2026-10-23 (implementación): `ListByBusiness` ya no usa un `JOIN LATERAL` correlacionado con `OR`; parte del inventario de la sucursal y elige una oferta por producto con `DISTINCT ON`. En `tienda2` (2,430 productos) esa parte bajó de ~620 ms a 2–15 ms; con el filtro por sucursal dentro del lateral llegaba a 1–2.6 s.
- 2026-10-23 (implementación): en la importación, un producto existente se identifica por nombre y presentación (simples), por base y clave (variantes) o por código de barras. Se conserva la regla de la 24: mismo nombre con un código distinto que no existe se rechaza con la indicación de usar variantes. Una fila de un producto que ya está en la sucursal y no cambia se omite con la advertencia «ya está en esta sucursal y no tiene cambios».
- 2026-10-23 (implementación): la importación guarda como responsable al usuario que la inicia (antes quedaba vacío en el historial de alta inicial).
- 2026-10-23 (implementación): al crear desde el formulario, la identidad de producto existente (mismo código, o mismo nombre y presentación sin distinguir mayúsculas ni acentos) aplica también dentro de la misma sucursal: «Este producto ya existe en esta sucursal».

## Pruebas ejecutadas (2026-10-23)

- `go build ./...`, `go vet ./internal/... ./cmd/...` y `gofmt` de los archivos tocados: correctos.
- Integración PostgreSQL (`tienda_security_test`): `TestSurtidoPorSucursal` correcta:
  - la sucursal B nueva lista vacía; A sin cambios; `ListCatalogNotInBranch(B)` trae los 3 productos;
  - `AddToBranch` con una familia agrega sus 2 variantes sin precio ni stock, es idempotente y no crea filas en `productos` ni `producto_negocio`;
  - importar en B el archivo de A: simples como `agregar_sucursal`, variantes ya agregadas como `actualizar`; precio y stock en B, A sin cambios, sin duplicar;
  - reimportar con otros valores: advertencias de stock, precio y costo; stock reemplazado con movimiento `AJUSTE_MANUAL` de −5; un evento `CARGA_MASIVA` con valores anterior y nuevo; filas sin cambios omitidas;
  - costo vacío conserva el costo vigente; reimportar el mismo archivo no escribe nada;
  - fila repetida en el archivo omitida; marca distinta: advertencia «se conservan los datos del catálogo»;
  - crear en B un producto de A (nombre sin acento) responde `ProductInOtherBranchError`; `AddProductToBranch` lo deja en B con precio y stock; agregarlo otra vez y crearlo en A (por nombre o por código) responden «ya existe en esta sucursal»;
  - «agregar todos» cuando B ya tiene todo: 0.
- Pruebas previas de integración (`TestProductEditWritesBranchPrice`, `TestCompraProveedoresVarios`, `TestCatalogVariantsTenancyIsolation`, `TestSecurityPhaseCatalogTenancyMigrationAndIsolation`): correctas.
- `go test ./internal/...`: correcta salvo `TestProductionAcceptsStrongOTPSecret`, falla previa registrada.
- Medición de solo lectura en `tienda2` con `EXPLAIN ANALYZE`: listado por sucursal 2–15 ms y variantes 9–25 ms.
- `ng test --include src/app/features/productos`: 55/55. Pruebas nuevas: página «Agregar del catálogo» (carga, búsqueda sin acentos, elegir visibles, agregar elegidos, agregar todos con confirmación, sin envío vacío); el propietario ve «Agregar del catálogo» y el miembro no; estado vacío de sucursal; crear un producto existente ofrece agregarlo y envía precio, costo y stock; reporte con las secciones «Se agregarán a esta sucursal» y «Se actualizarán».
- `ng test` completo: 255/261; las 6 fallas son las previas de `alta-inicial.spec.ts` (`window.matchMedia`).
- `ng build`: correcto. `graphify update .`: correcto.
- Sin prueba automática: el aviso «Agregar a esta sucursal» dentro de la plantilla del formulario (se probó el método) y la etiqueta «Carga masiva» del historial de precios.

## Pendientes

- Reiniciar el backend activo.
- Prueba manual (escritorio y móvil):
  - la sucursal SUR muestra solo sus productos y Centro no cambia;
  - como propietario, «Agregar del catálogo»: elegir algunos y agregarlos; aparecen sin precio;
  - Carga masiva en SUR con productos de Centro: «Se agregarán a esta sucursal»; al importar aparecen con precio y stock y el catálogo no crece;
  - reimportar con otro stock y precio: «Se actualizarán» con advertencias; el historial de Validar costos y precios muestra «Carga masiva»;
  - crear en SUR un producto que existe en Centro: aparece «Agregar a esta sucursal»;
  - como colaborador: no ve «Agregar del catálogo» y su lista muestra solo su sucursal.
- Aceptación final del usuario.

## Cierre (2026-10-23)

- Aceptación final del usuario: «acepto la 22, 23, 27, 28, 29 y 31». Cerrada y archivada en `historial/`.
