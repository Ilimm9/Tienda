# Plan 2: Invitar empleados a una sucursal

## Estado y control

- Estado: `Aprobada`; aprobación explícita del usuario el 2026-10-01 para implementar el flujo de invitación.
- Fecha: 2026-10-01. Alcance actual: implementación parcial, ver «Alcance aprobado y seguimiento».
- Archivo directo en `capabilities/thrs`, sin carpetas adicionales, por solicitud del usuario.
- Este borrador no sustituye la capability aprobada requerida antes de implementar.
- Dependencias: empleados, RBAC e invitaciones de capability 01; correo de 07; [alta inicial](01-thrs-respaldo-2026-10-04-alta-inicial-empresa-sucursal.md).

## Objetivo

Administrador registra empleado, elige sucursal y rol, comparte enlace y empleado completa su cuenta. Al aceptar, obtiene acceso operativo únicamente a sus sucursales autorizadas; no se le pide crear empresa propia.

## Referencias y hallazgos

- Revisión transversal de los 36 Markdown existentes, con prioridad a [base.MD](base-modelo-datos.md), fases 4 a 8, sesiones/OTP de 06 y correo de 07.
- Tienda ya tiene empleados, membresías, invitaciones, roles y asignaciones; no construir un sistema paralelo.
- Invitación actual dura 7 días, guarda SHA-256 del token y entrega token una sola vez; correo es opcional y enlace copiable funciona si falla SMTP.
- Hoy invitación no captura sucursal; aceptación enlaza membresía/empleado/rol, pero no asigna sucursal.
- Contexto lista todas las sucursales activas del negocio; validar sucursal hoy no recibe identidad del usuario.
- Login, registro y OTP no conservan de forma completa el retorno enviado desde invitación; cubrir esta interrupción del flujo.
- Aceptación actual consulta vigencia antes de transacción; reforzar bloqueo y revalidación dentro de ella.
- Capability 12 permite validar precios a cualquier sesión del negocio: incompatible con delegación segura; ajuste incluido expresamente abajo.

## Referencia del proyecto de Juanjo

Ruta revisada: `/home/thrs/Projects/didacticapp/didacticapp-backnestjs/src/coadministradores`.

- `services/invitaciones-coadministrador.service.ts`: token aleatorio de 32 bytes, SHA-256, expiración, bloqueo pesimista, reenvío que revoca enlace anterior y auditoría.
- `constants/coadministrador.constants.ts`: vigencia de 72 horas.
- `coadministradores.service.ts` y configuración: construyen enlace con `token` y envían correo después de persistir.
- `coadministrador-invitaciones-public.controller.ts`: consulta pública y aceptación con límites de frecuencia.
- `aceptacion-coadministrador.service.ts`: muestra organización/correo enmascarado; recibe nombre, apellidos y contraseña; crea cuenta, enlaza organización y activa perfil en una transacción.
- Ese proyecto marca correo verificado al aceptar y rechaza correos con cuenta existente.
- Adoptar experiencia de enlace y personalización, transacciones y rotación; adaptar a sesiones, OTP y cuentas globales de Tienda. No copiar arquitectura NestJS ni tablas `dd_coadministrador_*`.

## Enfoque recomendado

Extender invitación existente con sucursal fija y personalización guiada. Usuario confirmó personalización desde enlace y verificación por correo: conservar OTP para cuenta nueva y permitir cuenta existente con mismo correo.

- Alternativa: copiar aceptación pública de Juanjo y usar enlace como verificación; cambia política de activación de Tienda y necesita decisión explícita.
- Alternativa: seguir enviando al registro genérico; reutiliza más UI, pero pierde continuidad y claridad sobre sucursal.
- Propuesta inicial: una sucursal por invitación; después administrador puede asignar más mediante flujo actual.
- Vigencia confirmada: **72 horas desde cada emisión o reenvío**, calculadas en servidor; al cumplirse el plazo, enlace deja de ser válido.

## Flujo del administrador

1. Abrir Equipo / Invitaciones o acción «Invitar» desde empleado o sucursal.
2. Seleccionar empleado existente sin cuenta vinculada; ofrecer enlace al alta de empleado si falta.
3. Exigir correo válido, sucursal activa del negocio y rol activo permitido.
4. Mostrar resumen: empleado, empresa, sucursal, rol y vigencia.
5. Guardar invitación y emitir enlace `/invitacion/:token` una sola vez.
6. Permitir copiarlo para compartir por el canal que el administrador elija; no enviar mensajes externos desde este trabajo.
7. Intentar correo con servicio existente; fallo conserva invitación y muestra opción de copiar.
8. Listar estados `pendiente`, `aceptada`, `expirada`, `cancelada`, con filtros por sucursal/empleado y «Sin aceptar».
9. Mostrar correo destinatario al administrador autorizado, fecha de emisión, vencimiento y fecha de aceptación cuando exista.
10. Invitación pendiente muestra «Aún no aceptada» y tiempo restante; vencida muestra «Expirada sin aceptar». No inferir aceptación por envío de correo o apertura del enlace.
11. Ofrecer «Reenviar invitación» y «Corregir correo y reenviar» para pendientes o expiradas; una cancelada puede reemitirse si empleado sigue elegible.
12. Reenvío cancela pendiente anterior y emite token nuevo de forma atómica, con otras 72 horas; no recuperar tokens originales desde DB.
13. Corregir correo exige formulario con correo actual/nuevo y confirmación; guardar cambio en empleado e invitación nueva dentro de la misma transacción.
14. Cancelar deja enlace inutilizable; para cambiar sucursal o rol, cancelar y emitir otra. Invitación aceptada no admite reenvío ni corrección por este flujo.

Correo anterior se conserva en fila histórica; no editar invitación emitida ni cuenta global ya creada. Si destinatario anterior alcanzó a registrar cuenta pero no aceptó, esa cuenta conserva sus datos y no obtiene acceso por el enlace invalidado. Nuevo correo sigue registro o login y OTP según corresponda.

## Flujo del invitado

1. Abrir enlace: mostrar empresa, sucursal, rol resumido, vigencia y correo enmascarado.
2. Sin cuenta: completar nombre/apellidos, teléfono opcional, contraseña y confirmación en pantalla dedicada.
3. Correo se obtiene de invitación en backend; no aceptar un correo sustituto enviado por formulario.
4. Crear cuenta pendiente y perfil usando contrato de aplicación de `cuenta`; enviar OTP mediante flujo actual.
5. Conservar retorno válido de invitación durante OTP, reenvío, errores y login; almacenamiento transitorio en `sessionStorage`, no token persistente en `localStorage`.
6. Verificar correo y volver a invitación para confirmar aceptación con sesión válida.
7. Con cuenta existente: iniciar sesión con correo invitado, sin duplicar cuenta ni cambiar contraseña; perfil existente se conserva.
8. Con sesión de otro correo: explicar diferencia y permitir cambiar sesión; no aceptar ni revelar correo completo.
9. Aceptar crea/enlaza membresía, activa empleado, asigna rol y sucursal, consume token y devuelve contexto autorizado.
10. Recargar contexto, seleccionar explícitamente empresa y sucursal de invitación y abrir `/inicio`.

Registro/OTP y aceptación son etapas separadas: cuenta creada sin aceptación no recibe acceso a empresa. Una falla posterior permite reanudar con la cuenta ya registrada. Datos personales confirmados se sincronizan al empleado al aceptar; número, puesto, fechas laborales y roles permanecen bajo control administrativo.

## Tablas: reutilización y ampliación

| Tabla | Cambio propuesto |
| --- | --- |
| `invitaciones_negocio` | Añadir `sucursal_id uuid` con relación a sucursal del mismo negocio; requerido para nuevas invitaciones. |
| `asignaciones_empleado_sucursal` | Crear/reactivar asignación elegida al aceptar; conservar único empleado/sucursal y principal activa única. |
| `empleados` | Reutilizar vínculo `membresia_id`, correo y estado; permitir corregir correo antes de vincular cuenta y reemitir; no añadir contraseña. |
| `usuarios`, `perfiles_usuario` | Reutilizar cuenta global y perfil personal; sin empresa o sucursal fija en usuario. |
| `membresias_negocio`, `roles`, `roles_membresia`, `permisos`, `permisos_rol` | Reutilizar acceso y permisos; rol obligatorio para nuevas invitaciones. |

**Tablas nuevas obligatorias: ninguna.** No crear `invitaciones_sucursal` ni membresías duplicadas. Una tabla `eventos_invitacion_negocio` sería ampliación futura si se requiere historial detallado de cada acción; esta versión conserva filas emitidas y estados actuales.

- Seguimiento de aceptación usa `estado`, `aceptado_en`, `aceptado_por_usuario_id`, `creado_en` y `expira_en` existentes; no necesita tracking de apertura ni tablas adicionales.
- Listado determina vencimiento desde reloj del servidor aunque estado persistido todavía diga pendiente; «Sin aceptar» agrupa pendientes y expiradas, y canceladas se consultan en historial.

- Relación compuesta propuesta `(negocio_id, sucursal_id)` con `sucursales(negocio_id, id)`; añadir unicidad de destino necesaria para FK.
- Comprobar también empleado y rol del mismo negocio; hash único e índice de listado `(negocio_id, sucursal_id, estado)`.
- Añadir índice único parcial para una invitación pendiente por `(negocio_id, empleado_id)`; expirar/cancelar anteriores antes de insertar.
- Migración aditiva e idempotente: detectar y resolver pendientes duplicadas conservando historial antes de crear índice.
- Invitaciones históricas aceptadas conservan `sucursal_id` nulo; no inventar asignaciones.
- Pendientes antiguas sin sucursal se cancelan y quedan visibles; administrador emite nuevo enlace con sucursal elegida. Informar impacto antes de aplicar migración.
- Antes de activar restricciones, reportar miembros activos sin empleado/asignación y resolverlos con administrador; no conceder todas las sucursales como compatibilidad automática.
- No usar sucursal principal como backfill automático de permisos; no editar `base.MD`.

## Aceptación y autorización

- En transacción bloquear invitación y empleado; volver a validar estado `pendiente`, `expira_en > ahora`, empresa/sucursal/rol activos y empleado elegible.
- Reenvío/corrección usa los mismos bloqueos y autorización vigente; si aceptación ganó la carrera, devolver `409` sin cambiar correo, cuenta ni acceso.
- Solo reemitir última invitación del empleado: comprobarlo con empleado bloqueado; fila anterior ya reemplazada no vuelve a generar enlaces desde historial ni por reintento.
- Validar y normalizar correo nuevo, exigir empleado sin cuenta vinculada y permiso `equipo.empleados.gestionar` además de invitación/asignación para corregirlo.
- Tras persistir reemisión, intentar SMTP. Si falla, devolver enlace nuevo copiable y error controlado; no restaurar validez del enlace anterior.
- Comparar correo normalizado con cuenta autenticada y verificada; segundo uso responde conflicto y no repite efectos.
- Crear/reutilizar membresía activa, vincular empleado, asignar rol, crear/reactivar asignación y marcar aceptación; rollback completo si falla cualquiera.
- Membresía suspendida/revocada no se reactiva automáticamente mediante enlace; requiere intervención administrativa.
- No reemplazar empleado ya enlazado con otra cuenta ni degradar propietario existente.
- Primera asignación activa queda principal; otras asignaciones existentes se conservan y principal solo cambia por acción administrativa.
- Usuario propietario puede operar todas las sucursales activas de su empresa. Miembro empleado necesita empleado activo y asignación vigente, además de permiso del módulo.
- Aplicar filtro por usuario en opciones, selector y toda API que lea/escriba datos de sucursal, incluidos inventario, compras, productos con stock y precios.
- UUID de otra empresa o sucursal no autorizada responde `404` sin exponer sus datos; permiso de módulo insuficiente responde `403`.
- Catálogos generales siguen perteneciendo al negocio: asignación de sucursal no convierte productos/proveedores en entidades exclusivas de ubicación.
- Agregar `precios.ver` y `precios.autorizar` al catálogo de permisos y proteger rutas actuales de precios; otorgarlos al propietario y solo a roles autorizados.
- Miembro con permisos administrativos no obtiene todas las sucursales por defecto; administración solo gestiona alcance autorizado.
- Emisor requiere permisos vigentes de invitación y asignación; solo delega sucursales accesibles y roles cuyos permisos puede otorgar. Nunca delega rol propietario.
- Retirar asignación o suspender empleado invalida acceso en próxima solicitud; limpiar selección/caché al recargar contexto.
- Consulta y registro públicos con límites de frecuencia; aceptación autenticada conserva CSRF. No registrar tokens, contraseñas ni enlaces en logs.

## Contratos HTTP propuestos

- Mantener prefijo `/api/v1` y rutas administrativas actuales.
- `POST /negocios/:negocioId/administracion/invitaciones`: `empleado_id`, `sucursal_id`, `rol_predeterminado_id` obligatorios; respuesta existente más sucursal, token único y `correo_enviado`.
- `GET` del mismo recurso añade filtros `sucursal_id` y `sin_aceptar`, nombre de sucursal, estado efectivo y campos existentes de emisión/vencimiento/aceptación al resumen; `DELETE /:invitacionId` conserva cancelación.
- Nuevo `POST /negocios/:negocioId/administracion/invitaciones/:invitacionId/reenviar`: cuerpo con `correo` opcional; ausente conserva destinatario, presente corrige correo del empleado autorizado. Mantiene sucursal/rol; actualiza empleado, invalida anterior e inserta nueva en una transacción.
- Respuesta de reenvío devuelve nueva invitación, token una sola vez, `expira_en` a 72 horas y `correo_enviado`; `409` si anterior fue aceptada o empleado ya quedó vinculado. Doble clic/reintento con la misma invitación anterior ya reemplazada responde conflicto y no genera más enlaces.
- `GET /invitaciones/:token`: datos públicos mínimos, correo enmascarado y acción de continuación; no devolver nombre completo del empleado ni correo completo.
- Nuevo `POST /invitaciones/:token/registro`: datos personales y contraseña; backend fija correo desde invitación, crea cuenta pendiente y devuelve desafío OTP. No consume invitación.
- `POST /invitaciones/:token/aceptar`: requiere sesión verificada del mismo correo; devuelve `aceptada`, `negocio_id`, `sucursal_id`.
- Conservar formato `codigo`, `mensaje`, `campos`; `404` inexistente, `410` vencida/cancelada, `409` aceptada o estado incompatible, `429` límite.
- Retorno después de autenticación permite solo rutas internas autorizadas; no aceptar redirecciones externas.

## Diseño visual y archivos futuros

- Compartir paleta clara del plan 1: índigo `#2C2E6A`, crema `#FAF8F3`, verde `#3F5147`, salvia, arena y durazno del login.
- Invitaciones: encabezado de sucursal visible, formulario dedicado, resumen previo, etiquetas de aceptación, vencimiento y acciones «Reenviar» / «Corregir correo y reenviar» según estado.
- Aceptación: tarjeta de bienvenida con empresa/sucursal, personalización guiada y progreso Cuenta / Verificación / Acceso.
- Marca aportada provisional; usar sus verdes/naranjas solo como inspiración secundaria, sin sustituir colores principales del login.
- Formularios principales en rutas dedicadas; móvil en una columna, foco visible, contraste y movimiento reducido. Oscuro fuera de revisión visual actual.
- Backend: `negocio` en cuatro capas para invitación/asignación/contexto; `cuenta` para registro/OTP mediante contratos de aplicación; `correo` para plantilla existente.
- Alcance expreso: composición HTTP, verificaciones de acceso en `compra`, `precio` y catálogo/inventario actualmente en raíz; no mover esos archivos ni rehacer sus modelos.
- Frontend: Equipo, aceptación, autenticación/retorno, contexto, navegación y permisos. No crear paquetes genéricos nuevos.
- Cambios amplían fases 6/7 de capability 01 y alcance de capability 12; aprobación futura debe incluir estas restricciones, sin alterar documentos aprobados ahora.

## Tareas, pruebas y aceptación

- [x] Confirmar personalización desde enlace conservando verificación de correo; decisión explícita del usuario.
- [x] Confirmar 72 horas, seguimiento de no aceptadas, reenvío y corrección de correo; solicitados explícitamente por el usuario.
- [x] Confirmar una sucursal inicial por invitación; alcance aprobado el 2026-10-01.
- [x] Implementar ampliación de invitación y migración compatible, con reporte de enlaces antiguos afectados.
- [x] Incorporar personalización desde enlace y continuidad durante registro/OTP/login.
- [ ] Implementar aceptación atómica, selección explícita y autorización por sucursal en UI/API. Hecho: aceptación atómica y selección explícita. Pendiente: autorización por sucursal.
- [ ] Incorporar permisos de precios y revisar roles que pueden delegarse; impedir otorgar `PROPIETARIO` mediante invitación de empleado.
- [ ] Probar cuenta nueva/existente, correo distinto, OTP fallido y retorno después de verificación.
- [ ] Probar token vencido, cancelado, reemitido y dos aceptaciones simultáneas; solo una produce efectos.
- [ ] Probar vencimiento exacto a 72 horas, filtro «Sin aceptar» y actualización de estado después de aceptar.
- [ ] Probar corrección de correo: histórico conservado, enlace anterior inválido, nuevo correo obligatorio y cuenta global anterior intacta.
- [ ] Probar carrera aceptación/reenvío, doble clic de reenvío, permiso insuficiente y corrección/reenvío de invitación ya aceptada.
- [ ] Probar rollback, rol desactivado, empleado suspendido, sucursal archivada y retirada de asignación.
- [ ] Probar API directa, selector y datos entre dos empresas y dos sucursales; incluir precios y lecturas agregadas.
- [ ] Probar SMTP fallido, enlace copiable, móvil/escritorio, teclado y contraste.
- [ ] Ejecutar Go, pruebas Angular/build con Node compatible y Graphify si existe grafo; registrar resultados y fallos previos.

Criterios: empleado completa cuenta desde enlace, acepta una vez y entra a sucursal indicada sin crear empresa; API impide operar sucursales no asignadas; permisos de módulo siguen vigentes; enlaces expiran a las 72 horas; administrador distingue no aceptadas y puede reenviar o corregir correo; reemisión invalida enlace anterior y renueva 72 horas; fallo de correo no elimina invitación; histórico y cuentas existentes se preservan.

## Evidencia y decisiones pendientes

- Revisión documental y contraste con código de ambos proyectos realizados el 2026-10-01; otro proyecto solo consultado.
- Pruebas funcionales, migraciones y validación visual pendientes: no se implementó código.
- Decisión confirmada: «Personalizar desde enlace y verificar correo (recomendado)». Mantener OTP actual.
- Decisión confirmada: vigencia de 72 horas y seguimiento de aceptación con reenvío o corrección de correo. Sigue pendiente confirmar alcance inicial de una sucursal.
- Revisión de este borrador: controles, contratos, modelo reutilizado, concurrencia y pruebas actualizados; límite de 300 líneas, espacios y estado `Borrador` comprobados; `git diff --check` correcto para archivos rastreados.
- Ninguna aprobación de implementación, Git, correo real o despliegue registrada para estos planes.

## Alcance aprobado y seguimiento

Decisiones del usuario, 2026-10-01:

- Este archivo actúa como capability aprobada; se mantiene en `capabilities/thrs` sin crear carpeta `NN-nombre`.
- Entrega actual: **flujo de invitación** (sucursal en invitación, migración, 72 horas, reenvío, corrección de correo, registro desde enlace, OTP, aceptación atómica con asignación).
- Fuera de esta entrega: filtro por sucursal asignada en contexto, inventario, compras, productos y precios; permisos `precios.ver` / `precios.autorizar`; revisión de roles delegables por permisos. Siguen pendientes dentro de este plan.
- Una sucursal por invitación, confirmado.
- Vistas nuevas siguen la opción A del diseño (crema, índigo, mosaico). Variables de color definidas solo en los componentes nuevos; sin tocar `styles.css`. Modo oscuro omitido por ahora.
- Componentes nuevos sin sufijo `.component` en archivo ni `Component` en clase (convención de Angular 20+).
- Alcance extra aprobado: recolorear la topbar con variables locales en su propio CSS. Icono y nombre de marca no cambian hasta que el colega del plan 1 suba su cambio.
- Pruebas locales con Mailpit; sin tocar producción ni configuración.

### Implementado el 2026-10-01

- Backend `negocio`: `sucursal_id` en invitación, vigencia de 72 horas, rol y sucursal obligatorios, rol propietario no delegable, filtros `sucursal_id` y `sin_aceptar`, estado efectivo por reloj del servidor.
- Migración `MigrateInvitacionSucursal`: columna, FK compuesta `(negocio_id, sucursal_id)`, índice único de pendiente por empleado, índice de listado. Cancela pendientes antiguas sin sucursal y registra cuántas en el log de arranque.
- `POST …/invitaciones/:invitacionId/reenviar` (con `correo` opcional), `POST /invitaciones/:token/registro` y `POST /invitaciones/:token/aceptar` con respuesta `aceptada`, `negocio_id`, `sucursal_id`.
- Aceptación con bloqueo de empleado e invitación y revalidación dentro de la transacción; crea asignación de sucursal; membresía suspendida o revocada ya no se reactiva por enlace.
- `cuenta`: `RegisterWithNames` para recibir nombres y apellidos separados; adaptador en `cmd/api` como contrato entre dominios.
- Límite de frecuencia por IP en memoria para consulta y registro públicos (`interfaces/http`).
- Frontend: `equipo/invitaciones/` (lista, nueva, corregir correo) y `invitacion/aceptar-invitacion` (cuenta, verificación, acceso), con variables locales en `invitaciones-tema.css`. Se retiraron `invitaciones.component.*` y `aceptar-invitacion.component.*`.
- Login y verificación genérica respetan el retorno a la invitación guardado en `sessionStorage`.
- Topbar: paleta crema/índigo con variables locales, solo en tema claro.

### Pruebas ejecutadas el 2026-10-01

- `go vet ./...`: sin observaciones. `go test ./...`: pasa todo salvo `internal/config` (`TestProductionAcceptsStrongOTPSecret`), paquete no modificado por este trabajo.
- Angular: `ng build` correcto; `ng test`: 130 pasan, 9 fallan en `productos.component.spec.ts` y `navigation.config.spec.ts`, archivos no modificados por este trabajo.
- Recorrido HTTP contra API local en base temporal y Mailpit (27 comprobaciones, todas correctas): emisión, validaciones, 72 horas, reenvío, enlace anterior inválido, reintento sobre reemplazada, corrección de correo con histórico, registro desde enlace, correo sustituto rechazado, OTP leído de Mailpit, sesión de otro correo, dos aceptaciones simultáneas (una 200, otra 409), contexto, rol, asignación principal y sincronización de datos personales.
- Revisión visual con Chromium sin interfaz: lista, nueva invitación, corregir correo, aceptación en escritorio y móvil, enlace vencido.

### Pendientes conocidos

- El empleado invitado todavía ve todas las sucursales activas del negocio en el contexto: falta el filtro por asignación (fuera de esta entrega).
- Permisos `precios.ver` / `precios.autorizar` y validación de que el emisor solo delegue roles cuyos permisos posee.
- Pruebas de integración con base para reenvío y aceptación; pruebas de componentes Angular de las vistas nuevas.
- No probado: fallo real de SMTP desde la interfaz, rollback forzado, sucursal archivada o rol desactivado entre emisión y aceptación, cuenta existente iniciando sesión desde el navegador.
- Correo de invitación sigue diciendo «días»; con 72 horas muestra 3 días.
- Graphify no está instalado en este equipo: el grafo no se actualizó.
- `PLAN_IMPLEMENTACION.md` no se tocó porque ya tenía cambios locales del usuario.
