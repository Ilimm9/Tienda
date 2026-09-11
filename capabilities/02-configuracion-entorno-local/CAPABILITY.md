# Capability 02: Configuración del entorno local

## Estado

`Verificada`

## Control

- Responsable: Usuario del proyecto
- Fecha de creación: 2026-09-11
- Aprobación de especificación e implementación: Confirmada explícitamente el 2026-09-11 mediante “Implement the plan”
- Dependencias: configuración actual de Go, Angular CLI y proxy de desarrollo

## Objetivo

Permitir que backend y frontend cambien sus puertos locales desde `backend/.env`, conservando `8080` para Go y `4200` para Angular como valores predeterminados y evitando problemas de CORS durante el desarrollo mediante un proxy de Angular.

## Alcance

- Mantener `backend/.env` como fuente local ignorada por Git.
- Documentar `APP_PORT`, `FRONTEND_PORT` y `FRONTEND_URL` en `backend/.env.example`.
- Iniciar Angular en `FRONTEND_PORT` mediante `npm start`.
- Redirigir `/api/*` desde Angular hacia Go usando `APP_PORT`.
- Usar `/api/v1` como URL del frontend en desarrollo y producción.
- Conservar el middleware CORS para clientes que accedan directamente al backend.
- No modificar automáticamente el archivo local `backend/.env`.

## Reglas y contratos

- `APP_PORT` acepta un puerto TCP entre 1 y 65535; su valor predeterminado es `8080`.
- `FRONTEND_PORT` acepta un puerto TCP entre 1 y 65535; su valor predeterminado es `4200`.
- `FRONTEND_URL` conserva el origen exacto permitido por CORS; su valor predeterminado es `http://localhost:4200`.
- El frontend consume siempre `/api/v1`, sin construir URLs locales absolutas.
- El proxy solo aplica a `ng serve`; la compilación de producción mantiene rutas relativas.
- El lanzador falla con un mensaje claro si un puerto configurado es inválido.
- Los argumentos adicionales de `npm start -- ...` se transmiten a Angular CLI.

## Flujo

1. Go carga `backend/.env` y escucha en `APP_PORT`.
2. `npm start` carga el mismo archivo y arranca Angular en `FRONTEND_PORT`.
3. El navegador solicita `/api/v1/...` al origen de Angular.
4. Angular reenvía `/api/*` a `http://localhost:APP_PORT`.
5. Las solicitudes directas al backend continúan sujetas a `FRONTEND_URL` y CORS.

## Pruebas

- Pruebas Go completas y de middleware CORS.
- Pruebas Angular completas y build de producción.
- Validación del lanzador con puertos predeterminados, personalizados e inválidos.
- Smoke de `/api/v1/health` a través del proxy con puertos personalizados disponibles.
- Smoke de una solicitud autenticada y `PATCH` cuando la base de datos local esté disponible.

## Criterios de aceptación

- Sin variables locales, Go usa `8080` y Angular `4200`.
- Con `APP_PORT=8081` y `FRONTEND_PORT=4201`, ambos servicios arrancan en esos puertos sin editar TypeScript.
- Las solicitudes del frontend llegan al backend mediante `/api/v1` sin errores CORS.
- Login, cookies y métodos mutables atraviesan el proxy sin cambiar sus contratos HTTP.
- `.env.example`, el README y el comportamiento ejecutable describen los mismos valores predeterminados.
- El `.env` real y los cambios no relacionados del usuario permanecen intactos.

## Registro de ejecución

- 2026-09-11: plan aprobado e implementación iniciada.
- 2026-09-11: `go test ./...` y `go vet ./...` completados correctamente.
- 2026-09-11: 25 archivos y 62 pruebas Angular completados correctamente.
- 2026-09-11: build de producción completado; conserva advertencias conocidas de presupuesto de bundle y CSS.
- 2026-09-11: lanzador y proxy validados con `APP_PORT=18080` y `FRONTEND_PORT=14200`; `/api/v1/health` respondió `200` a través de Angular.
- 2026-09-11: cookie `HttpOnly` y solicitud `PATCH` atravesaron correctamente el proxy.
- 2026-09-11: puertos inválidos rechazados y build confirmado sin URL absoluta local de API.
- Pendiente: aceptación final del usuario para cerrar la capability.
