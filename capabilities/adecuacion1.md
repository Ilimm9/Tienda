# Adecuación 1 · Plantilla de productos y Configuración > Vistas

Estado: **Plan, con decisiones pendientes** · Fecha: 2026-11-03 · Origen: observaciones de la Lic. (puntos 1 y 2)

## Objetivo

1. Simplificar el archivo Excel de carga masiva de productos: quitar columnas sin uso y ordenar las que sirven para identificar el producto.
2. Permitir que cada negocio oculte secciones que no usa (primero **Categorías** y **Marcas**) desde un nuevo menú **Configuración > Vistas**.

## Situación actual

- Plantilla (`backend/internal/application/catalog_import.go`, 3 lugares + pruebas), 14 columnas en este orden:
  `Nombre | Categoría | Marca | Descripción | Presentación | Contenido | Unidad de contenido | Unidad de medida | Precio de venta | Costo | Stock inicial | Código de barras | Variantes | Proveedor`
- La lectura valida encabezados **por posición** (`readImportRowsWithOptions`): un archivo con otro orden se rechaza con «se esperaba la columna X».
- El archivo de errores descargable usa el mismo formato más la columna `Errores`.
- Producto simple se identifica por **nombre + presentación** normalizados (o código).
- Menú (`frontend/src/app/layout/navigation.config.ts`): Catálogo > Productos, Marcas, Categorías, Unidades de medida. No existe sección Configuración.
- Ya existe `configuraciones_negocio` (una fila por negocio) en el dominio `negocio`.

## Parte A · Plantilla Excel

Nuevo orden propuesto (pedido por la Lic.):

`Código de barras | Nombre | Variantes | Precio de venta | Costo | Stock inicial | Unidad de medida | Proveedor`

Más las columnas que se decidan en D2 y D3 (al final, opcionales).

Cambios:
- Una sola constante de encabezados de productos (hoy está repetida en 3 lugares) para que plantilla, lectura y archivo de errores no se desalineen.
- Quitar `Marca`, `Descripción`, `Contenido` (y `Unidad de contenido`, ver D1) de plantilla, lectura y archivo de errores.
- Actualizar hoja de instrucciones de la plantilla.
- Vista previa / resultado de carga masiva (frontend): quitar columnas eliminadas y respetar el nuevo orden.
- Pruebas Go de importación y specs Angular del reporte de importación.

## Parte B · Configuración > Vistas

- Nuevo grupo de menú **Configuración** con hijo **Vistas** (`/configuracion/vistas`), al final del menú.
- Pantalla: lista de secciones ocultables, cada una con interruptor, qué se oculta y cuántos productos la usan. Botón «Guardar cambios». Diseño: tablero «11 · Configuración: vistas» en el canvas de secciones.
- Al ocultar **Categorías**: desaparece del menú, su ruta redirige a Inicio, se ocultan campo en formulario de producto, filtro y columna de la lista.
- Al ocultar **Marcas**: igual (menú, ruta, campo, filtro).
- Ocultar **no borra datos**: las asociaciones existentes se conservan y vuelven a verse al mostrarla de nuevo. La API sigue aceptando los campos (es preferencia visual, no permiso).
- **Productos** y **Unidades de medida** no se pueden ocultar (las usan compras e inventario).

Contrato propuesto:
- Columna `secciones_ocultas text[] not null default '{}'` en `configuraciones_negocio` (valores permitidos: `categorias`, `marcas`).
- `GET /api/negocios/{id}/configuracion/vistas` → `{ "secciones_ocultas": ["categorias"] }`.
- `PUT /api/negocios/{id}/configuracion/vistas` con el mismo cuerpo; rechaza valores no permitidos (400).
- El frontend carga las vistas junto con el contexto del negocio y filtra menú, rutas y campos con ellas.

## Decisiones a tomar

| # | Decisión | Opciones | Recomendación |
| --- | --- | --- | --- |
| D1 | `Unidad de contenido` | Quitarla junto con Contenido · dejarla | **Quitarla**: sin Contenido no significa nada. |
| D2 | `Presentación` (no está en la lista de la Lic.) | Quitarla · dejarla opcional después de Nombre | **Dejarla después de Nombre**: es parte de la identidad del producto (Coca 600 ml vs Coca 2 L); sin ella, productos con mismo nombre chocan. |
| D3 | `Categoría` en el Excel | Quitarla · dejarla opcional al final | **Al final, opcional**: quien usa categorías no pierde la carga masiva; quien no, la deja vacía. |
| D4 | Alcance de «retirar Marca» | Solo del Excel · también ocultable en Vistas | **Excel + ocultable en Vistas**. Formulario conserva el campo mientras Marcas esté visible. |
| D5 | Archivos con la plantilla anterior | Rechazar con mensaje claro · aceptar ambos formatos | **Rechazar** con «Esta es la plantilla anterior; descarga la nueva». Aceptar ambos complica lectura y errores. |
| D6 | `Descripción` en el formulario de producto | Dejarla · quitarla | **Dejarla** (opcional); la Lic. solo pidió quitarla del Excel. |
| D7 | Alcance de Vistas | Por negocio · por sucursal · por usuario | **Por negocio**: todo el equipo ve lo mismo. |
| D8 | Quién cambia Vistas | Solo propietario · nuevo permiso `configuracion.gestionar` | **Solo propietario** (igual que crear productos, entrega 31). |
| D9 | Valor inicial en negocios nuevos | Todo visible · Marcas oculta | **Todo visible**; cada negocio decide. |
| D10 | Secciones ocultables en esta entrega | Categorías y Marcas · también Proveedores u otras | **Categorías y Marcas**; agregar más es solo ampliar la lista permitida. |

## Tareas

- [ ] Resolver D1–D10.
- [ ] Backend: constante única de encabezados; nuevo orden; quitar columnas; instrucciones; mensaje de plantilla anterior.
- [ ] Backend: columna `secciones_ocultas`, endpoints GET/PUT con validación y regla de propietario.
- [ ] Frontend: carga masiva (vista previa, resultado, descarga de errores) con nuevas columnas.
- [ ] Frontend: grupo Configuración > Vistas, pantalla, servicio, filtro de menú y guard de rutas ocultas.
- [ ] Frontend: ocultar campo/filtro/columna de Categoría y Marca en Productos según vistas.
- [ ] Actualizar `registro-productos.md` y `PLAN_IMPLEMENTACION.md`.

## Pruebas

- Go: plantilla descargada tiene el nuevo orden; importación válida con nuevo formato; plantilla anterior rechazada con mensaje; archivo de errores con nuevas columnas.
- Go: GET/PUT vistas; valor no permitido → 400; no propietario → 403; otro negocio → aislamiento.
- Angular: menú sin Categorías/Marcas cuando están ocultas; ruta oculta redirige; formulario y lista sin esos campos; pantalla Vistas guarda y refresca menú.
- Manual: descargar plantilla, llenar y subir; ocultar y mostrar Categorías con productos que ya tienen categoría (se conserva).

## Criterios de aceptación

- La plantilla descargada empieza con `Código de barras | Nombre | Variantes | Precio de venta | Costo | Stock inicial | Unidad de medida | Proveedor` y no contiene Marca, Descripción ni Contenido.
- Un negocio puede ocultar Categorías y Marcas desde Configuración > Vistas; el cambio aplica a todo su equipo sin perder datos.
