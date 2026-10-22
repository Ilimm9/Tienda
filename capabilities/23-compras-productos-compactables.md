# Productos de Registrar compra en contenedores compactables

- Estado: **Verificada** (pendiente de prueba manual y aceptación final).
- Aprobación explícita: 2026-10-21. El usuario eligió «Contraer los anteriores» y aprobó el plan.
- Fecha: 2026-10-21.
- Organización conforme a [AGENTS.md](../AGENTS.md). Solo frontend de `compras`; sin cambios de API ni de datos enviados.

## Objetivo

En «Mercancía recibida», cada producto de la compra aparece en su propio contenedor, que se puede compactar, para distinguir y revisar con claridad varios productos.

## Reglas de interfaz

- Cada detalle es una tarjeta con encabezado siempre visible: chevron, número, nombre del producto (o «Producto nuevo» / «Sin producto»), empaques y subtotal.
- El encabezado es un botón con `aria-expanded` y `aria-controls`; un clic alterna entre abierto y compacto.
- Al agregar un producto, todos los anteriores se compactan y el nuevo queda abierto.
- «Quitar» es un botón de icono en el encabezado (`aria-label="Quitar producto N"`), deshabilitado con un solo producto. Quitar uno no cambia el estado de los demás, porque se guarda por control y no por índice.
- Un detalle inválido y tocado muestra la etiqueta «Incompleto». Al intentar guardar, los detalles inválidos o con incidencias mayores al total se abren automáticamente.
- Compactar solo oculta los campos; los valores se conservan y se envían igual.

## Fuera de alcance

- Contraer o expandir todos con un botón.
- Reordenar productos.
- Cambios en el panel lateral «Resumen de productos».

## Archivos

- `frontend/src/app/features/compras/compras.component.ts`: `contraidos`, `estaContraido`, `alternarDetalle`, `addDetail`, `removeDetail` y apertura de inválidos en `submit`.
- `frontend/src/app/features/compras/compras.component.html`
- `frontend/src/app/features/compras/compras.component.css`: tarjeta y encabezado; se retiraron las reglas `.detail-actions`, que quedaron sin uso.
- `frontend/src/app/features/compras/compras.component.spec.ts`

## Pruebas ejecutadas (2026-10-21, Node 22.23.2)

- `ng test --include src/app/features/compras`: 8/8 correctas. Pruebas nuevas:
  - contrae los anteriores al agregar, con resumen, `aria-expanded` y alternancia;
  - conserva el estado de los demás al quitar uno;
  - abre los incompletos al intentar guardar.
- `ng build`: correcto. El aviso de presupuesto de `compras.component.css` sube de 5.36 kB a 6.48 kB; el límite de error es 16 kB.

## Correcciones relacionadas

- 2026-10-21 (plan aprobado por el usuario): la tabla de compras mostraba `fecha_recepcion` cruda (`2026-10-21T00:00:00Z`). Es una fecha civil (`type:date`), así que se formatea con `date: 'dd MMM yyyy' : 'UTC' : 'es-MX'` → «21 oct 2026». No se usa `fechaMexico`, porque en UTC-6 mostraría el día anterior ([11](historial/11-fechas-auditoria.md)). Prueba nueva en `compras.component.spec.ts`; `ng test --include src/app/features/compras` 9/9 y `ng build` correctos.

## Pendientes

- Prueba manual en navegador, escritorio y móvil: agregar 3 productos, abrir y cerrar, quitar uno, y guardar con un producto incompleto compactado.
- Aceptación final del usuario.
