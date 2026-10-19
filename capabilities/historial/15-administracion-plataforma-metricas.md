# Capability 15: Administración de plataforma y métricas de uso

## Estado

`En implementación`

Aprobada explícitamente por el usuario el 2026-10-15 mediante la instrucción: “Implement the plan”. La implementación comenzó en la misma fecha.

## Objetivo

Crear un backoffice global, separado del RBAC de cada negocio, para que cuentas operativas sin tiendas consulten cuántos negocios, sucursales, empleados y cuentas existen, además de su actividad reciente.

## Alcance

- RBAC global con el rol inicial `ADMIN_LECTURA`, independiente de membresías y roles tenant.
- Comando interno para asignar, retirar y listar administradores de plataforma.
- Panel de solo lectura con resumen, listado de negocios y detalle de usuarios/sucursales.
- Conteos por sucursal separados en personal asignado y personal con cuenta activa.
- Actividad diaria agregada por usuario y negocio, sin IP, agente, ruta ni cuerpo.
- Acceso del panel sin contexto de negocio y redirección apropiada para administradores sin tiendas.

Quedan fuera mutaciones de cuentas, negocios, sucursales o roles tenant; suplantación de usuarios; acceso técnico a IP/agente; y métricas comerciales.

## Reglas y decisiones

- `usuarios` continúa siendo la cuenta global. El acceso de plataforma no crea membresías.
- Un rol de negocio nunca concede acceso global y el rol global nunca concede acceso a datos tenant fuera de las APIs del backoffice.
- La asignación `ADMIN_LECTURA` exige cuenta activa, correo verificado y cero membresías de negocio.
- “Usuario activo” es una cuenta distinta con actividad autenticada exitosa en un negocio dentro del periodo.
- Los propietarios se reportan a nivel negocio como acceso total; no se duplican en cada sucursal.
- La actividad se atribuye a negocio, no a sucursal. Los agregados se conservan 13 meses.
- El panel expone únicamente nombre, correo, estado, roles, sucursales y último acceso.

## Modelo de datos

- `roles_plataforma`: catálogo global sembrado; código único, nombre, descripción y estado.
- `roles_usuario_plataforma`: asignación única usuario/rol con fecha y origen.
- `auditoria_plataforma`: historial de altas y retiros del rol, motivo, objetivo, origen y fecha.
- `actividad_usuario_diaria`: único por fecha/usuario/negocio; primera actividad, última actividad y solicitudes.

## Contratos

- `GET /api/v1/auth/me` añade `roles_plataforma: string[]`.
- `GET /api/v1/plataforma/resumen?desde=YYYY-MM-DD&hasta=YYYY-MM-DD`.
- `GET /api/v1/plataforma/negocios?q=&estado=&page=&page_size=`.
- `GET /api/v1/plataforma/negocios/:negocioId`.
- `GET /api/v1/plataforma/negocios/:negocioId/usuarios`.
- Comando: `admin-plataforma grant|revoke|list`, con `--email` y `--reason` en mutaciones.

Todos los endpoints de plataforma exigen sesión válida y `ADMIN_LECTURA`. Las listas usan `{items, total, page, page_size}` y las fechas se guardan en UTC.

## Pantallas

- `/administracion-plataforma`: resumen y actividad del periodo.
- `/administracion-plataforma/negocios`: búsqueda, estado y paginación.
- `/administracion-plataforma/negocios/:negocioId`: sucursales, personal, cuentas y actividad.

El shell es independiente del selector y guard de contexto tenant.

## Pruebas y criterios de aceptación

- Migración y sembrado idempotentes.
- Comando idempotente y rechazo de cuenta inexistente, no verificada o con membresías.
- Middleware devuelve `401`, `403` y `200` según sesión y rol global.
- Conteos distinguen asignaciones finalizadas, empleados sin cuenta y cuentas inactivas.
- Actividad deduplica usuarios en ventanas de 7/30 días y no almacena datos técnicos sensibles.
- Un administrador sin tiendas abre directamente el panel y no puede consumir endpoints tenant.
- Pruebas Go, `go vet`, pruebas Angular y build correctos.

## Seguimiento

- [x] Alcance y decisiones aprobados.
- [x] Capability registrada e indexada.
- [x] Modelo, migración y comando.
- [x] API, autorización y actividad.
- [x] Panel Angular y navegación.
- [x] Pruebas y verificaciones parciales.
- [x] Actualización de graphify.

## Resultados de verificación

- Backend nuevo compilado y pruebas focalizadas correctas con `GOCACHE=/private/tmp/tienda-go-cache`.
- `tsc -p tsconfig.app.json --noEmit` correcto; las pruebas nuevas del panel (5) y rutas (5) pasan. La suite completa mantiene seis fallos preexistentes de `alta-inicial` relacionados con `window.matchMedia` y la expectativa de flecha visual.
- `npm run build` queda pendiente de un runner estable: Node 14 es incompatible y Node 22.23.2 aborta durante el bundle optimizado con código 134 en este entorno.
