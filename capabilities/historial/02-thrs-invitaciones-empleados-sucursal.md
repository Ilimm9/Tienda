# Plan 2: Invitar empleados a una sucursal

Resumen al 2026-10-04. Versión completa anterior en `.respaldo-2026-10-04/`.

## Estado

- `Aprobada` el 2026-10-01 para el **flujo de invitación**. Implementado ese día.
- Lo que quedó fuera de esa entrega continúa en el [plan 5](05-thrs-permisos-frontend-formularios-equipo.md).
- Dependencias: empleados, RBAC e invitaciones de capability 01; correo de 07.

## Objetivo

El administrador elige empleado, sucursal y rol, y comparte un enlace. El empleado completa su cuenta, verifica su correo y entra solo a la sucursal indicada, sin crear empresa.

## Decisiones confirmadas

- Personalización desde el enlace, conservando OTP. Cuenta existente con el mismo correo inicia sesión.
- Vigencia de 72 horas por emisión o reenvío, calculada en servidor.
- Una sucursal por invitación. Para cambiar sucursal o rol: cancelar y emitir otra.
- Rol `PROPIETARIO` no delegable. Membresía suspendida o revocada no se reactiva por enlace.
- **Tablas nuevas: ninguna.** `invitaciones_negocio` ganó `sucursal_id`.

## Implementado

- Backend `negocio`: `sucursal_id`, rol y sucursal obligatorios, filtros `sucursal_id` y `sin_aceptar`, estado efectivo por reloj del servidor.
- Migración `MigrateInvitacionSucursal`: columna, FK compuesta `(negocio_id, sucursal_id)`, índice único de pendiente por empleado; cancela pendientes antiguas sin sucursal.
- Rutas: `POST …/invitaciones/:invitacionId/reenviar` (`correo` opcional), `GET /invitaciones/:token`, `POST /invitaciones/:token/registro`, `POST /invitaciones/:token/aceptar`.
- Aceptación atómica: bloqueo de empleado e invitación, revalidación en transacción, membresía, rol y asignación de sucursal.
- Reenvío invalida el enlace anterior; corregir correo conserva el histórico y exige además `equipo.empleados.gestionar`.
- Límite de frecuencia por IP en consulta y registro públicos. `cuenta.RegisterWithNames` como contrato entre dominios.
- Frontend: `equipo/invitaciones/` (lista, nueva, corregir correo) y `invitacion/aceptar-invitacion`; retorno a la invitación en `sessionStorage` durante login y OTP.
- Aspecto actual: opción G (plan 4) con Tailwind (capability 05).

## Pruebas ejecutadas

- `go vet` limpio; `go test ./...` pasa salvo `internal/config` (`TestProductionAcceptsStrongOTPSecret`, fallo previo).
- `ng build` correcto; `ng test` con 9 fallos previos en `productos.component.spec.ts` y `navigation.config.spec.ts`.
- Recorrido HTTP con Mailpit, 27 comprobaciones correctas, incluidas dos aceptaciones simultáneas (200 y 409).

## Pendientes resueltos en el plan 5 (2026-10-04)

- [x] Autorización por sucursal en contexto y API.
- [x] Permisos `precios.ver` y `precios.autorizar`.
- [x] El emisor solo delega roles cuyos permisos posee.
- [x] Permisos reflejados en el frontend, validaciones de formularios y mejora visual de Roles.
- [x] Correo de invitación indica 72 horas.
- [ ] Pruebas de integración con base (reenvío, aceptación) y de componentes Angular; SMTP caído, rollback, sucursal archivada y rol desactivado.
