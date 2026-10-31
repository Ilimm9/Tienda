# Capability 24: Correcciones a la carga masiva de productos

- Estado: **Cerrada** el 2026-10-22 con aceptación final del usuario («la 24 está aceptada»), tras la prueba manual.
- Aprobación explícita: 2026-10-22, el usuario aprobó los puntos 1 a 6 («procede con la implementación»), con estos ajustes: el costo se redondea a 2 decimales (punto 6) y el stock negativo se sigue rechazando.
- Fecha: 2026-10-22.
- Origen: prueba local con `thrs/docs-prueba/06-productos-lista-clienta.xlsx` (948 productos de la clienta): 505 correctos, 443 filas con error y 80 advertencias se muestran como listas largas sin orden; la vista pide una sucursal que ya está en el contexto.
- Diseño: canvas «Tienda — Alta inicial e invitaciones», artboard **H1 · Carga masiva: vista previa** (interactivo, con datos reales del archivo 06). Colores provisionales: los actuales.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Frontend `features/productos`; backend en `application` (importación de productos), `domain/product_view.go` y `interfaces/http/product_handler.go`. Sin cambios de esquema.

## Objetivo

Que la vista previa y el resultado de la carga masiva se lean de un vistazo aunque el archivo tenga cientos de filas, que los errores se puedan corregir en Excel y volver a subir, y que la sucursal se tome del contexto.

## Alcance y reglas

### 1. Sucursal desde el contexto (solo frontend)
- Se quita el select «Sucursal para el inventario inicial».
- La importación usa siempre `contexto.sucursal()`; se muestra como texto: «Se importará en **{sucursal}**. Para cargar en otra, cámbiala en el menú superior.»
- Si cambia la sucursal del contexto con una vista previa abierta, se descarta la vista previa y hay que validar de nuevo (evita importar en otra sucursal de la que se validó).
- El backend no cambia: sigue recibiendo `sucursal_id` y validando que sea una sucursal activa del usuario.

### 2. Secciones compactables
- Resumen con 4 tarjetas: Listos para importar, Filas con error, Advertencias, Catálogos nuevos. Al hacer clic en una, se abre su sección.
- Cuatro secciones compactables, cada una con su encabezado (chevron, título, contador y descripción corta), un `<button aria-expanded aria-controls>`:
  - **Errores**: abierta por defecto si hay errores. Filtros por columna (Todos, Stock inicial, Costo, Código de barras… se generan a partir de los errores recibidos). Tabla con Fila, Producto, Columna, Valor en el archivo y Qué corregir.
  - **Advertencias**: compactada. Incluye el aviso de `costos_faltantes`, que hoy se muestra en rojo como si fuera un error.
  - **Catálogos nuevos**: compactada.
  - **Correctos**: compactada. Tabla con Fila, Producto, Categoría, Precio, Costo y Stock.
- Una sección sin elementos no se muestra.
- La misma presentación aplica al resultado final de la importación (errores y advertencias).

### 3. Paginación
- Más de 10 filas: paginación de 10 en 10 del lado del cliente, con «1–10 de N», anterior/siguiente y números de página abreviados (1 … 4 5 6 … 45).
- 10 filas o menos: sin paginación.
- Cambiar el filtro de columna regresa a la página 1.

### 4. Descargar errores en Excel
- Botón «Descargar errores (.xlsx)» en la sección Errores.
- Nuevo endpoint `POST /negocios/:negocioId/catalogo/productos/importacion/errores` (multipart: `archivo` y `sucursal_id`; mismo grupo y permiso que `validar-importacion`). Vuelve a validar el archivo (lectura más reglas de BD) y responde un XLSX con:
  - solo las filas inválidas u omitidas, con sus valores originales en el formato de la plantilla de 14 columnas;
  - una columna final **Errores** con los motivos de la fila separados por `; `;
  - el nombre `errores-<archivo original>.xlsx`.
- El archivo descargado se puede corregir y subir tal cual: la lectura solo valida los primeros 14 encabezados, así que la columna extra se ignora (verificado en `readImportRowsWithOptions`).
- Sin estado en el servidor: no se guarda el archivo entre la validación y la descarga.

### 5. Datos nuevos en la vista previa (backend)
- `CatalogImportIssue` agrega `nombre` (`omitempty`) para mostrar el producto en errores y advertencias. Las importaciones de catálogos no cambian.
- `ProductImportPreview` agrega `filas_correctas`: `[{fila, nombre, categoria, precio_venta, costo, stock_inicial}]`, solo de las filas insertables. Con 948 filas el payload crece unos 60 KB; se acepta.

### 6. Costo redondeado a 2 decimales
- El Excel de la clienta guarda `8.380000000000001` aunque se vea `8.38`; antes se rechazaba por tener más de 4 decimales (20 filas del archivo 06).
- Decisión del usuario (2026-10-22): el costo de la importación se redondea a 2 decimales. Cualquier número finito no negativo se acepta y se redondea (`8.380000000000001` → 8.38; `12.12345` → 12.12). Se siguen rechazando el texto, los negativos y los valores fuera de rango (más de 10 dígitos enteros).
- Modifica la regla de precisión de la importación de la [capability 19](../19-historial-costos-precios-sucursal.md) (antes: hasta 4 decimales o error). La instrucción de la plantilla se actualiza.

### 7. Stock negativo
- Decisión del usuario (2026-10-22): se sigue rechazando con el error actual. Sin cambios de código.

## Fuera de alcance
- Proveedor por defecto para que los productos importados aparezcan en Registrar compra.
- Cambio de colores; el diseño de H1 usa la paleta actual.
- Paginación o filtros del lado del servidor.

## Contratos
- `POST …/catalogo/productos/validar-importacion`: respuesta + `filas_correctas[]`; `errores[]` y `advertencias[]` + `nombre`.
- `POST …/catalogo/productos/importacion/errores` → `200` con `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`; `400` si el archivo no es válido; `404` si la sucursal no es del usuario; `409` o mensaje si no hay errores que descargar.
- Frontend: `ProductImportIssue.nombre?`, `ProductImportPreview.filas_correctas`.

## Archivos estimados
| Archivo | Cambio |
|---|---|
| `frontend/.../productos/productos.component.html` | Quitar el select de sucursal; resumen; 4 secciones; paginador |
| `frontend/.../productos/productos.component.ts` | Estado abierto/cerrado, página y filtro por sección; sucursal desde el contexto; descarga |
| `frontend/.../productos/productos.component.css` | Estilos de tarjetas, secciones y paginador |
| `frontend/.../productos/productos.service.ts` | `downloadImportErrors()` como `Blob` |
| `frontend/.../productos/product.models.ts` | Campos nuevos |
| `frontend/.../productos/productos.component.spec.ts` | Pruebas nuevas |
| `frontend/.../productos/import-report.component.{ts,html,css,spec.ts}` | **Nuevo** (ver Decisiones): secciones, filtros y paginación |
| `frontend/src/styles/uploads.css` | Estilo del aviso de sucursal (`.import-branch`) |
| `backend/internal/application/product_service_test.go` | Prueba de vista previa y descarga con repositorio simulado |
| `backend/internal/domain/product_view.go` | `Nombre` en issue; `FilasCorrectas` en preview |
| `backend/internal/application/catalog_import.go` y `product_service.go` | Llenar `nombre` y `filas_correctas`; generar el XLSX de errores; (6) redondeo del costo |
| `backend/internal/interfaces/http/product_handler.go` y `cmd/api/main.go` | Handler y ruta de descarga |
| `backend/internal/application/catalog_import_test.go` | Pruebas nuevas |

Tamaño aproximado: frontend mediano (la mayor parte del trabajo), backend chico.

## Pruebas y criterios de aceptación
- Backend:
  - la vista previa del archivo 06 devuelve `filas_correctas` con 505 elementos y `nombre` en cada error;
  - el XLSX de errores contiene solo filas inválidas u omitidas, 14 columnas más `Errores`, y al subirlo se vuelven a mostrar exactamente esos errores;
  - (6) `8.380000000000001` se importa como 8.38 y `12.12345` como 12.12; `-1` y texto se rechazan.
- Frontend:
  - no aparece el select de sucursal y la petición usa la sucursal del contexto;
  - las secciones se abren y cierran con `aria-expanded`, y Errores empieza abierta;
  - con más de 10 filas hay paginación y con 10 o menos no;
  - el filtro por columna reinicia la página;
  - cambiar la sucursal del contexto descarta la vista previa;
  - el botón descarga el archivo.
- `go test ./internal/...`, `ng test --include src/app/features/productos` y `ng build` en verde.
- Prueba manual con el archivo 06 en escritorio y en móvil.

## Tareas
- [x] Aprobación del usuario (2026-10-22).
- [x] Backend: `nombre` y `filas_correctas`.
- [x] Backend: endpoint de descarga de errores.
- [x] (6) Costo redondeado a 2 decimales.
- [x] Frontend: sucursal desde el contexto.
- [x] Frontend: resumen, secciones, filtros y paginación.
- [x] Frontend: descarga.
- [x] Pruebas y registro de resultados.
- [x] Prueba manual y aceptación final (2026-10-22).

## Decisiones
- 2026-10-22: el usuario pidió secciones compactables (correctos, errores, advertencias), descargar los errores en Excel, paginar a partir de unas 10 filas y tomar la sucursal del contexto.
- Propuesta: el Excel de errores reutiliza la plantilla para que el ciclo «descargar, corregir, subir» no requiera copiar datos.
- Propuesta: se agrega «Catálogos nuevos» como cuarta sección compactable porque en el archivo 06 ocupa 29 filas.
- 2026-10-22 (implementación): las secciones viven en un componente nuevo `productos/import-report.component`, que se usa tanto en la vista previa como en el resultado; así no se duplica el marcado en `productos.component.html`. No cambia el alcance.
- 2026-10-22 (implementación): los tres handlers de importación de productos (vista previa, importar y descarga de errores) comparten `productImportUpload`, que valida la sucursal y lee el archivo con los mismos mensajes que antes.
- 2026-10-22 (implementación): si falla la descarga se muestra un aviso (`FeedbackService`) y no `importError`, para no ocultar la vista previa.
- 2026-10-22 (revisión visual del usuario: «la parte superior se ve algo fea»): la parte de arriba se acomoda como en el diseño H1. «Descargar plantilla» pasa a la tarjeta de intro; con archivo elegido, la sucursal y el archivo quedan lado a lado y la zona de arrastre se reduce a una fila (`.file-dropzone--compact`, solo en productos; Catálogo no cambia); el recuadro «Vista previa de importación» se cambia por un encabezado sin fondo («517 se insertarían · 2 omitidas por duplicado»), porque las tarjetas ya muestran los conteos. `ng test --include src/app/features/productos` 37/37 y `ng build` correctos.

## Pruebas ejecutadas (2026-10-22)

- `go build ./...` y `go vet` de `application`, `interfaces/http` y `cmd`: correctos.
- `go test ./internal/application/`: correcto. Pruebas nuevas o ajustadas:
  - `TestParseProductImportRoundsCostToTwoDecimals`: `8.380000000000001` → 8.38, `12.12345` → 12.12, `25.1234` → 25.12; negativo y texto se rechazan; los errores traen `nombre`.
  - `TestProductImportErrorsWorkbookKeepsOnlyRejectedRows`: solo filas rechazadas, 14 columnas más `Errores` con los motivos unidos por `; `, y el archivo se vuelve a leer ignorando la columna extra.
  - `TestPreviewProductImportListsValidRowsAndNamesErrors`: `filas_correctas` y `nombre` también en errores de BD; `ErrNoImportErrors` si no hay errores.
  - Reemplaza `TestParseProductImportRequiresValidCost` (regla anterior de 4 decimales).
- `go test ./...`: correcto salvo `internal/config` `TestProductionAcceptsStrongOTPSecret`, falla previa ya registrada.
- Archivo real `thrs/docs-prueba/06-productos-lista-clienta.xlsx` (lectura, sin BD): 519 válidas (antes 505), 429 inválidas (428 por stock negativo y 1 por el código `P1`); ya no hay errores de costo. El Excel de errores trae las 429 filas y al leerlo de nuevo da 429 inválidas.
- `ng test --include src/app/features/productos` (Node 24): 37/37 correctas. Pruebas nuevas: sin selector de sucursal y `sucursal_id` tomado del contexto; descartar la vista previa al cambiar de sucursal (no con el mismo ID); descarga del Excel; secciones con `aria-expanded`; paginación solo con más de 10; filtro que regresa a la página 1; costos faltantes como advertencia.
- `ng test` completo: 234/240. Las 6 fallas están en `alta-inicial.spec.ts` (`window.matchMedia is not a function`, del `p-select` del alta inicial), son previas y ajenas a esta capability; ningún archivo de `alta-inicial` cambió.
- `ng build`: correcto. Aviso de presupuesto: `import-report.component.css` mide 5.19 kB (aviso a 4 kB, error a 16 kB), igual que otros componentes.
- `graphify update .`: no se ejecutó porque `graphify` no está instalado en este equipo.

## Prueba manual (2026-10-22)

Hecha por el usuario con la lista de pasos A–G (sucursal desde el contexto, resumen y secciones, paginación y filtros, costo redondeado, descargar errores y volver a subir, importación y resultado, móvil): todo correcto.

- A: el selector de sucursal solo está en la barra superior. Al cambiar de sucursal la aplicación redirige a Inicio y recarga las vistas, así que la vista previa no persiste con otra sucursal; el riesgo de importar en una sucursal distinta a la validada no se presenta.
- Observación surgida en la prueba, fuera de esta capability: en otra sucursal se ve todo el catálogo del negocio (agotado y sin precio) y se percibe como duplicado. Se atiende en la [capability 28](28-surtido-por-sucursal.md).

## Pendientes

- Ninguno de esta capability. Comportamiento conocido: la descarga vuelve a validar contra la BD; si el catálogo cambió entre la vista previa y la descarga, el Excel refleja el estado actual.
