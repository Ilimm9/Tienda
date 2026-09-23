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
- [ ] Exponer proveedor en los contratos de producto y persistir la relación.
- [ ] Integrar selector, columna y filtros de Productos.
- [ ] Extender plantilla, parser y validación de importación XLSX.
- [ ] Ejecutar pruebas, actualización del grafo y registrar resultados.

## Criterios de aceptación

- Un producto puede guardarse con o sin proveedor, y nunca con un proveedor ajeno/inactivo.
- El listado muestra el proveedor y los filtros distinguen proveedor concreto de «Sin proveedor».
- Crear o importar una familia deja el mismo proveedor en todas sus variantes.
- La plantilla incluye `Proveedor`; una fila que lo nombre incorrectamente se reporta inválida sin bloquear las demás.

## Resultados

Pendiente de verificaciones al completar la implementación.
