# Capability 33: Puesto Cajero por defecto y «Crear nuevo» en selectores de puesto y rol

- Estado: **Verificada** (pendiente de prueba manual y aceptación final).
- Aprobación explícita: 2026-10-23 («Por default puedes tambien crear el puesto cajero con el rol sugerido de cajero… cuando se crea el negocio» y «en la parte donde escoger un puesto o un rol te indique crear uno nuevo y te redirija a su seccion… podrias hacer esas adecuaciones»).
- Fecha: 2026-10-23.
- Organización conforme a [AGENTS.md](../AGENTS.md). Dominio `negocio` (backend) y `features/equipo` (frontend). Base de conocimiento afectada: [gestion-colaboradores.md](gestion-colaboradores.md). Amplía la [30](30-rediseno-equipo.md) sin cambiar su contrato.

## Alcance y reglas

### 1. Puesto Cajero al crear el negocio
- En la misma transacción que crea el negocio y siembra el rol Cajero (`sembrarRolCajeroDeNegocio`), se crea el puesto «Cajero» en `puestos_negocio`, activo, con `rol_sugerido_id` = ese rol.
- Solo para negocios nuevos: no se siembra en negocios existentes ni al arrancar la API. Si después se borra o renombra, no se vuelve a crear.
- Sin cambios de endpoints.

### 2. «Crear nuevo…» en los selectores
Cada selector agrega al final la opción «+ Crear nuevo puesto…» o «+ Crear nuevo rol…», visible solo con el permiso para crear; elegirla lleva a la sección:

| Selector | Opción | Destino | Permiso |
| --- | --- | --- | --- |
| Alta/edición de empleado, «Puesto» | Crear nuevo puesto | `/equipo/puestos` | `equipo.empleados.gestionar` |
| Alta de empleado paso 3, «Rol y permisos» | Crear nuevo rol | `/roles-permisos/nuevo` | `roles.gestionar` |
| Nueva invitación, «Rol» | Crear nuevo rol | `/roles-permisos/nuevo` | `roles.gestionar` |
| Puestos, «Rol sugerido al invitar» (alta de puesto) | Crear nuevo rol | `/roles-permisos/nuevo` | `roles.gestionar` |

- Si el formulario de empleado tiene cambios sin guardar, se pide confirmación antes de salir; al cancelar, el selector vuelve a su valor anterior.
- Fuera de alcance: los selectores de rol por fila en Puestos y el selector múltiple de reactivación de empleado.

## Pruebas y criterios de aceptación

- Integración (`CAJERO_TEST_DATABASE_URL`): crear negocio deja el puesto Cajero activo con el rol Cajero sugerido.
- Frontend: la opción aparece solo con permiso y navega a su destino; con cambios sin guardar pide confirmación.
- `go build`, `go vet`, `go test ./...`, `ng test` de los componentes tocados y `ng build`.

## Decisiones

- 2026-10-23: se reutiliza la transacción de alta del negocio; no se agrega migración de datos para negocios existentes porque el usuario pidió el puesto «cuando se crea el negocio».

## Tareas

- [x] Backend: sembrar puesto Cajero.
- [x] Frontend: opción «Crear nuevo» en los cuatro selectores.
- [x] Pruebas y verificación.

## Verificación

- 2026-10-23: `go build`, `go vet` correctos. `go test ./...`: todo pasa salvo `TestProductionAcceptsStrongOTPSecret` (`internal/config`), falla previa ajena a este cambio.
- 2026-10-23: `TestCrearNegocioSiembraRolCajero` en base temporal `tienda_cajero_test` (migrada con `database.Init`, borrada al terminar): pasa; verifica rol y puesto Cajero activo con el rol sugerido.
- 2026-10-23: `ng test` de `features/equipo` y `features/home`: 42 de 42 (incluye: opción visible, navegación directa a Puestos, confirmación con datos capturados y regreso al valor anterior, rol «Crear nuevo» a Nuevo rol). `ng build` correcto.
- Pendiente: prueba manual en navegador y aceptación final. Negocios ya existentes no reciben el puesto (por alcance).
