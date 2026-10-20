# Arreglos de invitaciones y permisos

- Estado: **Cerrada**. Aceptación final explícita del usuario: 2026-10-20.
- Inicio de implementación: 2026-10-19, tras registrar estado `Aprobada`.
- Aprobación explícita: 2026-10-19, el usuario solicitó «Lo puedes implementar por favor» y autorizó pruebas con Mailpit.
- Fecha: 2026-10-19. Verificación final: 2026-10-20.
- Decisiones de los seis temas y tres reglas adicionales confirmadas por el usuario el 2026-10-19; implementación autorizada posteriormente en la misma fecha.
- Alcance de esta entrega: implementar los seis arreglos, agregar pruebas y verificar contra PostgreSQL aislado y Mailpit local. No ampliar automáticamente permisos de roles existentes ni publicar/desplegar cambios.
- Capability 20: `capabilities/historial/20-arreglos-invitaciones.md`; se evita ocupar el ID 19 previsto por el borrador de precios por sucursal. Organización conforme a [AGENTS.md](../../AGENTS.md).

## Objetivo

Que el invitado reciba el rol correcto, vea las secciones que puede usar y opere únicamente dentro del negocio y la sucursal activa autorizada. Aunque tenga varias sucursales asignadas, debe cambiar de contexto para consultar las existencias de otra. Corregir los seis puntos de la revisión, conservando el flujo de registro, verificación y aceptación existente.

## Funcionamiento que se conserva

- La invitación selecciona empleado, sucursal y un rol del negocio; no selecciona permisos individuales.
- Al aceptar se vinculan cuenta, membresía, empleado, rol y sucursal en una transacción. El correo debe coincidir; el enlace vence, no puede reutilizarse y el rol propietario no se delega por invitación.
- Los permisos efectivos son la unión de los roles activos de la membresía. Editar un rol afecta a quienes lo tengan; no se congela una copia al invitar.
- Interfaz, rutas y API verifican permisos. La interfaz se actualiza al aceptar; la autorización del servidor sigue siendo obligatoria en cada operación.
- Los roles pertenecen al negocio. Asignar otra sucursal no crea automáticamente un rol diferente para ella.

## 1. Reenvío y delegación del rol — prioridad alta

**Problema:** crear verifica que el emisor pueda otorgar el rol; reenviar solo verifica permiso de envío y, al corregir correo, gestión de empleados. El repositorio valida que el rol esté activo, pero no que el nuevo emisor pueda delegarlo.

**Propuesta:** comprobar otra vez los permisos actuales del emisor y que pueda otorgar el rol, tanto al reenviar como al corregir destinatario. Reutilizar la validación de delegación existente y ejecutarla antes de modificar correo, cancelar el enlace anterior o insertar el nuevo. El rechazo debe dejar datos y enlace anterior intactos y no enviar correo.

- **Decisión confirmada: opción A.** Validar dentro de la transacción de reemisión, con el rol y la invitación que se están modificando. No restringir el reenvío únicamente al emisor original: otro usuario autorizado puede hacerlo si también puede delegar el rol.
- **Regla adicional confirmada:** al aceptar también se revalidan el acceso del emisor, su permiso de envío y su capacidad actual para otorgar el rol. Si perdió acceso o el rol ahora excede sus permisos, rechazar sin crear ni modificar membresía, roles o asignaciones del invitado. La comprobación forma parte de la transacción de aceptación; no basta la validación realizada al emitir.
- Este rechazo no cancela automáticamente el enlace: permanece sujeto a su estado y vencimiento. Alguien autorizado debe revisar o reemitir la invitación antes de conceder acceso.

Fuente: [servicio](../../backend/internal/application/negocio/invitacion_service.go), [reemisión y aceptación](../../backend/internal/infrastructure/negocio/invitacion_repository.go), [delegación de roles](../../backend/internal/infrastructure/negocio/rol_repository.go).

## 2. Suspensión o terminación del empleado — prioridad alta

**Problema:** cambia el estado del empleado y pierde acceso operativo a sucursales, pero su membresía puede seguir activa; el cálculo de permisos del negocio no comprueba el estado del empleado. Algunas operaciones sin validación de sucursal siguen disponibles.

- **Decisión confirmada: opción A.** Bloquear todo acceso a ese negocio para el colaborador suspendido o terminado, desde la validación central de acceso y permisos. No puede observar datos ni operar; el negocio y sus asignaciones dejan de aparecer en su contexto, selectores y listados personales. Su cuenta y accesos a otros negocios siguen intactos.
- Conservar historial para administradores autorizados; ocultar el acceso al empleado no implica borrar registros laborales, roles ni asignaciones históricas.
- Comprobar el estado en servidor en cada petición; rechazar también URLs y llamadas API directas. Al refrescar el contexto del navegador, retirar negocio y asignaciones no accesibles y limpiar datos de la vista afectada. El servidor bloquea nuevas consultas desde la baja; el refresco de la interfaz sigue el mecanismo del tema 5, sin prometer borrado remoto instantáneo de información ya renderizada.
- Tratar explícitamente propietarios sin ficha de empleado, cuentas de plataforma y membresías legacy; no bloquearlas por un `JOIN` indiscriminado. No permitir dejar al negocio sin propietario efectivo.
- **Reactivación confirmada:** una suspensión temporal recupera los roles y asignaciones anteriores que sigan vigentes; una baja requiere seleccionar y confirmar explícitamente los roles y sucursales antes de habilitar nuevamente el acceso. Cambiar solo el estado laboral de terminado a activo no recupera accesos anteriores de forma automática.
- La reactivación no debe reactivar una membresía revocada o suspendida por otra causa, ni recuperar roles o asignaciones retirados de forma independiente.

Fuente: [estado del empleado](../../backend/internal/infrastructure/negocio/empleado_repository.go), [permisos efectivos](../../backend/internal/infrastructure/negocio/rol_repository.go), [contexto y sucursales](../../backend/internal/infrastructure/negocio/contexto_repository.go).

## 3. Inventario visible por sucursal — prioridad alta

**Problema:** el selector limita sucursales asignadas, pero el listado de Productos devuelve inventario y totales de otras sucursales del mismo negocio, incluidos los de variantes.

- **Decisión confirmada: combinar autorización por asignación con vista exclusiva de la sucursal activa.** No se suman ni se devuelven existencias de todas las sucursales asignadas.
- Ejemplo: tiene A, B y C asignadas y está en A; ve exclusivamente inventario de A. Para consultar B debe cambiar el contexto a B, y el servidor vuelve a comprobar su asignación vigente.
- Precio/costo por sucursal se mantienen dentro del alcance de su plan específico; este arreglo exige que stock, totales y estados de existencias correspondan solo a la sucursal activa. El catálogo de productos sigue perteneciendo al negocio: no duplicar fichas para filtrar inventario.
- Pasar y validar negocio/sucursal activa en las consultas correspondientes; el servidor no confía en IDs del navegador y nunca devuelve inventario de una sucursal no autorizada, ni de otra asignada que no esté seleccionada. Sin contexto válido o sucursal accesible, no usar existencias globales como reemplazo.
- Al cambiar de contexto, limpiar o recargar los datos de la sucursal anterior y evitar que respuestas tardías vuelvan a mostrarlos. Cubrir productos simples, variantes, agregados y contratos consumidos por edición y ajustes.
- Las vistas operativas basadas en contexto muestran una sucursal a la vez; el propietario conserva su alcance autorizado para seleccionar sucursales, sin mezclar existencias de otras en la vista activa.

Fuente: [handler de Productos](../../backend/internal/interfaces/http/product_handler.go), [consultas de inventario](../../backend/internal/infrastructure/product_repository.go), [contratos de productos](../../backend/internal/domain/product_view.go).

## 4. Dependencias entre permisos — prioridad media

**Problema:** puede guardarse un rol con `precios.autorizar` sin `precios.ver`. El permiso se asigna, pero la sección no aparece porque su entrada exige lectura. También hay acciones que necesitan lectura de otras secciones: crear invitación consulta empleados y roles.

- **Decisión confirmada: opción A.** Seleccionar automáticamente los permisos de lectura necesarios al marcar una acción y mostrar qué permisos se añadieron. Al quitar lectura con acciones dependientes seleccionadas, exigir confirmación para retirar esas acciones o conservar lectura.
- Definir dependencias con las rutas y llamadas reales, no solo por el sufijo del permiso: por ejemplo, emitir invitación desde la interfaz también requiere ver invitaciones, empleados y roles.
- Aplicar la normalización en frontend y backend para nuevas altas y ediciones de permisos: la respuesta de la API refleja los permisos finales. No permitir que completar dependencias otorgue permisos que el emisor no puede conceder; en ese caso, rechazar sin guardar parcialmente.
- **Roles existentes: decisión confirmada.** Listar primero combinaciones incompletas, permisos que se agregarían y personas afectadas. Obtener autorización específica antes de ampliar esos permisos; no completar lecturas masivamente de forma automática ni interpretar la aprobación futura de código como autorización de esa ampliación.

Fuente: [editor de roles](../../frontend/src/app/features/roles-permisos/rol-form.component.ts), [servicio de roles](../../backend/internal/application/negocio/rol_service.go), [rutas de Equipo](../../frontend/src/app/features/equipo/equipo.routes.ts), [menú](../../frontend/src/app/layout/navigation.config.ts).

## 5. Refresco de permisos en sesiones abiertas — prioridad media

**Problema:** el contexto se conserva en memoria. Un permiso recién agregado puede no aparecer hasta recargar página; al retirarlo, la API rechaza la acción y el manejo de `403` actualiza el contexto.

- **Decisión confirmada: opción A.** Reutilizar `ContextoService.recargar()` después de editar/asignar roles en la sesión actual y cuando el usuario vuelva a una pestaña visible. Deduplicar solicitudes, conservar contexto válido y manejar errores. No agregar consultas periódicas, WebSocket ni nueva infraestructura.
- Mantener recuperación ante `403 ACCESO_DENEGADO`. Si cambian permisos durante la edición de un formulario, conservar borrador cuando sea posible, explicar el cambio y dejar que el servidor rechace operaciones no autorizadas.
- Este mecanismo actualiza la sesión de otro usuario cuando vuelve a su pestaña; no promete sincronización instantánea mientras permanece abierta en primer plano. Si ya no tiene acceso al negocio, no conservarlo como selección ni mostrar sus sucursales como asignaciones actuales.

Fuente: [contexto](../../frontend/src/app/contexto/contexto.service.ts), [recuperación ante acceso denegado](../../frontend/src/app/contexto/acceso-denegado.service.ts), [interceptor](../../frontend/src/app/features/auth/auth.interceptor.ts).

## 6. Roles ofrecidos y resumen de permisos — prioridad media

**Problema:** el selector excluye roles inactivos y propietario, pero ofrece otros roles que el emisor no puede otorgar. La emisión falla al guardar y el resumen muestra solo el nombre del rol.

- **Decisión confirmada: opción A.** Ofrecer únicamente roles delegables y mostrar un resumen legible de sus permisos antes de emitir. No mostrar roles que el emisor no puede conceder. Reutilizar la consulta de roles y la regla de delegación del backend; definir el contrato exacto en la capability, sin duplicar reglas de seguridad en JavaScript.
- Crear una invitación conserva sus permisos actuales de sección; no se amplía quién puede consultar roles o empleados. La API vuelve a validar al guardar porque el rol puede cambiar después de cargar la pantalla.
- Agrupar el resumen por sección y distinguir lectura, registro, edición y autorización. No presentar el resumen como una copia inmutable de permisos.

Fuente: [selector de invitaciones](../../frontend/src/app/features/equipo/invitaciones/invitacion-nueva.ts), [plantilla](../../frontend/src/app/features/equipo/invitaciones/invitacion-nueva.html), [listado de roles](../../backend/internal/infrastructure/negocio/rol_repository.go).

## Tablas y límites de organización

| Tablas | Responsabilidad en el plan |
| --- | --- |
| `invitaciones_negocio` | Rol, destinatario, emisor, vigencia y estado del enlace. |
| `empleados`, `membresias_negocio` | Estado laboral y acceso al negocio. |
| `roles`, `permisos`, `permisos_rol`, `roles_membresia` | Delegación, permisos efectivos y dependencias acordadas. |
| `asignaciones_empleado_sucursal`, `sucursales` | Alcance autorizado y vigencia de sucursales. |
| `inventario_sucursal`, `producto_negocio`, `producto_variantes` | Existencias visibles, variantes y totales. |

No se justifican tablas nuevas. Usar servicios, contratos y repositorios existentes; conservar dirección de dependencias y organización por dominio. La integración de Productos con Negocio usa el contrato de aplicación de contexto, nunca un repositorio importado de otro dominio.

## Secuencia propuesta y pruebas

1. **Formalización:** trasladar las decisiones confirmadas a la capability, concretar contratos y archivos de Negocio/Productos/interfaz y obtener aprobación explícita de implementación.
2. **Seguridad:** corregir reenvío, política de suspensión y aislamiento de existencias; agregar pruebas por cada caso antes de declarar cierre.
3. **Interfaz y permisos:** dependencias, refresco de contexto y selector con resumen; comprobar regresión de rutas, menús y acciones.
4. **Verificación integral:** aceptar una invitación con cuenta nueva y existente; confirmar permisos esperados en navegador y API con PostgreSQL de prueba aislado.
5. **Documentación:** registrar resultados y pendientes; actualizar [base de colaboradores](../gestion-colaboradores.md), [base de productos](../registro-productos.md) e [índice general](../../PLAN_IMPLEMENTACION.md). Ejecutar `graphify update .` si se modifica código y la herramienta está disponible; documentar bloqueos reales.

Casos mínimos de aceptación:

- [x] Reenviar o corregir correo con un rol no delegable devuelve rechazo sin cambiar datos ni enviar correo.
- [x] Reenvío permitido invalida enlace anterior y conserva rol/sucursal correctos; enlace vencido, cancelado o usado no se acepta.
- [x] Aceptaciones concurrentes y aceptación contra reenvío producen un único resultado válido, sin vínculos parciales; rollback conserva coherencia.
- [x] Si el emisor pierde acceso o el rol supera sus permisos antes de aceptar, la aceptación se rechaza sin acceso parcial, aunque el enlace todavía no haya vencido.
- [x] Suspendido o terminado no puede observar ni operar datos del negocio; negocio y asignaciones desaparecen de su contexto. APIs directas lo bloquean, sin afectar otros negocios ni dejar un propietario desprotegido.
- [x] Reactivar una suspensión recupera solo accesos todavía vigentes; después de una baja se exige reasignación explícita. Ningún caso recupera accesos revocados por otra causa.
- [x] Invitado con A/B/C asignadas y contexto A recibe solo stock, totales y variantes de A. Solo al cambiar a B puede ver B; una sucursal no asignada se rechaza, y una respuesta tardía no restaura la vista anterior.
- [x] Las lecturas dependientes se agregan automáticamente al elegir una acción y coinciden en UI/API, sin superar permisos delegables; los roles existentes se listan y conservan sus permisos hasta autorizar específicamente la corrección.
- [x] Los permisos nuevos aparecen al refrescar contexto; los retirados se bloquean en API aunque el navegador conserve datos anteriores.
- [x] Selector y resumen coinciden con roles que la API permite otorgar; cambios entre selección y envío se rechazan correctamente.
- [x] Invitado nuevo o existente ve las secciones de sus roles activos; los permisos se suman y no se sustituyen por accidente.
- [x] Pruebas focalizadas Go/Angular, integración PostgreSQL, build y revisión manual de invitación/OTP/correo registran resultados y bloqueos por separado.

## Decisiones confirmadas

Confirmadas por el usuario: opción A en temas 1, 2, 4, 5 y 6; tema 3 exige sucursal asignada **y** activa, consultando una sola por contexto. Estas decisiones forman parte del alcance cuya implementación fue autorizada explícitamente.

También confirmadas:

- Revalidar permisos actuales del emisor al aceptar y rechazar si ya no puede otorgar el rol.
- Al reactivar, suspensión recupera accesos vigentes y baja exige reasignación explícita.
- Listar roles existentes incompletos y obtener autorización específica antes de ampliar sus permisos. La lectura automática de nuevas altas y ediciones no autoriza una ampliación masiva retroactiva.

No quedan preguntas funcionales abiertas de esta ronda. La aprobación explícita de implementación quedó registrada; la ampliación de permisos de roles existentes conserva su autorización separada.

## Línea base anterior a la implementación

- Revisión anterior: 108 pruebas Go y 49 pruebas Angular focalizadas correctas; tres comprobaciones temporales reprodujeron omisión de delegación en reenvío y combinación de permisos que oculta Precios.
- Los hallazgos de suspensión e inventario se sustentan en lectura del código; no se confirmó su comportamiento contra una base desplegada.
- No hay prueba integral de aceptación contra PostgreSQL real, recorrido completo en navegador ni entrega real de correo en esta revisión. Las suites existentes no sustituyen esos casos.
- Resultados anteriores son la línea base; las verificaciones de implementación se registran abajo.

## Contratos y archivos autorizados

- **Negocio:** modelos y reglas de `backend/internal/domain/negocio/`, servicios de `application/negocio/`, repositorios/migraciones/pruebas de `infrastructure/negocio/`, handlers de `interfaces/http/negocio/` y rutas de `backend/cmd/api/main.go`. Se autoriza cubrir todas las consultas de membresía del dominio para impedir accesos de empleados dados de baja, conservando propietarios y cuentas legacy sin empleado.
- **Productos:** `backend/internal/domain/product_view.go`, `application/product_service.go`, `infrastructure/product_repository.go`, `interfaces/http/product_handler.go` y sus callers/pruebas. `GET /negocios/:id/catalogo/productos` requerirá `sucursal_id` válida y accesible; devuelve el mismo contrato con inventario, agregados y variantes solo de esa sucursal. El frontend enviará esa selección en cada carga.
- **Roles:** catálogo de permisos expone dependencias de lectura; crear/editar permisos normaliza dependencias antes de validar delegación y persistir. Listado de roles puede filtrar delegables y devolver códigos para su resumen; reporte de dependencias faltantes es de solo lectura e incluye afectados. No corrige roles existentes sin autorización específica.
- **Baja/reactivación:** baja finaliza asignaciones y bloquea acceso conservando roles y registros históricos; `POST /negocios/:id/administracion/empleados/:empleadoId/reactivar` recibe `roles` y `sucursales` no vacíos, exige gestionar empleados, asignar roles y editar asignaciones y reemplaza los accesos expresamente elegidos en transacción. El PATCH de estado no puede reactivar acceso de una baja por sí solo. Si nunca tuvo cuenta, puede volver a `pendiente` para una nueva invitación que elija rol/sucursal explícitos, sin conceder acceso hasta aceptarla. Suspensión conserva los accesos todavía vigentes y bloquea su uso. No se reactiva una membresía revocada por otra causa ni se da de baja al último propietario efectivo.
- **Interfaz:** `frontend/src/app/contexto/`, `features/equipo/`, `features/invitacion/`, `features/roles-permisos/`, `features/productos/`, navegación/rutas y pruebas conectadas. Los cambios visuales reutilizan componentes, estilos y formularios existentes.
- **Verificación:** pruebas Go de dominio/aplicación/HTTP e integración PostgreSQL con sufijo `_test`; pruebas Angular y recorrido real de invitación, OTP y permisos con Mailpit. Scripts de QA temporal en `/tmp`; pruebas permanentes junto a sus dominios. No crear carpetas genéricas ni dependencias nuevas.

## Tareas de implementación

- [x] Formalizar alcance y registrar aprobación explícita.
- [x] Delegación transaccional en reenvío y aceptación.
- [x] Bloqueo de baja/suspensión y reactivación segura.
- [x] Inventario exclusivo de sucursal activa y manejo de cambio de contexto.
- [x] Dependencias de permisos, diagnóstico de roles existentes y roles delegables.
- [x] Refresco de contexto y resumen de permisos en invitación.
- [x] Pruebas unitarias, integración, build y Mailpit/navegador.
- [x] Actualizar bases, índice, grafo y resultados finales.

## Resultados de implementación

Entrega local verificada el 2026-10-20. Sin tablas nuevas, dependencias adicionales, migraciones de permisos antiguos, commits ni despliegue. Se conserva el borrador independiente de precios por sucursal.

Cambios: bloqueo central por estado laboral; reenvío/aceptación revalidan delegación dentro de una transacción serializada por negocio; reactivación explícita de bajas; consultas de Productos exigen sucursal accesible y suman únicamente sus existencias, incluyendo todas las variantes; interfaz descarta respuestas tardías, actualiza contexto y muestra roles delegables con permisos legibles.

Contratos concretos:

- `GET /negocios/:id/administracion/roles?delegables=true`: roles activos otorgables; incluye `permisos`, `permisos_detalle` y `permisos_faltantes`. El reporte de Roles cruza miembros para mostrar personas afectadas, sin modificar permisos.
- `GET /negocios/:id/administracion/permisos`: cada permiso incluye `requiere` con códigos de sus dependencias.
- `PATCH /negocios/:id/administracion/roles/:rolId`: completar lecturas faltantes de un rol anterior exige `autorizar_dependencias_existentes: true`, además de los permisos de gestión/delegación. Editar solo el nombre conserva exactamente sus permisos.
- `GET /negocios/:id/catalogo/productos?sucursal_id=:uuid`: sin sucursal válida devuelve 400; sin acceso devuelve 403/404 según validación. Inventario, suma, estado y variantes corresponden exclusivamente a esa sucursal.
- `POST /negocios/:id/administracion/empleados/:empleadoId/reactivar`: `roles` y `sucursales` explícitos; no restaura membresías revocadas independientemente.

Verificaciones ejecutadas:

| Verificación | Resultado y evidencia |
| --- | --- |
| Go: aplicación, infraestructura y HTTP de Negocio | Correctas; `/tmp/tienda-invitaciones-go-final.log`. |
| PostgreSQL real, base aislada `tienda_invitaciones_test` | `TestInvitacionesPermisosSucursal`, tres ejecuciones correctas. Cubre doble aceptación, aceptación contra reenvío, rol ampliado, emisor revocado, rollback, consentimiento de roles antiguos, baja/suspensión, otros negocios, membresía revocada, último propietario y stock/estados de dos variantes en dos sucursales. `/tmp/tienda-invitaciones-integracion.log`. |
| Angular excluyendo únicamente la suite histórica de alta inicial | 44 archivos, 207 pruebas correctas. Incluye dependencias, contexto y respuestas tardías de sucursal. `/tmp/tienda-invitaciones-angular-final.log`. |
| Compilación Go y `go vet ./...` | Correctos; binario temporal y `/tmp/tienda-invitaciones-vet.log`. |
| Angular, build de producción | Correcto; persisten avisos previos de presupuesto de CSS/bundle. `/tmp/tienda-invitaciones-build-final.log`. |
| API y Mailpit reales | Registro/OTP, correo de invitación y reenvío, enlace anterior inválido, registro del invitado, aceptación única, permisos exactos, sección no permitida, sucursal no asignada, stock A/B, suspensión/baja y reactivación explícita correctos. Solo direcciones ficticias `@tienda.local`, SMTP local 1025. `/tmp/tienda-mailpit-qa.log`. |
| Chromium sobre el build final | Menú del invitado, Productos, selector sin propietario, permisos legibles, lectura automática, vista móvil 390 px, refresco de permisos y suspensión correctos; cero excepciones JavaScript. `/tmp/tienda-invitaciones-browser.log`; captura `/tmp/tienda-invitacion-permisos-movil.png`. |
| Grafo y formato | `graphify update .` correcto, extracción AST local sin API; `git diff --check` correcto. |

Límites y pendientes explícitos:

- Suite Go global conserva el fallo anterior `TestProductionAcceptsStrongOTPSecret` de configuración OTP. Las demás pruebas ejecutadas pasan. Registro: `/tmp/tienda-invitaciones-go-completo.log`.
- Suite Angular global: 215 correctas y 6 fallos previos en `alta-inicial.spec.ts`, más 2 errores de esa suite por `window.matchMedia`. También conserva expectativas antiguas de icono y paso final. Registro: `/tmp/tienda-invitaciones-angular-completo.log`. No se alteró ese flujo ajeno al plan.
- Mailpit prueba entrega SMTP local; no certifica proveedores externos ni recepción Gmail/Outlook. El entorno se reinició durante QA y se reconstruyeron los temporales para repetir la verificación final.
- No se ensayó carga productiva ni se modificó la instancia Docker habitual. Base `tienda_invitaciones_test`, API de QA y archivos `/tmp/tienda-*` eliminados el 2026-10-20 tras la aceptación final; las rutas de evidencia citadas ya no existen.
- Código verificado no equivale a aceptación final ni autoriza ampliar roles existentes, publicar o desplegar.
