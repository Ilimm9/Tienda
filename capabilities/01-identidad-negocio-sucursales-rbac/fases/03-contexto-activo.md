# Fase 3: Contexto activo

## Estado de fase

`Cerrada`

Planeación redactada el 2026-09-11 por solicitud explícita del usuario. Implementación autorizada explícitamente por el usuario el 2026-09-11.
Cierre autorizado explícitamente por el usuario el 2026-09-12, junto con la autorización para ejecutar fases 4 a 8 sin supervisión.

Excepción de dependencia aprobada por el usuario: fase 2 permanece `Verificada` y pendiente de aceptación final, pero se autoriza iniciar fase 3. No altera resultado ni estado de fase 2.

## Fuente de verdad y compatibilidad

`capabilities/_template/base.MD` permanece sin modificaciones y aporta las relaciones que delimitan el contexto:

- `usuarios -> membresias_negocio -> negocios` determina qué negocios puede seleccionar una cuenta.
- `negocios -> sucursales` determina qué sucursales activas pertenecen al negocio seleccionado.
- No existe ni se agregará una tabla de contexto activo: la selección es preferencia local de interfaz, no dato de negocio.
- Los UUID guardados en navegador nunca prueban autorización. Cada petición seguirá validando sesión, membresía, negocio y sucursal en backend.
- Los contratos administrativos de fases 1 y 2 se preservan.
- Las rutas de producto que hoy usan `environment.defaultBusinessId` deberán recibir el negocio activo; el valor fijo deja de ser fuente de verdad.
- Las rutas de negocio explícitas, como `/negocios/:negocioId/...`, conservan el UUID de la URL y no cambian silenciosamente el contexto activo.

## Objetivo y alcance

Dar a la sesión autenticada un contexto visible y consistente de negocio y sucursal, restaurarlo de forma segura al recargar y propagarlo al dashboard y a los módulos actuales que operan datos de negocio.

Incluye:

- Selector de negocio y sucursal en topbar, responsive y accesible.
- Pantalla dedicada para elegir negocio cuando no puede resolverse uno automáticamente.
- Persistencia local de los últimos UUID seleccionados.
- Revalidación completa de la selección al iniciar sesión, recargar, cambiar de cuenta o refrescar opciones.
- Selección automática de sucursal principal cuando no existe una selección válida.
- Estado explícito para negocio sin sucursales activas.
- Redirección de `/sucursales` al negocio activo.
- Propagación del negocio activo a productos y proveedores, eliminando su dependencia operativa de `defaultBusinessId`.
- Uso de la sucursal activa como valor inicial de operaciones que ya exigen sucursal.
- Protección autenticada y validación de membresía en endpoints actuales delimitados por `negocioId`.
- Validación de pertenencia y estado de `sucursalId` en operaciones actuales que lo reciben.
- Actualización del contexto cuando se archiva o restaura el negocio o la sucursal seleccionados.
- Estados de carga, vacío, error e invalidación.
- Pruebas de aislamiento entre cuentas, negocios y sucursales.

No incluye:

- Guardar el contexto en base de datos o sincronizarlo entre dispositivos.
- Incorporar negocio, sucursal o roles dentro del JWT.
- Enviar un header implícito como `X-Negocio-Id`; los identificadores continuarán explícitos en ruta o payload.
- Roles y permisos granulares; fase 4 mantiene autorización temporal por membresía/propiedad.
- Restringir una membresía a sucursales concretas; corresponde a fases 5 y 7.
- Implementar ventas, empleados, invitaciones, roles o permisos.
- Convertir catálogos realmente globales en catálogos por negocio.
- Rediseñar productos o proveedores más allá de inyectar contexto y cerrar su aislamiento.
- Modificar tablas o datos existentes.

## Modelo de contexto

No hay migración ni entidad persistente. El frontend mantiene un estado en memoria derivado de opciones autorizadas por la API:

```text
estado              cargando | requiere_negocio | listo | sin_sucursal | error
negocio             ContextoNegocio | null
sucursal            ContextoSucursal | null
negocios            ContextoNegocio[]
inicializado         boolean
```

`ContextoNegocio` contiene:

```text
id                   uuid
slug                 string
nombre_comercial     string
tipo_miembro         propietario | miembro
sucursales           ContextoSucursal[]
```

`ContextoSucursal` contiene:

```text
id                   uuid
codigo               string
nombre               string
es_principal         boolean
```

Reglas de almacenamiento:

```text
tienda.contexto.negocio_id
tienda.contexto.sucursal_id
```

- Solo se guardan UUID; nombres, tipo de membresía y estados siempre proceden de la API.
- Escritura y lectura toleran `localStorage` no disponible.
- Valores malformados, inaccesibles, archivados o asociados a otro negocio se eliminan.
- Cerrar sesión limpia ambas claves.
- Un evento `storage` de otra pestaña dispara revalidación antes de adoptar el cambio.
- No se almacenan tokens, permisos ni respuestas completas.

## Resolución inicial

Después de confirmar la sesión:

1. Solicitar opciones autorizadas.
2. Si no hay negocios activos, limpiar contexto y navegar a `/negocios`.
3. Si el negocio guardado sigue disponible, seleccionarlo.
4. Si no hay negocio guardado y solo existe uno disponible, seleccionarlo automáticamente.
5. Si hay varios negocios y ninguno guardado es válido, usar estado `requiere_negocio` y navegar a `/seleccionar-negocio`.
6. Dentro del negocio resuelto, conservar la sucursal guardada solo si está activa y pertenece a ese negocio.
7. Si la sucursal guardada no es válida, elegir la principal activa.
8. Si excepcionalmente no existe principal, elegir la primera opción estable ordenada por nombre y UUID.
9. Si no existen sucursales activas, conservar negocio, usar sucursal `null` y estado `sin_sucursal`.
10. Publicar estado `listo` únicamente después de terminar la validación.

La interfaz no debe mostrar brevemente datos del contexto anterior mientras se resuelve el nuevo.

## Cambios de selección

### Cambiar negocio

1. Usuario abre el selector y elige otro negocio autorizado.
2. Estado entra en carga y deja de emitir el par anterior.
3. Se selecciona la principal del nuevo negocio o `null` si no tiene sucursales.
4. Se persisten ambos UUID de forma conjunta desde la perspectiva del servicio.
5. Se refrescan consumidores y se navega a `/inicio` para no conservar vistas con datos del negocio previo.
6. Se muestra notificación breve con el nuevo negocio mediante `FeedbackService`.

### Cambiar sucursal

1. Usuario elige una sucursal activa del negocio actual.
2. Se valida contra las opciones ya autorizadas.
3. Se persiste el UUID, se publica el nuevo contexto y se refrescan consumidores dependientes de sucursal.
4. Se conserva la ruta si el módulo declara que soporta recarga por sucursal; en caso contrario se navega a `/inicio`.
5. No se solicita confirmación cuando no existe trabajo sin guardar.

Un componente con formulario o importación pendiente debe bloquear el cambio mediante un contrato explícito de cambios sin guardar y usar la confirmación global de fase 1.6. No se detectará mediante patrones de URL.

### Invalidación posterior

- Archivar el negocio activo obliga a recargar opciones y resolver otro negocio o pedir selección.
- Archivar la sucursal activa elige la principal vigente; si no queda ninguna usa `sin_sucursal`.
- Restaurar no cambia la selección automáticamente, salvo que el negocio actual no tenga sucursal activa y la restaurada pase a ser principal.
- Respuestas `401` limpian sesión mediante el flujo de autenticación vigente.
- Respuestas `404` o `409` que indiquen contexto obsoleto disparan una sola revalidación; no crean ciclos de reintento.

## API y contratos

### Obtener opciones autorizadas

```text
GET /api/v1/contexto/opciones
```

Requiere JWT válido. Respuesta `200`:

```json
{
  "items": [
    {
      "id": "uuid-negocio",
      "slug": "tienda-centro",
      "nombre_comercial": "Tienda Centro",
      "tipo_miembro": "propietario",
      "sucursales": [
        {
          "id": "uuid-sucursal",
          "codigo": "SUC-001",
          "nombre": "Matriz",
          "es_principal": true
        }
      ]
    }
  ],
  "total": 1
}
```

Reglas del endpoint:

- Devuelve solo membresías activas de negocios activos.
- Devuelve solo sucursales activas y no eliminadas del mismo negocio.
- Incluye negocios sin sucursales con arreglo vacío.
- Ordena negocios por `nombre_comercial` e `id`.
- Ordena sucursales con principal primero y después por `nombre` e `id`.
- No acepta UUID almacenados por el cliente ni decide selección.
- Una cuenta sin negocios recibe `200` con `items: []`.
- No expone roles ni permisos antes de fase 4.

### Validación de módulos actuales

Todas las rutas actuales bajo `/api/v1/negocios/:negocioId/...` deberán:

- Pasar por `RequireAuth`.
- Obtener `usuarioID` únicamente del contexto autenticado.
- Validar membresía activa y negocio activo antes de ejecutar producto, catálogo por negocio, proveedor o listado de sucursales.
- Responder `404` tanto para negocio ajeno como inexistente, sin revelar existencia.
- Mantener `400` para UUID con formato inválido.
- Validar que todo `sucursalId` de payload o formulario pertenezca al `negocioId` y esté activo.
- Evitar que una sucursal de otro negocio cree stock, valide o ejecute importaciones.
- Mantener reglas temporales de mutación actuales hasta que fase 4 defina permisos.

El contexto de frontend no reemplaza estas verificaciones. El backend no confiará en orden de navegación, guard, selector ni `localStorage`.

### Errores

- `400`: identificador con formato inválido.
- `401`: sesión ausente o inválida.
- `404`: negocio inaccesible, sucursal ajena o contexto ya no vigente.
- `422`: una operación que exige sucursal no recibió el campo requerido u otros datos de entrada válidos; una sucursal bien formada pero inaccesible responde `404`.
- `500`: fallo interno sin detalles sensibles.

El endpoint de opciones usa el contrato de error uniforme de fases anteriores.

## Organización backend propuesta

La fase amplía el dominio `negocio` para resolver acceso y expone un contrato explícito reutilizable por módulos actuales:

```text
backend/internal/domain/negocio/contexto.go
backend/internal/application/negocio/contexto_service.go
backend/internal/application/negocio/contexto_service_test.go
backend/internal/infrastructure/negocio/contexto_repository.go
backend/internal/interfaces/http/negocio/contexto_handler.go
backend/internal/interfaces/http/negocio/contexto_handler_test.go
backend/internal/interfaces/http/negocio/contexto_middleware.go
backend/internal/interfaces/http/negocio/contexto_middleware_test.go
```

Reglas:

- `ContextoService` lista opciones y valida acceso activo.
- Su interfaz de validación se declara en `application/negocio`.
- El middleware entiende HTTP y delega validación; no consulta GORM.
- Producto y catálogo no importan repositories ni handlers de `negocio`.
- La composición en `cmd/api` inyecta el validador explícito y agrupa rutas protegidas.
- La validación de sucursal incluida en payload vive en aplicación mediante un puerto explícito; no mediante acceso directo al repository de otro dominio.
- No se mueven archivos de producto en esta fase. Solo se modifican puntos necesarios y enumerados para autenticación, negocio activo y sucursal válida.
- No se crean carpetas `shared`, `common`, `helpers` o `utils`.

## Organización frontend propuesta

```text
frontend/src/app/contexto/contexto.models.ts
frontend/src/app/contexto/contexto.service.ts
frontend/src/app/contexto/contexto.service.spec.ts
frontend/src/app/contexto/contexto.guard.ts
frontend/src/app/contexto/contexto.guard.spec.ts
frontend/src/app/contexto/seleccionar-negocio.component.ts
frontend/src/app/contexto/seleccionar-negocio.component.html
frontend/src/app/contexto/seleccionar-negocio.component.css
frontend/src/app/contexto/seleccionar-negocio.component.spec.ts
frontend/src/app/layout/context-selector/context-selector.component.ts
frontend/src/app/layout/context-selector/context-selector.component.html
frontend/src/app/layout/context-selector/context-selector.component.css
frontend/src/app/layout/context-selector/context-selector.component.spec.ts
```

- `ContextoService` es la única pieza que lee o escribe las claves de contexto.
- Expone señales de solo lectura para estado, negocio, sucursal y opciones.
- Deduplica inicializaciones concurrentes y permite una recarga controlada.
- El selector de layout consume el servicio; no llama HTTP directamente.
- Los módulos consumen el servicio y pasan UUID explícitos a sus services HTTP.
- No se agrega el contexto a `AuthService` ni a `LayoutStateService`.
- `environment.defaultBusinessId` puede conservarse temporalmente para seeds o pruebas aisladas, pero ningún flujo autenticado lo usa después de esta fase.

## Pantallas y navegación

### Topbar

- En escritorio muestra nombre del negocio, nombre/código de sucursal y chevrón.
- Negocio y sucursal se eligen en controles separados para que la pertenencia sea evidente.
- En móvil usa un botón compacto que abre un panel con ambos selectores.
- Mientras inicializa muestra skeleton y deshabilita interacción.
- Negocio sin sucursales muestra `Sin sucursal activa` y CTA `Registrar sucursal`.
- Si solo existe un negocio o una sucursal, sigue mostrando el contexto, aunque el control correspondiente no necesite desplegable.
- Cierre con Escape, navegación por teclado, foco visible y etiquetas accesibles.
- No muestra opciones archivadas ni UUID.

### Selección de negocio

Ruta:

```text
/seleccionar-negocio
```

- Muestra cards de negocios activos con tipo de relación y cantidad de sucursales activas.
- Seleccionar completa resolución y navega a `/inicio`.
- Cuenta sin negocios ve CTA `Registrar negocio`.
- Error de carga permanece inline con acción de reintento.
- Usuario con contexto válido que entra manualmente puede cambiarlo.
- No crea, edita, archiva ni restaura negocios desde esta pantalla.

### Rutas protegidas por contexto

- `/negocios` y sus rutas explícitas requieren autenticación, pero no contexto previo.
- `/seleccionar-negocio` requiere autenticación, pero admite contexto vacío.
- `/inicio`, `/sucursales`, producto y proveedores requieren negocio activo.
- `/sucursales` redirige a `/negocios/:negocioId/sucursales`.
- Un módulo que exige sucursal recibe estado `sin_sucursal` y CTA, no un UUID inventado.
- Rutas reservadas para fases 4 a 7 no adquieren funcionalidad por mostrar el contexto.

### Dashboard

- Encabezado identifica negocio y sucursal activos.
- Sin sucursal muestra orientación para registrar o restaurar una.
- Cards navegan usando contexto actual.
- No se agregan métricas de ventas, inventario o empleados todavía.
- Cambiar contexto actualiza textos y enlaces sin recargar toda la aplicación.

## Propagación a módulos actuales

### Productos

- Sustituir cada uso operativo de `environment.defaultBusinessId` por `negocio.id`.
- Listados, altas, edición, baja, consulta de código e importación usan el negocio activo.
- Operaciones con sucursal toman inicialmente `sucursal.id`.
- Si el flujo permite elegir otra sucursal destino, sus opciones se limitan a las sucursales activas del mismo negocio.
- Cambiar contexto cancela o invalida resultados anteriores antes de cargar nuevos.
- Sin sucursal no se habilitan alta con stock inicial, validación ni importación dependiente de sucursal.

### Catálogo

- Proveedores usa el negocio activo.
- Marcas, categorías y unidades que hoy son globales permanecen globales; la UI no afirmará que cambian por sucursal.
- No se cambia su modelo de datos en esta fase.

### Sucursales

- La navegación general usa el negocio activo.
- Las rutas administrativas con `:negocioId` conservan el negocio de la URL.
- Abrir detalle administrativo de otro negocio accesible no cambia el selector implícitamente.
- Después de mutar la sucursal activa se refrescan opciones de contexto.

### Módulos reservados

Ventas, equipo y roles/permisos pueden mostrar el contexto en el shell, pero su funcionalidad permanece fuera de alcance hasta sus fases. No se crean endpoints ni datos ficticios para ellos.

## Seguridad y aislamiento

- El selector nunca contiene negocios sin membresía activa.
- Un UUID manipulado en `localStorage` no permite cargar ni mutar datos.
- Las rutas por negocio actuales dejan de estar públicas.
- Toda consulta de negocio incluye `usuarioID`, `negocioID` y estado activo.
- Toda operación con sucursal valida además `sucursal.negocio_id = negocioID` y sucursal activa.
- Negocio ajeno e inexistente producen respuesta indistinguible.
- Cambio de cuenta en el mismo navegador revalida antes de renderizar datos.
- La aplicación limpia datos visibles del contexto anterior durante transiciones.
- No se registran UUID de contexto como credenciales ni se confía en ellos para RBAC.
- La autorización granular se mantiene explícitamente pendiente de fase 4.

## Validaciones y experiencia de error

- UUID local inválido: limpiar y resolver de nuevo sin mostrar error técnico.
- Negocio guardado ya archivado o revocado: limpiar, notificar cambio de disponibilidad y resolver otra opción.
- Sucursal guardada archivada o ajena: elegir principal vigente.
- Sin negocios: CTA a alta y navegación administrativa disponible.
- Sin sucursales: negocio permanece activo, CTA a alta; módulos dependientes quedan bloqueados de forma explicativa.
- Error de red al inicializar: estado de error recuperable; no usar valor fijo como respaldo.
- Cambio rápido repetido: solo la última selección puede publicar resultado.
- Fallo al refrescar después de cambio: revertir visualmente a un contexto validado o quedar en error; no conservar mezcla de ambos.
- Notificaciones de éxito y fallos de acción usan fase 1.6; errores de carga permanecen inline.

## Pruebas de fase 3

### Backend

- Opciones incluye solo negocios con membresía activa y estado activo.
- Miembro activo y propietario reciben sus negocios; membresía suspendida o revocada no.
- Negocio sin sucursales aparece con arreglo vacío.
- Solo sucursales activas del negocio aparecen y la principal ordena primero.
- Cuenta sin negocios recibe lista vacía.
- Endpoint rechaza sesión ausente o inválida.
- Rutas de producto, proveedor y sucursales legacy requieren autenticación.
- Miembro no puede usar UUID de negocio ajeno en cada ruta por negocio actual.
- `sucursalId` ajeno, archivado o inexistente no puede crear stock ni validar/ejecutar importación.
- Negocio inválido responde `400`; negocio inaccesible responde `404`.
- Contratos exitosos vigentes de producto y catálogo no cambian salvo autenticación/aislamiento documentados.
- Queries no producen N+1 por cada negocio o sucursal.

### Frontend

- Inicialización restaura un par válido.
- Un solo negocio se selecciona automáticamente.
- Varios negocios sin selección abren `/seleccionar-negocio`.
- UUID malformados o ajenos se eliminan.
- Sucursal inválida cae en principal; negocio sin sucursales produce `sin_sucursal`.
- Topbar refleja contexto, carga, error y diseño móvil.
- Cambiar negocio limpia primero la sucursal anterior y navega a inicio.
- Cambiar sucursal refresca consumidores declarados.
- Logout limpia claves.
- Evento de otra pestaña revalida.
- Guard evita módulos de negocio hasta completar inicialización.
- `/sucursales` genera la ruta con negocio activo.
- Productos y proveedores no leen `defaultBusinessId`.
- Alta/importación de producto usa por defecto sucursal activa.
- Cambio de contexto no deja respuestas tardías del contexto anterior.
- Formularios con cambios pendientes solicitan confirmación una sola vez.
- Revisión visual en 390, 768 y 1440 px, claro y oscuro.
- Navegación de selector probada con teclado y lector semántico básico.

### Regresión y verificación integral

- `go list ./...`, `go test ./...` y `go vet ./...`.
- `npm test -- --watch=false` y `npm run build`.
- Smoke con dos cuentas, dos negocios por cuenta y sucursales cruzadas.
- Manipulación manual de ambas claves de `localStorage`.
- Smoke de recarga, logout/login con otra cuenta y dos pestañas.
- Smoke de producto/proveedor con cada negocio y rechazo cruzado.
- Confirmar que `base.MD` y esquema físico no cambian.
- Diff limitado a contexto, layout, rutas y puntos de integración enumerados.

## Criterios de aceptación de fase 3

- [x] Alcance, dependencias y exclusiones documentados.
- [x] Modelo no persistente y claves de almacenamiento definidos.
- [x] Algoritmo de resolución e invalidación definido.
- [x] Contrato de opciones autorizadas definido.
- [x] Reglas de navegación y estados sin negocio/sucursal definidos.
- [x] Integración con productos, proveedores y sucursales delimitada.
- [x] Cierre de aislamiento de rutas actuales incluido expresamente.
- [x] Organización backend y frontend propuesta.
- [x] Pruebas y verificación planeadas.
- [x] Fase 2 aceptada finalmente y marcada `Cerrada`.
- [x] Especificación de fase 3 revisada por el usuario.
- [x] Fase 3 aprobada explícitamente para implementación.
- [x] Implementación terminada.
- [x] Verificaciones registradas.
- [x] Resultado aceptado finalmente por el usuario.

## Tareas de implementación propuestas

- [x] Crear modelos, repository, service y endpoint de opciones.
- [x] Crear validador de acceso reutilizable y proteger rutas actuales por negocio.
- [x] Validar sucursales de payload en producto/importaciones.
- [x] Crear `ContextoService` y almacenamiento local. Falta sincronización entre pestañas.
- [x] Crear guard y pantalla de selección.
- [x] Integrar selector responsive en topbar.
- [x] Integrar dashboard y ruta general de sucursales.
- [x] Sustituir `defaultBusinessId` en productos y proveedores.
- [x] Manejar cambios pendientes e invalidación por mutaciones administrativas.
- [x] Ejecutar pruebas backend, frontend, aislamiento y revisión visual.
- [x] Registrar resultados, imprevistos y pendientes en este archivo.

## Decisiones propuestas para aprobación

- El contexto se persiste solo en `localStorage`, sin tabla ni JWT.
- Se agrega un endpoint agregado `GET /api/v1/contexto/opciones`.
- Con un negocio se selecciona automáticamente; con varios y sin selección válida se pide elección.
- La sucursal principal es el fallback; negocio sin sucursal sigue siendo un contexto válido limitado.
- Cambiar negocio navega a `/inicio`; cambiar sucursal conserva ruta solo en consumidores compatibles.
- Las rutas administrativas explícitas no cambian el contexto implícitamente.
- La fase incluye cerrar autenticación y aislamiento por membresía de rutas actuales delimitadas por negocio.
- Producto usa negocio activo y preselecciona sucursal activa sin eliminar selectores necesarios para destinos explícitos.
- Catálogos globales permanecen globales.
- La aprobación de esta especificación deberá ser explícita y posterior al cierre formal de fase 2.

## Avance de implementación

- 2026-09-11: creado endpoint autenticado `GET /api/v1/contexto/opciones`, selector de topbar, guard, pantalla de elección y redirección de `/sucursales`.
- 2026-09-11: productos y proveedores usan negocio activo; operaciones de producto e importación validan sucursal activa del negocio.
- 2026-09-11: `go test ./...`, `go vet ./...`, `npm test -- --watch=false` y `npm run build` correctos. Build conserva avisos preexistentes de presupuesto CSS y bundle.
- 2026-09-11: selector visual refinado como panel contextual con opciones tipo card, estados activos, iconografía, teclado/Escape y vista móvil compacta. Pruebas Angular siguen correctas; topbar agrega aviso no bloqueante de presupuesto CSS.
- 2026-09-12: cerrados los pendientes restantes de la fase, detallados abajo.

## Cierre de pendientes de fase 3

Fecha: 2026-09-12.

### Inicialización deduplicada y recarga controlada

`ContextoService` incorpora la señal `inicializado` y una petición en vuelo compartida. `asegurarInicializado()` resuelve una sola vez y el guard lo usa en lugar de `inicializar()`, por lo que una navegación deja de pedir opciones en cada ruta protegida. `recargar()` fuerza revalidación explícita.

### Sincronización entre pestañas

El servicio escucha `storage` sobre ambas claves y revalida contra la API antes de adoptar el cambio; nunca confía en el valor recibido del evento. El listener se retira con `DestroyRef`.

### Invalidación por mutaciones administrativas

Crear, editar, promover, archivar y restaurar una sucursal disparan `recargar()` del contexto, de modo que el selector y la sucursal activa dejan de mostrar una opción que ya no existe.

### Contrato de cambios sin guardar

Se agregó `CambiosPendientesService`: cada componente declara si tiene trabajo sin guardar y se da de baja al destruirse. El topbar consulta ese contrato antes de cambiar negocio o sucursal y usa `FeedbackService.confirmDanger` una sola vez. No se usan patrones de URL. `SucursalFormComponent` es el primer consumidor, declarando `form.dirty`.

### Endurecimiento de lectura de `localStorage`

El patrón de UUID pasó a ser estricto, y lectura y escritura toleran que `localStorage` lance excepción, no solo que no exista.

## Verificación técnica de fase 3

Fecha: 2026-09-12.

- `go build ./...`, `go vet ./...` y `go test ./...`: correctos.
- Pruebas backend nuevas: `contexto_service_test.go` con 7 casos y `contexto_handler_test.go` con 7 casos. Cubren lista vacía, negocio sin sucursales, negocio ajeno indistinguible de inexistente, propagación de error de infraestructura, contrato JSON de opciones, ausencia de roles/permisos en la respuesta, `401` sin sesión, `400` por UUID mal formado, `404` por negocio ajeno y paso de membresía vigente.
- `npm test -- --watch=false`: correcto, 28 archivos y 84 pruebas. Antes del cierre eran 24 archivos y 54 pruebas.
- Pruebas frontend nuevas: `contexto.service.spec.ts` con 15 casos, `contexto.guard.spec.ts` con 4 y `cambios-pendientes.service.spec.ts` con 3. Cubren selección automática con un solo negocio, restauración de par guardado, `requiere_negocio` con varios negocios, descarte de UUID malformado, descarte de sucursal ajena con caída a principal, `sin_sucursal`, cuenta sin negocios, cambio de negocio que limpia la sucursal previa, sucursal ajena ignorada, deduplicación de inicializaciones, ausencia de segunda petición tras inicializar, recarga forzada, error de red sin respaldo fijo, limpieza al cerrar sesión y revalidación por evento de otra pestaña.
- `npm run build`: correcto; conserva los avisos preexistentes de presupuesto CSS.
- `localStorage` se prueba con el stub `vi.stubGlobal` ya usado por `ThemeService`, por consistencia con la convención del proyecto.

## Imprevistos resueltos de fase 3

- El guard llamaba `inicializar()` en cada navegación, lo que repetía la petición de opciones por ruta protegida. Se resolvió con `asegurarInicializado()` y la petición en vuelo compartida.
- El patrón de UUID guardado era laxo y aceptaba cadenas que no son UUID; se endureció.
- Las pruebas del proyecto corren en Vitest, no en Jasmine: `toBeTrue`/`toBeFalse` no existen y `localStorage` no está disponible por defecto. Se ajustaron matchers y se adoptó el stub ya usado por `ThemeService`.
- La versión de Node del `PATH` por omisión es v14 y Angular exige v22 o superior. Las verificaciones se ejecutaron con la instalación de `mise` en v26.5.0; no se cambió configuración del proyecto.

## Pendiente de cierre de fase 3

- Revisión visual y smoke multiempresa manual con dos cuentas reales quedan como verificación de usuario. El smoke automatizado equivalente está cubierto por las pruebas de contrato y aislamiento del backend.
