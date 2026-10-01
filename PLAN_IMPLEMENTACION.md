# Plan de implementación

Este archivo registra el orden, dependencia y estado de las capabilities del proyecto Tienda. Ninguna funcionalidad puede implementarse sin especificación aprobada.

## Flujo de trabajo

1. Describir la funcionalidad.
2. Crear `capabilities/NN-nombre/CAPABILITY.md` con estado `Borrador`.
3. Completar alcance, decisiones, contratos, pruebas y criterios de aceptación.
4. Revisar la propuesta con el usuario.
5. Registrar aprobación explícita y cambiar estado a `Aprobada`.
6. Implementar y cambiar estado a `En implementación`.
7. Ejecutar verificaciones y registrar resultados.
8. Cambiar estado a `Verificada` cuando todos los criterios pasen.
9. Cambiar estado a `Cerrada` después de la revisión final del usuario.

## Estados permitidos

- `Borrador`: definición inicial incompleta o no revisada.
- `En revisión`: lista para recibir decisiones o correcciones.
- `Aprobada`: autorizada explícitamente para implementación.
- `En implementación`: trabajo de código activo.
- `Verificada`: implementación terminada y pruebas registradas.
- `Cerrada`: aceptada finalmente por el usuario.

## Registro de capabilities

| ID | Capability | Estado | Dependencias | Aprobación | Notas |
| --- | --- | --- | --- | --- | --- |
| 01 | [Identidad, negocios, sucursales y RBAC](capabilities/01-identidad-negocio-sucursales-rbac/CAPABILITY.md) | En implementación | Autenticación actual | Fases 1, 1.5, 1.6, 2 y 3 aprobadas | Fases 1, 1.5 y 1.6 cerradas; fase 2 verificada y pendiente de aceptación; fase 3 en implementación por autorización explícita con excepción documentada; documentos separados por fase; `base.MD` preservado |
| 02 | [Configuración del entorno local](capabilities/02-configuracion-entorno-local/CAPABILITY.md) | Verificada | Configuración Go y Angular CLI | Aprobada el 2026-09-11 | Implementación verificada; pendiente de aceptación final |
| 03 | [Carga de archivos unificada](capabilities/03-carga-archivos-unificada/CAPABILITY.md) | En implementación | Carga masiva existente de productos y catálogo | Aprobada explícitamente por el usuario el 2026-09-14 | Estado visual inmediato, eliminación y arrastrar/soltar |
| 04 | [Resumen de cambios de productos](capabilities/04-productos-importacion-variantes/CAPABILITY.md) | — | Catálogo de productos e importación XLSX existentes | — | Resumen consolidado de los cambios antes documentados en las capabilities 04–14 |
| 05 | [Tailwind y sistema visual](capabilities/05-tailwind-sistema-visual/CAPABILITY.md) | En revisión | Tema, layout y estilos actuales del frontend | Pendiente | Integración incremental; piloto en shell e inicio |
| 06 | [Preparación segura para producción en AWS Lightsail](capabilities/06-preparacion-produccion-segura/CAPABILITY.md) | En implementación | Autorización, configuración, contenedores, datos e infraestructura actuales | Fases 1, 2 y 3 aprobadas el 2026-09-20 | Fases 1, 2 y 3 implementadas y verificadas; fase 3 usa Mailpit local y no SES; base legacy `5433`, fases 4–7 y despliegue pendientes |
| 08 | [Proveedores como sección independiente](capabilities/08-proveedores-seccion-independiente/CAPABILITY.md) | En implementación | Navegación, rutas y pantalla de catálogo actuales | Aprobada explícitamente el 2026-09-22 | Proveedores se separa de Catálogo, usa `/proveedores` y elimina la ruta anterior |
| 09 | [Proveedor opcional en productos](capabilities/09-productos-proveedores/CAPABILITY.md) | En implementación | Productos, proveedores e importación XLSX | Aprobada explícitamente el 2026-09-23 | Un proveedor opcional por producto, incluidos filtros, formulario y carga masiva; pendiente de pruebas Angular con Node compatible |
| 10 | [Compras y recepción de mercancía](capabilities/10-compras-proveedores/CAPABILITY.md) | En implementación | Proveedores, productos, unidades e inventario | Revisiones aprobadas explícitamente el 2026-09-28 | Mercancía antes de impuestos y pago; total e importe pagado automáticos |
| 11 | [Fechas de auditoría y presentación en Ciudad de México](capabilities/10-fechas-auditoria/CAPABILITY.md) | En implementación | Timestamps UTC expuestos por la API y vistas Angular | Aprobada explícitamente el 2026-09-23 | UTC desde Go para auditoría y presentación fija en `America/Mexico_City`; pendiente ejecutar Angular con Node compatible |
| 12 | [Validación de costos y precios por sucursal](capabilities/12-validacion-costos-precios-sucursal/CAPABILITY.md) | En implementación | Compras, productos, sucursales y RBAC | Aprobada explícitamente el 2026-09-30 | Recepción crea propuestas; administrador de sucursal edita y autoriza por renglón |

## Restricciones vigentes

- Git solo lectura hasta confirmación final del usuario.
- Sin commits, push, despliegues ni publicaciones.
- Código real permanece en `backend/` y `frontend/`.
- Cualquier desviación de una capability aprobada detiene esa parte del trabajo hasta revisión.
