# Área de secciones administrativas

## Control

- Estado: En implementación.
- Fecha de creación: 2026-09-21.
- Aprobación: autorización explícita del usuario para implementar el plan el 2026-09-21.
- Dependencias: shell autenticado y estilos actuales del frontend.

## Objetivo

Ampliar de forma uniforme el área útil de las secciones administrativas, sin modificar el `main` ni el sidebar, y aclarar el panel expandido de variantes en Productos.

## Alcance y reglas

- Declarar en estilos globales un ancho máximo de sección de `1440px` y padding lateral adaptable de `16px` a `32px`.
- Aplicar esos tokens a Productos, Catálogo, Equipo, Negocios, Sucursales, Roles/Permisos, Inicio y placeholders administrativos.
- Conservar el comportamiento responsive: el contenido no excede el viewport y el sidebar mantiene sus dimensiones actuales.
- Aclarar el fondo y borde del panel de variantes respetando los temas claro y oscuro.
- Representar la ausencia de selección en los filtros de Categoría, Marca y Estado con `null`, para que PrimeNG no muestre su icono de limpieza antes de elegir un valor.
- No modificar contratos HTTP, modelos, columnas ni comportamiento de tablas.

## Contratos

- Variables CSS públicas: `--section-max-width` y `--section-padding-inline` en `frontend/src/styles.css`.
- Los contenedores de sección consumen esos tokens mediante `width: min(100%, var(--section-max-width))` y padding lateral común.
- Los filtros opcionales de Productos usan `string | null`; `null` significa que no hay filtro activo.

## Tareas y decisiones

- [x] Definir el ancho máximo global en `1440px`.
- [x] Definir padding lateral fluido entre `16px` y `32px`.
- [x] Mantener sin cambios el `main` y el sidebar.
- [x] Aplicar tokens a todos los contenedores administrativos en alcance.
- [x] Aclarar panel desplegable de variantes.
- [x] Corregir el estado vacío de los filtros opcionales de Productos.
- [x] Añadir pruebas unitarias para el estado y la limpieza de filtros.
- [x] Ejecutar verificaciones disponibles del frontend.

## Pruebas y criterios de aceptación

- Las vistas administrativas usan hasta `1440px`, permanecen centradas y muestran padding lateral de al menos `16px`.
- El sidebar expandido y colapsado no cambia de tamaño ni provoca desbordamiento.
- En móvil y tablet, las vistas conservan ancho fluido sin scroll horizontal añadido.
- El panel expandido de variantes se percibe más claro y mantiene contraste en ambos temas.
- `npm run build` y `npm test` del frontend finalizan correctamente, o se registran fallos preexistentes.

## Resultados

- `git diff --check`: correcto, sin errores de espacios.
- `graphify update .`: ejecutado; el grafo de código fue actualizado.
- Se añadió una prueba que verifica el estado inicial `null`, la selección y la limpieza de Categoría, Marca y Estado.
- `npm run build`: no ejecutable en este entorno. Angular CLI requiere Node `v22.22.3` o superior y la terminal dispone de Node `v14.15.0`.
- `npm test -- --watch=false`: no ejecutable por la misma incompatibilidad de Node.
- Pendiente conocido: ejecutar build y pruebas con una versión de Node compatible antes de cambiar el estado a `Verificada`.
