# Fase 2: Sucursales

## Estado de fase

`Cerrada`

Especificación reconstruida el 2026-09-11 a partir de `capabilities/_template/base.MD`, referencias visuales, código vigente y decisiones del usuario. Autorizada explícitamente para implementación el 2026-09-11.
Implementación y verificación técnica terminadas el 2026-09-11. Cierre autorizado explícitamente por el usuario el 2026-09-12.

## Fuente de verdad y compatibilidad

`capabilities/_template/base.MD` es la base inicial del modelo y no debe modificarse. Fase 2 seguirá su tabla `sucursales` siempre que no implique pérdida de datos ni romper producto.

Orden de decisión:

1. Preservar datos existentes y trabajo ajeno.
2. Mantener contratos actuales consumidos por producto.
3. Aplicar nombres, tipos y relaciones de `base.MD`.
4. Documentar cualquier excepción temporal necesaria.

Excepciones conocidas:

- Base define `nombre varchar(160)`, mientras tabla legacy admite 180. La aplicación aceptará máximo 160, pero migración no estrechará destructivamente la columna existente. Una DB nueva usará límite 160 cuando producto deje el modelo legacy.
- Campo `direccion text` legacy permanecerá físicamente mientras producto use su modelo actual. Administración usará `direccion_id` y no volverá a escribir texto legacy.
- Modelo legacy de producto permanecerá intacto. Dominio `negocio` tendrá el modelo administrativo de la misma tabla como puente temporal documentado, sin duplicar filas ni crear otra tabla.
- GET simple de sucursales usado por producto permanecerá sin cambios. CRUD administrativo tendrá rutas separadas.

## Objetivo y alcance

Permitir administrar las ubicaciones operativas de un negocio accesible: listar, crear, consultar, editar, establecer principal, archivar y restaurar sucursales con dirección estructurada.

Incluye:

- CRUD administrativo aislado por negocio.
- Código interno único e inmutable.
- Una sola sucursal principal activa por negocio.
- Archivado lógico mediante `activo` y `eliminado_en`.
- Migración y backfill de sucursales legacy.
- Listado responsive con cards.
- Formularios y detalle mediante rutas dedicadas.
- Acceso desde detalle de negocio.
- Notificaciones y confirmaciones mediante fase 1.6.
- Compatibilidad explícita con endpoint y seed usados por producto.

No incluye:

- Contexto global de negocio o sucursal; corresponde a fase 3.
- RBAC granular; corresponde a fase 4.
- Empleados asignados; corresponde a fases 5 y 7.
- Límites de sucursales por plan, suscripciones o facturación.
- Inventario, terminales, ventas, folios ni configuración de tickets.
- Edición de archivos de producto, catálogo, proveedores o inventario.
- Eliminación física.

## Flujo funcional

### Entrada desde negocio

1. Usuario abre detalle de un negocio.
2. Acción `Administrar sucursales` abre `/negocios/:negocioId/sucursales`.
3. Si no existen activas, CTA abre formulario de primera sucursal.
4. Ruta general `/sucursales` redirige a `/negocios` hasta implementar contexto global en fase 3.

### Creación

1. Propietario abre `/negocios/:negocioId/sucursales/nueva`.
2. Captura código, nombre, teléfono opcional y dirección opcional.
3. Puede solicitar que sea principal.
4. Si es la primera activa, backend la fuerza como principal.
5. Backend crea dirección y sucursal atómicamente.
6. Frontend muestra `Sucursal registrada` y navega al detalle.

### Cambio de principal

1. Propietario edita una sucursal secundaria activa y marca `Sucursal principal`.
2. Backend degrada principal anterior y promueve nueva dentro de misma transacción.
3. No se permite desmarcar directamente la principal; debe promoverse otra.

### Archivado y restauración

- Archivar requiere SweetAlert2 y conserva todos los datos.
- Si principal tiene otras sucursales activas, archivado responde conflicto hasta promover otra.
- Si principal es la última activa, archivado se permite y negocio queda sin principal.
- Restauración no pide confirmación.
- Si no existe principal activa, restaurada se vuelve principal; de otro modo vuelve como secundaria.

## Modelo de datos

### `sucursales`

```text
id                  uuid primary key
negocio_id          uuid not null references negocios(id)
codigo              varchar(40) not null
nombre              varchar(160) not null lógico; ver excepción legacy
telefono            varchar(30) null
direccion_id        uuid null references direcciones(id)
es_principal        boolean not null default false
activo              boolean not null default true
creado_en           timestamptz not null
actualizado_en      timestamptz not null
eliminado_en        timestamptz null
```

Índices y constraints:

- Único case-insensitive `(negocio_id, lower(codigo))`.
- Índice `(negocio_id, nombre)`.
- Índice `(negocio_id, activo)` para listados.
- Único parcial por negocio donde `es_principal = true AND activo = true`.
- Código no puede quedar vacío después de normalizar.
- Principal archivada siempre queda con `es_principal = false`.

Dirección reutiliza `direcciones` de fase 1. Cada sucursal posee su propia fila; compartir estructura no significa compartir registro.

## Migración y backfill

Migración será aditiva, transaccional e idempotente:

1. Agregar `codigo`, `direccion_id`, `es_principal` y `eliminado_en` si faltan.
2. Generar códigos deterministas `SUC-001`, `SUC-002`, etc. por negocio, ordenando por `creado_en` e `id`.
3. Resolver colisiones con códigos existentes sin sobrescribirlos.
4. Para `direccion` legacy no vacía, crear `direcciones` con país `MX` y texto original en `referencias`; asignar `direccion_id`.
5. Elegir como principal la sucursal activa más antigua de cada negocio que no tenga principal.
6. Desmarcar principales archivadas o duplicadas antes de crear índice único parcial.
7. Establecer `codigo` como obligatorio después del backfill.
8. Crear claves foráneas e índices solo si no existen.
9. Conservar columna `direccion` legacy y registros sembrados.

No truncar nombres, no borrar direcciones y no regenerar códigos en ejecuciones posteriores.

## Reglas de negocio

- Código se recorta, convierte a mayúsculas y acepta 2 a 40 caracteres con patrón `[A-Z0-9][A-Z0-9_-]{1,39}`.
- Código es único por negocio sin distinguir mayúsculas y no aparece en input de actualización.
- Nombre se recorta y requiere entre 2 y 160 caracteres.
- Nombres repetidos están permitidos; código diferencia sucursales.
- Teléfono vacío se guarda como `null`; máximo 30 caracteres.
- Dirección vacía no crea fila.
- Al actualizar dirección: campo ausente conserva, `null` elimina relación y objeto reemplaza valores.
- Crear siempre produce sucursal activa; no existe selector para crear archivada.
- Negocio debe estar activo para crear, editar, promover, archivar o restaurar sucursal.
- Membresía activa permite lectura.
- Solo `tipo_miembro = propietario` permite mutaciones hasta fase 4.
- Sucursal solicitada debe pertenecer al negocio de la ruta; de otro modo responde no encontrada.
- Archivar última sucursal activa está permitido.
- Archivar o restaurar una sucursal ya en ese estado produce conflicto.

## API administrativa

Todas las rutas usan JWT y validan membresía:

```text
GET    /api/v1/negocios/:negocioId/administracion/sucursales
POST   /api/v1/negocios/:negocioId/administracion/sucursales
GET    /api/v1/negocios/:negocioId/administracion/sucursales/:sucursalId
PATCH  /api/v1/negocios/:negocioId/administracion/sucursales/:sucursalId
DELETE /api/v1/negocios/:negocioId/administracion/sucursales/:sucursalId
POST   /api/v1/negocios/:negocioId/administracion/sucursales/:sucursalId/restaurar
```

Endpoint legacy preservado sin cambios:

```text
GET /api/v1/negocios/:negocioId/sucursales
```

Sigue devolviendo opciones activas `{id,nombre}` para producto. No se mueve a handler administrativo en esta fase.

### Listado

Query:

```text
estado=activo|archivado   default activo
buscar=texto              opcional, nombre o código
```

Respuesta `200`:

```json
{
  "items": [
    {
      "id": "uuid",
      "negocio_id": "uuid",
      "codigo": "SUC-001",
      "nombre": "Matriz Reforma",
      "telefono": "+52 55 1234 5678",
      "direccion_resumida": "Ciudad de México, CDMX",
      "es_principal": true,
      "estado": "activo",
      "creado_en": "timestamp",
      "actualizado_en": "timestamp"
    }
  ],
  "total": 1
}
```

### Creación

Input:

```json
{
  "codigo": "SUC-001",
  "nombre": "Matriz Reforma",
  "telefono": "+52 55 1234 5678",
  "es_principal": true,
  "direccion": {
    "codigo_pais": "MX",
    "estado": "Ciudad de México",
    "municipio": "Miguel Hidalgo",
    "ciudad": "CDMX",
    "colonia": "Polanco",
    "codigo_postal": "11560",
    "calle": "Av. Paseo de la Reforma",
    "numero_exterior": "123",
    "numero_interior": "Piso 4",
    "referencias": "Frente a la fuente principal"
  }
}
```

Respuesta `201`: detalle completo.

### Actualización

Input parcial; al menos un campo:

```json
{
  "nombre": "Matriz Centro",
  "telefono": null,
  "es_principal": true,
  "direccion": null
}
```

`codigo` nunca se acepta. Campo omitido conserva valor; `telefono: null` lo limpia; `direccion: null` desvincula y elimina su fila huérfana dentro de transacción.

Respuesta `200`: detalle actualizado.

### Detalle

Respuesta incluye todos los campos de sucursal, dirección completa, `estado` derivado y `eliminado_en`.

### Estados HTTP

- `400`: UUID o JSON inválido.
- `401`: sesión ausente o inválida.
- `403`: miembro sin capacidad de mutación.
- `404`: negocio inaccesible o sucursal fuera del negocio.
- `409`: código duplicado, estado repetido o violación de principal.
- `422`: validación de campos.
- `500`: error interno sin detalles sensibles.

## Organización backend

Fase amplía dominio `negocio`:

```text
backend/internal/domain/negocio/sucursal.go
backend/internal/application/negocio/sucursal_service.go
backend/internal/application/negocio/sucursal_service_test.go
backend/internal/infrastructure/negocio/sucursal_repository.go
backend/internal/infrastructure/negocio/sucursal_migration.go
backend/internal/interfaces/http/negocio/sucursal_handler.go
backend/internal/interfaces/http/negocio/sucursal_handler_test.go
```

Reglas:

- Service posee validación, autorización e invariantes de principal.
- Repository posee queries y transacciones.
- Handler solo traduce HTTP.
- Migración de sucursal se ejecuta después de fase 1 y antes de AutoMigrate legacy.
- Compatibilidad con modelo raíz es temporal y no autoriza agregar comportamiento de sucursal a producto.
- No modificar `product_handler.go`, `product_repository.go`, modelos, servicios, componentes ni pruebas de producto.

## Frontend

### Rutas

```text
/negocios/:negocioId/sucursales
/negocios/:negocioId/sucursales/nueva
/negocios/:negocioId/sucursales/:sucursalId
/negocios/:negocioId/sucursales/:sucursalId/editar
```

`/sucursales` redirige a `/negocios` mientras no exista contexto global.

### Listado

- Solo cards responsive; no tabla.
- Tabs `Activas` y `Archivadas`.
- Búsqueda por nombre o código.
- Card muestra código, nombre, ubicación resumida, teléfono, estado y distintivo `Principal` o `Secundaria`.
- Propietario ve editar, archivar o restaurar según estado.
- Miembro solo ve detalle.
- Estado vacío activo invita a crear primera sucursal.
- No muestra empleados asignados, límite de plan ni controles de suscripción.

### Formulario

- Ruta compartida para alta y edición.
- Código requerido y editable solo durante creación.
- Nombre requerido; teléfono y dirección opcionales.
- Checkbox `Sucursal principal`; primera sucursal informa que será principal automáticamente.
- No incluye toggle activa: alta siempre activa y archivado usa acción separada.
- Vista previa lateral en escritorio y resumen compacto debajo del formulario en móvil.
- Errores de campos y carga permanecen inline.

### Detalle

- Identificación, contacto, dirección, estado, principal/secundaria y timestamps.
- Edición disponible solo para propietario y negocio activo.
- Zona de riesgo con archivar o restaurar.
- Enlace de regreso conserva filtro mediante navegación normal; no introduce contexto persistente.

### Integración con negocio y feedback

- Detalle de negocio siempre muestra `Administrar sucursales`.
- CTA de primera sucursal deja de estar deshabilitado y abre alta.
- Éxitos usan Sonner: registrada, actualizada, principal cambiada, archivada y restaurada.
- Fallos de acciones usan Sonner; fallos de carga o validación permanecen inline.
- Archivado usa `FeedbackService.confirmDanger` con nombre de sucursal.
- Restauración no solicita confirmación.
- No importar directamente Sonner o SweetAlert2 en componentes.

## Seguridad y aislamiento

- Grupo administrativo protegido con `RequireAuth`.
- Service recibe siempre `usuarioID`, `negocioID` y, cuando corresponda, `sucursalID`.
- Toda consulta filtra simultáneamente negocio y sucursal.
- Lectura exige membresía activa aun si negocio o sucursal están archivados.
- Mutación exige propiedad activa y negocio activo.
- No distinguir mediante mensajes si UUID pertenece a otro negocio.
- Transacciones bloquean filas involucradas al cambiar principal para evitar dos principales concurrentes.
- Constraint parcial de DB actúa como última defensa de concurrencia.

## Pruebas de fase 2

### Backend

- Migración desde esquema legacy conserva registros y texto de dirección.
- Backfill produce códigos deterministas sin colisiones.
- Segunda ejecución no cambia datos ni crea direcciones duplicadas.
- Primera sucursal se vuelve principal.
- Crear/promover principal degrada anterior atómicamente.
- Dos promociones concurrentes no dejan múltiples principales.
- Código se normaliza, es único por negocio e inmutable.
- Mismo código puede existir en negocios distintos.
- Propietario muta; miembro activo solo consulta.
- Usuario ajeno y sucursal de otro negocio reciben `404`.
- Negocio archivado bloquea mutaciones.
- Principal con alternativas no puede archivarse.
- Última activa sí puede archivarse.
- Restauración elige principal solamente si no existe otra.
- Dirección crear, reemplazar y eliminar no deja filas huérfanas.
- Endpoint legacy de producto conserva ruta, respuesta y filtro activo.
- Códigos HTTP y cuerpos de error coinciden con contrato.

### Frontend

- Rutas cargan lista, alta, detalle y edición correctos.
- `/sucursales` redirige a negocios.
- Búsqueda y tabs envían filtros correctos.
- Cards muestran estado y principal/secundaria.
- Miembro no ve mutaciones.
- Código se bloquea en edición.
- Primera sucursal informa promoción automática.
- Cancelar SweetAlert no ejecuta HTTP.
- Confirmar archivado ejecuta una petición y un toast.
- Restaurar no solicita confirmación y muestra toast.
- Formularios preservan errores inline y no duplican toast.
- Revisión visual 390, 768 y 1440 px en claro/oscuro.

### Verificación integral

- `go list ./...`, `go test ./...` y `go vet ./...`.
- `npm test -- --watch=false` y `npm run build`.
- Migración dos veces sobre clon legacy y comparación de datos.
- Smoke test autenticado: listar, crear, obtener, editar, promover, archivar y restaurar.
- Smoke del GET legacy usado por producto.
- Diff confirma cero cambios en producto, catálogo, proveedores e inventario.

## Criterios de aceptación de fase 2

- [x] `base.MD` confirmado como fuente inicial y preservado sin cambios.
- [x] Navegación por negocio definida.
- [x] Listado solo con cards definido.
- [x] Código inmutable definido.
- [x] Regla de principal único definida.
- [x] Backfill legacy definido.
- [x] Archivado de última activa permitido.
- [x] Compatibilidad API de producto separada.
- [x] Datos futuros de empleados y suscripción excluidos.
- [x] Especificación completa revisada por el usuario.
- [x] Fase aprobada explícitamente para implementación.
- [x] Migración implementada e idempotente.
- [x] Backend, frontend y seguridad implementados.
- [x] Contrato legacy de producto preservado.
- [x] Pruebas y verificaciones pasan.
- [x] Usuario acepta resultado final.

## Verificación técnica de fase 2

Fecha: 2026-09-11.

- `go list ./...`: correcto, sin ciclos de imports.
- `go test ./...`: correcto; incluye service, contratos HTTP y prueba de integración PostgreSQL de sucursales.
- `go vet ./...`: correcto.
- Migración ejecutada dos veces sobre esquema legacy temporal: códigos `SUC-001...` estables, direcciones legacy preservadas, direcciones no duplicadas y principal única.
- Prueba transaccional en PostgreSQL: promoción degrada principal anterior, principal con alternativas no se archiva, restauración respeta principal existente y código solo colisiona dentro del mismo negocio.
- Smoke HTTP autenticado sobre DB temporal: registro, login, creación de negocio, listado, alta de dos sucursales, detalle, edición, promoción, rechazo de cambio de código, archivo, restauración, búsqueda y estados `200/201/204/400/409` esperados.
- Endpoint legacy `GET /api/v1/negocios/:negocioId/sucursales`: respondió correctamente y excluyó sucursal archivada.
- `npm test -- --watch=false`: correcto, 24 archivos y 54 pruebas.
- `npm run build`: correcto; conserva avisos no bloqueantes de presupuesto CSS, incluidos los nuevos estilos de sucursales.
- Revisión visual en 1440 px claro, 768 px oscuro y 390 px claro: rutas correctas, sin scroll horizontal y sin excepciones de ejecución.
- Diff revisado: ningún archivo de producto, catálogo, proveedores o inventario fue modificado.
- `capabilities/_template/base.MD` fue preservado sin cambios.
- Las dos bases PostgreSQL temporales, sus usuarios y datos de smoke fueron eliminados; la DB compartida no fue modificada.

## Imprevistos resueltos de fase 2

- El seed legacy habría intentado insertar una sucursal sin el nuevo código obligatorio. Se mantuvieron sus IDs y datos de desarrollo, agregando `SUC-001` y la marca principal sin tocar código de producto.
- Para evitar que `codigo` u otros campos internos fueran ignorados silenciosamente, el handler administrativo usa decodificación JSON estricta y rechaza campos desconocidos con `400`.
- Una revisión inicial reutilizó un servidor Angular/HMR anterior y mostró un error circular. Se repitió en servidor y Chromium aislados: las tres rutas cargaron sin errores de runtime; no era causado por fase 2.

## Cierre de fase 2

- 2026-09-12: el usuario autorizó explícitamente el cierre de la fase junto con la ejecución no supervisada de fases 4 a 8. Estado `Verificada` cambiado a `Cerrada`.

## Decisiones resueltas de fase 2

- Especificación completa e implementación aprobadas explícitamente el 2026-09-11.

---
