# Plan 4: Adaptar la app a la opción G

Resumen al 2026-10-04. Versión completa anterior en `.respaldo-2026-10-04/`.

## Estado

- `Verificada`; aprobada el 2026-10-02. Pendiente la revisión final del usuario.
- Solo frontend. Trabajada en `thrs-vG` y presente en `thrs` (commit «propuesta final»).
- Referencia: fila «Versión G» del canvas [Tienda — Alta inicial e invitaciones](https://claude.ai/artifact/MKcds35194SagmaVT5sb84), láminas G1 a G8.
- Es el **sistema visual vigente**. Capability 05 añadió después Tailwind 4 sin Preflight sobre estos tokens y migró la sección Equipo.

## Tokens vigentes (`styles.css`)

| Variable | Valor | Uso |
| --- | --- | --- |
| `--primary` / `-hover` | `#4f46c8` / `#3f37a8` | Acción principal, enlaces, paso actual |
| `--primary-soft` / `-border` / `-tint` | `#eceafb` / `#dcd8f6` / `#f5f3fd` | Activo del menú, anillo de foco, aviso informativo |
| `--brand-ink` / `-raised` / `-line` / `-muted` | `#2c2e6a` / `#34367a` / `#45478a` / `#b9badf` | Topbar y sus elementos |
| Base y estados | Ver [plan 3](03-thrs-paleta-tipografia-global.md) | Crema, texto, éxito, aviso, peligro |

Preset PrimeNG con escala violeta (600 = `#4f46c8`), highlight lavanda y campos con radio de 10 px.

## Implementado

- **Shell**: topbar índigo con marca Stockion, migas dentro de la barra, contexto y cuenta con avatar; sin segunda barra. Menú lateral crema con activo lavanda y barra violeta de 3 px.
- **Migas**: no repiten niveles, anteponen el grupo del menú, recortan con elipsis y en móvil muestran solo el nivel actual.
- **Kit global** `styles/componentes.css`: página, botones (principal, secundario, fantasma, chico, icono), campos con icono y estados, select con flecha propia, formulario por secciones con barra de acciones, avisos (lavanda, salvia, mantequilla, rojo), pastillas con punto, avatares, filtro segmentado, tabla nativa, pasos.
- **Tablas** `styles/tables.css`: encabezado crema en mayúscula chica, hover suave, página activa lavanda, acciones a la derecha.
- **Pantallas** según G5 a G8: Invitaciones, Nueva invitación, Corregir correo (página con aspecto de diálogo) y Aceptar invitación. Inicio con tarjeta de contexto y tarjetas de sección.
- `.status-tabs` segmentado en Negocios, Sucursales y Empleados.
- Logo temporal `frontend/public/stockion-logo.svg`.

## Decisiones tomadas

- Sin búsqueda en Invitaciones. Contador solo en la vista activa del filtro.
- Pantallas de acceso: heredan tokens; mosaico y animación intactos.
- Paginador y filas seleccionadas de PrimeNG en lavanda suave.

## Pruebas

- `ng build` sin errores; `ng test`: 136 pasan, 9 fallos previos (`productos.component.spec.ts`, `navigation.config.spec.ts`).
- Capturas a 1440 px, 1024 px y 390 px con API simulada.

## Pendiente

- [ ] Fase 7: revisar Negocios, Sucursales, Proveedores, Compras, Precios y Ventas con datos reales; retirar `--tone-peach` y `--tone-sand`.
- [ ] **Roles y permisos** sigue con CSS propio y distribución anterior; se atiende en el [plan 5](05-thrs-permisos-frontend-formularios-equipo.md).
- [ ] Contraste medido y navegación con teclado sobre la topbar índigo.
- [ ] Asistente G2 a G4 e Inicio sin empresa (G1): [plan 1](01-thrs-alta-inicial-empresa-sucursal.md).
- [ ] `graphify update .`: el comando no está instalado en este equipo.
