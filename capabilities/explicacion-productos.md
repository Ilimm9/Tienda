# Cómo calcula y guarda el sistema el costo de un producto

Documento para explicarle a la licenciada cómo funciona hoy el costo de compra, con un ejemplo, y qué cambia con la actualización de compras.

## La regla en una línea

**Costo por pieza = (lo que pagaste por el producto + sus impuestos) ÷ piezas que llegaron buenas**

Después:

**Precio sugerido = costo por pieza × (1 + margen), redondeado hacia arriba al peso**

## Ejemplo 1: una caja de Gomilocas

Compra: 1 caja de 12 Gomilocas a **$240.00** sin impuestos ($20.00 por pieza). Lleva IEPS 8 % e IVA 16 %. Margen de la sucursal: 30 %.

| Paso | Cálculo | Resultado |
|---|---|---|
| Subtotal (lo que dice el ticket sin impuestos) | 1 × $240.00 | $240.00 |
| IEPS 8 % | $240.00 × 0.08 | $19.20 |
| IVA 16 % (se calcula sobre subtotal + IEPS) | ($240.00 + $19.20) × 0.16 | $41.47 |
| Total pagado | $240.00 + $19.20 + $41.47 | **$300.67** |
| Costo por pieza (12 buenas) | $300.67 ÷ 12 | **$25.06** |
| Precio sugerido (30 %) | $25.06 × 1.30 = $32.57, se sube al peso | **$33.00** |

En la pantalla de Validar costos la sucursal ve $25.06 de costo y $33.00 de precio sugerido. Puede cambiar el margen o escribir otro precio antes de validar.

### Según los impuestos que se marquen

Misma caja de $240.00 (12 piezas, todas buenas):

| Impuestos | Costo por pieza | Precio sugerido (30 %) |
|---|---|---|
| Ninguno | $20.00 | $26.00 |
| Solo IVA 16 % | $23.20 | $31.00 |
| IEPS 8 % + IVA 16 % | $25.06 | $33.00 |

### Si llegan piezas dañadas

Si de las 12 llegan 2 dañadas, se pagaron igual las 12 pero solo se venden 10. El sistema reparte lo pagado entre las buenas:

$300.67 ÷ 10 = **$30.07** por pieza, precio sugerido **$40.00**.

Las piezas faltantes cuentan igual que las dañadas.

## Dónde se guarda cada número

| Dato | Dónde | Valor del ejemplo 1 |
|---|---|---|
| Costo por pieza **sin** impuestos | Renglón de la compra (`costo_pieza`). Solo informativo | $20.00 |
| Costo por pieza **con** impuestos | Propuesta que espera validación (`costo_capturado`) | $25.06 |
| Precio sugerido | Propuesta (`precio_sugerido`) | $33.00 |
| Costo de las piezas que entraron | Movimiento de inventario y lote | $25.06 |
| **Costo oficial del producto en la sucursal** | Precios de la sucursal (`costo_vigente`). **Cambia solo cuando alguien valida** | $25.06 |
| Precio de venta oficial | Precios de la sucursal (`precio_venta`). Cambia solo al validar | El que se autorice |

Mientras nadie valide, la sucursal sigue vendiendo con su costo y precio anteriores. La mercancía sí entra al inventario desde que se guarda la compra.

## Lo que hoy está mal y se corrige

### Ticket con productos que llevan impuestos distintos

Hoy el IEPS y el IVA se capturan **una sola vez por compra** y se reparten entre todos los productos según lo que costó cada uno. En un ticket mixto, eso le carga impuestos a productos que no los llevan.

Ejemplo: en el mismo ticket, la caja de Gomilocas ($240.00, IEPS 8 % e IVA 16 %) y una caja de 12 litros de leche ($264.00, sin IEPS e IVA 0 %).

| Producto | Costo por pieza **hoy** | Costo por pieza **correcto** | Precio sugerido hoy → correcto |
|---|---|---|---|
| Gomilocas | $25.06 | $25.06 | $33.00 → $33.00 |
| Leche 1 L | **$27.56** | **$22.00** | **$36.00 → $29.00** |

La leche sale $5.56 más cara por pieza porque absorbe parte del IEPS y del IVA de las Gomilocas. Con la actualización, **cada producto lleva su propio IEPS y su propio IVA**, y el sistema recuerda los impuestos de cada producto para la siguiente compra.

### Precio con centavos

Hoy la pantalla de validar propone precios como $21.04. Con la actualización se usa una sola regla: siempre se redondea **hacia arriba al peso** ($22.00). Si la sucursal quiere otro precio, lo escribe a mano.

## ¿Es correcto este cálculo?

Depende de dos cosas que hay que confirmar con la licenciada:

1. **¿La tienda recupera (acredita) el IVA de sus compras?**
   - **No lo recupera** (compra con tickets sin factura, o su régimen no le permite acreditar): el IVA que paga sí es parte de su costo. **El cálculo actual es correcto.**
   - **Sí lo recupera** (pide factura y lo descuenta del IVA que cobra al vender): su costo real es **sin IVA**. En el ejemplo 1 serían $21.60 por pieza en vez de $25.06, y los precios sugeridos salen 16 % más altos de lo necesario.
2. **¿Los precios de sus tickets ya traen el IVA incluido?**
   - Si ya lo traen y se marca IVA en la compra, el sistema lo **suma dos veces**. En ese caso no hay que marcar IVA en esos productos.
   - Por ahora el sistema supone que los precios del ticket **no** traen IVA.

### Diferencias con su hoja de Excel

| Punto | Su Excel | El sistema |
|---|---|---|
| IVA | No lo suma al costo | Lo suma si se marca |
| Piezas dañadas | Divide entre todas las piezas | Divide entre las piezas buenas |
| Costo por pieza del ejemplo 1 | ($240 × 1.08) ÷ 12 = **$21.60** | **$25.06** |

Con sus respuestas se ajusta el sistema para que dé el mismo número que su hoja, o se le explica por qué el número del sistema es el correcto.
