# Capability 18: Variantes para productos repetidos en importación

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-16 mediante la instrucción: “Implement the plan”.

## Alcance

- Identificar el caso de un producto simple repetido por nombre y presentación con código de barras distinto.
- Solicitar un atributo en la columna `Variantes` en lugar de tratarlo como un segundo producto simple.
- Actualizar las instrucciones de la plantilla XLSX.

## Reglas y contratos

- Un producto simple conserva la identidad por nombre y presentación normalizados.
- Un código distinto para esa identidad requiere `Variantes` con el formato `Atributo: valor`.
- Un código de barras repetido conserva su error independiente.
- Las rutas, columnas y modelo de variantes no cambian.

## Tareas y decisiones

- [x] Confirmar que las variantes deben declararse en el XLSX y no inferirse del código.
- [x] Ajustar el mensaje de validación y la plantilla.
- [x] Agregar prueba y ejecutar verificaciones.

## Criterios de aceptación

- Un segundo código para el mismo producto simple muestra un error en `Variantes` con la instrucción requerida.
- Filas con atributos de variante mantienen el flujo de familia existente.
- Un código de barras idéntico sigue reportándose como duplicado.

## Resultados

- `go test ./internal/application ./internal/infrastructure`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,933 nodos y 8,313 relaciones. La herramienta informó cambios de comunidades y archivos no clasificables, sin fallos de extracción.
