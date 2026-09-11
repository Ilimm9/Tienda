# Plan de implementación

Este archivo registra el orden, dependencia y estado de las capabilities del proyecto Tienda. Ninguna funcionalidad puede implementarse sin especificación aprobada.

## Flujo de trabajo

1. Describir la funcionalidad.
2. Crear `capabilities/NN-nombre/CAPABILITY.md` con estado `Borrador`.
3. Completar alcance, decisiones, contratos, pruebas y criterios de aceptación.
4. Revisar la propuesta con el usuario.
5. Registrar aprobación explícita y cambiar estado a `Aprobada`.
6. Implementar y cambiar estado a `En implementación`.
7. Ejecutar verificaciones y registrar resultados.
8. Cambiar estado a `Verificada` cuando todos los criterios pasen.
9. Cambiar estado a `Cerrada` después de la revisión final del usuario.

## Estados permitidos

- `Borrador`: definición inicial incompleta o no revisada.
- `En revisión`: lista para recibir decisiones o correcciones.
- `Aprobada`: autorizada explícitamente para implementación.
- `En implementación`: trabajo de código activo.
- `Verificada`: implementación terminada y pruebas registradas.
- `Cerrada`: aceptada finalmente por el usuario.

## Registro de capabilities

| ID | Capability | Estado | Dependencias | Aprobación | Notas |
| --- | --- | --- | --- | --- | --- |
| 01 | [Identidad, negocios, sucursales y RBAC](capabilities/01-identidad-negocio-sucursales-rbac/CAPABILITY.md) | En implementación | Autenticación actual | Fase 1 aprobada el 2026-09-10 | Fase 1 verificada, pendiente aceptación final; fases 2 a 8 pendientes |

## Restricciones vigentes

- Git solo lectura hasta confirmación final del usuario.
- Sin commits, push, despliegues ni publicaciones.
- Código real permanece en `backend/` y `frontend/`.
- Cualquier desviación de una capability aprobada detiene esa parte del trabajo hasta revisión.
