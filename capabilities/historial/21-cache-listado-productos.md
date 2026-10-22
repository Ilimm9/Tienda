# Lista de productos en memoria con refresco silencioso

- Estado: **Cerrada**. Aceptación final explícita del usuario: 2026-10-21 («es correcto»).
- Aprobación explícita: 2026-10-21. El usuario eligió «Mostrar al instante y refrescar en silencio» y aprobó el plan.
- Fecha: 2026-10-21.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Solo frontend del dominio `productos`; sin cambios de API.

## Objetivo

Al cambiar de ventana o de sección y regresar a Productos, no esperar a que se vuelvan a descargar todos los productos (más de 1500). La lista se muestra al instante y se actualiza en segundo plano; se recarga con espera solo cuando es necesario, por ejemplo después de crear, editar, desactivar, ajustar stock o hacer una carga masiva.

## Causas encontradas

1. `ContextoService` revalida el contexto al recuperar el foco o la visibilidad de la ventana (`focus` / `visibilitychange`) y asigna objetos nuevos de negocio y sucursal con los mismos IDs. El `effect` de `ProductosComponent` dependía del objeto completo, así que cada regreso a la ventana vaciaba la tabla y descargaba todo otra vez.
2. La lista vivía solo en el componente; al navegar a otra sección se perdía.

## Alcance y reglas

- `ProductosService` (root) guarda en memoria la última lista por `negocio:sucursal` (`listaGuardada`). Se actualiza con cada respuesta de `listByBusiness`.
- La caché se vacía (`olvidarListas`) después de `create`, `update`, `adjustInventory` y `deactivate`, y cuando una importación termina en estado `completada`. Así, después de esos cambios, la lista llega fresca con indicador de carga.
- La caché se vacía cuando `ContextoService.negocio()` pasa a `null`, por ejemplo al cerrar sesión con `contexto.limpiar()`.
- `ProductosComponent` recarga solo si cambia el ID de negocio o de sucursal (`computed` de IDs), no cuando el contexto se revalida con los mismos valores.
- Con lista guardada: se muestra de inmediato sin indicador de carga y se refresca en segundo plano. Los cambios de Compras o Precios se reflejan al llegar la respuesta.
- Si falla el refresco en segundo plano, se conservan los datos guardados y se muestra un aviso (`FeedbackService.error`), no el error de pantalla completa.
- Se conserva la protección contra respuestas tardías de otro negocio o sucursal.
- No se persiste en `localStorage`; al recargar el navegador se vuelve a cargar con espera.

## Fuera de alcance

- Conservar la búsqueda, los filtros y la página al regresar.
- Paginación en el servidor.
- Endpoint de detalle para editar (la edición sigue pidiendo datos frescos).

## Archivos

- `frontend/src/app/features/productos/productos.service.ts`
- `frontend/src/app/features/productos/productos.component.ts`
- `frontend/src/app/features/productos/productos.service.spec.ts`
- `frontend/src/app/features/productos/productos.component.spec.ts`

## Pruebas ejecutadas (2026-10-21, Node 22.23.2)

- `ng test --include src/app/features/productos`: 30/30 correctas. Pruebas nuevas:
  - guarda la lista por negocio y sucursal;
  - la olvida tras desactivar o ajustar inventario, y solo cuando la importación termina;
  - la olvida al perder el negocio activo;
  - muestra la lista guardada sin indicador de carga y la reemplaza al llegar el refresco;
  - si falla el refresco, conserva la lista sin error de pantalla;
  - no recarga cuando la sucursal se revalida con el mismo ID.
- Se corrigió en `productos.service.spec.ts` el payload de prueba, al que le faltaba `costo_proveedor`, obligatorio desde el merge `c8030a1`. Impedía compilar la suite.
- `ng test` completo: 222 correctas y 7 fallidas al primer intento. Las 6 de `alta-inicial.spec.ts` son la falla ya registrada de `window.matchMedia` en jsdom. La 7.ª era la prueba nueva, corregida y en verde después.
- `ng build`: correcto. Se mantiene la advertencia previa de presupuesto CSS en `compras.component.css`.

## Cierre

- Comportamiento confirmado por el usuario el 2026-10-21 y registrado como aceptación final.
- Sin pendientes propios. La falla de `alta-inicial` (`window.matchMedia`) sigue registrada en planes actuales.
