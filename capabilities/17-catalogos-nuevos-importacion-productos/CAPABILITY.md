# Capability 17: Alta autorizada de catálogos durante importación

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-16 mediante la instrucción: “Implement the plan”.

## Alcance

- Detectar marcas, categorías, unidades de medida y proveedores nuevos al validar un XLSX de productos.
- Mostrar los valores nuevos agrupados en una tabla antes de importar.
- Requerir una autorización global para crear esos catálogos y registrar los productos asociados.

## Reglas y contratos

- Marca, categoría, unidad y proveedor siguen siendo opcionales: una celda vacía no crea ni asocia catálogo.
- La vista previa expone `catalogos_nuevos` con tipo, nombre y cantidad de productos que lo usan.
- Si existen catálogos nuevos, la importación requiere `autorizar_catalogos_nuevos=true`; sin ella no se persiste nada.
- Marcas, categorías raíz y proveedores nuevos se crean solo con nombre.
- Una unidad nueva usa tipo `CANTIDAD`, factor `1`, permite fracción, dos decimales y un símbolo técnico único autogenerado; podrá editarse después.
- Los duplicados, formatos inválidos y códigos de barras inválidos conservan sus validaciones.

## Tareas y decisiones

- [x] Confirmar autorización global y campos de catálogo opcionales.
- [x] Definir unidad mínima neutra y editable.
- [x] Extender contratos de vista previa e importación.
- [x] Integrar tabla y confirmación en Productos.
- [x] Ejecutar pruebas y actualizar el grafo.

## Criterios de aceptación

- La previsualización lista cada catálogo nuevo una sola vez con su número de productos.
- Sin autorización no se importan productos que requieran catálogos nuevos.
- Con autorización se crean los catálogos y los productos quedan asociados a ellos.
- Los productos con referencias de catálogo vacías se importan sin asociación.

## Resultados

- `go test ./internal/application ./internal/infrastructure ./internal/interfaces/http`: correcto.
- `go test ./...` desde `backend/`: los paquetes relacionados son correctos; conserva un fallo preexistente en `internal/config`, `TestProductionAcceptsStrongOTPSecret`.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,931 nodos y 8,308 relaciones. La herramienta informó cambios de comunidades y 57 archivos no clasificables, sin fallos de extracción.
- `prettier --check` sobre los archivos de Productos: reporta diferencias de formato históricas en esos archivos; no se reformataron para evitar una reescritura masiva no relacionada.
