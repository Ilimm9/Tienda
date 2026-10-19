# Pendientes

Lista única de lo que quedó abierto en los planes 1 a 7. Actualizada el 2026-10-05.

Estado de ramas en esa fecha: `main`, `production` y `thrs` apuntan a `b9aff37`. El servidor aún no se actualiza.

## Antes de desplegar en el servidor

- [ ] **Intentos de OTP**: `config.go` usa 10 por omisión desde el commit `a827745` (2026-09-30); la prueba `TestProductionAcceptsStrongOTPSecret` espera 5 y falla. Decidir el valor correcto y alinear código y prueba.
- [ ] **Correo real**: enviar los cuatro correos a Gmail, Outlook y Apple Mail, también en modo oscuro. Solo se revisaron en Chromium.
- [ ] **Base de datos**: si se conserva, ejecutar antes el reporte de miembros activos sin sucursal asignada y de correos de empleado duplicados. Si se borra y se recrea, no hace falta.
- [ ] Ejecutar `pasos-update-server.md`. El repositorio no tiene CI: subir a `production` no despliega.
- [ ] Compilar las imágenes de producción (`compose.production.yaml`); no se probó.

## Coordinación con Ilu

- [ ] Su commit `13f410d favicon` en `origin/ilu` sigue sin integrar. Chocará en `frontend/src/index.html`, `frontend/public/favicon.ico` y `frontend/package-lock.json`.
- [ ] Probar de punta a punta, como propietario: registrar una compra, carga masiva de productos y autorizar precios. La validación de sucursal asignada corre antes de su lógica.
- [ ] 8 pruebas fallan en `productos.component.spec.ts` desde antes de estos planes.
- [ ] Plan de Catálogo, Proveedores, Compras y Precios con el kit: en pausa. Esas cuatro plantillas no usan ninguna clase del kit y repiten `primary-button`, `secondary-button`, `eyebrow` y `form-grid` en varios CSS.

## Marca (plan 7)

- [ ] `.env.example` y `docker-compose.yml` siguen con `MAIL_FROM_NAME` por omisión `Tienda`.
- [ ] Títulos «proyecto Tienda» en `AGENTS.md`, `DOCUMENTACION.md` y `frontend/README.md`.
- [ ] Logo definitivo: el actual es el temporal pulido.
- [ ] Identificadores técnicos sin renombrar por decisión: cookie `tienda_session`, claves `tienda.*` del navegador, módulo Go `tienda/backend`, dominio `tienda.mergemakers.com`, clases `tienda-alerta`.

## Alta inicial (plan 1)

- [ ] Probar en navegador: restaurar una empresa archivada, rechazo `403 SOLO_PROPIETARIOS` y aviso de invitado sin sucursal. Solo tienen pruebas unitarias.
- [ ] Revisar una por una las capturas móviles.
- [ ] El selector de moneda del asistente corta su texto («MXN · Peso m…»).
- [ ] `aceptar-invitacion` conserva su propia barra; falta pasarla a `barra-minima` del kit.
- [ ] Lista y detalle de Negocios y Sucursales siguen con CSS propio.
- [ ] Los formularios de Negocio y Sucursal crecieron en HTML (6.0 → 12.6 kB y 5.8 → 10.9 kB).

## Permisos y cuentas (plan 5)

- [ ] Perfiles anteriores sin apellido quedan con `primer_apellido` vacío.
- [ ] Retirar la columna `apellidos` de `perfil_usuarios` en una migración posterior.
- [ ] «Mínimo un rol» al asignar se valida solo en el frontend; la API acepta dejar a un miembro sin roles.
- [ ] Reenviar una invitación conserva su rol sin volver a comprobar que el emisor pueda otorgarlo.
- [ ] El endpoint `…/mis-permisos` quedó sin uso desde el frontend.
- [ ] Un empleado invitado que quiera abrir un negocio propio necesita otra cuenta. Revisar si se permite con un permiso especial.
- [ ] El menú no marca «Roles y permisos» como activo en sus subpantallas.
- [ ] Ventas no tiene permisos: la sección aún no existe.
- [ ] Nombre de un empleado ya vinculado a una cuenta: quién lo edita.
- [ ] La columna «Acceso» de Empleados no muestra «Invitación enviada»: la lista no trae ese dato.

## Invitaciones (plan 2)

- [ ] Pruebas de integración con base para reenvío y aceptación.
- [ ] Pruebas de carrera entre aceptación y reenvío, y de rollback.
- [ ] Casos sin probar: SMTP caído desde la interfaz, sucursal archivada o rol desactivado entre emisión y aceptación.

## Kit y estilos (planes 3, 4 y 6)

- [ ] Equipo y Roles quedaron en 79 kB; la meta era 60 kB. Bajar más exige componentes Angular (`<app-campo>`, `<app-boton>`).
- [ ] Capturas «antes» para comparar lado a lado: no se tomaron.
- [ ] Retirar `--tone-peach` y `--tone-sand` si ya nadie los usa.
- [ ] Revisión visual con datos reales de Negocios, Sucursales, Proveedores, Compras, Precios y Ventas.
- [ ] Login: la fila «Contraseña / ¿Olvidaste tu contraseña?» queda pegada al campo.

## Calidad general

- [ ] Medir contraste con herramienta: texto 4.5:1; bordes de campo y foco 3:1.
- [ ] Navegación completa con teclado, incluida la topbar índigo.
- [ ] Pruebas de componentes Angular para las pantallas de Equipo y Roles.
- [ ] Presupuesto de estilos: `topbar.component.css` (6.5 kB), `productos.component.css` (8.0 kB), `compras.component.css` (5.2 kB) y `aceptar-invitacion.css` (4.2 kB) superan los 4 kB.
- [ ] El paquete inicial pesa 749 kB; el presupuesto es 500 kB.
- [ ] `graphify` no está instalado en este equipo: el grafo no se actualiza.
- [ ] Planes 1, 5, 6 y 7 siguen «pendientes de aceptación final».
