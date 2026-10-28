# Capability 36: El administrador de plataforma nunca entra al asistente ni al shell de tiendas

- Estado: **En implementación**.
- Aprobación explícita: 2026-10-26, mediante la aprobación del plan en Claude Code. El usuario lo reportó así: su cuenta de administrador cae en `/configuracion-inicial` al entrar.
- Fecha: 2026-10-26.
- Base: [15 · Administración de plataforma](historial/15-administracion-plataforma-metricas.md) y [35](35-plataforma-cuentas-tiendas.md). Frontend `features/plataforma` y `app.routes.ts`.

## Problema

Solo el login manda a `ADMIN_LECTURA` a `/administracion-plataforma`. Si el administrador abre `/` o `/inicio`, recarga la página o usa un marcador, `altaInicialGuard` lo trata como una cuenta sin negocios (`requiere_empresa`) y lo manda al asistente para crear una tienda. Un administrador de plataforma no puede tener membresías, y el backend ya lo bloquea con `RejectPlatformAdmin`.

## Alcance y regla

- Una cuenta con `ADMIN_LECTURA` que intente entrar a `/configuracion-inicial` o a cualquier ruta del shell de tiendas se redirige a `/administracion-plataforma`.
- Las demás cuentas no cambian: una cuenta sin tienda sigue viendo el asistente.
- No se modifican el backend, `altaInicialGuard`, `ContextoService` ni el login.

## Implementación

- `fueraDePlataformaGuard` en `features/plataforma/plataforma.guard.ts` usa `auth.me()`. Ante un error devuelve `true`, porque `authGuard` ya se encarga de mandar a `/login`.
- `app.routes.ts`:
  - `configuracion-inicial` usa `[authGuard, fueraDePlataformaGuard, asistenteGuard]`;
  - el shell de tiendas agrega `canActivate: [fueraDePlataformaGuard]`, que se evalúa una vez al entrar al shell.

## Pruebas y criterios de aceptación

- Unitarias del guard: con `ADMIN_LECTURA` redirige al panel; una cuenta de negocio pasa; sin sesión pasa y deja decidir a `authGuard`.
- Manual: con la cuenta administradora, `/`, `/inicio`, `/configuracion-inicial` y `/catalogo/productos` terminan en `/administracion-plataforma`. Con una cuenta normal sin tienda, `/inicio` sigue llevando al asistente.

## Tareas

- [x] Capability e índices.
- [x] Guard, rutas y pruebas.
- [ ] Prueba manual y aceptación final del usuario.

## Resultados de verificación (2026-10-26)

- Angular (Node 22.23.2): las specs de `features/plataforma`, `contexto` y `features/auth` pasan: 16 archivos, 78 de 78 pruebas, incluidas las 3 nuevas de `fueraDePlataformaGuard`.
- `ng build --configuration development`: correcto.
- Prueba manual en navegador: pendiente del usuario.
