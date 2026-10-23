# Capability 32: Inicio con bienvenida, primeros pasos y consejos

- Estado: **Verificada** (pendiente de prueba manual y aceptación final).
- Aprobación explícita: 2026-10-23 («okey por favor implementalo»), sobre la propuesta del artboard **J1 · Inicio: bienvenida, pasos y consejos** del canvas de diseño.
- Fecha: 2026-10-23.
- Organización conforme a [AGENTS.md](../AGENTS.md). Solo frontend: `features/home`. Sin cambios de backend ni de contratos.

## Objetivo

Que la pantalla Inicio, con empresa y sucursal listas, dé una bienvenida cálida, muestre los siguientes pasos del propietario y enseñe consejos reales de la app (por ejemplo, cargar productos sin dar de alta antes marcas o proveedores).

## Alcance y reglas

Aplica solo cuando `estadoAlta` es `listo`. Las vistas de alta inicial (`requiere_empresa`, `requiere_sucursal`, `sin_asignacion`) no cambian.

1. **Bienvenida**
   - Fecha de hoy y saludo según la hora en `America/Mexico_City`: «Buenos días» (5:00–11:59), «Buenas tardes» (12:00–18:59), «Buenas noches» (resto).
   - Sin nombre: la sesión solo expone `correo` (ver Decisiones).
   - Frase con la sucursal y la empresa activas; pastillas con el código de la sucursal, «Sucursal activa» y el tipo de miembro (Propietario o Miembro).
   - Mosaico decorativo con las clases existentes `mosaico mosaico--chico`.
2. **Primeros pasos** (solo `tipo_miembro === 'propietario'`)
   - Cinco pasos: empresa y sucursal (siempre hechos en este estado), productos en la sucursal activa, equipo (al menos un empleado) y primera compra.
   - Fuentes, sin endpoints nuevos: `ProductosService.listByBusiness(negocio, sucursal)`, `EmpleadoService.listar(negocio)` y `ComprasService.list(negocio)`; un paso está hecho si `total > 0` (o `items.length` si falta `total`).
   - Si alguna consulta falla, la sección no se muestra. Si los cinco están hechos, tampoco.
   - Cada paso pendiente enlaza a su pantalla: `/productos/importar`, `/equipo/empleados/nuevo`, `/compras/nueva`.
3. **Consejos** («¿Sabías que…?»), para todos, con anterior/siguiente y contador «n de 5». Textos verificados contra [registro-productos.md](registro-productos.md) y [modulo-compras.md](modulo-compras.md):
   - Carga productos sin preparar catálogos: los nombres nuevos de marcas, categorías, unidades y proveedores se crean con una sola confirmación.
   - El código de barras es opcional.
   - Una sucursal nueva empieza vacía; «Agregar del catálogo».
   - En Registrar compra, «Crear nuevo…»; un producto puede tener varios proveedores.
   - El precio es por sucursal y queda en el historial al editar.
4. **Ir directo a**: Productos (`catalogo.ver`), Compras (`compras.ver`), Sucursales (`sucursales.ver`), Equipo (`equipo.empleados.ver`) y Negocios (`puedeVerNegocios`), con los mismos permisos de hoy.

## Pruebas y criterios de aceptación

- `home.component.spec.ts`: saludo por hora; pasos calculados con datos simulados; sección oculta para miembro, con error o con todo hecho; consejos avanzan y regresan en ciclo; accesos filtrados por permiso; vistas de alta inicial sin cambios.
- `ng test` de `features/home` y `ng build` correctos.

## Decisiones

- 2026-10-23: el diseño saluda por nombre («Buenos días, Pedro»), pero la sesión (`AuthenticatedUser`) solo trae `id`, `correo` y `roles_plataforma`. Mostrar el nombre requiere ampliar la respuesta de sesión del dominio `cuenta`; queda fuera de esta capability y el saludo va sin nombre.
- 2026-10-23: las listas de productos, empleados y compras no están paginadas; para el Inicio se descargan completas. Aceptable con volúmenes actuales; si crecen, conviene un endpoint de resumen.

## Tareas

- [x] Componente, plantilla y estilos de Inicio.
- [x] Pruebas.
- [x] Verificación (`ng test`, `ng build`).

## Verificación

- 2026-10-23: `ng test --include src/app/features/home` con Node 24: 11 de 11 correctas (saludo por hora, bienvenida y accesos, pasos con datos simulados, ocultos para miembro, con todo hecho y con error, consejos en ciclo, vistas de alta inicial).
- 2026-10-23: `ng build` correcto. `home.component.css` rebasa el presupuesto de 4 kB por 1.48 kB (advertencia, como otros componentes).
- Pendiente: prueba manual en navegador con un propietario real y aceptación final.
