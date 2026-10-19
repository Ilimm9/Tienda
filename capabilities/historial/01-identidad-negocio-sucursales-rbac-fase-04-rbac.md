# Fase 4: RBAC

## Estado de fase

`Verificada`

Idea preliminar registrada el 2026-09-12. Implementación autorizada explícitamente por el usuario el 2026-09-12 junto con las fases 5 a 8, para ejecución no supervisada.

Decisión del usuario del 2026-09-12: el modelo actual de roles debe **coincidir con `base.MD`**, no coexistir con una forma paralela.

## Fuente de verdad y compatibilidad

`base.MD` define `permisos`, `roles`, `permisos_rol` y `roles_membresia`. El backend contiene hoy una forma previa simplificada que diverge y debe converger.

### Estado real verificado del código (2026-09-12)

- `domain/negocio/rol.go`: `Rol{id, negocio_id, nombre varchar(100), descripcion, creado_en, actualizado_en}`.
- `domain/negocio/membresia_negocio.go`: incluye `rol_id *uuid` (rol único por membresía) y `empleado_id *uuid`.
- No existen `permisos` ni `permisos_rol`.
- `Rol` aparece únicamente en la lista de AutoMigrate de `internal/database/database.go`. Ningún service, repository, handler, seed ni prueba lo consume.

Consecuencia: la tabla `roles` es andamiaje vacío sin consumidores. Por eso esta fase **reemplaza su forma** para igualar `base.MD`, en lugar de mantener un puente legacy indefinido como se hizo con `sucursales` en fase 2. La migración sigue siendo idempotente y defensiva por si alguna base ya tiene filas.

### Divergencias a cerrar

| Tabla | Hoy | `base.MD` | Acción |
| ----- | --- | --------- | ------ |
| `roles` | sin `codigo` | `codigo varchar(60)` | agregar; backfill desde `nombre` normalizado; único `(negocio_id, codigo)` |
| `roles` | `nombre varchar(100)` | `nombre varchar(120)` | ampliar a 120; ampliar no es destructivo |
| `roles` | sin `es_rol_sistema` | `es_rol_sistema boolean not null default false` | agregar |
| `roles` | sin `activo` | `activo boolean not null default true` | agregar |
| `roles` | sin `creado_por_usuario_id` | `creado_por_usuario_id uuid` | agregar con FK a `usuarios` |
| `membresias_negocio` | `rol_id uuid` | no existe | backfillear a `roles_membresia` y eliminar columna |
| `permisos` | no existe | catálogo global | crear y sembrar |
| `permisos_rol` | no existe | `(rol_id, permiso_id)` único | crear |
| `roles_membresia` | no existe | `(membresia_negocio_id, rol_id)` único | crear |

`membresias_negocio.empleado_id` también sobra respecto a `base.MD`, pero su retiro pertenece a fase 5 porque la dirección correcta del vínculo es `empleados.membresia_id`.

### Excepción de tipo ya vigente

`base.MD` declara enums de PostgreSQL. El proyecto viene implementándolos como `varchar` con default y validación en service desde fases 1 y 2. Esta fase mantiene esa convención por consistencia; no es una divergencia introducida aquí y no cambia semántica ni nombres de valores.

## Modelo objetivo

### `permisos`

```text
id                uuid primary key
codigo            varchar(120) not null unique
codigo_modulo     varchar(30) not null
nombre            varchar(160) not null
descripcion       text null
creado_en         timestamptz not null
```

Catálogo global, no pertenece a ningún negocio y no es editable por el usuario. Se siembra desde código en cada arranque de forma idempotente por `codigo`. Módulos previstos: `negocios`, `sucursales`, `equipo`, `roles`, `catalogo`, `ventas`.

### `roles`

```text
id                     uuid primary key
negocio_id             uuid not null references negocios(id)
codigo                 varchar(60) not null
nombre                 varchar(120) not null
descripcion            text null
es_rol_sistema         boolean not null default false
activo                 boolean not null default true
creado_por_usuario_id  uuid null references usuarios(id)
creado_en              timestamptz not null
actualizado_en         timestamptz not null
```

Único case-insensitive `(negocio_id, lower(codigo))`, siguiendo el mismo criterio ya usado en `sucursales.codigo`.

### `permisos_rol`

```text
id          uuid primary key
rol_id      uuid not null references roles(id)
permiso_id  uuid not null references permisos(id)
creado_en   timestamptz not null
```

Único `(rol_id, permiso_id)`.

### `roles_membresia`

```text
id                       uuid primary key
membresia_negocio_id     uuid not null references membresias_negocio(id)
rol_id                   uuid not null references roles(id)
asignado_por_usuario_id  uuid null references usuarios(id)
asignado_en              timestamptz not null
```

Único `(membresia_negocio_id, rol_id)`. Nombre de columna conforme a `base.MD`: `membresia_negocio_id`, no `membresia_id`.

## Migración y backfill

Aditiva, transaccional e idempotente, en el estilo de `sucursal_migration.go`:

1. Ampliar `roles.nombre` a `varchar(120)`.
2. Agregar `codigo`, `es_rol_sistema`, `activo`, `creado_por_usuario_id` a `roles` con `IF NOT EXISTS`.
3. Backfillear `codigo` para filas existentes desde `nombre` normalizado en mayúsculas sin acentos; resolver colisiones por negocio con sufijo numérico, sin sobrescribir códigos ya presentes.
4. Establecer `codigo` obligatorio después del backfill; crear el único parcial case-insensitive.
5. Crear `permisos`, `permisos_rol` y `roles_membresia` si faltan.
6. Sembrar el catálogo `permisos` por `codigo`, sin borrar códigos desconocidos ya presentes.
7. Sembrar en cada negocio existente un rol de sistema `PROPIETARIO` (`es_rol_sistema = true`) con todos los permisos vigentes.
8. Asignar ese rol vía `roles_membresia` a toda membresía activa con `tipo_miembro = propietario`.
9. Backfillear `membresias_negocio.rol_id` no nulo hacia `roles_membresia`, sin duplicar filas.
10. Eliminar la columna `membresias_negocio.rol_id` y su campo en el struct, una vez backfilleada. Al no tener consumidores, no queda excepción pendiente para fase 8.

No borrar roles existentes, no regenerar códigos en ejecuciones posteriores y no reasignar roles ya asignados.

## Objetivo y alcance (idea)

Sustituir la autorización actual basada en `tipo_miembro = propietario` por permisos explícitos, permitiendo varios roles por membresía y roles personalizados por negocio.

Incluye (idea):

- CRUD de roles por negocio, con asignación de permisos del catálogo global.
- Asignar y quitar uno o varios roles a una membresía.
- Servicio que resuelve permisos efectivos de la membresía activa y expone verificación por código de permiso, más middleware que lo aplique en rutas administrativas.
- Reemplazar los checks `tipo_miembro = propietario` de fases 1 a 3 por checks de permiso, conservando la invariante de al menos una membresía propietaria activa por negocio.
- Rol de sistema `PROPIETARIO` no editable, no desactivable y no eliminable.
- Pantallas reales en `roles-permisos`, hoy placeholder en `features/roles-permisos/roles-permisos.routes.ts`: listado de roles, detalle y edición de permisos por rol, asignación de roles a miembros.

No incluye (idea):

- Permisos a nivel de sucursal o de registro individual.
- Compartir roles entre negocios; salvo el rol de sistema, todo rol pertenece a un negocio.
- Retirar `membresias_negocio.empleado_id`; corresponde a fase 5.

## Cadena de autorización prevista

```text
usuario -> membresía activa -> roles_membresia -> roles -> permisos_rol -> permisos
```

## Organización backend prevista

Dominio `negocio`, según convención de la capability:

```text
backend/internal/domain/negocio/rol.go            (reformado)
backend/internal/domain/negocio/permiso.go
backend/internal/application/negocio/rol_service.go
backend/internal/infrastructure/negocio/rol_repository.go
backend/internal/infrastructure/negocio/rol_migration.go
backend/internal/interfaces/http/negocio/rol_handler.go
backend/internal/interfaces/http/negocio/permiso_middleware.go
```

## Implementación de fase 4

Fecha: 2026-09-12.

### Backend

```text
backend/internal/domain/negocio/permiso.go                        (nuevo: catálogo y códigos)
backend/internal/domain/negocio/rol.go                            (reformado a base.MD)
backend/internal/domain/negocio/membresia_negocio.go              (sin rol_id)
backend/internal/application/negocio/rol_service.go               (nuevo)
backend/internal/application/negocio/rol_service_test.go          (nuevo)
backend/internal/application/negocio/sucursal_service.go          (autorización por permiso)
backend/internal/infrastructure/negocio/rol_migration.go          (nuevo)
backend/internal/infrastructure/negocio/rol_repository.go         (nuevo)
backend/internal/infrastructure/negocio/rol_migration_integration_test.go (nuevo)
backend/internal/infrastructure/negocio/sucursal_repository.go    (expone PermisosEfectivos)
backend/internal/interfaces/http/negocio/rol_handler.go           (nuevo)
backend/internal/interfaces/http/negocio/permiso_middleware.go    (nuevo)
backend/internal/interfaces/http/negocio/permiso_middleware_test.go (nuevo)
backend/internal/database/database.go                             (MigratePhaseFour y modelos)
backend/cmd/api/main.go                                           (composición y rutas)
```

`MigratePhaseFour` se ejecuta antes y después de `AutoMigrate`, igual que la migración de fase 2, para cubrir bases existentes y bases nuevas.

### API

```text
GET    /api/v1/negocios/:negocioId/administracion/permisos
GET    /api/v1/negocios/:negocioId/administracion/mis-permisos
GET    /api/v1/negocios/:negocioId/administracion/roles
POST   /api/v1/negocios/:negocioId/administracion/roles
GET    /api/v1/negocios/:negocioId/administracion/roles/:rolId
PATCH  /api/v1/negocios/:negocioId/administracion/roles/:rolId
DELETE /api/v1/negocios/:negocioId/administracion/roles/:rolId
GET    /api/v1/negocios/:negocioId/administracion/miembros
PUT    /api/v1/negocios/:negocioId/administracion/miembros/:membresiaId/roles
```

`PUT` de roles reemplaza el conjunto completo de la membresía; no acumula asignaciones.

### Frontend

```text
frontend/src/app/features/roles-permisos/rol.models.ts
frontend/src/app/features/roles-permisos/rol.service.ts
frontend/src/app/features/roles-permisos/rol.service.spec.ts
frontend/src/app/features/roles-permisos/roles.component.{ts,html,css}
frontend/src/app/features/roles-permisos/rol-form.component.{ts,html,css}
frontend/src/app/features/roles-permisos/miembro-roles.component.{ts,html,css}
frontend/src/app/features/roles-permisos/roles-permisos.routes.ts
```

La sección dejó de ser placeholder. Las rutas exigen contexto activo, y los botones de gestión y asignación aparecen según los permisos efectivos devueltos por `mis-permisos`; el backend vuelve a validarlos.

### Sustitución de la regla temporal de fases 1 a 3

`SucursalService.acceso` ya no compara `tipo_miembro = propietario`. Ahora exige negocio activo y el permiso correspondiente:

| Operación | Permiso |
| --------- | ------- |
| Crear sucursal | `sucursales.crear` |
| Editar o promover principal | `sucursales.editar` |
| Archivar y restaurar | `sucursales.archivar` |
| Listar, detalle | solo membresía activa |

Como el rol de sistema `PROPIETARIO` concentra todos los permisos y la migración lo asigna a toda membresía propietaria activa, el comportamiento observable para propietarios no cambia; lo que se habilita es que un miembro con roles pueda operar.

La consulta de permisos efectivos vive en `permisosEfectivos` dentro de `infrastructure/negocio` y la comparten `RolRepository` y `SucursalRepository`, para no duplicar SQL ni cruzar dominios.

## Verificación técnica de fase 4

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test ./...`: correctos.
- Prueba de integración PostgreSQL sobre base desechable `tienda_fase4_test`: migración ejecutada dos veces con resultado idéntico; catálogo de 18 permisos sembrado; rol de sistema creado una vez por negocio con los 18 permisos; propietario activo con rol de sistema asignado; rol legacy conservado con código derivado `ENCARGADO_DE_PISO`; `rol_id` backfilleado a `roles_membresia` y después eliminado de `membresias_negocio`; único por negocio case-insensitive confirmado, incluido que el mismo código sí puede repetirse en otro negocio.
- `rol_service_test.go`: 14 casos. Cubren permiso concedido y ausente, lectura sin permiso, miembro no propietario que sí puede gestionar por permiso, validación de código y nombre, código reservado del rol de sistema, código duplicado, inmutabilidad del rol de sistema al editar y al eliminar, rol con miembros no eliminable, actualización sin campos, negocio archivado que bloquea escritura, membresía ajena, protección del último propietario y deduplicación de roles.
- `permiso_middleware_test.go`: 4 casos con `200`, `403`, `401` y `400`.
- `npm test -- --watch=false`: correcto, 29 archivos y 90 pruebas.
- `npm run build`: correcto, con los avisos preexistentes de presupuesto CSS.
- La base compartida `tienda` no fue modificada; la migración se aplicará en su próximo arranque de la API.

## Imprevistos resueltos de fase 4

- `Optional[T]` del proyecto usa el campo `Set`, no `Present`; se ajustó el service.
- Reemplazar el check de propietario exigía que `SucursalService` conociera los permisos. En lugar de duplicar la consulta o cruzar repositories, se amplió su puerto `SucursalRepository` con `PermisosEfectivos` y ambos repositories comparten la misma función de infraestructura.
- El stub de pruebas de sucursales devuelve el catálogo completo cuando el miembro es propietario, que es exactamente lo que produce el rol de sistema; así las pruebas de fase 2 siguieron siendo válidas sin reescribirlas.

---
