# Fase 7: Asignaciones empleado-sucursal

## Estado de fase

`Verificada`

Idea preliminar registrada el 2026-09-12. Implementación autorizada explícitamente por el usuario el 2026-09-12, para ejecución no supervisada.

## Fuente de verdad y compatibilidad

`base.MD` define `asignaciones_empleado_sucursal`: `negocio_id`, `empleado_id`, `sucursal_id`, `es_principal`, `activo`, `asignado_en`, `finalizado_en`; único `(empleado_id, sucursal_id)`. Tabla nueva, sin equivalente legacy. Depende de `empleados` (fase 5) y `sucursales` (fase 2).

## Objetivo y alcance (idea)

Relación muchos-a-muchos entre empleado y sucursal, con una sucursal principal por empleado, para saber dónde opera cada persona.

Incluye (idea):

- Asignar/retirar sucursales a un empleado desde su detalle (tab "Sucursales asignadas").
- Marcar una asignación como principal; invariante: máximo una principal activa por empleado, igual patrón que sucursal principal por negocio en fase 2.
- Validar que la sucursal pertenezca al mismo `negocio_id` que el empleado.
- Finalización lógica de una asignación (`activo=false`, `finalizado_en`) sin borrar historial.

No incluye (idea):

- Restringir el contexto activo (fase 3) o el login a las sucursales asignadas; queda para una capability futura.
- Horarios, turnos o disponibilidad por sucursal.
- Cambiar el aislamiento de ventas/productos por sucursal más allá del dato de asignación.

## Organización backend prevista

Dominio `negocio`: `domain/negocio/asignacion_empleado_sucursal.go`, `application/negocio/asignacion_empleado_sucursal_service.go`, `infrastructure/negocio/asignacion_empleado_sucursal_repository.go` + migración, `interfaces/http/negocio/asignacion_empleado_sucursal_handler.go`.

## Implementación de fase 7

Fecha: 2026-09-12.

### Backend

```text
backend/internal/domain/negocio/asignacion_empleado_sucursal.go                   (nuevo)
backend/internal/application/negocio/asignacion_empleado_sucursal_service.go      (nuevo)
backend/internal/application/negocio/asignacion_empleado_sucursal_service_test.go (nuevo)
backend/internal/infrastructure/negocio/asignacion_empleado_sucursal_repository.go (nuevo)
backend/internal/infrastructure/negocio/asignacion_empleado_sucursal_integration_test.go (nuevo)
backend/internal/interfaces/http/negocio/asignacion_empleado_sucursal_handler.go  (nuevo)
backend/internal/database/database.go                                             (modelo en AutoMigrate)
backend/cmd/api/main.go                                                           (composición y rutas)
```

Tabla nueva sin forma legacy: se crea desde el struct, sin migración de convergencia.

### API

```text
GET    /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId/sucursales?estado=
POST   /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId/sucursales
POST   /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId/principal
DELETE /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId/sucursales/:asignacionId
```

Autorización: `equipo.asignaciones.ver` para leer y `equipo.asignaciones.editar` para modificar.

### Reglas de principal

Se replica el criterio ya usado para la sucursal principal del negocio en fase 2, ahora por empleado:

- La primera asignación activa se vuelve principal automáticamente.
- Promover degrada a la anterior dentro de la misma transacción.
- Retirar la principal promueve a la asignación activa más antigua.
- Sin asignaciones activas el empleado simplemente queda sin principal.
- Reasignar una sucursal retirada reactiva la fila existente en lugar de duplicarla, porque el único es `(empleado_id, sucursal_id)`.

### Aislamiento

Una sucursal de otro negocio se rechaza antes de tocar la tabla y el handler responde `404`, indistinguible de una sucursal inexistente. El empleado también se valida contra el negocio de la ruta antes de exponer o modificar sus asignaciones.

### Frontend

```text
frontend/src/app/features/equipo/asignacion.models.ts
frontend/src/app/features/equipo/asignacion.service.ts
frontend/src/app/features/equipo/asignacion.service.spec.ts
frontend/src/app/features/equipo/empleado-sucursales.component.{ts,html,css}
frontend/src/app/features/equipo/equipo.routes.ts
```

El selector de sucursales se alimenta del contexto activo de fase 3 y excluye las ya asignadas.

## Verificación técnica de fase 7

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test ./...`: correctos.
- Prueba de integración PostgreSQL sobre base desechable `tienda_fase7_test`, con el índice único parcial activo: primera asignación principal automática, segunda como secundaria sin romper la invariante, promoción que degrada la anterior, retiro de la principal que promueve a otra activa, reactivación sin duplicar fila y estado final sin principal cuando no queda ninguna activa.
- `asignacion_empleado_sucursal_service_test.go`: 9 casos. Cubren permiso de lectura y de edición, sucursal de otro negocio, sucursal obligatoria, duplicado activo, propagación de la solicitud de principal, empleado ajeno y negocio archivado.
- `npm test -- --watch=false`: correcto, 32 archivos y 106 pruebas.
- `npm run build`: correcto.

## Imprevistos resueltos de fase 7

- La prueba de integración descubrió un defecto real y latente: `Pluck` de GORM sobre un destino `uuid.UUID` falla al escanear con el error `converting driver.Value type string ... to a uint8`. Afectaba a `asegurarPrincipal` de esta fase y también a `QuedaPropietarioConRolSistema` de fase 4, que nunca se había ejercitado contra PostgreSQL. Ambos lugares ahora leen `id::text` y convierten explícitamente.

---
