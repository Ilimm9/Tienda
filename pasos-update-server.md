# Configurar correo y actualizar Stockion

Ejecuta estos comandos tú en la terminal SSH de Lightsail. No compartas el contenido del `.env`, la API key ni el token del túnel.

## 1. Configurar Resend antes de actualizar

Primero verifica en Resend el dominio de envío `send.mergemakers.com` y crea una API key con permiso de envío para ese dominio. Agrega en Cloudflare los registros que indique Resend; conserva los registros existentes de SES.

Abre el archivo con **nano**:

```bash
sudo nano /opt/stockion/.env.production
```

Edita las líneas existentes; no dupliques variables. Sustituye el marcador de contraseña por tu API key real:

```dotenv
SMTP_HOST=smtp.resend.com
SMTP_PORT=587
SMTP_TLS=starttls
SMTP_USERNAME=resend
SMTP_PASSWORD=REEMPLAZAR_POR_API_KEY_DE_RESEND
SMTP_TIMEOUT=10s
MAIL_FROM=no-reply@send.mergemakers.com
MAIL_FROM_NAME=Stockion
SES_CONFIGURATION_SET=
```

Si verificaste otro dominio, ajusta `MAIL_FROM` para usarlo. No cambies `DB_PASSWORD`, `OTP_HMAC_SECRET` ni la configuración del túnel.

Guardar: **Ctrl+O**, **Enter**. Salir: **Ctrl+X**.

Valida sin imprimir secretos:

```bash
sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 config --quiet
```

Sin salida y sin error significa que Compose acepta la configuración; todavía no comprueba las credenciales SMTP. Si falla, corrige antes de continuar.

Aplica sólo al backend: no requiere compilar ni reiniciar PostgreSQL. Habrá una interrupción breve de la API.

```bash
sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 up -d --no-deps --force-recreate backend

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 ps backend

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 logs --tail 100 backend
```

Espera a que el backend indique `healthy`. Antes de compartir logs, revisa que no contengan datos sensibles.

### Probar envío real

1. Abre https://stockion.mergemakers.com y registra una cuenta de prueba con un correo tuyo que no esté registrado.
2. Confirma que aparece el paso de OTP y llega el código. Revisa también spam.
3. Introduce el código y comprueba que puedes entrar.
4. En Resend → **Emails**, revisa el envío y su estado. Un backend `healthy` no garantiza entrega del correo.

Si falla, revisa logs del backend y el detalle del envío en Resend. Comprueba dominio verificado, API key, remitente y cuota disponible. No desactives OTP para solucionar el envío.

Referencia: [SMTP de Resend](https://resend.com/docs/send-with-smtp).

deshacer. Si hay alguna duda, saca respaldo primero (comando al final).

Te dejo los dos casos; el orden importa: primero se detiene el backend, porque Postgres no deja borrar una base con conexiones abiertas.

En el servidor (producción)

Define un atajo para no repetir rutas:

C="sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml"

1. Detener backend:
   $C stop backend

2. Borrar y volver a crear la base (usa usuario y nombre que ya tiene el contenedor):
   $C exec postgres sh -c 'dropdb -U "$POSTGRES_USER" --if-exists "$POSTGRES_DB" && createdb -U "$POSTGRES_USER" "$POSTGRES_DB"'

3. Levantar backend; al arrancar crea tablas, permisos y migraciones sobre la base vacía:
   $C up -d backend
$C logs -f backend
   Busca la línea API escuchando. Ctrl+C para salir del log.

En tu máquina (local)

1. Detén tu go run (Ctrl+C) y el contenedor de backend:
   docker stop tienda-backend-1

2. Borrar y crear:
   docker exec tienda-postgres-1 sh -c 'dropdb -U postgres --if-exists tienda && createdb -U postgres tienda'

3. Arranca de nuevo:
   docker start tienda-backend-1 # o: go run ./cmd/api

Si quieres entrar a mano

docker exec -it tienda-postgres-1 psql -U postgres # local
$C exec postgres sh -c 'psql -U "$POSTGRES_USER" postgres' # servidor

Dentro de psql:
\l -- lista bases
DROP DATABASE tienda; -- nombre real según \l
CREATE DATABASE tienda;
\q
Conéctate a la base postgres, no a la que vas a borrar. Si dice «is being accessed by other users», falta detener backend, o usa DROP DATABASE tienda WITH (FORCE);.

Respaldo opcional antes de borrar

$C exec postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' > respaldo-$(date +%F).sql

Qué resuelve y qué no

Con base limpia desaparecen dos pendientes: ya no hay miembros sin sucursal asignada ni correos duplicados que revisar; todas las migraciones corren sobre tablas vacías.

Siguen en pie:

- Intentos de OTP (código 10, prueba 5).
- Probar un correo real en Gmail/Outlook.
- En producción la base queda sin ningún usuario: la primera cuenta se crea desde «Registrarse» y pasa por el asistente de empresa y sucursal.
- Sesiones abiertas dejan de servir; todos vuelven a iniciar sesión.

## 2. Actualizar código desde production

Sólo aplica cuando los cambios de código ya estén publicados en `origin/production`. Para cambiar únicamente SMTP, basta la sección anterior.

Ejecuta cada comando por separado. Si uno falla, detente; no reemplaces contenedores con una compilación fallida.

### Revisar y descargar cambios

```bash
sudo -u stockion -H git -C /opt/stockion status --short
```

Si aparecen cambios locales, revísalos antes de continuar; no los sobrescribas.

```bash

sudo -u stockion -H git -C /opt/stockion pull --ff-only origin production

```

### Validar configuración

```bash

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 config --quiet

```

### Compilar sin detener la aplicación actual

```bash

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 build --pull

```

### Reemplazar contenedores

```bash

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 up -d --remove-orphans

```

### Verificar

```bash

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 ps

```

Luego revisa errores:

```bash

sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 logs --tail 100 backend frontend gateway cloudflared

```

Finalmente abre:

https://stockion.mergemakers.com

Haz recarga forzada con Ctrl+Shift+R.

Estos comandos no borran el volumen `stockion_postgres_data`. Conserva el mismo proyecto y archivo Compose. No ejecutes `down -v`, `docker volume rm` ni limpiezas de volúmenes: esas operaciones sí pueden eliminar la base de datos.
