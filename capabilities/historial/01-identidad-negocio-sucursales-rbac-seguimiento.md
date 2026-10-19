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
- [x] Fase 2 detallada, aprobada, implementada y verificada.
- [x] Capability dividida por fases para limitar el contexto documental.
- [x] Fase 3 redactada y puesta en revisión, sin autorización de implementación.
- [x] Fase 3 terminada, verificada y cerrada.
- [x] Fase 2 aceptada y cerrada.
- [x] Fases 4 a 8 detalladas, implementadas y verificadas.
- [x] Modelos legacy de `roles` y `empleados` convergidos con `base.MD`.
- [x] Autorización por `tipo_miembro` sustituida por permisos en toda la capability.
- [x] Auditoría de aislamiento multiempresa ejecutada de extremo a extremo.
- [x] `capabilities/_template/base.MD` preservado sin cambios.

# Registro de aprobación

| Fecha      | Alcance                         | Decisión    |
| ---------- | ------------------------------- | ----------- |
| 2026-09-10 | Estructura documental por fases | Aprobada    |
| 2026-09-10 | Fase 1: Negocios                | Aprobada explícitamente |
| 2026-09-11 | Fase 2: Sucursales              | Aprobada explícitamente; implementación verificada |
| 2026-09-11 | Fase 3: Contexto activo          | Redacción solicitada; en revisión y sin aprobación de implementación |
| 2026-09-12 | Fases 4 a 8: idea preliminar     | Redactadas a solicitud del usuario, sin implementar |
| 2026-09-12 | Convergencia con `base.MD` de `roles` y `empleados` | Solicitada explícitamente por el usuario |
| 2026-09-12 | Cierre de fase 3 y ejecución no supervisada de fases 4 a 8 | Autorizada explícitamente por el usuario |
| 2026-09-12 | Fase 2: Sucursales               | Cerrada |
| 2026-09-12 | Fase 3: Contexto activo          | Implementada, verificada y cerrada |
| 2026-09-12 | Fase 4: RBAC                     | Implementada y verificada |
| 2026-09-12 | Fase 5: Empleados                | Implementada y verificada |
| 2026-09-12 | Fase 6: Invitaciones             | Implementada y verificada |
| 2026-09-12 | Fase 7: Asignaciones empleado-sucursal | Implementada y verificada |
| 2026-09-12 | Fase 8: Integración y endurecimiento | Implementada y verificada; pendiente aceptación final del usuario |
