# Ajuste: carga por lista en una sola transacción

Creado e implementado: 2026-11-11. Pendiente: prueba manual del usuario.

> **Cambio respecto al plan:** al probar contra PostgreSQL apareció el índice único `idx_categorias_negocio_nombre_ci` (negocio + nombre). El usuario eligió respetarlo: **el nombre de categoría es único en todo el negocio** (no por padre). Las secciones de abajo que hablan de «único dentro del mismo padre» quedan sustituidas por esta regla; las rutas `A > B` siguen sirviendo para ubicar y crear ancestros.

## Contexto

La carga masiva de marcas, categorías, unidades y proveedores acepta pegar una lista (un registro por renglón). Hoy el frontend (`frontend/src/app/features/catalogo/catalogo.component.ts`, `agregarLista()`) crea renglón por renglón con el `POST` individual:

- N requests por lista (más recargas de categorías para resolver padres).
- Sin transacción: si algo falla a la mitad, lo creado se queda.
- Validaciones distintas en el navegador, el alta individual y el XLSX.

El backend ya tiene importadores transaccionales por catálogo (`ProductRepository.ImportBrands/ImportCategories/ImportUnits/ImportProviders` en `backend/internal/infrastructure/product_repository.go`), usados por el XLSX. Este ajuste los reutiliza y unifica las reglas.

## Alcance

- Endpoint JSON por catálogo que recibe la lista y usa el mismo importador del XLSX.
- `agregarLista()` envía una sola petición y muestra un reporte final.
- Categorías: nombre único **dentro del mismo padre** en todas las vías (individual, XLSX y lista).
- Validación en backend de RFC (proveedores) y tipo (unidades) en todas las vías.
- Sin cambios de esquema (sin índices nuevos: la regla se valida en código para no chocar con duplicados que ya existan en producción).

Fuera de alcance: productos, edición masiva, borrado masivo.

## Reglas

1. **Omitir y seguir.** Renglones inválidos, repetidos o ya existentes se omiten y el resto se crea. Todo lo creado va en una sola transacción: si falla la base de datos, no se guarda nada.
2. **Reporte final.** Se muestra cuántos se crearon, cuántos se omitieron y, por renglón omitido, su número, nombre y motivo. Los omitidos por error se pueden devolver a la caja para corregirlos.
3. **Comparación de nombres:** sin mayúsculas, acentos ni espacios dobles, igual en navegador y backend (`catalogImportKey` agrega el colapso de espacios).
4. **Categorías, únicas por padre.**
   - `Bebidas > Frías` y `Lácteos > Frías` pueden convivir; dos `Frías` bajo `Bebidas`, no.
   - Aplica en el alta individual, en la edición (renombrar o mover a otro padre) y en ambos importadores.
   - Los duplicados que ya existan en producción no se tocan; la regla solo aplica a altas y ediciones nuevas.
5. **Categorías con ruta.**
   - En la lista, cada renglón manda su ruta completa (`["Bebidas", "Frías", "Agua"]`). El backend recorre la ruta: usa el ancestro existente bajo el padre correcto o lo crea en la misma transacción. Un ancestro repetido en varias rutas se crea una sola vez.
   - En el XLSX, la columna `Categoría padre` acepta un nombre o una ruta `A > B`. Si un nombre simple coincide con varias categorías, el renglón se rechaza con «Hay varias categorías con ese nombre; escribe la ruta completa (Padre > Hija)».
6. **RFC (proveedores):** opcional; si viene, se pasa a mayúsculas y debe cumplir `^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$` (mismo patrón que negocio). Aplica en individual (alta y edición), XLSX y lista.
7. **Tipo de unidad:** `PESO`, `VOLUMEN`, `LONGITUD`, `AREA`, `CANTIDAD` o `EMPAQUE`. El XLSX y la lista aceptan también `peso`, `Área`, etc. Aplica en individual (alta y edición), XLSX y lista; las unidades existentes con otro tipo no se tocan.
8. **Unidades repetidas:** por nombre **o** símbolo (regla actual del importador; el navegador la adopta).
9. Mismo permiso que la importación XLSX (grupo `catalogoGestion`). Límite de 1000 renglones y 1 MB por petición.

## Contrato

```
POST /catalogo/{marcas|categorias|unidades|proveedores}/lista
```

Mismo prefijo y middleware que `/catalogo/{seccion}/importar`.

Body (`fila` = número de renglón en la caja):

```jsonc
// marcas
{ "renglones": [{ "fila": 1, "nombre": "Bimbo" }] }
// categorias
{ "renglones": [{ "fila": 1, "ruta": ["Bebidas", "Frías"] }] }
// unidades
{ "renglones": [{ "fila": 1, "nombre": "Kilogramo", "simbolo": "kg", "tipo": "PESO", "factor_a_base": 1000, "decimales": 3 }] }
// proveedores
{ "renglones": [{ "fila": 1, "nombre": "Sigma", "rfc": "SIG010101AB1", "telefono": "5512345678", "email": "ventas@sigma.mx" }] }
```

Respuesta `200`: `domain.CatalogImportResult` (`procesadas`, `creadas`, `omitidas`, `invalidas`, `errores[{fila, campo, motivo}]`). En categorías, `creadas` incluye los ancestros creados.

Errores: `400` (JSON inválido, lista vacía o más de 1000 renglones), `403` (permiso) y `500` (falla de base de datos, sin guardar nada).

## Tareas

### Backend

1. `application/catalog_import.go`:
   - Separar la validación por renglón de la lectura del XLSX (`validateBrandRow`, `validateCategoryRow`, `validateUnitRow`, `validateProviderRow`), con RFC y tipo (reglas 6 y 7).
   - `parse*Import` y la lista usan estas funciones.
2. `domain`: `CatalogImportCategoryRow` agrega `Ruta []string` (struct en memoria, no tabla). `CategoriaPadre` se sigue leyendo en el XLSX y se convierte a ruta.
3. `ProductRepository.ImportCategories`: reescribir la resolución por llave `(padre, nombre)` recorriendo rutas; crear ancestros faltantes; resolver un padre ambiguo del XLSX con el error de la regla 5.
4. `ProductRepository.CreateCategory/UpdateCategory`: rechazar nombre repetido bajo el mismo padre (`409` con mensaje claro).
5. `CreateUnit/UpdateUnit` y `CreateProvider/UpdateProvider`: validar tipo y RFC.
6. `catalogImportKey`: colapsar espacios dobles.
7. `ProductService.Import*List(businessID, rows)`, el handler `importCatalogList(c, section)` y las 4 rutas `/lista` en `cmd/api/main.go`.

### Frontend

8. `catalogo.service.ts`: `importarLista(negocioId, seccion, renglones)`.
9. `lista-pegada.ts`:
   - Categorías por ruta completa (ya lo hace) y comparación por `(padre, nombre)`.
   - Unidades por nombre o símbolo.
   - Exponer los renglones que se envían.
10. `catalogo.component.ts`: `agregarLista()` envía una sola petición y guarda el resultado.
11. `catalogo.component.html/.css`:
    - Reporte final en el modal: creadas, omitidas y tabla de omitidos con renglón, nombre y motivo.
    - Botón «Corregir omitidos» que devuelve a la caja los renglones con error.
    - Toast de resumen.
12. Formularios individuales: mostrar los nuevos mensajes (`409` de categoría repetida, RFC y tipo).

### Documentación

13. Actualizar `capabilities/historial/registro-productos.md` (carga masiva y regla de categorías).

## Pruebas

- **Backend unit:**
  - Validadores por renglón (vacíos, factor ≤ 0, decimales, correo, RFC y tipo).
  - `catalogImportKey` con espacios dobles.
- **Backend integración (base temporal):**
  - Lista con existentes, repetidos e inválidos: crea solo los válidos y el reporte trae cada omitido.
  - Categorías `A > B > C` con `A` existente: crea `B` y `C` con los padres correctos.
  - `Bebidas > Frías` y `Lácteos > Frías`: crea ambas.
  - XLSX con padre ambiguo: el renglón se rechaza; con ruta `A > B` como padre, se crea.
  - Alta y edición individual de una categoría repetida bajo el mismo padre: `409`.
  - Error forzado de base de datos: no queda nada (rollback).
- **Handler:** `400` con lista vacía, `400` con más de 1000 renglones y `403` sin permiso.
- **Frontend spec:**
  - Unidad repetida por símbolo.
  - Categoría igual bajo otro padre aceptada.
  - `agregarLista()` hace una sola llamada y pinta el reporte.
- `go test ./...`, `ng test` y `ng build`.

## Criterios de aceptación

- Pegar 200 marcas hace un solo request y crea solo las nuevas.
- El modal muestra el reporte con los omitidos y sus motivos; «Corregir omitidos» devuelve solo los renglones con error.
- Si la base de datos falla, no se crea nada y se avisa.
- Las mismas reglas (categoría por padre, RFC, tipo) aplican en individual, XLSX y lista.

## Decisiones (usuario, 2026-11-11)

- Omitir y seguir, con reporte final.
- ~~Categorías: nombre único dentro del mismo padre (opción B).~~ Sustituida: nombre único en el negocio, por el índice existente (usuario, 2026-11-11).
- Validar RFC y tipo de unidad en el backend.
- Planes sin número, con prefijo `thrs-`; `capabilities/` es una guía opcional.

## Resultado de la implementación

- Backend: `domain/catalog_rules.go` (RFC, tipo, longitudes, `ErrCategoriaDuplicada`), `application/catalog_list.go` (validadores compartidos XLSX/lista y `Import*List`), `infrastructure/catalog_category_import.go` (árbol de categorías, `ImportCategoryPaths`, nuevo `ImportCategories`), handler `importCatalogList` y 4 rutas `/lista`.
- Frontend: `lista-pegada.ts` con las mismas reglas, `importarLista` en el servicio, `agregarLista()` con una sola petición y reporte en la ventana.
- Pruebas: unitarias Go (validadores, servicio, handler), integración contra PostgreSQL temporal (rutas, XLSX por nombre/ruta/orden, nombre único individual, rollback), specs de Angular, `ng build`, recorrido HTTP completo (registro → negocio → 4 listas) y prueba en navegador (vista previa, reporte, «Corregir», aviso de nombre repetido).
- Integración: `LISTA_TEST_DATABASE_URL=<base _test migrada> go test ./internal/infrastructure -run 'TestImportCateg|TestCategoriaIndividual|TestImportLista'`.
