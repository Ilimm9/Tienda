# Capability 27: Compras con cualquier producto y varios proveedores

- Estado: **Cerrada** el 2026-10-23 con aceptación final del usuario.
- Aprobación explícita: 2026-10-22, el usuario pidió «continúa con la corrección del punto dos en la capability 26» y aprobó el plan con este alcance.
- Fecha: 2026-10-22.
- Origen: observación 2 de la [26](26-observaciones-ilian.md): Registrar compra no muestra productos sin proveedor.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Backend en `compra` (`application`, `infrastructure`) y productos (`domain`, `infrastructure`); frontend `features/compras`. Cambio de esquema menor: índices de `producto_proveedor`.

## Causa (verificada en código)

- `compra.Repository.ListProductos` (`backend/internal/infrastructure/compra/compra_repository.go:34`) solo lista productos con fila en `producto_proveedor` para el proveedor elegido, y `buildDetail` rechaza con «el producto no está asociado al proveedor».
- El vínculo solo se crea al importar con la columna Proveedor llena, al editar el producto o al crear el producto desde la compra. Los 948 productos de la clienta vienen sin proveedor, así que ninguno aparece en Compras.
- `ProductoProveedor` tiene un índice único por `producto_negocio_id` (`idx_producto_proveedor_unico`): un producto solo puede tener un proveedor. `replaceProductProvider` borra todos los vínculos al crear, editar o importar.

## Objetivo

Que en Registrar compra se pueda comprar cualquier producto del negocio a cualquier proveedor, y que el producto quede ligado a cada proveedor que lo surte.

## Alcance y reglas

### 1. Varios proveedores por producto
- Un producto comercial (`producto_negocio`) puede tener varios proveedores; el par producto-proveedor no se repite.
- Como máximo un proveedor es **principal**. Es el que se muestra y se edita en Productos.
- Se elimina el índice único `idx_producto_proveedor_unico` y se crea `idx_producto_proveedor_principal` (único por producto donde `es_principal`). Migración idempotente; los datos actuales ya cumplen ambas reglas.

### 2. Productos: proveedor principal sin borrar los demás
- Crear, editar o importar un producto fija el proveedor principal: el elegido se marca como principal (se crea el vínculo si no existía) y los demás dejan de ser principales, pero se conservan.
- Si el proveedor va vacío, el producto queda sin principal y conserva los proveedores que ganó por compras.

### 3. Selector de Registrar compra
- Muestra todos los productos activos del negocio, en dos grupos: «De este proveedor» primero y «Otros productos» después, cada uno ordenado por nombre. «Crear nuevo…» se mantiene.
- Tiene buscador por nombre, SKU y código de barras.
- Si el producto elegido es de «Otros productos», se avisa «Se ligará a este proveedor al guardar».

### 4. Vínculo al guardar la compra
- Dentro de la transacción de la compra, cada producto comprado queda ligado al proveedor de la compra si no lo estaba. Es principal solo si el producto no tenía principal.
- Si ya estaba ligado, no cambia nada. El principal existente nunca cambia por una compra.

## Fuera de alcance
- Ver o administrar los proveedores secundarios en Editar producto. **Decidido por el usuario el 2026-10-22:** en Productos solo se ve y se edita el principal; los demás funcionan en Compras.
- Costo o código por proveedor.
- Surtido por sucursal ([28](28-surtido-por-sucursal.md)). Una compra ya crea el `inventario_sucursal` de su sucursal, así que es compatible.

## Contratos
- `GET …/compras/productos?proveedor_id=…` → todos los productos activos del negocio; cada uno agrega `del_proveedor: bool`. Orden: `del_proveedor` primero, luego nombre.
- `POST …/compras` ya no responde «el producto no está asociado al proveedor»; crea el vínculo.
- Frontend: `PurchaseProduct.del_proveedor`.

## Archivos estimados
| Archivo | Cambio |
|---|---|
| `backend/internal/domain/catalog_models.go` | Quitar el índice único por producto |
| `backend/internal/infrastructure/catalog_tenant_migration.go` | Borrar índice viejo; índice de principal único |
| `backend/internal/infrastructure/product_repository.go` | `setPrincipalProvider` en lugar de `replaceProductProvider` |
| `backend/internal/application/compra/compra_service.go` | `ProductoComprable.DelProveedor` |
| `backend/internal/infrastructure/compra/compra_repository.go` | Listado completo; vínculo al guardar |
| `frontend/.../compras/compras.{models,component}.ts`, `.html`, spec | Grupos, buscador y aviso |
| Pruebas de integración en `backend/internal/database/` | Nuevas |

Tamaño aproximado: chico a mediano.

## Pruebas y criterios de aceptación
- Un producto sin proveedor aparece en Compras como «Otros productos»; los del proveedor salen primero.
- Comprar un producto sin proveedor lo liga como principal.
- Comprar el mismo producto a un segundo proveedor: la compra se guarda, se crea un vínculo no principal y el principal se conserva; el producto aparece en «De este proveedor» para ambos.
- Editar el producto cambiando el principal conserva el otro proveedor.
- La base admite dos proveedores por producto y rechaza dos principales.
- `go test ./internal/...`, `ng test` de compras y productos y `ng build` en verde (salvo fallas previas registradas).
- Prueba manual: compra con la lista de la clienta y con dos proveedores.

## Tareas
- [x] Aprobación del usuario (2026-10-22).
- [x] Esquema: índices de `producto_proveedor`.
- [x] Productos: proveedor principal sin borrar los demás.
- [x] Compras: listado completo con `del_proveedor` y vínculo al guardar.
- [x] Frontend: grupos, buscador y aviso.
- [x] Pruebas, registro de resultados y bases `modulo-compras.md` y `registro-productos.md`.
- [x] Prueba manual y aceptación final del usuario (2026-10-23).

## Decisiones
- 2026-10-22: el usuario eligió mostrar todos los productos, primero los del proveedor elegido, con buscador.
- 2026-10-22: el usuario pidió que el producto quede ligado al proveedor de la compra y que otro proveedor también pueda surtirlo, es decir, varios proveedores por producto.
- 2026-10-22: el usuario decidió que en Productos solo se vea y edite el proveedor principal.
- 2026-10-22 (implementación): `setPrincipalProvider` reemplaza a `replaceProductProvider` en crear, editar, variantes e importación: quita la marca de principal a los demás y hace upsert del elegido (`ON CONFLICT` sobre producto + proveedor).
- 2026-10-22 (implementación): `linkProvider` bloquea la fila de `producto_negocio` (`FOR UPDATE`) antes de ligar, para que dos compras simultáneas no elijan dos principales.
- 2026-10-22 (implementación): «Crear nuevo…» pasa al inicio del selector (grupo «Nuevo»), porque con cientos de productos al final quedaba fuera de vista.

## Pruebas ejecutadas (2026-10-22)

- `go build ./...` y `go vet ./internal/...`: correctos.
- Integración PostgreSQL (`tienda_security_test` local): `TestCompraProveedoresVarios` correcta:
  - el listado trae todos los productos, con los del proveedor primero y `del_proveedor` correcto;
  - comprar un producto sin proveedor lo liga como principal;
  - comprarlo a un segundo proveedor agrega un vínculo no principal; aparece como «del proveedor» para ambos; recomprar no duplica el vínculo;
  - editar el producto cambiando el principal conserva el otro proveedor; dejarlo vacío deja ambos sin principal; una compra a un proveedor ya ligado no lo vuelve principal;
  - la base acepta un segundo proveedor y rechaza un segundo principal.
- `go test ./internal/...` con la base de integración: correcta salvo `TestProductionAcceptsStrongOTPSecret`, falla previa registrada (incluye las pruebas de la 29 y de variantes).
- `ng test --include src/app/features/compras`: 11/11, con dos pruebas nuevas (grupos con el proveedor primero y aviso «Se ligará a este proveedor al guardar»; grupos vacíos omitidos).
- `ng test` completo: 240/246; las 6 fallas son las previas de `alta-inicial.spec.ts` (`window.matchMedia`).
- `ng build`: correcto.
- `graphify update .`: correcto.

## Corrección: selector lento (2026-10-22)

- El usuario reportó en su prueba: «funciona pero tarda mucho en darme una lista de los productos».
- Medición en la base local `tienda2` (2,430 productos activos en un negocio): `EXPLAIN ANALYZE` de `ListProductos` tarda 63 ms y usa índices. La consulta no es la causa.
- Causa: el `p-select` de producto dibujaba todas las opciones (unas 2,400 con grupos) al abrirse y en cada búsqueda.
- Cambio: scroll virtual en el selector (`virtualScroll`, `virtualScrollItemSize` 38, `scrollHeight` 320px); solo se dibujan las opciones visibles. Grupos, buscador y aviso no cambian.
- Pruebas: `ng test --include src/app/features/compras` 12/12 (nueva: el selector de producto usa scroll virtual); `ng test` completo 241/247, con las 6 fallas previas de `alta-inicial`; `ng build` correcto; `graphify update .` correcto.

## Pendientes

- Reiniciar el backend activo: al arrancar corre la migración de índices de `producto_proveedor`.
- Prueba manual:
  - el selector de productos abre y filtra sin pausa con todo el catálogo;
  - en Registrar compra, al elegir un proveedor aparecen todos los productos con buscador, primero «De este proveedor»;
  - comprar un producto importado sin proveedor: queda ligado y en Productos se ve ese proveedor;
  - comprar el mismo producto a otro proveedor: se guarda, aparece en «De este proveedor» para ambos y el proveedor en Productos no cambia.
- Aceptación final del usuario.

## Cierre (2026-10-23)

- Aceptación final del usuario: «acepto la 22, 23, 27, 28, 29 y 31». Cerrada y archivada en `historial/`.
