# Plan de implementación

Índice general actualizado el 2026-10-20. Reorganización documental solicitada por el usuario; conserva aprobaciones y estados anteriores sin cerrar entregas automáticamente.

## Contexto por tema

- [Registro y alta inicial](capabilities/flujo-registro.md).
- [Gestión de colaboradores](capabilities/gestion-colaboradores.md).
- [Productos y catálogos](capabilities/registro-productos.md).
- [Compras, inventario y precios](capabilities/modulo-compras.md).

## Estado del trabajo

[Planes actuales](capabilities/planes-actuales.md) concentra pendientes, fuentes y límites de autorización: producción segura, correo real/revisión Resend, verificación de plataforma, marca, sistema visual y revisiones de entregas.

Capability 20: [Arreglos de invitaciones y permisos](capabilities/historial/20-arreglos-invitaciones.md). **Cerrada** el 2026-10-20 con aceptación final del usuario. Aprobada el 2026-10-19: seis temas implementados, PostgreSQL/Mailpit/Chromium comprobados; ampliar permisos de roles existentes requiere autorización específica adicional. ID 19 reservado para el borrador de precios por sucursal.

Los planes previos, fases, guías, respaldos y capturas se conservan en [capabilities/historial/](capabilities/historial/), sin subcarpetas. El [índice anterior](capabilities/historial/00-plan-implementacion-anterior.md) conserva el registro detallado de IDs y aprobaciones; algunos estados estaban desactualizados respecto de sus documentos y no se reinterpretan como cierre.

## Convención y ciclo

1. Crear plan directamente en `capabilities/NN-nombre.md` (siguiente ID libre: `21`; `19` reservado para precios), sin nuevas carpetas.
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
