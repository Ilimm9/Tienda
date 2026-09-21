# Plan de endurecimiento de la aplicación antes de producción

## Estado

Fases 1, 2 y 3 implementadas y verificadas mediante pruebas automatizadas. La migración de la base local legacy del puerto `5433` está detenida de forma segura hasta resolver o eliminar sus catálogos sin propietario. Fases 4 a 7 pendientes de aprobación explícita.

## Relación con la capability

Este documento desarrolla únicamente los cambios de aplicación de la [Capability 05](CAPABILITY.md). La creación de la instancia, DNS, TLS y demás trabajo de AWS Lightsail se realizará después y no forma parte de estas fases inmediatas.

## Decisiones confirmadas

- El primer lanzamiento tendrá únicamente el frontend Angular y la API Go bajo el mismo origen.
- No habrá por ahora aplicación móvil, API pública para terceros ni servicios distribuidos que necesiten portar credenciales entre clientes.
- La autenticación web dejó de usar JWT autocontenido y ahora utiliza sesiones opacas revocables del lado servidor.
- No se implementaron access token y refresh token en esta etapa. Su reevaluación quedó condicionada a la aparición de un cliente móvil, una API externa o una arquitectura distribuida.
- Marcas, categorías y unidades de medida pertenecen a un negocio; no son catálogos globales compartidos.
- Amazon Simple Email Service (SES) será el proveedor de correo para OTP de verificación y, posteriormente, recuperación e invitaciones.
- Las fases 1, 2 y 3 fueron autorizadas explícitamente el 2026-09-20. Las fases 4 a 7 no están autorizadas.

## Por qué no se implementó access token + refresh token

Para una SPA y una API del mismo origen, el navegador puede enviar una cookie de sesión `HttpOnly` sin exponer la credencial a JavaScript. Una sesión opaca permite consultar su estado en PostgreSQL y revocarla inmediatamente.

Usar JWT corto más refresh token exigiría además:

- rotación de refresh tokens;
- detección de reutilización;
- tratamiento de solicitudes concurrentes y varias pestañas;
- dos expiraciones y dos credenciales;
- persistencia de revocación, que elimina la principal ventaja de un JWT autocontenido.

La sesión opaca cubre el caso actual con menos estados y menos superficie de ataque. Los tokens de autenticación nunca se almacenarán en `localStorage` ni `sessionStorage`.

## Dependencias y orden obligatorio

```text
Fase 0: alcance y línea base — completada
    ↓
Fase 1: rutas, permisos y catálogo por negocio — implementada
    ↓
Fase 2: sesiones opacas y CSRF — implementada
    ↓
Fase 3: registro pendiente y OTP — implementada y verificada
    ↓
Fase 4: integración con Amazon SES
    ↓
Fase 5: recuperación e invitaciones por correo
    ↓
Fase 6: límites, timeouts y configuración segura
    ↓
Fase 7: auditoría y verificación integral
```

SES se dejó para una fase posterior porque primero se cerraron las rutas y las sesiones. Verificar un correo no habría corregido una API con operaciones sin autorización.

## Fase 0 — alcance, conflictos y línea base (completada)

### Resultado

Se revisó el estado previo antes de cambiar contratos o datos:

- Se confirmó que el primer cliente será Angular y la API bajo el mismo origen.
- Se confirmó que marcas, categorías y unidades pertenecen a cada negocio.
- Se inventariaron las rutas públicas, autenticadas y administrativas relevantes para las fases 1 y 2.
- Se identificaron las mutaciones globales de catálogo, la falta de permiso RBAC conectado y el uso anterior de JWT.
- Se revisaron los modelos, asociaciones, pruebas, configuración y contenedores involucrados.
- Se preservó el trabajo existente de otras capabilities y no se ejecutaron operaciones Git destructivas.
- Se distinguieron las dos bases PostgreSQL locales: `5433` para la ejecución directa y `5434` para Docker Compose.
- Se estableció una base de pruebas antes de modificar autorización, catálogo y sesiones.

## Fase 1 — rutas, permisos y aislamiento de catálogos (implementada)

### Resultado

Se cerraron las mutaciones anónimas del catálogo y se aplicó autorización en el backend:

- Se eliminaron las rutas administrativas globales `/api/v1/catalogo/*`.
- Las rutas de productos, marcas, categorías, unidades, proveedores e importaciones quedaron bajo `/api/v1/negocios/:negocioId`.
- `RequireAuth` protege los endpoints privados y `RequireNegocioActivo` confirma la membresía activa.
- Las lecturas del catálogo exigen `catalogo.ver`.
- Las altas, cambios, desactivaciones e importaciones exigen `catalogo.gestionar`.
- La autorización ya no depende de guards, botones ni datos guardados por Angular.

El catálogo quedó modelado por negocio:

- Se agregó `negocio_id` a marcas, categorías, unidades, productos, categorías de producto y códigos de producto.
- Todas las consultas, actualizaciones e importaciones filtran por negocio.
- Una marca, categoría, unidad o categoría padre de otro negocio es rechazada.
- Los nombres, códigos, símbolos y códigos de barras son únicos dentro de cada negocio, no globalmente.
- PostgreSQL respalda estas reglas con columnas obligatorias, índices por negocio y claves foráneas compuestas.
- Angular consume las nuevas rutas por negocio utilizando el negocio activo del contexto.

### Contratos HTTP activos

```text
GET|POST   /api/v1/negocios/:negocioId/catalogo/marcas
PATCH      /api/v1/negocios/:negocioId/catalogo/marcas/:id

GET|POST   /api/v1/negocios/:negocioId/catalogo/categorias
PATCH      /api/v1/negocios/:negocioId/catalogo/categorias/:id

GET|POST   /api/v1/negocios/:negocioId/catalogo/unidades-medida
PATCH      /api/v1/negocios/:negocioId/catalogo/unidades-medida/:id
```

Productos, proveedores, plantillas e importaciones siguen la misma jerarquía y sus permisos correspondientes.

### Migración y estado de datos

- Se implementó una migración idempotente que completa propietarios conocidos y activa las restricciones después del backfill.
- La migración se detiene antes de modificar datos si encuentra productos compartidos o catálogos cuyo negocio no puede determinarse con seguridad.
- La prueba de integración sobre `tienda_security_test` pasó desde un esquema vacío y también al ejecutar la migración por segunda vez.
- Se verificó que dos negocios pueden tener una marca con el mismo nombre y que uno no puede modificar la marca del otro.
- La base legacy usada por `backend/.env` en el puerto `5433` contiene tres negocios y catálogos globales. Por ello la migración se detuvo correctamente con `catálogos sin dueño y múltiples negocios disponibles`.
- El código de la fase está terminado; esa base local requiere ser reconstruida o recibir una asignación explícita de sus datos antes de arrancar con el esquema nuevo.

## Fase 2 — sesiones opacas revocables y CSRF (implementada)

### Resultado

La sesión web dejó de usar JWT:

- Se eliminó la emisión, interpretación y dependencia de JWT.
- Se creó `sesiones_usuario` con usuario, hashes de sesión y CSRF, IP, agente de usuario, emisión, expiración, última actividad y revocación.
- Los tokens se generan con un CSPRNG de 256 bits, se entregan una sola vez y PostgreSQL conserva únicamente sus hashes SHA-256.
- `RequireAuth` comprueba que la sesión exista, no esté revocada ni expirada y que el usuario continúe activo y con correo verificado.
- `/auth/me` usa la identidad resuelta por el middleware.
- Logout revoca la sesión en PostgreSQL antes de eliminar las cookies.
- Dos sesiones del mismo usuario son independientes y revocar una no elimina la otra.
- `RevokeAll` quedó disponible para conectarlo a cambio y recuperación de contraseña cuando esos flujos sean implementados.
- Una tarea periódica elimina sesiones expiradas o revocadas después de la retención configurada.

### Duraciones y cookies activas

- Sesión normal: 24 horas de expiración absoluta.
- Sesión con `recordarme`: 30 días de expiración absoluta y máximo 7 días sin actividad.
- La última actividad se actualiza como máximo una vez cada 5 minutos.
- Las duraciones se validan al arrancar; un valor vacío o inválido impide iniciar la API.
- Producción utiliza `__Host-tienda_session` con `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/` y sin `Domain`.
- Desarrollo conserva el nombre `tienda_session` y permite HTTP local de forma explícita.
- Las cookies JWT anteriores quedaron invalidadas y los usuarios deben iniciar sesión nuevamente al activar esta versión.

### Protección CSRF activa

- El login entrega una cookie `XSRF-TOKEN` separada y ligada criptográficamente a la sesión almacenada.
- Angular envía `X-XSRF-TOKEN` en las mutaciones relativas al mismo origen.
- El backend compara cookie, cabecera y hash de la sesión antes de aceptar `POST`, `PUT`, `PATCH` o `DELETE` autenticados.
- El backend también valida que `Origin` coincida exactamente con `FRONTEND_URL` en todos los métodos con efecto, incluidos login y registro.
- CORS quedó limitado al origen configurado y permite explícitamente la cabecera CSRF.

### Verificación realizada

- Las pruebas cubren generación aleatoria, persistencia por hash, expiración, inactividad, revocación y usuario deshabilitado o sin verificar.
- Se comprobaron los atributos de las cookies de producción.
- Se probaron CSRF válido, incorrecto y ligado a otra credencial, además del rechazo de origen ajeno.
- La integración confirmó que una sesión revocada no puede reutilizarse y que otra sesión del mismo usuario permanece vigente.
- `go test ./...`, `go vet ./...`, las 108 pruebas Angular y la construcción de ambas imágenes Docker terminaron correctamente.

### Compatibilidad con la fase 3

Las cuentas activas anteriores recibieron una marca de verificación mediante una migración ejecutada una sola vez. Desde la implementación de la fase 3, todo registro web nuevo queda en `pendiente_verificacion` y sólo pasa a `activo` al validar correctamente el OTP.

## Fase 3 — registro pendiente y OTP de correo

**Estado: implementada y verificada el 2026-09-20.** El envío local usa Mailpit mediante un adaptador SMTP explícito; Amazon SES permanece fuera de alcance hasta la fase 4.

### Objetivo

Impedir que una cuenta nueva se active antes de demostrar control del correo registrado.

### Modelo

Crear un desafío o token de autenticación con:

- `id` opaco usado como identificador del desafío;
- `usuario_id`;
- propósito `verificacion_correo`;
- hash/HMAC del OTP;
- número de intentos fallidos;
- fecha de último envío;
- `expira_en`;
- `usado_en`;
- `creado_en`.

Un OTP numérico tiene poca entropía. No basta con SHA-256: se almacenará un HMAC con secreto del servidor para que una copia de la base no permita probar rápidamente el millón de combinaciones.

### Flujo de registro

1. `POST /api/v1/auth/register`
   - normaliza correo;
   - crea o reutiliza de forma controlada una cuenta `pendiente_verificacion`;
   - persiste contraseña con bcrypt;
   - genera desafío y OTP;
   - intenta enviar el correo;
   - responde `202 Accepted` sin crear sesión.

2. `POST /api/v1/auth/verificar-correo`
   - recibe identificador del desafío y OTP;
   - verifica propósito, vigencia, uso e intentos;
   - marca `correo_verificado_en`;
   - cambia el usuario a `activo`;
   - invalida los demás desafíos de verificación;
   - crea la primera sesión opaca.

3. `POST /api/v1/auth/reenviar-verificacion`
   - responde de forma uniforme;
   - invalida el OTP anterior;
   - genera un desafío nuevo;
   - aplica espera mínima y límites por correo e IP.

### Reglas iniciales

- OTP de 6 dígitos generado criptográficamente.
- Vigencia de 10 minutos.
- Máximo 5 intentos por desafío.
- Mínimo 60 segundos entre envíos.
- Máximo 5 envíos por hora por correo y por IP.
- Un solo uso.
- Nunca registrar OTP, HMAC, contraseña o cuerpo completo del correo.
- Login de usuario pendiente responde igual que otras credenciales no utilizables.
- Solicitar OTP para un correo ya verificado no modifica ni bloquea la cuenta.
- Las respuestas no confirman innecesariamente si el correo existe.

### Frontend

- Registro conduce a una pantalla de verificación.
- El correo normalizado puede mostrarse parcialmente oculto.
- Entrada de OTP accesible y pegable, sin seis inputs que dificulten lectores de pantalla.
- Contador de reenvío es informativo; el backend decide el límite real.
- Recargar la página conserva únicamente el identificador no secreto del desafío cuando sea necesario.
- Verificación exitosa carga usuario y contexto y redirige al flujo inicial.

### Pruebas obligatorias

- registro crea cuenta pendiente y no sesión;
- OTP correcto activa una vez;
- OTP incorrecto incrementa intentos;
- expirado, usado o agotado se rechaza;
- reenvío invalida el anterior;
- límites por correo e IP;
- concurrencia: dos verificaciones no activan dos veces;
- usuario ya verificado no se degrada a pendiente;
- errores externos no enumeran cuentas;
- el frontend no conserva OTP ni credenciales.

### Criterios de aceptación

- Ninguna cuenta registrada por web queda activa sin correo verificado.
- OTP en claro sólo existe durante generación y envío.
- La verificación crea una sesión opaca, no un JWT.
- Ataques de repetición y fuerza bruta están limitados y auditados.

## Fase 4 — integración con Amazon Simple Email Service

### Objetivo

Enviar OTP transaccional sin acoplar el dominio `cuenta` al SDK de AWS.

### Contrato de aplicación

Definir una interfaz en `application/cuenta`, conceptualmente:

```text
EnviadorCorreo
└── EnviarOTPVerificacion(ctx, destinatario, codigo, expiraEn)
```

Implementaciones:

- doble/fake determinista para pruebas;
- implementación local explícita para desarrollo;
- adaptador SES v2 en `infrastructure/cuenta`.

La capa de dominio y aplicación no importará paquetes AWS.

### Estrategia inicial de envío

Para el primer volumen, el envío será síncrono y acotado por timeout:

1. Persistir cuenta pendiente y desafío.
2. Intentar envío por SES.
3. Si SES falla, conservar la cuenta pendiente.
4. Responder con error reintentable sin revelar detalles AWS.
5. Permitir reenvío, que invalida el desafío anterior.

No se introducirá outbox o cola hasta que volumen, latencia o disponibilidad lo justifiquen. Esta decisión evita almacenar temporalmente el OTP en claro para un worker.

### Configuración SES necesaria

- Verificar el dominio remitente.
- Configurar DKIM, SPF y DMARC.
- Solicitar acceso de producción; una cuenta nueva permanece inicialmente en sandbox.
- Usar identidad y región explícitas por ambiente.
- Conceder a la aplicación únicamente permiso de envío desde la identidad aprobada.
- No incrustar credenciales AWS en código, imagen o frontend.
- Definir configuration set para eventos transaccionales.
- Registrar identificador de mensaje, latencia y resultado sin registrar contenido ni OTP.
- Preparar recepción de rebotes y quejas antes de aumentar volumen.

En el sandbox de SES sólo se puede enviar a destinatarios verificados, con cuotas iniciales de 200 mensajes cada 24 horas y uno por segundo. La salida del sandbox y la verificación de identidad son dependencias externas, no pruebas de código.

### Manejo de fallos

- Timeout, throttling o error 5xx de SES no activa la cuenta.
- Reintentos automáticos dentro de una petición serán limitados para no duplicar correos.
- El usuario puede pedir reenvío después de la ventana establecida.
- Rebote permanente marca el correo como no entregable para evitar dañar reputación.
- Quejas y rebotes generan eventos operativos, no contienen OTP.

### Pruebas obligatorias

- adaptador con cliente SES simulado;
- remitente, destinatario, asunto y contenido correctos;
- timeout, throttling y error permanente;
- jamás registrar código o contenido sensible;
- configuración ausente falla al arrancar en producción;
- prueba manual con mailbox simulator y direcciones verificadas en sandbox.

### Criterios de aceptación

- Pruebas no dependen de AWS real.
- Producción no inicia sin remitente, región y credenciales válidas por el mecanismo aprobado.
- Fallo de correo deja un estado recuperable por reenvío.
- Métricas y eventos permiten detectar degradación de entrega.

## Fase 5 — recuperación de contraseña e invitaciones

### Objetivo

Reutilizar la infraestructura de correo sin reutilizar credenciales entre propósitos.

### Recuperación de contraseña

- `POST /auth/solicitar-recuperacion` siempre devuelve respuesta uniforme.
- Token aleatorio de un solo uso, con propósito propio y expiración corta.
- El enlace se construye desde una URL base permitida, nunca desde `Host` recibido sin validar.
- La pantalla de recuperación usa `Referrer-Policy: no-referrer`.
- Nueva contraseña aplica la misma política que registro.
- Recuperación correcta revoca todas las sesiones.
- Se envía notificación posterior sin incluir la contraseña.

Para recuperación se prefiere un enlace con token aleatorio largo en lugar de OTP numérico, porque reduce fuerza bruta y sigue la guía de OWASP. El token se guarda como hash y se consume una sola vez.

### Invitaciones

- Mantener los tokens aleatorios y hash actuales.
- Enviar el enlace por SES además de permitir copiarlo desde la interfaz administrativa.
- No registrar el token o URL completa.
- Confirmar que la cuenta autenticada tiene el mismo correo invitado.
- Rate limit por negocio, usuario remitente y destinatario.

### Criterios de aceptación

- Verificación, recuperación e invitación tienen propósitos, expiraciones y validaciones independientes.
- Cambiar contraseña invalida sesiones existentes.
- Solicitar recuperación no revela si la cuenta existe.
- El error de correo no corrompe invitaciones ni cuentas.

## Fase 6 — protección contra abuso y agotamiento

### Objetivo

Evitar que solicitudes válidas pero excesivas agoten CPU, memoria, disco, conexiones o servicios de terceros.

### Cambios previstos

- Rate limiting por IP, usuario, negocio y endpoint según costo.
- `429 Too Many Requests` con `Retry-After`.
- Límites más estrictos para login, registro, OTP, recuperación, invitaciones, imports y consultas externas.
- `http.Server` con `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` e `IdleTimeout`.
- Límites generales de cuerpo, cabeceras, JSON y multipart en proxy y API.
- Pool PostgreSQL con máximos, tiempos de vida y métricas.
- Propagación de `context.Context` y cancelación a DB y proveedores externos.
- Límite de tamaño al decodificar respuestas externas.
- Apagado ordenado de HTTP y DB.
- Configuración de producción que falla cerrada si faltan secretos o valores obligatorios.
- Rotación de logs y alertas de disco/inodos.

### Criterios de aceptación

- Solicitudes lentas, grandes o repetitivas no crecen recursos sin límite.
- Reiniciar el proceso permite terminar solicitudes dentro de una ventana definida.
- Un proveedor externo lento no bloquea indefinidamente la petición.
- Ningún secreto crítico tiene fallback de desarrollo en producción.

## Fase 7 — auditoría y verificación integral

### Objetivo

Demostrar que las fases funcionan juntas y dejar evidencia útil para detectar y responder a incidentes.

### Eventos auditables

- login correcto y fallido;
- bloqueo de cuenta;
- creación, uso y revocación de sesión;
- OTP emitido, reenviado, agotado, expirado o validado, sin guardar su valor;
- solicitud y consumo de recuperación;
- cambio de contraseña;
- cambio de roles y permisos;
- denegación por negocio o permiso;
- importación y cambios de catálogo;
- envío, rebote, queja y throttling de SES.

Cada evento tendrá fecha, actor cuando exista, acción, resultado, request ID y metadatos mínimos. No incluirá contraseña, cookie, token, OTP, secreto, cuerpo sensible ni datos personales innecesarios.

### Verificación final

- Matriz completa de rutas y roles.
- Pruebas IDOR/BOLA entre dos negocios reales de prueba.
- CSRF y atributos de cookie.
- Revocación de sesión capturada.
- OTP expirado, reutilizado, agotado y concurrente.
- Anti-enumeración en login, reenvío y recuperación.
- Fallos simulados y sandbox de SES.
- `go test ./...` y `go vet ./...`.
- Pruebas y build Angular con Node 24/npm 11.
- SCA de Go y npm, escaneo de secretos y análisis estático.
- Revisión de migración, rollback de aplicación e idempotencia.

### Criterios de aceptación

- Todos los controles P0 de aplicación de la capability 05 están cerrados.
- No existen vulnerabilidades críticas o altas conocidas sin excepción explícita.
- Logs permiten investigar eventos sin exponer credenciales.
- La evidencia y resultados quedan registrados en la capability.

## Tamaño relativo y entregas

Las fases se aprobarán e implementarán por separado. No se agruparán todas en un solo cambio.

| Fase | Tamaño relativo | Entrega verificable |
| --- | --- | --- |
| 0 | Pequeño | Inventario, conflictos resueltos y pruebas de línea base |
| 1 | Grande | Autorización completa y catálogos aislados por negocio |
| 2 | Mediano/grande | Sesión revocable y CSRF sin JWT |
| 3 | Mediano | Registro pendiente y OTP probado con doble de correo |
| 4 | Mediano | Adaptador SES y validación en sandbox |
| 5 | Mediano | Recuperación segura e invitaciones por correo |
| 6 | Mediano | Límites, timeouts y configuración fail-closed |
| 7 | Mediano | Auditoría, escaneos y regresión integral |

La fase 1 es la más riesgosa porque combina autorización y migración multiempresa. Si el inventario encuentra productos compartidos entre negocios, se separará la migración de catálogo en una capability propia antes de tocar datos.

## Fuera de alcance de este plan inmediato

- Crear o configurar la instancia Lightsail.
- DNS, certificados, firewall, backups del servidor o monitoreo de infraestructura.
- Aplicación móvil o API para terceros.
- OAuth/OIDC con proveedores sociales.
- Access token y refresh token.
- MFA de segundo factor permanente después del login; el OTP inicial verifica correo, no constituye MFA continua.
- Alta disponibilidad o tolerancia a caída de una instancia.

## Referencias

- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [OWASP Email Validation and Verification Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Email_Validation_and_Verification_Cheat_Sheet.html)
- [OWASP Forgot Password Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
- [Amazon SES: solicitar acceso de producción](https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html)
- [Amazon SES: cuotas](https://docs.aws.amazon.com/ses/latest/dg/quotas.html)
- [AWS SDK for Go v2: configuración y credenciales](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)

## Registro de decisiones

- 2026-09-19: el usuario confirmó que el primer cliente será únicamente Angular bajo el mismo origen.
- 2026-09-19: se eligió sesión opaca revocable y se descartó access/refresh token para esta etapa.
- 2026-09-19: el usuario confirmó que marcas, categorías y unidades pertenecen a cada negocio.
- 2026-09-19: SES y OTP se implementarán después de cerrar rutas, RBAC y sesiones.
- 2026-09-19: el usuario solicitó conservar este plan dentro de la capability; esto no constituye aprobación de implementación.
- 2026-09-20: el usuario aprobó explícitamente la implementación de las fases 1 y 2. El resto del plan permanece sin autorización.
- 2026-09-20: el usuario aprobó explícitamente la fase 3; quedó implementada con proveedor SMTP local Mailpit, sin configurar ni integrar SES.
