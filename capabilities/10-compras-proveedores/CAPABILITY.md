# Capability 10: Compras y recepción de mercancía

## Estado

En implementación.

## Aprobación

Revisión de recepción aprobada explícitamente por el usuario el 2026-09-28 mediante la instrucción: “Implement the plan”.

Formato visual de producto nuevo aprobado explícitamente por el usuario el 2026-09-29 mediante la instrucción: “Implement the plan”.

Validación visual de compra aprobada explícitamente por el usuario el 2026-09-29 mediante la instrucción: “Implement the plan”.

## Alcance

- Registrar compras recibidas de proveedores por sucursal y documento.
- Guardar importes del ticket, forma de pago, observaciones y usuario autenticado como capturista y receptor.
- Registrar empaques, piezas, costos, descuentos, incidencias, lote y caducidad por producto.
- Aumentar únicamente las piezas buenas y dejar movimientos de inventario trazables.
- Permitir alta rápida de producto simple y de unidad de medida durante la recepción.
- Calcular IVA (16 %) a partir de su check de inclusión y permitir ajustar el importe de IEPS cuando se incluya.
- Derivar las piezas por empaque desde la unidad de empaque y mostrar anomalías solo cuando se soliciten.
- Mostrar la captura en una sola superficie blanca, con el formulario y los productos a la izquierda y un resumen global fijo a la derecha en escritorio.
- Mostrar el alta rápida de producto con el mismo panel visual usado para la nueva unidad de empaque, conservando su creación al confirmar la compra.
- Marcar en rojo los campos obligatorios inválidos después de interactuar con ellos y desactivar las acciones de guardar o confirmar hasta que sus formularios sean válidos.

## Reglas y contratos

- Un documento es único por negocio, proveedor, tipo y folio.
- El subtotal coincide con los renglones netos; impuestos es IEPS más IVA; total es subtotal más impuestos.
- Las piezas buenas son total menos dañadas y faltantes; solo ellas entran al stock.
- La unidad elegida para una recepción debe ser de tipo `EMPAQUE` y tener un factor entero positivo; dicho factor define las piezas por empaque.
- El cliente envía `incluye_iva`, `incluye_ieps` e `ieps`; IVA se calcula en el servidor como 16 % de `subtotal + IEPS` cuando ambos impuestos están activos, o solo del subtotal cuando IEPS no aplica. Al activar IEPS el cliente propone 8 % del subtotal, pero el usuario puede ajustar el importe no negativo y el servidor persiste ese valor.
- El importe pagado se propone con el total calculado, pero el usuario puede modificarlo.
- La mercancía se captura antes del bloque final; el importe pagado es de solo lectura y siempre coincide con el total calculado por el servidor.
- Los productos existentes deben estar asociados al proveedor. Un producto nuevo requiere nombre y código de barras, recibe SKU automático, precio de venta inicial cero y se asocia al proveedor.
- Ticket/factura adjunto queda fuera de alcance hasta contar con almacenamiento de archivos.
- Una compra confirmada es inmutable.
- La sucursal receptora se toma del contexto activo y no se puede cambiar desde el formulario de compra.
- El lateral fijo contiene únicamente un resumen de los productos capturados, sin impuestos ni total.
- Los impuestos, totales y la confirmación se muestran al final del formulario. En móvil el resumen deja de ser fijo y se muestra después del formulario.
- Nombre y código de barras son obligatorios únicamente para un producto nuevo; al elegir un producto existente dejan de bloquear la compra.

## Tareas

- [x] Definir cabecera y detalle de recepción con el usuario.
- [x] Implementar modelos, transacción y endpoints de compras.
- [x] Añadir navegación, formulario y listado de compras.
- [x] Ejecutar verificaciones y actualizar el grafo.
- [x] Ajustar la pantalla de Compras al sistema visual general, tomando Catálogo como referencia.
- [x] Reemplazar importes manuales de impuestos por checks y cálculo en servidor.
- [x] Obtener piezas por empaque desde la unidad y condicionar anomalías.
- [x] Integrar creación rápida desde selects y resumen tipo ticket por renglón.
- [x] Mover mercancía antes del bloque de impuestos y fijar el importe pagado al total.
- [x] Sustituir los resúmenes por renglón por un resumen acumulado, lateral y fijo en escritorio.
- [x] Limitar el lateral al resumen de productos y mover impuestos, totales y confirmación al final del formulario.
- [x] Permitir ajustar IEPS sin permitir editar IVA.
- [x] Corregir la actualización visual de IVA al activar su check, conservando su suma al total.
- [x] Calcular IVA sobre subtotal más IEPS cuando ambos aplican.
- [x] Aplicar al producto nuevo el formato visual del formulario de nueva unidad.
- [x] Mostrar validación visual y bloquear acciones con campos obligatorios inválidos.

## Criterios de aceptación

- Una compra válida incrementa el inventario de la sucursal por sus piezas buenas y crea movimientos `COMPRA_ENTRADA`.
- Lote/caducidad son opcionales y se enlazan al movimiento cuando se informan.
- No se permite documento repetido, producto ajeno al proveedor ni cantidades inconsistentes.
- La pantalla muestra cálculos de empaques, costos y totales antes de confirmar.
- El usuario puede crear producto o unidad desde su select y el formulario respectivo aparece inmediatamente debajo.
- El panel de nuevo producto muestra nombre y código de barras con la misma estructura visual que el panel de nueva unidad.
- Los campos obligatorios tocados se muestran en rojo y los botones de guardar unidad o confirmar compra permanecen desactivados mientras sean inválidos.
- El resumen lateral muestra cada producto capturado; los impuestos y total aparecen al final del formulario.

## Resultados visuales

- Los estilos de Compras viven en `compras.component.css`, no en el decorador TypeScript.
- El contenedor respeta `--section-max-width` y `--section-padding-inline`.
- El estado inicial muestra una tarjeta blanca centrada con una sola acción global para registrar una compra.
- Los botones reutilizan los tamaños y colores de Catálogo (`primary-button` y `secondary-button`), sin sobrescribir el tema global.
- `tsc --noEmit -p frontend/tsconfig.spec.json` y `git diff --check`: correctos.

## Resultados

- `go test ./...`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check` y `git diff --cached --check`: correctos.
- La lista de Compras conserva una única acción de alta: centrada en el estado vacío y en el encabezado cuando ya existen registros.
- La estructura visual de Compras reutiliza las medidas, colores y espaciados de Catálogo.
- La sucursal receptora se preselecciona desde el contexto activo.
- Pendiente conocido: ejecutar pruebas Angular completas con Node `v22.22.3` o superior; este entorno mantiene Node `v14.15.0`.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto tras la revisión.
- `go test ./...`: la recepción compila; queda fallo ajeno en `internal/config.TestProductionAcceptsStrongOTPSecret`.
- Reordenamiento verificado con `tsc --noEmit`; `go test ./...` mantiene exclusivamente el fallo preexistente de configuración OTP.
- Rediseño del resumen global: `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json` y `git diff --check` correctos. `npm test -- --watch=false --runInBand` no inicia porque el entorno usa Node `v14.15.0` y Angular requiere `v22.22.3` o superior.
- Revisión de IEPS y distribución: `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`, `git diff --check` y `go test ./internal/application/compra ./internal/infrastructure/compra ./internal/interfaces/http/compra` correctos. Las pruebas Angular continúan bloqueadas por Node `v14.15.0`.
- Regla IVA sobre IEPS: compilación TypeScript, revisión de espacios y paquetes Go de compras correctos.
- Formato visual de producto nuevo: `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json` y `git diff --check` correctos. La prueba Angular del panel permanece pendiente por Node `v14.15.0`.
- Validación visual: `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json` y `git diff --check` correctos. `npm test -- --watch=false --runInBand` continúa bloqueado por Node `v14.15.0`.
