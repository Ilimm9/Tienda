# Capability 34: Registrar compra con ventanas para crear, datos reordenados y datos semilla

- Estado: **Verificada** (pendiente de prueba manual y aceptación final).
- Aprobación explícita: 2026-10-23 («si la tienes claro dale al plan»). Decisiones: incluye «Proveedor general» (desactivable); unidades Pieza (1), Caja 12 y Caja 24; solo negocios nuevos, porque los datos actuales son de prueba y el usuario recreará la base (`docker compose down -v`). **Todo es visual**: no se modifica el cálculo de subtotal, impuestos y total ni el envío de la compra (lógica existente de Ilian).
- Fecha: 2026-10-23.
- Diseño: canvas de diseño, fila «Compras»: **K1 · Registrar compra: ventanas para crear** (interactivo) y **K2 · Compras vacías: guía y datos semilla**.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Frontend `features/compras`; backend del catálogo (raíz de `application`/`infrastructure`, donde viven unidades y proveedores) y alta de negocio (`negocio`). Base de conocimiento afectada: [modulo-compras.md](../modulo-compras.md) y [registro-productos.md](../registro-productos.md).

## Problema

1. Crear proveedor, producto o unidad desde Registrar compra abre un formulario incrustado dentro del renglón o de la tarjeta: empuja el contenido, no tiene cerrar y se confunde con los datos de la compra.
2. Un negocio nuevo no tiene unidades `EMPAQUE`: el selector «Unidad de empaque» sale vacío y no hay forma evidente de comprar 8 piezas sueltas. Las unidades que crea la carga masiva son `CANTIDAD` y Compras no las muestra.
3. La pantalla vacía de Compras no orienta.

## Recomendación

- **Ventana modal para crear** (decisión del usuario, 2026-10-23): el panel lateral ocupa todo el alto y queda vacío con formularios de 1 a 4 campos. El modal se ajusta a su contenido.
- **Semilla mínima y real, no datos de ejemplo.** Sembrar solo catálogos que el negocio va a usar tal cual: formas de compra y un proveedor genérico. No sembrar productos ni compras ficticias: se mezclarían con inventario y costos reales y habría que borrarlos.
- **Guía en el estado vacío** que explique los 3 pasos y muestre lo que ya viene listo.

## Alcance y reglas

### 1. Ventanas modales para crear (frontend)
- Un único componente de ventana en `features/compras` (sobre `p-dialog` de PrimeNG, ya instalado, o el patrón de diálogo de la casa) con tres variantes y **los mismos campos de hoy**:
  - **Nuevo proveedor**: nombre.
  - **Nuevo producto**: nombre y código de barras (opcional), con el aviso «se crea sin precio de venta».
  - **Nueva unidad de empaque**: nombre, símbolo y piezas por empaque, más atajos Pieza·1, Caja 6, Caja 12 y Caja 24 que solo prellenan esos tres campos.
- Aspecto: centrada, ancho máximo 520 px, alto según contenido (con desplazamiento interno si no cabe), radio 18 px, sombra `--shadow-dialog`, fondo oscurecido. Encabezado con icono, título, subtítulo y X; pie con «Cancelar» y «Guardar y usar». En teléfono ocupa el ancho con márgenes de 16 px.
- Se abre con el botón «Nuevo…» junto a cada selector (proveedor, producto, unidad de empaque), en lugar de la opción «Crear nuevo…» dentro de la lista.
- Cierra con la X, «Cancelar», clic en el fondo o `Esc`; devuelve el foco al botón que la abrió; `role="dialog"` y `aria-modal`. Mientras está abierta, la compra no se puede enviar.
- «Guardar y usar» crea el registro con los endpoints actuales y lo deja elegido en el selector de origen. Los errores del servidor se muestran dentro de la ventana.
- Se eliminan las secciones incrustadas `purchase-provider-form`, `purchase-product-form` y `purchase-unit-form`; sus campos pasan a la ventana sin perder ninguno.
- Permisos sin cambio: proveedor y unidad exigen `catalogo.gestionar`; producto nuevo sigue disponible como hoy (regla de la 31).

### 2. Datos de la compra y renglones (frontend)
- **No se quita ningún campo** del formulario actual; todos se siguen enviando igual.
- «¿A quién le compraste?», visibles: proveedor, sucursal receptora, tipo de documento (Factura, Ticket), folio, **fecha del documento** (hoy por omisión, como ya ocurre; ayuda: «La que trae impresa la factura o el ticket»), **fecha de recepción** (hoy por omisión; ayuda: «Cuando la mercancía llegó a la sucursal») y **forma de pago**.
- **Número de cuenta** aparece solo con forma de pago Transferencia (ayuda «CLABE o cuenta de origen») o Tarjeta («Últimos 4 dígitos»). Al cambiar a otra forma de pago se oculta y se limpia para no enviar un valor que ya no aplica.
- Observaciones queda en una sección plegable opcional.
- Renglón: se conservan las etiquetas actuales **Unidad de empaque**, **Empaques**, **Costo por empaque** y **Descuento**. Cada unidad muestra su factor: «Caja 12 (12 piezas)». Renglón nuevo con «Pieza» preseleccionada si existe.
- Debajo del renglón: «Entran N piezas a <sucursal> · subtotal $X».
- Lote, caducidad y anomalías (dañadas, faltantes) siguen como opcionales en una sección plegable.
- Sin cambios de contrato: `unidad_medida_id`, `cantidad_empaques`, `costo_empaque`, `numero_cuenta`, `fecha_documento`, etc.

### 3. Datos semilla (backend)
- Al crear un negocio, en la misma transacción que el rol y el puesto Cajero:
  - Unidades `EMPAQUE`, enteras, activas: **Pieza** (`pz`, factor 1), **Caja 12** (`cj12`, 12), **Caja 24** (`cj24`, 24).
  - Proveedor **Proveedor general** activo.
- Negocios existentes: sin migración (decisión del 2026-10-23; la base se recrea). Las unidades usan `ON CONFLICT DO NOTHING`: si un nombre o símbolo ya existe en el negocio, esa fila se omite.
- Son datos editables: se pueden renombrar, desactivar o borrar y no se vuelven a sembrar.
- Integración entre dominios: `application/negocio` define la interfaz `SembradorCatalogo`; la implementa `infrastructure.CatalogSeeder` y se conecta en `cmd/api/main.go` con `NegocioService.ConSembrador`. Se llama justo después de crear el negocio, en su propia transacción: si falla, se registra en el log y el alta no se revierte ni devuelve error (evita que un reintento duplique el negocio).

### 4. Compras vacías con guía (frontend)
- Sin compras: tarjeta «Tu primera compra toma un minuto» con 3 pasos y botón «Registrar primera compra» (solo con `compras.registrar`).
- Tarjeta «Ya viene listo» con las formas de compra y el proveedor sembrados, leídos de los catálogos reales (no fijos en la plantilla); se oculta si el negocio ya no tiene ninguno.

## Fuera de alcance

- Convertir a `EMPAQUE` las unidades `CANTIDAD` que crea la carga masiva (pendiente aparte).
- Ligar a cada producto su forma de compra habitual.
- Regresar automáticamente al formulario tras crear desde Equipo (capability 33).

## Pruebas y criterios de aceptación

- Backend:
  - Integración: crear negocio deja las 3 unidades `EMPAQUE` y «Proveedor general».
  - Semilla: una segunda corrida no falla ni duplica unidades.
- Frontend (`compras.component.spec.ts` y del panel):
  - La ventana abre y cierra (X, Cancelar, fondo, `Esc`), regresa el foco y no hay formularios incrustados.
  - «Guardar y usar» elige el registro nuevo; los errores aparecen dentro de la ventana.
  - Número de cuenta aparece solo con Transferencia o Tarjeta y se limpia al cambiar.
  - Pieza viene preseleccionada y el resumen «Entran N piezas» es correcto (8 × Pieza = 8; 2 × Caja 12 = 24).
  - El cuerpo enviado conserva todos los campos actuales.
  - El estado vacío muestra la guía y lo sembrado.
- `go build`, `go vet`, `go test ./...`, `ng test`, `ng build`. Prueba manual de 360 px a 1440 px.

## Decisiones

1. Modal, no panel lateral (2026-10-23).
2. Semilla con «Proveedor general», desactivable (2026-10-23).
3. Unidades Pieza (1), Caja 12 y Caja 24 (2026-10-23).
4. Solo negocios nuevos; sin migración (2026-10-23).
5. Nuevo producto no llama a la API: como hoy, la ventana solo llena `producto_nuevo`, `nombre_producto_nuevo` y `codigo_barras_nuevo` del renglón y el producto se crea al guardar la compra.
6. La ventana usa `<dialog>` nativo con `showModal()` (Esc, foco y fondo inerte del navegador) y el estilo `.dialogo` de la casa; sin dependencias nuevas.

## Tareas

- [x] Aprobación del usuario.
- [x] Backend: semilla en alta de negocio + pruebas.
- [x] Frontend: ventanas modales, datos reordenados con número de cuenta condicional, valores por defecto, estado vacío + pruebas.
- [x] Actualizar `modulo-compras.md`, `registro-productos.md` y `PLAN_IMPLEMENTACION.md`.
- [x] Verificación.

## Verificación

- 2026-10-23: sin cambios en `subtotal`, `ieps`, `iva`, `taxes`, `total`, `pieces`, `detailSubtotal` ni `submit` (cuerpo enviado igual). Se quitaron solo las opciones «Crear nuevo…» de los selectores, sustituidas por botones «Nuevo…».
- 2026-10-23: `go build`, `go vet` correctos; `go test ./...` pasa salvo `TestProductionAcceptsStrongOTPSecret` (falla previa registrada). Prueba unitaria: la semilla recibe el negocio creado y su falla no rompe el alta.
- 2026-10-23: en base temporal `tienda_semilla_test` (migrada con `database.Init`, borrada al terminar): `TestCrearNegocioSiembraRolCajero` y `TestCatalogSeederSiembraUnidadesYProveedor` pasan (3 unidades `EMPAQUE` con factores 1/12/24, «Proveedor general» activo; segunda corrida sin errores ni duplicados).
- 2026-10-23: `ng test` completo: 275 de 281; las 6 fallas son de `alta-inicial.spec.ts`, ya registradas como previas en [planes actuales](../planes-actuales.md). `features/compras`: 15 de 15 (ventana de producto sin formulario incrustado, cierre por fondo, Pieza preseleccionada, etiquetas con factor, «Entran 8 piezas», 2 × Caja 12 = 24, número de cuenta condicional y limpieza). `ng build` correcto; `compras.component.css` rebasa el presupuesto de 4 kB (advertencia).
- Pendiente: prueba manual en navegador de 360 a 1440 px tras recrear la base (`docker compose down -v` y reconstruir) y aceptación final.

## Historial de revisión

- 2026-10-23: el usuario pide forma de pago en la sección principal y número de cuenta solo cuando aplique; conservar las etiquetas Unidad de empaque / Empaques / Costo por empaque / Descuento; modales en lugar de panel lateral; no quitar campos. Diseño K1 y plan actualizados.
