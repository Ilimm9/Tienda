# Capability 16: Código de barras opcional en importación

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-16 mediante la instrucción: “Implement the plan”.

## Alcance

- Conservar la columna `Código de barras` como parte obligatoria de la plantilla XLSX de productos.
- Tratar como ausencia de código las celdas que se ven vacías pero contienen solamente espacios o marcas Unicode invisibles.
- Crear esos productos sin un registro asociado en `producto_codigos`.

## Reglas y contratos

- Un archivo sin el encabezado `Código de barras` continúa siendo inválido.
- Un valor vacío, espacios o solo marcas invisibles no crea código de barras ni advertencia.
- Los códigos con contenido visible conservan las validaciones vigentes: solo de 1 a 14 dígitos, advertencia si tienen menos de 8 y control de duplicados.
- No cambian endpoints, contratos HTTP, base de datos, plantilla ni formulario manual de productos.

## Tareas y decisiones

- [x] Confirmar que el encabezado debe mantenerse y que la fila 327 presenta una celda visualmente vacía.
- [x] Limitar la normalización a la columna de código de barras durante la importación de productos.
- [x] Definir que las marcas invisibles solo equivalen a vacío si no hay contenido visible en la celda.
- [x] Agregar prueba de regresión y ejecutar verificaciones focalizadas.
- [x] Actualizar el grafo y registrar resultados.

## Criterios de aceptación

- Una celda vacía o formada solo por espacios o marcas invisibles importa el producto sin código de barras.
- Un código no numérico visible sigue produciendo el error de validación correspondiente.
- La ausencia del encabezado sigue rechazando el archivo.
- La importación de códigos numéricos válidos conserva su comportamiento actual.

## Resultados

- `go test ./internal/application -run 'TestParseProductImport' -count=1`: correcto.
- `go test ./...` desde `backend/`: los paquetes relacionados y la nueva prueba son correctos. Conserva un fallo preexistente en `internal/config`: `TestProductionAcceptsStrongOTPSecret` rechaza su secreto OTP de prueba.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2,925 nodos y 8,291 relaciones. La herramienta informó cambios de comunidad y omitió 57 archivos sin extensión compatible, sin fallos de extracción.
