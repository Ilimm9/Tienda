# Fase 1: Negocios

## Estado de fase

`Cerrada`

Autorizada explícitamente para implementación el 2026-09-10.
Implementación y verificación técnica terminadas el 2026-09-10; resultado aceptado por el usuario el 2026-09-11.

## Objetivo y alcance

Crear administración de negocios accesibles por usuario autenticado. Diferenciar propiedad y colaboración sin implementar todavía roles detallados.

Incluye:

- Listar negocios donde usuario tiene membresía activa.
- Diferenciar `propietario` y `miembro`.
- Crear negocio y membresía propietaria atómicamente.
- Consultar detalle.
- Editar datos generales.
- Archivar y restaurar negocio.
- Crear configuración operativa predeterminada.
- Registrar dirección fiscal/comercial opcional.
- Mostrar CTA para crear primera sucursal, sin implementar sucursales en esta fase.

No incluye:

- CRUD de sucursales.
- Selector global de negocio o sucursal.
- Roles, permisos y autorización granular.
- Empleados e invitaciones.
- Configuración de tickets, suscripciones, planes o secuencias documentales.
- Eliminación física.

## Flujo funcional

### Usuario sin negocios

1. Usuario autenticado entra en `/negocios`.
2. API devuelve lista vacía.
3. Vista explica que necesita crear un negocio.
4. Acción principal abre `/negocios/nuevo`.
5. Usuario completa datos obligatorios y opcionales.
6. Backend crea dirección opcional, negocio, configuración y membresía propietaria en una transacción.
7. Frontend muestra detalle y CTA para crear sucursal principal en fase 2.

### Usuario con negocios

1. Vista muestra cards de negocios con tipo de relación y estado.
2. Usuario abre detalle de un negocio.
3. Propietario puede editar, archivar o restaurar.
4. Miembro puede consultar, pero no modificar durante esta fase.
5. Negocios archivados se muestran únicamente al activar filtro correspondiente.

## Modelo de datos de fase 1

### `direcciones`

```text
id                  uuid primary key
codigo_pais         varchar(2) not null default 'MX'
estado              varchar(120) null
municipio           varchar(120) null
ciudad              varchar(120) null
colonia             varchar(150) null
codigo_postal       varchar(12) null
calle               varchar(180) null
numero_exterior     varchar(30) null
numero_interior     varchar(30) null
referencias         text null
creado_en           timestamptz not null
actualizado_en      timestamptz not null
```

Una dirección vacía no debe crearse. Dirección compartida entre entidades significa reutilizar estructura, no compartir misma fila.

### `negocios`

```text
id                       uuid primary key
slug                     varchar(120) not null unique
nombre_comercial         varchar(180) not null
razon_social             varchar(220) null
rfc                      varchar(20) null
telefono                 varchar(30) null
correo                   varchar(254) null
direccion_id             uuid null references direcciones(id)
archivo_logo_id          uuid null, reservado sin relación activa en esta fase
codigo_moneda            varchar(3) not null default 'MXN'
zona_horaria             varchar(80) not null default 'America/Mexico_City'
estado                   varchar(30) not null default 'activo'
creado_por_usuario_id    uuid null temporalmente references usuarios(id)
creado_en                timestamptz not null
actualizado_en           timestamptz not null
archivado_en             timestamptz null
```

Índices:

- Único case-insensitive para `slug` normalizado.
- Índice no único para `rfc` normalizado cuando no sea nulo.
- Índice para `creado_por_usuario_id`.

### `configuraciones_negocio`

Se crea una fila 1:1 con valores predeterminados. En fase 1 no tendrá pantalla ni endpoints propios.

```text
id                                                   uuid primary key
negocio_id                                           uuid not null unique references negocios(id)
metodo_costeo                                        varchar(40) not null default 'promedio_ponderado'
permitir_inventario_negativo                         boolean not null default false
requerir_autorizacion_cambio_precio                  boolean not null default true
requerir_autorizacion_eliminar_partida_venta         boolean not null default false
permitir_multiples_cajas_abiertas_por_empleado       boolean not null default false
permitir_cambio_caja_con_autorizacion                boolean not null default true
habilitar_cortes_parciales_caja                      boolean not null default false
monto_corte_parcial_caja                             numeric(18,2) null
hora_corte_parcial_caja                              time null
precios_mayoreo_habilitados                          boolean not null default true
dias_devolucion_envase_predeterminados                integer null
creado_en                                            timestamptz not null
actualizado_en                                       timestamptz not null
```

### `membresias_negocio`

```text
id                 uuid primary key
negocio_id         uuid not null references negocios(id)
usuario_id         uuid not null references usuarios(id)
tipo_miembro       varchar(30) not null default 'miembro'
estado             varchar(30) not null default 'activo'
se_unio_en         timestamptz null
suspendido_en      timestamptz null
revocado_en        timestamptz null
creado_en          timestamptz not null
actualizado_en     timestamptz not null
```

Restricciones:

- Único `(negocio_id, usuario_id)`.
- `tipo_miembro`: `propietario` o `miembro` durante esta fase.
- `estado`: `activo`, `suspendido` o `revocado`.
- Campo legado `rol_id` quedará nullable hasta migración RBAC de fase 4.

## Migración segura

La migración de fase 1 debe ser automática e idempotente:

1. Crear `direcciones` y `configuraciones_negocio` si no existen.
2. Agregar columnas nuevas nullable a `negocios`.
3. Copiar `negocios.nombre` a `nombre_comercial` donde falte.
4. Generar slug único y estable desde nombre; resolver colisiones con sufijo corto.
5. Completar `codigo_moneda`, `zona_horaria`, `estado` y timestamps.
6. Conservar `creado_por_usuario_id` nullable durante transición. Todo negocio nuevo debe recibir usuario creador desde sesión autenticada; únicamente seed legado de desarrollo puede permanecer en `NULL`.
7. Agregar campos nuevos a membresías y traducir registros existentes a `tipo_miembro='miembro'`, `estado='activo'`.
8. Volver nullable `rol_id`; conservar columna hasta fase 4.
9. Crear índices y restricciones después del backfill.
10. Mantener temporalmente columna `negocios.nombre` para compatibilidad; escritura nueva debe sincronizarla mientras exista código legado.

No usar únicamente `AutoMigrate` para renombrar columnas, cambiar nulabilidad o ejecutar backfill. Añadir migración explícita y conservar `database.Init` como ejecutor automático al iniciar.

### Decisión resuelta: seed de desarrollo

- Conservar temporalmente negocio y sucursal sembrados porque siguen en uso por otro desarrollador.
- Permitir `creado_por_usuario_id = NULL` únicamente para negocio legado creado por seed.
- No inventar propietario ni crear membresía para datos legados.
- Aplicación debe exigir creador y membresía propietaria para todo negocio nuevo.
- Seed continúa condicionado a `APP_ENV=development`.
- Antes de producción, retirar datos sembrados, comprobar que no existan negocios sin creador y aplicar `NOT NULL` a `creado_por_usuario_id` mediante migración posterior.

## Reglas de negocio

- Usuario debe estar autenticado para cualquier endpoint.
- Crear negocio genera membresía `propietario` para usuario autenticado.
- Operación de creación completa es transaccional.
- Slug se genera en servidor, se normaliza a minúsculas y no cambia automáticamente al editar nombre.
- Nombre comercial es obligatorio después de recortar espacios.
- RFC es opcional; si existe, se guarda en mayúsculas y sin espacios laterales.
- Correo es opcional; si existe, se normaliza a minúsculas.
- Moneda queda `MXN` y zona horaria `America/Mexico_City` por defecto.
- Solo propietario activo puede editar, archivar o restaurar durante fase 1.
- Miembro activo puede listar y consultar detalle.
- Membresía suspendida o revocada no concede acceso.
- Negocio archivado no aparece en listado activo ni admite operación normal.
- Archivar negocio no elimina membresías, productos, inventario ni futuros datos relacionados.
- Último propietario no puede revocarse; gestión de propietarios se desarrollará con RBAC.
- No exponer negocios mediante identificador si usuario no tiene membresía activa.

## API y contratos

Base: `/api/v1`

### Listar negocios accesibles

`GET /negocios?estado=activo|archivado`

Respuesta `200`:

```json
{
  "items": [
    {
      "id": "uuid",
      "slug": "mi-tienda",
      "nombre_comercial": "Mi Tienda",
      "tipo_miembro": "propietario",
      "estado": "activo",
      "tiene_sucursales": false
    }
  ],
  "total": 1
}
```

### Crear negocio

`POST /negocios`

Entrada:

```json
{
  "nombre_comercial": "Mi Tienda",
  "razon_social": "Mi Tienda SA de CV",
    "rfc": "RFC123456ABC",
    "telefono": "5555555555",
    "correo": "contacto@mitienda.mx",
    "codigo_moneda": "MXN",
    "zona_horaria": "America/Mexico_City",
  "direccion": {
    "codigo_pais": "MX",
    "estado": "Ciudad de México",
    "municipio": "Cuauhtémoc",
    "codigo_postal": "06000",
    "calle": "Reforma",
    "numero_exterior": "100"
  }
}
```

Respuesta `201`: detalle completo del negocio creado. No aceptar `slug`, `estado`, `creado_por_usuario_id` ni tipo de membresía desde cliente.

### Consultar detalle

`GET /negocios/:negocioId`

Respuesta `200`: datos generales, dirección, relación del usuario y `tiene_sucursales`.

### Actualizar negocio

`PATCH /negocios/:negocioId`

Permite actualizar datos generales y dirección. Campos omitidos se conservan. `null` explícito elimina campo opcional o dirección cuando contrato lo permita. No cambia slug automáticamente.

### Archivar negocio

`DELETE /negocios/:negocioId`

Respuesta `204`. Actualiza `estado='archivado'` y `archivado_en`; no elimina filas.

### Restaurar negocio

`POST /negocios/:negocioId/restaurar`

Respuesta `200` con detalle restaurado.

### Errores uniformes

```json
{
  "codigo": "NEGOCIO_NO_ENCONTRADO",
  "mensaje": "No fue posible encontrar el negocio solicitado.",
  "campos": {}
}
```

Códigos HTTP requeridos:

- `400`: entrada o UUID inválido.
- `401`: sesión ausente o inválida.
- `403`: membresía sin capacidad requerida.
- `404`: negocio inexistente o no visible para usuario.
- `409`: slug/RFC incompatible o transición de estado inválida.
- `422`: campos válidos en formato, pero regla de negocio incumplida.
- `500`: fallo inesperado sin filtrar detalle interno.

## Backend

Componentes planeados:

- `domain/negocio.go`: entidades, inputs y vistas de negocio, dirección, configuración y membresía mínima.
- `application/negocio_service.go`: casos de uso, errores e interfaz `NegocioRepository`.
- `infrastructure/negocio_repository.go`: consultas GORM y transacciones.
- `interfaces/http/negocio_handler.go`: binding, sesión y respuestas HTTP.
- Middleware reutilizable para extraer usuario autenticado desde cookie JWT y colocarlo en contexto.
- Registro de rutas y dependencias en `cmd/api/main.go`.

El service debe decidir quién puede operar y coordinar creación. Repository no decide códigos HTTP ni permisos.

## Frontend y experiencia

Rutas:

```text
/negocios
/negocios/nuevo
/negocios/:negocioId
/negocios/:negocioId/editar
```

Vistas:

- Listado responsive con cards y card primaria “Crear negocio”.
- Filtros `Activos` y `Archivados`.
- Etiqueta visible `Propietario` o `Colaborador`.
- Formulario de alta/edición en página completa.
- Detalle con datos fiscales, contacto, dirección y estado.
- Confirmación explícita antes de archivar.
- Acción de restaurar visible solo para propietario en listado archivado.
- Estado vacío con explicación y CTA.
- Después de crear, detalle muestra CTA “Crear sucursal principal”; durante fase 1 indica que configuración de sucursales aún no está disponible.

Campos visibles del formulario:

- `Nombre comercial *`: obligatorio.
- `Moneda base *`: obligatoria, con `MXN` preseleccionado.
- `Zona horaria *`: obligatoria, con `America/Mexico_City` preseleccionada.
- `Razón social`, `RFC`, `Teléfono` y `Correo electrónico`: opcionales.
- Dirección completa: opcional como bloque. Si usuario captura cualquier dato de dirección, `País *` queda obligatorio y usa `MX` por defecto.
- `Estado/Entidad`, `Municipio/Delegación`, `Ciudad`, `Colonia`, `Código postal`, `Calle`, `Número exterior`, `Número interior` y `Referencias`: opcionales.

Campos no editables o no visibles:

- `slug`: generado por servidor y mostrado como lectura informativa.
- `estado`, `creado_por_usuario_id`, membresía propietaria y timestamps: internos.
- Configuración operativa: creada con valores predeterminados, sin formulario en fase 1.
- Logo: fuera de alcance hasta definir almacenamiento de archivos; usar placeholder visual.

Referencias visuales:

- `capabilities/img/visily-mis-negocios.png`
- `capabilities/img/visily-registrar-nuevo-negocio.png`
- `capabilities/img/visily-datos-del-negocio.png`
- `capabilities/img/visily-seleccionar-negocio.png`

Requisitos:

- Mantener shell, breadcrumbs, variables CSS y estilo visual existente.
- Funcionar en escritorio, tablet y móvil sin scroll horizontal.
- Controles táctiles de al menos 44 px.
- Foco visible, labels asociados y mensajes accesibles mediante `role="alert"` o `role="status"`.
- No introducir framework CSS nuevo.

## Seguridad y aislamiento

- No confiar en `negocioId` enviado por frontend sin comprobar membresía.
- Toda consulta de detalle o modificación debe filtrar negocio mediante usuario/membresía.
- Para evitar enumeración, negocio ajeno responde `404`; usuario autenticado sin permiso dentro de negocio visible responde `403`.
- JWT debe validar algoritmo esperado, firma y expiración antes de extraer `sub`.
- `creado_por_usuario_id` siempre proviene del token validado.
- No devolver hashes, tokens ni detalles SQL.
- Dirección se modifica solo junto al negocio autorizado.
- Operaciones de crear y actualizar deben usar contexto de request en GORM cuando se implemente.

## Validaciones

- `nombre_comercial`: requerido, 2 a 180 caracteres después de normalización.
- `codigo_moneda`: requerido, tres letras mayúsculas; `MXN` por defecto.
- `zona_horaria`: requerida, identificador IANA válido; `America/Mexico_City` por defecto.
- `razon_social`: máximo 220 caracteres.
- `rfc`: opcional, mayúsculas, 12 o 13 caracteres y formato fiscal mexicano básico.
- `correo`: opcional, formato email, máximo 254 caracteres.
- `telefono`: opcional, máximo 30 caracteres; conservar prefijo internacional.
- `codigo_pais`: dos letras ISO, `MX` por defecto.
- `codigo_postal`: opcional, máximo 12 caracteres.
- Campos de dirección respetan longitudes del esquema.
- String vacío en campo opcional se normaliza a `null`.
- PATCH sin campos modificables devuelve `400`.
- No aceptar cambios directos a identificadores, creador, timestamps o estado mediante PATCH general.

## Pruebas de fase 1

### Service

- Crear negocio normaliza datos y crea propietario.
- Fallo en cualquier inserción revierte transacción.
- Usuario miembro puede consultar, pero no editar.
- Usuario ajeno no puede descubrir negocio.
- Propietario puede archivar y restaurar.
- Negocio archivado queda fuera del listado activo.
- Slug duplicado recibe resolución estable sin sobrescribir otro negocio.

### Repository y migración

- Migración funciona sobre esquema actual con datos existentes.
- Segunda ejecución no cambia resultado ni falla.
- `nombre` existente se copia a `nombre_comercial`.
- Índices y restricciones se crean después de backfill.
- Creación atómica genera negocio, configuración y membresía.
- Consultas siempre aíslan por usuario y negocio.

### HTTP

- JSON y UUID inválidos devuelven `400`.
- Sesión inválida devuelve `401`.
- Acceso prohibido y recurso invisible usan `403`/`404` según regla.
- Respuestas coinciden con contratos y no filtran campos internos.

### Frontend

- Estados vacío, carga, error, activo y archivado.
- Navegación listado, alta, detalle y edición.
- Validaciones visibles y accesibles.
- Acciones cambian según propiedad.
- Cards y formularios funcionan a 390 px, tablet y escritorio.

### Regresión

- Registro, login, logout y `/auth/me` siguen funcionando.
- Productos continúan consultando negocio existente durante transición.
- Build y pruebas completas de backend/frontend pasan.

## Criterios de aceptación de fase 1

- [x] Se resolvió decisión de transición para propietario de datos existentes.
- [x] Contratos API fueron revisados y aprobados.
- [x] Modelo y migración fueron revisados y aprobados.
- [x] UI y navegación fueron revisadas y aprobadas.
- [x] Fase recibió aprobación explícita para implementación.
- [x] Migración automática conserva datos existentes y es idempotente.
- [x] CRUD con archivado funciona según permisos.
- [x] Aislamiento multiempresa está probado.
- [x] Responsive y accesibilidad están verificadas.
- [x] Pruebas backend/frontend pasan.
- [x] Usuario acepta resultado final.

## Verificación técnica de fase 1

Fecha: 2026-09-10.

- `go test ./...`: correcto.
- `npm test -- --watch=false`: 16 archivos y 29 pruebas correctas.
- `npm run build`: correcto; permanecen avisos no bloqueantes de presupuesto CSS.
- Migración PostgreSQL sobre clon del esquema legacy: correcta en primera y segunda ejecución.
- Backfill comprobado: `nombre` pasó a `nombre_comercial`, slug legacy fue generado y creador legacy permaneció `null`.
- Transacción comprobada: negocio nuevo creó una configuración y una membresía `propietario`.
- API comprobada: crear, consultar, actualizar parcialmente, archivar, filtrar archivados y restaurar.
- Preflight CORS comprobado para `PATCH`, incluyendo origen permitido, credenciales, headers y métodos.
- Aislamiento comprobado: segundo usuario recibió `404` al consultar negocio ajeno.
- UI comprobada en 1440, 768 y 390 px, sin scroll horizontal.
- Base y snapshot temporales de validación fueron eliminados; base compartida no fue modificada.

## Decisiones resueltas de fase 1

- Seed legado permanece temporalmente con creador nullable y sin membresía.
- Campos obligatorios visibles se marcan con `*`; opcionales pueden omitirse.
- Campos internos no aparecen como controles editables.
- Restauración vive en mismo listado `/negocios`, usando filtro `Archivados`; cada card autorizada muestra acción `Restaurar`.

## Observaciones pendientes de fase 1

- Falta aceptación visual y funcional final del usuario antes de cambiar fase a `Cerrada`.

---
