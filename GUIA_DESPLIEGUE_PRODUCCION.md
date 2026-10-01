# Guía de despliegue de Stockion en producción

Esta guía resume la preparación y el despliegue de Stockion en AWS Lightsail usando Docker Compose, Cloudflare Tunnel y Amazon SES. No contiene credenciales reales.

## 1. Arquitectura

```text
Usuario
  |
  v
Cloudflare: DNS, HTTPS, CDN, WAF y mitigación DDoS
  |
  v
Cloudflare Tunnel: conexión cifrada saliente desde Lightsail
  |
  v
Caddy interno: rutas, límites y cabeceras
  |-------------------|
  v                   v
Frontend Angular      API Go
                          |
                          v
                  PostgreSQL privado
```

Funciones principales:

- **Cloudflare Tunnel:** evita publicar `80/443` y oculta el origen.
- **Caddy:** enruta `/api/*` al backend y el resto al frontend.
- **Frontend:** sirve archivos estáticos Angular como usuario no-root.
- **Backend:** procesa autenticación, negocio y correo transaccional.
- **PostgreSQL:** persiste datos en volumen Docker, sin puerto público.
- **Amazon SES:** envía OTP, recuperación, invitaciones y avisos.

Todos los cambios en GitHub, Cloudflare, AWS y SSH los ejecuta el propietario. Nunca se comparten contraseñas, tokens, credenciales SMTP ni llaves privadas.

## 2. Estado actual

### Completado

- Lightsail con Ubuntu 24.04 LTS, IP estática y Docker Compose.
- UFW activo: sólo `22/tcp` entrante; salida permitida.
- SSH: root y contraseña desactivados; llaves activas; máximo 3 intentos.
- Actualizaciones automáticas de seguridad activas.
- Rama `production` creada desde `thrs` y publicada.
- Compose de producción, Caddy y frontend no-root implementados.
- Dominio e identidad SES verificados en `us-east-2`.
- Registros DKIM, SPF, MAIL FROM y DMARC configurados en Cloudflare.
- Credenciales SMTP antiguas rotadas; nuevas descargadas.
- Tunnel `stockion-production` creado.
- Token del Tunnel guardado fuera del repositorio.
- Conector probado: DNS, QUIC, HTTP/2 y API de Cloudflare correctos.

### Pendiente

- Ilián debe agregar deploy key GitHub de sólo lectura.
- Clonar rama `production` en Lightsail.
- Crear `.env.production` con secretos nuevos.
- Construir y levantar el stack.
- Crear Public Hostname de Cloudflare y sustituir registro A anterior.
- Probar flujos completos.
- Solicitar salida del sandbox de SES.

## 3. Archivos importantes

- `compose.production.yaml`: servicios, redes, límites y healthchecks.
- `.env.production.example`: variables requeridas sin secretos.
- `deploy/Caddyfile`: gateway HTTP interno.
- `deploy/Dockerfile.caddy`: Caddy no-root y sin capacidades innecesarias.
- `frontend/Dockerfile.production`: imagen Angular/Nginx no-root.
- `frontend/nginx.production.conf`: servicio estático interno.

No usar `docker-compose.yml` para producción: incluye puertos y herramientas de desarrollo.

## 4. Deploy key de GitHub

La llave privada debe existir únicamente en Lightsail:

```bash
mkdir -p /home/thrs/.ssh
chmod 700 /home/thrs/.ssh
ssh-keygen -t ed25519 -C "stockion-lightsail-deploy" -f /home/thrs/.ssh/stockion_github -N ""
```

Mostrar únicamente llave pública:

```bash
cat /home/thrs/.ssh/stockion_github.pub
```

El propietario de `Ilimm9/Tienda` debe agregarla en:

```text
Settings > Deploy keys > Add deploy key
Title: stockion-lightsail
Allow write access: desactivado
```

Nunca copiar ni mostrar `/home/thrs/.ssh/stockion_github`.

Después de recibir permiso, crear `/home/thrs/.ssh/config` con:

```sshconfig
Host github-stockion
    HostName github.com
    User git
    IdentityFile /home/thrs/.ssh/stockion_github
    IdentitiesOnly yes
```

Aplicar permisos:

```bash
chmod 600 /home/thrs/.ssh/config
ssh -T github-stockion
```

Antes de aceptar por primera vez la identidad de GitHub, comparar la huella mostrada con las huellas publicadas en la documentación oficial de GitHub.

Clonar sólo producción:

```bash
mkdir -p /home/thrs/apps
git clone --branch production --single-branch git@github-stockion:Ilimm9/Tienda.git /home/thrs/apps/stockion
cd /home/thrs/apps/stockion
```

## 5. Secretos y variables

El token Cloudflare está en:

```text
/home/thrs/.config/stockion/cloudflare-tunnel-token
```

La imagen oficial `cloudflared` usa UID/GID `65532`:

```bash
sudo chown 65532:65532 /home/thrs/.config/stockion/cloudflare-tunnel-token
sudo chmod 600 /home/thrs/.config/stockion/cloudflare-tunnel-token
sudo stat -c '%a %u:%g %n' /home/thrs/.config/stockion/cloudflare-tunnel-token
```

Resultado esperado: `600 65532:65532`.

Crear configuración de producción:

```bash
cd /home/thrs/apps/stockion
cp .env.production.example .env.production
chmod 600 .env.production
nano .env.production
```

Valores principales:

```dotenv
APP_ENV=production
FRONTEND_URL=https://stockion.mergemakers.com
DB_USER=stockion
DB_PASSWORD=<contraseña nueva y única>
DB_NAME=stockion
OTP_HMAC_SECRET=<secreto aleatorio de al menos 32 caracteres>

SMTP_HOST=email-smtp.us-east-2.amazonaws.com
SMTP_PORT=587
SMTP_TLS=starttls
SMTP_USERNAME=<SMTP Username nuevo>
SMTP_PASSWORD=<SMTP Password nuevo>
MAIL_FROM=no-reply@mergemakers.com
MAIL_FROM_NAME=Stockion

CLOUDFLARE_TUNNEL_TOKEN_FILE=/home/thrs/.config/stockion/cloudflare-tunnel-token
```

Generar valores aleatorios sin reutilizarlos:

```bash
openssl rand -base64 36
openssl rand -base64 48
```

No usar credenciales locales, antiguas o enviadas por correo. No almacenar el CSV de SES dentro del repositorio.

## 6. Validación y arranque

Desde `/home/thrs/apps/stockion`:

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml config --quiet
sudo docker compose --env-file .env.production -f compose.production.yaml build --pull
sudo docker compose --env-file .env.production -f compose.production.yaml up -d
sudo docker compose --env-file .env.production -f compose.production.yaml ps
```

Todos los servicios deben aparecer `Up`; PostgreSQL, backend, frontend y gateway deben terminar como `healthy`.

Logs iniciales:

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=100
```

No publicar logs que contengan información sensible.

## 7. Cloudflare Tunnel y DNS

Cuando el stack esté saludable:

1. Cloudflare > **Networking > Tunnels**.
2. Abrir `stockion-production`.
3. Agregar **Public Hostname**.
4. Subdomain: `stockion`.
5. Domain: `mergemakers.com`.
6. Service type: `HTTP`.
7. URL: `gateway:8080`.
8. Guardar.

El registro A anterior `stockion.mergemakers.com -> 52.7.215.247` entra en conflicto y debe eliminarse al crear el hostname del Tunnel. Cloudflare creará la ruta DNS correspondiente.

No abrir `80/443` en UFW ni Lightsail. Tunnel usa conexiones salientes. Conservar únicamente SSH `22`.

## 8. Verificación pública

```bash
curl -fsS https://stockion.mergemakers.com/api/v1/health
curl -fsSI https://stockion.mergemakers.com/
sudo ss -lntp
sudo docker compose --env-file .env.production -f compose.production.yaml ps
```

Resultado esperado:

- API responde `{"estado":"ok"}`.
- HTTPS válido.
- Cabeceras defensivas presentes.
- Ningún puerto de aplicación o PostgreSQL publicado en host.
- Tunnel aparece `Healthy` en Cloudflare.

Probar desde navegador:

1. Registro crea cuenta pendiente.
2. OTP llega por SES y activa cuenta.
3. Login y logout funcionan.
4. Recuperación de contraseña funciona.
5. Invitación de empleado llega por correo.
6. Reinicio de contenedores conserva datos PostgreSQL.

Mientras SES esté en sandbox, destinatarios deben estar verificados.

## 9. Sacar Amazon SES del sandbox

Hacerlo en región **US East (Ohio), `us-east-2`**, después de que sitio HTTPS funcione:

1. SES > **Account dashboard**.
2. **View Get set up page**.
3. **Request production access**.
4. Mail type: **Transactional**.
5. Website URL: `https://stockion.mergemakers.com`.
6. Agregar correo de contacto.
7. Aceptar reglas anti-spam y enviar solicitud.

Descripción sugerida:

```text
Stockion sends low-volume transactional emails triggered by user actions:
email verification OTPs, password resets, employee invitations, and security
notifications. We do not send marketing email or use purchased mailing lists.
Recipients register directly or are invited by an authorized administrator of
their organization. Bounces and complaints are monitored through Amazon SES,
and failed addresses are not repeatedly contacted.
```

Tras aprobación, probar un destinatario no verificado y revisar métricas de reputación.

## 10. Operación normal

Estado y logs:

```bash
cd /home/thrs/apps/stockion
sudo docker compose --env-file .env.production -f compose.production.yaml ps
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=200 backend gateway cloudflared postgres
```

Reinicio controlado:

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml restart
```

Actualización de aplicación:

```bash
git switch production
git pull --ff-only origin production
sudo docker compose --env-file .env.production -f compose.production.yaml build --pull
sudo docker compose --env-file .env.production -f compose.production.yaml up -d
sudo docker compose --env-file .env.production -f compose.production.yaml ps
```

Revisar siempre estado, espacio y logs antes y después.

## 11. Diagnóstico rápido

### Cloudflare muestra Tunnel caído

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=100 cloudflared
sudo stat -c '%a %u:%g %n' /home/thrs/.config/stockion/cloudflare-tunnel-token
```

El token debe ser legible por `65532:65532` y tener permiso `600`.

### Sitio devuelve 502

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml ps
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=100 gateway backend frontend
```

Confirmar servicios `healthy`; no reiniciar borrando datos.

### Correo no llega

Revisar:

- Región y endpoint `us-east-2`.
- `SMTP_TLS=starttls` y puerto `587`.
- Credenciales SMTP nuevas, no access keys normales.
- Remitente verificado.
- Restricciones del sandbox.
- Logs del backend sin publicar secretos.

### Riesgo de disco

```bash
df -h /
sudo docker system df
sudo journalctl --disk-usage
```

No borrar volúmenes para liberar espacio.

## 12. Acciones prohibidas o peligrosas

- No ejecutar `docker compose down -v`: elimina PostgreSQL.
- No ejecutar `docker volume rm` sobre volumen de Stockion.
- No pegar secretos en chat, GitHub, capturas, logs o correo.
- No publicar PostgreSQL, backend, Caddy, Nginx o métricas.
- No activar escritura en deploy key.
- No usar cuenta root de AWS para operación diaria.
- No desactivar UFW para diagnosticar.

## 13. Limitación aceptada

Esta demo no tiene respaldos de base de datos por decisión temporal. Una falla de disco, corrupción, eliminación o migración defectuosa puede causar pérdida total e irreversible.

Antes de almacenar datos reales o depender comercialmente del servicio deben implementarse respaldos externos y una restauración probada.
