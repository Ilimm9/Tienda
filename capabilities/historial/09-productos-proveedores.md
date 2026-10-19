# Capability 09: Proveedor opcional en productos

## Estado

En implementación.

## Aprobación

Aprobada explícitamente por el usuario el 2026-09-23 mediante la instrucción: “Implement the plan”.

## Alcance

- Asociar opcionalmente un proveedor activo del mismo negocio a cada producto.
- Mostrar y filtrar productos por proveedor, incluido el filtro «Sin proveedor».
- Permitir seleccionar el proveedor al crear y editar productos.
- Agregar la columna opcional `Proveedor` a la plantilla e importación XLSX de productos.
- Aplicar el proveedor de una familia a todas sus variantes comerciales.

## Reglas y contratos

- Un producto comercial tiene como máximo un proveedor; un proveedor puede pertenecer a varios productos.
- `proveedor_id` es opcional en creación y edición; si se envía debe ser un proveedor activo del negocio actual.
- La respuesta del listado de productos incorpora `proveedor` y `proveedor_id`.
- La columna final de la plantilla XLSX es `Proveedor`; se resuelve por nombre normalizado contra proveedores activos del negocio.
- Un proveedor inexistente vuelve inválida únicamente su fila XLSX. Una celda vacía conserva la asociación como ausente.
- No se migran ni reinterpretan relaciones históricas: la base se considera nueva para esta capability.

## Tareas y decisiones

- [x] Definir relación opcional de un proveedor por producto y alcance para variantes.
- [x] Confirmar importación XLSX por nombre e invalidez por fila desconocida.
- [x] Exponer proveedor en los contratos de producto y persistir la relación.
- [x] Integrar selector, columna y filtros de Productos.
- [x] Extender plantilla, parser y validación de importación XLSX.
- [x] Ejecutar pruebas, actualización del grafo y registrar resultados.

## Criterios de aceptación

- Un producto puede guardarse con o sin proveedor, y nunca con un proveedor ajeno/inactivo.
- El listado muestra el proveedor y los filtros distinguen proveedor concreto de «Sin proveedor».
- Crear o importar una familia deja el mismo proveedor en todas sus variantes.
- La plantilla incluye `Proveedor`; una fila que lo nombre incorrectamente se reporta inválida sin bloquear las demás.

## Resultados

- `go test ./...`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check` y `git diff --cached --check`: correctos.
- `graphify update .`: correcto; actualizó la extracción estructural del proyecto. El grafo reportó que su conjunto de comunidades cambió y conserva etiquetas previas para 63 comunidades; es una advertencia de rotulado, no un fallo de integridad.
- Pendiente conocido: las pruebas de Angular y el build de producción requieren Node `v22.22.3` o superior, mientras el entorno dispone de Node `v14.15.0`.
