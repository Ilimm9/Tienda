# Productos: alta, carga masiva y "agregar del catálogo" en ventanas modales

Creado: 2026-11-11. Estado: pendiente de implementar.

## Objetivo

Que las tres pantallas de alta de productos se abran como **ventanas modales grandes encima de la lista de productos**, con el aire fresco del diseño propuesto, y corregir la lógica de variantes.

- Producto nuevo y Editar producto (`/catalogo/productos/nuevo`, `/catalogo/productos/:productoId/editar`)
- Agregar del catálogo del negocio (`/catalogo/productos/agregar-del-catalogo`)
- Carga masiva (`/catalogo/productos/importar`)

## Regla principal: no se retira ni se esconde nada

- **Ningún campo, botón, aviso ni comportamiento actual se elimina ni se oculta.** El diseño es una guía visual. Si el diseño no muestra algo que hoy existe, se conserva igual.
- Esto incluye: búsqueda por código de barras con imagen y fuentes, SKU interno, nombre, categoría con "+ Nueva", marca con "+ Nueva", proveedor, presentación, contenido, unidad de contenido con "Nueva unidad", precio base, motivo del cambio de precio, costo de proveedor, el switch de variantes, el editor de variantes, el stock y su edición, Sucursal (cuando hay más de una), descripción, Multimedia, el aviso "ya está en el catálogo del negocio" y los mensajes de error.
- Solo se reacomoda y se reestiliza. Única reubicación pedida: **Marca** pasa a "Detalles opcionales" junto a Descripción y a Sucursal, que se queda donde ya está.
- Obligatorios con `*`; opcionales con "(opcional)" en gris, como en el modal de Proveedor.
- Sin cambios de backend ni de esquema.

## Diseño de referencia

- Canvas: https://claude.ai/artifact/6yBhhs6XuSabAE6X61HBPT → página **Secciones**
- **Tablero 05 · Catálogo**:
  - En el botón "Agregar producto", la flecha abre el menú de "Producto nuevo", "Del catálogo del negocio" y "Carga masiva". Así se ven las ventanas encima de la lista.
  - Se prueba con Play o pantalla completa.
- **Tablero 12 · Propuesta: ventanas modales por dispositivo**:
  - Arriba, en "Ventana", elegir Producto nuevo, Del catálogo del negocio o Carga masiva de productos.
  - En "Pantalla", elegir Escritorio, Tablet o Móvil para ver cómo queda cada uno.
- El diseño es referencia: se cambia el código Angular para parecerse, nunca el diseño. El texto y los campos reales mandan sobre el diseño (ver regla principal).

## Archivos a tocar

| Archivo | Qué |
|---|---|
| `frontend/src/app/features/productos/productos.component.html` | Lista siempre visible; el formulario (create/edit) y la importación van dentro de `<app-modal>` |
| `frontend/src/app/features/productos/productos.component.ts` | Modo por ruta para abrir el modal correspondiente; cerrar = volver a `/catalogo/productos`; lógica de variantes (abajo) |
| `frontend/src/app/features/productos/productos.component.css` | Estilo nuevo del formulario dentro del modal |
| `frontend/src/app/features/productos/productos.routes.ts` | `agregar-del-catalogo` carga `ProductosComponent` con `data.mode: 'agregar'`. Mismas URL, guards, permisos y breadcrumbs |
| `frontend/src/app/features/productos/agregar-catalogo.component.*` | Se vuelve embebible dentro del modal: sin encabezado de página, emite `cerrar` y `agregado` en lugar de navegar |
| `frontend/src/app/features/productos/import-report.component.*` | Reporte de carga con contadores como pestañas (abajo) |
| `frontend/src/styles/componentes.css` | Clase única para selects (ver Selects) |
| `frontend/src/app/shared/ui/modal/modal.component.ts` | Tamaño nuevo y pantalla completa en móvil (abajo). Sin romper los usos actuales (`catalogo.component.html`) |
| `*.spec.ts` de los anteriores | Ajustar y agregar pruebas |

## Modales

### Comportamiento común

- Se usa el `ModalComponent` compartido (`app-modal`), con `[sucio]` para pedir confirmación si hay cambios y `[ocupado]` mientras se guarda.
- Las rutas actuales se conservan para que los enlaces directos y el botón atrás sigan funcionando:
  - Entrar a la ruta muestra la lista con el modal abierto.
  - Cerrar (X, fondo, Cancelar o Esc) navega a `/catalogo/productos`.
  - Guardar navega a `/catalogo/productos` y refresca la lista.
- Estructura del modal:
  - Encabezado blanco con título, subtítulo y X.
  - Cuerpo con scroll propio.
  - Pie fijo con nota a la izquierda y botones a la derecha (Cancelar / acción principal).
- Tamaños de `ModalComponent`:
  - Se agrega `amplio: 1180`.
  - Se agrega el input `completoEnMovil` (por defecto `false`). Con `true`, a ≤640 px el modal ocupa toda la pantalla sin bordes redondeados y los botones del pie se apilan a todo lo ancho.
  - Asignación: Producto nuevo y Editar → `grande` (940); Agregar del catálogo → `grande`; Carga masiva → `amplio`. Los tres con `completoEnMovil`.
  - Tablet: el modal ocupa casi toda la pantalla, con 16 px de margen.

### Estilo (aplica a los tres)

- Fondo blanco. **Sin cajas internas** (quitar el aspecto `family-card`, `barcode-lookup-card` y `shared-data-card` dentro del modal). Los bloques se separan con un título de sección en morado (`#3F37A8`, 14.5 px, semibold) y una línea fina `#EEF0F5`, como en el modal de Proveedor.
- Colores:
  - Blanco y morado de la marca (`#4F46C8` / `#3F37A8`; seleccionado `#F7F6FE`).
  - Color frío de apoyo, azul: `#EAF2FB` con texto `#1D4F7A` o `#1D5FA8`.
  - Bordes grises fríos: `#E3E5EE`, inputs `#D5D8E3`, encabezados de tabla `#F6F7FB`.
  - **Sin fondos beige ni café** (`#FAF8F3`, `#F3EFE6`, `#FDF9EE`, amarillos) dentro de los modales.
  - Verde, rojo y ámbar solo para estados (margen, error, advertencia), en texto o puntos pequeños, sin fondos grandes.
- La casilla "Seguir agregando" y similares van sin fondo de color.

### 1. Producto nuevo / Editar producto (tablero 05 "Producto nuevo"; tablero 12 "Producto nuevo")

Orden de secciones (todos los campos actuales se conservan):

1. **Identificación**
   - Código de barras (opcional) + "Buscar datos". Se conserva la búsqueda con imagen, las fuentes y la limpieza al cambiar el código.
   - Aviso "Encontramos este producto" en azul frío.
   - Tipo de producto como dos tarjetas de radio: **Producto simple / Con variantes**. Reemplazan visualmente el checkbox "Este producto tiene variantes" y usan el mismo control `tiene_variantes`, que sigue bloqueado si el producto editado ya tiene variantes.
2. **Datos del producto**: Nombre *, Categoría (+ Nueva), Proveedor, Presentación, Contenido y Unidad de contenido (+ Nueva unidad), SKU interno si hoy aplica. Los formularios en línea de "Nueva categoría" y "Nueva unidad" se conservan, con el estilo nuevo.
3. **Precio e inventario**
   - Costo proveedor * y Precio base *.
   - Línea de margen calculado debajo (verde si gana, rojo si pierde, gris si falta dato).
   - Motivo del cambio de precio cuando aplica.
   - Producto simple: Stock inicial o Stock actual *, con el botón "Editar stock" en edición.
   - Con variantes: el editor de variantes completo (ver Variantes).
4. **Detalles opcionales**
   - Plegable (`<details>`) con chevron; resumen "Marca, sucursal y descripción".
   - Contiene Marca (+ Nueva), Sucursal * (solo si `branches().length > 1` y no se edita, igual que hoy) y Descripción.
   - Debe abrirse solo si alguno de sus campos tiene error de validación.
5. **Multimedia**: igual que hoy (deshabilitada, "próximamente"), con el estilo nuevo.
6. Avisos y errores (`catalogLoadErrors`, `formError`, `productoEnOtraSucursal` con "Agregar a esta sucursal"), igual que hoy, al final del cuerpo.

- Pie: "Faltan N datos obligatorios (*)" o "Listo para guardar", más Cancelar y Guardar producto (o el texto actual en edición).
- Quitar la vista previa lateral del diseño viejo si existiera; el diseño nuevo no la tiene.

### 2. Agregar del catálogo del negocio (tablero 05 "Del catálogo del negocio"; tablero 12)

- El aviso "Entran sin precio ni stock…" queda como una línea pequeña con ícono morado, sin caja.
- Buscador con borde gris frío.
- Lista o tabla blanca:
  - Encabezado `#F6F7FB`.
  - Fila elegida con fondo `#F5F3FD` y checkbox morado.
  - "Lo vende" como etiqueta azul fría.
  - En móvil, cada fila es una tarjeta: nombre + "presentación · categoría · sucursal".
- Pie: "N de M elegidos", "Elegir todos / Quitar todos", Cancelar y "Agregar N productos". Se conserva toda la lógica actual de `agregar-catalogo.component.ts`.

### 3. Carga masiva (tablero 05 "Carga masiva"; tablero 12 "Carga masiva de productos")

- "Se importa en [sucursal activa]" pasa a una línea pequeña (12.5 px, gris, ícono de ubicación). No se cambia la sucursal aquí.
- Archivo en una sola fila: ícono azul frío, nombre, filas leídas, enlace "Plantilla" y botón quitar. Antes de elegir archivo se conserva el selector actual.
- Resultado (`import-report`):
  - Los contadores (Listos, Filas con error, Advertencias, Catálogos nuevos) son **pestañas**: tarjetas blancas con un punto de color y el número en color. La elegida lleva borde morado de 2 px y fondo `#F7F6FE`.
  - Debajo hay **un solo panel** con la tabla de la pestaña elegida. Se conservan los filtros por columna, la paginación de 10 en 10 y "Descargar errores".
  - Pestaña inicial: Filas con error si hay alguna; si no, Listos.
  - En móvil, las tablas se vuelven tarjetas.
- Pie: "Se importan N productos; las filas con error se omiten.", Cancelar y "Autorizar e importar N".

## Selects (mejora visual)

Hoy algunos `p-select` ocupan más espacio que los inputs de al lado o se ven mal. Hay 11 `<p-select>` y 1 `<select>` en `productos.component.html`, con estilos regados: `productos.component.css` (`.field--error .p-select`, `.filter-field p-select`), `styles/componentes.css` (`.formulario-alta__select`) y estilos propios en `catalogo` y `compras`.

- **Misma caja que un input:** alto 44 px, radio 10 px, borde `#D5D8E3`, padding horizontal 12 px, mismo tamaño y color de texto. El placeholder va en gris `#8A90A0`. El foco usa borde morado `#4F46C8` y anillo `#E4E1FA`. El error usa el mismo rojo que los inputs.
- **Ancho:** ocupa su celda de la cuadrícula (`width: 100%`) y nunca la desborda. Textos largos se cortan con `…`; no crecen la caja. Quitar los `[style]="{ width: '100%' }"` en línea a favor de una clase.
- **Ícono:** chevron discreto gris de 14–16 px. La X de `showClear` se alinea y no empuja el texto ni cambia el alto.
- **Panel desplegable:**
  - Fondo blanco, borde `#E3E5EE`, radio 12 px y sombra suave.
  - Cada opción mide al menos 40 px de alto. Hover `#F5F3FD`; seleccionada en morado `#3F37A8` con check.
  - Ancho mínimo igual al del select y alto máximo con scroll.
  - Dentro del modal, usar `appendTo="body"` para que el panel no quede cortado por el scroll del cuerpo.
- **Un solo lugar:** definir el estilo en una clase reutilizable (por ejemplo `.campo-select`) en `styles/componentes.css`, junto a `.formulario-alta__select`, o unificarla con esa. Aplicarla a todos los selects de los tres modales: categoría, marca, proveedor, unidad de contenido, sucursal, tipo de unidad, filtros y los de "Nueva …".
  - No cambiar los selects de otras pantallas salvo que compartan la misma clase y se vean igual o mejor.
  - Revisar Catálogo y Compras después del cambio.
- No se quita ni se cambia ninguna opción ni el comportamiento de los selects (filtrado, limpiar, crear nueva). Solo cambia el estilo.
- Revisar a 1440, 820 y 390 px que en la misma fila los selects y los inputs queden alineados y del mismo alto.

## Variantes (lógica, `productos.component.ts`)

Problemas actuales (ver captura del usuario: valores `enchilado`, `fuego`, `another`, con precio 0):

1. `removeDimensionValue()` y `removeDimension()` solo cambian `variantDimensions`; **las combinaciones ya generadas se quedan en la tabla**.
2. La leyenda dice "El precio base se aplica automáticamente a las variantes", pero en la práctica las variantes quedan en 0.

Comportamiento esperado:

- **La tabla responde al momento.** Al quitar un valor (✕ del chip) o una característica completa, se quitan en ese instante las filas que ya no corresponden. Al agregar un valor o característica, se agregan al momento las combinaciones nuevas.
  - Implementar una sola función `sincronizarCombinaciones()` y llamarla desde `addDimensionValue`, `removeDimensionValue`, `addDimension` y `removeDimension`.
  - La función calcula las combinaciones actuales; conserva las filas que siguen existiendo (con su código, precio y stock capturados); quita las que ya no existen; agrega las que faltan.
  - Al agregar una característica nueva sin valores, la tabla no cambia hasta que tenga al menos un valor.
- Las filas existentes en edición (con `id`) que ya no correspondan se quitan del formulario como hoy lo hace la ✕ de fila (`removeVariant`). Si quitar valores deja la tabla vacía, **no** se desmarca `tiene_variantes` (hoy `removeVariant` lo hace al quedar en 0; la sincronización no debe hacerlo).
- El botón **"Generar combinaciones" se conserva**. Sirve para volver a crear filas borradas a mano con la ✕ de la fila.
- **Precio base por defecto:**
  - Toda fila nueva toma el precio base actual.
  - Al cambiar el precio base, se actualizan las variantes cuyo precio no se ha editado a mano (control `pristine` o igual al precio base anterior).
  - Las editadas a mano se respetan. Hoy `applyBasePrice()` pisa todas.
  - Reproducir primero por qué queda en 0. Sospechas: precio base capturado después de generar y el `valueChanges` solo aplica con `tiene_variantes`; `syncVariantControlPriceEditing()` pone 0 en edición sin permiso; el valor llega como texto.
  - Si el usuario no tiene permiso de cambiar precios, se mantiene la regla actual y la leyenda lo dice.
- La leyenda debe describir lo que de verdad pasa: "Las variantes nuevas toman el precio base; puedes cambiarlo en cada una."

## Pruebas y cierre

- `productos.component.spec.ts`:
  - Quitar un valor quita sus filas.
  - Quitar una característica regenera las combinaciones conservando datos.
  - Agregar un valor agrega filas.
  - Fila nueva toma el precio base.
  - Cambiar el precio base no pisa precios editados.
  - La tabla vacía no desmarca variantes.
  - Las rutas `nuevo`, `importar`, `agregar-del-catalogo` y `:id/editar` abren el modal sobre la lista.
  - Cerrar regresa a `/catalogo/productos`.
- `modal.component` (si tiene spec), `import-report` y `agregar-catalogo`: ajustar a los cambios.
- `npm test` y `npm run build` en `frontend/` sin errores.
- Revisar en navegador a 1440, 820 y 390 px contra el tablero 12, incluidos los selects (alto, ancho y panel no cortado).
- `graphify update .` al terminar.
- Git: solo lectura hasta confirmación final del usuario (AGENTS.md).
