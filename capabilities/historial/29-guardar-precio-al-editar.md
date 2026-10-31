# Capability 29: Guardar el precio al editar un producto

- Estado: **Cerrada** el 2026-10-23 con aceptación final del usuario.
- Aprobación explícita: 2026-10-22, el usuario indicó «empieza con la 29 corrección de precios», con dos ajustes: el motivo es opcional y solo el administrador (permiso `precios.autorizar`) puede cambiar el precio.
- Fecha: 2026-10-22.
- Origen: observaciones 1 y 4 de la [26](26-observaciones-ilian.md), confirmadas por el usuario el 2026-10-22: al editar un producto, el precio no se guarda.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Backend en productos (`domain`, `application`, `infrastructure`, `interfaces/http`); escribe en `precios_sucursal` mediante un contrato explícito, sin importar el repositorio del dominio `precio`. Frontend `features/productos`. Sin cambios de esquema.

## Causa (verificada en código)

- La [19](../19-historial-costos-precios-sucursal.md) movió los montos de `producto_negocio` a `precios_sucursal`. La creación y la importación se adaptaron; la edición no.
- `ProductRepository.Update` (`backend/internal/infrastructure/product_repository.go:1093`) recibe `precio_venta`, solo valida que no sea negativo y **nunca lo escribe**. Lo mismo ocurre con el precio de las variantes existentes (`updateProductFamily`).
- `UpdateProductInput` (`backend/internal/domain/product_view.go:111`) no trae `sucursal_id`, así que el backend no sabe en qué sucursal aplicar el precio.
- El formulario (`productos.component.ts:1344`) sí envía el precio. Resultado: el guardado sale bien y el precio queda igual, sin mostrar error.
- Variantes nuevas al editar (`createFamilyVariantOnUpdate`, también usada por `convertSimpleProductToFamily`) crean `producto_negocio` pero no `precios_sucursal`: quedan «sin precio» aunque se capturó uno (observación 4).

## Objetivo

Que el precio capturado al editar un producto o sus variantes se guarde en la sucursal activa, quede en el historial y respete el permiso de precios.

## Alcance y reglas

### 1. Sucursal en la edición
- `PATCH /catalogo/productos/:productoId` recibe `sucursal_id` (sucursal activa del contexto). El backend valida que sea una sucursal activa del usuario, igual que el listado.

### 2. Precio de producto simple y variantes existentes
- Si el precio enviado es **distinto** del vigente en `precios_sucursal` de esa sucursal:
  - con fila vigente: se actualiza `precio_venta` y se registra un evento `AJUSTE_PRECIO_MANUAL` con precio anterior, precio nuevo, usuario y motivo;
  - sin fila vigente (producto «sin precio»): se crea con costo 0 y evento `ALTA_INICIAL`, igual que `createInitialBranchPrice` (`product_repository.go:754`).
- Si el precio no cambió, no se escribe nada ni se crea evento.
- El costo no se edita desde aquí (sigue llegando por compras o ajuste en Validar costos y precios).
- Precio redondeado a 2 decimales; negativo o no finito se rechaza.
- Solo afecta a la sucursal activa; las demás sucursales no cambian.

### 3. Variantes nuevas al editar
- Las variantes creadas al editar (familia existente o conversión de simple a familia) crean su precio en la sucursal activa con `createInitialBranchPrice`, igual que en la creación.

### 4. Motivo del cambio
- Campo **«Motivo del cambio de precio (opcional)»**, visible solo para quien puede cambiar el precio y solo cuando algún precio difiere del original. Se envía como `motivo_precio`.
- Si va vacío, el historial guarda «Edición del producto».

### 5. Permiso
- Cambiar el precio al editar exige `precios.autorizar` además de la gestión de catálogo. Sin ese permiso:
  - el campo precio (y la columna precio de variantes) se muestra deshabilitado con la nota «Necesitas permiso para cambiar precios»;
  - el backend responde `403` si recibe un precio distinto del vigente.
- Editar los demás datos del producto no requiere `precios.autorizar`.

## Fuera de alcance
- Permiso de montos al **crear** productos (resto de la observación 3 de la 26 y de la tarea de la 19).
- Editar el costo desde el formulario de producto.
- Surtido por sucursal ([28](28-surtido-por-sucursal.md)) y Compras (27).

## Relación con la capability 19
- Cubre la parte de **edición** de la tarea abierta de la 19: «Cambiar creación y edición de productos y variantes para escribir exclusivamente `precios_sucursal` de la sucursal activa; exigir `precios.autorizar` cuando envíen montos». La parte de creación queda en la 19.
- **Decidido el 2026-10-22 por el usuario:** la parte de edición queda trasladada a la 29; en la 19 queda anotada.

## Contratos
- `PATCH …/catalogo/productos/:productoId` agrega `sucursal_id` (obligatorio) y `motivo_precio?`. Respuestas nuevas: `403` si cambia el precio sin `precios.autorizar`; `400` si falta o no es válida la sucursal.
- Sin cambios en el listado ni en el endpoint de ajuste manual de precios.
- Frontend: `UpdateProductRequest.sucursal_id`, `motivo_precio?`.

## Archivos estimados
| Archivo | Cambio |
|---|---|
| `backend/internal/domain/product_view.go` | `SucursalID`, `MotivoPrecio` y usuario en `UpdateProductInput` |
| `backend/internal/infrastructure/product_repository.go` | Escribir precio de simple y variantes; precio de variantes nuevas |
| `backend/internal/application/product_service.go` | Permiso `precios.autorizar` cuando cambia el precio |
| `backend/internal/interfaces/http/product_handler.go` | Sucursal activa, usuario y permiso |
| Pruebas en `application` e `infrastructure` | Nuevas |
| `frontend/.../productos/productos.component.{ts,html}` | Enviar sucursal; campo motivo; precio deshabilitado sin permiso |
| `frontend/.../productos/product.models.ts` y spec | Campos nuevos y pruebas |

Tamaño aproximado: chico a mediano.

## Pruebas y criterios de aceptación
- Editar un producto simple cambiando el precio: el listado muestra el precio nuevo en la sucursal activa, el historial tiene un evento `AJUSTE_PRECIO_MANUAL` con motivo, y otra sucursal no cambia.
- Editar sin cambiar el precio: no se crea evento.
- Producto «sin precio» en la sucursal: al editar con precio se crea su precio vigente.
- Variante existente: su precio se guarda igual que un simple.
- Variante nueva al editar y conversión de simple a familia: la variante queda con el precio capturado en la sucursal activa.
- Sin `precios.autorizar`: campo deshabilitado; el backend responde `403` si recibe un precio distinto; editar nombre u otros datos sigue funcionando.
- `go test ./internal/...`, `ng test --include src/app/features/productos` y `ng build` en verde (salvo fallas previas registradas).
- Prueba manual: editar precio, verlo en la lista y en el historial de Validar costos y precios.

## Tareas
- [x] Aprobación del usuario (2026-10-22).
- [x] Decidir la relación con la 19 (2026-10-22: parte de edición trasladada).
- [x] Backend: sucursal y motivo en la edición.
- [x] Backend: guardar precio de simple y variantes existentes con historial.
- [x] Backend: precio de variantes nuevas al editar.
- [x] Backend: permiso `precios.autorizar` al cambiar precio.
- [x] Frontend: sucursal, motivo y campo deshabilitado sin permiso.
- [x] Pruebas, registro de resultados y base `registro-productos.md`.
- [x] Prueba manual del usuario (2026-10-22): correcta.
- [x] Aceptación final del usuario (2026-10-23).

## Decisiones
- 2026-10-22: el usuario confirmó la observación 1 de la 26 (el precio no se guarda al editar).
- Propuesta: guardar el precio en lugar de deshabilitar el campo, porque el usuario espera poder cambiarlo desde la edición.
- Propuesta: reutilizar el evento `AJUSTE_PRECIO_MANUAL`, para que el historial de la 19 no tenga huecos.
- 2026-10-22: el usuario decidió que el motivo sea opcional (por defecto «Edición del producto») y que solo el administrador, con `precios.autorizar`, pueda cambiar el precio.
- 2026-10-22 (implementación): una variante nueva con precio 0 queda «sin precio» en la sucursal; un producto sin precio vigente que se guarda con precio 0 tampoco crea fila. Así no aparecen precios en cero que nadie capturó.
- 2026-10-22 (implementación): el handler resuelve `precios.autorizar` con `RolService.Autorizar` (interfaz `permisoAutorizador` en `ProductHandler`) y lo pasa como `PuedeCambiarPrecio`; el repositorio decide si hubo cambio dentro de la transacción y devuelve `domain.ErrPricePermissionRequired` (HTTP `403`). Una edición rechazada no deja ningún cambio (ni nombre ni precio).
- 2026-10-22 (implementación): el cambio de precio recalcula `margen` cuando hay costo vigente. El motivo admite hasta 200 caracteres (la columna admite 500).
- 2026-10-22 (implementación): sin permiso, el formulario conserva el precio vigente de las variantes existentes y pone en 0 (sin precio) el de las variantes nuevas, para que agregar variantes no requiera el permiso.
- 2026-10-22 (implementación): habilitar o deshabilitar el precio base no emite `valueChanges`; de lo contrario `applyBasePrice` reescribía el precio de todas las variantes al abrir la edición (lo detectó una prueba nueva).

## Pruebas ejecutadas (2026-10-22)

- `go build ./...` y `go vet` de `interfaces/http`, `infrastructure`, `domain` y `cmd`: correctos.
- Integración PostgreSQL (`SECURITY_TEST_DATABASE_URL` → base local nueva `tienda_security_test` en Postgres.app, puerto 5432): `TestProductEditWritesBranchPrice` correcta, con estos casos:
  - cambio con permiso: precio redondeado a 2 decimales, margen recalculado, un evento `AJUSTE_PRECIO_MANUAL` con «Edición del producto» y usuario; la sucursal B no cambia;
  - mismo precio: sin evento; motivo capturado: se guarda recortado;
  - sin permiso: `ErrPricePermissionRequired` y la transacción no deja cambios; sin permiso y sin cambio de precio: correcto;
  - producto sin precio en la sucursal: con 0 no crea fila; con precio crea `ALTA_INICIAL` con costo 0;
  - variante existente con cambio de precio, variante existente sin cambio, variante nueva con precio y variante nueva con 0 (sin precio); variante nueva con precio sin permiso: rechazada;
  - sucursal de otro negocio y motivo de más de 200 caracteres: rechazados.
- `TestCatalogVariantsTenancyIsolation` ajustada (ahora envía sucursal y permiso) y correcta. Toda `go test ./internal/...` con la base de integración: correcta salvo `TestProductionAcceptsStrongOTPSecret`, falla previa registrada.
- `ng test --include src/app/features/productos` (Node 22): 41/41. Pruebas nuevas: payload con `sucursal_id` y motivo opcional recortado; sin motivo cuando el precio no cambia; sin permiso, precio bloqueado y variantes nuevas sin precio; al abrir una familia se conserva el precio de cada variante.
- `ng test` completo: 244/250; las 6 fallas son las previas de `alta-inicial.spec.ts` (`window.matchMedia`).
- `ng build`: correcto, con avisos de presupuesto previos.
- No hay prueba automática del handler (mapeo a `403` y `400` sin sucursal); se cubre con la prueba manual.
- `graphify update .`: correcto (sin cambios de topología en el grafo).

## Prueba manual (2026-10-22)

- El usuario confirmó: «ya se puede editar de manera correcta el producto». La edición guarda el precio.
- No se reportó por separado la prueba como colaborador sin `precios.autorizar` (precio bloqueado).

## Pendientes

- Aceptación final del usuario (con ella, la 29 pasa a Cerrada y se archiva en `historial/`).

## Cierre (2026-10-23)

- Aceptación final del usuario: «acepto la 22, 23, 27, 28, 29 y 31». Cerrada y archivada en `historial/`.
