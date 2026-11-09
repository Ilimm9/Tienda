# Pasos de operación en el servidor Stockion

Ejecuta estos comandos tú en la terminal SSH de Lightsail. No compartas el contenido del `.env`, la API key ni el token del túnel.

## Índice

1. [Actualizar código desde production](#1-actualizar-código-desde-production) ← uso habitual
2. [Configurar correo (Resend)](#2-configurar-correo-resend) 🆕 solo al cambiar proveedor o API key
3. [⚠️ Borrar y recrear la base de datos](#3-️-borrar-y-recrear-la-base-de-datos)
4. [Conexión remota a PostgreSQL](#4-conexión-remota-a-postgresql)
5. [Cerrar la conexión remota (seguridad)](#5-cerrar-la-conexión-remota-seguridad) 🔁 cada vez
6. [Respaldo de la base y copia al equipo local](#6-respaldo-de-la-base-y-copia-al-equipo-local)

Marcadores usados en los pasos: 🆕 **solo la primera vez** · 🔁 **cada vez**.

Atajo usado en varias secciones para no repetir rutas:

```bash
C="sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml"
```

---

## 1. Actualizar código desde production

Sólo aplica cuando los cambios de código ya estén publicados en `origin/production`. Para cambiar únicamente SMTP, basta la sección 2.

Ejecuta cada comando por separado. Si uno falla, detente; no reemplaces contenedores con una compilación fallida.

### 1.1 Revisar y descargar cambios

```bash
sudo -u stockion -H git -C /opt/stockion status --short
```

Si aparecen cambios locales, revísalos antes de continuar; no los sobrescribas.

```bash
sudo -u stockion -H git -C /opt/stockion pull --ff-only origin production
```

### 1.2 Validar configuración

```bash
sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 config --quiet
```

### 1.3 Compilar sin detener la aplicación actual

```bash
sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 build --pull
```

### 1.4 Reemplazar contenedores

```bash
sudo docker compose \
 --env-file /opt/stockion/.env.production \
 -f /opt/stockion/compose.production.yaml \
 up -d --remove-orphans
```

### 1.5 Verificar

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

---

## 2. Configurar correo (Resend)

### 2.1 Preparar Resend

Primero verifica en Resend el dominio de envío `send.mergemakers.com` y crea una API key con permiso de envío para ese dominio. Agrega en Cloudflare los registros que indique Resend; conserva los registros existentes de SES.

### 2.2 Editar variables

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

### 2.3 Validar y aplicar

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

### 2.4 Probar envío real

1. Abre https://stockion.mergemakers.com y registra una cuenta de prueba con un correo tuyo que no esté registrado.
2. Confirma que aparece el paso de OTP y llega el código. Revisa también spam.
3. Introduce el código y comprueba que puedes entrar.
4. En Resend → **Emails**, revisa el envío y su estado. Un backend `healthy` no garantiza entrega del correo.

Si falla, revisa logs del backend y el detalle del envío en Resend. Comprueba dominio verificado, API key, remitente y cuota disponible. No desactives OTP para solucionar el envío.

Referencia: [SMTP de Resend](https://resend.com/docs/send-with-smtp).

---

## 3. ⚠️ Borrar y recrear la base de datos

> [!CAUTION]
> **Esto elimina TODOS los datos de producción y no se puede deshacer.**
> Usuarios, negocios, sucursales, productos, compras y sesiones desaparecen.
> Saca respaldo primero (3.1) y hazlo solo si estás completamente seguro.

Te dejo los dos casos; el orden importa: primero se detiene el backend, porque Postgres no deja borrar una base con conexiones abiertas.

### 3.1 Respaldo antes de borrar

```bash
$C exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' > respaldo-$(date +%F).sql
```

### 3.2 En el servidor (producción)

1. Detener backend:

   ```bash
   $C stop backend
   ```

2. Borrar y volver a crear la base (usa usuario y nombre que ya tiene el contenedor):

   ```bash
   $C exec postgres sh -c 'dropdb -U "$POSTGRES_USER" --if-exists "$POSTGRES_DB" && createdb -U "$POSTGRES_USER" "$POSTGRES_DB"'
   ```

3. Levantar backend; al arrancar crea tablas, permisos y migraciones sobre la base vacía:

   ```bash
   $C up -d backend
   $C logs -f backend
   ```

   Busca la línea `API escuchando`. Ctrl+C para salir del log.

> Después de recrear la base, los usuarios `soporte_ro` y `soporte_rw` siguen existiendo (los roles son del servidor, no de la base), pero pierden sus permisos sobre las tablas nuevas. Repite los `GRANT` y `ALTER DEFAULT PRIVILEGES` de la sección 4.1 (sin los `CREATE ROLE`).

### 3.3 En tu máquina (local)

1. Detén tu go run (Ctrl+C) y el contenedor de backend:

   ```bash
   docker stop tienda-backend-1
   ```

2. Borrar y crear:

   ```bash
   docker exec tienda-postgres-1 sh -c 'dropdb -U postgres --if-exists tienda && createdb -U postgres tienda'
   ```

3. Arranca de nuevo:

   ```bash
   docker start tienda-backend-1 # o: go run ./cmd/api
   ```

### 3.4 Si quieres entrar a mano

```bash
docker exec -it tienda-postgres-1 psql -U postgres # local
$C exec postgres sh -c 'psql -U "$POSTGRES_USER" postgres' # servidor
```

Dentro de psql:

```sql
\l -- lista bases
DROP DATABASE tienda; -- nombre real según \l
CREATE DATABASE tienda;
\q
```

Conéctate a la base postgres, no a la que vas a borrar. Si dice «is being accessed by other users», falta detener backend, o usa `DROP DATABASE tienda WITH (FORCE);`.

### 3.5 Qué resuelve y qué no

Con base limpia desaparecen dos pendientes: ya no hay miembros sin sucursal asignada ni correos duplicados que revisar; todas las migraciones corren sobre tablas vacías.

Siguen en pie:

- Intentos de OTP (código 10, prueba 5).
- Probar un correo real en Gmail/Outlook.
- En producción la base queda sin ningún usuario: la primera cuenta se crea desde «Registrarse» y pasa por el asistente de empresa y sucursal.
- Sesiones abiertas dejan de servir; todos vuelven a iniciar sesión.

---

## 4. Conexión remota a PostgreSQL

PostgreSQL no tiene puerto público. El acceso se hace por SSH y un puente temporal ligado a `127.0.0.1` del servidor. Nunca abrir `5432` en Lightsail ni UFW del servidor, ni publicar con `0.0.0.0` (Docker se salta UFW).

Datos reales: base `stockion`, usuario de la aplicación `stockion`, red interna `stockion_data`.

Flujo de cada sesión (detalle en 4.3 a 4.5):

1. Servidor: levantar `pg-puente`.
2. Local: abrir túnel SSH.
3. pgAdmin: conectar.
4. Al terminar: sección 5.

### 4.1 🆕 Usuarios de soporte (solo la primera vez)

- `soporte_ro`: solo lectura, uso diario.
- `soporte_rw`: lectura y edición de filas, solo para correcciones. Ninguno puede `DROP`, `ALTER` ni crear tablas, por lo que respetan la regla de esquema aditivo.

Abrir psql como usuario de la aplicación:

```bash
$C exec postgres sh -c 'psql -U "$POSTGRES_USER" "$POSTGRES_DB"'
```

Dentro de psql (`\password` pide la clave sin dejarla en el historial; generarla con `openssl rand -base64 24`):

```sql
-- Solo lectura (uso diario)
CREATE ROLE soporte_ro LOGIN;
\password soporte_ro
GRANT CONNECT ON DATABASE stockion TO soporte_ro;
GRANT USAGE ON SCHEMA public TO soporte_ro;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO soporte_ro;
ALTER DEFAULT PRIVILEGES FOR ROLE stockion IN SCHEMA public GRANT SELECT ON TABLES TO soporte_ro;
ALTER ROLE soporte_ro SET default_transaction_read_only = on;
ALTER ROLE soporte_ro SET statement_timeout = '30s';

-- Edición (solo para correcciones; no puede borrar tablas ni cambiar el esquema)
CREATE ROLE soporte_rw LOGIN;
\password soporte_rw
GRANT CONNECT ON DATABASE stockion TO soporte_rw;
GRANT USAGE ON SCHEMA public TO soporte_rw;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO soporte_rw;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO soporte_rw;
ALTER DEFAULT PRIVILEGES FOR ROLE stockion IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO soporte_rw;
ALTER DEFAULT PRIVILEGES FOR ROLE stockion IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO soporte_rw;
ALTER ROLE soporte_rw SET statement_timeout = '30s';
\q
```

`DEFAULT PRIVILEGES` cubre las tablas que la aplicación cree en migraciones futuras. Tablas creadas antes de este paso ya quedan cubiertas por `ON ALL TABLES`.

Consulta rápida sin interfaz gráfica:

```bash
$C exec postgres psql -U soporte_ro -d stockion
```

### 4.2 🆕 Configuración local (solo la primera vez)

En el equipo local, `~/.ssh/config`:

```
Host stokion
    HostName 52.7.215.247
    User thrs
    IdentityFile ~/.ssh/tienda_prod
    IdentitiesOnly yes
```

El servidor solo acepta llave; la contraseña está desactivada. La llave `tienda_prod` tiene passphrase. Probar con `ssh stokion`. Si `HostName` deja de responder, confirmar la IP pública desde la consola de Lightsail con `curl -s checkip.amazonaws.com`.

pgAdmin local corre en Docker (`localhost:5052`, contenedor `tienda-pgadmin-1`, red `tienda_default`). No subir la llave privada a pgAdmin: el túnel lo abre el equipo local y pgAdmin se conecta a él por el gateway de esa red.

Obtener el gateway (actualmente `172.22.0.1`; cambia si se recrea la red):

```bash
docker network inspect tienda_default --format '{{range .IPAM.Config}}{{.Gateway}}{{end}}'
```

Permitir en el firewall **local** que la red de pgAdmin alcance el túnel (solo si `sudo ufw status` está activo; en el servidor no se toca ufw):

```bash
sudo ufw allow from 172.22.0.0/16 to 172.22.0.1 port 15432 proto tcp
```

Registrar en pgAdmin, **Register → Server**:

- **SSH Tunnel:** desactivado.
- **Connection:** host `172.22.0.1`, puerto `15432`, maintenance DB `stockion`, usuario `soporte_ro` (o `soporte_rw` en un servidor registrado aparte, solo para correcciones).

### 4.3 🔁 Levantar el puente (servidor, cada vez)

`stockion_data` es `internal: true` y no admite puertos publicados. El puente se publica por la red `bridge` por defecto y se conecta a `stockion_data` para alcanzar `postgres`:

```bash
sudo docker run -d --rm --name pg-puente -p 127.0.0.1:15432:5432 \
  alpine/socat TCP-LISTEN:5432,fork,reuseaddr TCP:postgres:5432
sudo docker network connect stockion_data pg-puente
```

Comprobar: `sudo docker ps | grep pg-puente`. El puente no forma parte de Compose y no sobrevive a un reinicio del servidor.

### 4.4 🔁 Abrir el túnel (local, cada vez)

En una terminal que se queda abierta. Pide la passphrase y luego no muestra nada: eso significa que está activo.

```bash
ssh -N -L 172.22.0.1:15432:127.0.0.1:15432 stokion
```

Escucha solo en el gateway de Docker, no en la red local. Verificar con `ss -ltn | grep 15432`.

### 4.5 🔁 Conectar

Abrir pgAdmin en `localhost:5052` y conectar al servidor registrado.

Alternativa con DBeaver o cliente nativo: `ssh -N -L 15432:127.0.0.1:15432 stokion` y conectar a `127.0.0.1:15432`, o usar el túnel SSH integrado del cliente con la llave `~/.ssh/tienda_prod`. En DBeaver elegir **Connection type: Production** (desactiva auto-commit y pide confirmación).

### 4.6 Antes de modificar datos

1. Respaldo:

   ```bash
   $C exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB"' > ~/respaldo-$(date +%F-%H%M).sql
   ```

2. Editar con `soporte_rw`; los cambios quedan pendientes hasta Save y Commit.
3. Evitar `DELETE`: puede romper relaciones o supuestos del backend. Preferir corregir o desactivar el registro.

### 4.7 Diagnóstico

- `Permission denied (publickey)`: la llave no está en `~/.ssh/authorized_keys` del servidor o falta `IdentityFile`. Comparar con `~/.ssh/tienda_prod.pub`.
- pgAdmin «Failed to create the SSH tunnel»: tiene activado SSH Tunnel con contraseña; desactivarlo y usar el túnel manual.
- Timeout en pgAdmin: túnel cerrado, gateway distinto o falta la regla ufw local.
- Conexión rechazada: `pg-puente` no está levantado o no está conectado a `stockion_data`.

---

## 5. Cerrar la conexión remota (seguridad)

Hazlo siempre al terminar; no dejes el puente ni el túnel abiertos.

1. pgAdmin: clic derecho en el servidor → **Disconnect Server**.
2. Local: en la terminal del túnel, **Ctrl+C**.
3. Servidor: detener el puente (se elimina solo por `--rm`):

   ```bash
   sudo docker stop pg-puente
   ```

4. Verificar que no quede nada abierto:

   ```bash
   sudo docker ps | grep pg-puente   # servidor: sin salida
   ss -ltn | grep 15432              # local: sin salida
   ```

5. Si guardaste respaldos `.sql` en el servidor o local, muévelos a un lugar seguro o bórralos: contienen todos los datos.

Opcional, si ya no usarás pgAdmin local contra producción, quitar la regla del firewall local:

```bash
sudo ufw delete allow from 172.22.0.0/16 to 172.22.0.1 port 15432 proto tcp
```

---

## 6. Respaldo de la base y copia al equipo local

Formato comprimido de `pg_dump` (`-Fc`): ocupa menos y se restaura con `pg_restore`. Usa `-T` en `exec`: sin él Docker asigna terminal y puede corromper el archivo.

Un respaldo que solo vive en el servidor no protege ante falla de disco: siempre copiarlo al equipo local y borrarlo del servidor.

### 6.1 🆕 Preparar carpetas (solo la primera vez)

Servidor:

```bash
mkdir -p ~/backups && chmod 700 ~/backups
```

Local:

```bash
mkdir -p ~/backups/stockion && chmod 700 ~/backups/stockion
```

### 6.2 🔁 Crear el respaldo (servidor, cada vez)

```bash
C="sudo docker compose --env-file /opt/stockion/.env.production -f /opt/stockion/compose.production.yaml"
F=~/backups/stockion-$(date +%F-%H%M).dump
$C exec -T postgres sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$F"
chmod 600 "$F"
```

### 6.3 🔁 Verificar (servidor, cada vez)

```bash
ls -lh "$F"
$C exec -T postgres pg_restore --list < "$F" | head -20
sha256sum "$F"
```

Correcto si `pg_restore --list` muestra cabecera (`dbname: stockion`, `Format: CUSTOM`, versión 17.x) y tablas. Anotar el hash de `sha256sum`.

### 6.4 🔁 Copiar al equipo local (local, cada vez)

```bash
scp stokion:~/backups/stockion-AAAA-MM-DD-HHMM.dump ~/backups/stockion/
sha256sum ~/backups/stockion/stockion-AAAA-MM-DD-HHMM.dump
```

El hash debe coincidir exactamente con el del servidor. Si no coincide, repetir la copia.

### 6.5 🔁 Borrar del servidor (cada vez, tras validar el hash)

```bash
rm ~/backups/stockion-AAAA-MM-DD-HHMM.dump
```

### 6.6 Restaurar en el equipo local para revisar (opcional)

El Postgres local (`tienda-postgres-1`) es 17.11, igual que producción. Se restaura en una base aparte, `stockion_prod`, para no tocar `tienda` (desarrollo). Todo se ejecuta en el equipo local.

`--no-owner --no-acl` omite dueños y permisos: los roles de producción (`stockion`, `soporte_*`) no existen en local y sin estas opciones aparecen errores de permisos.

1. 🔁 Indicar el archivo y comprobar su hash (debe coincidir con el del servidor, paso 6.3):

   ```bash
   B=~/backups/stockion/stockion-AAAA-MM-DD-HHMM.dump
   sha256sum "$B"
   ```

2. 🔁 Comprobar que la base destino no existe (sin salida = no existe):

   ```bash
   docker exec tienda-postgres-1 psql -U postgres -Atc "select datname from pg_database where datname='stockion_prod'"
   ```

   Si ya existe de una restauración anterior, borrarla primero (paso 6).

3. 🔁 Crear la base y restaurar:

   ```bash
   docker exec tienda-postgres-1 createdb -U postgres stockion_prod
   docker exec -i tienda-postgres-1 pg_restore -U postgres -d stockion_prod --no-owner --no-acl < "$B"
   echo "restore exit=$?"
   ```

   `restore exit=0` significa que terminó sin errores.

4. 🔁 Verificar tablas y filas:

   ```bash
   docker exec tienda-postgres-1 psql -U postgres -d stockion_prod -Atc "select count(*) from information_schema.tables where table_schema='public'"
   docker exec tienda-postgres-1 psql -U postgres -d stockion_prod -c "select relname as tabla, n_live_tup as filas from pg_stat_user_tables order by n_live_tup desc limit 15"
   ```

   El primero da el total de tablas; el segundo las 15 con más filas (conteo aproximado de estadísticas de Postgres).

5. 🔁 Ver en pgAdmin (`localhost:5052`): clic derecho en el servidor local → **Refresh** → `stockion_prod` → Schemas → public → Tables.

6. 🔁 Borrar cuando ya no se necesite (`--force` cierra conexiones abiertas, por ejemplo de pgAdmin):

   ```bash
   docker exec tienda-postgres-1 dropdb -U postgres --force stockion_prod
   ```

Alternativa desde pgAdmin (sin terminal para restaurar): copiar el archivo al almacenamiento de pgAdmin, porque el contenedor no ve los archivos del equipo:

```bash
docker cp ~/backups/stockion/stockion-AAAA-MM-DD-HHMM.dump tienda-pgadmin-1:/var/lib/pgadmin/storage/
docker exec -u root tienda-pgadmin-1 chown 5050 /var/lib/pgadmin/storage/stockion-AAAA-MM-DD-HHMM.dump
```

En pgAdmin: crear la base `stockion_prod` → clic derecho → **Restore…** → Format **Custom or tar**, Filename en `/var/lib/pgadmin/storage/` → pestaña **Data Options**, «Do not save»: **Owner** y **Privileges** → **Restore**. Al terminar, borrar la copia del contenedor:

```bash
docker exec -u root tienda-pgadmin-1 rm /var/lib/pgadmin/storage/stockion-AAAA-MM-DD-HHMM.dump
```

> Los `.dump` contienen todos los datos de producción. Guardarlos en un lugar seguro o cifrado; no subirlos al repositorio ni compartirlos.
