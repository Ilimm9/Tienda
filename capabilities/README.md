# Capabilities

Contexto rápido y planeación vigente. Actualizado: 2026-10-20.

## Leer por tema

- [Flujo de registro](flujo-registro.md): cuenta, verificación, acceso y alta inicial.
- [Gestión de colaboradores](gestion-colaboradores.md): empleados, invitaciones, sucursales y permisos.
- [Registro de productos y catálogos](registro-productos.md): altas, variantes, importación y existencias.
- [Módulo de compras](modulo-compras.md): recepción, impuestos, inventario y autorización de precios.
- [Planes actuales](planes-actuales.md): pendientes documentados y límites de aprobación.

## Organización

- Única subcarpeta: [historial/](historial/), plana. Conserva planes anteriores, fases, respaldos, guías, modelo inicial y capturas; nada se elimina.
- Cada `NN-nombre/CAPABILITY.md` anterior ahora es `historial/NN-nombre.md`. Fases y anexos agregan un sufijo; los planes personales conservan número y prefijo `thrs` para evitar colisiones.
- Los planes nuevos se escriben directamente como `NN-nombre.md`; continuar desde `21`; `19` reservado para precios y `20` cerrado. Incluir alcance, reglas, contratos, pruebas, criterios de aceptación, tareas y decisiones.
- Solo aprobación explícita del usuario permite estado `Aprobada` e implementación. Estados: `Borrador`, `En revisión`, `Aprobada`, `En implementación`, `Verificada` y `Cerrada`.
- Al terminar una entrega, actualizar su base temática y mover su plan a `historial/`. Los pendientes y la aceptación final siguen explícitos; archivar no cambia estados.

Las cuatro bases describen comportamiento y tablas del código local, no certifican el estado de una base desplegada. Mantenerlas breves y actualizarlas en cada cambio del tema. Para solicitar cambios, indicar tema, comportamiento actual, resultado esperado y reglas o datos afectados.

Índice general: [PLAN_IMPLEMENTACION.md](../PLAN_IMPLEMENTACION.md). El [modelo inicial](historial/base-modelo-datos.md) se conserva sin modificar; para nombres reales de tablas prevalecen los modelos y migraciones del código.
