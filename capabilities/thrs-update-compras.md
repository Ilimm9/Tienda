# Actualización de compras: impuestos por producto, precio al peso, validar por ticket, Excel y editar/suspender/reactivar

Origen: pruebas de `thrs/docs-prueba/compras-prueba-oaxaca.md` (Depósito del Valle / OAXACA) y la hoja `thrs/compras_valle.xlsx`.
Explicación para la clienta de cómo se calcula el costo: `capabilities/explicacion-productos.md`.

Este plan es autocontenido: está pensado para que un agente nuevo lo implemente completo sin más contexto.

## Reglas que no se negocian

- **Base de datos en producción, con un cliente registrando compras.** Solo cambios aditivos: columnas nuevas nulas o con valor por defecto, tablas nuevas, índices nuevos y valores nuevos en columnas `varchar`. Prohibido borrar o renombrar columnas o tablas, o cambiar un tipo. Las compras ya guardadas se deben seguir leyendo y calculando igual que hoy.
- Leer y respetar `AGENTS.md`: capas por dominio (`domain`, `application`, `infrastructure`, `interfaces/http`), Git solo lectura hasta que el usuario confirme, `graphify update .` al terminar.
- Los tableros del canvas son la **referencia visual**. Se implementan en Angular; el canvas no se edita.
- Estilo de la app: modales todo blancos, acento morado `#4F46C8` / lila `#F1EFFD`, crema solo en avisos chicos, pesos de letra 500–600 (no 700/800), kit global `frontend/src/styles/componentes.css` (`pagina`, `encabezado-pagina`, `tabla-tarjeta`, `tabla-barra`, `segmentado`, `pastilla`, `tabla-paginador`). No esconder campos que ya existen.

## Vistas de referencia (canvas «Tienda», página «Secciones»)

Link: https://claude.ai/artifact/6yBhhs6XuSabAE6X61HBPT

Los tableros son interactivos (botón Play). Lo que hay que copiar de cada uno:

### Tablero «07 · Compras» (`project/07-Compras.dc.html`)

En Tweaks, la prop `rol` cambia entre «Propietario», «Encargado con permiso» y «Cajero sin permiso» para ver qué botones aparecen.

**Pantalla lista** (`frontend/src/app/features/compras/compras.component.*`):

- Encabezado con «Descargar Excel» (secundario) y «Registrar compra» (principal).
- Aviso crema: «Hay N costos nuevos de M compras por revisar…», con enlace a Validar costos.
- Barra de filtros: buscar por folio o proveedor, Proveedor, Desde, Hasta (fecha de recepción) y el conteo «N de M compras».
- Al marcar compras aparece una barra lila: «N compras seleccionadas · Quitar selección · Descargar seleccionadas».
- Tabla: checkbox, Recepción, Proveedor (con «N pz · N productos»), Documento (tipo + folio, con pastilla «IEPS» si aplica), Pago, Total, Costos («N por validar», «Validados» o «Cancelados»), Estado («Activa» o «Suspendida») y Acciones.
- Acciones: ver siempre. Editar y suspender si la compra está activa y hay permiso. **Reactivar** (ícono de flechas circulares) si está suspendida y hay permiso.
- Una compra suspendida se ve en gris con el total tachado.
- Sin permiso: nota con candado al pie de la tabla.

**Pantalla «Registrar compra»:**

- 1 · ¿A quién le compraste?: proveedor (con «Nuevo proveedor»), sucursal, tipo, folio, fecha del documento, fecha de recepción, forma de pago y observaciones plegables.
- 2 · ¿Qué recibiste?: una tarjeta por producto con:
  - unidad de empaque, empaques, costo por empaque y descuento;
  - **franja IEPS**: casilla «Aplica IEPS», tasa (8, 25, 26.5, 30, 53, 160 %) e «IEPS de este producto $X»; sin IEPS dice «no se le suma al costo»;
  - **franja IVA**: selector 16 % / 0 % e «IVA de este producto $X»; con 0 % dice «no se le suma al costo»;
  - la línea «Entran N piezas · costo por pieza con impuestos $X»;
  - plegable con lote, caducidad, piezas dañadas y faltantes.
  - El texto de la sección explica que IEPS e IVA se llenan con lo guardado en el catálogo.
- 3 · Revisa y guarda:
  - Productos y Subtotal.
  - Bloque IEPS con desglose por tasa («8 % sobre $588.00 · 2 productos $47.04») y el enlace «Ajustar importe del ticket», que abre el campo «Importe IEPS del ticket» y avisa que la diferencia se reparte.
  - Bloque IVA con desglose por tasa.
  - Total y el botón «Confirmar compra y agregar stock».

**Ventanas modales** (formato del tablero 12):

- **Detalle**:
  - Tabla con Producto, Empaques, Piezas buenas, Subtotal, IEPS, **IVA**, Costo por pieza y estado del costo.
  - Resumen de Subtotal, IEPS, IVA y Total.
  - Botones: Excel, Suspender y Editar si está activa; **Reactivar** y Descargar Excel si está suspendida.
- **Editar**:
  - Documento (tipo, folio, fechas, forma de pago; el proveedor no se cambia).
  - Productos con empaques, costo, **IEPS (selector con «Sin IEPS») e IVA**.
  - Si algún costo ya se validó, aviso con candado y los productos quedan bloqueados.
  - Motivo opcional.
- **Suspender**: cuadro rojo con lo que pasa (salen las piezas, se cancelan los costos pendientes, los ya validados no se tocan, «se puede reactivar después; el folio sigue ocupado») y motivo obligatorio.
- **Reactivar**: cuadro lila con lo que pasa (vuelven a entrar las piezas, los costos cancelados regresan a Validar costos como pendientes, vuelve a «Activa» con el mismo folio) y motivo opcional.
- **Descargar Excel**:
  - Alcance «Seleccionadas (N)» o «Por proveedor o fechas».
  - Hojas: «hoja 1» siempre, COMPROBACIÓN opcional, «Incluir compras suspendidas».
  - Resumen de compras, renglones e importe, aviso de renglones sin validar y la lista de columnas A–AD.
- **Nuevo proveedor**: solo el nombre.
- **Agregar producto**: buscador del catálogo y selectores IEPS e IVA, con la nota de que se llenan con lo guardado en el producto y se recuerdan.

### Tablero «08 · Validar costos y precios» (`project/08-Validar-costos.dc.html`)

El archivo trae copias de otras pantallas (catálogo, roles). **Solo cuenta la pantalla «Validar costos y precios»** (`frontend/src/app/features/precios/precios.component.*`):

- Encabezado con subtítulo (costo con impuestos y precio redondeado al peso) y «Descargar Excel».
- Filtros: buscar por folio, proveedor o producto; Proveedor; Desde; Hasta; y «N tickets · N productos».
- Tres tarjetas: «Por validar N en M tickets», «Con costo más alto» y «Ganancia estimada de lo recibido».
- Barra de selección: «Margen para seleccionados [ %]», «Aplicar» y «Validar seleccionados».
- **Una tarjeta por compra**. La cabecera lleva:
  - checkbox «elegir todo el ticket» y botón para plegar;
  - documento («Factura PRUEBA-RIC-001»), pastilla del proveedor y pastilla de impuestos («IEPS 8 % · IVA 16 %»);
  - fecha, productos por validar y total del documento;
  - ganancia estimada, «Ver compra» y «Validar ticket (N)».
- Tabla dentro de la tarjeta:
  - **Producto** con variante («Gomilocas · Dientes»), nota (unidad, dañadas, descuento, impuestos) y pastilla «Producto nuevo».
  - **Piezas**.
  - **Costo por pieza**: anterior tachado, flecha, nuevo, y pastilla «▲ x %», «▼ x %», «Sin cambio» o «Primer costo».
  - **Margen %**, editable.
  - **Precio al público**, editable, con paso de $1. Debajo: «Sugerido $X (al peso) · hoy $Y».
  - **Ganas por pieza** y total.
  - Botón **Validar**.
- Al cambiar el margen, el precio se recalcula **redondeado hacia arriba al peso**. Escribir un precio a mano recalcula el margen.
- Vacío: «Todo validado» con «Ir a compras».
- Modal Excel: alcance «Tickets del filtro» o «Tickets seleccionados», hoja COMPROBACIÓN opcional, resumen de tickets, renglones y sin validar, y nota de antes y después de validar.

## Cómo funciona hoy (para no romperlo)

- **Formulario de compra** (`compras.component.ts`): casillas `incluye_ieps` + importe `ieps` (propone 8 % del subtotal) e `incluye_iva`; las dos casillas arrancan apagadas.
- **Backend** (`backend/internal/infrastructure/compra/compra_repository.go`):
  - `ieps` = importe capturado; `iva = (subtotal + ieps) × 0.16`; `impuestos = ieps + iva`.
  - `buildDetail` guarda `costo_pieza = subtotal ÷ total_piezas`, **sin impuestos** y contando dañadas. Es solo informativo.
  - `createPriceProposal` y `receiveDetail` **repiten** el mismo cálculo: `costo final = (subtotal_renglón + impuestos × subtotal_renglón ÷ subtotal_compra) ÷ piezas_buenas`. Es decir, los impuestos se **prorratean** entre todos los renglones.
  - El costo final va a `propuestas.costo_capturado`, `lotes.costo_unitario`, `movimientos_inventario.costo_unitario` (tipo `COMPRA_ENTRADA`) y, **al validar**, a `precios_sucursal.costo_vigente`.
  - `precio_sugerido = math.Ceil(costo_final × (1 + margen))`. El margen sale de `precios_sucursal.margen`, si no de `configuracion_sucursal.margen_predeterminado`, y si no es 0.30.
- **Validar costos** (`frontend/src/app/features/precios/precios.component.ts`): **ignora** `precio_sugerido` y calcula `costo + round(costo × margen)` con centavos. Esa es la regla que se elimina.
- `ListPending` (`backend/internal/infrastructure/precio/precio_repository.go:18`) devuelve una lista plana, sin compra, y el nombre como `COALESCE(pr.nombre, vp.nombre)`, sin variante.
- **Tablas de impuestos que ya existen y nadie usa** (`backend/internal/domain/catalog_models.go:201-218`):
  - `impuestos` (`nombre`, `tipo`, `tasa numeric(7,4)`, `activo`);
  - `producto_impuestos` (`producto_negocio_id`, `impuesto_id`, índice único por par).
- Permisos: `CatalogoPermisos` en `backend/internal/domain/negocio/permiso.go:84`. El rol de sistema PROPIETARIO recibe todos con `sembrarRolPropietario` (`backend/internal/infrastructure/negocio/rol_migration.go:124`).
- `excelize` ya está en `go.mod`.

## Decisiones tomadas (2026-11-11)

1. **Precio sugerido: regla del backend.** Redondeo hacia arriba al peso, y la pantalla lo muestra y autoriza tal cual. El precio manual se sigue permitiendo.
2. **IEPS e IVA por producto** (por renglón de compra), recordados en el producto para la siguiente compra.
3. **El importe manual de IEPS se queda** (refrescos con cuota por litro, redondeos del ticket). La diferencia se reparte entre los renglones con IEPS.
4. **IVA por renglón** con tasas 16 % y 0 %.
5. Por ahora se supone que los precios del ticket **no** traen IVA (el sistema lo suma). Falta que la clienta confirme (ver preguntas abiertas).
6. **Suspender se puede reactivar.**
7. «FECHA» del Excel = **fecha de recepción**.
8. Validar costos agrupado por compra y el Excel: aprobados como estaban.
9. Permisos nuevos, columnas y tablas nuevas: aprobados, siempre aditivos.

## 1. Arreglos chicos

### 1.1 Variante en Validar costos

- En `ListPending`, el nombre se arma con la variante: `producto · pv.clave` (ej. «Gomilocas · Dientes»). Revisar la consulta actual: hoy junta `pr.nombre`/`vp.nombre` con `COALESCE`; se agrega el `JOIN` a la variante (`pv`) y su `clave`.
- Las variantes son productos propios: cada una tiene su costo y precio, y se valida por separado (un renglón por `producto_negocio_id`, como hoy).

### 1.2 Una sola regla de precio sugerido

- **Backend**: se deja `ceil`, pero se corrige el error de punto flotante: `10 × 1.1 = 11.000000000000002`, y `math.Ceil` da 12. La regla queda `math.Ceil(round4(costo × (1 + margen)))`, con una sola función, por ejemplo `precioSugerido(costo, margen)` en el dominio `precio`, que usan compras y precios.
- **Frontend** (`precios.component.ts`):
  - Al cargar: precio = `precio_sugerido` del backend, y el margen se muestra como hoy.
  - Al cambiar el margen (en el renglón o con «Aplicar» a seleccionados): precio = `Math.ceil(round4(costo × (1 + margen/100)))`, la misma fórmula.
  - Al escribir el precio a mano: se respeta y se recalcula el margen.
  - Campo de precio con `step="1"` y la leyenda «Sugerido $X (al peso) · hoy $Y».
  - Eliminar `finalPrice = costo + ganancia` como fuente del precio. «Ganas por pieza» = precio − costo.

## 2. IEPS e IVA por producto

### Esquema (todo aditivo)

**Catálogo de tasas.** Se sembra de forma idempotente en la tabla existente `impuestos`, con `tasa` como fracción, igual que `margen`:

| nombre | tipo | tasa |
|---|---|---|
| IVA 16 % | IVA | 0.16 |
| IVA 0 % | IVA | 0 |
| Sin IEPS | IEPS | 0 |
| IEPS 8 % | IEPS | 0.08 |
| IEPS 25 % | IEPS | 0.25 |
| IEPS 26.5 % | IEPS | 0.265 |
| IEPS 30 % | IEPS | 0.30 |
| IEPS 53 % | IEPS | 0.53 |
| IEPS 160 % | IEPS | 1.60 |

**Impuestos del producto.** Se usa la tabla existente `producto_impuestos`, con a lo sumo un renglón de tipo IVA y uno de tipo IEPS por `producto_negocio`.

- Al guardar una compra se recuerdan las tasas usadas: si ya hay un renglón de ese tipo, se **actualiza** su `impuesto_id`; si no, se inserta.
- «Sin IEPS» es un impuesto con tasa 0, para nunca tener que borrar renglones.

**Renglón de compra.** Columnas nuevas en `compras_proveedor_detalles`:

- `impuestos_por_renglon bool not null default false`: marca las compras nuevas. Las viejas quedan en `false` y se siguen calculando con el prorrateo.
- `tasa_ieps numeric(8,4) not null default 0`
- `ieps numeric(14,2) not null default 0`
- `tasa_iva numeric(8,4) not null default 0`
- `iva numeric(14,2) not null default 0`

**Totales de la compra.** `compras_proveedor.ieps`, `iva` e `impuestos` se siguen llenando, con la **suma** de los renglones.

### Backend (cambio mínimo)

- **`application/compra`**: `CrearDetalleInput` gana `tasa_ieps *float64` y `tasa_iva *float64`, ambos opcionales.
  - Si **algún** renglón los trae, la compra es por renglón y se usan en todos (lo que falte = 0).
  - Si ninguno los trae, se usa el modo viejo (`incluye_ieps`, `ieps`, `incluye_iva`). Así no se rompe ningún cliente.
- **Importe manual**: `CrearCompraInput` gana `ieps_ajustado *float64`.
  - Si viene y es distinto de la suma calculada, la diferencia se reparte entre los renglones con `tasa_ieps > 0`, proporcional a su subtotal. El último renglón absorbe los centavos para que la suma cuadre exacta.
  - Validar que no sea negativo, NaN ni infinito.
- **`buildDetail`** calcula por renglón:
  - `ieps = round2(subtotal × tasa_ieps)`, más su parte del ajuste;
  - `iva = round2((subtotal + ieps) × tasa_iva)`.
- **Una sola función de costo** que reemplaza el cálculo duplicado de `createPriceProposal` y `receiveDetail`: `costoFinal(compra, detalle)`.
  - Si `detalle.impuestos_por_renglon`: `(subtotal + ieps + iva) ÷ piezas_buenas`.
  - Si no: el prorrateo actual, sin cambios.
- Al guardar, se actualizan los `producto_impuestos` de cada producto (ver arriba).
- **Lectura de productos para el formulario de compra**: el endpoint que usa el buscador de productos de compras devuelve `tasa_ieps` y `tasa_iva` del producto, o `null` si no tiene.
- **Detalle de compra** (GET): devuelve `tasa_ieps`, `ieps`, `tasa_iva` e `iva` por renglón. En compras viejas, `ieps` e `iva` del renglón se calculan con el prorrateo, solo para mostrarlos.

### Frontend (`compras.component.*`, `compras.models.ts`, `ventana-compra.component.ts`)

- Cada renglón lleva la franja IEPS (casilla y tasa) y la franja IVA (16 % / 0 %), como en el tablero 07.
- Al agregar un producto, las tasas se llenan con las del producto. Si el producto no tiene, IEPS = sin IEPS e IVA = 16 %, que es lo más común en botanas, dulces y refrescos.
- Resumen con desglose por tasa de IEPS e IVA, y el «Ajustar importe del ticket» del IEPS.
- Se dejan de mostrar las casillas globales `incluye_ieps` / `incluye_iva`. Los campos se mantienen en el modelo para compatibilidad.
- «Costo por pieza con impuestos» por renglón, con la misma fórmula que el backend.

### Fuera de alcance por ahora

- Editar los impuestos del producto desde Catálogo: se recuerdan desde compras. Se puede agregar después.
- IEPS de refrescos por cuota por litro: se captura con el importe manual.
- Opción por negocio «¿Acredito IVA?» o «¿Los precios del ticket traen IVA?»: espera la respuesta de la clienta.

## 3. Validar costos agrupado por compra

- **`ListPending`** agrega por propuesta:
  - `compra_id`, `tipo_documento`, `folio_documento`, `proveedor`, `fecha_documento`, `fecha_recepcion` y `total` de la compra;
  - nombre con variante (1.1) y `codigo_barras`;
  - nota del renglón: unidad × empaques, dañadas, faltantes, descuento, tasas de IEPS e IVA y si es producto nuevo.
- **Filtros por query**: `proveedor_id`, `desde`, `hasta` (fecha de recepción) y `q` (folio, proveedor o producto).
- **Pantalla**: según el tablero 08, una tarjeta por compra, «Validar ticket (N)», «Validar seleccionados» y «Margen para seleccionados».
- **Endpoint opcional** `PATCH /negocios/:negocioId/precios/compras/:compraId/autorizar`: valida el ticket entero en una transacción, con los precios de cada renglón en el body. Si no se hace, el frontend manda N llamadas al endpoint actual.

## 4. Descargar Excel de compras

- Botón «Descargar Excel» en Compras (lista y detalle) y en Validar costos, con el modal del tablero 07 / 08.
- **Endpoint**: `GET /negocios/:negocioId/compras/exportar?sucursal_id=&proveedor_id=&desde=&hasta=&compra_id=...&incluir_suspendidas=&comprobacion=` responde un `.xlsx` hecho con `excelize`.
- Requiere `compras.ver` y `precios.ver` (el archivo lleva costos y precios). Valida la sucursal con `AccesoSucursal`.
- Por omisión excluye las compras suspendidas.

### Hoja «hoja 1»: una fila por producto comprado

Mismos encabezados que `compras_valle.xlsx` (tabla `Tabla2`):

| Col | Encabezado | Fuente |
|---|---|---|
| A | NO. DE TIKET | `tipo_documento` + `folio_documento` |
| B | FECHA | **`fecha_recepcion`** |
| C | CODIGO | `detalle.codigo_barras` (vacío si no tiene) |
| D | NOMBRE DEL PRODUCTO | producto + variante |
| E | PROVEEDOR | proveedor de la compra |
| F | SE LE APLICA IEPS? | «SI 8 %» / «NO» (`tasa_ieps`); en compras viejas, «SI» si la compra tuvo IEPS |
| G | COSTO TOTAL SIN IMPUESTO | `detalle.subtotal` (ya con descuento) |
| H | COSTO TOTAL CON IMPUESTO | subtotal + IEPS del renglón |
| I | CANTIDAD COMPRADA | `cantidad_empaques` |
| J | UNIDAD DE MEDIDA | nombre de la unidad |
| K | CANTIDAD | `piezas_por_empaque` |
| L | TOTAL DE PIEZAS | `total_piezas` |
| M | COSTO POR CAJA | fórmula H / I |
| N | COSTO POR PIEZA | costo final del sistema, `costoFinal` (ver pregunta abierta 1) |
| O | COSTO UNITARIO EN SISTEMA | `propuesta.costo_anterior` (foto al comprar) |
| P | DIFERENCIA DE COSTO | fórmula N − O |
| Q | OBSERVACIONES | fórmula: «SUBIR EL P.V» / «BAJÓ EL P.C» / «SIN CAMBIO» |
| R | P.V EN SISTEMA | `propuesta.precio_anterior` (foto) |
| S | NUEVO P.V | el autorizado si ya se validó; si no, `precio_sugerido` |
| T | PRODUCTO NUEVO | `detalle.producto_nuevo` SI/NO |
| U | NOMBRE DEL NUEVO PRODUCTO | nombre, si es nuevo |
| V | GRAMAJE O NO. DE PIEZAS | `contenido` + `unidad_contenido` del producto |
| W | PRECIO DE VENTA | precio autorizado (solo productos nuevos) |
| X | PIEZAS DAÑADAS | `piezas_danadas` |
| Y | PIEZAS FALTANTES | `piezas_faltantes` |
| Z | PIEZAS BUENAS | `piezas_buenas` |
| AA | DESCUENTO | `descuento` |
| AB | IVA | «16 %» / «0 %» y su importe |
| AC | ESTADO DE VALIDACIÓN | «Pendiente» / «Validado el dd/mm por X» / «Cancelado» |
| AD | ESTADO DE LA COMPRA | «Activa» / «Suspendida» |

Notas:

- O y R salen de la **foto** de la propuesta, no de `precios_sucursal`. Así no cambian al validar.
- En el Excel original, «PROVEEDOR» es la «Categoría de Punto de venta». Aquí va el proveedor real.

### Hoja «COMPROBACIÓN»: una fila por compra

Columnas: No. de ticket, productos registrados, importe sin impuestos, IEPS, IVA, importe con impuestos, total del documento, **diferencia** y nota (observaciones, descuento). Con fórmulas `SUMIF` sobre `Tabla2`, como la original.

No se exportan `BASE`, `DATOS`, «PRODUCTOS SACAR DEL SISTEMA» ni «LISTA DE COMPRAS».

## 5. Permisos, editar, suspender y reactivar

### Permisos nuevos en `CatalogoPermisos` (módulo `compras`)

- `compras.editar`, «Editar compras». Depende de `compras.ver` y `catalogo.ver`.
- `compras.suspender`, «Suspender y reactivar compras». Depende de `compras.ver`.

El PROPIETARIO los recibe solo, por el `CROSS JOIN permisos` de `sembrarRolPropietario`. Los demás roles se configuran en Roles y permisos (la matriz se arma desde el catálogo). El frontend oculta los botones si falta el permiso; el backend lo exige con `RequierePermiso`.

### Esquema (aditivo)

**`compras_proveedor`**, columnas nuevas:

- `estado varchar(20) not null default 'ACTIVA'` (`ACTIVA` | `SUSPENDIDA`)
- `motivo_suspension text null`
- `suspendido_por_usuario_id uuid null`
- `suspendido_en timestamptz null`
- `reactivado_por_usuario_id uuid null`
- `reactivado_en timestamptz null`
- `actualizado_en timestamptz null`
- `editado_por_usuario_id uuid null`

**`compras_proveedor_detalles`**: `cancelado bool not null default false`, para el renglón que se quita al editar.

**Tabla nueva `compras_proveedor_eventos`**:

- `id`, `compra_id`, `usuario_id`, `motivo text null`, `creado_en`;
- `tipo varchar(20)`: `EDICION` | `SUSPENSION` | `REACTIVACION`;
- `antes jsonb`, `despues jsonb`.

**Valores nuevos** (sin cambiar tipos):

- `propuestas.estado` gana `CANCELADA`.
- `movimientos_inventario.tipo` gana `COMPRA_AJUSTE`, `COMPRA_SUSPENSION` y `COMPRA_REACTIVACION`.
- El movimiento original `COMPRA_ENTRADA` **nunca** se modifica: cada cambio agrega un movimiento nuevo (kardex).

### Editar: `PATCH /negocios/:negocioId/compras/:compraId` (requiere `compras.editar`)

- Solo compras ACTIVA.
- **Encabezado** (tipo, folio, fechas, forma de pago, cuenta, observaciones): siempre editable. Respeta el folio único. El proveedor no se cambia.
- **Productos** (cantidades, costo, descuento, dañadas, faltantes, tasas IEPS e IVA, agregar o quitar renglón): solo si **ninguna** propuesta de la compra está autorizada. Si alguna lo está, se bloquea el detalle y se explica por qué.
- **Stock**: la diferencia de piezas buenas se registra como `COMPRA_AJUSTE`. Si bajar piezas deja stock negativo (ya se vendieron), se rechaza con un mensaje claro.
- **Propuestas**: se actualizan en el mismo registro (`compra_detalle_id` es único) con el nuevo costo, piezas y precio sugerido. Renglón quitado: `cancelado = true` y su propuesta pasa a `CANCELADA`.
- **Lote del renglón**: se actualizan cantidad y costo unitario.
- Todo en una transacción. Evento `EDICION` con antes y después.

### Suspender: `POST …/compras/:compraId/suspender` (requiere `compras.suspender`, body `{ "motivo": "..." }`)

- Motivo obligatorio. Solo compras ACTIVA.
- Revierte el stock: un `COMPRA_SUSPENSION` por renglón con las piezas buenas en negativo. Si algún producto no alcanza, se rechaza y se listan los productos.
- Las propuestas **pendientes** pasan a `CANCELADA`. Sus ids se guardan en `evento.despues` para poder reactivarlas.
- Las propuestas ya autorizadas **no** revierten costo ni precio (pudo haber ventas a ese precio). Se avisa en el modal.
- El lote baja su `cantidad_actual`.
- La compra sigue en la lista como «Suspendida», fuera de los totales y del Excel por omisión. El folio sigue ocupado.

### Reactivar: `POST …/compras/:compraId/reactivar` (requiere `compras.suspender`, body `{ "motivo": "..." }` opcional)

- Solo compras SUSPENDIDA.
- Vuelve a sumar el stock: un `COMPRA_REACTIVACION` por renglón no cancelado, más el lote.
- Las propuestas canceladas **por esa suspensión** (ids del último evento `SUSPENSION`) vuelven a `PENDIENTE`. Los valores de `costo_anterior` y `precio_anterior` de la foto se quedan como estaban.
- `estado = 'ACTIVA'`, con `reactivado_por` y `reactivado_en`. Evento `REACTIVACION`.

Los tres endpoints validan la sucursal con `AccesoSucursal`, como `Get`.

## Orden de implementación

1. Arreglos chicos: variante, y precio sugerido con una sola regla y el arreglo de punto flotante.
2. IEPS e IVA por renglón: esquema, catálogo de tasas, backend, formulario. Es la base para que el Excel diga SI/NO bien.
3. Validar costos agrupado por compra, con filtros.
4. Exportar Excel (hoja 1 + COMPROBACIÓN).
5. Permisos, estado, eventos, editar, suspender y reactivar.

Cada punto se puede entregar por separado.

## Pruebas

**Backend**

- `precioSugerido`: `10 × 1.10 = 11` (no 12); `25.0558 × 1.30` da 33; `20 × 1.30` da 26.
- IEPS e IVA mixtos: Gomilocas $240 con IEPS 8 % e IVA 16 % da $25.0558 por pieza; leche $264 sin IEPS e IVA 0 % da $22.00. Ver `explicacion-productos.md`.
- Ajuste manual del IEPS: la diferencia solo va a los renglones con IEPS y la suma cuadra al centavo.
- Compras viejas (`impuestos_por_renglon = false`): mismo costo que antes del cambio.
- Cliente viejo, sin tasas en los renglones: se comporta como hoy.
- Las tasas usadas quedan en `producto_impuestos` y se devuelven en la siguiente búsqueda del producto.
- Editar:
  - encabezado con folio duplicado: falla;
  - detalle con una propuesta autorizada: falla;
  - bajar piezas con stock ya vendido: falla;
  - cambiar el costo actualiza la propuesta, el precio sugerido y el lote.
- Suspender:
  - sin permiso: 403;
  - propietario sin habérsele asignado el permiso: pasa;
  - revierte el stock;
  - las pendientes pasan a CANCELADA y no aparecen en Validar costos.
- Reactivar:
  - vuelve a sumar el stock;
  - solo regresan a PENDIENTE las canceladas por esa suspensión;
  - suspender, reactivar y suspender otra vez funciona.
- Excel:
  - filtros por proveedor, fechas y compra;
  - B es la fecha de recepción;
  - O y R iguales antes y después de autorizar;
  - COMPROBACIÓN cuadra con C1–C7 del documento de pruebas.

**Frontend**

- Specs de `compras.component`: IEPS e IVA por renglón, desglose y ajuste.
- Specs de precios: agrupación por compra, precio al peso, precio manual.

**Al final**: `go test ./...`, las pruebas de Angular y `graphify update .`.

## Preguntas abiertas para la clienta

1. **¿Acredita IVA? ¿Los precios de sus tickets ya traen IVA?**
   - Si acredita, el costo debería ir sin IVA.
   - Si los tickets ya lo traen, no se marca IVA en esos productos.
   - Hoy el sistema suma IVA y divide entre las piezas buenas; su Excel no suma IVA y divide entre todas las piezas.
   - Con la respuesta se decide si se agrega una opción por negocio. **No bloquea el plan.**
2. ¿Quiere editar los impuestos del producto desde Catálogo, además de desde compras?
