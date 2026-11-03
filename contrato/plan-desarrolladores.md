# Plan de desarrollo: Stockion

- **Para:** thrs e Ilian.
- **Estado:** propuesta para revisión. Las asignaciones se confirman entre ambos.
- **Arranque:** el usuario confirmó que este plan todavía no se ha iniciado. El POS y las ramas citadas en el diagnóstico son antecedentes, no tareas de este plan ya comenzadas.
- **Fecha:** actualizado el 2026-11-02.
- **Base confirmada:** 20 horas semanales por persona y cierre en un máximo de 4 meses. En [preguntas](preguntas.md), thrs propone 20–30 h semanales por persona; el aumento y las fechas siguen pendientes de Ilian. No se cuenta con 30 h para prometer entregas. El reparto y el calendario siguen como propuesta; no son aprobación de capabilities.
- **Preferencia confirmada de reparto:** thrs conserva almacenamiento y asume endurecimiento de autenticación; Ilian se enfoca en ventas. El resto se organiza por secciones completas para trabajar en paralelo en cada fase.
- **Regla:** este plan no autoriza implementar. Cada tarea necesita su `capabilities/NN-nombre.md` en estado `Aprobada` (ver `AGENTS.md`). ID candidato según el diagnóstico de ramas: 38; confirmar y reservar con Ilian antes de usarlo.

**Detalle en otros archivos:**

- [Plan maestro](plan-maestro.md): diagnóstico, hallazgos de seguridad, arquitectura sin internet y criterios.
- [Servicios externos](servicios-externos.md): almacenamiento, tarjeta, recargas y pasarela de pagos.
- [Ideas de pentest](ideas-pentest.md): pruebas de seguridad propias.
- [Plan para la Lic. Mari](plan-lic.md): compromiso con la clienta.

## Las fases

**Fase 0** nombra la base existente que la Lic. debe revisar y aceptar; no añade semanas a este plan ni acredita que todas las capabilities previas estén cerradas. El POS previo sigue pendiente de integración y validación antes de sustituir Odoo. Las Fases 1–3 constituyen la entrega de cuatro meses; Fase 4 es el inventario interno posterior.

| Fase | Objetivo | Ventana propuesta |
| --- | --- | --- |
| 1. Control confiable | Datos reales seguros y la Lic. decidiendo desde el celular. Odoo sigue en caja | Construcción S1–8; piloto S9–12 |
| 2. Mostrador sin fallas | POS completo, caja, modo sin internet y reportes. Se apaga Odoo tras superar las compuertas | Base técnica desde S2; construcción hasta S12; transición S13–14 |
| 3. Revisión final | Seguridad y pruebas desde cada entrega; cierre sin módulos nuevos | Transversal S1–16; cierre S15–16 |
| 4. Innovación y plataforma | **Interna, no se reporta a la Lic.** Soporte, imágenes, inteligencia y cobros integrados | Después de S16; sin fecha comprometida |

**Límite: 4 meses; horizonte interno de 16 semanas.** Son 320 horas por persona y 640 del equipo. Cuatro meses calendario no equivalen exactamente a 16 semanas: fijar fecha de inicio y fecha límite en D-19. Ausencias, reuniones, aprobaciones, revisión, pruebas y despliegues consumen esta misma capacidad.

**Fechas propuestas por thrs, pendientes de Ilian y de agenda con la Lic.:** inicio 2026-11-10; fecha límite 2027-03-10. El hito a 16 semanas cae el 2027-03-02, ocho días antes del máximo. Este margen de calendario no agrega automáticamente horas a las 640 ni permite consumir la reserva dos veces. Con 30 h por persona serían 960 h teóricas en 16 semanas; es un escenario pendiente, no la capacidad comprometida.

El cálculo anterior era 33–42 semanas, equivalentes a 1,320–1,680 horas del equipo con la misma dedicación. Pasar a 640 horas exige reducir entre 52 % y 62 % de ese tiempo previsto. **El alcance completo todavía no tiene una estimación que demuestre que cabe.** Los agentes de código son apoyo, no una reducción de horas garantizada. No se elimina alcance ni se extiende el plazo automáticamente: cualquier ajuste debe resolverse explícitamente en D-20.

### Presupuesto de capacidad propuesto

La capacidad se controla por persona. Se retira el reparto anterior de horas por paquetes: asignaba horas de seguridad a Ilian y dividía secciones que ahora tendrán un único responsable. **No hay todavía estimaciones de esfuerzo verificadas por tarea.**

| Persona | Capacidad semanal | Trabajo planificable propuesto | Reserva propuesta | Total en 16 semanas |
| --- | --- | --- | --- | --- |
| thrs | 20 h | 256 h | 64 h | 320 h |
| Ilian | 20 h | 256 h | 64 h | 320 h |
| **Equipo** | **40 h** | **512 h** | **128 h** | **640 h** |

Seguridad, coordinación y pruebas consumen estas horas. La igualdad de capacidad no demuestra igualdad de carga: thrs concentra administración en Fase 1 e Ilian concentra mostrador en Fase 2. Ilian adelanta su propio núcleo sin conexión mientras thrs completa administración; thrs prepara sus módulos de Fase 2 mientras Ilian completa mostrador. Estimar horas por persona y semana antes de confirmar viabilidad. Un exceso se resuelve en D-20; no se traspasan archivos de otro responsable para simular equilibrio.

Las exportaciones E1-17, E1-11.11 y E2-12, incorporadas en esta versión, deben incluirse en esa estimación. Compartir formato evita inconsistencias, pero no hace gratuito su desarrollo ni lo convierte en trabajo ya incluido en las cifras históricas.

## Ajustes incorporados respecto al maestro original

1. **La Etapa 3 cambió.** Ahora es revisión final. Lo de plataforma vendible pasa a [Fase 4](#fase-4-innovación-y-plataforma-interna), interna y no comprometida con la Lic.
2. **Contrato y esquema de ventas antes de la línea base.** Hay un POS previo como base; este plan aún no ha arrancado. Ilian entrega contrato y modelo estabilizados de E1-05; thrs los incorpora a E1-03 sin esperar la aceptación integral de Ventas. Después se valida la integración fiscal real con E1-06. Así no se forma un ciclo entre ventas, migraciones e impuestos.
3. **Tarea nueva E1-16,** almacenamiento. El plan maestro necesita respaldos externos y foto de ticket, pero no definía dónde guardarlos.
4. **La estimación anterior no justifica el nuevo plazo.** La Fase 1 acumulaba unas 35 semanas-persona al sumar tamaños máximos, sin ser una estimación detallada. Desglosar horas restantes y validar capacidad al terminar S2; no aplicar un factor de aceleración supuesto.
5. **Granel y venta por caja entran desde el contrato de ventas** (D-5).
6. **Plus nuevos:** recargas y servicios, pasarela de pagos y planes de uso.
7. **Reparto por secciones:** autenticación, todas las tareas de Compras y administración quedan con thrs; ventas, clientes y envases, inventario, migración de Odoo y motor sin conexión quedan con Ilian. El reparto no cambia el alcance funcional.

## Cómo evitamos conflictos

| Carril | Dueño | Zonas |
| --- | --- | --- |
| **Base y administración** | thrs | Infraestructura, CI, almacenamiento, respaldos, `cuenta`, seguridad del servidor, `database` y línea base; catálogo, impuestos, `precio`, **Compras completo**, bitácora, buzón, Inicio administrativo, aprobación móvil, terminales y reportes |
| **Mostrador y existencias** | Ilian | `venta`, `features/ventas`, caja, cobros, devoluciones, clientes y envases, `inventario`, conteos, transferencias y migración de Odoo; **motor sin conexión completo**, cliente y servidor de sincronización, IndexedDB/outbox, service worker, manifest y hardware |

Cada sección incluye sus modelos, contratos de aplicación, persistencia, rutas, pantallas y pruebas cuando corresponda. Los nombres de futuros dominios y sus rutas se fijan en sus capabilities aprobadas; esta tabla no autoriza crear carpetas ni reorganizar dominios existentes. Inicio administrativo pertenece a thrs; el inicio operativo del vendedor pertenece a Ilian dentro de Ventas.

### Paralelo por fase

| Fase | thrs | Ilian | Punto de integración |
| --- | --- | --- | --- |
| 1. Control confiable | Seguridad, almacenamiento, respaldos, migraciones, Compras/Precios y autorizaciones | Ajuste de ventas, inventario y migración de Odoo; prueba de hardware y adelanto de su núcleo sin conexión | Dinero e impuestos, movimientos de inventario, permisos y publicaciones al buzón |
| 2. Mostrador sin fallas | Terminales y credenciales, reportes y compras sin conexión | Sincronización completa, venta local, caja, cobros, devoluciones, clientes/envases, equipo y conteos sin conexión | Identidad de terminal, envases como producto de catálogo, consultas para reportes y transporte de compras/conteos |
| 3. Revisión final | Seguridad, aislamiento, carga del servidor, recuperación y cierre administrativo | E2E operativos, pruebas sin conexión, hardware y cierre de mostrador/inventario | Revisión cruzada de evidencia y validación integral con la Lic. |

**Archivos compartidos:**

| Archivo | Regla |
| --- | --- |
| `backend/internal/database/database.go` | thrs mantiene arranque y línea base. Ilian entrega el esquema de sus dominios; no modifica este archivo en paralelo |
| `backend/internal/domain/catalog_models.go` y persistencia heredada de catálogo | thrs mantiene archivos existentes. Ilian acuerda cambios necesarios para inventario mediante contrato; no modifica a la vez el modelo compartido |
| `backend/cmd/api/main.go` | thrs integra registros y dependencias. Ilian entrega el registro de rutas de su módulo en su propia función |
| `domain/negocio/permiso.go`, `rol.go`, `contexto/permisos.ts` | thrs integra permisos y rol Capturista; Ilian define necesidades de Ventas/Inventario y las verifica |
| `app.routes.ts`, `sidebar` e Inicio administrativo | thrs integra navegación. Ilian mantiene rutas internas y entrada del vendedor dentro de Ventas |
| Service worker, manifest, configuración PWA y base IndexedDB | Ilian es único responsable. thrs consume el contrato local para Compras, terminales y avisos; no modifica el motor |
| `PLAN_IMPLEMENTACION.md` y `DOCUMENTACION.md` | thrs consolida índices y entrega. Ilian aporta avances y documentación de sus módulos |
| `capabilities/` y archivos nuevos de migración | Cada responsable escribe la capability y las migraciones de su dominio; reservan IDs y nombres antes. thrs coordina orden y aplicación de migraciones |

Estos responsables aplican a las modificaciones previstas por este plan. Se conserva cualquier trabajo previo de colaboradores y se respeta el alcance de cada capability. Una migración de inventario puede requerir revisión de catálogo; eso es una integración coordinada, no dos personas editando el mismo archivo.

**Reglas de trabajo:**

- Cambios pequeños e integración frecuente. La estrategia de ramas se acuerda entre ambos; crear ramas, integrar o realizar cualquier escritura Git requiere la confirmación exigida por `AGENTS.md`.
- Entre dominios solo se comparten identificadores e interfaces de `application`, nunca repositories.
- **Compras tiene un solo responsable: thrs**, incluidos dinero, aprobación, cuadre, adjuntos, correcciones y órdenes. E1-09 y E1-10 se integran después de E1-06; Ilian no edita esa sección.
- **Sincronización tiene un solo responsable: Ilian**, desde copia local hasta recepción de lotes en servidor. Terminales y autenticación siguen con thrs y exponen un contrato explícito.
- Revisar trabajo ajeno no transfiere su propiedad. El revisor registra el hallazgo y el responsable corrige su sección. Cambiar un archivo ajeno requiere acuerdo e inclusión expresa en capability aprobada.

### Contratos que permiten avanzar sin esperar toda la otra sección

Antes de construir cada integración, su capability define formatos, validaciones, errores, ejemplos, versión y responsable. Ambos revisan ese contrato; cada quien implementa su lado en sus propios archivos.

| Contrato | Responsable del contrato y proveedor | Consumidor | Se puede construir en paralelo | Integración que sí debe esperar |
| --- | --- | --- | --- | --- |
| Importes, impuestos, precio autorizado, unidad y presentación | thrs, E1-06; Ilian define necesidades de venta E1-05 | Ventas e inventario de Ilian | Pantalla y casos de venta con ejemplos acordados | Venta final y persistencia con desglose real |
| Registrar y consultar movimientos de existencias | Ilian, E1-11 | Compras y catálogo de thrs | Compras contra interfaz y movimientos de ejemplo | Recepción real, kardex y conciliación |
| Publicar tareas al buzón y registrar auditoría | thrs, E1-04/07 | Ventas, inventario y sincronización de Ilian | Publicadores con proveedor sustituido en pruebas | Incidencias visibles y auditadas en conjunto |
| Identidad, permisos, terminal y credencial revocable | thrs, E1-03b/E2-02 | Venta local y sincronización de Ilian | Consumidores con identidades de prueba | Acceso local y sincronización con validación real |
| Cola local y sobre de operaciones por dominio | Ilian, E2-01/03 | Compras y terminales de thrs; conteos y envases de Ilian | Formularios y procesamiento de cada dominio con contrato de transporte | Reintentos reales, recuperación y ausencia de duplicados |
| Importar catálogo y registrar envase como producto ligado | thrs, catálogo | Migración E1-13 y envases E2-07 de Ilian | Migración y envases con productos de ejemplo | Carga real del catálogo de Odoo y saldo de envases sobre productos reales |
| Consulta de ventas, cortes y movimientos para reportes | Ilian para sus datos; thrs para compras/precios | Reportes de thrs | Reportes con datos de ejemplo y consultas acordadas | Cuadre de cifras y carga sobre información real |

Un proveedor sustituido o datos de ejemplo permiten desarrollar y probar cada lado; no cuentan como integración terminada. No se importan repositories ajenos ni se escribe directamente en tablas del otro dominio para evitar una espera.

## Fase 1: Control confiable

Tamaños heredados: **S** ≤ 3 días · **M** ≤ 2 semanas · **L** ≤ 4 semanas. Son referencias anteriores; no deben sumarse como si fueran horas disponibles en este calendario. Sustituirlos por horas restantes, dependencia y criterio probado en cada capability.

### Trabajo de thrs

| Código | Tarea | Descripción | Qué se espera | Tam. | Depende de |
| --- | --- | --- | --- | --- | --- |
| E1-01 | Estabilización | CI con PostgreSQL, dependencias, secretos y pruebas de sus secciones; Ilian corrige fallas de sus módulos | CI obligatorio y verde; integración ejecutada, no omitida; cero vulnerabilidades altas alcanzables | M | — |
| E1-03b | Endurecer autenticación | Límites por IP, mensajes uniformes, contraseña segura en recuperación, tiempos del servidor, tokens ocultos en logs; integrar rol Capturista y permisos acordados con Ilian | Hallazgos de plan maestro §6 cerrados y con prueba; permisos centrales consistentes | M | E1-01 para cierre; contrato de permisos con E1-05 |
| E1-16 | Almacenamiento | Preparar R2 y buckets con autorización de contratación aparte; interfaz `AlmacenArchivos` con API S3 | Subir y leer archivo privado con URL firmada, aislado por negocio; portabilidad de operaciones básicas a S3 probada y retención configurada por proveedor | S | — |
| E1-02 | Respaldos y monitoreo | Respaldo cifrado externo, restauración, staging, errores centralizados y alertas | Restauración cronometrada y documentada; alerta real probada. **Sin esto no entran datos reales** | M | E1-16 |
| E1-03 | Migraciones versionadas | Propuesta: adoptar `goose` (D-6 pendiente), línea base con el esquema estabilizado de ventas, retirar `AutoMigrate` al arrancar | Base nueva y base existente llegan al mismo esquema; respaldo previo y reversa | M | E1-01 y entrega de contrato/modelo E1-05; no exige aceptación fiscal final de Ventas |
| E1-04 | Bitácora | Registro por negocio, solo inserción: quién, qué, cuándo, antes y después | Acciones sensibles registradas; nadie puede editarlas | M | E1-03 |
| E1-06 | Impuestos y dinero | Impuestos por producto, dinero exacto en compras y precios, existencias en `numeric` para granel, corregir SKU | Desglose correcto validado con contador; existencias sin error de redondeo; SKU duplicado rechazado | L | E1-03 |
| E1-07 | Buzón de decisiones | Tabla `tareas_negocio`, interfaz para publicar, bandeja por prioridad y sucursal. Tipos: costo o precio por autorizar, producto nuevo, diferencia contra ticket, diferencia de conteo, transferencia incompleta, cancelación, corte con diferencia e incidencia de sincronización. Los tipos se definen en Fase 1; los eventos de caja y sincronización se habilitan al integrar sus módulos en Fase 2 | Módulos publican sin importar repositories; la Lic. resuelve desde la bandeja | M | E1-04 |
| E1-08 | Validación y aprobación | Producto nuevo `PENDIENTE_VALIDACION`, aprobación masiva, reglas y alerta de margen | Producto sin validar no se vende; aprobar por proveedor o porcentaje | M | E1-07 |
| E1-09 | Cuadre contra ticket | Total del documento con diferencia explicada, foto adjunta, nota de corrección autorizada | Diferencia sin explicar no se confirma; corrección deja rastro | M | E1-06, E1-16, contrato E1-07 |
| E1-10 | Órdenes de compra | Órdenes y sugerencia de compra según ventas y mínimos, mediante consultas acordadas con Ilian | Orden se convierte en recepción; sugerencia explicable | M | Contrato E1-05 y E1-06; ventas reales para validación final |
| E1-12 | Inicio administrativo | Pendientes del día y vista consolidada de sucursales | El propietario ve lo urgente al entrar | S–M | E1-07 |
| E1-17 | Exportación administrativa | Descarga a Excel de compras, productos y precios, por sucursal o consolidado; formato común de exportación acordado con Ilian; respeta permisos (costos solo con permiso) y registra la descarga en bitácora | La Lic. descarga su información para análisis propio; nadie exporta datos de otro negocio | S–M | E1-04, E1-06 |
| E1-14 | Aprobación móvil y avisos | Vistas móviles del buzón, suscripciones y envío de push; consumir la base PWA de Ilian E2-01 | La Lic. instala, recibe aviso y aprueba a 360 px; manifest, worker y motor local los mantiene Ilian | M | E1-07; base PWA E2-01 para instalación y push |

### Trabajo de Ilian

Resumen de Ilian en Fase 1. El desglose detalla el alcance de esta versión y no se cuenta dos veces. Sus entregables, incluidas las nuevas exportaciones, deben estimarse antes de afirmar que caben en el tamaño heredado.

| Código | Tarea | Qué se espera al cerrar | Tam. | Depende de |
| --- | --- | --- | --- | --- |
| E1-05 | **Ajustar e integrar ventas** (plan maestro §5.3) | Ventas integrada con contrato listo para sincronizar y vender por pieza, caja o granel; doble envío no duplica; prueba manual con lector | M | Reglas de granel/cajas; contrato de E1-06 para construir y validación real para cerrar |
| E1-11 | **Inventario** (plan maestro §5.2) | Transferencia real cerrada; diferencias llegan al buzón; movimientos expuestos a Compras | L | Contratos de movimientos y E1-07; E1-06 para precisión final y buzón real para aceptación |
| E1-13 | **Migración de Odoo** | Catálogo real con conteo inicial en ambas sucursales; negativos identificados resueltos (215 en el diagnóstico original, por contrastar con exportación vigente); exportación aceptada por Odoo | M | E1-06, E1-11; contrato de importación de catálogo de thrs |

#### E1-05 desglosado: ajustar e integrar ventas

**Hito A, contrato y modelo (S1–4).** Lo necesita thrs para la línea base E1-03. **Hito B, integración real (después de E1-06).**

| Sub | Qué hacer | Hallazgo §5.3 | Qué se espera | Hito | Depende de |
| --- | --- | --- | --- | --- | --- |
| E1-05.1 | Escribir la capability de Ventas con contrato versionado: campos, errores, ejemplos y `schema_version` | — | Contrato revisado por thrs | A | Respuesta de la Lic. sobre granel/cajas |
| E1-05.2 | Idempotencia: `id_cliente` UUID con índice único; un reintento devuelve la venta existente | #1 | Doble envío no duplica, con prueba | A | E1-05.1 |
| E1-05.3 | Folio por terminal `{sucursal}-{terminal}-000123`, en lugar de consecutivo por negocio | #2 | Folios únicos sin consultar al servidor | A | Formato de código de terminal acordado con thrs (E2-02) |
| E1-05.4 | Columnas `terminal_id` y `turno_id` en la venta | #3 | Cada venta trazable a caja y turno | A | Identificadores acordados con E2-02 y E2-04 |
| E1-05.5 | Fechas `vendida_en` y `recibida_en` en lugar de solo `creado_en` | #4 | Se distingue cuándo se vendió y cuándo llegó | A | E1-05.1 |
| E1-05.6 | Modo sincronización: una venta legítima sin existencia se conserva y abre incidencia, con validación de identidad y pertenencia de D-22 | #5 | Stock insuficiente no borra una venta ocurrida; control local avisa y la conciliación respeta el único flujo local primero | A modelo; B buzón real | Interfaz de buzón E1-07 (proveedor sustituido mientras tanto); D-22 |
| E1-05.7 | Retirar de la capability el `CHECK stock_actual >= 0`; control preventivo de existencias en mostrador sin crear otro flujo de persistencia para ventas con internet | #6 | Ningún `CHECK` bloquea la sincronización; venta local primero también cuando hay red | A | Contrato E1-05.1 |
| E1-05.8 | Línea de venta con foto de impuestos, en lugar de IVA fijo 16 % | #7 | Desglose por línea igual al que calcula E1-06 | A forma; B valores reales | Contrato E1-06 (A); E1-06 implementado (B) |
| E1-05.9 | Tabla `venta_pagos` y campo de descuento, aunque la pantalla de cobro espere a E2-05 | #8 | Una venta admite varias formas de pago | A | E1-05.1 |
| E1-05.10 | Cantidades decimales con unidad y presentación: pieza, caja (factor de conversión) y granel | #9 | A: tipos, precisión y factores acordados; B: existencia descontada correctamente con persistencia real | A contrato/modelo; B integración | D-5 y contrato E1-06 (A); existencias `numeric` implementadas en E1-06 (B) |
| E1-05.11 | Lista de permisos que requiere Ventas (vender, cancelar, ver costos, descuento); thrs los integra en E1-03b | §5.4 | thrs recibe la lista; Ilian verifica el resultado | A | Contrato de permisos con E1-03b |
| E1-05.12 | Entregar a thrs el esquema estabilizado de ventas para la línea base, sin editar `database.go` | — | thrs incorpora el esquema en E1-03 sin esperar la implementación fiscal o `numeric` de E1-06 | **Fin de A** | Partes de contrato/modelo del hito A de E1-05.2 a E1-05.10; no hitos B |
| E1-05.13 | Pendientes propios: listado sin límite fijo de 1000 (paginación) y elemento «Ventas» del menú con permiso (thrs integra en `sidebar`) | §5.3 | Listado paginado; menú oculto sin permiso | B | E1-03b para el permiso |
| E1-05.14 | Validar la integración fiscal real con E1-06 y prueba manual con lector | — | Desglose real correcto y captura con lector comprobada | B | E1-06 terminado |

El inicio del vendedor (01b) no entra en E1-05: queda en E2-03b.

#### E1-11 desglosado: inventario

| Sub | Qué hacer | Qué se espera | Depende de |
| --- | --- | --- | --- |
| E1-11.1 | Capability del dominio `inventario` y contrato de movimientos (registrar y consultar) para Compras de thrs | Contrato revisado por thrs; Compras puede construir contra él | E1-05.1 |
| E1-11.2 | Movimientos como fuente de verdad: se registra `-2`, nunca `stock = 8`; `stock_actual` queda como proyección | Existencia = suma de movimientos, con prueba | E1-11.1; existencias `numeric` de E1-06 |
| E1-11.3 | Kardex por producto y sucursal con filtro por tipo de movimiento (datos de `movimientos_inventario`) | Historial consultable y cuadrado contra existencia | E1-11.2 |
| E1-11.4 | Mermas con motivo, costo perdido y responsable | Merma descuenta existencia y queda auditada | E1-11.2; bitácora E1-04 para auditoría real |
| E1-11.5 | Conteos físicos por zona y a ciegas; la diferencia requiere autorización | El contador no ve la existencia esperada; diferencia autorizada genera movimiento | E1-11.2 |
| E1-11.6 | Publicar diferencias de conteo e incidencias al buzón | Diferencias visibles en la bandeja de la Lic. | Interfaz E1-07 (sustituida mientras tanto); buzón real para aceptación |
| E1-11.7 | Transferencias entre sucursales: solicitud, envío, **en tránsito**, recepción y diferencia | Transferencia real cerrada entre ambas sucursales | E1-11.2 |
| E1-11.8 | Alertas de mínimo y de producto sin movimiento; edición masiva de `stock_minimo` (hoy sembrado en 0) | Alertas útiles con mínimos reales | E1-11.2; cambio de modelo acordado con thrs (ver nota) |
| E1-11.9 | Decidir lotes y caducidad: FEFO real o retirar la promesa (hoy la venta no descuenta `lotes.cantidad_actual`) | Decisión registrada en la capability; implementada si se elige FEFO | Consulta con la Lic. |
| E1-11.10 | Movimientos de carga inicial y de ajuste para E1-13 | Existencias iniciales y negativos se corrigen con movimientos, nunca escribiendo `stock_actual` | E1-11.2; antes de S8 |
| E1-11.11 | Exportar a Excel existencias, kardex, mermas, conteos y transferencias con el formato común de E1-17 | Descarga por sucursal o consolidada, con permisos y registro en bitácora | E1-11.3; formato de E1-17 |

**Nota:** `stock_actual` y `stock_minimo` viven en `catalog_models.go`, archivo que mantiene thrs. Cualquier cambio a esos campos se pide por contrato; Ilian no edita el modelo compartido.

#### E1-13 desglosado: migración de Odoo

| Sub | Qué hacer | Qué se espera | Depende de |
| --- | --- | --- | --- |
| E1-13.1 | Capability de migración: fuente, mapeo de columnas, reglas de limpieza, ensayo en staging y reversa | Plan de migración revisado por thrs | E1-05.1, E1-11.1 |
| E1-13.2 | Importar catálogo real usando el importador XLSX existente o un contrato de importación de thrs; sin editar `catalog_models.go` | Catálogo completo con impuestos y SKU válidos de E1-06 | Contrato de importación de catálogo de thrs; E1-06 |
| E1-13.3 | Contrastar el diagnóstico de 215 negativos con exportación vigente; resolver negativos y conteo inicial en ambas sucursales con movimientos de E1-11.10 | Existencia inicial real, sin negativos sin explicar | E1-11.10; disponibilidad de la Lic. para contar |
| E1-13.4 | Exportar a Excel con el formato que acepta Odoo, mientras Odoo siga en caja | Archivo aceptado por Odoo en prueba real | E1-13.2 |
| E1-13.5 | Ensayo completo en staging con respaldo previo; carga en producción solo tras autorización y compuertas operativas y de privacidad | Migración repetible y cronometrada antes del piloto S9 | E1-02, E1-13.2 a E1-13.4, revisión legal inicial y autorización de operación externa |

#### Adelanto de Fase 2 durante Fase 1

Ilian adelanta partes de **E2-08, E2-01 y E2-03a/03b** bajo sus capabilities aprobadas. Se contabilizan una sola vez, en Fase 2; no son tareas adicionales.

| Sub | Qué hacer | Qué se espera | Semanas | Depende de |
| --- | --- | --- | --- | --- |
| E2-08.1 | Prueba temprana del equipo real: lector, impresora térmica, cajón y báscula si existe | Lista de equipo compatible; insumo para D-3 | 1–2 | Inventario de hardware de la Lic. |
| E2-01.1 | Base PWA: `@angular/service-worker`, manifest e iconos; excepción `no-cache` para `ngsw-worker.js` y `ngsw.json` en `nginx.production.conf` y Cloudflare | PWA instalable que recibe actualizaciones. thrs la consume en E1-14 | 1–4 | Capability E2-01 aprobada |
| E2-01.2 | IndexedDB (Dexie) con versiones de esquema; `navigator.storage.persist()` y monitoreo de cuota | Datos locales sobreviven al cierre del navegador | 1–4 | E2-01.1 |
| E2-01.3 | Copia inicial `GET /sync/snapshot` paginada y comprimida, sin costos para usuarios sin `ventas.ver_costos` | 5,000 productos descargados; búsqueda local < 100 ms | 3–8 | Contratos E1-05/06 |
| E2-01.4 | Bajada incremental `GET /sync/cambios` con cursor y tombstones; prueba con transacciones confirmadas en orden inverso | Ningún cambio se salta | 5–8 | Ilian define el registro de cambios y su cursor; thrs acuerda publicación transaccional desde sus dominios. La auditoría E1-04 no sustituye automáticamente ese registro |
| E2-01.5 | Contrato local para otros módulos: persistencia y cola sin editar el motor | thrs puede construir Compras sin red y terminales | 5–8 | E2-01.2 |
| E2-03a.1 | Servidor `POST /sync/lote` con `sync_operaciones` único por terminal + `op_id`; respuesta `DUPLICADA` | Reintento no duplica | 3–8 | Contrato de identidad E2-02 (identidades de prueba) |
| E2-03b.1 | Recorrido mínimo: venta escrita en local, outbox que no se borra sin ACK y reintento en staging | Venta local con reintento demostrada en staging (salida S3–4) | 3–4 | E2-01.2, E2-03a.1 |

La base PWA de E2-01 no depende del buzón ni de E1-14: Ilian la publica primero y thrs la consume para aprobación móvil. El resto de E2-01/03 (latido de conectividad, Web Locks, contador de pendientes, actualización sin venta en curso, incidencias y versiones) sigue en Fase 2.

### Piloto coordinado

| Código | Tarea | Coordinador | Aporte de cada sección | Qué se espera |
| --- | --- | --- | --- | --- |
| E1-15 | Piloto con la Lic. | thrs | thrs sigue Compras/Precios y datos; Ilian sigue existencias, conteos y transferencias. Cada quien corrige sus módulos | Cuatro semanas de compras reales con administración, catálogo y exportaciones E1-17/E1-11.11 listos, compuertas operativas y revisión legal inicial cumplidas; observaciones registradas |

### Orden sugerido

| Semanas | thrs: base y administración | Ilian: mostrador y existencias | Evidencia conjunta de salida |
| --- | --- | --- | --- |
| 1–2 | E1-01/16/02, contrato de dinero/permisos y preparación E1-03b; revisión legal inicial | Contrato y ajuste E1-05; prueba de hardware E2-08 y persistencia E2-01 | Restauración y alerta antes de datos reales; contratos iniciales, hardware y horas por persona estimadas, D-20 |
| 3–4 | E1-03 tras entrega de contrato/modelo E1-05, E1-03b/06/04; contratos de buzón y credenciales | Entregar contrato/modelo E1-05 y validar con E1-06; base E1-11 y recorrido mínimo E2-01/03 | Migraciones verificadas y venta local con reintento en staging; contratos listos para consumidores |
| 5–8 | E1-07/08/09/10/12/14/17, en secuencia según dependencias; contrato de importación de catálogo; E2-02 | E1-11, E1-13 y E2-01/03; empezar E2-04/05/06 al tener contrato estable | Administración, inventario y datos migrados listos para piloto S9; identidad y motor local integrados |
| 9–12 | Coordinar E1-15; E2-09/10a y correcciones administrativas | E2-03b/04/05/06/07/08/10b/12; correcciones de inventario del piloto | Piloto de compras completo y 13 escenarios aprobados antes de sustituir Odoo |
| 13–14 | Conciliar compras, reportes y exportaciones; capacitación administrativa | Coordinar E2-11, instalación y capacitación de cajas; seguimiento de cortes y envases | Una semana completa en ambas sucursales sin Odoo, con conciliación y aceptación |
| 15–16 | Seguridad, aislamiento, carga del servidor, restauración y cierre documental | E2E, pruebas sin conexión/hardware y cierre operativo | Cero errores graves abiertos; evidencia de recuperación y entrega aceptada |

E1-05 distingue dos hitos: primero contrato/modelo estabilizados para E1-03, después pruebas e integración real con E1-06. La aceptación integral de Ventas no bloquea la creación de la línea base. La prueba técnica temprana no autoriza saltar dependencias: adelantar código productivo requiere contratos estables y capabilities aprobadas. Si una compuerta falla, Odoo sigue activo y se registra riesgo de incumplimiento; no se declara terminado ni se mueve la fecha por defecto.

### Seguimiento para sostener los cuatro meses

- Cada persona registra hasta 20 h semanales, incluyendo coordinación y pruebas. Máximo una tarea de construcción activa por persona, además de revisión y soporte al piloto.
- Revisión semanal: horas consumidas y restantes, bloqueos, evidencia y reserva disponible. Un bloqueo se escala al siguiente día hábil.
- S1: resolver hardware, fiscalidad con contador, disponibilidad de la Lic. y definición de alcance. S2: validar presupuesto por bloque y ruta crítica. Si no cabe, resolver D-20 de inmediato.
- Sin nuevos plus durante estas 16 semanas. Cualquier cambio aprobado debe mostrar cómo conserva el límite de horas y fecha.
- Seguridad, aislamiento, restauración y aceptación se comprueban durante la construcción; F3 repite sobre la versión final.

## Fase 2: Mostrador sin fallas

| Código | Tarea | Responsable | Descripción | Qué se espera | Tam. | Depende de |
| --- | --- | --- | --- | --- | --- | --- |
| E2-01 | Núcleo local | Ilian | Service worker, manifest, IndexedDB, copia y bajada; contrato para que otros módulos usen persistencia y cola sin editar el motor | PWA base instalable, consulta sin red, 5,000 productos en menos de 100 ms | L | Contratos de E1-05/06; no depende de E1-14 |
| E2-02 | Terminales y PIN | thrs | Registro, credencial de dispositivo, revocación y módulo de PIN; usa persistencia local acordada con Ilian | Terminal revocable; contrato de identidad consumible por sincronización; cola con sesión vencida comprobada en integración | L | Contrato de identidad para construir; E2-01 y E2-03 para cierre integrado |
| E2-03a | Sincronización en servidor | Ilian | `/sync/lote`, bajada de cambios, idempotencia, incidencias y versiones; validar identidad mediante contrato de E2-02 y delegar operaciones a cada dominio | Reintento no duplica; operación legítima se conserva para conciliación; tratamiento de datos no confiables definido en D-22 | L | Contratos de E2-01/02, movimientos y buzón; proveedores reales para cierre |
| E2-03b | Venta local primero | Ilian | Venta escrita en local, outbox y contador de pendientes; inicio operativo del vendedor dentro de Ventas | Un solo camino de código; cola intacta al cerrar el navegador | L | Contratos de E2-01/02/03a para construir; integración completa para cierre |
| E2-04 | Caja y turnos | Ilian | Apertura, movimientos, corte a ciegas y diferencia autorizada; corte con diferencia se publica al buzón | Corte cuadrado por forma de pago | L | E2-03b |
| E2-05 | Cobros | Ilian | Tarjeta con terminal independiente y referencia manual, transferencia y pagos mixtos. Sin conexión de la terminal no hay cobro con tarjeta: el pago queda bloqueado y la caja ofrece efectivo; nunca se registra tarjeta sin autorización confirmada. Preferir terminal con datos móviles propios. El responsable revisa los detalles de D-15 | Corte separa cada forma de pago; registro local sin red, cobro bancario confirmado por separado | S–M | E2-03b |
| E2-06 | Devoluciones y presentaciones | Ilian | Devolución, descuento con límite por rol, mayoreo, pantalla de venta por caja y a granel; cancelación se publica al buzón | Encargado no cancela sus propias ventas; caja y granel descuentan la existencia correcta | M | E2-03b |
| E2-07 | Clientes y envases | Ilian | Módulo de clientes y saldos de envases; entregar, recibir y cobrar envases desde la caja. El envase como producto ligado se pide a catálogo de thrs por contrato; el saldo con proveedor, si la Lic. lo confirma, se integra con Compras de thrs mediante contrato | Saldo de envases correcto por cliente y flujo integrado en Ventas | L | E2-03b; contrato de envase en catálogo con thrs; respuesta de la Lic. sobre envases con proveedor |
| E2-08 | Equipo de caja | Ilian | Impresora térmica, cajón, lector, etiquetas y báscula si se confirma | Probado con el equipo real; decide D-3 | M | E2-03b |
| E2-09 | Reportes y panel | thrs | 5 cortes con Excel, por sucursal o consolidado; consultas acordadas, sin importar repositories de Ventas | Totales cuadran contra ventas; no afectan la operación | L | Contratos de consulta para construir; E2-04 y datos reales para cierre |
| E2-10a | Compras sin red | thrs | Borrador de compra dentro de Compras; consumir cola y transporte de Ilian sin modificarlos | Se envía al reconectar sin duplicar | Parte de E2-10 (M) | Contrato E2-01/03 para construir; motor real para cierre |
| E2-10b | Conteos sin red | Ilian | Conteo dentro de Inventario usando el mismo motor local | Diferencia y base de conteo conservadas; envío sin duplicados | Parte de E2-10 (M) | E1-11 y E2-01/03 |
| E2-12 | Exportación de ventas | Ilian | Descarga a Excel de ventas, detalle de tickets, cortes de caja, devoluciones y envases con el formato común de E1-17; sin costos para quien no tenga permiso | La Lic. descarga ventas y tickets por periodo y sucursal; totales cuadran con los cortes | S–M | E2-04; formato de E1-17 |
| E2-11 | Salida de Odoo | Ilian, coordinador | Piloto de caja, luego todas; capacitación de mostrador y seguimiento de cortes. thrs valida datos, reportes y administración | Una semana en ambas sucursales sin Odoo; aceptación escrita | M | Todas |

E2-10a y E2-10b desglosan el alcance existente de E2-10, sin duplicar su estimación. E2-12 exporta datos detallados de ventas; E2-09 de thrs arma reportes consolidados y su propio Excel. Cada dominio exporta sus datos; el formato común no autoriza un paquete `shared` o `utils`. El inicio de vendedor pendiente queda bajo Ventas en E2-03b; su alcance exacto requiere la capability correspondiente.

**Punto de coordinación:** Ilian mantiene el contrato de sincronización cliente/servidor; thrs revisa identidad, autorización e integración con sus dominios antes de construir sus consumidores. Los adaptadores de Compras y terminales pertenecen a thrs; los de ventas, envases y conteos, a Ilian; el transporte genérico y sus registros de operaciones pertenecen a Ilian y se ubican en el dominio concreto aprobado, no en un paquete `shared` o `utils`.

**Salida de fase:** los 13 escenarios de plan maestro §7.4 en verde y 10 productos capturados con lector en menos de 30 segundos. **Odoo no se apaga antes.**

Además, E2-12 debe exportar ventas, tickets, cortes, devoluciones y envases con permisos y cifras conciliadas con E2-09. Antes de retirar Odoo se confirma internamente la vía de facturación existente, sin ofrecer CFDI de Stockion (D-10).

## Fase 3: Revisión final

Sin módulos nuevos. Solo se corrige, se prueba y se endurece.

| Código | Tarea | Responsable | Descripción | Qué se espera |
| --- | --- | --- | --- | --- |
| F3-01 | Revisión de seguridad | thrs | Coordinar pruebas propias, sin pentest externo por ahora ([ideas](ideas-pentest.md)); registrar hallazgos y corregir los de sus módulos; Ilian corrige los de mostrador/inventario | Sin críticas ni altas abiertas; casos de aislamiento y dinero probados |
| F3-02 | Aislamiento | thrs | Pruebas automáticas entre negocios y sucursales; evaluar RLS | Ningún acceso cruzado; decisión de RLS documentada |
| F3-03 | Pruebas de punta a punta | Ilian | Mantener escenarios E2E de compra, autorización, venta, caja, inventario y sincronización; thrs aporta casos de administración y mantiene configuración central de CI | Flujos críticos en CI; fallas corregidas por el responsable de cada módulo |
| F3-04 | Repetir pruebas sin internet | Ilian | Los 13 escenarios sobre la versión final | Todos en verde, con evidencia |
| F3-05 | Carga y rendimiento | thrs | Carga del servidor, consultas y memoria en entorno equivalente; Ilian mide respuesta local, cola y hardware dentro de F3-04 | Metas cumplidas; ajustes de código hechos por el responsable de cada sección |
| F3-06 | Simulacro de desastre | thrs | Restaurar desde cero, medir RPO y RTO, documentar procedimiento | Metas acordadas cumplidas por cronómetro |
| F3-07 | Cerrar pendientes | thrs, coordinador | thrs reúne evidencia de 25/30/32/33/34 y administración; Ilian de 37 y mostrador/inventario. Cada quien corrige su sección; la aceptación la concede el usuario | Sin errores graves abiertos y estados respaldados por aceptación explícita |
| F3-08a | Deuda de riesgo administrativa | thrs | Corregir riesgos concretos en cuenta, catálogo, dinero, Compras, Precios e infraestructura | Cambios justificados por un riesgo y dentro de capability aprobada |
| F3-08b | Deuda de riesgo operativa | Ilian | Corregir riesgos concretos en ventas, inventario, caja y sincronización | Cambios justificados por un riesgo y dentro de capability aprobada |
| F3-09 | Legal mínimo | thrs, coordinador con abogado | Recopilar condiciones de ambos; preparar documentos antes de datos reales y revisar al cierre | Revisión inicial previa al piloto y versión final revisada por abogado |
| F3-10 | Entrega | thrs, coordinador | thrs consolida manual, `DOCUMENTACION.md` y acta; Ilian entrega guía de mostrador y evidencia de su sección | Validación integral de ambos y aceptación final de la Lic. |

F3-08a/b dividen el alcance previo de F3-08. Una tarea con coordinador tiene un solo responsable del documento o registro final; no convierte todos los archivos involucrados en propiedad compartida. Revisiones cruzadas y prueba final conjunta validan la entrega, mientras cada corrección sigue en su sección.

## Decisiones tomadas y propuestas pendientes

| ID | Decisión | Resultado |
| --- | --- | --- |
| D-2 | ¿Ajustar ventas antes de integrar? | **Sí** |
| D-5 | ¿Granel y venta por caja desde ya? | **Sí, ambos** |
| D-7 | Dedicación semanal | **20 horas por persona; 40 del equipo**, confirmado por el usuario el 2026-11-02 |
| D-19 | Fechas de trabajo | **Propuesta de thrs, pendiente de Ilian:** 2026-11-10 a 2027-03-10; hito interno de 16 semanas el 2027-03-02 |
| D-10 | Facturación electrónica | **No se ofrece a la Lic. por ahora.** Queda como F4-23 interna; se retiró la pregunta de su plan |
| D-14 | Almacenamiento | **Cloudflare R2 para iniciar.** S3 queda como alternativa o segunda copia; verificar compatibilidad y configurar retención por proveedor |
| D-15 | Tarjeta | **Terminal independiente en Fase 2.** Los detalles los revisa el responsable de E2-05 |
| D-16 | Recargas y servicios | **Posible plus,** en Fase 4 (F4-22) |
| D-17 | ¿Pentest externo? | **Por ahora no.** Pruebas propias; retomarlo antes del primer cliente externo |
| D-18 | Plazo máximo | **4 meses**, confirmado por el usuario el 2026-11-02; calendario interno de 16 semanas propuesto |

D-7 conserva 20 h como base; la propuesta posterior de thrs de 20–30 h y D-19 siguen pendientes de acuerdo con Ilian. Por eso D-19 se registra arriba como propuesta, no como decisión conjunta tomada. D-20 debe estimar la carga, incluidas exportaciones. Detalle comercial y demás decisiones abiertas: [preguntas](preguntas.md). Diagnóstico histórico: [revisión del plan](revision-plan.md). Opciones: [servicios externos](servicios-externos.md).

## Fase 4: Innovación y plataforma (interna)

> **Solo para desarrolladores.** No se presenta ni se reporta a la Lic. Mari esta fase, sus códigos, prioridades ni estrategia SaaS. El [plan para la Lic.](plan-lic.md) puede enumerar algunos posibles extras, sin comprometer su desarrollo ni fecha. Fase 4 no forma parte de los 4 meses ni de las 640 horas. Cada idea necesita su capability aprobada antes de implementarse. Son hipótesis: validar demanda antes de construir.

Los códigos E3 del plan maestro se conservan entre paréntesis. Responsable sugerido según el carril de cada sección; se confirma al priorizar.

### F4-A. Plataforma y soporte para nosotros

Ya existe una base: dominio `plataforma` con panel de lectura de negocios y métricas. Se trata de hacerlo robusto, no de empezar de cero.

| Código | Tarea | Qué incluye | Resp. sugerido | Depende de |
| --- | --- | --- | --- | --- |
| F4-01 | **Tablero de soporte robusto** (E3-12) | Estado por negocio, sucursal y terminal: última sincronización, cola pendiente, incidencias, errores, versión instalada; búsqueda de negocio; ayuda y tickets; estado del servicio | thrs; Ilian aporta métricas de sincronización | E1-02, E1-04, E2-03 |
| F4-02 | Acceso temporal de soporte | Acceso autorizado por el cliente, con vencimiento, alcance limitado y registro completo en bitácora | thrs | F4-01, E1-04 |
| F4-03 | MFA para propietario y administradores de plataforma | Pendiente del plan maestro §5.4 sin tarea asignada | thrs | E1-03b |
| F4-04 | Infraestructura SaaS (E3-07) | DB administrada, separar servicios, pruebas de carga con 50 negocios, RLS, SLO | thrs | E1-02, F3-02 |
| F4-05 | Planes de uso (E3-08) | Límites por plan, cobro recurrente con Stripe propuesto, suspensión y gracia; precios y proveedor se confirman antes de cobrar | thrs | F4-19 para definir planes; F4-07/08 antes de activar cobros reales |
| F4-06 | Alta autoservicio (E3-10) | Giros, importador guiado, datos demo; negocio operando en menos de 1 hora | thrs | F4-05 |
| F4-07 | Legal SaaS (E3-11) | Aviso de privacidad, términos, contrato de encargado, exportación y baja de datos | thrs con abogado | F3-09 |
| F4-08 | Pentest externo (E3-13) | Antes del primer cliente externo y de activar cobros reales de planes | thrs coordina | Versión candidata en staging, incluidos planes probados sin cobros reales; remediación antes de activación |

### F4-B. Catálogo e imágenes

| Código | Tarea | Qué incluye | Resp. sugerido | Depende de |
| --- | --- | --- | --- | --- |
| F4-09 | **Catálogo propio de imágenes en S3/R2** | Subida de imagen de producto, validación de tipo y tamaño, redimensión y miniaturas, aislamiento por negocio, URL firmada o CDN; sustituye a `ImagenURL` externa | thrs | E1-16 |
| F4-10 | Imágenes sin internet | Miniaturas en caché local para que la caja las muestre sin red, con límite de cuota | Ilian | F4-09, E2-01 |
| F4-11 | Catálogo maestro por código de barras | Biblioteca compartida de productos comunes (EAN, nombre, imagen, presentación) para dar de alta en segundos; cada negocio conserva sus precios y datos | thrs | F4-09, F4-06 |
| F4-12 | Refactor de catálogo (E3-09) | Catálogo a dominio propio y componentes grandes; al tocar la zona | thrs | E1-06 |

### F4-C. Inteligencia para el dueño

| Código | Tarea | Qué incluye | Resp. sugerido | Depende de |
| --- | --- | --- | --- | --- |
| F4-13 | Centro de decisiones (E3-01) | Protector de margen, capital atorado, sugerencia de compra, comparador de proveedores | thrs | E2-09 |
| F4-14 | Control antifraude | Cancelaciones, diferencias de caja y de conteo por empleado; patrones atípicos | Ilian | E2-04, E1-11, E1-04 |
| F4-15 | Resumen diario por WhatsApp (E3-02) | Ventas, margen, faltantes y pendientes; costo por mensaje | thrs | F4-13 |
| F4-16 | Captura de ticket con foto, IA (E3-03) | Lee el ticket y prellena la compra; medir precisión con tickets reales antes de prometer | thrs | E1-09 |
| F4-17 | Pronóstico de demanda y caducidad | Predicción por producto y temporada; alerta de caducidad próxima si se adoptó FEFO (E1-11.9) | Ilian | E2-09, E1-11 |
| F4-18 | Asistente de preguntas | El dueño pregunta en lenguaje natural («¿cuánto vendí ayer en la sucursal 2?») sobre consultas de solo lectura y aisladas por negocio | thrs | E2-09, F3-02 |
| F4-19 | Comparativo de competencia (E3-05) | Probar Eleventa, SICAR, Loyverse y otros; precios, planes y términos de servicios de terceros | Ambos | — |

### F4-D. Cobros, servicios y operación

| Código | Tarea | Qué incluye | Resp. sugerido | Depende de |
| --- | --- | --- | --- | --- |
| F4-20 | Terminal integrada | Mercado Pago Point por API y Webhooks; elimina la captura manual de referencia. Sigue requiriendo conexión para autorizar | Ilian | E2-05 |
| F4-21 | Pasarela de pagos | Cobros en línea de cada negocio; comparar Mercado Pago y Stripe | Ilian | Contrato de cobros E2-05 y revisión de seguridad/legal antes de activación; no requiere integrar antes una terminal Point |
| F4-22 | Recargas y pago de servicios | Taecel; requiere internet; cobrar solo en efectivo | Ilian | E2-05 |
| F4-23 | CFDI 4.0 (E3-04) | Por PAC; factura por ticket, autofactura con QR, factura global. Interna; no se ofrece a la Lic. por ahora (D-10) | thrs | E1-06 |
| F4-24 | Gastos, cuentas por pagar, crédito y promociones (E3-06) | Flujo de caja, saldos con proveedor, fiado, combos y descuentos por fecha | thrs; promociones en venta con Ilian | E2-04 |
| F4-25 | Pedidos de clientes | Catálogo público del negocio y pedidos para recoger o domicilio | Ilian | F4-09, F4-21 |
| F4-26 | Lealtad | Puntos o beneficios por cliente frecuente, ligado a clientes de E2-07 | Ilian | E2-07 |

**Prioridad sugerida al terminar Fase 3:** F4-01/02/03 (soporte seguro para operar a la Lic.), F4-09 (imágenes), F4-13/14 (núcleo de venta según plan maestro §9). El resto, según demanda validada.

## Primeros pasos

1. thrs e Ilian ratifican asignaciones, fechas propuestas D-19 y base de 20 h de D-7; acuerdan si el aumento propuesto a 20–30 h es viable. D-20 se estima con la capacidad realmente acordada.
2. Enviar preguntas a la Lic. Mari, sobre todo granel y cajas, antes de cerrar E1-05.
3. Ilian escribe la capability de E1-05.
4. thrs escribe las capabilities de E1-01, E1-16 y E1-02.
5. Desglosar horas por bloque, dependencias y trabajo verificable; resolver D-20 antes de comprometer el alcance completo. Ninguna capability cambia a `Aprobada` por esta actualización documental.
6. Acordar contratos de dinero/impuestos, movimientos, buzón, identidad, transporte y consultas; reservar los archivos centrales y migraciones antes de iniciar implementación. No hay tareas de este plan ya comenzadas.
