# Capability 30: Rediseño de Equipo (alta en 3 pasos, puestos e invitaciones por sucursal)

- Estado: **Verificada** (pendiente de prueba manual y aceptación final).
- Aprobación explícita: 2026-10-22 («implementa el plan»). El usuario acepta la paginación del lado del cliente y ajusta el alcance:
  - **Solo vista y un poco de backend.** El backend se limita a la tabla de puestos, que se llama **`puestos_negocio`**, y a sus endpoints. «Para no entrar en detalles posterior».
  - En el asistente se puede volver al paso 1 para editar y al paso 2 desde el 3. Sin el paso 2 no se puede invitar.
  - No se retira ningún campo del formulario actual (segundo nombre, segundo apellido, fecha de contratación).
- Fecha: 2026-10-22.
- ID 30 elegido por el usuario para no chocar con trabajo de Ilian (27 reservado para Compras; 28 y 29 libres).
- Diseño: canvas «Stockion - Disenio», artboard interactivo **I1 · Equipo: alta en 3 pasos, invitaciones y puestos** (tercera fila). Colores provisionales: los actuales.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Dominio `negocio` en todas las capas; frontend `features/equipo`. Base de conocimiento afectada: [gestion-colaboradores.md](../gestion-colaboradores.md).

## Objetivo

Dar de alta a un empleado e invitarlo sin salir de un solo flujo. Usar un teléfono mexicano de 10 dígitos, un número de empleado generado y puestos de un catálogo. Mantener los filtros actuales y paginar las listas largas.

## Alcance y reglas

### 1. Teléfono de 10 dígitos (México)
- Sin lada ni prefijo `+52`. Se guardan solo dígitos; exactamente 10. Vacío sigue permitido.
- Frontend: se aceptan solo dígitos, máximo 10, y se muestran como `55 1234 5678`. Error: «Escribe 10 dígitos, sin lada.»
- Backend (fuente de verdad): `Crear` y `Actualizar` de empleado y el registro por invitación (`POST /invitaciones/:token/registro`) quitan espacios y guiones y validan `^\d{10}$`; si no, responden 400.
- Datos existentes: una migración idempotente normaliza `empleados.telefono`:
  - quita todo lo que no sea dígito y un `52` inicial cuando quedan 12 dígitos;
  - si el resultado no tiene 10 dígitos, lo deja como estaba y lo reporta en el log, sin borrar el valor.
  - Al editar ese empleado se le pedirá corregirlo.

### 2. Alta en 3 pasos
Ruta `/equipo/empleados/nuevo` como asistente con indicador de pasos y vista previa lateral:
1. **Datos**: nombre, apellidos, correo (opcional), teléfono, puesto (select del catálogo) y número de empleado. Guarda el empleado (`POST empleados`, sin cambios de contrato salvo el teléfono).
2. **Sucursal**: una o varias sucursales y la principal. Por defecto, la sucursal del contexto. Usa el endpoint actual de asignaciones. «Omitir» deja al empleado sin sucursal.
3. **Invitación**: el rol se sugiere a partir del puesto y se puede cambiar. «Enviar invitación» usa el `POST invitaciones` actual con la sucursal principal. Se bloquea si no hay correo o sucursal, con un aviso que explica por qué. «Terminar sin invitar» cierra el flujo.
- Pantalla final con resumen y enlaces a «Ver empleado» y «Dar de alta otro».
- Cada paso persiste al avanzar. Si se abandona a la mitad, el empleado queda creado y se completa desde su detalle (no hay rollback).
- La edición de empleado sigue usando el formulario actual con las reglas 1 y 3.

### 3. Generar número de empleado
- Botón «Generar» junto al campo. Llama a `GET /negocios/:negocioId/administracion/empleados/siguiente-numero?sucursal_id=` y responde `{ "numero": "SC-003" }`.
- Prefijo: las iniciales del nombre de la sucursal, ignorando «Sucursal» y conectores (de, del, la, las, los, y). Ejemplos: «Sucursal Centro» → `SC`, «Sucursal Norte» → `SN`. Si queda una sola palabra, se toman sus 2 primeras letras más una `S` inicial. Todo en mayúsculas y sin acentos.
- Consecutivo: el máximo `PREFIJO-NNN` existente en el negocio + 1, con 3 dígitos (`SC-010`; de `SC-999` pasa a `SC-1000`).
- Es una sugerencia: el campo sigue editable. La unicidad la garantiza `ExisteNumero` al guardar (409 si choca; el frontend sugiere regenerar).
- Si el usuario aún no eligió sucursal (paso 1), se usa la sucursal del contexto.

### 4. Catálogo de puestos
- Nueva tabla `puestos`: `id`, `negocio_id`, `nombre` varchar(120), `rol_sugerido_id` (nulo permitido), `activo`, auditoría. Nombre único por negocio sin distinguir mayúsculas ni acentos.
- Es del negocio y aplica a todas las sucursales. **No muestra cuántos empleados tiene cada puesto**, por decisión del usuario.
- `empleados.puesto_id` (nulo permitido) reemplaza al texto libre `puesto`. La respuesta sigue exponiendo `puesto` (nombre) para no romper la lista.
- Migración: crea un puesto por cada valor distinto de `empleados.puesto` (recortado y agrupando mayúsculas y minúsculas), enlaza `puesto_id` y conserva la columna anterior hasta la aceptación final.
- Endpoints, con el permiso de administración de empleados actual:
  - `GET /negocios/:negocioId/administracion/puestos`
  - `POST /negocios/:negocioId/administracion/puestos` `{nombre, rol_sugerido_id?}` (409 si es duplicado)
  - `PATCH .../puestos/:puestoId` `{nombre?, rol_sugerido_id?, activo?}`
- No hay borrado físico, solo desactivar. Un puesto inactivo no aparece en el select, pero se conserva en los empleados que ya lo tienen.
- Pantalla «Puestos» en Equipo: agregar (con aviso de duplicado), editar nombre y rol sugerido, y desactivar.

### 5. Empleados: filtros, búsqueda y paginación
- Se conservan el buscador (nombre, número o correo) y las pestañas de estado: Todos, Pendientes, Activos, Suspendidos y Terminados, con su contador.
- Columnas: Empleado (con número), Puesto, Contacto (correo y teléfono), Sucursal, Estado, Acceso y Acciones.
- Paginación de 10 en 10 cuando hay más de 10 resultados, con «1–10 de N» y anterior/siguiente. Cambiar el filtro o la búsqueda regresa a la página 1.

### 6. Invitaciones: por sucursal del contexto
- **Se quita el selector de sucursal.** La lista siempre envía `sucursal_id` = sucursal del contexto (el backend ya acepta este filtro) y se muestra «Solo {sucursal} · cámbiala en el menú superior».
- Cambiar la sucursal del contexto recarga la lista.
- Se conservan las pestañas Sin aceptar (n), Aceptadas, Canceladas y Todas, el botón «Nueva invitación», la columna Emitida, las acciones, el pie y las 3 tarjetas informativas.
- Paginación igual a la de la regla 5.

### Decisión pendiente para aprobar
- **Paginación del lado del cliente** (propuesta): las APIs actuales ya devuelven la lista filtrada y un negocio pequeño tiene decenas o pocos cientos de registros. Es el mismo patrón que la capability 24.
- Si se prefiere paginación en el servidor, se agregan `page` y `page_size` y la respuesta `{items, total}`, como en `plataforma/handler.go`. Eso cambia el contrato y obliga a ajustar los consumidores.

### 7. Rol «Cajero» predeterminado (anexo aprobado 2026-10-22)
- Solicitud: sembrar un rol Cajero básico y editable.
- Los roles son por negocio, así que el usuario aprobó crearlo **al crear cada negocio**, en la misma transacción que el rol Propietario (`sembrarRolPropietarioDeNegocio`).
- No hay backfill en el arranque: si el negocio lo borra, no reaparece.
- Datos: código `CAJERO`, nombre «Cajero», descripción «Rol básico de caja; ajústalo a tu operación.».
  - `es_rol_sistema = false`: se puede editar, agregar permisos, desactivar o borrar.
  - Permisos: `catalogo.ver`, `compras.ver`, `compras.registrar`.
- Implementación:
  - `sembrarRolCajeroDeNegocio` en `rol_migration.go`, llamada desde `NegocioRepository.Crear`;
  - constantes `CodigoRolCajero` y `PermisosRolCajero` en `domain/negocio/rol.go`.
- Verificado con `TestCrearNegocioSiembraRolCajero` (requiere `CAJERO_TEST_DATABASE_URL`) sobre la base temporal `tienda_cajero_test`, migrada por la API y borrada al terminar. El negocio nuevo tiene el rol Cajero, activo, que no es de sistema y tiene exactamente esos 3 permisos.

## Fuera de alcance
- Contar empleados por puesto.
- Ver invitaciones de otras sucursales desde una sola pantalla.
- Teléfonos extranjeros o con lada.
- Cambios de permisos o de roles.

## Archivos estimados
- Backend:
  - `domain/negocio/empleado.go` y `domain/negocio/puesto.go` (nuevo).
  - `application/negocio/empleado_service.go` (teléfono y siguiente número), `invitacion_service.go` (teléfono en el registro) y `puesto_service.go` (nuevo).
  - `infrastructure/negocio/puesto_repository.go`, `puesto_migration.go`, `empleado_telefono_migration.go` y `empleado_repository.go`.
  - `interfaces/http/negocio/puesto_handler.go`, `empleado_handler.go` y `cmd/api/main.go` (rutas).
- Frontend (`features/equipo`):
  - `empleado-form` (teléfono, generar número, select de puesto).
  - Nuevo `empleado-alta` (asistente de 3 pasos).
  - `empleados.component` (paginación).
  - `invitaciones/invitaciones.ts` (sin selector, sucursal del contexto, paginación).
  - Nuevo `puestos/` (pantalla y servicio).
  - `equipo.routes.ts`, modelos y servicios.
  - Página de registro por invitación (teléfono de 10 dígitos).

## Pruebas
- Go unitarias:
  - validación de teléfono (10 dígitos, vacío, con espacios, 9 y 11 dígitos, `+52`);
  - prefijo de sucursal (casos «Sucursal Centro», una palabra, acentos, conectores);
  - consecutivo (sin previos, con huecos, 999 → 1000);
  - puesto duplicado sin distinguir mayúsculas ni acentos.
- Go integración (PostgreSQL temporal):
  - migración de teléfonos (normaliza y conserva los inválidos);
  - migración de puestos (agrupa valores);
  - `siguiente-numero` y el 409 por número repetido.
- Angular (vitest):
  - máscara de teléfono;
  - asistente (avanza, omite sucursal, bloquea la invitación sin correo);
  - paginación y reinicio a la página 1 al filtrar;
  - invitaciones enviando `sucursal_id` del contexto y sin selector;
  - Puestos (agregar, duplicado, desactivar).
- Manual con Chromium: el flujo completo de alta e invitación, que el correo llegue a Mailpit, y anchos de 360 y 1440 px.

## Criterios de aceptación
1. No se puede guardar un teléfono que no tenga 10 dígitos, ni desde la interfaz ni desde la API.
2. Desde «Nuevo empleado» se llega a «Invitación enviada» sin salir del asistente.
3. «Generar» propone `SC-NNN` único para Sucursal Centro.
4. El puesto se elige de un catálogo de todo el negocio, sin conteos; los valores anteriores quedan migrados.
5. Empleados conserva el buscador y las pestañas de estado, y pagina a partir de 11 resultados.
6. Invitaciones no tiene selector de sucursal, muestra solo la del contexto, conserva sus pestañas y pagina.

## Decisiones de implementación (2026-10-22)
Alcance reducido por el usuario a «solo vista y un poco de backend». Por eso:
- **Backend hecho:** tabla `puestos_negocio` (AutoMigrate más el índice único `uq_puestos_negocio_nombre` sobre `(negocio_id, lower(nombre))`), `PuestoService` (reutiliza la autorización de empleados) y `GET/POST/PATCH /administracion/puestos`. La API también valida el teléfono del empleado (10 dígitos) al crearlo y editarlo.
- **No se hizo (respecto de lo escrito arriba):**
  - No existe el endpoint `siguiente-numero`: el número se calcula en el frontend con la lista de empleados (`numero-empleado.ts`) y la unicidad la garantiza `ExisteNumero` al guardar.
  - Prefijo: `S` + iniciales (hasta 3) de las palabras de la sucursal, sin «Sucursal» ni conectores. Ejemplos: «Sucursal Centro» → `SC`, «Matriz de Querétaro» → `SMQ`, y sin palabras → `SUC`.
  - No hay migración de teléfonos existentes: al editar se muestran como 10 dígitos (se quita la lada 52 si quedan 12) y la API rechaza los inválidos.
  - El registro por invitación (teléfono del perfil) no cambia.
  - Empleados no gana la columna Sucursal: el resumen no la trae y requeriría backend.
- **Empleados:** ya tenía buscador, pestañas de estado y paginación (`p-table`, 10 por página a partir de 11). Sin cambios.
- **Invitaciones:** se quitó el selector y siempre se envía `sucursal_id` del contexto; se recarga al cambiar la sucursal. Paginación de 10 en el cliente.
- **Asistente** (`empleado-form`, solo al crear):
  - Se conservan todos los campos: segundo nombre, segundo apellido y fecha de contratación.
  - Se puede volver desde el paso 2 o 3 al 1 (guarda con PATCH, no duplica) y del 3 al 2.
  - El paso 3 explica por qué no puede invitar (falta sucursal, falta correo o falta permiso) y ofrece volver al paso correspondiente o «Terminar sin invitar».
  - En el paso 2 las sucursales ya asignadas no se pueden quitar; eso se hace en «Sucursales asignadas».
- Puestos: nueva pantalla en el menú de Equipo con alta, renombrado en línea, rol sugerido y Activar/Desactivar.

### Ajuste 2026-10-22: puesto vinculado por id
A petición del usuario («realiza la vinculación por id»; base recreada y sin producción):
- `empleados.puesto_id` (uuid, nullable) con FK `fk_empleados_puesto` → `puestos_negocio(id)` `ON DELETE RESTRICT`. La columna de texto `empleados.puesto` se elimina (`DROP COLUMN IF EXISTS`), sin migrar valores.
- API: `puesto_id` en crear/editar. Debe ser un puesto activo del negocio (422 `campos.puesto_id`); conservar el puesto actual se permite aunque esté inactivo. Las respuestas traen `puesto_id` y `puesto` (nombre por JOIN, solo lectura).
- Frontend: select por id; un puesto inactivo asignado aparece como «(inactivo)»; el rol sugerido se busca por id.
- Verificado: base temporal con columna `puesto_id` + FK; segundo arranque idempotente; `go test` (nueva `TestEmpleadoServicePuestoDebeSerActivoDelNegocio`) y specs de equipo en verde; `ng build` correcto.

## Pruebas ejecutadas (2026-10-22)
- `go build ./...` correcto.
- `go test ./...`: correcto salvo `TestProductionAcceptsStrongOTPSecret`, que ya fallaba antes. Pruebas nuevas en `puesto_service_test.go`: permisos, normalización, duplicado, nombre corto, rol ajeno, desactivar y teléfono de 10 dígitos.
- Base temporal `tienda_e2e30` (borrada al terminar):
  - la API crea `puestos_negocio` y su índice;
  - el índice rechaza «Cajero»/«CAJERO» en el mismo negocio;
  - `GET /puestos` sin sesión responde 401.
  - La base local `tienda` no cambió.
- Angular:
  - `ng build` correcto (solo los avisos de presupuesto que ya existían).
  - `ng test`: 242 de 248. Los 6 fallos de `alta-inicial.spec.ts` (`matchMedia`) ya existían.
  - Pruebas nuevas: `numero-empleado.spec.ts`, `empleado-form.component.spec.ts` (flujo completo, omitir sucursal y volver, sin correo, teléfono) e `invitaciones.spec.ts` (sucursal del contexto, recarga y paginación).
- **Pendiente:** prueba manual en el navegador con sesión real (pantallas de 360 y 1440 px) y aceptación del usuario.

## Tareas
- [x] Aprobación del usuario (paginación en el cliente; alcance reducido).
- [x] Backend: `puestos_negocio`, endpoints y validación de teléfono, con pruebas.
- [x] Frontend: asistente, Puestos, invitaciones por contexto y paginación, con pruebas.
- [x] Actualizar `gestion-colaboradores.md`.
- [ ] Verificación manual y aceptación final.
