# Adecuación 2 · Presentaciones de venta sobre un mismo inventario

Estado: **Plan, pendiente confirmación de la Lic. y decisiones** · Fecha: 2026-11-03 · Origen: observaciones de la Lic. (punto 3, ejemplo Corona Mega)

## Objetivo

Un producto original (ej. Corona Mega) se compra por empaque (cartón de 12) y se vende de varias formas (pieza, promoción de 2, cartón completo), cada una con su precio. Todas descuentan del **mismo inventario en piezas**. El **costo por pieza** lo calcula el sistema a partir de las compras.

## Situación actual

- Compras ya convierten empaques a piezas: unidad `EMPAQUE` con factor (Pieza 1, Caja 12, Caja 24), compartida por todo el negocio.
- Existe `producto_unidades` (`nombre`, `factor`, `codigo_barras`, `activo`) por producto, sin precio ni uso en flujos.
- `configuraciones_negocio.metodo_costeo` = `promedio_ponderado` por defecto; `movimientos_inventario` y `lotes` guardan `costo_unitario`. No hay costo promedio calculado y guardado.
- Precios por sucursal en `precios_sucursal` (producto o variante), con autorización de cambios.
- **Ventas no está construido** (pantalla de relleno).
- Riesgo actual: registrar «Corona pieza» y «Corona cartón» como variantes crea inventarios separados. Es incorrecto para este caso.

## Alcance

1. **Presentaciones de venta por producto**: nombre (Pieza, Promo 2, Cartón), factor en piezas, código de barras propio opcional, activa/inactiva. La presentación base (factor 1) existe siempre.
2. **Precio por presentación y sucursal**, con las mismas reglas de autorización de `precios_sucursal`.
3. **Inventario solo en piezas**: vender una presentación descuenta `factor × cantidad` del producto original.
4. **Costo calculado**: cada recepción de compra recalcula el costo por pieza por sucursal (promedio ponderado). El costo de una presentación = costo por pieza × factor (solo se muestra, no se captura).
5. **Formulario de producto**: sección «Formas de venta» con lista editable; muestra margen por presentación.
6. **Compras**: siguen registrándose sobre el producto original; sin cambio de flujo.
7. **Ventas** (cuando se construya): buscar por código de presentación o producto, elegir presentación, descontar piezas.

Fuera de alcance: módulo de ventas/caja completo, promociones por fecha, combos de productos distintos.

## Contrato propuesto (borrador)

- Reutilizar `producto_unidades` como presentaciones (agregar `es_base bool`, `orden int`), o tabla nueva `producto_presentaciones` (ver D1).
- `precios_sucursal` gana `presentacion_id` nullable; `null` = presentación base.
- Nueva tabla o columnas `costo_sucursal` (`producto_negocio_id`, `sucursal_id`, `costo_promedio`, `actualizado_en`) actualizada al confirmar una compra.
- Endpoints: CRUD de presentaciones bajo el producto; precios por presentación en el endpoint de precios existente.

## Decisiones a tomar

| # | Decisión | Opciones | Recomendación |
| --- | --- | --- | --- |
| D1 | Dónde guardar presentaciones | Reutilizar `producto_unidades` · tabla nueva | **Reutilizar `producto_unidades`**: ya tiene nombre, factor y código. |
| D2 | «Promoción de 2» | Presentación permanente (factor 2) · promoción con vigencia | **Presentación permanente** ahora; promociones con fecha después. |
| D3 | Precio por presentación | Por sucursal · igual en todo el negocio | **Por sucursal**, como el precio actual. |
| D4 | Método de costo | Promedio ponderado · último costo | **Promedio ponderado** (ya es el valor de `metodo_costeo`). |
| D5 | Columna `Costo` del Excel | Dejarla como costo inicial · quitarla | **Dejarla**: solo para el stock inicial; después lo calcula el sistema. |
| D6 | Productos con variantes | Presentaciones por variante · solo productos simples | **Solo simples** en esta entrega (decidir si se amplía). |
| D7 | Código de barras de presentación | Único en el negocio junto con códigos de producto · libre | **Único en el negocio**: el escáner debe saber qué vende. |
| D8 | Venta de presentación sin piezas suficientes | Bloquear · permitir según `permitir_inventario_negativo` | **Respetar `permitir_inventario_negativo`**. |
| D9 | Carga masiva de presentaciones | Hoja aparte en el Excel · solo desde formulario | **Solo formulario** en esta entrega. |
| D10 | Orden respecto a Ventas | Hacer modelo + formulario ahora · esperar módulo de ventas | **Modelo, formulario, precios y costo ahora**; el descuento en venta se conecta al construir Ventas. |

## Tareas

- [ ] Confirmación de la Lic. del ejemplo (mensaje enviado).
- [ ] Resolver D1–D10.
- [ ] Backend: modelo de presentaciones, validaciones (factor entero > 0, base única, código único).
- [ ] Backend: precio por presentación y sucursal con autorización.
- [ ] Backend: costo promedio por sucursal recalculado al recibir compra.
- [ ] Frontend: sección «Formas de venta» en producto (nombre, factor, código, precio, costo calculado, margen).
- [ ] Frontend: lista de productos muestra presentaciones bajo el producto.
- [ ] Actualizar `registro-productos.md`, `modulo-compras.md`, `PLAN_IMPLEMENTACION.md`.

## Pruebas

- Go: crear/editar/desactivar presentaciones; factor inválido; código repetido; precio por sucursal; costo promedio tras 2 compras con costos distintos (ej. 12 a $180 y 12 a $204 → $16.00 por pieza).
- Go: aislamiento por negocio y sucursal.
- Angular: formulario de formas de venta, costo y margen calculados.
- Manual: Corona Mega con Pieza ($20), Promo 2 ($38), Cartón ($216); comprar 2 cartones; inventario 24 piezas; costo por pieza correcto.

## Criterios de aceptación

- Un producto tiene varias formas de venta con precio propio y un solo inventario en piezas.
- El costo por pieza se actualiza solo al recibir compras; nadie lo captura por presentación.
