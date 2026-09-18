# Capability 03: Carga de archivos unificada

## Estado

En implementación.

## Aprobación

Aprobada explícitamente por el usuario el 2026-09-14 mediante la instrucción: “Implement the plan”.

## Alcance

- Aplicar una interacción consistente a las cargas masivas de Productos y Catálogo (marcas, categorías y unidades).
- Mostrar inmediatamente el archivo elegido o soltado como listo: nombre visible y estado visual verde.
- Permitir sustituir el archivo al elegir o soltar otro.
- Incluir un botón accesible con icono para eliminar el archivo preparado localmente.
- Restablecer archivo, selector nativo, error, resultado y vista previa al sustituir o eliminar el archivo.
- Mantener deshabilitada la interacción durante una validación o importación activa.

## Fuera de alcance

- Cambios a endpoints, contratos HTTP, formato XLSX o límites de archivo.
- Revertir importaciones ya ejecutadas.

## Reglas y contratos

- La selección desde el explorador y el evento `drop` entregan el primer archivo al mismo flujo de estado local.
- El botón de eliminar no abre el selector de archivos y devuelve la zona a su estado inicial.
- Reemplazar o eliminar un archivo invalida cualquier resultado de validación o importación mostrado.
- La carga permanece disponible mediante clic y teclado, y el icono de eliminación tiene etiqueta accesible.

## Pruebas y criterios de aceptación

- Seleccionar un archivo, soltarlo y sustituirlo actualiza inmediatamente el archivo local y limpia estado previo.
- Eliminar el archivo lo remueve y limpia el valor del `input`.
- Las zonas de Productos y Catálogo muestran el estado preparado con el archivo seleccionado y admiten arrastrar/soltar.
- Las pruebas de frontend y el build de producción finalizan correctamente.

## Decisiones y seguimiento

- 2026-09-14: se aplica a todos los cargadores existentes, no solo a Productos.
- 2026-09-14: al elegir o soltar otro archivo, este sustituye al anterior.
- 2026-09-14: se añadieron pruebas unitarias de `drop`, sustitución, limpieza y eliminación para ambos componentes.
- 2026-09-14: `npm test -- --watch=false` y `npm run build` no iniciaron porque el entorno tiene Node.js `v14.15.0`; el proyecto requiere Node 24 según `.nvmrc` y Angular exige al menos Node 22.22.3.
- Pendiente: ejecutar las verificaciones con Node 24 y cambiar el estado a `Verificada` al concluir.
