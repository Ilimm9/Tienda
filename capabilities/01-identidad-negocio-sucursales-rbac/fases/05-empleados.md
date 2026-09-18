# Fase 5: Empleados

## Estado de fase

`Verificada`

Idea preliminar registrada el 2026-09-12. Implementación autorizada explícitamente por el usuario el 2026-09-12, para ejecución no supervisada junto con las fases 4 y 6 a 8.

Decisión del usuario del 2026-09-12: el modelo actual de empleados debe **coincidir con `base.MD`**, no coexistir con una forma paralela.

## Fuente de verdad y compatibilidad

`base.MD` define `empleados` con alcance de negocio directo y nombre propio de la persona, sin depender de que exista una cuenta.

### Estado real verificado del código (2026-09-12)

- `domain/negocio/empleado.go`: `Empleado{id, perfil_id uuid not null, numero varchar(50) unique global, puesto varchar(120), estado varchar(30) default 'activo', creado_en, actualizado_en}`.
- `domain/negocio/membresia_negocio.go`: incluye `empleado_id *uuid`, es decir el vínculo apunta en dirección contraria a `base.MD`.
- `Empleado` aparece únicamente en la lista de AutoMigrate de `internal/database/database.go`. Ningún service, repository, handler, seed ni prueba lo consume.

Consecuencia: la tabla `empleados` es andamiaje vacío sin consumidores. Esta fase **reemplaza su forma** para igualar `base.MD`. La migración sigue siendo idempotente y defensiva por si alguna base ya tiene filas: si existen, se conservan y se completan con valores derivados antes de imponer restricciones.

### Divergencias a cerrar

| Campo hoy | `base.MD` | Acción |
| --------- | --------- | ------ |
| `perfil_id uuid not null` | no existe | eliminar; la identidad de cuenta llega por `membresia_id` |
| sin `negocio_id` | `negocio_id uuid not null` | agregar con FK; es el eje de aislamiento |
| sin `membresia_id` | `membresia_id uuid unique` | agregar con FK; se llena al aceptar invitación en fase 6 |
| `numero varchar(50)` único global | `numero_empleado varchar(40)` | renombrar, reducir a 40, retirar único global |
| — | único `(negocio_id, numero_empleado)` | crear |
| sin nombre propio | `nombre`, `segundo_nombre`, `primer_apellido`, `segundo_apellido` | agregar; `nombre` y `primer_apellido` obligatorios |
| sin contacto | `correo varchar(254)`, `telefono varchar(30)` | agregar; índice `(negocio_id, correo)` |
| sin fechas laborales | `contratado_en date`, `terminado_en date` | agregar |
| sin autoría | `creado_por_usuario_id uuid not null` | agregar con FK a `usuarios` |
| `estado` default `'activo'` | default `'pendiente'` | cambiar default |

El retiro de `membresias_negocio.empleado_id` pertenece a esta fase, porque aquí se invierte el vínculo hacia `empleados.membresia_id`, tal como manda `base.MD`.

### Excepción de tipo ya vigente

`base.MD` declara `estado_empleado` como enum de PostgreSQL. El proyecto viene implementando esos enums como `varchar` con default y validación en service desde fases 1 y 2. Esta fase mantiene esa convención; los valores conservan los nombres de `base.MD`.

## Modelo objetivo

### `empleados`

```text
id                     uuid primary key
negocio_id             uuid not null references negocios(id)
membresia_id           uuid null unique references membresias_negocio(id)
numero_empleado        varchar(40) null
nombre                 varchar(100) not null
segundo_nombre         varchar(100) null
primer_apellido        varchar(100) not null
segundo_apellido       varchar(100) null
correo                 varchar(254) null
telefono               varchar(30) null
puesto                 varchar(120) null
estado                 varchar(30) not null default 'pendiente'
contratado_en          date null
terminado_en           date null
creado_por_usuario_id  uuid not null references usuarios(id)
creado_en              timestamptz not null
actualizado_en         timestamptz not null
```

Índices y constraints:

- Único `(negocio_id, numero_empleado)` cuando `numero_empleado` no es nulo.
- Índice `(negocio_id, correo)`.
- Índice `(negocio_id, estado)` para listados.
- `membresia_id` único global; una membresía no puede representar a dos empleados.

Estados previstos: `pendiente` sin cuenta vinculada, `activo` con membresía vigente, `suspendido`, `terminado` con `terminado_en`.

## Migración y backfill

Aditiva, transaccional e idempotente, en el estilo de `sucursal_migration.go`:

1. Agregar todas las columnas faltantes con `IF NOT EXISTS`.
2. Renombrar `numero` a `numero_empleado` solo si la columna origen existe y la destino no.
3. Retirar el índice único global de `numero`.
4. Para filas preexistentes, derivar `negocio_id`, `nombre` y `primer_apellido` desde el perfil referido por `perfil_id`; si no puede resolverse, dejar la fila marcada y no imponer `not null` hasta que exista dato válido.
5. Eliminar `perfil_id` una vez migrado.
6. Backfillear `empleados.membresia_id` desde `membresias_negocio.empleado_id`, respetando la unicidad.
7. Eliminar la columna `membresias_negocio.empleado_id` y su campo en el struct.
8. Crear FK e índices solo si no existen.

No borrar filas, no truncar nombres y no reasignar números de empleado en ejecuciones posteriores.

## Objetivo y alcance (idea)

CRUD de personas laborales por negocio, registrables sin cuenta de usuario, con vínculo opcional posterior a una membresía cuando acepten invitación en fase 6.

Incluye (idea):

- Alta de empleado con nombre completo, número interno opcional único por negocio, puesto, correo y teléfono opcionales.
- Cambio de estado y baja lógica con `terminado_en`; sin eliminación física.
- Detalle de empleado como punto de entrada a invitar, en fase 6, y asignar sucursales, en fase 7.
- Pantallas reales en `equipo/empleados`, hoy placeholder en `features/equipo/equipo.routes.ts`: listado, alta, detalle, edición y baja.
- Aislamiento estricto por `negocio_id`, con el mismo patrón de fase 2.
- Autorización por permiso del módulo `equipo`, resuelto en fase 4.

No incluye (idea):

- Crear cuenta o membresía automáticamente; eso solo ocurre al aceptar invitación en fase 6.
- Asignación de sucursales, que corresponde a fase 7.
- Asignación de roles, que se hace sobre la membresía en fases 4 y 6.
- Nómina, horarios o datos laborales fuera de los campos de `base.MD`.

## Organización backend prevista

```text
backend/internal/domain/negocio/empleado.go            (reformado)
backend/internal/application/negocio/empleado_service.go
backend/internal/infrastructure/negocio/empleado_repository.go
backend/internal/infrastructure/negocio/empleado_migration.go
backend/internal/interfaces/http/negocio/empleado_handler.go
```

## Implementación de fase 5

Fecha: 2026-09-12.

### Backend

```text
backend/internal/domain/negocio/empleado.go                        (reformado a base.MD)
backend/internal/domain/negocio/membresia_negocio.go               (sin empleado_id)
backend/internal/application/negocio/empleado_service.go           (nuevo)
backend/internal/application/negocio/empleado_service_test.go      (nuevo)
backend/internal/infrastructure/negocio/empleado_migration.go      (nuevo)
backend/internal/infrastructure/negocio/empleado_repository.go     (nuevo)
backend/internal/infrastructure/negocio/empleado_migration_integration_test.go (nuevo)
backend/internal/interfaces/http/negocio/empleado_handler.go       (nuevo)
backend/internal/database/database.go                              (MigratePhaseFive)
backend/cmd/api/main.go                                            (composición y rutas)
```

### API

```text
GET    /api/v1/negocios/:negocioId/administracion/empleados?estado=&buscar=
POST   /api/v1/negocios/:negocioId/administracion/empleados
GET    /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId
PATCH  /api/v1/negocios/:negocioId/administracion/empleados/:empleadoId
```

Autorización por permiso de fase 4: `equipo.empleados.ver` para lectura y `equipo.empleados.gestionar` para escritura, más negocio activo.

No existe `DELETE`: la baja es lógica mediante `estado = terminado`, que además registra `terminado_en` cuando no se envía explícitamente.

### Frontend

```text
frontend/src/app/features/equipo/empleado.models.ts
frontend/src/app/features/equipo/empleado.service.ts
frontend/src/app/features/equipo/empleado.service.spec.ts
frontend/src/app/features/equipo/empleados.component.{ts,html,css}
frontend/src/app/features/equipo/empleado-form.component.{ts,html,css}
frontend/src/app/features/equipo/empleado-detalle.component.{ts,html,css}
frontend/src/app/features/equipo/equipo.routes.ts
```

`equipo/empleados` dejó de ser placeholder. El detalle enlaza a invitar, que se habilita en fase 6, y a sucursales asignadas, que se habilita en fase 7.

## Verificación técnica de fase 5

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test ./...`: correctos.
- Prueba de integración PostgreSQL sobre base desechable `tienda_fase5_test`: migración ejecutada dos veces con resultado idéntico; la fila legacy conservó su identificador y derivó `nombre` y `primer_apellido` desde `perfiles_usuario`; `numero` renombrado a `numero_empleado`; `negocio_id` derivado de la membresía; vínculo invertido a `empleados.membresia_id`; `membresias_negocio.empleado_id` y `empleados.perfil_id` eliminados; número único por negocio pero repetible entre negocios; y varios empleados sin número conviviendo gracias al único parcial.
- `empleado_service_test.go`: 12 casos. Cubren permiso de lectura, estado de filtro inválido, normalización de número y correo, alta sin número ni cuenta, validación de nombre, apellido, correo y fecha, número duplicado, negocio archivado, actualización sin campos, estado inválido, limpieza explícita con `null`, empleado de otro negocio y armado del nombre completo.
- `npm test -- --watch=false`: correcto, 30 archivos y 95 pruebas.
- `npm run build`: correcto.

## Imprevistos resueltos de fase 5

- `base.MD` deja `numero_empleado` opcional, de modo que el único por negocio debe ser parcial: de otro modo dos empleados sin número colisionarían. La migración crea el índice con `WHERE numero_empleado IS NOT NULL` y hay una prueba que lo verifica.
- Las restricciones `NOT NULL` solo se aplican si ninguna fila quedó sin resolver, para que una base con datos legacy incompletos no rompa el arranque de la API.

---
