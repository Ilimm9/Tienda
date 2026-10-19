# Gestión de colaboradores

Base vigente según código local. Actualizado: 2026-10-19.

## Flujo y reglas

1. El propietario o un miembro autorizado registra un empleado dentro del negocio; puede existir sin cuenta de acceso.
2. Para invitarlo se eligen empleado con correo, sucursal y rol disponibles del mismo negocio. El enlace vence a las 72 horas y solo se persiste su hash.
3. El destinatario inicia sesión o registra/verifica una cuenta con ese correo. La aceptación vincula empleado y membresía, otorga rol y asigna sucursal en una transacción.
4. Reenviar reemplaza el token anterior; cancelar impide aceptar. Suspender/terminar personal y finalizar asignaciones afectan su acceso operativo.

**Cuenta** identifica a la persona que inicia sesión; **membresía** permite entrar al negocio; **empleado** es su ficha laboral; **asignación** habilita sucursales. El rol aporta permisos mediante la membresía. El propietario tiene acceso total al negocio; un miembro opera según permisos y sucursales autorizadas.

API e interfaz comprueban permisos; ocultar un botón no basta. No se pueden mezclar empleados, roles ni sucursales de otros negocios. Los roles de sistema y la continuidad de un propietario efectivo tienen protección específica. El rol global `ADMIN_LECTURA` pertenece al módulo de plataforma y no concede acceso operativo a tiendas.

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `usuarios`, `perfil_usuarios` | Cuenta y datos personales; ver [registro](flujo-registro.md). |
| `membresias_negocio` | Vínculo único de cuenta y negocio, tipo de miembro y estado. |
| `empleados` | Ficha del negocio; `membresia_id` opcional hasta vincular cuenta. |
| `invitaciones_negocio` | Correo, empleado, sucursal, rol, hash del token y estado. |
| `asignaciones_empleado_sucursal` | Empleado/sucursal, principal, vigencia y estado. |
| `roles`, `permisos`, `permisos_rol`, `roles_membresia` | Roles del negocio, permisos y asignaciones a membresías. |
| `negocios`, `sucursales` | Límites de pertenencia y operación. |

## Dónde cambiar y qué revisar

- Backend: [servicios de negocio](../backend/internal/application/negocio/), [persistencia](../backend/internal/infrastructure/negocio/) y [handlers](../backend/internal/interfaces/http/negocio/).
- Frontend: [invitaciones](../frontend/src/app/features/invitacion/) y [permisos/contexto](../frontend/src/app/contexto/); rutas registradas en [app.routes.ts](../frontend/src/app/app.routes.ts).
- Comprobar aislamiento entre negocios, correo destinatario, token reemplazado, aceptación repetida, rollback y revocación. Revisiones aún abiertas en [planes actuales](planes-actuales.md).
- Antecedentes: [identidad y RBAC](historial/01-identidad-negocio-sucursales-rbac.md), [invitaciones](historial/02-thrs-invitaciones-empleados-sucursal.md) y [permisos y formularios](historial/05-thrs-permisos-frontend-formularios-equipo.md).
