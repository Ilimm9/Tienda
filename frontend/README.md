# Frontend de Tienda

Frontend Angular 22 para registro, inicio de sesión y página inicial protegida.

## Requisitos

- Node.js 24 LTS mediante NVM.
- Backend disponible en el puerto `APP_PORT` configurado en `backend/.env` (`8080` por defecto).

## Desarrollo

```bash
nvm use
npm install
npm start
```

Abrir el puerto `FRONTEND_PORT` (`http://localhost:4200` por defecto).

`npm start` lee, en este orden de prioridad: variables exportadas, `frontend/.env`
(opcional, ver `.env.example`), `backend/.env` y valores por defecto.
`frontend/.env` acepta `FRONTEND_PORT` y `BACKEND_URL` (destino del proxy `/api`).

Los puertos y el origen CORS se configuran juntos:

```env
APP_PORT=8080
FRONTEND_PORT=4200
FRONTEND_URL=http://localhost:4200
```

Angular consume `/api/v1` y reenvía `/api/*` al backend mediante el proxy de desarrollo.

## Variables y secretos

`frontend/.env` solo configura el servidor de desarrollo; no entra al bundle.
La configuración pública de la app vive en
`src/environments/environment.ts` y `src/environments/environment.prod.ts`.

Nunca poner en el frontend `SMTP_USERNAME`, `SMTP_PASSWORD`, claves AWS,
`JWT_SECRET`, `OTP_HMAC_SECRET` ni credenciales de base de datos: todo valor
incluido en el bundle Angular queda visible para el navegador.

Para Docker, copiar `.env.docker.example` a `.env` en la raíz del repositorio.
Para ejecutar el backend con `go run`, copiar `backend/.env.example` a
`backend/.env`. Esos son los únicos archivos locales que reciben secretos.

## Verificación

```bash
npm test -- --watch=false
npm run build
```

Desarrollo y producción consumen `/api/v1`. En producción, el servidor o proxy inverso debe publicar el backend bajo esa ruta.
