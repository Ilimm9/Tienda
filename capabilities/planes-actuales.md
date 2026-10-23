# Planes actuales

Consolidación documental: 2026-10-19. Los planes y entregas anteriores están en [historial/](historial/); sus estados, aprobaciones y resultados originales se conservan. Archivar no equivale a cerrar ni aceptar una entrega.

Este documento reúne pendientes ya registrados. **Su vigencia técnica debe confirmarse antes de retomarlos**: aquí no se repitieron pruebas de aplicación ni se aprobó nuevo alcance.

## Trabajo pendiente y autorización

| Tema y fuente | Pendiente registrado | Límite de aprobación |
| --- | --- | --- |
| [06 · Producción segura](historial/06-preparacion-produccion-segura.md) y [plan de aplicación](historial/06-preparacion-produccion-segura-plan-aplicacion.md) | Resolver DB legacy `5433` sin asignar catálogos a otro negocio; completar respaldos, restauración, límites y operación de fases 4–7. | Fases 1–3 y cambios locales de demo aprobados. Resto y servicios externos conservan revisión/autorización separada. |
| [06 · Revisión Resend](historial/06-preparacion-produccion-segura.md) y [07 · Correo](historial/07-servicio-correo-transaccional.md) | Propuesta Resend/SES pendiente de aprobación; entrega real de OTP e invitación, cuotas, rechazo/timeout y SPF/DKIM/DMARC por comprobar. | Correo 07 fases 1–3 aprobado. La revisión Resend no queda aprobada por esta reorganización; configuración externa corresponde al usuario. |
| [15 · Plataforma](historial/15-administracion-plataforma-metricas.md) | Recuperar build Angular estable y suite global; verificar fallos registrados de alta inicial y configuración OTP. Código y pruebas focalizadas ya entregados. | Alcance aprobado el 2026-10-15; no autoriza nuevos perfiles ni operación externa. |
| [07 · Marca y producción](historial/07-thrs-marca-stockion-correos-produccion.md) | Revisar marca/configuración pendiente y publicación. | Aprobación reducida a código; Docker, configuración, Git y despliegue no se incluyen. |
| [05 · Sistema visual](historial/05-tailwind-sistema-visual.md) | Conciliar borrador original con Tailwind y opción G ya entregados; expansión del kit a Catálogo/Proveedores/Compras/Precios registrada en pausa. | El borrador amplio conserva `En revisión`. No asumir aprobada su expansión por entregas parciales posteriores. |

## Verificaciones y revisiones abiertas

- **Aplicación:** verificación del 2026-10-20 en entrega 20: build de producción correcto, 207 pruebas Angular fuera de alta inicial correctas; suite global conserva 6 fallos de alta inicial y el fallo Go de configuración OTP. Resolver el fallo histórico `TestProductionAcceptsStrongOTPSecret`; ejecutar suites Go/Angular y build en entorno compatible. Historial también registra fallos de `productos.component.spec.ts`, alta inicial y presupuestos de estilos/bundle; no se certifica su presencia actual.
- **Productos y compras:** probar como propietario importación, recepción y autorización de precios de punta a punta. Verificaciones Angular pendientes en [08](historial/08-proveedores-seccion-independiente.md), [09](historial/09-productos-proveedores.md), [10](historial/10-compras-proveedores.md), [11](historial/11-fechas-auditoria.md), [12](historial/12-validacion-costos-precios-sucursal.md), [13](historial/13-acceso-animado-stockion.md) y [14](historial/14-panel-editorial-alta-inicial.md).
- **Productos (lista):** [21 · Lista en memoria con refresco silencioso](historial/21-cache-listado-productos.md) cerrada el 2026-10-21 con aceptación final del usuario.
- **Productos (carga masiva):** [24 · Correcciones a la carga masiva](historial/24-correccion-masiva-producto.md) cerrada el 2026-10-22 con prueba manual correcta y aceptación final del usuario.
- **Productos (surtido por sucursal):** [28 · Surtido por sucursal](historial/28-surtido-por-sucursal.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Productos (precio al editar):** [29 · Guardar el precio al editar](historial/29-guardar-precio-al-editar.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Observaciones 26:** [observaciones a los cambios de Ilian](historial/26-observaciones-ilian.md) cerradas el 2026-10-23; quedan solo las fallas previas de pruebas automáticas descritas en «Aplicación».
- **Productos (crear):** [31 · Solo el propietario crea productos](historial/31-crear-productos-solo-propietario.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Compras (productos y proveedores):** [27 · Compras con cualquier producto y varios proveedores](historial/27-compras-cualquier-producto.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Compras (producto nuevo):** [22 · Código de barras opcional](historial/22-compras-producto-sin-codigo.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Compras (captura):** [23 · Productos compactables](historial/23-compras-productos-compactables.md) cerrada el 2026-10-23 con aceptación final del usuario.
- **Colaboradores:** [20 · Arreglos de invitaciones](historial/20-arreglos-invitaciones.md) cerrada el 2026-10-20: delegación, concurrencia/rollback en PostgreSQL, baja/reactivación, dependencias, refresco y selector; OTP/invitación/reenvío con Mailpit y navegador comprobados. Aceptación final registrada. SMTP caído, operación externa, membresías legacy sin roles y edición de nombres vinculados conservan su revisión histórica independiente. Los roles existentes incompletos requieren autorización específica antes de ampliarlos.
- **Cuentas:** revisar perfiles legacy con primer apellido vacío, retiro futuro de `apellidos` y posibilidad de que un invitado cree negocio propio; cualquier migración o cambio de reglas requiere plan aprobado.
- **Alta inicial:** navegador para restauración de negocio, `403 SOLO_PROPIETARIOS`, invitado sin sucursal, selector de moneda y capturas móviles; revisar barra de aceptar invitación y estilos de Negocios/Sucursales.
- **Interfaz:** menú activo de Roles, estado «Invitación enviada», pruebas de Equipo/Roles, contraste, teclado y revisión con datos reales. Limpieza de tokens y tamaño de estilos siguen como deuda, sin autorizar nuevos componentes.
- **Marca (logo):** [25 · Logo vigente en correos y aplicación](25-logo-vigente-correos-app.md) aprobada y verificada localmente el 2026-10-22. Derivados de `trace.svg`, invitación y favicons actualizados; Go, build temporal y Mailpit/Chromium a 600/360 px correctos. Pendientes aceptación final, reinicio del backend activo para cargar el PNG embebido y revisión en clientes reales. Correos anteriores conservan su imagen.
- **Marca y operación:** correo real en Gmail/Outlook/Apple Mail, remitente y títulos pendientes; identificadores técnicos conservados por decisión. Revisar reporte de datos legacy, imágenes de producción y pasos de actualización antes de cualquier despliegue autorizado.
- **Coordinación y aceptación:** confirmar vigencia de la integración pendiente con Ilu antes de tocar archivos compartidos. Entregas verificadas pendientes de aceptación final conservan ese requisito, incluidos los planes personales 1, 4, 5 y 6.

Lista detallada anterior, sin pérdida de puntos: [pendientes al 2026-10-05](historial/thrs-pendientes-2026-10-05.md). Ideas futuras de identidad: [seguimiento original](historial/01-identidad-negocio-sucursales-rbac-seguimiento.md); no son trabajo autorizado.

## Al retomar o terminar un cambio

1. Leer base temática y fuente histórica; confirmar qué pendiente sigue vigente.
2. Crear `NN-nombre.md` con alcance, reglas, contratos, pruebas y aceptación; obtener aprobación explícita antes de implementar nuevo alcance. Una aprobación histórica solo cubre lo descrito en su fuente.
3. Registrar tareas, decisiones, pruebas y resultados en el plan; actualizar base temática y este índice al entregar.
4. Mover el plan terminado a `historial/`, conservando pendientes y aceptación final cuando corresponda.

## Verificación de esta reorganización

- Se conservaron 113 documentos y evidencias anteriores, incluido el índice general previo; modelo inicial e imágenes conservan su contenido original.
- Se revisaron estructura plana, nombres sin colisiones, conservación de archivos y enlaces locales de las bases e índices.
- Solo documentación: sin pruebas de aplicación, consultas a DB, cambios funcionales ni operaciones Git de escritura. `graphify-out/graph.json` no existe en este checkout; no hubo cambios de código que requieran reconstruirlo.
