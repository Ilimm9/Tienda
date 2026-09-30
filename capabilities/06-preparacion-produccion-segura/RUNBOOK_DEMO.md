# Runbook de demo robusta

## Límites de responsabilidad

El propietario ejecuta todos los cambios en GitHub, Cloudflare y Lightsail. Ningún token, contraseña, credencial SMTP o llave privada se comparte por chat, captura o correo.

## Orden obligatorio

1. Revisar y aceptar los archivos locales de producción.
2. Autorizar por separado creación de rama, commit y push de `production` desde `thrs`.
3. Crear una deploy key de sólo lectura para el repositorio privado y guardar la llave privada únicamente en Lightsail.
4. Clonar la rama `production` en Lightsail.
5. Crear `.env.production`, permisos `0600`, usando `.env.production.example`.
6. Rotar las credenciales SMTP usadas durante pruebas y colocar las nuevas sólo en `.env.production`.
7. Crear en Cloudflare un túnel remoto llamado `stockion-production`.
8. Crear `secrets/cloudflare-tunnel-token`, permisos `0600`, con sólo el token del túnel.
9. En Cloudflare, publicar `stockion.mergemakers.com` hacia `http://gateway:8080`. El registro A anterior deberá retirarse al crear el hostname del túnel.
10. Construir y levantar el Compose de producción.
11. Verificar aplicación, correo, reinicio y superficie pública.
12. Cerrar `80/443` en Lightsail y UFW después de confirmar el túnel. Mantener SSH por llave y restringido.
13. Solicitar acceso de producción de Amazon SES en `us-east-2`.

## Comandos del servidor

Estos comandos se ejecutan manualmente desde la carpeta del repositorio:

```bash
chmod 600 .env.production secrets/cloudflare-tunnel-token
sudo docker compose --env-file .env.production -f compose.production.yaml config --quiet
sudo docker compose --env-file .env.production -f compose.production.yaml build --pull
sudo docker compose --env-file .env.production -f compose.production.yaml up -d
sudo docker compose --env-file .env.production -f compose.production.yaml ps
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=100
```

No ejecutar `down -v`: elimina el volumen de PostgreSQL y todos los datos.

## Verificación mínima

```bash
curl -fsS https://stockion.mergemakers.com/api/v1/health
curl -fsSI https://stockion.mergemakers.com/
sudo ss -lntp
sudo docker compose --env-file .env.production -f compose.production.yaml ps
```

Resultado esperado: salud `200`, HTTPS válido, todos los contenedores activos y ningún puerto de aplicación o PostgreSQL publicado en el host.

Después se prueban manualmente registro, OTP, login, recuperación, invitación y correo a un destinatario no verificado cuando SES salga del sandbox.

## Actualización

La actualización queda pendiente de la rama y política Git aprobadas. Antes de cada cambio se revisarán espacio, estado de contenedores y migraciones. Sin respaldos, una migración destructiva no tiene recuperación de datos.

## Recuperación de servicio

```bash
sudo docker compose --env-file .env.production -f compose.production.yaml restart
sudo docker compose --env-file .env.production -f compose.production.yaml logs --tail=200 backend gateway cloudflared postgres
```

Si una imagen nueva falla, volver al commit anterior requiere autorización Git explícita. No borrar volúmenes para solucionar un fallo.
