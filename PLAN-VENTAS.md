# Plan de implementación: Ventas en Stockion

Oct 28, 2026 · @Ilu

Este plan lleva a código el módulo de ventas de Stockion en 6 fases: datos, captura de venta, cobro, historial, accesos rápidos e inicio de vendedor. Cada fase termina con algo que se puede probar y un commit propio.

## Alcance y referencias del diseño

Entra: capturar una venta por código de barras o nombre, editar la lista, verificar costos, cobrar, descontar inventario, ver y cancelar tickets, y el acceso rápido a la venta. Queda fuera por ahora: facturación, ventas a crédito, clientes registrados, corte de caja e impresora de tickets física.

Todo sale del lienzo «Tienda — Alta inicial e invitaciones», página Secciones:

| Tablero | Qué define |
| --- | --- |
| 04b · Nueva venta | Captura, lista editable, verificar costos, cobro y pantalla de venta terminada |
| 04 · Ventas | Historial con totales por periodo, detalle del ticket, reimprimir y cancelar |
| 01 · Inicio | Botón «Nueva venta» en la bienvenida y en «Ir directo a» |
| 01b · Inicio para vendedor | Inicio y menú reducido para el puesto Cajero |
| Menú lateral (todas) | Botón «Nueva venta» con atajo F2 arriba del menú |
| 10 · Roles y permisos | Permisos que deciden qué menú e inicio ve cada usuario |

Estilo: Inter, fondo #FAF8F3, barra superior #2C2E6A, acento #4F46C8, tarjetas blancas con borde #E6E0D2 y radio de 16 px. Si el proyecto ya tiene componentes con este estilo, se reutilizan en lugar de crear otros.

## Modelo de datos y reglas de negocio

Una venta guarda una foto de precio y costo de cada producto al momento de cobrar, para que cambios futuros de precio no alteren tickets pasados. Claude Code debe adaptar estos nombres a las tablas que ya existen (productos, existencias por sucursal, usuarios).

| Entidad | Campos principales |
| --- | --- |
| Venta | id, folio (V-000232, consecutivo por empresa), sucursal, caja, usuario que cobró, fecha y hora, subtotal, IVA, total, forma de pago, recibido, cambio, referencia, estado (pagada o cancelada), motivo y usuario de cancelación |
| Línea de venta | venta, producto, código de barras, nombre, cantidad, precio unitario, costo unitario, importe |
| Movimiento de inventario | producto, sucursal, cantidad (negativa al vender, positiva al cancelar), tipo, venta de origen, fecha |

Reglas:

- **Búsqueda:** si el texto son solo dígitos (6 o más), se busca el código exacto; si no, por nombre o SKU sin importar acentos ni mayúsculas. Solo productos con precio en la sucursal activa.
- **Escaneo repetido:** suma 1 a la línea existente en lugar de crear otra.
- **Existencias:** no se puede agregar más de lo que hay en la sucursal; los agotados aparecen pero no se pueden agregar.
- **Precios con IVA incluido:** IVA = total − total ÷ 1.16; subtotal = total − IVA. Montos en centavos (enteros) o decimal exacto, nunca flotantes.
- **Verificar costos:** muestra costo y margen = (precio − costo) ÷ precio. Margen negativo = alerta en rojo; menor a 20 % = ámbar. Solo lo ve quien tenga el permiso de ver costos.
- **Cobro en efectivo:** «Confirmar cobro» se bloquea si recibido < total. Tarjeta y transferencia guardan la referencia opcional.
- **Confirmar cobro es una sola transacción:** se vuelven a validar existencias, se crea la venta con sus líneas, se descuenta el inventario y se asigna el folio. Si algo falla, no se guarda nada y se avisa qué producto ya no alcanza.
- **Cancelar venta:** marca la venta como cancelada, regresa las piezas al inventario con un movimiento positivo y pide motivo.
- **Inicio por puesto:** si el usuario solo tiene permisos de ventas, entra a 01b y ve el menú reducido; si no, entra al Inicio general.

## Fases de implementación

Cada fase depende de la anterior; no se empieza la siguiente sin cumplir su «Terminado cuando».

### Fase 0 · Revisión del proyecto

- [x] Identificar stack, estructura de carpetas, cómo se manejan sesión, sucursal activa y permisos.
- [x] Ubicar tablas existentes de productos, precios por sucursal, costos y existencias.
- [x] Proponer ajustes a este plan y la lista de archivos a crear o modificar.

**Terminado cuando:** apruebas el plan ajustado. No se escribe código en esta fase.

### Fase 1 · Datos y lógica de venta

- [x] Migraciones de Venta, Línea de venta y Movimiento de inventario.
- [x] Servicio de búsqueda de productos (código exacto y por nombre) con precio y existencia de la sucursal.
- [x] Servicio «registrar venta» transaccional con folio consecutivo y descuento de inventario.
- [x] Servicio «cancelar venta» que regresa piezas.
- [x] Permisos nuevos: ventas.registrar, ventas.ver, ventas.cancelar, ventas.ver\_costos.

**Terminado cuando:** las pruebas de la sección de pruebas pasan sin interfaz.

### Fase 2 · Pantalla de captura (tablero 04b)

- [x] Campo de captura con foco automático; Enter agrega, Escape limpia; el lector de códigos funciona como teclado.
- [x] Desplegable de coincidencias con precio y existencia.
- [x] Lista con cantidad +/−, eliminar con «Deshacer» y la última fila resaltada.
- [x] Interruptor «Verificar costos» con costo, margen, utilidad y alerta de venta bajo costo.
- [x] Resumen con piezas, subtotal, IVA y total.
- [x] La lista no se pierde si se recarga la página por accidente.

**Terminado cuando:** puedes capturar 10 productos solo con el lector, sin tocar el mouse.

### Fase 3 · Cobro y venta terminada

- [x] Ventana de cobro: efectivo con billetes rápidos y cambio (tarjeta y transferencia quedan pendientes).
- [x] Confirmar llama al servicio de la fase 1; muestra errores de existencia si los hay.
- [x] Pantalla final con folio, total, cambio, «Imprimir ticket» (versión imprimible del navegador) y «Nueva venta».

**Terminado cuando:** una venta queda guardada, el inventario baja y el folio sube en 1.

### Fase 4 · Historial de ventas (tablero 04)

- [x] Lista con totales de hoy, semana y mes, filtros por folio, cajero y forma de pago.
- [x] Detalle del ticket con reimprimir y cancelar (con motivo).
- [x] Un cajero solo ve sus propias ventas.

**Terminado cuando:** cancelar un ticket lo tacha y regresa sus piezas.

### Fase 5 · Accesos rápidos

- [x] Botón «Nueva venta» con F2 arriba del menú lateral en todas las pantallas.
- [x] Botón en la bienvenida de Inicio y primera tarjeta de «Ir directo a».
- [x] F2 abre la venta desde cualquier pantalla, salvo cuando el foco está en un formulario con cambios sin guardar.

**Terminado cuando:** desde cualquier pantalla llegas a la venta con una tecla.

### Fase 6 · Inicio para vendedor (tablero 01b)

- [ ] Redirección al inicio de vendedor según permisos. *(Pendiente: fase cerrada por decisión del usuario el 2026-10-28.)*
- [x] Menú reducido por permisos (Inicio y Ventas para el Cajero). *(«Mis ventas» y «Consultar precios» pendientes.)*
- [ ] Números del turno, últimas ventas, consulta de precio y aviso de productos por agotarse. *(Pendiente.)*
- [x] Bloquear por URL las secciones que el cajero no tiene permitidas, no solo ocultarlas.

**Terminado cuando:** un usuario cajero no puede abrir ninguna sección fuera de ventas.

## Pruebas y casos borde

Estas pruebas se escriben en la fase 1 y se corren al final de cada fase; los datos de ejemplo del lienzo (Coca-Cola 600 ml, Leche Lala, Galletas Marías agotadas) sirven como datos semilla.

| Caso | Resultado esperado |
| --- | --- |
| Escanear un código existente dos veces | Una línea con cantidad 2 |
| Escanear un código que no existe | Aviso «No encontramos el código…», la lista no cambia |
| Agregar un producto agotado | Aviso de sin existencias, no se agrega |
| Pasar del stock con el botón + | El botón se desactiva al llegar al tope |
| Otro cajero vende la última pieza mientras capturas | Al confirmar, error que nombra el producto; no se guarda nada |
| Efectivo con recibido menor al total | «Confirmar cobro» bloqueado y muestra «Falta» |
| Total con IVA de $64.00 | Subtotal $55.17, IVA $8.83 |
| Dos ventas confirmadas al mismo tiempo | Folios distintos y consecutivos |
| Cancelar un ticket | Estado cancelada, piezas de vuelta, total tachado |
| Producto con precio menor al costo y «Verificar costos» activo | Margen en rojo y alerta con enlace a Validar costos |
| Cajero abre la URL de Compras | Acceso denegado o redirección a su inicio |
| Usuario sin ventas.ver\_costos | No ve el interruptor «Verificar costos» |
