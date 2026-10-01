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
- Deploy key de sólo lectura instalada para el usuario de servicio `stockion`.
- Rama `production` clonada en `/opt/stockion`.
- Stack productivo construido y levantado desde PostgreSQL vacío.
- Public Hostname `stockion.mergemakers.com` conectado a `http://gateway:8080`.
- Aplicación pública operativa por HTTPS, sin puertos de aplicación expuestos.
- Volumen persistente `stockion_postgres_data` creado y conservado entre despliegues.
- PrimeIcons compatibles con la CSP estricta y locale Angular `es-MX` registrado.
- Cloudflare Web Analytics/RUM desactivado para evitar inyectar un beacon bloqueado por la CSP.

### Pendiente

- Mantener `.env.production` y credenciales fuera de Git.
- Completar pruebas funcionales de los módulos principales.
- Esperar resolución del caso de acceso a producción de Amazon SES.
- Mientras SES permanezca en sandbox, verificar manualmente cada destinatario de prueba.
- Implementar respaldos externos y restauración probada antes de almacenar información crítica.

## 3. Archivos importantes

- `compose.production.yaml`: servicios, redes, límites y healthchecks.
- `.env.production.example`: variables requeridas sin secretos.
- `deploy/Caddyfile`: gateway HTTP interno.
- `deploy/Dockerfile.caddy`: Caddy no-root y sin capacidades innecesarias.
- `frontend/Dockerfile.production`: imagen Angular/Nginx no-root.
- `frontend/nginx.production.conf`: servicio estático interno.

No usar `docker-compose.yml` para producción: incluye puertos y herramientas de desarrollo.

## 4. Repositorio privado y usuario de servicio

El repositorio se administra con el usuario de servicio `stockion`; los operadores humanos usan `sudo -u stockion`. La llave privada existe únicamente en Lightsail:

```bash
sudo stat -c '%a %U:%G %n' /var/lib/stockion/.ssh/stockion_github
```

Permiso esperado: `600 stockion:stockion`. Nunca mostrar ni copiar la llave privada.

Validar autenticación y acceso de sólo lectura:

```bash
sudo -u stockion -H ssh \
  -i /var/lib/stockion/.ssh/stockion_github \
  -o IdentitiesOnly=yes \
  -T git@github.com

sudo -u stockion -H git -C /opt/stockion status
sudo -u stockion -H git -C /opt/stockion branch --show-current
sudo -u stockion -H git -C /opt/stockion ls-remote --heads origin production
```

La deploy key registrada en GitHub debe mantener `Allow write access` desactivado. Los operadores, incluida Ilián, deben iniciar sesión con su propia cuenta y tener autorización `sudo` para operar como `stockion`; no necesitan ser propietarios de `/opt/stockion`.

No ejecutar Git como `root`, compartir la llave privada ni aplicar `chmod -R 777`.

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
cd /opt/stockion
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

Desde `/opt/stockion`:

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml config --quiet
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml build --pull
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml up -d --remove-orphans
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml ps
```

Todos los servicios deben aparecer `Up`; PostgreSQL, backend, frontend y gateway deben terminar como `healthy`.

Logs iniciales:

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml logs --tail=100
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

Para conservar la CSP `script-src 'self'`, mantener desactivado el beacon JavaScript en **Analytics & Logs > Web Analytics > Manage site > Automatic setup > Disable**. Si no aparece esa opción, crear una Configuration Rule para `stockion.mergemakers.com` con **Disable Real User Monitoring (RUM)**. Esto no afecta Tunnel, DNS ni las métricas de tráfico obtenidas en el edge.

## 8. Verificación pública

```bash
curl -fsS https://stockion.mergemakers.com/api/v1/health
curl -fsSI https://stockion.mergemakers.com/
sudo ss -lntp
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml ps
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

Mientras SES esté en sandbox, cada destinatario externo de prueba debe agregarse en **SES > Configuration > Identities > Create identity > Email address** dentro de `us-east-2`. El propietario del correo debe abrir el mensaje de AWS y confirmar el enlace antes de recibir OTP.

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

La solicitud fue enviada y AWS pidió información adicional; se respondió en el mismo caso con URL, tipo transaccional, volumen, origen de destinatarios, manejo de rebotes/quejas y muestra de correo. No crear solicitudes duplicadas mientras el caso siga abierto.

Tras aprobación, confirmar que Account dashboard ya no indique sandbox, probar un destinatario no verificado y revisar métricas de reputación.

## 10. Operación normal

Estado y logs:

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml ps
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml logs --tail=200 backend frontend gateway cloudflared postgres
```

Reinicio controlado:

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml restart
```

Actualización de aplicación:

```bash
sudo -u stockion -H git -C /opt/stockion status --short
sudo -u stockion -H git -C /opt/stockion pull --ff-only origin production
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml config --quiet
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml build --pull
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml up -d --remove-orphans
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml ps
```

Construir antes de ejecutar `up` evita reemplazar contenedores si la compilación falla. `up -d --remove-orphans` conserva el volumen nombrado `stockion_postgres_data`; no usar `down -v`.

Revisar siempre estado, espacio y logs antes y después. Tras actualizar Angular, hacer recarga forzada en el navegador.

## 11. Diagnóstico rápido

### Cloudflare muestra Tunnel caído

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml logs --tail=100 cloudflared
sudo stat -c '%a %u:%g %n' /home/thrs/.config/stockion/cloudflare-tunnel-token
```

El token debe ser legible por `65532:65532` y tener permiso `600`.

### Sitio devuelve 502

```bash
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml ps
sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml logs --tail=100 gateway backend frontend
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

### Angular muestra `NG0701`

Confirmar que la versión desplegada incluye `import '@angular/common/locales/global/es-MX';` antes del bootstrap, reconstruir frontend y hacer recarga forzada. El locale mexicano debe tener una prueba específica que ejecute `FechaMexicoPipe`.

### Consola muestra `VM... reportAllChanges ... startTime`

Si el HTML público ya no contiene `static.cloudflareinsights.com`, este mensaje proviene del medidor Web Vitals inyectado por Chrome DevTools, no de Stockion. Actualizar Chrome, reiniciarlo y probar sin DevTools. Como alternativa temporal, desactivar `chrome://flags/#soft-navigation-heuristics`.

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
- No ejecutar `docker system prune --volumes`.
- No cambiar el nombre del proyecto Compose ni `postgres_data` sin una migración planificada.
- No pegar secretos en chat, GitHub, capturas, logs o correo.
- No publicar PostgreSQL, backend, Caddy, Nginx o métricas.
- No activar escritura en deploy key.
- No usar cuenta root de AWS para operación diaria.
- No desactivar UFW para diagnosticar.
- No ejecutar Git como `root`; usar `sudo -u stockion -H git -C /opt/stockion ...`.

## 13. Limitación aceptada

Esta demo no tiene respaldos de base de datos por decisión temporal. Una falla de disco, corrupción, eliminación o migración defectuosa puede causar pérdida total e irreversible.

Antes de almacenar datos reales o depender comercialmente del servicio deben implementarse respaldos externos y una restauración probada.
