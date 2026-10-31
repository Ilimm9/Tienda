# Gestión de colaboradores

Base vigente según código local. Actualizado: 2026-10-22.

## Flujo y reglas

1. El propietario o un miembro autorizado registra un empleado dentro del negocio; puede existir sin cuenta de acceso.
2. Para invitarlo se eligen empleado con correo, sucursal y rol disponibles del mismo negocio. El enlace vence a las 72 horas y solo se persiste su hash.
3. El destinatario inicia sesión o registra/verifica una cuenta con ese correo. La aceptación vincula empleado y membresía, otorga rol y asigna sucursal en una transacción.
4. Reenviar reemplaza el token anterior; cancelar impide aceptar. Emisión, reenvío y aceptación comprueban que el emisor conserve acceso y pueda otorgar el rol actual; un rechazo no deja vínculos parciales.

**Cuenta** identifica a la persona que inicia sesión; **membresía** permite entrar al negocio; **empleado** es su ficha laboral; **asignación** habilita sucursales. El rol aporta permisos mediante la membresía. El propietario tiene acceso total al negocio; un miembro opera según permisos y sucursales autorizadas.

API e interfaz comprueban permisos; ocultar un botón no basta. No se pueden mezclar empleados, roles ni sucursales de otros negocios. Los roles de sistema y la continuidad de un propietario efectivo tienen protección específica. El rol global `ADMIN_LECTURA` pertenece al módulo de plataforma y no concede acceso operativo a tiendas.

Al crear un negocio se siembran el rol de sistema Propietario y el rol editable **Cajero** (`catalogo.ver`, `compras.ver`, `compras.registrar`); junto con el puesto **Cajero** que lo sugiere ([33](historial/33-puesto-cajero-y-crear-desde-selector.md)); ninguno se vuelve a sembrar si lo borran. El selector ofrece solo roles delegables y resume permisos por sección. Crear o editar permisos agrega las lecturas necesarias; no puede otorgar permisos ajenos al emisor. Roles anteriores incompletos muestran lecturas faltantes y personas afectadas: corregirlos exige confirmación específica. Editar solo el nombre no amplía accesos.

Suspensión o baja ocultan el negocio y todas sus asignaciones al empleado y bloquean API, conservando historial para administradores. Suspensión recupera accesos aún vigentes; baja exige reasignar roles/sucursales explícitos desde **Reactivar y reasignar acceso**. No restaura una membresía revocada por otra causa ni permite suspender al último propietario. Sin cuenta vinculada puede prepararse otra invitación, sin conceder acceso hasta aceptarla.

Alta en 3 pasos ([30 · Rediseño de Equipo](historial/30-rediseno-equipo.md)): datos → sucursales (la del contexto propuesta como principal) → invitación con la principal y el rol sugerido por el puesto. Se puede volver al paso 1 (actualiza, no duplica) o al 2; sin sucursal o sin correo no se invita. Teléfono: 10 dígitos sin lada (API los exige). Número de empleado: botón «Generar» `S`+iniciales de la sucursal + consecutivo (`SC-003`), editable y único. Puesto: `empleados.puesto_id` → `puestos_negocio` (catálogo común a todas las sucursales, sin conteos); al asignar debe estar activo, conservar uno inactivo se permite. Invitaciones muestra solo la sucursal del contexto y pagina de 10 en 10. Los selectores de puesto y de rol (alta de empleado, nueva invitación, alta de puesto) terminan en «+ Crear nuevo…», visible con `equipo.empleados.gestionar` o `roles.gestionar`, que lleva a Puestos o a Nuevo rol; con datos sin guardar en el empleado se pide confirmación.

El contexto se recarga tras cambiar roles/asignaciones y al volver a la pestaña; no hay sincronización instantánea mientras otra sesión permanece en primer plano. Implementación y pruebas: [20 · Arreglos de invitaciones](historial/20-arreglos-invitaciones.md).

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `usuarios`, `perfil_usuarios` | Cuenta y datos personales; ver [registro](flujo-registro.md). |
| `membresias_negocio` | Vínculo único de cuenta y negocio, tipo de miembro y estado. |
| `empleados` | Ficha del negocio; `membresia_id` opcional hasta vincular cuenta; `puesto_id` opcional (FK a `puestos_negocio`). |
| `invitaciones_negocio` | Correo, empleado, sucursal, rol, hash del token y estado. |
| `asignaciones_empleado_sucursal` | Empleado/sucursal, principal, vigencia y estado. |
| `puestos_negocio` | Catálogo de puestos del negocio: nombre único sin distinguir mayúsculas, rol sugerido opcional, activo. |
| `roles`, `permisos`, `permisos_rol`, `roles_membresia` | Roles del negocio, permisos y asignaciones a membresías. |
| `negocios`, `sucursales` | Límites de pertenencia y operación. |

## Dónde cambiar y qué revisar

- Backend: [servicios de negocio](../backend/internal/application/negocio/), [persistencia](../backend/internal/infrastructure/negocio/) y [handlers](../backend/internal/interfaces/http/negocio/).
- Frontend: [equipo](../frontend/src/app/features/equipo/), [invitaciones](../frontend/src/app/features/invitacion/) y [permisos/contexto](../frontend/src/app/contexto/); rutas registradas en [app.routes.ts](../frontend/src/app/app.routes.ts).
- Comprobar aislamiento entre negocios, correo destinatario, token reemplazado, aceptación repetida, rollback y revocación. Revisiones aún abiertas en [planes actuales](planes-actuales.md).
- Antecedentes: [identidad y RBAC](historial/01-identidad-negocio-sucursales-rbac.md), [invitaciones](historial/02-thrs-invitaciones-empleados-sucursal.md) y [permisos y formularios](historial/05-thrs-permisos-frontend-formularios-equipo.md).
