# Registro de productos y catálogos

Base vigente según código local. Actualizado: 2026-11-03.

## Flujo y reglas

- Registrar productos simples o variantes dentro del negocio activo. `productos` describe el artículo; `producto_negocio` guarda su oferta comercial y SKU; existencias se separan por sucursal.
- Crear un producto y la carga masiva son solo del propietario del negocio (`403 SOLO_PROPIETARIOS` para los demás); «Crear nuevo…» en Registrar compra sigue disponible y crea el producto sin precio ([31](historial/31-crear-productos-solo-propietario.md)).
- Crear, editar o importar un producto fija su proveedor principal (el que se muestra en Productos) sin borrar los demás proveedores que lo surten; Registrar compra liga cada producto comprado a su proveedor ([27](historial/27-compras-cualquier-producto.md)).
- Marca, categoría, unidad y proveedor son catálogos asociados. Catálogo permite gestionarlos; Proveedores tiene sección propia. Validar referencias del mismo negocio y dependencias antes de borrar.
- Plantilla de carga masiva ([adecuación 1](adecuacion1.md)): `Código de barras | Nombre | Variantes | Precio de venta | Costo | Stock inicial | Unidad de medida | Categoría | Proveedor`. Los encabezados deben coincidir exactamente en nombre y orden, sin columnas extra (solo se tolera `Errores` al final, para volver a subir el archivo de errores); la plantilla anterior se rechaza con «es la plantilla anterior; descarga la plantilla nueva». Marca, Descripción, Presentación, Contenido y Unidad de contenido ya no se importan; el formulario conserva Presentación y Descripción (opcionales) y Marca mientras la sección esté visible.
- La importación XLSX primero valida y muestra errores, advertencias y catálogos nuevos; después procesa una tarea con progreso. SKU vacío se genera por negocio. La columna de código de barras sigue obligatoria, pero sus celdas pueden estar vacías.
- La carga masiva usa la sucursal del contexto (sin selector propio). La vista previa y el resultado agrupan Errores, Advertencias, Catálogos nuevos y Correctos en secciones compactables, con filtro por columna y paginación de 10 en 10; los errores se descargan en XLSX con el formato de la plantilla más la columna `Errores`, listo para corregir y volver a subir. El costo importado se redondea a 2 decimales; el stock negativo se rechaza ([24](historial/24-correccion-masiva-producto.md)).
- Marcas, categorías, unidades y proveedores vacíos no crean asociaciones. Si aparecen nombres nuevos, exigir confirmación global `autorizar_catalogos_nuevos=true`; sin ella no persistir la importación que los requiere.
- Identidad en la importación: **el código de barras manda**; si ya existe, la fila es ese producto (si el nombre difiere se avisa y se conservan los datos del catálogo). Sin código, se busca por nombre normalizado (también coincide «nombre presentación»); varios productos con ese nombre exigen el código. Un nombre existente con un código nuevo exige `Variantes` con `Atributo: valor`. Las variantes se agregan a la familia existente con ese nombre aunque tenga marca o presentación; cada variante conserva SKU, código, precio e inventario propios. Códigos repetidos conservan su error independiente.
- El listado exige `sucursal_id` activa y accesible: stock, totales, estado y detalle de variantes muestran solo esa sucursal. Con varias asignadas se debe cambiar de contexto; sin selección válida no se usan existencias globales. Respuestas tardías no restauran datos del contexto anterior. El frontend guarda en memoria la última lista por negocio y sucursal: al regresar la muestra de inmediato y la refresca en segundo plano; se vacía al crear, editar, desactivar, ajustar inventario, completar una importación o cerrar sesión, y solo un cambio real de negocio o sucursal recarga la vista ([21](historial/21-cache-listado-productos.md)). El catálogo sigue siendo del negocio; precios por sucursal conservan su plan independiente.
- Editar un producto guarda su precio de venta, y el de sus variantes, en `precios_sucursal` de la sucursal activa. Solo con `precios.autorizar`: el cambio crea un evento `AJUSTE_PRECIO_MANUAL` con motivo opcional (por defecto «Edición del producto»). Sin cambio de precio no se escribe nada. Las variantes nuevas reciben su precio en la sucursal activa, o quedan sin precio si se captura 0 ([29](historial/29-guardar-precio-al-editar.md)).
- **Surtido por sucursal** ([28](historial/28-surtido-por-sucursal.md)): un producto o variante pertenece a la sucursal cuando tiene fila en `inventario_sucursal` (aunque el stock sea 0). La lista muestra solo los de la sucursal activa; una sucursal nueva empieza vacía. El propietario agrega productos del catálogo con «Agregar del catálogo» (sin precio ni stock). Crear un producto que ya existe en otra sucursal ofrece «Agregar a esta sucursal» con el precio, costo y stock capturados.
- La carga masiva no duplica productos del negocio: un producto existente que no está en la sucursal se agrega con su precio, costo y stock; si ya está, se reemplaza el stock (movimiento `AJUSTE_MANUAL`) y se actualizan precio y costo (evento `CARGA_MASIVA`). Los cambios se avisan como advertencias, una celda de costo vacía conserva el costo y los datos compartidos del catálogo no cambian. El reporte agrega «Se agregarán a esta sucursal» y «Se actualizarán».
- **Configuración > Vistas** ([adecuación 1](adecuacion1.md)): el propietario oculta Categorías y/o Marcas para todo el negocio (`configuraciones_negocio.secciones_ocultas`, texto separado por comas; `GET/PUT /api/v1/negocios/:negocioId/configuracion/vistas`, PUT solo propietario `403 SOLO_PROPIETARIOS`, sección no permitida `400 SECCION_NO_VALIDA`). Oculta = sin menú, ruta redirige a Inicio, sin campo en formulario, filtro ni columna en Productos y sin columna Categoría en «Agregar del catálogo». Es preferencia visual: no borra asociaciones y la API sigue aceptando los campos. Negocio nuevo: todo visible.
- Inventario inicial y ajustes pertenecen a sucursal autorizada. Las entradas por recepción se describen en [compras](modulo-compras.md).

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `productos`, `producto_negocio` | Datos del artículo, vínculo comercial, SKU y valores del negocio. |
| `marcas`, `categorias`, `producto_categorias` | Clasificación y asociaciones de categorías. |
| `unidades_medida`, `producto_unidades` | Unidades, empaques y conversiones del producto. Un negocio nuevo trae Pieza (1), Caja 12 y Caja 24 como `EMPAQUE` ([34](historial/34-compras-panel-lateral-y-semilla.md)). |
| `proveedores`, `producto_proveedor` | Proveedores del negocio y asociación de suministro. Un producto puede tener varios proveedores y como máximo un principal ([27](historial/27-compras-cualquier-producto.md)). |
| `producto_codigos`, `producto_imagenes` | Códigos y recursos visuales del artículo. |
| `familias_producto`, `producto_variantes`, `producto_variante_atributos` | Agrupación de variantes y valores de sus atributos. |
| `importacion_productos` | Estado, progreso y resultado de la importación. |
| `configuraciones_negocio.secciones_ocultas` | Secciones ocultas del negocio (`categorias`, `marcas`). |
| `inventario_sucursal`, `movimientos_inventario`, `lotes` | Existencias, trazabilidad y caducidad por sucursal. |
| `impuestos`, `producto_impuestos` | Catálogo de impuestos y vínculos del artículo. |

## Dónde cambiar y qué revisar

- Backend: [modelos](../backend/internal/domain/catalog_models.go), [aplicación](../backend/internal/application/) y [repositorio de productos](../backend/internal/infrastructure/product_repository.go).
- Frontend: [Productos](../frontend/src/app/features/productos/), [Catálogo](../frontend/src/app/features/catalogo/) y [Proveedores](../frontend/src/app/features/proveedores/).
- Comprobar duplicados, códigos vacíos, variantes, autorización de catálogos nuevos y aislamiento de negocio/sucursal; pendientes en [planes actuales](planes-actuales.md).
- Antecedentes: [productos](historial/04-productos-importacion-variantes.md), [proveedor](historial/09-productos-proveedores.md), [código opcional](historial/16-productos-codigo-barras-opcional.md), [catálogos nuevos](historial/17-catalogos-nuevos-importacion-productos.md) [variantes repetidas](historial/18-variantes-importacion-productos-repetidos.md) y [correcciones a la carga masiva](historial/24-correccion-masiva-producto.md).

Aislamiento y totales verificados con PostgreSQL, API y navegador: [20 · Arreglos de invitaciones](historial/20-arreglos-invitaciones.md).
