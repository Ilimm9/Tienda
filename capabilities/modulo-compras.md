# Módulo de compras

Base vigente según código local. Actualizado: 2026-10-19.

## Flujo y reglas

1. Elegir proveedor y documento `TICKET` o `FACTURA`, folio y fechas; la sucursal receptora viene del contexto activo.
2. Capturar productos del proveedor, empaques, costo, descuento y, si aplica, piezas dañadas/faltantes, lote y caducidad. La unidad debe ser `EMPAQUE` con factor entero positivo. Se permite producto nuevo: requiere nombre y el código de barras es opcional; si se captura, no puede repetirse en el negocio ([22](22-compras-producto-sin-codigo.md)); crear proveedor desde Compras exige `catalogo.gestionar`. Cada producto capturado aparece en un contenedor compactable que resume nombre, empaques y subtotal; al agregar otro, los anteriores se compactan, y los incompletos se abren al intentar guardar ([23](23-compras-productos-compactables.md)). En el listado, la fecha de recepción es fecha civil y se muestra en UTC con formato «dd MMM yyyy», sin convertir a hora de Ciudad de México.
3. Revisar mercancía, impuestos y pago. Piezas buenas = empaques × piezas por empaque − dañadas − faltantes. Subtotal suma renglones netos; IEPS se captura cuando aplica y el servidor calcula IVA al 16 % sobre subtotal + IEPS. Total e importe pagado coinciden.
4. Confirmar una transacción: cabecera, detalle, entrada de piezas buenas, movimientos `COMPRA_ENTRADA`, lotes opcionales y propuestas de costo/precio. Una compra confirmada es inmutable; el documento no puede repetirse para negocio/proveedor/tipo/folio.
5. En Precios, consultar y autorizar cada propuesta por sucursal. Recibir mercancía aumenta stock; costo y precio vigentes cambian al autorizar, con historial de valores y responsable.

Las cantidades, importes y pertenencia a negocio/proveedor/sucursal se validan en servidor. El costo por pieza procede del ticket y es de consulta en la bandeja; se ajustan los valores de autorización permitidos. Las rutas exigen permisos `compras.ver`, `compras.registrar`, `precios.ver` o `precios.autorizar` según operación. Adjuntar ticket/factura queda fuera del flujo actual.

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `compras_proveedor` | Negocio, proveedor, sucursal, documento, fechas, impuestos, pago y responsables. |
| `compras_proveedor_detalles` | Producto comercial, empaques/piezas, costos, descuento, incidencias y lote. |
| `proveedores`, `producto_proveedor`, `producto_negocio`, `unidades_medida` | Origen de mercancía, productos comprables y presentación. |
| `inventario_sucursal`, `movimientos_inventario`, `lotes` | Entrada de piezas buenas y trazabilidad por sucursal. |
| `propuestas_costo_precio` | Valores recibidos, propuesta pendiente y auditoría de autorización. |
| `precios_sucursal`, `configuracion_precios_sucursal` | Costo/precio autorizados y configuración del margen. |

## Dónde cambiar y qué revisar

- Backend: [aplicación de compras](../backend/internal/application/compra/), [persistencia](../backend/internal/infrastructure/compra/) y [precios](../backend/internal/application/precio/).
- Frontend: [Compras](../frontend/src/app/features/compras/) y [Precios](../frontend/src/app/features/precios/).
- Comprobar documento repetido, piezas inconsistentes, impuestos, rollback y autorización sin afectar otra sucursal. Pendientes en [planes actuales](planes-actuales.md).
- Antecedentes: [compras y recepción](historial/10-compras-proveedores.md) y [costos y precios](historial/12-validacion-costos-precios-sucursal.md).
