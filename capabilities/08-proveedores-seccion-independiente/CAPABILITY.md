# Capability 08: Proveedores como sección independiente

## Estado

En implementación.

## Aprobación

Aprobada explícitamente por el usuario el 2026-09-22 mediante la instrucción: “Implement the plan”.

Ampliación aprobada explícitamente por el usuario el 2026-09-22 mediante la instrucción: “Implement the plan”, tras solicitar formularios de página completa y carga masiva.

## Alcance

- Separar Proveedores de los elementos hijos de Catálogo en el menú lateral.
- Añadir Proveedores como entrada principal del menú lateral, con la ruta `/proveedores`.
- Reutilizar la pantalla y operaciones actuales de proveedores, incluidos el guardia de contexto y el breadcrumb.
- Eliminar la ruta anterior `/catalogo/proveedores`.
- Sustituir el modal de alta y edición por rutas de formulario de página completa.
- Incorporar carga masiva XLSX de proveedores, con plantilla, validación, resultado por fila e inserción transaccional.

## Fuera de alcance

- Redirección o compatibilidad para marcadores de `/catalogo/proveedores`.

## Reglas y contratos

- `NAVIGATION_ITEMS` expone Proveedores como elemento raíz con icono `pi pi-truck` y ruta `/proveedores`.
- La ruta `/proveedores` usa `contextoGuard`, conserva `breadcrumb: 'Proveedores'` y carga `CatalogoComponent` con `section: 'proveedores'`.
- `CATALOGO_ROUTES` no declara una ruta `proveedores`.
- La URL eliminada cae en el fallback global existente hacia Inicio.
- Las rutas `/proveedores/nuevo`, `/proveedores/:id/editar` y `/proveedores/importar` usan `CatalogoComponent` y `contextoGuard`.
- La plantilla XLSX de proveedores usa, en orden: `Nombre`, `Razón social`, `RFC`, `Teléfono`, `Correo`, `Dirección`; solo Nombre es obligatorio.
- `POST /negocios/:negocioId/catalogo/proveedores/importar` recibe `archivo` XLSX de hasta 5 MB y devuelve `CatalogImportResult`.
- Los nombres normalizados repetidos dentro del archivo o ya existentes se omiten y se reportan; no se actualizan registros existentes.

## Tareas y decisiones

- [x] Confirmar que Proveedores tendrá una ruta propia: `/proveedores`.
- [x] Confirmar que `/catalogo/proveedores` se elimina sin redirección.
- [x] Mover Proveedores a una entrada raíz de la navegación.
- [x] Registrar la ruta raíz y retirar la ruta hija de Catálogo.
- [x] Añadir pruebas de navegación y rutas, y ejecutar la comprobación estática.
- [x] Actualizar el grafo del proyecto y registrar resultados.
- [x] Reemplazar el modal de Proveedores por formularios de página completa.
- [x] Implementar la importación XLSX de proveedores en frontend y backend.
- [x] Añadir pruebas de rutas, UI, plantilla, parser e importación de proveedores.
- [x] Corregir títulos singulares de formularios sin recortar el texto visible.

## Pruebas y criterios de aceptación

- El menú muestra Proveedores fuera de Catálogo y dirige a `/proveedores`.
- `/proveedores` requiere un contexto activo y carga la sección de proveedores.
- `/catalogo/proveedores` no se encuentra entre las rutas de Catálogo.
- Las pruebas del frontend y el build de producción finalizan correctamente, o los bloqueos del entorno se registran aquí.

## Resultados

- `./node_modules/.bin/tsc --noEmit -p tsconfig.spec.json`: correcto.
- `git diff --check` y `git diff --cached --check`: correctos, sin errores de espacios.
- `npm test -- --watch=false`: no inicia; Angular CLI requiere Node `v22.22.3` o superior y el entorno dispone de Node `v14.15.0`.
- `npm run build`: no inicia por la misma incompatibilidad de Node.
- `graphify update .`: correcto; actualizó `graphify-out/graph.json`, `graph.html` y `GRAPH_REPORT.md`.
- Pendiente conocido: ejecutar las pruebas de Angular y el build con Node compatible antes de cambiar el estado a `Verificada`.
- `go test ./...`: correcto; incluye las pruebas de plantilla y parser de proveedores.
- `./node_modules/.bin/tsc --noEmit -p tsconfig.spec.json`: correcto después de añadir las rutas y formularios de proveedores.
- Los formularios usan títulos singulares explícitos; evita recortes como “Proveedore” o “Unidades de medid”.
