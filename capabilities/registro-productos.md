# Registro de productos y catálogos

Base vigente según código local. Actualizado: 2026-10-19.

## Flujo y reglas

- Registrar productos simples o variantes dentro del negocio activo. `productos` describe el artículo; `producto_negocio` guarda su oferta comercial y SKU; existencias se separan por sucursal.
- Marca, categoría, unidad y proveedor son catálogos asociados. Catálogo permite gestionarlos; Proveedores tiene sección propia. Validar referencias del mismo negocio y dependencias antes de borrar.
- La importación XLSX primero valida y muestra errores, advertencias y catálogos nuevos; después procesa una tarea con progreso. SKU vacío se genera por negocio. La columna de código de barras sigue obligatoria, pero sus celdas pueden estar vacías.
- Marcas, categorías, unidades y proveedores vacíos no crean asociaciones. Si aparecen nombres nuevos, exigir confirmación global `autorizar_catalogos_nuevos=true`; sin ella no persistir la importación que los requiere.
- Un producto simple se identifica por nombre y presentación normalizados. Otro código para esa identidad exige `Variantes` con `Atributo: valor`; cada variante conserva SKU, código, precio e inventario propios. Códigos repetidos conservan su error independiente.
- Inventario inicial y ajustes pertenecen a sucursal autorizada. Las entradas por recepción se describen en [compras](modulo-compras.md).

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `productos`, `producto_negocio` | Datos del artículo, vínculo comercial, SKU y valores del negocio. |
| `marcas`, `categorias`, `producto_categorias` | Clasificación y asociaciones de categorías. |
| `unidades_medida`, `producto_unidades` | Unidades, empaques y conversiones del producto. |
| `proveedores`, `producto_proveedor` | Proveedores del negocio y asociación de suministro. |
| `producto_codigos`, `producto_imagenes` | Códigos y recursos visuales del artículo. |
| `familias_producto`, `producto_variantes`, `producto_variante_atributos` | Agrupación de variantes y valores de sus atributos. |
| `importacion_productos` | Estado, progreso y resultado de la importación. |
| `inventario_sucursal`, `movimientos_inventario`, `lotes` | Existencias, trazabilidad y caducidad por sucursal. |
| `impuestos`, `producto_impuestos` | Catálogo de impuestos y vínculos del artículo. |

## Dónde cambiar y qué revisar

- Backend: [modelos](../backend/internal/domain/catalog_models.go), [aplicación](../backend/internal/application/) y [repositorio de productos](../backend/internal/infrastructure/product_repository.go).
- Frontend: [Productos](../frontend/src/app/features/productos/), [Catálogo](../frontend/src/app/features/catalogo/) y [Proveedores](../frontend/src/app/features/proveedores/).
- Comprobar duplicados, códigos vacíos, variantes, autorización de catálogos nuevos y aislamiento de negocio/sucursal; pendientes en [planes actuales](planes-actuales.md).
- Antecedentes: [productos](historial/04-productos-importacion-variantes.md), [proveedor](historial/09-productos-proveedores.md), [código opcional](historial/16-productos-codigo-barras-opcional.md), [catálogos nuevos](historial/17-catalogos-nuevos-importacion-productos.md) y [variantes repetidas](historial/18-variantes-importacion-productos-repetidos.md).
