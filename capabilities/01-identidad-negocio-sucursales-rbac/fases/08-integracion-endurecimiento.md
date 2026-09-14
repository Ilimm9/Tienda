# Fase 8: Integración y endurecimiento

## Estado de fase

`Verificada`

Idea preliminar registrada el 2026-09-12. Implementación autorizada explícitamente por el usuario el 2026-09-12, para ejecución no supervisada.

## Objetivo y alcance (idea)

Cierre transversal de la capability: verificar que negocios, sucursales, contexto activo, RBAC, empleados, invitaciones y asignaciones funcionan como un solo sistema coherente, sin fugas entre negocios/sucursales y sin cabos sueltos de fases previas.

Incluye (idea):

- Reemplazar cualquier check residual `tipo_miembro = propietario` que no haya migrado a permisos (fase 4) durante fases 2-3.
- Verificar que las convergencias con `base.MD` cerradas en fases 4 y 5 quedaron completas: `roles` con `codigo`/`es_rol_sistema`/`activo`, `empleados` con `negocio_id`/`membresia_id`, y sin columnas `rol_id` ni `empleado_id` en `membresias_negocio`.
- Retirar excepciones temporales documentadas si su condición de retiro ya se cumplió: puente legacy de sucursal en dominio negocio (fase 2), campos sombra `Nombre`/`EmailLegacy` en `negocios` (fase 1) y `direccion` texto legacy en `sucursales` (fase 2).
- Contrastar el esquema final contra `capabilities/_template/base.MD` tabla por tabla para el alcance de esta capability, y documentar toda divergencia que subsista con su motivo; la convención `varchar` en lugar de enums de PostgreSQL es la divergencia conocida y aceptada.
- Navegación integrada: enlaces cruzados entre empleado, invitación, membresía, roles y sucursales asignadas; breadcrumbs consistentes.
- Auditoría de aislamiento: pruebas cruzadas de fuga de datos entre negocios distintos y entre sucursales distintas, para cada módulo de la capability.
- Uniformar contratos de error (`400/401/403/404/409/422/500`) en todos los endpoints de la capability.
- Regresión de producto (catálogo, ventas) para confirmar que no se rompió nada al introducir contexto/RBAC.
- Cierre documental: actualizar `CAPABILITY.md` y `SEGUIMIENTO.md` con estado final de todas las fases.

No incluye (idea):

- Funcionalidad nueva; solo integración, limpieza y endurecimiento de lo construido en fases 1-7.

## Implementación de fase 8

Fecha: 2026-09-12.

### Checks residuales de `tipo_miembro` sustituidos

Quedaban tres en `NegocioService`, heredados de fase 1. Ahora usan permisos:

| Operación | Permiso |
| --------- | ------- |
| Editar negocio | `negocios.editar` |
| Archivar negocio | `negocios.archivar` |
| Restaurar negocio | `negocios.archivar` |

`NegocioRepository` ganó `PermisosEfectivos`, igual que los repositories de sucursal, empleado, invitación y asignación. Los cinco comparten la misma función `permisosEfectivos` de `infrastructure/negocio`.

El único uso restante de `"propietario"` fuera de pruebas es la creación de la membresía al registrar un negocio, que es correcto: quien crea el negocio es su propietario.

### Defectos de integración encontrados y corregidos

Estos no aparecían en las pruebas unitarias de cada fase; sólo surgieron al ejercitar el sistema completo:

1. **Negocio creado en runtime sin rol de sistema.** La migración de fase 4 siembra el rol `PROPIETARIO` al arrancar, pero un negocio creado después quedaba sin él. Como la autorización ya depende de permisos, su propietario se habría quedado sin acceso a su propio negocio hasta el siguiente reinicio. `NegocioRepository.Crear` ahora siembra el rol, sus permisos y la asignación a la membresía dentro de la misma transacción, con la función compartida `sembrarRolPropietarioDeNegocio`.
2. **Segundo arranque roto.** `MigratePhaseOne` ejecutaba `ALTER TABLE membresias_negocio ALTER COLUMN rol_id DROP NOT NULL` sin condición, y fase 4 elimina esa columna. El segundo `Init` fallaba con `column "rol_id" does not exist`. La sentencia quedó condicionada a que la columna exista.
3. **Índice único parcial ausente en el esquema real.** GORM no puede declarar `WHERE es_principal = true AND activo = true` desde los tags del struct, así que la invariante de una sola sucursal principal por empleado no tenía defensa en base. Se agregó `MigratePhaseSeven` con ese índice y con el índice `(negocio_id, sucursal_id)` de `base.MD`.
4. **`Pluck` sobre `uuid.UUID`.** Documentado en fase 7; afectaba también a `QuedaPropietarioConRolSistema` de fase 4.

### Auditoría de aislamiento

Se agregó `internal/database/integracion_capability_test.go`, que levanta el esquema real con `Init` sobre una base desechable y recorre la capability completa con dos cuentas y dos negocios:

- `Init` es idempotente: ejecutarlo dos veces no falla ni cambia el esquema.
- Un negocio recién creado entrega a su propietario los 18 permisos del catálogo.
- Una cuenta ajena no puede leer negocio, sucursales, empleados ni roles de otra.
- Una sucursal o un empleado de otro negocio no se encuentran, aunque el UUID exista.
- Una sucursal de otro negocio no puede asignarse a un empleado propio.
- Un miembro activo sin roles puede leer pero no mutar en ningún módulo.
- Al concederle un rol con `sucursales.crear`, la misma operación procede, y sigue sin poder tocar el módulo de equipo.
- El negocio no puede quedarse sin propietario efectivo.
- Las columnas que `base.MD` no declara (`membresias_negocio.rol_id`, `membresias_negocio.empleado_id`, `empleados.perfil_id`) no reaparecen.

### Contraste final contra `base.MD`

Tablas de la capability que ahora coinciden con `base.MD`: `negocios`, `direcciones`, `configuraciones_negocio`, `sucursales`, `membresias_negocio`, `empleados`, `permisos`, `roles`, `permisos_rol`, `roles_membresia`, `invitaciones_negocio`, `asignaciones_empleado_sucursal`.

Divergencias que subsisten, todas conocidas y documentadas:

| Divergencia | Motivo | Estado |
| ----------- | ------ | ------ |
| Enums de PostgreSQL implementados como `varchar` con validación en service | Convención adoptada desde fases 1 y 2; los valores conservan los nombres de `base.MD` | Aceptada |
| `sucursales.nombre` admite 180 en base y 160 en la aplicación | Fase 2 evitó estrechar una columna con datos | Aceptada |
| `sucursales.direccion` texto legacy | Producto aún usa su modelo legacy; administración solo escribe `direccion_id` | Aceptada |
| `negocios.nombre` y `negocios.email` como campos sombra | Compatibilidad con el modelo legacy de producto | Aceptada |
| El negocio sembrado por `SeedDevelopment` no tiene membresías | Es andamiaje de desarrollo para flujos legacy de producto; al no tener miembros, RBAC no aplica sobre él | Aceptada |

`capabilities/_template/base.MD` no fue modificado en ninguna fase.

## Verificación técnica de fase 8

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test -count=1 ./...`: correctos, incluidas las cinco pruebas de integración PostgreSQL de las fases 2, 4, 5, 7 y 8.
- `npm test -- --watch=false`: correcto, 32 archivos y 106 pruebas. Al inicio de esta sesión eran 24 archivos y 54 pruebas.
- `npm run build`: correcto, con los avisos preexistentes de presupuesto CSS.
- Las cinco bases temporales `tienda_fase*_test` fueron eliminadas al terminar. La base compartida `tienda` no fue tocada en ningún momento.

## Pendiente de cierre de fase 8

- Revisión visual y smoke manual del usuario sobre la aplicación en ejecución.
- Al primer arranque de la API contra la base compartida se aplicarán las migraciones de fases 4, 5 y 7, que incluyen eliminar `membresias_negocio.rol_id`, `membresias_negocio.empleado_id` y `empleados.perfil_id`. Las tres eran andamiaje sin consumidores y el backfill ocurre antes del borrado, pero conviene respaldar la base antes de ese arranque.

---
