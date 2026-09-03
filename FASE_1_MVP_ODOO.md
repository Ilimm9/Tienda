# Fase 1 — MVP operativo con Odoo

> Fuente de verdad para alcance, requerimientos, seguridad, ejecución y aceptación.

| Campo             | Valor                                                                                |
| ----------------- | ------------------------------------------------------------------------------------ |
| Proyecto          | Plataforma de Administración de Negocios                                             |
| Entrega           | Fase 1 — Base operativa y recepción prioritaria                                      |
| Duración          | 6 semanas                                                                            |
| Equipo            | Socio A y Socio B, ambos full-stack                                                  |
| Modelo de trabajo | Features verticales con revisión cruzada                                             |
| Integración Odoo  | Intercambio de archivos XLSX, no API directa                                         |
| Infraestructura   | Aplicación en Docker, PostgreSQL administrado y almacenamiento S3 compatible privado |
| Estado            | Planeación lista para desglose                                                       |

## 1. Objetivo operativo

Entregar una primera versión utilizable que permita cargar el catálogo vigente de Odoo, crear la estructura mínima del negocio, registrar recepciones reales de mercancía y detectar productos nuevos o variaciones de costo y presentación.

La plataforma sugerirá precios, pero nunca modificará el precio vigente sin autorización. Una recepción confirmada generará una entrada de inventario y actualizará las existencias de la sucursal exactamente una vez.

### 1.1 Flujo principal

```text
Cuenta segura
  → Negocio y sucursales
  → Importar XLSX de Odoo
  → Validar y aplicar catálogo
  → Registrar proveedor
  → Capturar recepción
  → Vincular o reportar productos
  → Analizar costo y presentación
  → Sugerir precio
  → Aprobar o rechazar
  → Confirmar entrada de inventario
  → Exportar cambios para Odoo
  → Auditar todo el flujo
```

### 1.2 Definición de terminado de la Fase 1

La fase estará terminada cuando, en staging y con un XLSX real:

- Una propietaria pueda verificar su cuenta, activar MFA, crear un negocio y registrar una o más sucursales.
- Un administrador pueda invitar empleados y asignar los roles definidos en este documento.
- Un administrador pueda cargar un XLSX de Odoo, revisar errores por fila y aplicar únicamente un lote válido.
- Repetir una importación o una solicitud no duplique productos, recepciones, precios ni movimientos.
- Un capturista pueda registrar proveedor, documento y partidas de una recepción sin poder aprobar precios.
- Cada partida quede clasificada como nueva, sin vincular, sin cambio, costo subió, costo bajó, cambio de presentación, incompleta o error.
- Un administrador pueda aprobar o rechazar la sugerencia y consultar el historial anterior y nuevo.
- Una recepción confirmada aumente existencias en la sucursal y cree movimientos inmutables de kardex.
- Se pueda descargar un XLSX o CSV con los cambios aprobados para su aplicación en Odoo.
- Las operaciones sensibles sean auditables por usuario, negocio, sucursal, fecha y `request_id`.
- La aplicación opere por HTTPS en contenedores endurecidos y se haya comprobado una restauración de PostgreSQL.

## 2. Alcance

### 2.1 Incluido

- Registro, verificación de correo, login, logout, recuperación de contraseña y revocación de sesiones.
- MFA TOTP obligatorio para propietario y administrador antes de operar en producción.
- Negocios, sucursales, membresías, empleados, invitaciones y RBAC mínimo.
- Catálogo local: marcas, categorías, unidades, productos, variantes, presentaciones, códigos, impuestos, costos y precios.
- Importación XLSX de Odoo con validación previa, errores descargables, aplicación controlada y mapeos externos.
- Proveedores y relación producto-proveedor.
- Recepción de mercancía, impuestos por partida e incidencias básicas.
- Comparación de costos, cambio de presentación, sugerencia de precio y autorización.
- Exportación por archivo de cambios aprobados hacia Odoo.
- Inventario mínimo: entradas por recepción, existencias por sucursal y consulta de kardex.
- Auditoría, idempotencia, logs estructurados, monitoreo, backups y despliegue Docker.

### 2.2 Fuera de alcance

- Ventas, tickets, cajas, pagos, clientes, apartados y envases retornables.
- Salidas de inventario, conteos físicos, ajustes manuales, mermas y transferencias.
- Integración directa con la API de Odoo.
- Modo offline completo, sincronización entre servidores locales y nube.
- Open Food Facts, catálogo global, imágenes de productos y reportes avanzados.
- Aplicación automática de precios sin autorización.
- Aplicación móvil nativa.

## 3. Usuarios, roles y separación de responsabilidades

| Rol                         | Capacidades mínimas                                                                                                                                               | Restricciones                                                                                         |
| --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Propietario / Administrador | Administrar negocio, sucursales y equipo; importar catálogo; administrar proveedores; ver y confirmar recepciones; aprobar precios; exportar; consultar auditoría | MFA obligatorio; todas las acciones sensibles quedan auditadas                                        |
| Capturista de recepción     | Consultar catálogo/proveedores; crear y editar borradores; enviar recepciones a revisión                                                                          | No puede importar/aplicar catálogo, aprobar precios, administrar permisos ni eliminar históricos      |
| Revisor de precios          | Consultar recepciones y análisis; revisar sugerencias                                                                                                             | No administra usuarios ni aplica importaciones; aprobación final solo si se le concede explícitamente |

Reglas transversales:

- Todo permiso se valida en el backend; ocultar un botón en Angular no es autorización.
- Toda consulta operativa se filtra por membresía y negocio autorizado.
- Una sucursal enviada por el cliente debe pertenecer al negocio autorizado.
- El creador de una solicitud no puede autoaprobarla cuando la política del negocio exija separación.
- Los permisos se asignan a roles; no se crearán excepciones directas por usuario en esta fase.

## 4. Requerimientos funcionales

### RF-001 — Cuenta y sesión segura

**Actor autorizado:** Usuario no autenticado y usuario autenticado.

**Flujo esperado:** Registrar cuenta, verificar correo, iniciar sesión, activar MFA cuando sea obligatorio, consultar sesión, cerrar la sesión actual, revocar otras sesiones y recuperar contraseña.

**Criterios de aceptación:**

- Un correo sin verificar no puede realizar operaciones administrativas.
- Después del umbral configurado de intentos fallidos se aplica bloqueo temporal y rate limiting.
- Los tokens de verificación y recuperación son de un solo uso, expiran y se almacenan mediante hash.
- Cerrar o revocar una sesión invalida su acceso posterior.
- El sistema no registra contraseñas, códigos MFA, cookies ni tokens en logs o auditoría.

**Feature Trello:** `F01`.

### RF-002 — Crear negocio y membresía propietaria

**Actor autorizado:** Usuario verificado.

**Flujo esperado:** Registrar nombre y datos básicos del negocio; crear atómicamente la membresía propietaria y su configuración inicial.

**Criterios de aceptación:**

- El alta es transaccional: no puede quedar un negocio sin propietario.
- El usuario puede pertenecer posteriormente a más de un negocio.
- Los datos de un negocio nunca aparecen al consultar otro negocio sin membresía.

**Feature Trello:** `F01`.

### RF-003 — Administrar sucursales

**Actor autorizado:** Propietario o administrador.

**Flujo esperado:** Crear, consultar, editar y desactivar sucursales asociadas al negocio.

**Criterios de aceptación:**

- El negocio debe tener una sucursal principal activa.
- Desactivar una sucursal no elimina recepciones, costos, precios ni movimientos históricos.
- Las recepciones solo se registran en sucursales activas y autorizadas.

**Feature Trello:** `F01`.

### RF-004 — Administrar equipo y permisos

**Actor autorizado:** Propietario o administrador con `empleados.administrar`.

**Flujo esperado:** Invitar por correo, aceptar invitación, crear la relación de empleado, asignar sucursales y otorgar un rol permitido.

**Criterios de aceptación:**

- Las invitaciones son aleatorias, de un solo uso, con expiración y asociadas al negocio.
- Un capturista puede enviar una recepción, pero recibe `403` al intentar aprobar precios.
- Revocar una membresía impide nuevas operaciones y revoca sus sesiones según la política definida.

**Feature Trello:** `F01`.

### RF-005 — Configurar origen Odoo

**Actor autorizado:** Propietario o administrador con `catalogo.importar`.

**Flujo esperado:** Registrar un sistema externo de tipo `intercambio_archivos_odoo` y su versión de mapeo.

**Criterios de aceptación:**

- El mapeo `Odoo XLSX v1` queda versionado y no cambia silenciosamente.
- Las futuras modificaciones del formato crean una nueva versión.

**Feature Trello:** `F03`.

### RF-006 — Cargar y validar XLSX de Odoo

**Actor autorizado:** Propietario o administrador con `catalogo.importar`.

**Precondición:** Existe un archivo real aprobado y un mapeo `Odoo XLSX v1` congelado.

**Flujo esperado:** Cargar el archivo, calcular su hash, almacenarlo de forma privada, validar estructura y generar una vista previa por fila sin modificar el catálogo.

**Reglas:**

- Aceptar únicamente `.xlsx` sin macros, hasta 10 MB y 50,000 filas; límites configurables por ambiente.
- Validar ID Odoo, SKU, código, nombre, categoría, marca, unidad, presentación, conversión, costo, precio, impuestos y estado.
- Clasificar cada fila como válida, duplicada, incompleta o rechazada.
- Producir un archivo de errores con fila, campo, código y mensaje.

**Criterios de aceptación:**

- Una carga no modifica productos ni precios.
- Un archivo inválido no bloquea el proceso ni agota memoria, CPU o disco temporal.
- El mismo archivo se reconoce por hash y no crea un segundo lote aplicado accidentalmente.

**Feature Trello:** `F03`.

### RF-007 — Aplicar importación al catálogo local

**Actor autorizado:** Propietario o administrador con `catalogo.importar_aplicar`.

**Flujo esperado:** Confirmar un lote validado y crear o actualizar marcas, categorías, unidades, productos, variantes, presentaciones, códigos, impuestos, costo inicial y precio vigente.

**Reglas:**

- Prioridad de matching: ID externo → código de barras → SKU → revisión manual.
- Nunca empatar automáticamente únicamente por nombre.
- Conservar `mapeos_producto_externo` para actualizaciones futuras.
- Aplicar en una transacción controlada o en chunks idempotentes con estado por fila.
- Registrar costo inicial en el historial y precio importado mediante una vigencia, no como valor sobrescribible.

**Criterios de aceptación:**

- Reaplicar el mismo lote no duplica información.
- Una fila fallida puede diagnosticarse sin perder la evidencia original.
- El resumen muestra creados, actualizados, omitidos y fallidos.

**Feature Trello:** `F03`.

### RF-008 — Consultar catálogo

**Actor autorizado:** Miembro con `catalogo.ver`.

**Flujo esperado:** Buscar por código, SKU, nombre o ID externo y consultar presentación, unidad, costo de referencia, precio vigente y proveedor relacionado.

**Criterios de aceptación:**

- La búsqueda nunca devuelve productos de otro negocio.
- Los importes usan decimal exacto; no se usa coma flotante para cálculos monetarios.
- El producto conserva estado activo/inactivo sin eliminar sus históricos.

**Feature Trello:** `F03`.

### RF-009 — Administrar proveedores

**Actor autorizado:** Administrador con `proveedores.administrar`; el capturista puede consultar.

**Flujo esperado:** Crear, consultar, editar y desactivar proveedor; relacionar sus SKU y presentaciones conocidas.

**Criterios de aceptación:**

- RFC y datos fiscales son opcionales en el MVP, pero se validan si se proporcionan.
- Desactivar un proveedor no elimina recepciones ni costos históricos.
- No se permiten duplicados evidentes dentro del mismo negocio según la regla aprobada.

**Feature Trello:** `F04`.

### RF-010 — Capturar recepción en borrador

**Actor autorizado:** Capturista, revisor o administrador con `compras.registrar`.

**Flujo esperado:** Seleccionar negocio, sucursal y proveedor; capturar tipo de documento, folio, fecha y partidas con producto/presentación, cantidad, piezas por empaque, descuentos, IVA, IEPS y costo.

**Criterios de aceptación:**

- El borrador puede editarse sin afectar costos, precios ni inventario.
- Se conservan tanto los valores escritos en el documento como los valores normalizados.
- Los totales calculados y declarados muestran cualquier diferencia.
- La recepción usa una secuencia legible y un UUID interno.

**Feature Trello:** `F04`.

### RF-011 — Resolver partidas e incidencias

**Actor autorizado:** Capturista o administrador.

**Flujo esperado:** Buscar coincidencia en el catálogo; vincular manualmente cuando sea necesario o reportar producto nuevo, faltante, daño, rechazo, presentación incorrecta o datos incompletos.

**Criterios de aceptación:**

- Un producto desconocido no se crea ni se vincula automáticamente por nombre.
- La línea conserva el código y descripción originales aunque posteriormente se vincule.
- Una línea incompleta no puede llegar a autorización ni confirmar inventario.

**Feature Trello:** `F04`.

### RF-012 — Analizar variación de costo

**Actor autorizado:** Capturista para consultar; revisor o administrador para revisar.

**Flujo esperado:** Normalizar el costo a unidad base, localizar el costo anterior y calcular diferencia absoluta y porcentual.

**Orden para seleccionar costo anterior:**

1. Misma presentación y mismo proveedor.
2. Misma variante, sucursal y proveedor.
3. Misma variante y sucursal con cualquier proveedor.
4. Costo base importado de Odoo.
5. Sin referencia: producto o costo nuevo, sin porcentaje contra cero.

**Clasificación:**

- `producto_nuevo` o `producto_sin_vincular` si no existe mapeo confiable.
- `sin_cambio` cuando la diferencia absoluta está dentro de la tolerancia configurada.
- `cambio_costo` con dirección `subio` o `bajo` según el signo.
- `cambio_presentacion` cuando cambia la conversión o empaque relevante.
- `incompleto` cuando faltan datos; `error` ante inconsistencias técnicas.

**Criterios de aceptación:** Todos los valores usados en el cálculo quedan registrados y son reproducibles.

**Feature Trello:** `F05`.

### RF-013 — Sugerir precio

**Actor autorizado:** Usuario con `precios.ver`.

**Flujo esperado:** Aplicar una regla versionada de margen, incremento o precio fijo y su redondeo para producir una sugerencia.

**Criterios de aceptación:**

- La sugerencia indica regla, costo base, impuestos considerados, margen y redondeo.
- El precio vigente no cambia al calcular ni al enviar a revisión.
- La fórmula fiscal debe estar firmada por la responsable operativa antes de habilitar el cálculo automático.

**Feature Trello:** `F05`.

### RF-014 — Aprobar o rechazar cambios

**Actor autorizado:** Propietario, administrador o revisor con permiso explícito.

**Flujo esperado:** Consultar comparación, aceptar precio sugerido, capturar otro precio permitido o rechazar con motivo.

**Criterios de aceptación:**

- Aprobar cierra la vigencia anterior y crea una nueva fila de precio dentro de una transacción.
- Se registra historial de costo aun cuando el precio sea rechazado, según el estado final de la recepción.
- No existe `UPDATE` destructivo del precio o costo histórico.
- Repetir la aprobación con la misma clave idempotente devuelve el resultado existente.

**Feature Trello:** `F05`.

### RF-015 — Exportar cambios hacia Odoo

**Actor autorizado:** Administrador con `catalogo.exportar`.

**Flujo esperado:** Seleccionar cambios aprobados no exportados, crear lote y descargar XLSX o CSV conforme a la plantilla aprobada.

**Criterios de aceptación:**

- El lote conserva valores anteriores, nuevos, usuario, fecha y recepción de origen.
- Las celdas que comiencen con caracteres de fórmula se neutralizan.
- Regenerar o reintentar un lote no duplica su aplicación lógica.
- El archivo es privado y se entrega mediante acceso temporal de corta duración.

**Feature Trello:** `F05`.

### RF-016 — Confirmar recepción y entrada de inventario

**Actor autorizado:** Administrador con `compras.confirmar`.

**Precondición:** Todas las partidas están completas y resueltas. Las revisiones de precio pueden continuar su propio flujo, pero no pueden alterar cantidades recibidas.

**Flujo esperado:** Confirmar la recepción y crear un movimiento de entrada por partida, convertido a unidad base, actualizando el saldo de la sucursal en la misma transacción.

**Criterios de aceptación:**

- Cada partida crea como máximo un movimiento mediante unicidad por documento origen.
- No se modifica directamente el saldo sin movimiento de kardex.
- Reintentar la confirmación no vuelve a incrementar existencias.
- El kardex muestra fecha, producto, cantidad, saldo, sucursal, recepción y usuario.

**Feature Trello:** `F06`.

### RF-017 — Auditoría y trazabilidad

**Actor autorizado:** Administrador con `auditoria.ver`; escritura exclusiva del sistema.

**Flujo esperado:** Registrar autenticación relevante, importación, aplicación, recepción, autorización, precio, exportación y movimiento de inventario.

**Criterios de aceptación:**

- La bitácora registra acción, entidad, identificador, negocio, sucursal, usuario, resultado, fecha y `request_id`.
- Los valores sensibles se redactan antes de persistirse.
- La aplicación no ofrece edición o eliminación de auditoría.

**Feature Trello:** Todas; consolidación en `F07`.

## 5. Modelo conceptual mínimo

Las migraciones exactas se diseñarán desde estas responsabilidades, evitando tablas genéricas sin integridad.

| Bloque         | Entidades mínimas                                                                                                                                          |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Identidad      | usuarios, perfiles_usuario, tokens_autenticacion, sesiones_usuario, eventos_inicio_sesion, factores_mfa                                                    |
| Organización   | negocios, sucursales, membresias_negocio, empleados, asignaciones_empleado_sucursal, invitaciones_negocio                                                  |
| RBAC           | permisos, roles, permisos_rol, roles_membresia                                                                                                             |
| Catálogo       | marcas, categorias, unidades_medida, productos, variantes_producto, presentaciones_producto, codigos_barras_producto                                       |
| Fiscal/precios | definiciones_impuesto, impuestos_predeterminados_producto, reglas_precio, precios_producto, historial_costos_producto                                      |
| Odoo           | sistemas_externos, mapeos_producto_externo, lotes_importacion_catalogo, filas_importacion_catalogo, lotes_exportacion_catalogo, filas_exportacion_catalogo |
| Recepción      | proveedores, productos_proveedor, recepciones_compra, partidas_recepcion_compra, impuestos_partida_recepcion, incidencias_recepcion_compra                 |
| Autorización   | solicitudes_autorizacion, acciones_autorizacion, revisiones_precio                                                                                         |
| Inventario     | existencias_inventario, movimientos_inventario                                                                                                             |
| Plataforma     | archivos, operaciones_idempotentes, bitacora_auditoria                                                                                                     |

### 5.1 Reglas de datos

- UUID como identificador público; secuencias legibles solo como referencia humana.
- `numeric`/decimal para dinero, cantidades, tasas y factores; nunca `float`.
- Fechas almacenadas en UTC y presentadas según zona horaria del negocio.
- Índices únicos parciales para categoría raíz, presentaciones predeterminadas y precio vigente.
- Unicidad por negocio para códigos/SKU según el contrato aprobado.
- Llaves foráneas e índices en todas las relaciones de negocio, sucursal y documento.
- Bajas lógicas para maestros con historial; no usar borrado en cascada sobre operación real.
- Actualización de existencia con bloqueo/transacción para evitar pérdida de incrementos concurrentes.
- Migraciones versionadas y repetibles; `AutoMigrate` no se ejecuta en producción.

## 6. Contratos HTTP mínimos

Prefijo: `/api/v1`. Las rutas bajo negocio exigen membresía y permiso en cada solicitud.

| Método y ruta                                                  | Propósito                                                |
| -------------------------------------------------------------- | -------------------------------------------------------- |
| `POST /auth/register`                                          | Registrar cuenta                                         |
| `POST /auth/verify-email`                                      | Verificar correo con token de un solo uso                |
| `POST /auth/login`                                             | Iniciar sesión y completar MFA cuando aplique            |
| `POST /auth/mfa/setup`                                         | Preparar activación TOTP                                 |
| `POST /auth/mfa/confirm`                                       | Confirmar TOTP y generar códigos de recuperación         |
| `POST /auth/password/forgot`                                   | Solicitar recuperación sin revelar existencia del correo |
| `POST /auth/password/reset`                                    | Cambiar contraseña con token válido                      |
| `GET /auth/me`                                                 | Consultar usuario, membresías y sesión                   |
| `POST /auth/logout`                                            | Revocar sesión actual                                    |
| `GET/POST /negocios`                                           | Listar o crear negocio                                   |
| `GET/PATCH /negocios/:negocioId`                               | Consultar o actualizar negocio autorizado                |
| `GET/POST /negocios/:negocioId/sucursales`                     | Administrar sucursales                                   |
| `GET/POST /negocios/:negocioId/invitaciones`                   | Administrar invitaciones                                 |
| `GET /negocios/:negocioId/catalogo/productos`                  | Buscar catálogo                                          |
| `POST /negocios/:negocioId/catalogo/importaciones`             | Cargar XLSX                                              |
| `GET /negocios/:negocioId/catalogo/importaciones/:id`          | Consultar validación                                     |
| `POST /negocios/:negocioId/catalogo/importaciones/:id/aplicar` | Aplicar lote válido                                      |
| `GET/POST /negocios/:negocioId/proveedores`                    | Administrar proveedores                                  |
| `GET/POST /negocios/:negocioId/recepciones`                    | Listar o crear recepción                                 |
| `PATCH /negocios/:negocioId/recepciones/:id`                   | Editar borrador                                          |
| `POST /negocios/:negocioId/recepciones/:id/enviar`             | Enviar a revisión                                        |
| `POST /negocios/:negocioId/recepciones/:id/confirmar`          | Confirmar y afectar inventario                           |
| `POST /negocios/:negocioId/revisiones-precio/:id/acciones`     | Aprobar o rechazar                                       |
| `POST /negocios/:negocioId/catalogo/exportaciones`             | Generar archivo para Odoo                                |
| `GET /negocios/:negocioId/inventario/existencias`              | Consultar saldo por sucursal                             |
| `GET /negocios/:negocioId/inventario/kardex`                   | Consultar movimientos                                    |

### 6.1 Formato de error

```json
{
  "codigo": "IMPORTACION_FILA_INVALIDA",
  "mensaje": "El archivo contiene filas que requieren corrección",
  "detalles": [{ "fila": 18, "campo": "codigo_barras", "motivo": "duplicado" }],
  "request_id": "uuid"
}
```

### 6.2 Reglas de contrato

- `401` para sesión ausente o inválida; `403` para membresía o permiso insuficiente.
- Respuestas de login, recuperación e invitación no revelan información innecesaria.
- `Idempotency-Key` obligatorio al aplicar importaciones, confirmar recepciones, aprobar y exportar.
- Paginación y límites máximos obligatorios en catálogos, recepciones, auditoría y kardex.
- Content types permitidos explícitamente; solicitudes desconocidas se rechazan.
- Todo error incluye `request_id`; los detalles internos solo aparecen en logs protegidos.

## 7. Seguridad

Objetivo: aplicar los controles pertinentes de OWASP ASVS nivel 2. Esto no sustituye una auditoría independiente ni permite afirmar que un sistema es invulnerable.

### SEG-001 — Autenticación y sesión

- Cookies de sesión `HttpOnly`, `Secure`, `SameSite` y con alcance mínimo.
- Producción same-origin detrás del reverse proxy; CORS cerrado por defecto.
- Protección CSRF para toda operación que cambia estado.
- Rotación del identificador de sesión al autenticar, elevar privilegios o cambiar contraseña.
- Sesiones almacenadas/revocables con expiración absoluta e inactividad.
- MFA TOTP obligatorio para roles privilegiados y códigos de recuperación de un solo uso.
- Migración progresiva de bcrypt existente a Argon2id después de un login válido.
- Respuestas uniformes para evitar enumeración de cuentas.

### SEG-002 — Autorización y multiempresa

- Denegar por defecto y comprobar permiso en cada caso de uso.
- Obtener identidad desde la sesión validada; nunca desde IDs de usuario enviados por el cliente.
- Comprobar que negocio, sucursal y entidad relacionada pertenecen al mismo contexto autorizado.
- Pruebas negativas para lectura, escritura, IDs adivinados y referencias cruzadas entre negocios.
- Auditoría obligatoria al cambiar membresía, rol, importación, precio, recepción o inventario.

### SEG-003 — Entradas, archivos y exportaciones

- Allowlist de campos, enums, tamaños, formatos, MIME y firma real del archivo.
- Rechazar XLSM, macros, enlaces externos, objetos incrustados y estructuras fuera del contrato.
- Limitar ZIP descomprimido, número de hojas, filas, columnas, longitud de celda, CPU y tiempo.
- Procesar fuera del proceso HTTP cuando el tamaño lo requiera y borrar temporales de forma segura.
- No ejecutar comandos, fórmulas, URLs, plantillas ni código proveniente del documento.
- Neutralizar `=`, `+`, `-`, `@`, tabulador y retorno al inicio de celdas exportadas.
- Almacenamiento privado y descargas mediante URL temporal o streaming autorizado.

### SEG-004 — API y servidor

- Timeouts HTTP, límite de body, rate limiting por IP/cuenta y máximo de conexiones.
- Validación del lado servidor, consultas parametrizadas y transacciones con contexto cancelable.
- Encabezados HSTS, `X-Content-Type-Options`, política de referrer, permisos y CSP probada.
- Sin endpoints de depuración, perfiles, métricas o documentación expuestos públicamente.
- Egress restringido; la aplicación no aceptará URLs arbitrarias ni actuará como proxy.
- Logs estructurados sin secretos, PII innecesaria, tokens, documentos completos o consultas sensibles.
- Secretos fuera de imágenes, Compose, repositorio, argumentos de build y salida de CI.

### SEG-005 — Cadena de suministro y operación

- Dependencias fijadas mediante lockfiles y revisión de actualizaciones.
- CI con pruebas, `govulncheck`, auditoría npm, escaneo de secretos, SBOM y Trivy.
- Imágenes mínimas, firmadas o identificadas por digest para producción.
- Parches críticos con SLA definido; rollback y despliegue reproducible.
- Alertas por errores de autenticación, rate limits, importaciones fallidas, disco, memoria y disponibilidad.

Referencias:

- [OWASP Application Security Verification Standard](https://owasp.org/www-project-application-security-verification-standard/)
- [OWASP Docker Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [Docker Engine Security](https://docs.docker.com/engine/security/)
- [Docker Compose in production](https://docs.docker.com/compose/how-tos/production/)

## 8. Docker e infraestructura

### INF-001 — Entorno reproducible de desarrollo

`compose.yaml` levantará:

- PostgreSQL local con volumen nombrado y healthcheck.
- API Go con configuración de desarrollo.
- Frontend Angular para desarrollo o una variante servida por Nginx.
- Almacenamiento S3 compatible local solo si se necesita probar archivos sin usar el proveedor remoto.
- Servicio de correo local para verificar correos sin enviar mensajes reales.

No se incluirán secretos reales. Un `.env.example` documentará únicamente nombres y valores seguros de ejemplo.

### INF-002 — Staging y producción endurecidos

- Imagen frontend multi-stage: compilar Angular y servir artefactos estáticos mediante Nginx no root.
- Imagen backend multi-stage: compilar Go de forma reproducible y ejecutar binario mínimo como UID sin privilegios.
- Reverse proxy como único punto público en `80/443`; redirección permanente a HTTPS.
- API accesible solo desde la red interna del proxy.
- PostgreSQL administrado fuera de Compose, con TLS, red privada y usuario de mínimo privilegio.
- Almacenamiento S3 compatible privado para importaciones, exportaciones y respaldos secundarios.
- Contenedores con `read_only`, `tmpfs` limitado, `cap_drop: ALL`, `no-new-privileges`, límites CPU/memoria/PIDs y healthchecks, agregando solo excepciones justificadas.
- Sin Docker socket, modo privilegiado, host networking ni puertos administrativos públicos.
- SSH limitado por firewall/VPN/IP autorizada; autenticación por llave y sin contraseña.
- Staging aislado de producción, con base, bucket, secretos y dominio independientes.

### INF-003 — Continuidad y recuperación

- Backups automáticos/PITR de PostgreSQL administrado y exportación secundaria cifrada.
- Versionado y lifecycle en objetos; archivos temporales expiran automáticamente.
- Simulacro documentado de restauración antes de la salida y posteriormente de forma periódica.
- Runbook de caída, rollback, compromiso de credenciales, falta de espacio y corrupción de importación.

## 9. Tarjetas maestras para Trello

### F01 — Identidad, negocio, sucursales y RBAC

**Responsable:** Socio A  
**Revisor:** Socio B  
**Fecha objetivo:** Fin de semana 2  
**Dependencias:** Ninguna

**Resultado:** Contexto multiempresa seguro con cuentas verificadas, MFA privilegiado, sesiones revocables, negocio, sucursales, equipo y permisos mínimos.

**Checklist inicial:**

- [ ] Crear threat model de identidad, sesiones y aislamiento.
- [ ] Diseñar migraciones versionadas para identidad y organización.
- [ ] Endurecer cookies, CSRF, bloqueo, recuperación y sesiones.
- [ ] Implementar verificación de correo y MFA TOTP.
- [ ] Implementar negocio, sucursales, membresías e invitaciones.
- [ ] Sembrar permisos y roles mínimos.
- [ ] Construir pantallas responsive de negocio, sucursales y equipo.
- [ ] Probar matriz RBAC y aislamiento entre dos negocios.
- [ ] Documentar flujos y decisiones.

**Criterio de cierre:** Cumple `RF-001` a `RF-004` y controles aplicables `SEG-001/002`.

### F02 — Docker, CI, publicación y continuidad

**Responsable:** Socio B  
**Revisor:** Socio A  
**Fecha objetivo:** Base en semana 1; cierre en semana 6  
**Dependencias:** Accesos al proveedor de infraestructura

**Resultado:** Desarrollo reproducible y staging/producción endurecidos con base administrada, almacenamiento privado, observabilidad y restauración probada.

**Checklist inicial:**

- [ ] Crear Dockerfiles multi-stage y `.dockerignore`.
- [ ] Crear Compose de desarrollo y override de producción.
- [ ] Separar redes, puertos, secretos y ambientes.
- [ ] Configurar TLS, headers, límites y healthchecks.
- [ ] Crear pipeline de pruebas y escaneos.
- [ ] Configurar migraciones previas al despliegue.
- [ ] Configurar logs, métricas, alertas y retención.
- [ ] Configurar backups y ejecutar restauración real.
- [ ] Crear runbooks y procedimiento de rollback.

**Criterio de cierre:** Se puede reconstruir, desplegar, observar, respaldar, restaurar y regresar de versión usando documentación versionada.

### F03 — Catálogo e importación XLSX de Odoo

**Responsable:** Socio B  
**Revisor:** Socio A  
**Fecha objetivo:** Fin de semana 3  
**Dependencias:** `F01`, XLSX real y aprobación del mapeo v1

**Resultado:** Catálogo local creado mediante validación y aplicación idempotente de una exportación real de Odoo.

**Checklist inicial:**

- [ ] Perfilar el XLSX real y congelar el contrato `Odoo XLSX v1`.
- [ ] Diseñar migraciones de catálogo, históricos, lotes y mapeos.
- [ ] Implementar almacenamiento privado, hash y límites del archivo.
- [ ] Implementar validación y previsualización sin efectos.
- [ ] Implementar descarga de errores por fila.
- [ ] Implementar matching seguro e idempotente.
- [ ] Implementar aplicación transaccional/por chunks.
- [ ] Crear pantallas de carga, progreso, errores y resultado.
- [ ] Probar duplicados, reintentos y archivos hostiles.

**Criterio de cierre:** Cumple `RF-005` a `RF-008` y `SEG-003`.

### F04 — Proveedores y recepción de mercancía

**Responsable:** Socio A  
**Revisor:** Socio B  
**Fecha objetivo:** Fin de semana 4  
**Dependencias:** `F01` y catálogo consultable de `F03`

**Resultado:** Un capturista registra un documento real, resuelve sus partidas e incidencias y lo envía a revisión sin afectar inventario ni precios prematuramente.

**Checklist inicial:**

- [ ] Diseñar migraciones de proveedores y recepción.
- [ ] Implementar permisos y endpoints acotados al negocio.
- [ ] Implementar formulario responsive con borrador servidor.
- [ ] Implementar cálculos y conciliación de totales.
- [ ] Implementar búsqueda/vinculación de catálogo.
- [ ] Implementar producto nuevo, faltantes y demás incidencias.
- [ ] Implementar estados y bloqueo de ediciones inválidas.
- [ ] Probar concurrencia, permisos y recuperación de errores.

**Criterio de cierre:** Cumple `RF-009` a `RF-011`.

### F05 — Análisis, autorización y exportación

**Responsable:** Socio B  
**Revisor:** Socio A  
**Fecha objetivo:** Fin de semana 5  
**Dependencias:** `F03`, `F04` y fórmula fiscal aprobada

**Resultado:** Variaciones explicables, sugerencias reproducibles, aprobación con separación de funciones y exportación segura para Odoo.

**Checklist inicial:**

- [ ] Obtener firma de la fórmula de costo, impuestos y tolerancia.
- [ ] Implementar selección de costo anterior y clasificación.
- [ ] Implementar reglas versionadas de precio y redondeo.
- [ ] Implementar bandeja y acciones de autorización.
- [ ] Implementar vigencias e históricos transaccionales.
- [ ] Implementar exportación y neutralización de fórmulas.
- [ ] Implementar pantallas de comparación y aprobación.
- [ ] Probar cálculos frontera, rechazo, reintento y concurrencia.

**Criterio de cierre:** Cumple `RF-012` a `RF-015`.

### F06 — Entradas, existencias y kardex mínimo

**Responsable:** Socio A  
**Revisor:** Socio B  
**Fecha objetivo:** Fin de semana 5  
**Dependencias:** Recepción completa de `F04`

**Resultado:** Confirmar una recepción genera una sola entrada y un saldo consultable por producto y sucursal.

**Checklist inicial:**

- [ ] Diseñar saldo, movimiento y unicidad por origen.
- [ ] Implementar confirmación transaccional.
- [ ] Implementar conversión a unidad base.
- [ ] Implementar consulta de existencias y kardex.
- [ ] Implementar pantallas responsive y filtros mínimos.
- [ ] Probar doble clic, reintento, concurrencia y rollback.

**Criterio de cierre:** Cumple `RF-016` sin incorporar salidas, ajustes ni conteos.

### F07 — Integración, hardening y salida controlada

**Responsables:** Socio A y Socio B  
**Fecha objetivo:** Fin de semana 6  
**Dependencias:** `F01` a `F06`

**Resultado:** Flujo completo validado con operación real controlada, controles de seguridad comprobados y rollback disponible.

**Checklist inicial:**

- [ ] Ejecutar E2E del flujo completo con copia segura del XLSX real.
- [ ] Ejecutar matriz RBAC y aislamiento multiempresa.
- [ ] Ejecutar casos idempotentes y concurrencia.
- [ ] Ejecutar pruebas de archivos hostiles y límites.
- [ ] Ejecutar escaneos SAST/SCA/secretos/contenedores y ZAP en staging.
- [ ] Corregir hallazgos críticos y altos antes de publicar.
- [ ] Restaurar un backup y documentar tiempos/resultados.
- [ ] Ejecutar aceptación con responsable operativa.
- [ ] Preparar rollback, soporte y monitoreo del lanzamiento.

**Criterio de cierre:** Todos los criterios de la sección 1.2 tienen evidencia y no existen vulnerabilidades críticas o altas abiertas sin aceptación formal del riesgo.

## 10. Calendario de seis semanas

| Semana | Socio A                           | Socio B                                  | Hito                                  |
| ------ | --------------------------------- | ---------------------------------------- | ------------------------------------- |
| 1      | Threat model, negocio y sucursal  | Docker dev/CI, perfilado XLSX y catálogo | Contratos y migraciones aprobados     |
| 2      | Sesiones, MFA, membresías y RBAC  | Validación y preview de importación      | Contexto seguro e importación visible |
| 3      | Proveedores y recepción base      | Aplicación idempotente del catálogo      | Catálogo real disponible              |
| 4      | Recepción completa e incidencias  | Motor de costos y sugerencias            | Recepción real analizada              |
| 5      | Entrada, saldo y kardex           | Autorización, históricos y exportación   | Flujo funcional completo              |
| 6      | Pruebas de negocio y correcciones | Infra, hardening, backup y rollback      | Aceptación y publicación controlada   |

## 11. Política del tablero Trello

### 11.1 Listas

```text
Backlog
Por hacer
En curso
En revisión
Bloqueado
Terminado
```

### 11.2 Reglas

- Las tarjetas `F01`–`F07` son features maestras; no se reemplazan con tareas pequeñas.
- Cada responsable desglosa su feature en tarjetas enlazadas o checklists antes de moverla a `Por hacer`.
- WIP máximo: una feature maestra `En curso` por socio.
- Toda feature requiere revisión del otro socio mediante pull request.
- Una tarjeta bloqueada indica causa, responsable externo y siguiente fecha de revisión.
- `Terminado` exige código integrado, migraciones, pruebas, documentación, controles de seguridad y evidencia de aceptación.
- Una corrección de seguridad crítica interrumpe el WIP normal.

### 11.3 Etiquetas

- `P0 Producción`
- `P1 Importante`
- `Seguridad`
- `Infraestructura`
- `Odoo`
- `Catálogo`
- `Recepción`
- `Precios`
- `Inventario`
- `Bloqueante`

### 11.4 Plantilla para subtareas

```md
## Resultado esperado

## Requerimientos relacionados

- RF-000
- SEG-000

## Criterios de aceptación

- [ ]

## Dependencias

## Pruebas requeridas

- [ ] Unitarias
- [ ] Integración
- [ ] Autorización/multiempresa
- [ ] Responsive o E2E, cuando aplique

## Evidencia

- Pull request:
- Capturas/logs:
- Decisiones:
```

## 12. Estrategia de pruebas

| Nivel       | Cobertura mínima                                                                            |
| ----------- | ------------------------------------------------------------------------------------------- |
| Unitarias   | Cálculos, validadores, permisos, estados y normalización                                    |
| Integración | PostgreSQL real, migraciones, transacciones, índices e idempotencia                         |
| Contrato    | Status codes, errores, paginación, headers y límites                                        |
| Seguridad   | Auth, CSRF, RBAC, aislamiento, enumeración, upload hostil y rate limiting                   |
| E2E         | Alta → negocio → importación → recepción → análisis → aprobación → inventario → exportación |
| Operación   | Contenedores, healthchecks, alertas, backup, restauración y rollback                        |
| UI          | Escritorio, tablet y móvil; teclado, foco, errores y ausencia de scroll horizontal          |

Casos obligatorios:

1. Un miembro de Negocio A no puede leer ni modificar IDs válidos de Negocio B.
2. Un capturista no puede aplicar importaciones ni aprobar precios usando la API directamente.
3. El mismo XLSX, recepción, aprobación o confirmación reenviados no duplican efectos.
4. Dos confirmaciones concurrentes producen un solo movimiento y un saldo correcto.
5. Un XLSX corrupto, enorme, comprimido maliciosamente o con fórmulas no compromete el servidor.
6. Fallar a mitad de una transacción no deja catálogo, precio o inventario parcial.
7. Restaurar el último backup produce una aplicación consistente y verificable.

## 13. Decisiones bloqueantes y riesgos

| Bloqueante/riesgo              | Tratamiento                                                   | Responsable                     | Fecha límite         |
| ------------------------------ | ------------------------------------------------------------- | ------------------------------- | -------------------- |
| Falta de XLSX real             | Obtener exportación representativa antes de programar parsing | Responsable operativa + Socio B | Semana 1             |
| Columnas Odoo variables        | Congelar y versionar contrato `Odoo XLSX v1`                  | Socio B, aprobado por operación | Semana 1             |
| Fórmula de IVA/IEPS/descuentos | Documentar ejemplos y obtener aprobación escrita              | Operación + Socio B             | Semana 3             |
| Alcance excesivo               | Aplicar estrictamente la sección 2.2                          | Ambos                           | Continuo             |
| Acceso cruzado multiempresa    | Guardas de autorización y suite negativa obligatoria          | Dueño de cada feature           | Antes de revisión    |
| Duplicados por reintento       | Idempotencia y restricciones únicas                           | Dueño de cada operación         | Antes de integración |
| Compromiso del servidor        | Hardening, mínimo privilegio, egress restringido y monitoreo  | Socio B                         | Antes de producción  |
| Pérdida de datos               | Backup automático y restauración comprobada                   | Socio B, revisa Socio A         | Semana 6             |

## 14. Control de cambios

- Este archivo es la fuente de verdad del alcance de Fase 1.
- Todo cambio de alcance debe registrar fecha, motivo, impacto, aprobador y features afectadas.
- Trello refleja ejecución; no sustituye reglas de negocio ni criterios de aceptación.
- Socio A y Socio B deben reemplazarse por los nombres reales al crear el tablero.
- Las decisiones del XLSX y de la fórmula fiscal se anexarán aquí antes de implementar sus respectivos bloques.
