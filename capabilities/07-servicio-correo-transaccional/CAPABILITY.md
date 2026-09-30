# Capability 07: Servicio de correo transaccional (SMTP + Amazon SES)

## Estado

En implementación. Fases 1, 2 y 3 implementadas y verificadas el 2026-09-25 (pendiente aceptación del usuario). Fase 4 (SES real, configuración manual) pendiente.

## Aprobación

Fases 1, 2 y 3 aprobadas explícitamente por el usuario el 2026-09-25 mediante la instrucción: “podrás implementar la fase 1, fase 2 y fase 3 por favor — sí será con smtp, sí recuperación de contraseña, logo embebido”. Dominio: `mergemakers.com`; MAIL FROM personalizado: `mail.mergemakers.com`. Región SES pendiente de confirmar.

## Contexto

- La fase 3 de la [Capability 06](../06-preparacion-produccion-segura/PLAN_APLICACION.md) dejó implementado el OTP de verificación con un adaptador SMTP mínimo (`backend/internal/infrastructure/cuenta/smtp_mailer.go`): texto plano, sin autenticación, sin TLS y ubicado dentro del dominio `cuenta`. Solo sirve para Mailpit.
- Las fases 4 y 5 de la Capability 06 planean SES e invitaciones/recuperación por correo, pero asumen un adaptador SES API v2 con el SDK de AWS.
- Referencia usada: `didacticapp-backnestjs/src/mail/` (NestJS). De ahí se toman estas ideas: un contrato de proveedor (`MailProvider`), un servicio con métodos por caso de uso (`sendVerificationCode`, `sendPasswordResetCode`), plantillas que devuelven `subject`, `text` y `html`, escape del HTML, configuración aislada y SMTP como transporte (`nodemailer`).
- Esta capability sustituye, en lo relativo al correo, las fases 4 y 5 de la Capability 06. Cuando se apruebe, se actualizarán `PLAN_APLICACION.md` de la 06 y `PLAN_IMPLEMENTACION.md` para apuntar aquí y conservar una única ubicación canónica.

## Objetivo

Crear un servicio de correo único, limpio y probado, que:

1. Envíe correos HTML y texto alternativo con plantillas propias de Tienda.
2. Use el mismo código en desarrollo (Mailpit) y en producción (Amazon SES por SMTP). Solo cambian las variables de entorno.
3. No acople `cuenta` ni `negocio` a SMTP ni a AWS.
4. Cubra: verificación de correo (OTP), invitación a negocio y, opcionalmente, recuperación de contraseña.
5. Deje documentado cómo configurar SES y cómo solicitar el acceso de producción para el demo.

## Decisión técnica principal

### Transporte: SMTP hacia SES, no SDK de AWS

| Opción | Ventajas | Desventajas |
| --- | --- | --- |
| **SMTP (`go-mail`) contra `email-smtp.<región>.amazonaws.com`** (recomendada) | El mismo adaptador sirve para Mailpit y SES. Sin SDK de AWS ni dependencias pesadas. Replica el patrón probado de didacticapp. Portable a otro proveedor (Resend, Postmark) cambiando solo el `.env`. | Errores menos tipados que la API. El identificador de mensaje se obtiene de la respuesta SMTP. |
| SDK `aws-sdk-go-v2/service/sesv2` | Errores tipados (throttling) y `MessageId` nativo. | Se necesita un segundo adaptador para desarrollo. Agrega unas 10 dependencias transitivas. En Lightsail no hay rol IAM de instancia, así que igual hacen falta llaves estáticas. |

Decisión propuesta: SMTP. Si más adelante se necesitan eventos de SES (SNS) o envíos masivos, se agrega un adaptador `sesv2` detrás del mismo contrato sin tocar los casos de uso.

### Librerías

| Librería | Uso | Motivo |
| --- | --- | --- |
| `github.com/wneessen/go-mail` (última `v0.7.x` al implementar) | Construir el mensaje MIME (`multipart/alternative` HTML + texto), STARTTLS/TLS, `AUTH`, `context`, timeouts, cabeceras seguras y embebido opcional de imágenes (CID). | Mantenida, sin dependencias externas, compatible con Go 1.25 y equivalente a `nodemailer`. Sustituye el ensamblado manual de cabeceras de `net/smtp`, que es frágil y no hace MIME. |
| `html/template` (stdlib) | Plantillas HTML. | Escapa automáticamente. Sustituye el `escapeHtml` manual de la referencia. |
| `text/template` (stdlib) | Versión en texto plano. | Accesibilidad, clientes sin HTML y filtros antispam. |
| `embed` (stdlib) | Incluir plantillas (y logo opcional) dentro del binario. | La imagen Docker no depende de archivos sueltos ni de `process.cwd()`. |

No se agrega nada en frontend.

## Alcance

- Nuevo dominio `correo` en `application` e `infrastructure` (sin `domain` ni `interfaces/http`, porque no expone endpoints).
- Adaptador SMTP con autenticación y política TLS configurable.
- Plantillas: base común, verificación OTP, invitación y recuperación (esta última solo si se aprueba la fase 3).
- Migrar el OTP actual al nuevo servicio y eliminar `infrastructure/cuenta/smtp_mailer.go`.
- Enviar el enlace de invitación por correo, conservando el botón de copiar enlace.
- Validación estricta de la configuración en `APP_ENV=production`.
- Guía de configuración de SES y del acceso de producción (sección final).

## Fuera de alcance

- Cola u outbox. El envío es síncrono con timeout, como decidió la fase 4 de la 06, para no guardar OTP en claro.
- Webhooks SNS de rebotes y quejas. Para el demo basta la lista de supresión de cuenta de SES, que viene activa por defecto.
- Correos de marketing, boletines o listas.
- Adjuntos de usuario (PDF de tickets, facturas).
- Despliegue en Lightsail: sigue en la Capability 06, fases 4 y 7.
- Rate limiting general: sigue en la fase 6 de la 06. Aquí solo se reutilizan los límites de OTP ya existentes.

## Arquitectura

```text
backend/internal/
├── application/
│   ├── correo/
│   │   ├── mensaje.go            # Mensaje{Para, Asunto, HTML, Texto, ReplyTo}
│   │   ├── transporte.go         # interface Transporte { Enviar(ctx, Mensaje) (Resultado, error) }
│   │   ├── plantillas.go         # interface Renderizador { Render(nombre, datos) (asunto, html, texto, error) }
│   │   ├── servicio.go           # Servicio: EnviarCodigoVerificacion, EnviarInvitacion, EnviarRecuperacion
│   │   └── servicio_test.go      # con Transporte y Renderizador falsos
│   ├── cuenta/
│   │   └── verification_service.go   # SIN CAMBIOS de contrato: sigue usando VerificationMailer
│   └── negocio/
│       └── invitacion_service.go     # nueva interfaz local InvitacionMailer
└── infrastructure/
    └── correo/
        ├── smtp_transporte.go        # go-mail; implementa correo.Transporte
        ├── smtp_transporte_test.go   # servidor SMTP falso en memoria (net.Listener)
        ├── renderizador.go           # html/template + text/template + embed
        ├── renderizador_test.go      # golden files y escape
        └── plantillas/
            ├── base.html.tmpl
            ├── verificacion.html.tmpl / verificacion.txt.tmpl
            ├── invitacion.html.tmpl   / invitacion.txt.tmpl
            └── recuperacion.html.tmpl / recuperacion.txt.tmpl
```

### Dirección de dependencias (conforme a `AGENTS.md`)

- `application/correo` no importa `go-mail` ni AWS. Solo define contratos y orquesta.
- `infrastructure/correo` implementa `Transporte` y `Renderizador`.
- `cuenta` y `negocio` **no importan** `correo`. Cada uno declara su propia interfaz pequeña:
  - `cuenta.VerificationMailer.SendVerificationOTP(ctx, recipient, code, expiresAt)` ya existe.
  - `negocio.InvitacionMailer.EnviarInvitacion(ctx, para, nombreNegocio, nombreInvitador, enlace, expiraEn)` es nueva.
- `correo.Servicio` cumple ambas interfaces de forma estructural. Se conectan en `cmd/api/main.go`, igual que hoy con `NewDevelopmentSMTPMailer`.

### Contratos (conceptuales)

```go
// application/correo
type Mensaje struct {
    Para    string
    Asunto  string
    HTML    string
    Texto   string
    ReplyTo string // opcional
}

type Resultado struct{ IDMensaje string }

type Transporte interface {
    Enviar(ctx context.Context, m Mensaje) (Resultado, error)
}

type Renderizador interface {
    Render(plantilla string, datos any) (asunto, html, texto string, err error)
}

// Métodos del Servicio (uno por caso de uso, como MailService de la referencia)
func (s *Servicio) SendVerificationOTP(ctx, para, codigo string, expiraEn time.Time) error
func (s *Servicio) EnviarInvitacion(ctx, para, negocio, invitador, enlace string, expiraEn time.Time) error
func (s *Servicio) EnviarRecuperacion(ctx, para, enlace string, expiraEn time.Time) error
```

### Cómo funciona un envío

1. El caso de uso (por ejemplo, `VerificationService.Register`) persiste primero el estado: la cuenta pendiente y el desafío con el HMAC del OTP.
2. Llama a `mailer.SendVerificationOTP(ctx, correo, codigo, expira)`.
3. `correo.Servicio` arma los datos de la plantilla (código, minutos, `FRONTEND_URL`, nombre de la app) y llama a `Renderizador.Render("verificacion", datos)`, que devuelve asunto, HTML y texto.
4. `Transporte.Enviar` usa `go-mail`: `From` = `MAIL_FROM_NAME <MAIL_FROM>`, más `To`, `Subject`, cuerpo alternativo, cabecera `X-SES-CONFIGURATION-SET` si está configurada, `DialAndSendWithContext` con timeout `SMTP_TIMEOUT`, STARTTLS obligatorio en producción y `AUTH PLAIN/LOGIN`.
5. Se registra en log `evento=correo_enviado plantilla=verificacion id_mensaje=... latencia_ms=...` **sin** destinatario completo (solo enmascarado), OTP, enlace ni cuerpo.
6. Si falla: se registra `correo_fallido` con la clase de error (timeout, auth, rechazo 5xx) y se devuelve error. El caso de uso conserva el estado recuperable y el usuario puede pedir reenvío. Esta regla ya existe en la fase 3.

### Plantillas

- `base.html.tmpl`: layout en tablas, 600 px, estilos en línea, colores del sistema visual de Tienda (Capability 05), nombre de la app en texto y pie. Sin fuentes remotas ni trackers.
- Logo: por defecto, solo el nombre en texto. Si se aprueba, se usa un PNG embebido con `embed` y adjunto inline por CID, como en la referencia. Nunca una URL remota, porque los clientes la bloquean.
- Cada plantilla define el bloque `asunto` y el cuerpo. Todo dato dinámico pasa por `html/template`.
- Enlaces: se construyen solo a partir de `FRONTEND_URL` validada, nunca desde el `Host` de la petición.
- Idioma: español (es-MX).

## Tablas: ¿alcanzan las actuales?

Conclusión: **sí. No se requieren migraciones** para el alcance propuesto.

| Necesidad | Tabla/campo actual | ¿Alcanza? | Nota |
| --- | --- | --- | --- |
| Destinatario de verificación | `usuarios.correo` varchar(255), `estado`, `correo_verificado_en` | Sí | Ya lo usa la fase 3. |
| OTP de verificación | `desafios_autenticacion` (`proposito`, `hash_otp` char(64), `intentos_fallidos`, `direccion_ip`, `ultimo_envio_en`, `expira_en`, `usado_en`) | Sí | Sin cambios. |
| Recuperación de contraseña | `desafios_autenticacion` con `proposito = 'recuperacion_contrasena'` | Sí | El enlace lleva `?desafio=<id>&token=<aleatorio 32 bytes>`. Se guarda `HMAC-SHA256(secreto, id‖token)` en `hash_otp` (64 hex). Se busca por `id` (PK) y se compara en tiempo constante, así que no hace falta índice nuevo. `varchar(50)` admite el propósito. |
| Revocar sesiones al recuperar | `usuarios.contrasena_cambiada_en`, `sesiones_usuario.revocado_en` | Sí | Ya existen desde la fase 2. |
| Invitación por correo | `invitaciones_negocio` (`correo`, `hash_token`, `expira_en`, `estado`) | Sí | El token en claro solo existe dentro de `Crear`, así que el correo se envía en esa misma petición. |
| Datos para la plantilla de invitación | `negocios.nombre_comercial`, `empleados.nombre` | Sí | Se leen en el mismo servicio `negocio`. |
| Registro de envíos y rebotes | — | No hace falta | Se usan logs estructurados y la lista de supresión de SES. Una tabla `envios_correo` o una columna `usuarios.correo_no_entregable_en` quedan para cuando se implementen webhooks SNS. |

Opcional, **no incluido** salvo que se decida: `invitaciones_negocio.correo_enviado_en timestamptz`, para mostrar en la interfaz "enviado por correo" frente a "solo enlace".

## Variables de entorno

Estado actual en `backend/.env.example`: `SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, `SMTP_TIMEOUT`. Propuesta final:

```dotenv
# ---------------------------------------------------------------
# Correo transaccional
# Desarrollo: Mailpit (bandeja en http://localhost:8025), sin auth ni TLS.
# Producción: Amazon SES por SMTP (ver capabilities/07-servicio-correo-transaccional).
# ---------------------------------------------------------------
SMTP_HOST=localhost
SMTP_PORT=1025
# none | starttls | tls   (producción: starttls con puerto 587, o tls con 465)
SMTP_TLS=none
# Vacíos en desarrollo. En producción: credenciales SMTP de SES (NO son las access keys de IAM).
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_TIMEOUT=10s
# Remitente. En producción debe pertenecer a un dominio verificado en SES.
MAIL_FROM=no-reply@tienda.local
MAIL_FROM_NAME=Tienda
# Opcional: buzón real para respuestas (soporte@mergemakers.com)
MAIL_REPLY_TO=
# Opcional: configuration set de SES (métricas/eventos). Vacío = no se envía la cabecera.
SES_CONFIGURATION_SET=
```

Renombres: `SMTP_FROM` pasa a `MAIL_FROM`. Durante una versión se acepta `SMTP_FROM` como alias, con aviso en log, para no romper `docker-compose.yml`. `FRONTEND_URL` ya existe y se reutiliza para los enlaces.

### De dónde sale cada valor

| Variable | Desarrollo | Producción (SES) |
| --- | --- | --- |
| `SMTP_HOST` | `localhost` (o `mailpit` en Compose) | Consola SES → *SMTP settings* → *SMTP endpoint*, por ejemplo `email-smtp.us-east-1.amazonaws.com` (la región donde verificaste el dominio). |
| `SMTP_PORT` | `1025` | `587` (STARTTLS). Alternativa: `465` con `SMTP_TLS=tls`. **No usar 25**: Lightsail y EC2 lo bloquean. |
| `SMTP_TLS` | `none` | `starttls` |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | vacíos | Consola SES → *SMTP settings* → **Create SMTP credentials**. Se genera un usuario IAM y se muestran usuario y contraseña SMTP **una sola vez** (descargar CSV). |
| `SMTP_TIMEOUT` | `10s` | `10s` |
| `MAIL_FROM` | `no-reply@tienda.local` | `no-reply@mergemakers.com` |
| `MAIL_FROM_NAME` | `Tienda` | Nombre comercial del demo. |
| `MAIL_REPLY_TO` | vacío | Buzón que sí leas; a AWS le gusta verlo en la solicitud de producción. |
| `SES_CONFIGURATION_SET` | vacío | Nombre del configuration set creado en SES (opcional). |

### Validación al arrancar en `APP_ENV=production`

El backend **no arranca** si:

- `SMTP_HOST` es `localhost`, `127.0.0.1` o `mailpit`;
- `SMTP_TLS=none`;
- `SMTP_USERNAME` o `SMTP_PASSWORD` están vacíos;
- `MAIL_FROM` está vacío, es inválido o termina en `.local`;
- `FRONTEND_URL` no es `https://`.

Los secretos nunca se imprimen en logs de error; solo el nombre de la variable.

## Fases

### Fase 1 — Servicio de correo y migración del OTP

- Agregar `go-mail` a `go.mod`.
- Crear `application/correo` e `infrastructure/correo` (contratos, servicio, transporte SMTP, renderizador, plantillas base y verificación).
- Ampliar `config.go` con las variables nuevas, el alias `SMTP_FROM` y las validaciones de producción.
- Conectar en `main.go` `correo.Servicio` como `VerificationMailer`.
- Eliminar `infrastructure/cuenta/smtp_mailer.go`.
- Actualizar `backend/.env.example` y `docker-compose.yml` (`MAIL_FROM`, `SMTP_TLS=none`).
- Resultado: el mismo OTP de hoy llega a Mailpit con HTML y texto.

### Fase 2 — Invitación por correo

- Interfaz `InvitacionMailer` en `application/negocio`.
- `InvitacionService.Crear` persiste la invitación, construye el enlace desde `FRONTEND_URL` e intenta el envío.
- Si el envío falla, la invitación **no se revierte**: la respuesta incluye `correo_enviado: false` y el enlace copiable sigue funcionando.
- Frontend (`invitaciones.component`): el mensaje cambia a "Invitación enviada a x***@dominio" o "No se pudo enviar el correo; copia el enlace".
- Reenviar una invitación queda fuera de alcance: requeriría regenerar el token.

### Fase 3 — Recuperación de contraseña

- `POST /api/v1/auth/solicitar-recuperacion` responde siempre `202` de forma uniforme.
- `POST /api/v1/auth/restablecer-contrasena` recibe desafío, token y contraseña nueva. Aplica la política de contraseña de registro, marca `usado_en`, actualiza `contrasena_cambiada_en`, revoca todas las sesiones y envía una notificación "tu contraseña cambió".
- Vigencia de 30 minutos, un solo uso y los mismos límites por correo/IP que el OTP.
- Frontend: pantallas "Olvidé mi contraseña" (`/recuperar-contrasena`) y "Nueva contraseña" (`/restablecer-contrasena`).
- El enlace lleva desafío y token en el **fragmento** (`#desafio=…&token=…`): el navegador no lo envía al servidor, a logs de proxy ni en `Referer`. La pantalla lo lee y lo retira de la URL con `history.replaceState`. Esto sustituye la cabecera `Referrer-Policy` prevista.
- Es la fase 5 de la 06; aquí queda como decisión explícita.

### Fase 4 — SES real (configuración manual, sin código)

Pasos en la guía de abajo. Se verifica con Mailpit apagado y SES en sandbox antes de solicitar producción.

## Reglas

- Nunca registrar OTP, tokens, enlaces completos, cuerpo del correo, `SMTP_PASSWORD` ni el destinatario completo.
- Timeout duro por envío (`SMTP_TIMEOUT`) y sin reintentos automáticos dentro de la petición, para no duplicar OTP. El reintento es el reenvío que hace el usuario.
- Errores hacia el cliente: genéricos y en español, sin detalles de SMTP o AWS.
- Validar `\r\n` en destinatario y asunto; `go-mail` también lo protege.
- Un solo remitente por ambiente; el `From` no se toma de la entrada del usuario.
- Las plantillas no cargan recursos remotos.

## Pruebas

### Automatizadas (sin red ni AWS)

- `application/correo`: con `Transporte` y `Renderizador` falsos, cada método arma el destinatario, asunto y datos correctos; los errores del transporte se propagan sin filtrar detalles.
- `infrastructure/correo/renderizador`:
  - golden files de cada plantilla (HTML y texto);
  - un nombre `<script>` sale escapado;
  - los enlaces usan `FRONTEND_URL`;
  - una plantilla inexistente devuelve error.
- `infrastructure/correo/smtp_transporte`: servidor SMTP falso en `net.Listener` que comprueba `MAIL FROM`, `RCPT TO`, `multipart/alternative`, la cabecera del configuration set cuando existe, el timeout (servidor que no responde), el rechazo 5xx y que un `\r\n` en el destinatario se rechace.
- `config`: producción falla con host local, TLS `none`, credenciales vacías, `MAIL_FROM` `.local` o `FRONTEND_URL` http; el alias `SMTP_FROM` funciona.
- Regresión: las pruebas existentes de `verification_service_test.go`, `auth_handler_test.go` e `invitacion_service_test.go` siguen verdes.
- `invitacion_service`: un fallo de correo conserva la invitación y devuelve `correo_enviado=false`.
- Logs: una prueba captura la salida y confirma que no aparecen el OTP, el token ni la contraseña.

### Manuales

- Compose con Mailpit: registro, OTP visible en `http://localhost:8025`, HTML correcto en escritorio y móvil, versión de texto presente.
- SES sandbox: envío a una dirección verificada propia y a `success@simulator.amazonses.com`, `bounce@simulator.amazonses.com` y `complaint@simulator.amazonses.com`.
- Revisar cabeceras en Gmail ("Mostrar original"): `SPF: PASS`, `DKIM: PASS`, `DMARC: PASS`.

### Comandos

```bash
cd backend && go test ./... && go vet ./...
cd frontend && npm test -- --watch=false && npm run build   # solo si se toca la fase 2/3 del frontend
```

## Criterios de aceptación

- Un único adaptador SMTP sirve para Mailpit y SES; cambiar de uno a otro solo requiere el `.env`.
- `cuenta` y `negocio` no importan `correo`, `go-mail` ni AWS.
- Todo correo lleva HTML y texto, con datos escapados y enlaces basados en `FRONTEND_URL`.
- Producción no arranca con configuración de correo insegura o incompleta.
- Un fallo de envío nunca activa cuentas ni corrompe invitaciones, y siempre queda un camino de reintento.
- Ningún secreto, OTP ni token aparece en logs.
- Correo recibido en Gmail con SPF, DKIM y DMARC en `PASS`.

## Guía: configurar Amazon SES para el demo

> Requisito previo: tener un **dominio propio** con acceso al DNS (Lightsail DNS, Route 53, Cloudflare o el registrador). No se puede salir del sandbox usando solo un `@gmail.com` como remitente.

### 1. Elegir región

Usar la misma región que la instancia de Lightsail (por ejemplo `us-east-1`). Identidades, credenciales SMTP, cuotas y acceso de producción **son por región**: todo lo que sigue se hace en esa región.

### 2. Verificar el dominio (identidad)

1. Consola AWS → **Amazon SES** → *Configuration* → **Identities** → *Create identity*.
2. Tipo **Domain** → escribir `mergemakers.com`.
3. En *Advanced DKIM settings*: **Easy DKIM**, `RSA_2048_BIT`, *DKIM signatures* activado.
4. Activar **Use a custom MAIL FROM domain** → `mail.mergemakers.com` y *Behavior on MX failure* → *Use default MAIL FROM*.
   `mail.mergemakers.com` **sí se usa**, pero no es el remitente visible: es el dominio del *Return-Path* (sobre SMTP), a donde llegan los rebotes. Permite que SPF quede alineado con `mergemakers.com` y que DMARC pase por SPF y no solo por DKIM. El remitente que ve el usuario sigue siendo `no-reply@mergemakers.com`. No hay que crear buzón ni variable de entorno para él; solo los registros DNS de abajo.
5. Crear. SES muestra los registros DNS que hay que publicar:
   - 3 registros **CNAME** de DKIM: `xxxx._domainkey.mergemakers.com` → `xxxx.dkim.amazonses.com`.
   - MAIL FROM: **MX** `mail.mergemakers.com` → `10 feedback-smtp.<región>.amazonses.com`.
   - MAIL FROM: **TXT** `mail.mergemakers.com` → `"v=spf1 include:amazonses.com ~all"`.
6. Agregar DMARC manualmente: **TXT** `_dmarc.mergemakers.com` → `"v=DMARC1; p=none; rua=mailto:dmarc@mergemakers.com"`. Empezar con `p=none` y subir a `quarantine` cuando todo pase.
7. Esperar a que la identidad pase a **Verified** (de minutos a 72 h).

### 3. Probar en sandbox

Toda cuenta nueva empieza en sandbox: solo puede enviar **a** direcciones o dominios verificados, con 200 correos cada 24 h y 1 por segundo.

1. *Identities* → *Create identity* → **Email address** → tu correo personal → confirmar el enlace que llega.
2. Enviar desde el backend a esa dirección y a los buzones del *mailbox simulator*.

### 4. Crear credenciales SMTP (valores para `.env`)

1. SES → **SMTP settings** → anotar el *SMTP endpoint* (`SMTP_HOST`) y los puertos (`587`).
2. **Create SMTP credentials** → se abre IAM y se crea un usuario, por ejemplo `ses-smtp-tienda-prod`.
3. Descargar el CSV: *SMTP user name* va a `SMTP_USERNAME` y *SMTP password* a `SMTP_PASSWORD`. **No se vuelve a mostrar.**
4. Endurecer la política del usuario IAM. Reemplazar la política por defecto por una limitada al remitente:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [{
       "Effect": "Allow",
       "Action": "ses:SendRawEmail",
       "Resource": "*",
       "Condition": { "StringEquals": { "ses:FromAddress": "no-reply@mergemakers.com" } }
     }]
   }
   ```

5. Las credenciales SMTP **no** son las *access keys* de IAM. Si generaste access keys de la cuenta raíz o de un administrador (por ejemplo en `credenciales-aws.txt`), no las uses en la app. Si alguna vez salieron del equipo, rótalas.

### 5. Configuration set (opcional, recomendado)

1. SES → *Configuration sets* → *Create* → `tienda-transaccional`.
2. *Suppression list*: usar la configuración de cuenta (rebotes y quejas).
3. Opcional: *Event destinations* hacia CloudWatch para ver entregas, rebotes y quejas.
4. Poner el nombre en `SES_CONFIGURATION_SET`.

Verificar además en *Account dashboard* → **Suppression list settings** que la supresión a nivel cuenta esté activa para *Bounce* y *Complaint*.

### 6. Solicitar acceso de producción

Consola SES → **Account dashboard** → **Request production access** (también en *Get set up*).

Antes de enviar la solicitud:

- El dominio ya está **Verified**, con DKIM en PASS.
- El demo está publicado en una URL pública con HTTPS (Capability 06, fases 4 y 7), con página de inicio que explique el producto y, idealmente, aviso de privacidad y términos.
- Ya probaste el envío en sandbox.

Campos del formulario:

- **Mail type**: `Transactional`.
- **Website URL**: `https://mergemakers.com`.
- **Additional contacts**: tu correo.
- **Preferred contact language**: English o Spanish.
- **Use case description**: redactarlo concreto. Borrador en inglés, que AWS procesa más rápido:

  > Tienda is a point-of-sale and inventory web application for small retail businesses in Mexico (currently in public demo at https://mergemakers.com).
  > We only send transactional emails triggered by a user action: (1) 6-digit email verification codes at sign-up, (2) invitations when a business owner invites an employee by email, and (3) password reset links. We never send marketing or bulk email and never use purchased lists.
  > Recipients are users who registered themselves on our site or whose employer explicitly entered their address. Every sign-up requires email verification before the account is activated.
  > Bounces and complaints: we rely on the SES account-level suppression list (bounces and complaints enabled) and a configuration set; addresses on the list are not retried. Sending is rate-limited per address and per IP (max 5 codes/hour).
  > Expected volume: under 500 emails/day during the demo, peak 2 emails/second. Sender: no-reply@mergemakers.com, with SPF, DKIM and DMARC configured on mergemakers.com.

- **Acknowledge** que cumplirás la política de uso aceptable.

AWS responde en unas **24 h**, a veces pide más detalles; se contesta en el mismo caso de Support. Al aprobar suben las cuotas (típicamente 50 000/día y 14/s) y se puede enviar a cualquier dirección. Si lo rechazan, suele deberse a un sitio sin contenido, una descripción vaga o una identidad sin DKIM.

### 7. Cargar valores en el servidor (Lightsail)

- Poner las variables en el `.env` del servidor (fuera del repositorio, `chmod 600`) o en el mecanismo de secretos que defina la 06. **Nunca** en la imagen Docker ni en el frontend.
- Lightsail permite salida por `587`; no se abre ningún puerto de entrada.
- Arrancar con `APP_ENV=production`: si falta algo, el backend falla al iniciar, como está diseñado.
- Prueba final: registrar una cuenta real, recibir el OTP y revisar SPF, DKIM y DMARC en Gmail.

## Decisiones tomadas

1. Transporte SMTP con `go-mail`.
2. La recuperación de contraseña se implementa aquí (fase 3).
3. Logo PNG embebido por CID.
4. Sin columna `invitaciones_negocio.correo_enviado_en`; la respuesta de creación informa `correo_enviado`.
5. Dominio `mergemakers.com`; remitente `no-reply@mergemakers.com`; MAIL FROM personalizado `mail.mergemakers.com`. Región: pendiente.

## Decisiones y seguimiento

- 2026-09-25: borrador creado a partir de la referencia `didacticapp-backnestjs/src/mail/` y del estado de la fase 3 de la Capability 06. No se modificó código.
- 2026-09-25: fases 1, 2 y 3 aprobadas por el usuario e implementadas. Capability registrada en `PLAN_IMPLEMENTACION.md`; `PLAN_APLICACION.md` de la 06 apunta aquí para las partes de correo de sus fases 4 y 5.

### Implementación real (2026-09-25)

Backend:

- `application/correo/correo.go`: contratos `Transporte` (devuelve `Message-ID`) y `Renderizador`, datos por plantilla y `Servicio` con `SendVerificationOTP`, `EnviarInvitacion`, `SendPasswordReset` y `SendPasswordChanged`. Registra `evento=correo_enviado|correo_fallido` con destinatario enmascarado y latencia; nunca el cuerpo, OTP, token ni enlace. `ReplyTo` quedó en la configuración del transporte y no en `Mensaje`.
- `infrastructure/correo/`: `renderizador.go` (html/template + text/template + `embed`; compila todo al arrancar), `smtp_transporte.go` (`go-mail` v0.8.1; `none|starttls|tls`, `AUTH PLAIN` solo con usuario, timeout duro, cabecera `X-SES-CONFIGURATION-SET` opcional, `Message-ID` con el dominio del remitente, errores clasificados sin datos personales), `plantillas/` (base, 4 plantillas HTML+texto y `logo.png` embebido por CID `logo@tienda`).
- El logo es un **placeholder** generado para Tienda (440×96 px). Se reemplaza sustituyendo `backend/internal/infrastructure/correo/plantillas/logo.png` por otro PNG del mismo tamaño.
- `cuenta`: `PasswordResetService` y `PasswordResetHandler`, rutas `POST /api/v1/auth/solicitar-recuperacion` y `POST /api/v1/auth/restablecer-contrasena`; propósito `recuperacion_contrasena` en `desafios_autenticacion`; `ResetPasswordWithChallenge` consume el desafío, cambia la contraseña, reinicia bloqueos y revoca sesiones en una sola transacción con bloqueo de fila. La contraseña nueva exige de 8 caracteres a 72 bytes (límite de bcrypt).
- `negocio`: interfaz `InvitacionMailer`; `InvitacionCreada.correo_enviado`. Un fallo de correo no revierte la invitación.
- `config`: `SMTP_TLS`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `MAIL_FROM` (con alias `SMTP_FROM`), `MAIL_FROM_NAME`, `MAIL_REPLY_TO`, `SES_CONFIGURATION_SET`, `PASSWORD_RESET_TTL` y validación de producción. Se eliminó `infrastructure/cuenta/smtp_mailer.go`.
- `go get go-mail@v0.8.1` subió como requisito mínimo `golang.org/x/crypto` 0.38→0.54, `x/text` 0.25→0.40, `x/net`, `x/sync` y `x/sys`; `go.mod` pasó de `go 1.25` a `go 1.25.0`. Compatible con `golang:1.25-alpine` del Dockerfile.
- Sin migraciones de base de datos.

Frontend: `AuthService.requestPasswordReset/resetPassword`, componentes `recover-password` y `reset-password`, enlace "¿Olvidaste tu contraseña?" en login y aviso en invitaciones según `correo_enviado`.

Configuración: `backend/.env.example` y `docker-compose.yml` (variables parametrizadas, Mailpit por defecto).

### Verificaciones ejecutadas (2026-09-25)

- `go vet ./...` y `go test -count=1 ./...`: todo en verde. Pruebas nuevas: servicio de correo (datos, enlaces, logs sin secretos), renderizador (4 plantillas, escape XSS, `javascript:` bloqueado, asunto sin saltos de línea), transporte contra servidor SMTP falso en memoria (multipart/related + alternative, logo inline, Reply-To, configuration set, rechazo 550, timeout, inyección de cabeceras, error de conexión), configuración (valores de desarrollo, alias, rechazos en producción), recuperación (emisión, uniformidad, límites, un solo uso, expiración, agotamiento, propósito cruzado, política de contraseña), handler HTTP e invitación con y sin fallo de correo.
- Integración PostgreSQL (`SECURITY_TEST_DATABASE_URL` → `tienda_security_test` en el puerto 5434): dos restablecimientos concurrentes producen un solo éxito, se revocan sesiones, se reinician intentos fallidos y un desafío de recuperación no activa cuentas. En verde.
- Frontend con Node 24.21.0: `ng test --watch=false` 35 archivos / 122 pruebas en verde; `ng build` correcto (salida en scratchpad porque `frontend/dist/` pertenece a root). Solo advertencias de presupuesto previas.
- Manual con Mailpit (`127.0.0.1:1025`): se enviaron los 4 correos con el transporte real. El logo llega inline (`logo@tienda`, `image/png`), la parte de texto está presente y el HTML check de Mailpit da ≈91–93 % de compatibilidad. Revisión visual con Chromium headless a 760 px y 375 px correcta. La prueba detectó que la invitación mostraba "6 días" por truncamiento; se corrigió redondeando y se reforzó la prueba. Los mensajes de prueba se borraron de Mailpit.

### Pendientes conocidos

- Fase 4: configurar SES (guía arriba) con la región definitiva y probar en sandbox.
- El límite por IP de solicitudes de recuperación solo cuenta desafíos emitidos; correos inexistentes no suman. El rate limiting general queda en la fase 6 de la Capability 06.
- Webhooks SNS de rebotes/quejas fuera de alcance; se usa la lista de supresión de SES.
- No se revisaron en navegador las pantallas nuevas del frontend; están cubiertas por pruebas unitarias.
- `backend/.env.example` sigue repitiendo `APP_PORT`, `FRONTEND_PORT` y `FRONTEND_URL` al final; no se tocó porque puede ser una configuración local intencional.
