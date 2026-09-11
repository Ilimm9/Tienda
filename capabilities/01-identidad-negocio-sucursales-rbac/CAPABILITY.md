# Capability 01: Identidad, negocios, sucursales y control de acceso

## Estado

`En implementación`

La estructura por fases fue aprobada el 2026-09-10. Fases 1, 1.5 y 1.6 fueron aprobadas explícitamente; fases posteriores requieren aprobación independiente.

## Control

- Responsable: Usuario del proyecto
- Fecha de creación: 2026-09-10
- Última revisión estructural: 2026-09-10
- Aprobación de estructura: Confirmada
- Aprobación de implementación: Fases 1, 1.5 y 1.6 aprobadas explícitamente el 2026-09-10
- Dependencias generales: autenticación existente con JWT en cookie

## Objetivo

Permitir que una cuenta global administre varios negocios y participe en otros, opere sucursales separadas y delegue acceso mediante empleados, invitaciones, roles y permisos. Todo dato debe quedar aislado por negocio y, cuando aplique, por sucursal.

## Usuario y problema

Una misma persona puede:

- Ser propietaria de uno o varios negocios.
- Colaborar en otros negocios con capacidades distintas.
- Administrar ubicaciones operativas independientes.
- Registrar empleados antes de que tengan cuenta.
- Invitar empleados y asignarles acceso controlado.

La cuenta global no debe guardar un negocio o rol fijo. La relación con cada negocio se representa mediante membresías.

## Decisiones generales aprobadas

- Trabajar una fase por vez. No detallar ni implementar la siguiente hasta cerrar la actual.
- Conservar autenticación JWT actual durante esta capability.
- Usar migraciones automáticas, seguras, aditivas y con backfill; no perder datos existentes.
- Usar rutas dedicadas para listados, altas, detalles y edición; no usar modales para formularios principales.
- Usar archivado lógico para negocios y sucursales.
- Mantener siempre al menos una membresía propietaria activa por negocio.
- Guardar último negocio y sucursal seleccionados en `localStorage`; la API siempre validará acceso vigente.
- Usar permisos globales, roles por negocio y varios roles por membresía.
- Registrar empleado antes de invitarlo.
- Entregar invitación inicial mediante enlace copiable; integración de correo queda fuera.
- Si invitado no tiene cuenta, completar registro y continuar aceptación con el mismo correo.

## Reglas de ejecución por fases

Cada fase debe completar, en orden:

1. Objetivo y alcance.
2. Modelo de datos y migración.
3. Reglas de negocio.
4. API y contratos.
5. Pantallas y navegación.
6. Seguridad y aislamiento.
7. Validaciones y errores.
8. Pruebas.
9. Criterios de aceptación.
10. Aprobación explícita.
11. Implementación y verificación.
12. Aceptación final.

No avanzar de sección si existe una decisión pendiente que cambie contratos, datos o experiencia. Si una fase requiere modificar un área fuera de su alcance, detener esa parte y documentar impacto antes de continuar.

## Convención de archivos

Backend usa identificador singular en minúsculas:

```text
negocio.go
negocio_handler.go
negocio_service.go
negocio_repository.go
negocio_service_test.go
negocio_repository_test.go

sucursal.go
sucursal_handler.go
sucursal_service.go
sucursal_repository.go
```

Reglas:

- Handler entiende HTTP, valida formato de entrada y traduce errores a estados HTTP.
- Service contiene autorización, reglas de negocio y coordinación transaccional.
- Repository contiene consultas y persistencia GORM.
- Interfaces de repositories se declaran en `application`, junto al service consumidor.
- Modelos e inputs compartidos viven en `domain`.
- No ampliar handlers o repositories de productos para alojar negocios, sucursales o RBAC.
- Pruebas conservan nombre del archivo probado y sufijo `_test.go`.

## Organización backend por dominios

Cada dominio funcional debe repetirse dentro de las capas donde tenga responsabilidades:

```text
backend/internal/
├── domain/<dominio>/
├── application/<dominio>/
├── infrastructure/<dominio>/
└── interfaces/http/<dominio>/
```

Reglas obligatorias:

- Carpetas de dominio usan nombre singular, descriptivo y en minúsculas.
- Archivos permanecen pequeños y agrupados por entidad o caso de uso relacionado.
- `domain/<dominio>` contiene entidades, tipos de valor, inputs y vistas propias del dominio.
- `application/<dominio>` contiene services, casos de uso, errores e interfaces requeridas por esos services.
- `infrastructure/<dominio>` contiene repositories, migraciones e integraciones que implementan contratos de aplicación.
- `interfaces/http/<dominio>` contiene handlers y middleware exclusivo del dominio.
- Código transversal con responsabilidad real puede permanecer en raíz; ejemplos actuales: configuración, conexión DB, CORS y composición en `cmd/api`.
- No crear carpetas comodín `shared`, `common`, `helpers` o `utils` para evitar decidir pertenencia.
- Un dominio no importa handlers, repositories ni implementaciones de otro dominio.
- Dependencias válidas siguen dirección `HTTP -> application -> domain`; infraestructura implementa puertos de application y usa domain.
- Integraciones entre dominios usan UUID, DTO o interfaces explícitas. Modelos de negocio guardan `usuario_id`, pero no necesitan importar entidad completa `Usuario`.
- Todo dominio introducido por fases posteriores debe declarar ubicación de sus archivos antes de implementación.
- Excepciones temporales de compatibilidad deben documentarse, no agregar comportamiento y tener condición clara de eliminación.

## Mapa de fases

| Fase | Tema                           | Estado      | Dependencia                     |
| ---- | ------------------------------ | ----------- | ------------------------------- |
| 1    | Negocios                       | Verificada  | Autenticación actual            |
| 1.5  | Organización por dominios      | Cerrada     | Negocios verificada             |
| 1.6  | Feedback global del frontend   | Verificada  | Fase 1.5 cerrada                |
| 2    | Sucursales                     | Pendiente   | Fases 1, 1.5 y 1.6 cerradas     |
| 3    | Contexto activo                | Pendiente   | Sucursales cerrada              |
| 4    | RBAC                           | Pendiente   | Contexto activo cerrado         |
| 5    | Empleados                      | Pendiente   | RBAC cerrado                    |
| 6    | Invitaciones                   | Pendiente   | Empleados y RBAC cerrados       |
| 7    | Asignaciones empleado-sucursal | Pendiente   | Empleados y sucursales cerrados |
| 8    | Integración y endurecimiento   | Pendiente   | Fases 1, 1.5, 1.6 y 2 a 7 cerradas |

---

# Fase 1: Negocios

## Estado de fase

`Verificada`

Autorizada explícitamente para implementación el 2026-09-10.
Implementación y verificación técnica terminadas el 2026-09-10; falta aceptación final del usuario.

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
- [ ] Usuario acepta resultado final.

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

# Fase 1.5: Organización backend por dominios

## Estado de fase

`Cerrada`

Autorizada explícitamente para implementación el 2026-09-10.
Implementación y verificación técnica terminadas el 2026-09-10; resultado aceptado por el usuario el 2026-09-10.

## Objetivo y alcance

Reorganizar autenticación e identidad de negocio en paquetes cohesivos antes de agregar sucursales, RBAC, empleados e invitaciones. Cambio será estructural: no modificará reglas, esquema DB, rutas HTTP ni contratos JSON.

Incluye:

- Crear dominios backend `cuenta` y `negocio` con mismo nombre dentro de cada capa aplicable.
- Mover código de autenticación, usuario y perfil al dominio `cuenta`.
- Mover negocio, dirección, configuración, membresía, rol y empleado al dominio `negocio`.
- Mover tests junto con paquete probado.
- Actualizar imports, aliases de paquetes y composición en `cmd/api/main.go`.
- Colocar migración de negocio dentro de infraestructura de `negocio`; inicialización DB seguirá coordinándola.
- Documentar reglas usadas por fases posteriores.
- Mantener puentes temporales mínimos solo cuando producto los necesite para compilar sin modificar sus archivos.

No incluye:

- Cambios funcionales de autenticación o negocios.
- Cambios de tablas, columnas, índices, constraints, seeds o datos.
- CRUD de sucursales, roles, empleados o invitaciones.
- Reorganización de frontend.
- Reorganización, edición o formateo de archivos de producto, catálogo, proveedores o inventario.
- Creación del futuro dominio `catalogo`; corresponde al trabajo del colaborador responsable.

## Dominios aprobados para esta capability

### `cuenta`

Responsable de identidad global y autenticación independiente de cualquier negocio.

```text
backend/internal/domain/cuenta/
├── usuario.go
└── perfil_usuario.go

backend/internal/application/cuenta/
├── auth_service.go
└── auth_service_test.go

backend/internal/infrastructure/cuenta/
└── usuario_repository.go

backend/internal/interfaces/http/cuenta/
└── auth_handler.go
```

Propiedad conceptual:

- `Usuario`
- `PerfilUsuario`
- Registro, login, logout y sesión JWT.
- Persistencia de cuenta y perfil global.

### `negocio`

Responsable de tenant, estructura organizacional y acceso dentro del negocio.

```text
backend/internal/domain/negocio/
├── negocio.go
├── direccion.go
├── configuracion_negocio.go
├── membresia_negocio.go
├── empleado.go
└── rol.go

backend/internal/application/negocio/
├── negocio_service.go
└── negocio_service_test.go

backend/internal/infrastructure/negocio/
├── negocio_repository.go
└── negocio_migration.go

backend/internal/interfaces/http/negocio/
└── negocio_handler.go
```

Propiedad conceptual:

- `Negocio`
- `Direccion`
- `ConfiguracionNegocio`
- `MembresiaNegocio`
- `Empleado`, aunque pueda vincularse con `PerfilUsuario` después de una invitación.
- `Rol` y futuras entidades de RBAC.
- Sucursales, invitaciones y asignaciones cuando sus fases sean aprobadas.

## Clasificación de archivos transversales

Permanecen fuera de dominios:

- `internal/config`: carga de configuración general.
- `internal/database/database.go`: conexión, orden de migraciones y registro general de modelos.
- `internal/interfaces/http/cors_middleware.go`: política HTTP transversal.
- `internal/interfaces/http/auth_middleware.go`: validación JWT y contexto de usuario transversal para handlers protegidos.
- `cmd/api/main.go`: composition root y registro de rutas.

Estos archivos no pueden acumular reglas de cuenta o negocio.

## Compatibilidad temporal con producto

Producto está fuera de alcance porque otro colaborador trabaja en esa sección.

Lineamientos:

- No editar archivos con prefijos `product_`, clientes externos de producto ni modelos de catálogo durante fase 1.5.
- Si producto referencia tipos movidos como `Usuario` o `Negocio`, conservar aliases de compatibilidad en paquete `domain` raíz.
- Alias solo reexporta tipo; no duplica structs, métodos, hooks ni tablas.
- Archivo de compatibilidad debe indicar que será eliminado cuando dominio `catalogo` migre sus imports.
- `Sucursal` legacy permanece donde está durante fase 1.5 para no intervenir trabajo de producto. Fase 2 definirá su traslado coordinado a `domain/negocio` mediante alias temporal si sigue siendo necesario.
- Ningún cambio de fase 1.5 puede alterar consultas, DTO, endpoints o pruebas pertenecientes a producto.

## Reglas contra ciclos

Direcciones permitidas:

```text
interfaces/http/cuenta   -> application/cuenta   -> domain/cuenta
infrastructure/cuenta    -> application/cuenta + domain/cuenta

interfaces/http/negocio  -> application/negocio  -> domain/negocio
infrastructure/negocio   -> application/negocio + domain/negocio

cmd/api -> todos los paquetes necesarios para composición
database -> modelos de domain + migraciones de infrastructure
```

Prohibido:

- `domain` importar `application`, `infrastructure`, `interfaces/http` o `database`.
- `application` importar handlers, repositories concretos o `cmd/api`.
- `cuenta` importar implementaciones de `negocio`, o viceversa.
- Compartir modelos completos solo para obtener un identificador.
- Duplicar entidades para evitar resolver una dependencia.
- Crear aliases permanentes sin propietario y condición de retiro.

Cuando negocio necesite usuario, usará `uuid.UUID` como `UsuarioID`. Cuando coordinación requiera datos de cuenta, service consumidor declarará interfaz mínima y composición inyectará implementación.

## Estrategia de implementación

1. Crear paquetes `cuenta` y mover modelos de cuenta con tests existentes intactos.
2. Mover application, repository y HTTP de autenticación; actualizar wiring.
3. Crear paquete domain `negocio` y separar entidades actuales por archivo.
4. Agregar aliases temporales requeridos exclusivamente por producto sin tocar sus archivos.
5. Mover service, repository, handler, migración y tests de negocio.
6. Actualizar inicialización DB y composition root.
7. Comprobar imports para detectar ciclos y dependencias invertidas.
8. Ejecutar pruebas, build y verificación de migración idempotente.
9. Registrar resultado y aliases pendientes en esta sección.

Cada paso debe dejar backend compilable antes de continuar. No usar movimientos masivos que mezclen cambios funcionales.

## Contratos preservados

- Mismos nombres de tablas y columnas GORM.
- Mismos endpoints `/api/v1/auth/*` y `/api/v1/negocios/*`.
- Mismos métodos HTTP, cookies, estados y cuerpos JSON.
- Mismos valores de seed y comportamiento de migración.
- Mismas reglas de autorización y aislamiento.
- Mismos contratos consumidos por frontend.

## Organización esperada en fases posteriores

- Fase 2 agrega sucursal dentro de las cuatro carpetas `negocio`; no crea dominio nuevo.
- Fase 3 agrega contexto activo en `application/negocio` y adaptadores HTTP correspondientes; persistencia local del frontend permanece en feature de negocios.
- Fase 4 agrega permisos, roles y membresías en `domain/negocio`, con services, repositories y handlers equivalentes.
- Fase 5 agrega empleados en archivos propios dentro de `negocio`; no crea paquete paralelo `empleados`.
- Fase 6 agrega invitaciones dentro de `negocio`; interacción con cuenta ocurre mediante interfaz de aplicación, no importando repository de cuenta.
- Fase 7 agrega asignaciones empleado-sucursal dentro de `negocio`.
- Fase 8 endurece integración respetando límites; no vuelve a concentrar archivos en raíces.
- Capabilities futuras ajenas a identidad crean su propio dominio, repetido por capa: por ejemplo `catalogo`, `inventario`, `ventas` o `compras`.
- Una fase puede ampliar dominio existente solo si entidades comparten invariantes y ciclo de vida. Si no, debe proponer dominio nuevo en su plan.

## Pruebas de fase 1.5

- `go test ./...` pasa después de cada bloque de movimientos.
- `go list ./...` no reporta ciclos de imports.
- Pruebas de autenticación, JWT, CORS y negocio conservan resultados.
- `npm test -- --watch=false` y `npm run build` confirman contratos frontend intactos.
- Migración ejecutada dos veces sobre clon legacy conserva esquema y datos.
- Smoke test cubre registro, login, listado, creación, edición, archivado y restauración.
- Revisión de diff confirma cero cambios en archivos de producto y catálogo.
- Revisión de esquema confirma cero cambios causados únicamente por reorganización.

## Criterios de aceptación de fase 1.5

- [x] `Empleado` fue clasificado dentro de dominio `negocio`.
- [x] Producto y catálogo quedaron expresamente fuera de alcance.
- [x] Estructura de paquetes fue revisada y aprobada.
- [x] Estrategia temporal de aliases fue revisada y aprobada.
- [x] Fase recibió aprobación explícita para implementación.
- [x] Código de cuenta quedó organizado por dominio y capa.
- [x] Código de negocio quedó organizado por dominio y capa.
- [x] No existen ciclos ni dependencias de capas invertidas.
- [x] Contratos, DB y comportamiento permanecen iguales.
- [x] Archivos de producto y catálogo no fueron modificados.
- [x] Pruebas y builds pasan.
- [x] Usuario acepta resultado final.

## Verificación técnica de fase 1.5

Fecha: 2026-09-10.

- `go list ./...`: correcto, sin ciclos de imports.
- `go test ./...`: correcto, incluyendo pruebas nuevas de `application/cuenta`.
- `go vet ./...`: correcto.
- `npm test -- --watch=false`: 16 archivos y 29 pruebas correctas.
- `npm run build`: correcto; permanecen avisos no bloqueantes de presupuesto CSS ya registrados.
- Migración reorganizada ejecutada dos veces sobre clon de DB: correcta e idempotente.
- Dumps de esquema origen y temporal: idénticos; solo variaron tokens aleatorios `restrict/unrestrict` de `pg_dump`.
- Smoke test: registro `201`, login `200`, negocio `201`, listado `200`, actualización `200`, archivado `204` y restauración `200`.
- Endpoint de listado de productos respondió `200` en prueba de solo lectura.
- Diff revisado: ningún archivo de producto, catálogo, proveedor o inventario fue modificado por fase 1.5.
- Base y snapshot temporales fueron eliminados; DB compartida permaneció sin cambios.

## Decisiones resueltas de fase 1.5

- Estructura y aliases temporales aprobados explícitamente el 2026-09-10.
- Middleware JWT permanece en raíz HTTP porque entrega contexto autenticado a varios dominios; moverlo a `cuenta` crearía dependencia entre adapters HTTP.
- `Sucursal` permanece temporalmente en dominio legacy de catálogo; negocio la consume solo para conteo mediante puente hasta fase 2.

---

# Fase 1.6: Feedback global del frontend

## Estado de fase

`Verificada`

Propuesta elaborada y autorizada explícitamente para implementación el 2026-09-10. Implementación y verificación técnica terminadas el 2026-09-10; falta aceptación final del usuario.

## Objetivo y alcance

Unificar las notificaciones transitorias y las confirmaciones del frontend antes de construir sucursales. `ngx-sonner` mostrará notificaciones y `sweetalert2` solicitará confirmaciones. Ambas bibliotecas usarán configuración, estilos y textos centralizados.

Dependencias ya instaladas y comprobadas en `frontend/package.json`:

- `ngx-sonner ^3.1.0`.
- `sweetalert2 ^11.26.25`.

Incluye:

- Montar una sola instancia global de `NgxSonnerToaster` en la raíz de la aplicación.
- Crear un servicio compartido que encapsule `ngx-sonner` y `sweetalert2`.
- Definir estilos globales profesionales compatibles con los temas claro y oscuro existentes.
- Sustituir los `window.confirm` actuales del dominio negocio.
- Aplicar notificaciones a autenticación y a mutaciones de negocio ya implementadas.
- Establecer reglas obligatorias para las pantallas frontend de fases posteriores.
- Agregar pruebas unitarias, de integración de componentes y revisión visual responsive.

No incluye:

- Cambios de backend, base de datos o contratos HTTP.
- Cambios funcionales a negocios, autenticación o navegación.
- Edición de producto, catálogo, proveedores o inventario, porque pertenecen al trabajo del colaborador responsable.
- Confirmaciones para acciones reversibles que no tengan consecuencias relevantes.
- Reemplazar mensajes persistentes de validación o errores de carga por notificaciones efímeras.

## Organización propuesta

```text
frontend/src/app/
├── app.component.ts
├── app.component.html
└── shared/
    └── feedback/
        ├── feedback.service.ts
        └── feedback.service.spec.ts

frontend/src/styles.css
```

Reglas:

- Componentes y features consumirán `FeedbackService`; no importarán directamente `toast` ni `Swal`.
- Solo `AppComponent` montará `<ngx-sonner-toaster />`.
- Configuración visual global permanecerá en `styles.css`; no se duplicará en CSS de componentes.
- `shared/feedback` se permite porque contiene una capacidad transversal concreta y nombrada, no funciona como carpeta comodín.
- Tipos de opciones propios del servicio permanecerán junto al servicio mientras no exista otro consumidor.

## Contrato propuesto de `FeedbackService`

El servicio expondrá una API pequeña y estable:

```text
success(titulo, descripcion?)
error(titulo, descripcion?)
info(titulo, descripcion?)
warning(titulo, descripcion?)
confirmDanger(opciones) -> Promise<boolean>
```

`confirmDanger` recibirá como mínimo:

```text
titulo
descripcion
textoConfirmar
textoCancelar opcional, predeterminado "Cancelar"
```

Reglas del contrato:

- Los componentes reciben `boolean`; no conocen `SweetAlertResult`.
- Confirmaciones destructivas no se cierran haciendo clic fuera del diálogo.
- Tecla `Escape` cancela y el foco regresa al control que abrió el diálogo.
- Se usará texto escapado (`titleText` y `text`), nunca HTML construido con datos del usuario.
- Errores mostrados al usuario no expondrán cuerpos HTTP, stack traces ni detalles internos.
- Textos estarán en español, serán breves y describirán resultado o acción siguiente.

## Configuración global propuesta de notificaciones

`NgxSonnerToaster` usará:

- Posición superior derecha.
- Tema enlazado a `ThemeService`, reaccionando a `light` y `dark`.
- Colores semánticos enriquecidos.
- Botón de cierre visible.
- Duración predeterminada de 4 segundos.
- Máximo de 4 notificaciones visibles.
- Separación de 24 px en escritorio y adaptación propia de la biblioteca a 16 px en móvil y áreas seguras.
- Una sola región global para evitar notificaciones duplicadas al navegar.

El estilo de Sonner se ajustará mediante sus variables CSS oficiales `--ngx-sonner-*` y selectores `data-sonner-*`, reutilizando colores, bordes, radios, tipografía y sombras ya definidos por Tienda.

## Configuración global propuesta de confirmaciones

SweetAlert2 se configurará desde el servicio mediante un `mixin` único:

- `buttonsStyling: false` para usar botones del diseño Tienda.
- Botón de cancelar visible y enfocado inicialmente en acciones destructivas.
- Botones en orden seguro: cancelar antes de confirmar.
- Cierre por clic exterior deshabilitado en acciones destructivas.
- Cierre con `Escape` habilitado como cancelación.
- Restauración del foco habilitada.
- Clases globales con prefijo `tienda-alerta` para popup, título, texto, acciones y botones.

`styles.css` importará la hoja base de SweetAlert2 y agregará la capa visual del proyecto:

- Fondo, texto, borde y sombra desde variables de tema existentes.
- Botón destructivo rojo y botón cancelar neutro.
- Estados `hover`, `disabled` y `focus-visible` con contraste suficiente.
- Controles con altura táctil mínima de 44 px.
- Diálogo responsive; botones apilados y de ancho completo en pantallas pequeñas.
- Overlay y `z-index` coherentes con menús, diálogos y notificaciones existentes.
- Respeto a `prefers-reduced-motion`.

## Reglas de uso

### Usar notificación

- Éxito después de una mutación confirmada por API.
- Error de una acción puntual que el usuario puede volver a intentar.
- Aviso breve que no requiere una decisión inmediata.

### Mantener mensaje dentro de la pantalla

- Validación de campos antes de enviar formulario.
- Errores asociados a un campo específico.
- Error de carga inicial que impide usar toda la vista.
- Información que debe permanecer visible para poder corregirla.

### Usar confirmación

- Archivar negocio u otra acción destructiva o de impacto relevante.
- No usar SweetAlert2 como mensaje informativo ni para anunciar éxito.
- Restaurar negocio no requiere confirmación porque es reversible; mostrará notificación de éxito.
- No mostrar simultáneamente toast y mensaje inline para el mismo error.
- Quedan prohibidos `window.alert` y `window.confirm` en código nuevo.

## Flujos actuales que se migrarán

### Autenticación

- Registro correcto: navegar a login y mostrar `Cuenta creada` mediante notificación global; retirar el mensaje duplicado basado en `?registrado=1`.
- Login correcto: navegar sin notificación de éxito para evitar ruido en una acción habitual.
- Logout: navegar siempre a login como hoy, sin notificación de éxito. Si falla, no afirmar que el cierre remoto ocurrió.
- Validaciones y errores de credenciales permanecerán inline porque requieren atención en el formulario.

### Negocios

- Crear: `Negocio registrado` después de respuesta correcta.
- Editar: `Cambios guardados` después de respuesta correcta.
- Archivar: reemplazar `window.confirm` por `confirmDanger`; notificar `Negocio archivado` solo después del éxito.
- Restaurar: ejecutar sin confirmación y notificar `Negocio restaurado` después del éxito.
- Errores de carga completos y validación permanecerán inline.
- Fallos de archivar o restaurar usarán notificación de error sin duplicar un mensaje inline.

## Aplicación obligatoria en fases posteriores

- Toda mutación frontend nueva debe definir su resultado exitoso, error persistente y, si corresponde, confirmación.
- Features usarán exclusivamente `FeedbackService` para notificaciones y confirmaciones.
- Componentes no agregarán estilos globales propios para Sonner o SweetAlert2.
- Acciones reversibles no pedirán confirmación por defecto.
- Acciones destructivas deben usar lenguaje específico; evitar textos genéricos como `¿Está seguro?`.
- Producto y catálogo adoptarán estas reglas únicamente cuando su responsable coordine el cambio; fase 1.6 no modificará sus archivos.

## Estrategia de implementación

1. Crear `FeedbackService` y sus pruebas unitarias.
2. Montar y configurar el toaster global enlazado al tema actual.
3. Importar base de SweetAlert2 y agregar estilos globales de Tienda para ambas bibliotecas.
4. Migrar flujos de autenticación definidos en esta fase.
5. Migrar creación, edición, archivado y restauración de negocios.
6. Eliminar `window.confirm` únicamente de los archivos en alcance.
7. Ejecutar pruebas, build, revisión de accesibilidad y revisión visual responsive en ambos temas.
8. Registrar comandos, resultados y pendientes en esta sección.

## Pruebas de fase 1.6

- `FeedbackService` delega correctamente `success`, `error`, `info` y `warning` a Sonner.
- `confirmDanger` devuelve `true` al confirmar y `false` al cancelar o cerrar con `Escape`.
- La aplicación renderiza exactamente un toaster global.
- El toaster cambia entre tema claro y oscuro junto con `ThemeService`.
- Cancelar archivado no ejecuta petición HTTP ni cambia navegación.
- Confirmar archivado ejecuta una sola petición; éxito y error muestran feedback correcto.
- Crear, editar y restaurar negocio muestran una sola notificación después de éxito.
- Registro conserva validaciones inline y muestra éxito una sola vez; login conserva sus validaciones y navega sin toast de éxito.
- Logout conserva navegación actual y no comunica éxito remoto cuando la API falla.
- Búsqueda estática confirma ausencia de nuevos `window.alert`, `window.confirm` e imports directos productivos de las bibliotecas fuera del servicio y raíz autorizados.
- Revisión visual en 390 px, 768 px y 1440 px, temas claro y oscuro.
- Revisión con teclado cubre foco inicial, tabulación, cancelación con `Escape` y retorno del foco.
- `npm test -- --watch=false` y `npm run build` pasan sin errores nuevos.
- Diff confirma cero modificaciones a producto, catálogo, proveedores e inventario.

## Criterios de aceptación de fase 1.6

- [x] Dependencias instaladas fueron identificadas y sus versiones registradas.
- [x] Estructura, contrato y reglas de uso fueron revisados por el usuario.
- [x] Estilo global profesional fue aprobado por el usuario.
- [x] Fase recibió aprobación explícita para implementación.
- [x] Existe un único toaster global sincronizado con el tema.
- [x] Notificaciones y confirmaciones pasan por `FeedbackService`.
- [x] Flujos actuales de autenticación y negocio fueron migrados sin cambiar contratos HTTP.
- [x] No quedan `window.confirm` en archivos de negocio.
- [x] Producto y catálogo permanecen intactos.
- [x] Pruebas, build, accesibilidad y revisión responsive pasan.
- [x] Resultados de verificación fueron registrados.
- [ ] Usuario acepta resultado final.

## Verificación técnica de fase 1.6

Fecha: 2026-09-10.

- `npm test -- --watch=false`: correcto, 19 archivos y 40 pruebas.
- `npm run build`: correcto.
- Build conserva tres avisos CSS preexistentes: dos presupuestos de `negocio-shared.css` y uno de `productos.component.css`.
- Import directo del build ESM de SweetAlert2 eliminó el aviso nuevo de CommonJS; declaración local conserva tipado mediante el paquete oficial.
- Búsqueda estática: ningún `window.alert` o `window.confirm` permanece en `src/app`.
- Búsqueda estática: imports productivos de Sonner y SweetAlert2 limitados a raíz y `FeedbackService`.
- Revisión visual: confirmación a 1440 px y 390 px en tema oscuro; toast real a 768 px en tema claro.
- Confirmación visual: foco inicial en `Cancelar`, botones apilados en móvil y sin desbordamiento horizontal.
- Prueba visual de registro interceptó la respuesta HTTP dentro del navegador; no escribió usuarios ni datos en backend.
- Diff revisado: ningún archivo de producto, catálogo, proveedores o inventario fue modificado.

## Imprevistos resueltos de fase 1.6

- jsdom no implementa `window.matchMedia`, requerido por `ngx-sonner`; se agregó stub limitado a la prueba de `AppComponent`.
- El import principal de SweetAlert2 generaba advertencia CommonJS en Angular. Se cambió al artefacto ESM oficial y se agregó declaración TypeScript local.
- La primera prueba visual apuntó a `/auth/registro` en vez de `/auth/register`; CORS detuvo la petición antes del backend. Se corrigió el interceptor y la verificación final se ejecutó sin persistencia.

## Decisiones resueltas de fase 1.6

- Contrato, configuración global e implementación aprobados explícitamente el 2026-09-10.
- Login y logout no mostrarán notificación de éxito; registro y mutaciones de negocio sí.

---

# Fase 2: Sucursales

## Estado de fase

`Pendiente`

Alcance reservado: CRUD con archivado, dirección, código único por negocio, sucursal principal, cards responsive y formulario por ruta. No detallar ni implementar hasta cerrar fases 1, 1.5 y 1.6.

Entidades previstas: `sucursales`, reutilizando estructura `direcciones`.

---

# Fase 3: Contexto activo

## Estado de fase

`Pendiente`

Alcance reservado: selector de negocio/sucursal en topbar, persistencia en `localStorage`, validación de membresía y propagación a dashboard y módulos actuales. No detallar ni implementar hasta cerrar fase 2.

---

# Fase 4: RBAC

## Estado de fase

`Pendiente`

Alcance reservado: permisos globales, roles por negocio, varios roles por membresía, migración de `membresias.rol_id` y autorización por permiso. No detallar ni implementar hasta cerrar fase 3.

Entidades previstas: `permisos`, `roles`, `permisos_rol`, `roles_membresia`.

Cadena prevista:

```text
usuario - membresía activa - roles - permisos
```

---

# Fase 5: Empleados

## Estado de fase

`Pendiente`

Alcance reservado: CRUD de personas laborales por negocio, sin exigir cuenta, con vínculo opcional posterior a membresía. No detallar ni implementar hasta cerrar fase 4.

Entidad prevista: `empleados`.

---

# Fase 6: Invitaciones

## Estado de fase

`Pendiente`

Alcance reservado: invitar empleado existente, enlace copiable, token almacenado como hash, registro de cuenta cuando sea necesario, aceptación con correo coincidente y creación transaccional de membresía/roles. No detallar ni implementar hasta cerrar fase 5.

Entidad prevista: `invitaciones_negocio`.

---

# Fase 7: Asignaciones empleado-sucursal

## Estado de fase

`Pendiente`

Alcance reservado: relación muchos-a-muchos entre empleados y sucursales, asignación principal y validación estricta de negocio. No detallar ni implementar hasta cerrar fase 6.

Entidad prevista: `asignaciones_empleado_sucursal`.

---

# Fase 8: Integración y endurecimiento

## Estado de fase

`Pendiente`

Alcance reservado: pruebas multiempresa completas, navegación integrada, errores uniformes, auditoría de aislamiento, regresión y documentación final. No detallar hasta cerrar fases 1 a 7.

---

# Backlog fuera de esta capability activa

No crear tablas, structs o endpoints todavía para:

- `tokens_autenticacion`
- `dispositivos_usuario`
- `sesiones_usuario`
- `eventos_inicio_sesion`
- `versiones_terminos`
- `aceptaciones_terminos_usuario`
- `planes_servicio`
- `suscripciones_negocio`
- `configuraciones_ticket`
- `secuencias_documentos`
- `personal_plataforma`

Estas piezas requieren capabilities propias o ampliación aprobada de alcance.

# Verificación documental

- [x] Capability movida fuera de `_template`.
- [x] Fases ordenadas por dependencia.
- [x] Convención de archivos documentada.
- [x] Decisiones generales registradas.
- [x] Tablas futuras retiradas de alcance activo.
- [x] Fase 1 detallada sin implementar código.
- [x] Decisiones pendientes de fase 1 resueltas.
- [x] Fase 1 aprobada para implementación.
- [x] Fase 1.5 agregada antes de sucursales.
- [x] Organización futura por dominios documentada.
- [x] Producto excluido de fase 1.5.

# Registro de aprobación

| Fecha      | Alcance                         | Decisión    |
| ---------- | ------------------------------- | ----------- |
| 2026-09-10 | Estructura documental por fases | Aprobada    |
| 2026-09-10 | Fase 1: Negocios                | Aprobada explícitamente |
