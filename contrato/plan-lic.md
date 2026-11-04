# Stockion: plan de trabajo

**Propuesta actualizada el 2 de noviembre de 2026.** Objetivo solicitado: entrega en un máximo de cuatro meses. Fechas propuestas: del 10 de noviembre de 2026 al 10 de marzo de 2027, pendientes de acuerdo entre el equipo y usted. Falta validar agenda de pruebas y viabilidad del alcance antes de convertirlo en compromiso contractual.

Stockion será la plataforma para administrar sus dos sucursales: compras, precios, inventario, ventas y caja, con autorizaciones desde su celular.

Partimos de lo que ya está construido (Fase 0) y trabajaremos 3 fases más con tareas en paralelo. Cada entrega se prueba y acepta; la preparación de la siguiente puede avanzar mientras usted prueba la anterior. La caja cambia de sistema únicamente después de superar las pruebas de operación.

## Fase 0: Base actual

Es la base que Stockion ya tiene construida para revisar con usted y sobre la que se construyen las siguientes fases. Presentarla no sustituye su prueba y aceptación:

- **Registro y acceso:** alta del negocio, inicio de sesión, verificación de correo y recuperación de contraseña.
- **Catálogo de productos:** alta de productos y catálogos, con carga masiva desde Excel.
- **Proveedores.**
- **Compras:** registro de compras y recepción de mercancía.
- **Comparación de costos y precios:** análisis de compra contra venta por sucursal, con historial.
- **Equipo de trabajo:** invitaciones a colaboradores, roles y permisos.

**Siguiente paso:** usted la revisa y nos confirma su aceptación.

Ventas aún requiere integración y pruebas antes de sustituir Odoo; no se presenta como caja lista para operar dentro de esta base. Fase 0 reconoce lo existente y no agrega semanas al calendario de las Fases 1–3.

## Fase 1: Control del negocio

Odoo sigue funcionando en caja mientras construimos esta fase.

- **Información protegida:** respaldos automáticos, recuperación probada y registro de quién cambió qué.
- **Buzón de pendientes:** todo lo que requiere su decisión, en un solo lugar y desde el celular. Le avisa de:
  - Compras: cambios de costo o precio por autorizar, productos nuevos por validar y diferencias contra el ticket.
  - Inventario: diferencias en conteos y transferencias que llegan incompletas.
  - Ventas y caja: cancelaciones, cortes con diferencia y ventas sin internet que requieren revisión.

Los avisos de ventas y caja se activan cuando esas funciones se integran en la Fase 2.
- **Compras y precios:** comparación contra el ticket, foto del ticket, órdenes de compra, autorización masiva de precios y validación de productos nuevos.
- **Impuestos por producto:** IVA e IEPS según lo que realmente vende.
- **Inventario:** historial por producto, conteos físicos, mermas, alertas de faltantes y transferencias entre sucursales.
- **Su catálogo real:** migración desde Odoo, limpieza de datos y conteo inicial en ambas sucursales.
- **Su información exportable:** compras, productos, precios e inventario se pueden descargar en Excel para su propio análisis.

**Terminamos cuando:** usted autoriza desde su celular y opera compras e inventario reales durante cuatro semanas sin pérdida de datos.

## Fase 2: Mejora de la sección de ventas

- **Venta rápida** con lector de código de barras, ticket y uso sin internet.
- **Venta por pieza, por caja y a granel.**
- **Caja por usuario:** apertura, cierre, cortes y diferencias.
- **Formas de pago:** efectivo, transferencia y tarjeta, cobrada con su terminal bancaria.
- **Devoluciones y descuentos** con límites por empleado.
- **Envases retornables** por cliente.
- **Reportes y panel:** por periodo, proveedor, producto, categoría y usuario.
- **Descargas a Excel:** reportes, ventas, detalle de tickets, cortes de caja, devoluciones y envases, además de compras, productos, precios e inventario de la Fase 1. Las descargas respetan los permisos de cada usuario.
- **Capacitación** y salida ordenada de Odoo, primero en una caja y después en todas.

**Sin internet significa:** las ventas admitidas en equipos preparados se guardan localmente y se sincronizan sin duplicados al regresar la conexión, dentro del tiempo máximo sin conexión que acordemos. Tiene estos límites:

- Cada equipo debe prepararse antes con internet.
- Dos cajas sin conexión no se ven entre sí: pueden vender la misma última pieza. El sistema lo avisa para revisarlo.
- Si un equipo se destruye antes de sincronizar, sus ventas pendientes se pierden.
- **Tarjeta:** el cobro lo hace su terminal bancaria, no Stockion. Si la terminal no tiene señal, no se puede cobrar con tarjeta y se ofrece otra forma de pago. En Stockion se anota el cobro solo después de que la terminal lo aprueba.
- **Necesitan internet:** autorizar desde su celular, entrar por primera vez en un equipo y los servicios en línea. Si un equipo se descompone o se queda sin espacio, se sigue la contingencia acordada, por ejemplo continuar en otra caja preparada. Eso no recupera automáticamente las ventas pendientes del equipo original.

**Terminamos cuando:** ambas sucursales operan una semana sin Odoo, con cortes cuadrados, y superamos la prueba de desconectar el internet durante un día real.

## Fase 3: Revisión final

No agrega módulos. Es para dejar firme lo construido.

- Revisión de seguridad y corrección de lo encontrado.
- Pruebas completas de todos los flujos y prueba de carga.
- Simulacro de recuperación ante una falla grave.
- Corrección de sus observaciones.
- Manual actualizado y entrega final.

**Terminamos cuando:** no queda ningún error grave abierto y usted firma la aceptación.

## Tiempo estimado

| Fase | Ventana propuesta desde el inicio |
| --- | --- |
| 1. Control del negocio | Construcción semanas 1–8; prueba de compras semanas 9–12 |
| 2. Mejora de la sección de ventas | Preparación en paralelo; construcción hasta semana 12; transición semanas 13–14 |
| 3. Revisión final | Pruebas durante todo el proyecto; cierre semanas 15–16 |
| **Límite total solicitado** | **Máximo cuatro meses; objetivo interno de 16 semanas** |

Las ventanas se solapan y no se suman. Debemos acordar inicio, fecha límite, equipo disponible y tiempos de sus pruebas. La validación del trabajo restante puede exigir decisiones de alcance o apoyo adicional; cualquier cambio se acuerda antes de comprometerlo. Si una prueba de seguridad u operación falla, no se retira Odoo ni se declara entregado el sistema por cumplir una fecha.

Con el inicio propuesto del 10 de noviembre, el hito interno de 16 semanas cae el 2 de marzo de 2027; el máximo propuesto es el 10 de marzo. Los pagos o el soporte posteriores se acuerdan por separado y no alargan el plazo de entrega.

## Posibles extras

No forman parte de estas fases. Los revisamos aparte si le interesan:

- Venta de recargas y pago de servicios.
- Cobros en línea con tarjeta.
- Terminal conectada directamente a la caja.
- Resumen por WhatsApp y lectura automática del ticket para prellenar compras. Adjuntar una foto como comprobante sí forma parte de la Fase 1.

## Lo que necesitamos de usted

1. ¿Cuántas cajas tiene por sucursal y con qué equipo? ¿Ya tiene impresora, cajón y báscula?
2. ¿Qué productos vende a granel y en qué unidad? ¿Qué presentaciones maneja por caja? ¿Vende a crédito o a mayoreo?
3. ¿Qué terminal bancaria usa y con qué banco?
4. ¿Cuánto tiempo y con qué frecuencia suele fallar el internet?
5. ¿Los envases se controlan con el cliente, con el proveedor o con ambos?
