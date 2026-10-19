# Plan de implementación

Índice general actualizado el 2026-10-19. Reorganización documental solicitada por el usuario; conserva aprobaciones y estados anteriores sin cerrar entregas automáticamente.

## Contexto por tema

- [Registro y alta inicial](capabilities/flujo-registro.md).
- [Gestión de colaboradores](capabilities/gestion-colaboradores.md).
- [Productos y catálogos](capabilities/registro-productos.md).
- [Compras, inventario y precios](capabilities/modulo-compras.md).

## Estado del trabajo

[Planes actuales](capabilities/planes-actuales.md) concentra pendientes, fuentes y límites de autorización: producción segura, correo real/revisión Resend, verificación de plataforma, marca, sistema visual y revisiones de entregas.

Los planes previos, fases, guías, respaldos y capturas se conservan en [capabilities/historial/](capabilities/historial/), sin subcarpetas. El [índice anterior](capabilities/historial/00-plan-implementacion-anterior.md) conserva el registro detallado de IDs y aprobaciones; algunos estados estaban desactualizados respecto de sus documentos y no se reinterpretan como cierre.

## Convención y ciclo

1. Crear plan directamente en `capabilities/NN-nombre.md` (siguiente ID: `19`), sin nuevas carpetas.
2. Definir alcance, reglas, contratos, pruebas, criterios de aceptación, tareas y decisiones; comenzar en `Borrador` o `En revisión`.
3. Registrar aprobación explícita del usuario antes de pasar a `Aprobada` e implementar.
4. Mantener avance y verificaciones; usar `En implementación` y después `Verificada` según corresponda.
5. Actualizar base temática y pendientes; archivar plan terminado como `historial/NN-nombre.md`. Estado `Cerrada` solo tras aceptación final explícita.

Bases de conocimiento se actualizan con cada cambio del tema; no sustituyen planes ni autorizaciones. Organización detallada en [capabilities/README.md](capabilities/README.md).

## Restricciones vigentes

- Git solo lectura hasta confirmación final del usuario.
- Sin commits, push, despliegues ni publicaciones sin autorización explícita.
- Código ejecutable permanece en `backend/` y `frontend/`.
- Desviaciones de un plan aprobado requieren revisión antes de implementarse.
