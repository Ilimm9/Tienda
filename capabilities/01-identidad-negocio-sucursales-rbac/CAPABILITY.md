# Capability 01: Identidad, negocios, sucursales y control de acceso

## Estado

`En implementación`

La estructura por fases fue aprobada el 2026-09-10. Fases 1, 1.5 y 1.6 fueron aprobadas explícitamente; fases posteriores requieren aprobación independiente.

## Control

- Responsable: Usuario del proyecto
- Fecha de creación: 2026-09-10
- Última revisión estructural: 2026-09-11
- Aprobación de estructura: Confirmada
- Aprobación de implementación: Fases 1, 1.5 y 1.6 aprobadas explícitamente el 2026-09-10; fase 2 aprobada explícitamente el 2026-09-11
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
| 1    | [Negocios](fases/01-negocios.md)                       | Cerrada     | Autenticación actual            |
| 1.5  | [Organización por dominios](fases/01.5-organizacion-backend.md)      | Cerrada     | Negocios verificada             |
| 1.6  | [Feedback global del frontend](fases/01.6-feedback-global-frontend.md)   | Cerrada     | Fase 1.5 cerrada                |
| 2    | [Sucursales](fases/02-sucursales.md)                     | Verificada  | Fases 1, 1.5 y 1.6 cerradas     |
| 3    | [Contexto activo](fases/03-contexto-activo.md)                | En implementación | Sucursales verificada; excepción aprobada |
| 4    | [RBAC](fases/04-rbac.md)                           | Pendiente   | Contexto activo cerrado         |
| 5    | [Empleados](fases/05-empleados.md)                      | Pendiente   | RBAC cerrado                    |
| 6    | [Invitaciones](fases/06-invitaciones.md)                   | Pendiente   | Empleados y RBAC cerrados       |
| 7    | [Asignaciones empleado-sucursal](fases/07-asignaciones-empleado-sucursal.md) | Pendiente   | Empleados y sucursales cerrados |
| 8    | [Integración y endurecimiento](fases/08-integracion-endurecimiento.md)   | Pendiente   | Fases 1, 1.5, 1.6 y 2 a 7 cerradas |

---

## Lectura por contexto

Para reducir contexto innecesario, este archivo conserva únicamente reglas y estado general:

- Cargar este `CAPABILITY.md` para conocer alcance transversal, dependencias y estado.
- Cargar solo el archivo de la fase activa dentro de `fases/`.
- Consultar [SEGUIMIENTO.md](SEGUIMIENTO.md) únicamente para backlog, verificación documental y aprobaciones históricas.
- Consultar `capabilities/_template/base.MD` como fuente inicial de modelo solo cuando la fase requiera contrastar entidades o relaciones; nunca modificarlo.

La especificación detallada, decisiones, pruebas y resultados de cada fase permanecen en su archivo correspondiente.
