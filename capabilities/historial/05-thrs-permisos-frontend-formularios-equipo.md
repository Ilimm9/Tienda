# Plan 5: Nombres, correos únicos, permisos en frontend y cierre de Equipo y Roles

## Estado y control

- Estado: `Verificada`; aprobada por el usuario el 2026-10-04 («plan aprobado, empieza») e implementada ese día. Pendiente de aceptación final. Fecha: 2026-10-04. Fase 0 añadida el mismo día a petición del usuario.
- Continúa los pendientes del [plan 2](02-thrs-invitaciones-empleados-sucursal.md) y usa el sistema visual del [plan 4](04-thrs-adaptacion-opcion-g.md) con Tailwind (capability 05).
- Dependencias: RBAC de capability 01, registro y OTP de capability 06, precios de capability 12.

## Objetivo

Que perfil y empleado guarden el nombre con la misma forma y ningún correo se registre dos veces; que cada persona vea y pueda usar solo lo que su rol permite, tanto en la API como en la interfaz; que los formularios de Equipo y Roles validen antes de enviar y marquen sus campos obligatorios; y que Roles, Empleados e Invitaciones compartan el mismo aspecto.

## Hallazgos

- El backend ya autoriza con 20 permisos (`RequierePermiso` y `autorizar` en servicios). El frontend no los conoce: `GET /contexto/opciones` solo devuelve `tipo_miembro`.
- Menú (`NAVIGATION_ITEMS`) y rutas son iguales para todos; solo existen `authGuard` y `contextoGuard`. El 403 se trata a mano en 9 componentes de Equipo y Roles, con textos distintos.
- `precios/propuestas` (listar y autorizar) solo exige sesión y negocio activo. `ListarOpcionesContexto` devuelve todas las sucursales activas a cualquier miembro.
- `invitacion-nueva` usa señales sin validación; `empleado-form` valida nombre y primer apellido, y el correo no valida formato. La convención actual es «(opcional)», sin asterisco.
- Roles y permisos conserva CSS propio (`roles-shared.css` y tres más): acciones apiladas, estado en minúscula, miembros identificados solo por correo.
- Una invitación puede delegar un rol con permisos que el emisor no tiene. Su correo dice «3 días» en lugar de «72 horas».
- `perfiles_usuario` guarda `nombres` y `apellidos`; `empleados` guarda cuatro campos. Al aceptar una invitación hay que partir o unir a ciegas.
- `empleados` no impide repetir un correo dentro del negocio: solo hay un índice normal `(negocio_id, correo)`.
- El registro con un correo ya usado responde igual que uno nuevo (desafío ficticio, sin correo enviado): la persona espera un código que nunca llega.

## Enfoque

El backend es la única fuente de verdad. El frontend recibe los permisos efectivos y los usa para ocultar, deshabilitar y redirigir; nunca para autorizar. Toda acción oculta sigue protegida en la API.

- Descartado: derivar permisos del nombre del rol, un endpoint por pantalla, o un paquete `shared/permisos`. El permiso pertenece al contexto del negocio y vive en `contexto/`.
- Descartado: quitar los nombres de `empleados`. El empleado existe antes de tener cuenta, puede no tenerla nunca, y el perfil es global mientras el empleado pertenece a un negocio.

## Contrato

`GET /api/v1/contexto/opciones` añade a cada negocio:

```json
{ "id": "…", "tipo_miembro": "miembro", "permisos": ["equipo.empleados.ver", "catalogo.ver"], "sucursales": [] }
```

- `permisos`: códigos efectivos de la membresía, ordenados. El propietario recibe el catálogo completo.
- `sucursales`: el propietario recibe todas las activas; el miembro, solo las de sus asignaciones vigentes.
- Sin rutas nuevas. Errores con el formato actual `codigo`, `mensaje`, `campos`.
- Permisos nuevos en el catálogo: `precios.ver` y `precios.autorizar`, sembrados de forma idempotente y otorgados al rol propietario.

## Fase 0 · Nombres del perfil y correos únicos

**Perfil de usuario** (dominio `cuenta`):

1. `perfiles_usuario` pasa a `nombres`, `primer_apellido` y `segundo_apellido` (opcional). `empleados` no cambia.
2. Migración aditiva e idempotente: crear las columnas, copiar `apellidos` completo a `primer_apellido`, dejar `segundo_apellido` nulo y retirar `apellidos` en un paso posterior ya verificado. La persona corrige después desde su perfil.
3. `RegisterWithNames` y el contrato entre dominios reciben los tres campos; se retira `Register(fullName)`, que adivina dónde termina el nombre.
4. Registro general y registro desde invitación piden nombres\*, primer apellido\* y segundo apellido.
5. Al aceptar una invitación: `primer_apellido` y `segundo_apellido` se copian tal cual al empleado; `nombres` se reparte en `nombre` (primera palabra) y `segundo_nombre` (resto).

**Correo único al registrar un empleado** (dominio `negocio`):

6. Crear o editar un empleado, y corregir el correo de una invitación, rechazan con `409 CORREO_EMPLEADO_DUPLICADO` y `campos.correo` un correo que ya tenga otro empleado del mismo negocio, sin importar mayúsculas.
7. Índice único parcial sobre `(negocio_id, lower(correo))` cuando hay correo. La migración reporta duplicados existentes y no crea el índice hasta resolverlos.
8. Un correo con cuenta global existente sigue permitido: esa persona acepta la invitación iniciando sesión (plan 2). Se rechaza si esa cuenta ya es miembro del negocio.

**Correo único al registrarse** (dominio `cuenta`):

9. Registro con un correo ya verificado responde `409 CORREO_YA_REGISTRADO`, con `campos.correo` y enlaces a iniciar sesión y recuperar contraseña. Una cuenta pendiente de verificación reenvía el código, como hoy.
10. Esto cambia la respuesta neutra de capability 06 y permite saber si un correo tiene cuenta. Se compensa con límite de frecuencia por IP en registro. Ver «Decisiones».
11. Los formularios muestran el error bajo el campo de correo, sin perder lo ya capturado.

## Fase 1 · Backend

1. `ContextoNegocio` incorpora `Permisos []string`; `ContextoService.ListarOpciones` los resuelve con el contrato de `RolService.PermisosEfectivos`.
2. `ListarOpcionesContexto` filtra sucursales por asignación vigente cuando `tipo_miembro` es `miembro`.
3. `ValidarSucursalActiva` recibe el usuario: sucursal no asignada responde `404`. Aplica a compras, precios y productos con existencias.
4. `precios.ver` protege `GET /precios/propuestas`; `precios.autorizar` protege `PATCH …/autorizar`.
5. Invitación y asignación de roles: rechazar con `403` un rol cuyos permisos no sean subconjunto de los del emisor.
6. Plantillas de correo de invitación: vigencia en horas.
7. Actualizar `contexto_handler_test.go` (hoy afirma que no se exponen permisos) y añadir pruebas de cada regla.

Migración: solo sembrado de dos permisos. Antes de activar el filtro de sucursales, listar miembros activos sin asignación para que el administrador los resuelva; no asignar todas por compatibilidad.

## Fase 2 · Base de permisos en frontend

Todo en `frontend/src/app/contexto/`:

| Pieza | Responsabilidad |
| --- | --- |
| `contexto.models.ts` | `permisos: string[]` en `ContextoNegocio`. |
| `permisos.ts` | Constantes con los códigos, espejo del catálogo del backend. Evita cadenas sueltas. |
| `ContextoService` | `permisos` (computado desde el negocio activo), `puede(codigo)`, `puedeAlguno(...codigos)`. |
| `permiso.guard.ts` | Lee `data.permiso` (uno o varios); sin permiso, navega a `/sin-acceso` sin cambiar la URL. Espera a `asegurarInicializado()`. |
| `sin-acceso.ts` | Página «Sin acceso» dentro del shell: qué sección se intentó abrir, a quién pedir el permiso y botón «Ir a Inicio». No revela datos. |
| `puede.directive.ts` | `*appPuede="'roles.gestionar'"`: quita el elemento del DOM si falta el permiso. |
| `NavigationItem.permiso` | El menú lateral filtra elementos; un grupo sin hijos visibles desaparece. |
| `auth.interceptor.ts` | 403 con `ACCESO_DENEGADO`: recarga el contexto y muestra un mensaje único. Los componentes dejan de tratar el 403. |

Comportamiento cuando un permiso se retira en caliente: la siguiente petición responde 403, el contexto se recarga, el menú se actualiza y la vista pasa a «Sin acceso».

## Fase 3 · Aplicar permisos

| Sección | Ver (ruta y menú) | Acciones condicionadas |
| --- | --- | --- |
| Negocios | `negocios.ver` | Editar: `negocios.editar`. Archivar: `negocios.archivar` |
| Sucursales | `sucursales.ver` | Nueva: `sucursales.crear`. Editar: `sucursales.editar`. Archivar: `sucursales.archivar` |
| Catálogo y Proveedores | `catalogo.ver` | Crear, editar, importar y desactivar: `catalogo.gestionar` |
| Compras | `compras.ver` | Registrar: `compras.registrar` |
| Precios | `precios.ver` | Autorizar: `precios.autorizar` |
| Empleados | `equipo.empleados.ver` | Nuevo y editar: `equipo.empleados.gestionar`. Sucursales: `equipo.asignaciones.ver` / `.editar` |
| Invitaciones | `equipo.invitaciones.ver` | Nueva, reenviar y cancelar: `equipo.invitaciones.enviar`. Corregir correo: además `equipo.empleados.gestionar` |
| Roles y permisos | `roles.ver` | Nuevo, editar y eliminar: `roles.gestionar`. Asignar: `roles.asignar` |

- Rutas de formulario (`/nuevo`, `/:id/editar`, `/nueva`) llevan el permiso de gestión, no el de lectura.
- Con lectura sin gestión, la lista se muestra sin botones de acción y el detalle en solo lectura.
- Inicio muestra solo las tarjetas de sección permitidas.
- Ventas no tiene permisos en el catálogo: queda visible y fuera de este plan.

## Fase 4 · Formularios y validación

Reglas comunes:

- Formularios reactivos con `Validators`; `invitacion-nueva` pasa de señales sueltas a `FormGroup`.
- Asterisco en la etiqueta de cada campo obligatorio, con `aria-hidden`, y `aria-required="true"` en el control. Leyenda «Los campos con * son obligatorios» al inicio. Los opcionales van sin asterisco y sin «(opcional)».
- El error aparece bajo el campo al salir de él o al intentar guardar, con icono, `aria-invalid` y `aria-describedby` (kit del plan 4).
- Al enviar con errores: marcar todo como tocado y llevar el foco al primer campo inválido.
- El botón de guardar se deshabilita solo mientras envía, no por formulario inválido.
- Los errores `campos` del backend se asignan al control correspondiente. Espacios recortados antes de validar y enviar. Los límites se alinean con los del backend; si alguno falta allí, se añade.

| Formulario | Obligatorios (*) | Otras validaciones |
| --- | --- | --- |
| Registro y registro desde invitación | Nombres, primer apellido, correo, contraseña, confirmación | Correo no registrado (fase 0); reglas de contraseña actuales |
| Empleado | Nombre, primer apellido | Correo con formato y máximo 254; teléfono de 10 dígitos; correo no repetido en el negocio (fase 0); fecha de contratación no futura |
| Nueva invitación | Empleado, sucursal, rol | El empleado debe tener correo; si falta, enlace a editarlo |
| Corregir correo | Correo nuevo, confirmación | Formato; distinto del actual; confirmación idéntica |
| Rol | Código, nombre, mínimo un permiso | Código en mayúsculas, dígitos y guion bajo; longitudes actuales |
| Asignar roles | Mínimo un rol | Sin rol propietario para miembros |
| Sucursales del empleado | Sucursal | No repetir una asignación vigente |

## Fase 5 · Mejora visual

- **Roles y permisos**: migrar a Tailwind y retirar `roles-shared.css`, `roles.component.css`, `rol-form.component.css` y `miembro-roles.component.css`.
  - Lista de roles: acciones en una fila con botones de icono (`aria-label` y `title`); pastillas de estado con punto y texto capitalizado; el rol de sistema muestra un candado y «No editable» en vez de una celda vacía.
  - Miembros: avatar con iniciales, nombre y correo debajo; pastillas de rol; «Asignar» como botón chico.
  - Dos tarjetas con el patrón de tabla del plan 4: barra superior, tabla y pie con total.
  - Formulario de rol: permisos agrupados por módulo en tarjetas, con «Seleccionar todo» por módulo y contador «n de m».
- **Empleados**: columna de acciones con botones de icono; estado vacío con CTA solo si hay permiso; detalle con secciones y acciones ordenadas por importancia.
- **Invitaciones**: acciones por fila según estado y permiso; aviso cuando el empleado no tiene correo.
- **Botones** en las tres secciones: una acción principal violeta por pantalla; secundarias con borde; destructivas en rojo y con confirmación; Cancelar como fantasma.
- **Estados** de carga, vacío, error y sin permiso con el mismo componente visual. Móvil a 390 px: tablas con desplazamiento horizontal contenido y acciones accesibles. Sin colores ni tokens nuevos.

## Fase 6 · Cierre y verificación

- Pruebas Go: migración de perfil, correo duplicado de empleado y de registro, contexto con permisos, filtro de sucursales, precios con y sin permiso, rol no delegable, sucursal no asignada (`404`). Pruebas pendientes del plan 2: carrera entre aceptación y reenvío, rollback, sucursal archivada, rol desactivado, SMTP caído.
- Pruebas Angular: `puede`, guard, directiva, filtrado del menú, interceptor de 403 y validaciones de cada formulario.
- Recorrido manual con tres cuentas: propietario, miembro con rol de solo lectura y cajero con una sucursal.
- `go vet`, `go test ./...`, `ng test`, `ng build`; registrar por separado los fallos previos. Actualizar este plan, el plan 2 y `PLAN_IMPLEMENTACION.md`.

## Archivos previstos

- Backend `cuenta`: perfil, registro, migración y pruebas; adaptador entre dominios en `cmd/api`. Frontend `features/auth/register` y `features/invitacion`.
- Backend `negocio` en sus cuatro capas: `permiso.go`, contexto, servicios de rol, invitación y asignación, repositorio de contexto, handlers y pruebas. `backend/cmd/api/main.go`: grupos de precios. Plantillas de `correo`.
- Verificación de sucursal en `compra`, `precio` y productos: solo la composición HTTP; sin mover ni rehacer sus modelos.
- Frontend: `contexto/`, `layout/navigation.config.ts`, `layout/sidebar`, `features/auth/auth.interceptor.ts`, todos los `*.routes.ts`, `features/equipo/`, `features/roles-permisos/`, `features/home/`.
- En Negocios, Sucursales, Catálogo, Proveedores, Compras y Precios: solo `*appPuede` en botones y `data.permiso` en rutas.

## Fuera de alcance

- Permisos de Ventas, alta inicial (plan 1) y tema oscuro.
- Cambiar los campos de nombre de `empleados`; edición del nombre de un empleado ya vinculado a una cuenta.
- Rediseño de Catálogo, Compras, Proveedores, Precios, Negocios y Sucursales.
- Varias sucursales por invitación; historial detallado de eventos de invitación. Tablas nuevas.

## Riesgos

- Apellidos existentes quedan completos en `primer_apellido` hasta que cada persona los corrija. Duplicados de correo ya guardados impiden crear el índice único; la migración los reporta primero.
- Un miembro activo sin asignación quedará sin sucursales al activar el filtro. Mitigación: reporte previo y aviso «Solicita una asignación al administrador».
- Constantes de permisos en frontend y backend pueden divergir. Mitigación: prueba que compara contra `GET …/roles/permisos`.
- `navigation.config.spec.ts` ya falla; se corrige en la fase 2. Presupuesto de 4 kB por componente: estilos a Tailwind o al kit global.

## Tareas

- [x] Aprobar el alcance y resolver las decisiones.
- [x] Fases 0 a 6 (2026-10-04). Detalle, desviaciones y pendientes en «Seguimiento», al final.

## Criterios de aceptación

- El perfil guarda nombres, primer apellido y segundo apellido; ningún perfil existente pierde datos.
- No se puede registrar ni editar un empleado con un correo que ya use otro empleado del negocio. Registrarse con un correo ya verificado muestra el error en el campo y ofrece iniciar sesión.
- Un miembro solo ve en el menú las secciones para las que tiene permiso de lectura; una URL directa sin permiso muestra la página «Sin acceso».
- Ningún botón de crear, editar, eliminar, invitar, asignar o autorizar aparece sin su permiso; la API responde `403` si se intenta de forma directa.
- Un empleado solo ve y opera sus sucursales asignadas; el propietario, todas.
- Validar precios exige `precios.ver` y autorizar exige `precios.autorizar`. Nadie delega por invitación ni por asignación un rol con permisos que no posee.
- Todo campo obligatorio de Equipo y Roles lleva asterisco y muestra su error antes de enviar; los errores del backend aparecen junto a su campo.
- Roles y permisos no tiene CSS propio y coincide en estructura con Empleados e Invitaciones.
- El correo de invitación indica 72 horas.

## Decisiones

- Tomadas por el usuario el 2026-10-04: filtro de sucursales por asignación incluido; permisos en frontend para todas las secciones, Ventas sin permisos porque aún no existe; obligatorios solo con asterisco; URL directa sin permiso muestra «Sin acceso».
- Fase 0: perfil con tres campos y `apellidos` completo a `primer_apellido`; `empleados` sin cambio; nombre de empleado con cuenta, sin considerar por ahora.
- Registro con correo ya usado: opción A, error explícito con límite de frecuencia por IP. Modifica la respuesta neutra de capability 06 por decisión del usuario.

## Seguimiento

### Implementado (2026-10-04)

- **Fase 0.** Tabla real `perfil_usuarios`: `nombres`, `primer_apellido`, `segundo_apellido`. `MigratePerfilApellidos` corre antes de AutoMigrate; `apellidos` queda como respaldo sin obligatoriedad. `MigrateEmpleadoCorreoUnico` crea `uq_empleados_negocio_correo` u omite y reporta duplicados. `POST /auth/register` responde `409 CORREO_YA_REGISTRADO` y tiene límite de 10 por minuto por IP.
- **Fase 1.** `permisos` y sucursales filtradas en `GET /contexto/opciones`; `ValidarSucursalActiva` exige asignación vigente; `precios.ver` y `precios.autorizar` (22 permisos); `ROL_NO_DELEGABLE` en invitación y asignación; correo de invitación en horas.
- **Fase 2.** En `contexto/`: `permisos.ts`, `puede` / `puedeAlguno` / `puedeEn`, `permisoGuard` y `permisoChildGuard`, `*appPuede`, `sin-acceso.ts`, `acceso-denegado.service.ts`; menú filtrado con `filtrarNavegacion`; el interceptor atiende el 403 `ACCESO_DENEGADO`.
- **Fase 3.** `data.permiso` en todas las rutas; botones de Negocios, Sucursales, Catálogo, Productos, Proveedores, Compras, Precios, Equipo, Roles e Inicio según permiso.
- **Fase 4.** Validación, asteriscos, foco al primer error y errores del servidor por campo en registro, registro desde invitación, empleado, nueva invitación, corregir correo, rol y asignar roles.
- **Fase 5.** Roles, formulario de rol y asignar roles en Tailwind; se retiraron sus cuatro archivos CSS. `GET …/miembros` devuelve `nombre`.

### Desviaciones y añadidos

- Equipo y Roles ya pedían `…/mis-permisos` por pantalla; ahora leen el contexto. El endpoint sigue existiendo, sin uso desde el frontend.
- El 403 no expulsa siempre: se recarga el contexto y solo se muestra «Sin acceso» si la ruta actual dejó de estar permitida. Los componentes conservan su mensaje de 403 para datos secundarios.
- `/negocios/:negocioId/...` se evalúa contra ese negocio. Un negocio archivado no está en el contexto: decide ser propietario y la API.
- Nueva invitación exige `equipo.invitaciones.enviar`, `equipo.empleados.ver` y `roles.ver`, porque la pantalla carga empleados y roles.
- Nueva invitación valida con señales, no con `FormGroup`; el comportamiento es el del plan.
- Teléfono del empleado: de 7 a 30 caracteres, como el backend; no se fijó en 10 dígitos.
- Añadido: un rol exige al menos un permiso también en el backend, y nadie agrega a un rol permisos que no posee (cierra la escalada por `roles.gestionar`).
- Añadido: el correo de un empleado no puede repetir el de un miembro del negocio, incluido el propietario.
- «Mínimo un rol» al asignar se valida solo en el frontend; la API aún acepta dejar a un miembro sin roles.
- `precio` ganó `BranchOfProposal` para conocer la sucursal de una propuesta; las compras de sucursales no asignadas se filtran en el handler.
- Reenviar una invitación conserva su rol sin volver a comprobar que el emisor pueda otorgarlo.

### Ajustes posteriores (2026-10-04, a petición del usuario)

- El formulario de negocio recarga el contexto al guardar: un negocio recién creado muestra de inmediato el menú completo del propietario. Archivar y restaurar también recargan.
- **Negocios deja de ser sección libre.** Registra negocios una cuenta sin negocios o quien ya es propietario de alguno; un empleado invitado recibe `403 SOLO_PROPIETARIOS`. La lista solo incluye negocios propios o donde el rol tiene `negocios.ver`. En el frontend: `puedeCrearNegocio`, `puedeVerNegocios`, `data.acceso` en rutas, menú, botones e Inicio.
- `GET /negocios/:negocioId` sigue respondiendo a cualquier miembro: las pantallas de sucursales lo usan para leer nombre y estado.

### Pruebas

- `go vet` limpio. `go test ./...` pasa salvo `internal/config` (`TestProductionAcceptsStrongOTPSecret`, fallo previo).
- `ng build` correcto. `ng test`: 153 pasan; 8 fallan en `productos.component.spec.ts` (previos). `navigation.config.spec.ts` quedó corregido.
- Recorrido HTTP en base temporal vacía con Mailpit y tres cuentas: 48 comprobaciones correctas (registro, correos únicos, contexto, sucursal asignada y no asignada, retiro de asignación, precios, roles no delegables).
- Revisión visual con Chromium sin interfaz, 11 capturas en `capturas-plan5/`, sin errores de consola: registro con correo repetido, Roles (1440 y 390 px), formularios con errores, menú de un miembro y «Sin acceso».

### Pendientes

- Perfiles anteriores sin apellido quedan con `primer_apellido` vacío; retirar la columna `apellidos` en una migración posterior.
- Pruebas de integración con base y de componentes Angular para las pantallas de Roles y Equipo; pendientes de carrera y rollback del plan 2.
- El menú no marca «Roles y permisos» como activo en sus subpantallas (comportamiento anterior).
- Contraste y navegación completa con teclado no medidos con herramienta.
- `graphify update .` no se ejecutó: el comando no está instalado en este equipo.
