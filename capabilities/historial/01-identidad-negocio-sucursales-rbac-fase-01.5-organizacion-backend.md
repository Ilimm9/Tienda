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
