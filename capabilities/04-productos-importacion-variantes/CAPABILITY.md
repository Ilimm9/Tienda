# Resumen de cambios de productos

## Control

- Fecha de creación: 2026-09-17.
- Dependencias: catálogo de productos e importación XLSX existentes.

## Objetivo

Documentar de forma resumida los cambios realizados en el apartado de productos, su importación masiva y el manejo de variantes.

## Cambios 
En este cambio se editan los archivos relacionados a productos; su carga masiva y edición 
- El SKU es opcional: si no se proporciona, se genera automáticamente un identificador consecutivo por negocio.
- La importación valida el archivo completo y reporta todos los errores y advertencias antes de ingresar productos.
- Se identifican SKU y códigos de barras ya registrados, así como los repetidos dentro del mismo archivo.
- La importación acepta códigos de barras numéricos de 1 a 14 dígitos; los de menos de 8 dígitos generan una advertencia.
- Las filas del mismo producto base se agrupan para crear o ampliar una familia de variantes. Cada variante conserva su propio SKU, código de barras, precio e inventario.
- Las familias de variantes usan una clave técnica segura y un código legible consecutivo para su identificación.
- La importación se procesa como una tarea con progreso y resultado; la búsqueda de imágenes se separa de la carga.
- Pueden coexistir productos simples y familias de variantes con el mismo nombre.
