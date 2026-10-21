# Capability 17: Alta autorizada de catálogos durante importación

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-16 mediante la instrucción: “Implement the plan”.

La corrección para devolver arreglos vacíos, en vez de `null`, en la vista previa fue aprobada explícitamente por el usuario el 2026-10-21 mediante “Implement the plan”.

La aclaración de `Contenido`, `Unidad de contenido` y `Unidad de medida` en la hoja `Instrucciones` fue aprobada explícitamente por el usuario el 2026-10-21 mediante “Implement the plan”.

## Alcance

- Detectar marcas, categorías, unidades de medida y proveedores nuevos al validar un XLSX de productos.
- Mostrar los valores nuevos agrupados en una tabla antes de importar.
- Requerir una autorización global para crear esos catálogos y registrar los productos asociados.
- Explicar en la plantilla la diferencia entre la cantidad y símbolo del empaque, y la unidad estandarizada del catálogo.

## Reglas y contratos

- Marca, categoría, unidad y proveedor siguen siendo opcionales: una celda vacía no crea ni asocia catálogo.
- La vista previa expone `catalogos_nuevos` con tipo, nombre y cantidad de productos que lo usan.
- `catalogos_nuevos`, `errores` y `advertencias` siempre se devuelven como arreglos; cuando no contienen elementos, su valor es `[]`, nunca `null`.
- Si existen catálogos nuevos, la importación requiere `autorizar_catalogos_nuevos=true`; sin ella no se persiste nada.
- Marcas, categorías raíz y proveedores nuevos se crean solo con nombre.
- Una unidad nueva usa tipo `CANTIDAD`, factor `1`, permite fracción, dos decimales y un símbolo técnico único autogenerado; podrá editarse después.
- La hoja `Instrucciones` aclara que `Contenido` y `Unidad de contenido` describen lo que trae el empaque (por ejemplo, `600` y `ml`), mientras `Unidad de medida` usa el nombre completo de la unidad del catálogo (por ejemplo, `Mililitro`).
- Los duplicados, formatos inválidos y códigos de barras inválidos conservan sus validaciones.

## Tareas y decisiones

- [x] Confirmar autorización global y campos de catálogo opcionales.
- [x] Definir unidad mínima neutra y editable.
- [x] Extender contratos de vista previa e importación.
- [x] Integrar tabla y confirmación en Productos.
- [x] Ejecutar pruebas y actualizar el grafo.
- [x] Normalizar listas vacías en la vista previa y proteger la interfaz contra respuestas heredadas con `null`.
- [x] Aclarar en la plantilla la diferencia entre unidad de contenido y unidad de medida, con un ejemplo de botella de 600 ml.

## Criterios de aceptación

- La previsualización lista cada catálogo nuevo una sola vez con su número de productos.
- Sin autorización no se importan productos que requieran catálogos nuevos.
- Con autorización se crean los catálogos y los productos quedan asociados a ellos.
- Los productos con referencias de catálogo vacías se importan sin asociación.
- La plantilla descargable explica, sin ambigüedad, cómo llenar `Contenido`, `Unidad de contenido` y `Unidad de medida`.

## Resultados

- `go test ./internal/application ./internal/infrastructure ./internal/interfaces/http`: correcto.
- `go test ./...` desde `backend/`: los paquetes relacionados son correctos; conserva un fallo preexistente en `internal/config`, `TestProductionAcceptsStrongOTPSecret`.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,931 nodos y 8,308 relaciones. La herramienta informó cambios de comunidades y 57 archivos no clasificables, sin fallos de extracción.
- `prettier --check` sobre los archivos de Productos: reporta diferencias de formato históricas en esos archivos; no se reformataron para evitar una reescritura masiva no relacionada.
- `go test ./internal/application -run 'Test.*Import'`: correcto tras normalizar listas vacías de vista previa.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.app.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,958 nodos y 8,395 relaciones.
- `go test ./internal/application -run TestProductImportTemplateIncludesInstructions`: correcto; verifica las nuevas aclaraciones de unidades por etiqueta.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,958 nodos y 8,395 relaciones.
