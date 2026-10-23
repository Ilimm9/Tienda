# Código de barras opcional al crear producto desde Registrar compra

- Estado: **Cerrada** el 2026-10-23 con aceptación final del usuario.
- Aprobación explícita: 2026-10-21, el usuario aprobó el plan.
- Fecha: 2026-10-21.
- Organización conforme a [AGENTS.md](../../AGENTS.md). Dominio `compra` (backend) y feature `compras` (frontend); sin cambios de modelo ni de API.

## Objetivo

Al registrar una compra y elegir «Crear nuevo…», poder dar de alta el producto solo con su nombre cuando en ese momento no se tiene el código de barras. El código se agrega después desde Editar producto, igual que en Productos ([16](16-productos-codigo-barras-opcional.md)).

## Reglas

- El producto nuevo desde Compras requiere nombre; el código de barras es opcional.
- Con código: se conserva la validación de duplicado por negocio («el código de barras ya está registrado») y se crea `producto_codigos` como principal (`GTIN`).
- Sin código: no se crea `producto_codigos` y el detalle de compra guarda `codigo_barras` vacío (`NULL`).
- El resto del alta no cambia: `productos`, `producto_negocio` con SKU `COMP-…` y vínculo con el proveedor de la compra.

## Fuera de alcance

- Detección de productos duplicados por nombre.
- Pedir el código de forma automática después.

## Archivos

- `backend/internal/infrastructure/compra/compra_repository.go`: validación del detalle y `createNewProduct`.
- `frontend/src/app/features/compras/compras.component.ts`: sin `Validators.required` en `codigo_barras_nuevo`.
- `frontend/src/app/features/compras/compras.component.html`: etiqueta «Código de barras (opcional)».
- `frontend/src/app/features/compras/compras.component.spec.ts`

## Pruebas ejecutadas (2026-10-21)

- `go build ./...`, `go vet ./internal/infrastructure/compra/` y `gofmt`: correctos.
- `go test ./internal/...`: correcto salvo `internal/config` `TestProductionAcceptsStrongOTPSecret`, falla previa ya registrada.
- `ng test --include src/app/features/compras` (Node 22.23.2): 5/5 correctas. La prueba ajustada verifica que el nombre es obligatorio y el código opcional.
- `ng build`: correcto.

## Pendientes

- No existen pruebas de repositorio para `compra` con base de datos; el alta sin código no se cubrió con prueba automática de backend.
- Prueba manual: registrar una compra con un producto nuevo sin código. Debe guardarse, aparecer en Productos sin código y aceptar el código después desde Editar producto. Con un código ya existente debe seguir rechazándose.
- Aceptación final del usuario.

## Prueba manual (2026-10-22)

- El usuario confirmó que las pruebas manuales pasaron. Pendiente la aceptación final.

## Cierre (2026-10-23)

- Aceptación final del usuario: «acepto la 22, 23, 27, 28, 29 y 31». Cerrada y archivada en `historial/`.
