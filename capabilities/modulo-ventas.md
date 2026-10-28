# Módulo de ventas

Base vigente según código local. Actualizado: 2026-10-28 (Fases 1 y 2 de la [37](37-ventas.md)).

## Flujo y reglas

1. **Buscar:** `GET /ventas/productos?sucursal_id&q`. Primero código exacto en `producto_codigos` (cualquier longitud). Cada coincidencia trae `existencia_baja` (existencia > 0 y ≤ `stock_minimo`). Un texto de solo dígitos con 6 o más que no existe responde vacío con `por_codigo = true`. Si no, por nombre sin acentos ni mayúsculas (todas las palabras deben aparecer, máximo 20). Solo productos activos con `precio_venta > 0` en la sucursal; existencia en piezas completas y `agotado` cuando es 0. Costo, margen `(precio − costo) ÷ precio` y nivel (`negativo`, `bajo` < 20 %, `ok`) solo con `ventas.ver_costos`.
2. **Registrar:** `POST /ventas` en una transacción: suma líneas repetidas, bloquea `inventario_sucursal` en orden de `producto_negocio_id`, revalida existencias (409 `EXISTENCIA_INSUFICIENTE` con nombre del producto, sin guardar nada), guarda la foto de nombre, código, precio y costo, descuenta stock con `movimientos_inventario` tipo `VENTA_SALIDA` (negativo, `referencia` = folio) y reserva al final el folio `V-000001` consecutivo por negocio (`venta_folio_consecutivos`). Una venta fallida no consume folio.
3. **Cobro:** IVA 16 % incluido sobre el total: `subtotal = round(total ÷ 1.16)`, `iva = total − subtotal`. `EFECTIVO` exige recibido ≥ total (400 `PAGO_INSUFICIENTE`) y calcula cambio; `TARJETA` y `TRANSFERENCIA` cobran el total y guardan referencia opcional.
4. **Consultar:** `GET /ventas` (`periodo=hoy|semana|mes` en el calendario del negocio con `resumen` de vendido, tickets, ticket promedio, efectivo y cancelados, y `desde`/`hasta`; filtros `sucursal_id`, `folio`, `usuario_id`, `forma_pago`, `estado`, `desde`, `hasta` RFC 3339) y `GET /ventas/:ventaId`. Con `ventas.ver`, todas las ventas de las sucursales accesibles; solo con `ventas.registrar`, únicamente las propias. Lo ajeno responde 404. Cada venta trae `atendio` (nombre del perfil o correo) y, si se canceló, `cancelada_por`.
5. **Cancelar:** `POST /ventas/:ventaId/cancelar` con `motivo` de lista (`ERROR_CAPTURA`, `COBRO_DUPLICADO`, `DEVOLUCION_INMEDIATA`, `OTRO` con `detalle` obligatorio) y `detalle` opcional (`ventas.cancelar`). Solo ventas del mismo día en la zona horaria del negocio (409 `VENTA_FUERA_DE_PLAZO`); cada venta trae `cancelable`. Marca `CANCELADA` con usuario y fecha, y regresa piezas con `VENTA_CANCELACION` positivo. Una segunda cancelación responde 409 `VENTA_YA_CANCELADA`.

Montos en centavos (`bigint`) dentro de ventas; la API los intercambia como pesos con dos decimales convertidos desde el texto JSON, sin flotantes. Cantidades enteras. Lotes no se descuentan. Sin caja, corte, crédito, clientes ni facturación.

## Permisos

| Permiso | Uso | Requiere |
| --- | --- | --- |
| `ventas.registrar` | Buscar, cobrar y ver sus propias ventas | — |
| `ventas.ver` | Ver todas las ventas de sus sucursales | — |
| `ventas.cancelar` | Cancelar con motivo | `ventas.ver` |
| `ventas.ver_costos` | Costo y margen al vender y en tickets | `ventas.registrar` |

El rol Cajero por defecto tiene solo `ventas.registrar` (ver [gestión de colaboradores](gestion-colaboradores.md)); Propietario recibe todos.

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `ventas` | Negocio, sucursal, `numero`/`folio` (único por negocio), usuario, importes `*_centavos`, piezas, forma de pago, referencia, estado y cancelación. |
| `venta_lineas` | Producto comercial, código y nombre al momento, cantidad, `precio_unitario_centavos`, `costo_unitario_centavos` (nulo sin costo), `importe_centavos`, orden. |
| `venta_folio_consecutivos` | Siguiente número de folio por negocio. |
| `inventario_sucursal`, `movimientos_inventario` | Existencia y trazabilidad (`VENTA_SALIDA`, `VENTA_CANCELACION`). |
| `precios_sucursal`, `producto_codigos` | Precio/costo vigentes y códigos de barras. |

## Dónde cambiar y qué revisar

- Backend: [dominio](../backend/internal/domain/venta/), [aplicación](../backend/internal/application/venta/), [persistencia](../backend/internal/infrastructure/venta/) y [HTTP](../backend/internal/interfaces/http/venta/); permisos en `domain/negocio/permiso.go` y migración del Cajero en `infrastructure/negocio/rol_migration.go`.
- Pruebas: unitarias en cada paquete `venta` e integración `internal/database/ventas_integration_test.go` (`VENTAS_TEST_DATABASE_URL`, base `tienda_ventas_test`).
- Frontend: `/ventas/nueva` (`features/ventas/nueva-venta.component.*`, `ventas.routes.ts`; permiso `ventas.registrar`). Captura con foco automático (Enter agrega, Escape limpia, flechas en coincidencias), escaneo repetido suma 1 sin pasar la existencia, existencia en ámbar si `existencia_baja`, quitar con «Deshacer», «Verificar costos» con `ventas.ver_costos` y resumen en centavos (`venta-calculos.ts`). La lista se guarda en `localStorage` (`tienda.venta.borrador.{negocio}.{sucursal}`). «Terminar venta» abre el cobro solo en efectivo (recibido, billetes rápidos, «Falta»/cambio; tarjeta y transferencia visibles pero pendientes), confirma con `POST /ventas` y muestra la venta terminada con «Imprimir ticket» (ticket de 72 mm solo al imprimir) y «Nueva venta». `/ventas` (`ventas.component.*`): periodo Hoy/Semana/Mes, métricas, búsqueda por folio o cajero, forma de pago y tabla con «Ver ticket»; `/ventas/:ventaId` (`ticket-venta.component.*`): líneas, resumen, Reimprimir y Cancelar con motivo de lista. Ambas con `ventas.ver` o `ventas.registrar` (`ventasLecturaGuard`). El ticket impreso es `ticket-impreso.component.ts`. Accesos: botón «Nueva venta» con F2 en el menú lateral, botón en la bienvenida y primer acceso de Inicio; F2 global (`atajo-nueva-venta.service.ts`) no navega con cambios pendientes ni con el foco en un formulario con datos.
