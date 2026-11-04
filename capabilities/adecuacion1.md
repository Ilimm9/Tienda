# Adecuación 1 · Plantilla de productos y Configuración > Vistas

Estado: **Verificada** (pendiente prueba manual y aceptación final) · Fecha: 2026-11-03 · Origen: observaciones de la Lic. (puntos 1 y 2)

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

Contrato implementado:
- Columna `secciones_ocultas varchar(255) not null default ''` en `configuraciones_negocio`, valores separados por comas (`categorias`, `marcas`). Se eligió texto en lugar de `text[]` para no depender de tipos de arreglo de PostgreSQL en GORM; la lista es corta y se normaliza (orden fijo, sin repetidos) en el dominio.
- `GET /api/v1/negocios/:negocioId/configuracion/vistas` → `{ "secciones_ocultas": ["categorias"], "secciones_ocultables": ["categorias", "marcas"] }`.
- `PUT` con `{ "secciones_ocultas": [...] }`: solo propietario (`403 SOLO_PROPIETARIOS`); valor no permitido `400 SECCION_NO_VALIDA`. Si el negocio no tiene fila de configuración, se crea.
- Frontend: `VistasService` carga las vistas del negocio activo (si falla, muestra todo); `vistaGuard` redirige a Inicio; el menú filtra por vista y propietario.

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

### Respuestas del usuario (2026-11-03)

- D1: se quitan Marca, Descripción, Contenido y Unidad de contenido.
- D2: Presentación sale del Excel; el tamaño va en el nombre. El formulario de producto conserva el campo.
- Orden final: `Código de barras | Nombre | Variantes | Precio de venta | Costo | Stock inicial | Unidad de medida | Categoría | Proveedor` (Categoría opcional).
- **P1 (coincidencia con productos existentes), resuelto con opción (a):** la importación reconoce productos por nombre + presentación (`productImportNamePresentationKey`) y las familias de variantes por nombre + presentación + marca + categoría (`productBaseKey`, `infrastructure/product_repository.go`). Sin Presentación ni Marca en el Excel, un producto creado en el formulario con presentación «600 ml» o con marca no coincidirá con el renglón importado y podría duplicarse. Opciones: (a) el código de barras manda: si existe, es ese producto; sin código, comparar solo por nombre; (b) dejar la regla y advertir. Recomendación: (a).
- D3: Categoría se queda, opcional, **antes de Proveedor**.
- D4: Marca sale del Excel y se puede ocultar en Vistas.
- D5: solo se acepta la plantilla nueva; las columnas deben coincidir exactamente.
- D6: Descripción se queda en el formulario como opcional.
- D7: por negocio. D8: solo propietario. D9: todo visible al inicio.
- P1: opción (a), el código de barras manda; sin código se compara solo por nombre.
- D10: solo Categorías y Marcas.

## Tareas

- [x] Resolver D1–D10 y P1.
- [x] Backend: constante única de encabezados; nuevo orden; quitar columnas; instrucciones; mensaje de plantilla anterior; columnas exactas (se tolera `Errores` al final para volver a subir el archivo de errores).
- [x] Backend: identidad por código y luego por nombre; variantes se agregan a la familia existente por nombre (`FamiliaClave`).
- [x] Backend: columna `secciones_ocultas`, endpoints GET/PUT con validación y regla de propietario.
- [x] Frontend: carga masiva sin cambios de pantalla (las columnas vienen de la API); «Agregar del catálogo» oculta Categoría.
- [x] Frontend: grupo Configuración > Vistas, pantalla, servicio, filtro de menú y guard de rutas ocultas.
- [x] Frontend: ocultar campo/filtro/columna de Categoría y Marca en Productos según vistas.
- [x] Actualizar `registro-productos.md` y `PLAN_IMPLEMENTACION.md`.

Hallazgos corregidos durante la verificación:
- La vista previa no mostraba las advertencias del repositorio (cambios de precio/stock, nombre distinto al del código): `prepareProductImport` solo copiaba errores. Ahora copia también advertencias (falla previa a esta adecuación).
- Mensaje de «nombre existente con código nuevo» aclarado: «ya existe un producto con ese nombre y este código no está registrado; si es otra versión, indica la variante…».

## Pruebas

- Go: plantilla descargada tiene el nuevo orden; importación válida con nuevo formato; plantilla anterior rechazada con mensaje; archivo de errores con nuevas columnas.
- Go: GET/PUT vistas; valor no permitido → 400; no propietario → 403; otro negocio → aislamiento.
- Angular: menú sin Categorías/Marcas cuando están ocultas; ruta oculta redirige; formulario y lista sin esos campos; pantalla Vistas guarda y refresca menú.
- Manual: descargar plantilla, llenar y subir; ocultar y mostrar Categorías con productos que ya tienen categoría (se conserva).

## Resultados (2026-11-03)

- Go (`go vet`, `go test ./...` con integración PostgreSQL en `tienda_security_test`): todo en verde salvo `TestProductionAcceptsStrongOTPSecret` (falla previa conocida).
- Angular (`ng test`): 289 correctas; 6 fallas previas de alta inicial (`window.matchMedia`). `ng build` correcto; aviso de presupuesto CSS en `vistas.component.css` (5.2 kB > 4 kB), igual que otros componentes.
- Chromium/Playwright contra copia temporal `tienda_e2e` (API :18080, Angular :4400), cuenta del propietario:
  - Vistas: ocultar Categorías la quita del menú, `/catalogo/categorias` redirige a Inicio, sin filtro/columna en Productos ni campo en el formulario; al guardar un producto su categoría se conserva en `producto_categorias`.
  - Ocultar Marcas: igual (menú, ruta, filtro, campo). Mostrar todo restaura menú y columnas.
  - Plantilla descargada: 9 columnas en el orden nuevo e instrucciones actualizadas.
  - Plantilla anterior: rechazada con el mensaje esperado.
  - Archivo nuevo: código existente con otro nombre se actualiza con aviso; nombre sin código coincide; producto nuevo se crea sin presentación ni marca; nombre existente con código nuevo pide variante. Archivo de errores con el formato nuevo + `Errores`. Importación real: 1 creado, 2 actualizados.
  - Ancho de teléfono (390 px): Vistas sin desplazamiento horizontal.
- Base `tienda` sin cambios (conteos y última actualización iguales); bases temporales eliminadas. Al iniciar la API normal, AutoMigrate agrega `secciones_ocultas`.

## Criterios de aceptación

- La plantilla descargada es `Código de barras | Nombre | Variantes | Precio de venta | Costo | Stock inicial | Unidad de medida | Categoría | Proveedor` y no contiene Marca, Descripción, Presentación ni Contenido.
- Un negocio puede ocultar Categorías y Marcas desde Configuración > Vistas; el cambio aplica a todo su equipo sin perder datos.
