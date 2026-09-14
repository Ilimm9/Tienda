# Fase 6: Invitaciones

## Estado de fase

`Verificada`

Idea preliminar registrada el 2026-09-12. Implementación autorizada explícitamente por el usuario el 2026-09-12, para ejecución no supervisada.

## Fuente de verdad y compatibilidad

`base.MD` define `invitaciones_negocio`: `negocio_id`, `empleado_id` opcional, `correo`, `rol_predeterminado_id`, `hash_token` único, `estado`, `invitado_por_usuario_id`, `expira_en`, `aceptado_por_usuario_id/aceptado_en`. Tabla nueva, sin equivalente legacy.

Decisiones generales ya aprobadas que aplican directamente: registrar empleado antes de invitarlo (fase 5); entrega solo por enlace copiable, sin integración de correo; si el invitado no tiene cuenta, completa registro y continúa aceptación con el mismo correo.

## Objetivo y alcance (idea)

Invitar a un empleado ya registrado a obtener acceso al negocio, mediante enlace con token, y resolver su aceptación creando o reutilizando membresía y roles.

Incluye (idea):

- Acción "Invitar" desde detalle de empleado (fase 5): genera token, guarda solo su hash, entrega enlace copiable con expiración (p. ej. 7 días).
- Estados: `pendiente`, `aceptada`, `expirada`, `cancelada`. Reenviar genera nuevo token e invalida el anterior.
- Pantalla pública de aceptación por token: si el correo no tiene cuenta, exige registro con ese correo antes de continuar; si ya tiene cuenta, exige iniciar sesión con ese correo.
- Aceptación transaccional: crea o reactiva `membresia_negocio` (tipo `miembro`), enlaza `empleados.membresia_id`, asigna `rol_predeterminado_id` vía `roles_membresia` (fase 4), marca invitación `aceptada`.
- Pantallas reales en `equipo/invitaciones` (hoy placeholder): listado por estado, generar/copiar enlace, cancelar, reenviar.

No incluye (idea):

- Envío de correo real; solo enlace copiable.
- Invitar a alguien sin registro previo como empleado (fase 5 es requisito).
- Permitir aceptación con correo distinto al invitado.

## Organización backend prevista

Dominio `negocio`: `domain/negocio/invitacion.go`, `application/negocio/invitacion_service.go`, `infrastructure/negocio/invitacion_repository.go` + migración, `interfaces/http/negocio/invitacion_handler.go`. Ruta pública de aceptación separada de rutas administrativas autenticadas.

## Implementación de fase 6

Fecha: 2026-09-12.

### Backend

```text
backend/internal/domain/negocio/invitacion.go                      (nuevo)
backend/internal/application/negocio/invitacion_service.go         (nuevo)
backend/internal/application/negocio/invitacion_service_test.go    (nuevo)
backend/internal/infrastructure/negocio/invitacion_repository.go   (nuevo)
backend/internal/interfaces/http/negocio/invitacion_handler.go     (nuevo)
backend/internal/database/database.go                              (modelo en AutoMigrate)
backend/cmd/api/main.go                                            (composición y rutas)
```

La tabla `invitaciones_negocio` es nueva y no tiene forma legacy, por lo que se crea con AutoMigrate a partir del struct; no requiere una migración de convergencia como fases 4 y 5.

### API

```text
GET    /api/v1/negocios/:negocioId/administracion/invitaciones?estado=
POST   /api/v1/negocios/:negocioId/administracion/invitaciones
DELETE /api/v1/negocios/:negocioId/administracion/invitaciones/:invitacionId
GET    /api/v1/invitaciones/:token                (público)
POST   /api/v1/invitaciones/:token/aceptar        (requiere sesión)
```

Autorización: `equipo.invitaciones.ver` para leer y `equipo.invitaciones.enviar` para emitir o cancelar.

### Manejo del token

- Se genera con `crypto/rand`, 32 bytes, codificado en base64 URL-safe.
- Solo se persiste su SHA-256 en `hash_token`; el valor en claro se devuelve una única vez, en la respuesta de creación.
- Vigencia de 7 días mediante `domain.DuracionInvitacion`.
- Emitir una invitación nueva cancela automáticamente las pendientes del mismo empleado, de modo que un enlace anterior deja de servir.
- El listado normaliza a `expirada` lo vencido antes de responder, para no mostrar como pendiente algo que ya no funciona.

### Aceptación transaccional

Dentro de una sola transacción se crea o reactiva la membresía como `miembro` activo, se vincula `empleados.membresia_id` con estado `activo`, se asigna el rol predeterminado mediante `roles_membresia` y se marca la invitación como `aceptada`. Reactivar una membresía existente no degrada el `tipo_miembro` de un propietario.

El correo se compara normalizado: el usuario en sesión debe ser el mismo correo invitado, o la respuesta es `403`.

### Frontend

```text
frontend/src/app/features/equipo/invitacion.models.ts
frontend/src/app/features/equipo/invitacion.service.ts
frontend/src/app/features/equipo/invitacion.service.spec.ts
frontend/src/app/features/equipo/invitaciones.component.{ts,html,css}
frontend/src/app/features/invitacion/aceptar-invitacion.component.{ts,html,css}
frontend/src/app/app.routes.ts                    (ruta pública /invitacion/:token)
```

`equipo/invitaciones` dejó de ser placeholder. El selector de empleados solo ofrece a quienes ya están registrados, tienen correo y aún no tienen cuenta vinculada, conforme a la decisión aprobada de registrar antes de invitar.

La ruta `/invitacion/:token` vive fuera del shell autenticado, porque el invitado puede no tener cuenta; cuando falta, la pantalla dirige a registro conservando correo y retorno. Al aceptar se recarga el contexto de fase 3 para que el negocio nuevo aparezca de inmediato en el selector.

## Verificación técnica de fase 6

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test ./...`: correctos.
- `invitacion_service_test.go`: 13 casos. Cubren permiso de envío, que solo se persista el hash y no el token, cancelación de pendientes previas al reemitir, vigencia inicial, empleado sin correo, empleado ya vinculado, rol de otro negocio, negocio archivado, correo distinto al invitado, correo coincidente con distinta capitalización y espacios, token expirado, token ya aceptado, token vacío, señal de cuenta faltante y unicidad del token generado.
- `npm test -- --watch=false`: correcto, 31 archivos y 101 pruebas.
- `npm run build`: correcto.

## Imprevistos resueltos de fase 6

- El listado podía mostrar como `pendiente` una invitación ya vencida. Se normaliza el estado en la base antes de listar, en lugar de calcularlo solo en la interfaz.

---
