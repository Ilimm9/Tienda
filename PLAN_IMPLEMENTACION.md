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
| 05 | [Tailwind y sistema visual](capabilities/05-tailwind-sistema-visual/CAPABILITY.md) | Verificada | Tema, layout y estilos actuales del frontend | Aprobada explícitamente el 2026-10-02 con alcance ajustado | Tailwind 4 sin Preflight con tokens G; sección Equipo migrada |
| 06 | [Preparación segura para producción en AWS Lightsail](capabilities/06-preparacion-produccion-segura/CAPABILITY.md) | En implementación | Autorización, configuración, contenedores, datos e infraestructura actuales | Fases 1, 2 y 3 y demo aprobadas; revisión Resend en revisión | Demo publicada por el usuario; propuesta Resend Free + compatibilidad SES documentada en la 06; contratos de correo canónicos en la 07; respaldos y gates generales pendientes |
| 08 | [Proveedores como sección independiente](capabilities/08-proveedores-seccion-independiente/CAPABILITY.md) | En implementación | Navegación, rutas y pantalla de catálogo actuales | Aprobada explícitamente el 2026-09-22 | Proveedores se separa de Catálogo, usa `/proveedores` y elimina la ruta anterior |
| 09 | [Proveedor opcional en productos](capabilities/09-productos-proveedores/CAPABILITY.md) | En implementación | Productos, proveedores e importación XLSX | Aprobada explícitamente el 2026-09-23 | Un proveedor opcional por producto, incluidos filtros, formulario y carga masiva; pendiente de pruebas Angular con Node compatible |
| 10 | [Compras y recepción de mercancía](capabilities/10-compras-proveedores/CAPABILITY.md) | En implementación | Proveedores, productos, unidades e inventario | Revisiones aprobadas explícitamente el 2026-09-28 | Mercancía antes de impuestos y pago; total e importe pagado automáticos |
| 11 | [Fechas de auditoría y presentación en Ciudad de México](capabilities/10-fechas-auditoria/CAPABILITY.md) | En implementación | Timestamps UTC expuestos por la API y vistas Angular | Aprobada explícitamente el 2026-09-23 | UTC desde Go para auditoría y presentación fija en `America/Mexico_City`; pendiente ejecutar Angular con Node compatible |
| 12 | [Validación de costos y precios por sucursal](capabilities/12-validacion-costos-precios-sucursal/CAPABILITY.md) | En implementación | Compras, productos, sucursales y RBAC | Aprobada explícitamente el 2026-09-30 | Recepción crea propuestas; administrador de sucursal edita y autoriza por renglón |
| 13 | [Acceso animado Stockion](capabilities/13-acceso-animado-stockion/CAPABILITY.md) | Verificada | Rutas y formularios de autenticación actuales | Aprobada explícitamente el 2026-10-01 | Shell compartido y mosaico animado con GSAP FLIP |

## Borradores solicitados en `capabilities/thrs`

Ubicación directa autorizada por el usuario el 2026-10-01 para planeación, sin carpetas adicionales ni implementación. Antes de implementar deben contar con capability aprobada conforme a `AGENTS.md`; estos borradores no autorizan cambios de código.

| Orden | Plan | Estado | Dependencias |
| --- | --- | --- | --- |
| 1 | [Alta inicial de empresa y sucursal](capabilities/thrs/01-alta-inicial-empresa-sucursal.md) | Borrador | Identidad/contexto 01, seguridad 06 y acceso visual 13 |
| 2 | [Invitaciones de empleados por sucursal](capabilities/thrs/02-invitaciones-empleados-sucursal.md) | Aprobada el 2026-10-01; flujo de invitación implementado | Plan 1, empleados/RBAC/invitaciones 01 y correo 07 |
| 3 | [Paleta clara y tipografía global](capabilities/thrs/03-paleta-tipografia-global.md) | Aprobada e implementada el 2026-10-02; colores sustituidos por el plan 4 | Opción A, plan 2 y acceso 13 |
| 4 | [Adaptar la app a la opción G](capabilities/thrs/04-adaptacion-opcion-g.md) | Verificada; pendiente de revisión final | Plan 3 |
| 5 | [Nombres, correos únicos, permisos en frontend y cierre de Equipo y Roles](capabilities/thrs/05-permisos-frontend-formularios-equipo.md) | Verificada; aprobada el 2026-10-04, pendiente de aceptación final | Plan 2, plan 4, RBAC 01, registro 06 y precios 12 |

## Restricciones vigentes

- Git solo lectura hasta confirmación final del usuario.
- Sin commits, push, despliegues ni publicaciones.
- Código real permanece en `backend/` y `frontend/`.
- Cualquier desviación de una capability aprobada detiene esa parte del trabajo hasta revisión.
