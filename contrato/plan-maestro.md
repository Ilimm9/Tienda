# Plan maestro de producto — Stockion

- **Estado:** Borrador para revisión de thrs e Ilian.
- **Fecha:** 2026-10-30.
- **Audiencia:** desarrolladores. La versión para la clienta se deriva después (sección 13).
- **Base analizada:** rama `thrs` en `da9e6fb` (igual que `origin/production`) y rama `origin/ventas` en `5cd61d7`.
- **Alcance de este documento:** orienta y prioriza. **No autoriza implementación.** Cada bloque sigue necesitando su `capabilities/NN-nombre.md` aprobado, conforme a `AGENTS.md`.

---

## 1. Resumen ejecutivo

**Dónde estamos.** Hay una base buena y poco común para esta etapa: multiempresa, sucursales, RBAC, equipo, catálogo con variantes, importación, compras con propuestas de precio e historial. Ilian agregó un POS funcional en `origin/ventas`, sin integrar y sin prueba manual.

**Qué falta para venderlo.** El mayor hueco no es de pantallas, es de confianza operativa:

1. **No hay respaldos** y la propuesta los prometió desde el mes 1.
2. **No hay nada de modo sin internet**: ni service worker, ni base local, ni terminales. Ventas no está diseñada para sincronizar.
3. **El IVA fijo de 16 % no sirve para esta tienda**: vende alimentos, cerveza, licores y cigarros, con tasas e IEPS distintos.
4. **Faltan los módulos que convierten un POS en plataforma de administración**: buzón de autorizaciones, caja, transferencias, conteos, reportes y bitácora.
5. **Promesas de la propuesta sin construir**: órdenes de compra, envases retornables, conteos, exportación a Odoo, aprobación masiva.

**Decisión más urgente.** Ajustar el contrato de ventas **antes** de integrar la rama de Ilian (sección 5.3). Hoy cuesta días porque no hay ventas reales guardadas. Con datos en producción cuesta una migración delicada.

**Ruta propuesta:**

| Etapa | Nombre | Resultado | Duración estimada* |
| --- | --- | --- | --- |
| 1 | **Control confiable** | La Lic. administra compras, precios e inventario de ambas sucursales y autoriza desde el celular. Odoo sigue como POS. | 8–10 semanas |
| 2 | **Mostrador sin fallas** | POS completo, caja, modo sin internet probado y reportes. Se apaga Odoo. | 10–12 semanas |
| 3 | **Plataforma vendible** | Centro de decisiones, CFDI, suscripciones, alta autoservicio, legal y pentest. Primer cliente externo. | 10–12 semanas |

\* Hipótesis: dos personas con dedicación principal. Si es parcial, multiplicar por 1.5–2. El total de 28–34 semanas (6.5–8 meses) **rebasa los 5 meses de la propuesta**. La Lic. ya dijo que acepta más tiempo si se cubren las necesidades reales. Recalibrar después de los dos primeros bloques: en este proyecto el código sale rápido y el cuello de botella es la validación manual.

---

## 2. Visión de producto

Observación central de la Lic.: *«no vendan su producto como un punto de venta, sino como una plataforma o sistema de administración de negocios»*.

**Posicionamiento de trabajo:** Stockion le dice al dueño qué decidir hoy y le permite decidirlo desde el celular, aunque la tienda se quede sin internet.

Principios para decidir alcance:

1. **Cada dato de operación debe terminar en una decisión.** Un ticket capturado no es el final: dispara cambio de costo, margen, autorización y sugerencia de compra.
2. **El mostrador nunca se detiene.** Una caída de internet no frena una venta.
3. **Nada sensible cambia sin responsable.** Toda autorización y cambio deja rastro.
4. **El dueño trabaja por excepción.** Ve lo que requiere atención, no pantallas para revisar.
5. **Se construye sobre operación real.** Lo que hoy resuelven los Excel de la Lic. es la mejor especificación disponible (sección 4.3).

---

## 3. Estado actual por módulo

Madurez: **A** usable y probado · **B** funciona con huecos · **C** parcial · **D** no existe.

| Módulo | Madurez | Qué hay | Qué falta |
| --- | --- | --- | --- |
| Cuenta y acceso | B | Sesión opaca con hash, CSRF, OTP, recuperación, bloqueo por intentos | Cambio de contraseña con sesión, gestión de sesiones, MFA |
| Negocios y sucursales | A | Alta, archivado, contexto activo | Vista consolidada entre sucursales |
| Equipo y permisos | A | Empleados, puestos, roles, invitaciones, asignaciones, 3 pasos | Turnos, PIN de caja, límites por rol, rol «Capturista» |
| Catálogo | B | Variantes, códigos, importación XLSX, surtido por sucursal | Corregir SKU, granel, equivalencias de empaque en venta, refactor |
| Proveedores | B | Sección propia, varios por producto | Comparación de costos, condiciones, cuentas por pagar |
| Compras | B | Recepción, IEPS, lote, caducidad, propuestas | Órdenes de compra, cuadre contra ticket, adjunto, validación de producto nuevo |
| Precios | B | Propuesta por sucursal, autorización por renglón, historial | Aprobación masiva, reglas de margen, autorización móvil |
| Inventario | C | Existencias, movimientos, ajuste, mínimo | Kardex visible, conteos, mermas, transferencias, alertas |
| Ventas | C | Solo en `origin/ventas`: captura, efectivo, ticket, historial, cancelación | Ver sección 5.3 |
| Caja y cortes | D | — | Todo |
| Clientes, crédito, envases | D | Solo un campo de configuración de días de envase | Todo |
| Buzón de tareas | D | — | Todo |
| Reportes | D | Métricas de plataforma y resumen de ventas | Todo lo del negocio |
| Sin internet | D | Solo borrador de venta en `localStorage` | Todo |
| Móvil | C | Responsive | PWA instalable, notificaciones, vistas de aprobación |
| Bitácora | D | Solo `auditoria_plataforma` y `ActividadUsuarioDiaria` | Bitácora por negocio |
| Facturación CFDI | D | Solo se captura si la compra es ticket o factura | Todo |
| Plataforma SaaS | C | Panel de lectura de negocios y métricas (ampliado en `main` y `ventas`) | Planes, cobro, soporte, baja de datos |

### 3.1 Verificaciones ejecutadas hoy (2026-10-30, rama `thrs`)

| Verificación | Resultado |
| --- | --- |
| `go test ./...` | 1 falla: `TestProductionAcceptsStrongOTPSecret`. Espera `OTPMaxAttempts = 5`, el código tiene default 10 (`config.go:94`). Producción fija 5 en compose. |
| Pruebas de integración | **Omitidas por falta de variable.** `internal/database` pasa en 0.005 s porque usa `t.Skip`. Un `go test` verde **no** ejercita PostgreSQL. |
| `go vet ./...` | Sin hallazgos |
| `ng test` | 275 de 281. Las 6 fallas son de `alta-inicial.spec.ts`: `window.matchMedia is not a function`. |
| `ng build` | Correcto, con advertencias de presupuesto de CSS |
| `govulncheck` | 13 vulnerabilidades alcanzables en 4 módulos (sección 6) |
| `npm audit --omit=dev` | 2 altas: `@angular/router` (afecta SSR, que no usamos) y `zod` |
| Rama `origin/ventas` | **No se ejecutó.** Su capability reporta 321 de 327 en Angular, `TestVentas` 11 de 11 y prueba manual pendiente. |

No hay CI (`.github/` no existe), ni pruebas E2E, ni pruebas en `application/compra`, `application/precio` ni sus handlers.

### 3.2 Estado de ramas

- `origin/ventas` desciende de `thrs` y `origin/main`: contiene 3 commits adicionales (`677b995`, `478e0c1`, `5cd61d7`). **No hay conflictos previsibles al integrar.**
- IDs ya usados: 35 y 36 (`main`), 37 (`ventas`). Siguiente libre: **38**.
- Este plan usa códigos `E1-01`, etc. El número `NN` se asigna al crear cada capability, para no chocar entre colaboradores.

---

## 4. Rastreo contra lo comprometido

### 4.1 Propuesta original ($60,000, 5 meses)

| Mes | Compromiso | Estado | Bloque |
| --- | --- | --- | --- |
| 1 | Importación de catálogo Odoo | Hecho (XLSX) | E1-13 migración real |
| 1 | Tickets con costo automático | Hecho | — |
| 1 | Detección de productos nuevos y cambios de precio | Hecho | — |
| 1 | Detección de errores de captura | Parcial | E1-09 |
| 1 | Aprobación masiva | **Falta**, es por renglón | E1-08 |
| 1 | Exportación a Excel para Odoo | **Falta** | E1-13 |
| 1 | Historial de compras y autorizaciones | Hecho | — |
| 2 | Catálogo interno | Hecho | — |
| 2 | Historial de costos y precios | Hecho | — |
| 2 | Trazabilidad de impuestos por producto | **Falta**, tabla `impuestos` sin uso real | E1-06 |
| 2 | Órdenes de compra | **Falta** | E1-10 |
| 3 | Existencias y movimientos | Parcial, sin kardex visible | E1-11 |
| 3 | Conteos físicos | **Falta** | E1-11 |
| 3 | Alertas de bajo inventario y sin movimiento | **Falta**, el mínimo se siembra en 0 | E1-11 |
| 3 | Envases retornables por cliente | **Falta** | E2-07 |
| 4 | Pantalla rápida y lector | Hecho en `ventas` | E1-05 |
| 4 | Métodos de pago | Parcial, solo efectivo en pantalla | E2-05 |
| 4 | Devoluciones | **Falta**, solo cancelación el mismo día | E2-06 |
| 4 | Apertura y cierre de caja por usuario | **Falta** | E2-04 |
| 5 | Panel ejecutivo | **Falta** | E2-09 |
| 5 | Reportes por periodo, proveedor, producto, categoría y usuario | **Falta** | E2-09 |
| 5 | Documentación y capacitación | Parcial: existe manual sin ventas | E2-11 |
| Todos | HTTPS y permisos | Hecho | — |
| Todos | **Respaldos automáticos** | **Falta**: la guía de despliegue lo declara como limitación aceptada | E1-02 |
| Todos | **Monitoreo continuo** | **Falta**: solo healthchecks | E1-02 |
| Todos | **Bitácora y recuperación documentada** | **Falta** | E1-04, E1-02 |

**Riesgo:** los tres últimos se ofrecieron «desde el mes 1». No guardar datos reales de la clienta hasta cerrar E1-02.

### 4.2 Observaciones de la Lic.

| Observación | Respuesta del plan | Bloque |
| --- | --- | --- |
| Manejo de dos sucursales | Ya hay precio y stock por sucursal. Faltan transferencias y vista consolidada. | E1-11, E1-12 |
| Trabajar sin internet | Arquitectura en sección 7 | E2-01 a E2-03 |
| Plataforma de administración | Sección 2 y centro de decisiones | E3-01 |
| Módulo de empleados con atributos | **Ya existe.** Mostrarlo en la propuesta. Agregar turnos y límites. | E2-04, E2-06 |
| Compra alimenta inventario y detecta cambio de precio | Ya existe. Falta cerrar el aviso y la aprobación móvil. | E1-07, E1-08 |
| El empleado captura, la Lic. valida | El flujo existe. Falta rol «Capturista», porque la migración de ventas le quita compras a Cajero. | E1-05, E1-08 |
| Producto nuevo validado por la Lic. | **Hoy no se cumple**: la compra lo crea directo (`compra_repository.go:149`). | E1-08 |
| Buzón de pendientes | No existe | E1-07 |
| Monitoreo y autorizaciones por celular | Responsive sí. PWA y notificaciones no. | E1-07, E1-14 |
| Diferenciarse de otros POS | Sección 9 | E3-01 a E3-03 |
| «Se tendría que probar, verificar y modificar» | Piloto y criterios de salida | E1-15, E2-11 |
| ¿$10 mil es costo único? ¿Cómo se paga? ¿5 meses? | Decisiones de negocio pendientes | Sección 12 y 13 |

### 4.3 Lo que revelan sus Excel (`thrs/compras_valle.xlsx`, `thrs/lista-productos.xlsx`)

Son requisitos reales que ninguna propuesta menciona:

| Hoja | Qué hace hoy la tienda | Traducción al producto |
| --- | --- | --- |
| `COMPROBACIÓN DE CAPTURA` | Compara ticket, cantidad e importe con y sin impuesto, con notas de descuento | Cuadre de captura contra ticket físico (E1-09) |
| `PRODUCTOS SACAR DEL SISTEMA` | Baja producto con motivo y registra si hubo cambio | Mermas y cambio con proveedor (E1-11) |
| `DATOS` | Motivos: abollado, dañado, caducado. Presentaciones de Paquete 2 a Paquete N | Catálogos iniciales sembrados |
| `LISTA DE COMPRAS` | Lista manual con existencia | Sugerencia de compra (E1-10) |
| `hoja 1` | Buscador de códigos y calculadora guiada de costo | Referencia para mejorar Registrar compra |

**Calidad de datos en hoja `BASE` (862 productos):**

- **215 con existencia negativa (25 %).** La importación actual rechaza stock negativo (capability 24), así que la migración directa falla.
- **128 sin código de barras.** Requiere búsqueda rápida por nombre, teclas favoritas y etiquetas propias.
- **12 sin costo.**
- Las categorías son en realidad marcas o distribuidores: Sabritas, Corona, Barcel, Coca Cola, Cigarros, Peñafiel, Bimbo, Licores.

**Giro real:** botanas, refrescos, cerveza, licores y cigarros. Esto confirma envases retornables, IEPS, venta por caja o cartón, y posible restricción de venta a menores como diferenciador de bajo costo.

---

## 5. Secciones a robustecer

### 5.1 Compras y precios

- **Cuadre contra ticket:** capturar total del documento y exigir explicación de diferencias.
- **Foto del ticket adjunta.** Hoy `modulo-compras.md` lo deja fuera y es evidencia de auditoría.
- **Producto nuevo con estado `PENDIENTE_VALIDACION`:** se recibe la mercancía, pero no se vende hasta validar nombre, precio e impuestos.
- **Aprobación masiva** con filtros por proveedor y porcentaje de cambio.
- **Reglas de precio por categoría:** margen objetivo y redondeo a precio comercial.
- **Alerta de margen:** la capability 37 registró 149 precios menores al costo y 5 en cero en una base local.
- **Corrección de compra confirmada:** hoy es inmutable y no existe reversa documentada. Se necesita nota de corrección con autorización.
- **IVA por renglón:** hoy es 16 % sobre subtotal + IEPS del documento completo.

### 5.2 Inventario

Hoy es la sección que más separa a Stockion de un sistema de administración real.

- Kardex por producto y sucursal con filtro por movimiento. Los datos ya existen en `movimientos_inventario`.
- Conteos físicos: por zona, a ciegas, con diferencia autorizada.
- Mermas con motivo, costo perdido y responsable.
- Transferencias entre sucursales: solicitud, envío, recepción y diferencia, con estado **en tránsito**.
- Alertas de mínimo y producto sin movimiento. `stock_minimo` se siembra en 0 y no hay pantalla de edición masiva.
- Lotes y caducidad: la venta no descuenta `lotes.cantidad_actual`, así que las caducidades capturadas no son accionables. Decidir FEFO o retirar la promesa.
- `stock_actual` es `float64` (`catalog_models.go:224`). Definir `numeric` para evitar acumulación de error y preparar granel.

### 5.3 Ventas (rama de Ilian)

**Buen trabajo y buena base:** centavos enteros, transacción con bloqueo ordenado, foto de precio y costo, prueba de concurrencia, ocultar costos desde la API. Se conserva.

**Cambios antes de integrar (E1-05), porque afectan el modelo de datos:**

| # | Hallazgo | Consecuencia | Cambio |
| --- | --- | --- | --- |
| 1 | `POST /ventas` sin clave de idempotencia | Reintento o doble envío puede duplicar venta, incluso con internet | `id_cliente` UUID con índice único |
| 2 | Folio consecutivo por negocio, asignado en servidor | Imposible sin conexión | Serie por terminal: `{sucursal}-{terminal}-000123` |
| 3 | Sin `terminal_id` ni `turno_id` | No hay corte ni trazabilidad por caja | Agregar columnas |
| 4 | Una sola fecha, `creado_en` | No distingue cuándo se vendió y cuándo llegó | `vendida_en` y `recibida_en` |
| 5 | Sin existencia, responde 409 | Una venta offline ya ocurrió y **no puede rechazarse** | Modo sincronización: acepta y crea incidencia |
| 6 | La capability sugiere `CHECK stock_actual >= 0` | Rompe la sincronización | **No agregarlo.** Controlar en la ruta en línea. |
| 7 | IVA fijo 16 % | Desglose incorrecto en ticket | Línea con foto de impuestos (E1-06) |
| 8 | Sin descuento ni forma de pago múltiple | Limita caja real | Modelar `venta_pagos` desde ya aunque la pantalla espere |
| 9 | Cantidades enteras | Sin granel ni venta por caja | Decidir antes de congelar el contrato |

**Pendientes propios, sin bloquear integración:** elemento «Ventas» del menú sin permiso, listado limitado a 1000, inicio de vendedor 01b, prueba manual con lector.

### 5.4 Cuenta, equipo y plataforma

- Cambiar contraseña con sesión activa, ver y cerrar sesiones. Hoy no hay ruta para ninguna.
- MFA obligatorio para propietario y administradores de plataforma.
- Límites por rol: descuento máximo, cancelación, ver costos. Un encargado no cancela sus propias ventas.
- Plantillas de rol: Encargado, Capturista, Cajero, Contador.
- Soporte con acceso temporal, autorizado por el cliente y auditado.

### 5.5 Deuda técnica

Atacar **solo al tocar la zona**, salvo E3-09.

- `product_repository.go` 2,555 líneas, `productos.component.ts` 1,560 líneas.
- Catálogo fuera de la regla de dominios de `AGENTS.md`: vive en la raíz de `application`, `infrastructure` y `domain`.
- Dinero en `float64` fuera de ventas: 77 usos de `float64` en dominio, con redondeos como `money()` y `money4()` en compras.
- **SKU roto:** `SKUInterno` sin etiqueta `column` (`catalog_models.go:182`) hace que GORM escriba en `sk_uinterno`, mientras el SQL consulta `sku_interno` (`product_repository.go:709`). La validación de duplicados no funciona.
- Estados de capabilities desactualizados en `PLAN_IMPLEMENTACION.md`; varias «Verificadas» sin aceptación final.
- `DOCUMENTACION.md` describe rutas de marcador y una carpeta `tienda/` antigua.

---

## 6. Seguridad y operación

**Lo que está bien y se conserva:** sesiones opacas con hash en DB, cookies `HttpOnly` y `SameSite=Strict`, CSRF validado contra la sesión, verificación de `Origin`, bcrypt, OTP con HMAC, token de invitación solo como hash, aislamiento por negocio y sucursal con 404, revocación de sesiones al recuperar contraseña, contenedores `read_only`, sin capacidades y con redes internas, CSP estricta en scripts. No encontré credenciales en el historial de Git (búsqueda por patrón `AKIA` y nombres de archivo).

**Esto es una revisión de código, no un pentest.**

| Sev. | Hallazgo | Evidencia | Acción | Bloque |
| --- | --- | --- | --- | --- |
| **Crítica** | Sin respaldos ni restauración probada | `GUIA_DESPLIEGUE_PRODUCCION.md` §13: «pérdida total e irreversible» | Respaldo cifrado externo, recuperación a un punto en el tiempo y simulacro | E1-02 |
| **Alta** | Migraciones sin versión, se ejecutan al arrancar, con `DROP COLUMN` y `AutoMigrate` | `database.go` | Herramienta versionada, respaldo previo, reversa | E1-03 |
| **Alta** | Sin bitácora de cambios por negocio | Solo `auditoria_plataforma` | Bitácora de solo inserción | E1-04 |
| **Alta** | 13 vulnerabilidades alcanzables. 7 en `excelize` 2.9.1, que procesa archivos subidos: un XLSX manipulado puede tumbar la API (límite 512 MB) | `govulncheck` | Actualizar `pgx` ≥ 5.9.2, `x/net` ≥ 0.60, `x/text` ≥ 0.41 y `excelize` a una versión que corrija las 7 (cuatro se corrigen después de 2.11.0; confirmar con `govulncheck`). Límites de filas y celdas. | E1-01 |
| **Alta** | Login sin límite por IP. Cinco intentos bloquean 15 minutos a la víctima, sin importar el origen | `main.go:126`, `auth_service.go:33-38` | Límite por IP y correo; bloqueo progresivo; captcha tras fallos | E1-03b |
| **Alta** | `verificar-correo`, `reenviar-verificacion`, `solicitar-recuperacion` y `restablecer-contrasena` sin límite en ruta | `main.go:127-130`. Tienen límites internos por desafío, no por IP para validar | Extender limitador. Pasarlo a almacenamiento compartido si hay réplicas | E1-03b |
| **Media** | Token de invitación en la URL, registrado por el logger de Gin y Caddy | `main.go:116`, `:178` | Redactar `/invitaciones/*` en logs o enviar por `POST` | E1-03b |
| **Media** | Enumeración: cuenta bloqueada o sin verificar responde distinto, antes de validar contraseña; correo inexistente no ejecuta bcrypt | `auth_service.go:26-31` | Mensaje y tiempo uniformes | E1-03b |
| **Media** | La recuperación no aplica `ContrasenaSegura` | Único uso en `verification_service.go:95` | Aplicar en todos los caminos. Subir mínimo a 10–12 y revisar contraseñas filtradas | E1-03b |
| **Media** | Gin en modo debug en producción y sin `ReadTimeout`, `WriteTimeout` ni cierre ordenado | `main.go:116`, `:260` | `http.Server` con límites, `Shutdown`, `GIN_MODE=release` | E1-03b |
| **Media** | Sin configuración de pool de DB | `database.Open` | `SetMaxOpenConns` y afines, sobre todo con 512 MB | E1-03b |
| **Media** | Sin HSTS visible | `deploy/Caddyfile` | Confirmar en Cloudflare o agregar | E1-03b |
| **Media** | Sin observabilidad: 8 `log.Print`, sin logs estructurados, errores centralizados ni métricas | Búsqueda en `backend/` y `frontend/package.json` | `slog`, ID de solicitud, Sentry o similar, alertas | E1-02 |
| **Media** | Aislamiento solo en aplicación | Sin RLS en PostgreSQL | Suite automática de aislamiento; evaluar RLS para SaaS | E3-07 |
| **Media** | Un solo servidor, DB en contenedor con 512 MB, sin staging | `compose.production.yaml` | Staging en E1; separar DB y definir RPO/RTO en E3 | E1-02, E3-07 |
| **Baja** | Imagen de login servida desde Unsplash | `frontend/src/environments/` | Servir local; también rompe modo offline | E2-01 |
| **Baja** | Consulta de códigos por `UPCitemdb` en URL `trial` | `.env.production.example` | Revisar términos comerciales antes de vender | E3-05 |
| **Higiene** | `credenciales-aws.txt` y un CSV SMTP en texto claro en carpetas locales | Ignorados por Git, no leí su contenido | Pasar a gestor de secretos. Rotar si alguna vez se compartieron | E1-01 |

**Metas operativas (por confirmar):**

| Meta | Etapa 1 | Etapa 3 |
| --- | --- | --- |
| RPO, pérdida máxima de datos | 24 h, idealmente 1 h | ≤ 5 min |
| RTO, tiempo de recuperación | 8 h | ≤ 1 h |
| Simulacro de restauración | Una vez antes de datos reales | Trimestral |
| Disponibilidad | Monitoreada | 99.5 %, con página de estado |

---

## 7. Arquitectura sin internet

### 7.1 Qué significa «funciona 1000 %»

Hay que convertir la frase en algo comprobable:

> **Ninguna venta se detiene ni se pierde por falta de internet o caída del servidor, y al reconectar todo se concilia sin duplicados ni intervención técnica.**

**Límites que debemos decir con franqueza a la Lic.:**

- Un equipo que nunca se conectó no puede iniciar. Requiere preparación previa.
- Dos cajas sin internet no se ven entre sí: pueden vender la misma última pieza.
- Si se destruye un equipo con ventas sin sincronizar, esas ventas se pierden, salvo que exista respaldo local (decisión D-3).
- Correos, cobros en línea, autorizaciones remotas y timbrado CFDI requieren internet: se encolan o esperan.

### 7.2 Qué funciona sin conexión

| Función | Sin internet | Regla |
| --- | --- | --- |
| Iniciar turno en equipo registrado | Sí | PIN local |
| Buscar producto y consultar precio | Sí | Última copia sincronizada, con antigüedad visible |
| Vender en efectivo | Sí | Folio por terminal |
| Vender con terminal bancaria independiente | Sí | Se captura referencia manual |
| Imprimir y reimprimir ticket | Sí | Según hardware (E2-08) |
| Cancelar venta del turno | Sí | Permiso de la copia local; revisión al sincronizar |
| Abrir y cerrar caja, movimientos | Sí | El corte definitivo se confirma al sincronizar |
| Devolución | Limitada | Ventas presentes en el equipo; lo demás espera |
| Capturar compra | Sí, como borrador | El stock entra al sincronizar y validar |
| Conteo físico | Sí | Se envía como diferencia con la base de la copia |
| Solicitar producto nuevo | Sí | Queda pendiente de validación |
| Autorizar precios | No | Exige estado actual del servidor |
| Reportes | Parcial | Solo del turno y la terminal |
| Ver otra sucursal | No | Última copia, marcada como desactualizada |
| Primer acceso, OTP, invitación, CFDI | No | — |

### 7.3 Diseño técnico

```text
Angular (UI)
   │
Capa de datos local ── IndexedDB (Dexie)
   │                     ├─ copia: productos, códigos, precios, existencias,
   │                     │         usuarios de la terminal, permisos, config
   │                     ├─ outbox: operaciones pendientes (no se borra sin ACK)
   │                     └─ turno: ventas y movimientos locales
   │
Motor de sincronización (Web Locks: un solo emisor entre pestañas)
   │        ▲
   ▼        │
POST /sync/lote         GET /sync/cambios?cursor=
   │        ▲
Go: sync_operaciones (único por terminal + op_id) ─► servicios de dominio
```

**Decisiones clave:**

1. **Local primero en mostrador.** La venta siempre se escribe en local y luego se sincroniza, aun con internet. Un solo camino de código evita el error clásico de que el modo offline sea el menos probado.
2. **ID generado por cliente** (UUID v7) en toda operación. Reintentar es seguro: el servidor responde `DUPLICADA`.
3. **Inventario por movimientos, nunca por valor absoluto.** Se envía `-2`, no `stock = 8`. Así las ventas de varias cajas se fusionan sin conflicto. `stock_actual` queda como proyección.
4. **El equipo no es confiable.** El servidor recalcula totales, valida pertenencia y compara el precio contra `precios_sucursal` e historial vigentes en esa fecha. Una diferencia no rechaza la venta, abre incidencia.
5. **Una venta ocurrida no se rechaza.** Stock negativo, precio viejo o permiso revocado generan incidencia en el buzón.
6. **Versionado de contrato.** Cada operación lleva `schema_version`. El servidor debe aceptar versiones anteriores durante la ventana máxima sin conexión. Un equipo puede volver tras días con cola vieja.

**Política de conflictos:**

| Dato | Autoridad | Resolución |
| --- | --- | --- |
| Venta, movimiento, gasto de caja | Terminal | Se agrega; idempotente |
| Existencia | Calculada | Suma de movimientos; negativo genera incidencia |
| Precio y costo | Servidor | La venta conserva foto; aviso si difiere |
| Catálogo | Servidor | La copia local se reemplaza |
| Permisos | Servidor | Se aplican al sincronizar; lo previo se revisa |
| Conteo | Servidor autoriza | Diferencia contra base; incidencia si hubo ventas en medio |
| Compra | Servidor | Documento repetido genera incidencia |
| Cierre de caja | Terminal propone | Servidor confirma al tener todas las ventas del turno |

**Identidad sin conexión:**

- **Terminal registrada** por el propietario: entidad `terminales` ligada a negocio y sucursal, revocable, con código para folio.
- **Credencial de dispositivo** independiente de la sesión de usuario. Si la cookie de 24 h vence offline, la cola debe poder sincronizar al volver. Evaluar llave WebCrypto no extraíble.
- **PIN de caja** por empleado autorizado en esa terminal: hash lento (PBKDF2 o Argon2id en WASM), intentos limitados, bloqueo local.
- **Ventana máxima sin conexión** configurable, sugerida en 72 h.
- **Datos mínimos:** los costos no se descargan a usuarios sin `ventas.ver_costos`. IndexedDB no está cifrado por defecto. Revocar terminal borra datos locales cuando se reconecta.

**Sincronización de bajada:**

- `GET /sync/snapshot` inicial paginado y comprimido (800–5,000 productos).
- `GET /sync/cambios` con cursor incremental y tombstones.
- **Cuidado:** una secuencia simple como cursor puede saltarse filas de transacciones abiertas. Usar bitácora de cambios escrita en la misma transacción y leer solo hasta un límite seguro (`pg_snapshot_xmin`).

**Cambios concretos ya identificados en el repo:**

- Agregar `@angular/service-worker`, manifest e iconos. Hoy no existe ninguno.
- `frontend/nginx.production.conf`: la regla `\.(js|css…)$` marcaría `ngsw-worker.js` como `immutable` durante un año, y **los clientes nunca recibirían actualizaciones**. Agregar excepción con `no-cache`, también para `ngsw.json`.
- Revisar caché de Cloudflare para esos mismos archivos.
- Servir localmente fuentes e imágenes externas.
- Detectar conectividad con latido a `/api/v1/health`. `navigator.onLine` no es confiable.
- `navigator.storage.persist()` y monitoreo de cuota.
- Una actualización del service worker no debe aplicarse con venta en curso. Requiere migraciones de IndexedDB.
- Contador permanente de «N sin sincronizar», y bloquear cierre de sesión con pendientes.

**Plataforma recomendada para cajas:** Chrome o Edge en Windows o Android, con PWA instalada. No usar iOS como caja: Safari tiene límites de almacenamiento y sin Background Sync. iPhone sí sirve para aprobar.

### 7.4 Criterios de aceptación

Automatizados con Playwright y un entorno que corte red, servidor y base de datos.

| # | Escenario | Resultado exigido |
| --- | --- | --- |
| 1 | 2 cajas, 8 horas sin red, 500 tickets | Al volver: 500 ventas, 0 duplicados, existencias exactas |
| 2 | Corte de red durante el envío, antes de recibir ACK | Reintento sin duplicar |
| 3 | Servidor caído a media venta | Venta intacta y cobrable |
| 4 | Cerrar navegador o apagar equipo con cola | Cola intacta al volver |
| 5 | Dos pestañas abiertas | Un solo emisor, folios sin repetir |
| 6 | Dos cajas venden la última pieza | Ambas aceptadas, incidencia creada |
| 7 | Precio cambia mientras una caja está offline | Venta con foto vieja, aviso |
| 8 | Cajero revocado mientras está offline | Ventas aceptadas y marcadas |
| 9 | Reloj del equipo adelantado 3 días | Fecha señalada, sin romper reportes |
| 10 | Despliegue nuevo con equipo offline y cola vieja | Cola procesada con su versión |
| 11 | Sesión vencida al reconectar | Cola sincroniza sin intervención |
| 12 | Catálogo de 5,000 productos | Búsqueda local < 100 ms en equipo modesto |
| 13 | Prueba real en tienda: desconectar el módem un día de operación | Sin incidencias no explicadas |

**Compuerta:** Odoo no se apaga hasta pasar los 13.

---

## 8. Módulos nuevos

### Buzón de decisiones (E1-07)

Bandeja única para el propietario, y la pieza más visible para la Lic. Tipos iniciales: cambio de costo o precio, producto nuevo, diferencia de cuadre, diferencia de conteo, transferencia con faltante, cancelación, incidencia de sincronización, corte con diferencia.

Contrato: tabla `tareas_negocio` con tipo, prioridad, sucursal, referencia, responsable, estado y vencimiento. Los módulos publican mediante interfaz de aplicación, sin importar repositories ajenos (regla de `AGENTS.md`).

### Caja y turnos (E2-04)

Apertura con fondo, entradas y salidas con motivo, corte a ciegas (el cajero cuenta sin ver lo esperado), diferencia con autorización, cierre por forma de pago, historial. Un turno por terminal.

### Clientes y envases retornables (E2-07)

Cliente opcional en venta. Envase como producto ligado: entregado, recibido, pendiente y cobrado por cliente, con los días de devolución que ya existen en `ConfiguracionNegocio`. Saldo de envases a nivel proveedor (Corona, Coca Cola): **validar con la Lic.**, porque también los devuelve al distribuidor.

### Reportes y tablero (E2-09)

Los cinco cortes prometidos. Cada uno con exportación a Excel y sucursal o consolidado. Base: ventas, costo foto y utilidad. Consultas sobre vistas o tablas resumen para no afectar la operación.

### Centro de decisiones (E3-01)

Ver sección 9.

### CFDI 4.0 (E3-04)

Por PAC, sin construir timbrado propio. Factura por ticket, autofactura con QR del ticket, factura global de público en general. **Depende de E1-06**: claves SAT por producto y desglose real de IVA e IEPS. Validar con contador; la investigación hecha solo llegó a fuentes secundarias.

### Otros

- Gastos y flujo de caja (E3-06): sin esto no hay utilidad real.
- Cuentas por pagar a proveedores (E3-06): la compra actual fuerza total igual a pagado.
- Crédito a clientes o «fiado» (E3-06): común en el giro, validar demanda.
- Promociones (E3-06): combos y descuentos por fecha.

---

## 9. Diferenciadores

**Advertencia:** son hipótesis. La búsqueda de hoy confirmó que hay jugadores establecidos y baratos: Eleventa en México desde 2001, SICAR orientado a abarrotes con pago único desde MX$4,940 según Capterra, Loyverse con 4.8 en Capterra, y ofertas de POS en la nube desde $390 al mes en Mercado Libre. **No pude verificar el modo offline ni los precios vigentes.** Antes de la presentación, hacer tabla comparativa probando los productos (E3-05). Competir por precio o por lista de funciones sería perder.

| # | Diferenciador | Por qué importa | Estado |
| --- | --- | --- | --- |
| 1 | **Del ticket a la decisión**: compra, cambio de costo, efecto en margen, autorización en el celular y precio actualizado | Es el dolor original de la Lic. | 70 % hecho, falta móvil |
| 2 | **Buzón de decisiones** | Administración por excepción | Nuevo |
| 3 | **Protector de margen**: alerta de venta bajo costo y de proveedor que subió | Dinero visible de inmediato | Datos ya disponibles |
| 4 | **Capital atorado y sugerencia de compra**: qué no se mueve, qué comprar, a qué proveedor | Decisión semanal real | Requiere ventas y compras |
| 5 | **Control antifraude**: cancelaciones, diferencias de caja y conteos por empleado | El dueño ausente es el cliente | Requiere caja y bitácora |
| 6 | **Comparador de proveedores** con historial de costo | Compras con datos | Datos ya disponibles |
| 7 | **Resumen diario por WhatsApp**: ventas, margen, faltantes y pendientes | El dueño no abre sistemas | Costo por mensaje |
| 8 | **Captura de ticket con foto (IA)**: lee el ticket y prellena la compra | Elimina el trabajo más tedioso | Medir precisión con tickets reales |
| 9 | **Envases retornables** | Nicho poco atendido en cerveza y refrescos | Verificar competencia |
| 10 | **Sin internet real** | Requisito de la Lic. | Etapa 2 |

**Recomendación:** vender 1, 2, 3 y 5 como núcleo. El 8 es el mejor gancho de demostración, pero debe probarse antes de prometerlo.

**Mercado inicial sugerido:** tiendas con 2–5 sucursales, dueño no siempre presente, alta rotación, cerveza y refrescos. El dolor de control crece con la segunda sucursal, y un POS básico ya no alcanza.

---

## 10. Etapas y bloques

Tamaño: **S** ≤ 3 días · **M** ≤ 2 semanas · **L** ≤ 4 semanas. Cada bloque es una o más capabilities.

### Etapa 1 — Control confiable

**Objetivo:** datos reales seguros y la Lic. decidiendo desde el celular. El POS corre solo como piloto.

| Código | Bloque | Tamaño | Depende de |
| --- | --- | --- | --- |
| E1-01 | Estabilización: suites verdes, integración en CI, GitHub Actions, dependencias, `govulncheck`, higiene de secretos | M | — |
| E1-02 | Respaldos, restauración probada, monitoreo, errores centralizados y staging | M | — |
| E1-03 | Migraciones versionadas | M | E1-01 |
| E1-03b | Endurecimiento de autenticación y servidor (sección 6) | M | E1-01 |
| E1-04 | Bitácora de auditoría por negocio | M | E1-03 |
| E1-05 | **Ajuste e integración de ventas** (sección 5.3), rol Capturista | M | E1-03 |
| E1-06 | Impuestos por producto; dinero exacto en compras y precios; corrección de SKU | L | E1-03 |
| E1-07 | Buzón de decisiones | M | E1-04 |
| E1-08 | Producto nuevo con validación, aprobación masiva, reglas de margen | M | E1-07 |
| E1-09 | Cuadre contra ticket, foto adjunta, corrección de compra | M | — |
| E1-10 | Órdenes de compra y sugerencia de compra | M | E1-05 |
| E1-11 | Kardex, mermas, conteos, alertas, transferencias | L | E1-07 |
| E1-12 | Inicio con pendientes y vista consolidada | S–M | E1-07 |
| E1-13 | Migración Odoo real, limpieza de negativos, exportación a Odoo | M | E1-06, E1-11 |
| E1-14 | PWA instalable, notificaciones push y aprobación móvil | M | E1-07 |
| E1-15 | Piloto con la Lic., lista de observaciones | S + continuo | Todos |

**Orden recomendado:**

1. **Semana 1–2:** E1-01, E1-02. Ninguna otra entrega va a datos reales antes.
2. **Semana 2–4:** E1-03, E1-03b, E1-05 e iniciar E1-06.
3. **Semana 4–8:** E1-04, E1-07, E1-08, E1-09, E1-14.
4. **Semana 6–10:** E1-10, E1-11, E1-12, E1-13, E1-15.

**Criterios de salida:**

- [ ] Restauración probada, documentada y cronometrada.
- [ ] CI obligatorio con integración sobre PostgreSQL; cero vulnerabilidades altas alcanzables.
- [ ] La Lic. autoriza precios y productos nuevos desde su celular, con aviso.
- [ ] Catálogo real migrado con conteo inicial en ambas sucursales.
- [ ] Una transferencia real entre sucursales cerrada.
- [ ] Cuatro semanas de compras reales sin pérdida de datos.

### Etapa 2 — Mostrador sin fallas

**Objetivo:** apagar Odoo con seguridad.

| Código | Bloque | Tamaño | Depende de |
| --- | --- | --- | --- |
| E2-01 | Núcleo local: service worker, IndexedDB, copia y sincronización de bajada | L | E1-05, E1-14 |
| E2-02 | Terminales, credencial de dispositivo, PIN | L | E2-01 |
| E2-03 | Venta local primero: outbox, `/sync/lote`, incidencias, pruebas de caos | L | E2-01, E2-02 |
| E2-04 | Caja y turnos | L | E2-03 |
| E2-05 | Tarjeta, transferencia y pagos mixtos | S–M | E2-03 |
| E2-06 | Devoluciones, descuentos con límite, mayoreo, venta por caja | M | E2-03 |
| E2-07 | Clientes y envases retornables | L | E2-03 |
| E2-08 | Impresora térmica, cajón, escáner, etiquetas | M | E2-03 |
| E2-09 | Reportes de 5 dimensiones y tablero | L | E2-04 |
| E2-10 | Compras y conteos sin conexión; inicio de vendedor | M | E2-03 |
| E2-11 | Piloto en una caja, luego todas; corte de Odoo; capacitación; manual | M + continuo | Todos |

**Criterios de salida:**

- [ ] Los 13 escenarios de la sección 7.4 en verde.
- [ ] Una semana con ambas sucursales sin Odoo, con cortes de caja cuadrados.
- [ ] Captura de 10 productos solo con lector, en menos de 30 segundos.
- [ ] Respuesta a escaneo < 100 ms.
- [ ] Aceptación escrita de la Lic.

### Etapa 3 — Plataforma vendible

**Objetivo:** un segundo negocio se da de alta solo y paga.

| Código | Bloque | Tamaño | Depende de |
| --- | --- | --- | --- |
| E3-01 | Centro de decisiones: margen, capital atorado, sugerencias, antifraude, comparador | L | E2-09 |
| E3-02 | Resumen diario y alertas por WhatsApp | S–M | E3-01 |
| E3-03 | Captura de ticket con foto: prueba de precisión, luego integración | M–L | E1-09 |
| E3-04 | CFDI 4.0 por PAC | L | E1-06 |
| E3-05 | Comparativo de competencia, términos de terceros, precios y planes | S | — |
| E3-06 | Gastos, flujo de caja, cuentas por pagar, crédito, promociones | L | E2-04 |
| E3-07 | Infraestructura: DB administrada, separar servicios, pruebas de carga, aislamiento, RLS, SLO | L | E1-02 |
| E3-08 | Suscripciones, límites por plan, cobro recurrente, suspensión y gracia | L | E3-05 |
| E3-09 | Refactor de catálogo a dominio propio y componentes grandes | L | E1-06 |
| E3-10 | Alta autoservicio: giros, importador guiado, datos demo | M | E3-08 |
| E3-11 | Legal: aviso de privacidad, términos, contrato de encargado, exportación y baja de datos | M | Abogado |
| E3-12 | Soporte: ayuda, tickets, estado del servicio, acceso temporal auditado | M | E1-04 |
| E3-13 | Pentest externo y remediación | M | Todo lo anterior |

**Criterios de salida:**

- [ ] Negocio nuevo operando en menos de 1 hora sin nuestra ayuda.
- [ ] Pentest sin críticas ni altas abiertas.
- [ ] Prueba de carga con 50 negocios simulados.
- [ ] Prueba documentada de aislamiento entre negocios.
- [ ] Cobro automático real funcionando.
- [ ] Legal revisado por abogado.
- [ ] Un cliente externo pagando.

---

## 11. Definición de terminado

Una capability no se cierra sin:

- [ ] Plan aprobado, con alcance, contratos, pruebas y criterios.
- [ ] Unitarias e integración **ejecutadas**, no omitidas.
- [ ] Autorización probada: sin permiso, otro negocio, otra sucursal.
- [ ] Flujo principal en E2E.
- [ ] Migración con reversa o plan de recuperación.
- [ ] Bitácora en acciones sensibles.
- [ ] Probado a 360 px y con teclado.
- [ ] Definición de comportamiento sin internet, aunque sea «no disponible».
- [ ] Base temática actualizada.
- [ ] Prueba manual y aceptación registradas.

**Regla de trabajo:** dos colaboradores no modifican la misma zona sin avisar. Reservar el ID de capability al crearla.

---

## 12. Riesgos y decisiones

### Riesgos

| Riesgo | Prob. | Impacto | Mitigación |
| --- | --- | --- | --- |
| Pérdida de datos antes de tener respaldos | Media | Crítico | E1-02 primero; bloquear datos reales |
| Sincronización con errores en dinero o stock | Alta | Crítico | Un solo camino de código, pruebas de caos, compuerta |
| Alcance mayor que precio y plazo acordados | **Alta** | Alto | Renegociar antes de firmar (sección 13) |
| Capacidad real de dos personas | Alta | Alto | Definir dedicación (D-7) |
| Cambio de fondo en Ventas con datos ya guardados | Media | Alto | E1-05 antes de integrar |
| Errores fiscales en desglose | Media | Alto | E1-06 con contador |
| Cliente único moldea el producto a su medida | Media | Medio | Configurar por negocio; validar con 2–3 tiendas más |
| Dependencia de un solo servidor | Media | Alto | Offline ayuda al mostrador; E3-07 |
| Datos iniciales sucios | Alta | Medio | Conteo inicial asistido (E1-13) |
| Costos de terceros: WhatsApp, IA, PAC | Media | Medio | Modelarlos en los planes |

### Decisiones pendientes

| ID | Decisión | Recomendación | Quién |
| --- | --- | --- | --- |
| D-1 | ¿Aceptar 3 etapas y orden? | Sí | thrs, Ilian |
| D-2 | ¿Ajustar ventas antes de integrar? | **Sí** | thrs, Ilian |
| D-3 | ¿PWA o envoltura de escritorio (Tauri)? | PWA primero. Tauri solo si impresión silenciosa, cajón o durabilidad en disco lo exigen. Prueba técnica al iniciar E2-01 | Equipo |
| D-4 | Ventana máxima sin conexión | 72 h | Lic. + equipo |
| D-5 | ¿Granel y venta por caja desde ya? | Preguntar a la Lic. antes de E1-05: cambia el contrato | Lic. |
| D-6 | Herramienta de migraciones | `goose` | Equipo |
| D-7 | Dedicación semanal real de cada quien | Necesaria para calendario | thrs, Ilian |
| D-8 | ¿Cuántas cajas por sucursal y qué hardware? | Preguntar a la Lic. | Lic. |
| D-9 | ¿Sin red entre cajas, o servidor local en tienda? | Sin servidor local; el riesgo se atiende con incidencias | Equipo |
| D-10 | ¿La Lic. necesita facturar desde el sistema y cuándo? | Preguntar. Podría adelantar E3-04 | Lic. |
| D-11 | Lotes y caducidad: ¿FEFO real? | Alertas sí; descuento por lote solo si lo pide | Lic. |
| D-12 | Nombre y marca comercial | Stockion, con búsqueda de marca ante IMPI antes de vender | thrs, Ilian |
| D-13 | Figura legal: quién firma, factura y responde | Definir antes del contrato. Carpeta `contrato/` aún no contiene contrato | thrs, Ilian |

### Preguntas para la Lic.

1. ¿Cuántas cajas y qué equipo hay por sucursal? ¿Impresora y cajón ya existen?
2. ¿Cuánto tiempo suele fallar el internet y con qué frecuencia?
3. ¿Vende a granel, por caja, a crédito o a mayoreo?
4. ¿Los envases se controlan con cliente, con proveedor o ambos?
5. ¿Necesita facturar? ¿Cuántas al mes?
6. ¿Qué decisiones toma cada semana y qué número revisa primero?
7. ¿Qué autorizaciones puede delegar a un encargado?
8. ¿Conocería a otros dueños dispuestos a probar el sistema?

---

## 13. Insumos para el documento de la clienta

Lo trabajamos después. Queda anotado para no perderlo.

**Formato que propuso Ilian (`thrs/propuestaClinicaDental.pdf`):** lista concreta de funciones por sección, hospedaje con precio anual, plazo, inversión separada por módulo, costos variables de terceros, IVA, garantía y tarifa por hora. Es más claro y más protector que la presentación actual, porque **dice exactamente qué se entrega**.

**Lo que debe responder y hoy no responde:**

| Tema | Falta definir |
| --- | --- |
| $10,000 de infraestructura | La propuesta cubre «durante el proyecto». ¿Cuánto cuesta después? |
| Calendario de pagos | La Lic. preguntó dos veces. Recomiendo por entrega aceptada, no por tiempo. |
| Plazo | Por etapas, con criterios de aceptación |
| Precio | **$50,000 fue calculado para un alcance mucho menor.** Decidir si se mantiene como precio de cliente fundador a cambio de caso de éxito y referencias, o si se reabre. Decidirlo con D-7 en la mano. |
| Fuera de alcance | Lista explícita |
| Cambios | Proceso y tarifa por hora |
| Garantía | Plazo y qué cubre |
| Terceros | WhatsApp, PAC, IA, SMS, dominio |
| IVA | Decir si está incluido |
| Datos | Son de la Lic.; exportación al terminar; confidencialidad |
| Propiedad intelectual | Ya propuesta: el equipo la conserva. Formalizar en contrato. |
| Soporte posterior | Horario, tiempo de respuesta, costo |
| Sin internet | Usar la definición y límites de 7.1. **No prometer «1000 %» sin esas condiciones.** |

**Mensaje clave para ella:** *«Ya tomamos sus observaciones: esto es una plataforma de administración, no otro punto de venta»*, mostrando la tabla 4.2 con lo ya construido, especialmente Equipo, Compras y Autorización de precios.

---

## 14. Siguientes pasos

1. thrs e Ilian revisan este documento y resuelven D-1, D-2 y D-7.
2. Hacer las preguntas de la sección 12 a la Lic., sobre todo granel y cajas, antes de congelar ventas.
3. Crear las capabilities de E1-01 y E1-02 y someterlas a aprobación.
4. Crear la capability de E1-05 con Ilian.
5. Después, redactar el documento para la clienta.

---

## Fuentes externas

Consultadas el 2026-10-30. Son referencias de orientación, no asesoría legal ni fiscal.

**Competencia**
- [Eleventa, texto de prensa](https://descargas.eleventa.com/mediakit/texto-prensa-eleventa.pdf)
- [Eleventa en Capterra](https://www.capterra.es/software/197855/eleventa)
- [SICAR en Capterra](https://www.capterra.es/software/201314/sicar-punto-de-venta)
- [SICAR, ficha de abarrotes](https://fichas.sicar.mx/PDFs/Ficha_Abarrotes.pdf)
- [Loyverse en Capterra](https://www.capterra.com/p/150632/Loyverse-POS/reviews/?page=4)
- [Listados de POS en Mercado Libre](https://listado.mercadolibre.com.mx/programa-puntos-de-venta-para-ferreteria)

**Protección de datos** (nueva LFPDPPP, publicada el 20 de marzo de 2025; verificar texto vigente en el DOF)
- [Hogan Lovells](https://www.hoganlovells.com/es/publications/mexicos-new-federal-data-protection-law-what-it-means-for-companies)
- [Garrigues](https://www.garrigues.com/es_ES/noticia/mexico-nueva-ley-federal-proteccion-datos-personales-posesion-particulares-introduce)
- [EY Law Flash](https://www.ey.com/content/dam/ey-unified-site/ey-com/es-mx/technical/tax/documents/ey-law-flash-nueva-lfpdppp.pdf)

**Fiscal** (validar con contador; no se confirmaron tasas por tipo de producto)
- [IDC: CFDI 4.0 con público en general](https://idconline.mx/fiscal-contable/2022/06/29/cfdi-40-por-operaciones-con-publico-en-general)
- [Guía de llenado CFDI global 4.0](https://bhrmx.com/wp-content/uploads/2022/01/GuiallenadoCFDIglobal4.0.pdf)
- [Facturama: factura con IEPS](https://facturama.mx/blog/que-significa/factura-con-ieps/)
