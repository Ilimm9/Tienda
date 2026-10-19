# Capability 12: Validación de costos y precios por sucursal

## Estado

Aprobada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-09-30 mediante la instrucción: “Implement the plan”.

La ampliación para validar costos por pieza, ganancia y presentaciones vendibles fue aprobada explícitamente el 2026-09-30 mediante la instrucción: “Implement the plan”.

La recuperación de valores desde la recepción de compra fue aprobada explícitamente el 2026-09-30 mediante la instrucción: “Implement the plan”.

## Alcance

- La recepción ingresa existencias y crea una propuesta pendiente por renglón.
- El administrador de la sucursal edita y autoriza costo y precio de venta desde un apartado independiente.
- Se conserva historial de valores capturados y autorizados por producto y sucursal.
- La validación muestra y autoriza el costo final por pieza; muestra ganancia por pieza y ganancia total.
- El costo por pieza procede del ticket de compra y solo se consulta en esta pantalla; no es editable durante la autorización.
- Las propuestas pendientes recuperan piezas, impuestos y costos desde la compra original, y el costo vigente desde la sucursal.
- El precio de venta actual prioriza el precio autorizado de sucursal y, si no existe, recupera el precio vigente del producto en catálogo.

## Reglas y contratos

- La propuesta no modifica el costo ni el precio vigente hasta ser autorizada.
- El precio sugerido usa costo final por pieza con impuestos, margen de sucursal o excepción por producto y redondeo al peso superior.
- No hay rechazo: una propuesta pendiente se edita y se autoriza por renglón.
- El apartado está disponible para toda sesión autenticada del negocio activo; no depende de permisos mientras se administra el esquema de roles.

## Pruebas y criterios de aceptación

- Una recepción aumenta inventario y crea propuestas pendientes.
- La autorización deja auditoría del valor original y del valor autorizado.
- Las propuestas y precios quedan aislados por sucursal.
- El administrador puede ajustar costo, margen y precio antes de autorizar.

## Tareas y resultados

- [x] Crear propuestas pendientes durante la recepción, sin bloquear el inventario.
- [x] Persistir precio y costo vigentes por producto y sucursal, con auditoría de la autorización.
- [x] Mantener el apartado separado de Compras sin restricción adicional por permisos o asignación.
- [x] Añadir la bandeja independiente de validación y navegación.
- [x] Recuperar piezas y costo final con impuestos desde el detalle de compra, y costo/precio vigentes desde la sucursal.
- [x] Normalizar importes entregados por la API y mostrar controles de edición visibles con el estilo de Prime.
- [x] Mantener como solo lectura el costo por pieza recuperado del ticket.
- [x] Mostrar el precio de venta actual del producto como referencia de la validación.
- [x] Mostrar el costo al público por pieza calculado junto a la ganancia.
- [x] Ejecutar `go test` de los paquetes nuevos, compilación TypeScript, revisión de espacios y actualización del grafo.

## Verificación

- `go test ./internal/infrastructure/precio ./internal/infrastructure/compra ./internal/application/precio ./internal/interfaces/http/precio`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; el grafo se reconstruyó con 2155 nodos y 6017 relaciones.
- Pendiente conocido: `go test ./...` conserva un fallo preexistente en `config.TestProductionAcceptsStrongOTPSecret`, ajeno a esta capability.
