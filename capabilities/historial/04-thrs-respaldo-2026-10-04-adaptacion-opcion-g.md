# Plan 4: Adaptar la app a la opción G

## Estado y control

- Estado: `Verificada`; aprobada explícitamente por el usuario el 2026-10-02 («aplica el plan, tienes mi permiso»). Pendiente de su revisión final.
- Fecha: 2026-10-02.
- Archivo directo en `capabilities/thrs`, igual que los planes 1 a 3.
- Solo frontend: estilos y plantillas. Sin cambios en backend, contratos, rutas ni permisos.
- Referencia visual: fila «Versión G · F pulida» del canvas [Tienda — Alta inicial e invitaciones](https://claude.ai/artifact/MKcds35194SagmaVT5sb84), láminas G1 a G8.
- Dependencias: [plan 3](03-thrs-respaldo-2026-10-04-paleta-tipografia-global.md) implementado en `thrs` (opción A). Se cruza con el [plan 1](01-thrs-respaldo-2026-10-04-alta-inicial-empresa-sucursal.md) (asistente y aviso de inicio) y el [plan 2](02-thrs-respaldo-2026-10-04-invitaciones-empleados-sucursal.md) (vistas de invitación).
- Rama base: `thrs`, no `thrs-vB`. Rama de trabajo `thrs-vG`, creada desde `thrs` con la autorización del usuario.

## Objetivo

Que la app se vea como la opción G: la base crema de A, con la topbar en índigo como único bloque oscuro, violeta para las acciones y tablas y formularios más claros. Sin colores nuevos más allá de esos.

## Qué cambia respecto a A

| Elemento | A (hoy en `thrs`) | G |
| --- | --- | --- |
| Topbar | Blanca, marca «Portal Administrativo» | Índigo `#2c2e6a`, texto blanco, marca Stockion |
| Migas | Segunda barra bajo la topbar | Dentro de la topbar, junto a la marca; la segunda barra desaparece |
| Acción principal | Índigo `#2c2e6a` | Violeta `#4f46c8` |
| Menú lateral, activo | Índigo sólido, texto blanco | Fondo lavanda `#eceafb`, texto violeta, barra izquierda de 3 px |
| Filtros por estado | Pastillas sueltas, activa en índigo | Control segmentado con contador, dentro de la tarjeta de la tabla |
| Tablas | Tarjeta con encabezado blanco | Encabezado en mayúscula chica sobre crema; barra de filtros arriba y pie con total |
| Campos | Borde simple | Icono a la izquierda, anillo de foco lavanda, ayuda, error y correcto con icono |
| Avisos | Bloques durazno, arena y salvia | Lavanda (informativo), salvia (correcto), mantequilla (atención), rojo (error) |
| Durazno y arena | En mosaico, iconos y avisos | Se retiran; el mosaico usa violeta, lavanda y salvia |

## Tokens

Los nombres de variable de `styles.css` se conservan. Cambian valores y se agregan los de la topbar.

| Variable | A | G | Uso |
| --- | --- | --- | --- |
| `--primary` | `#2c2e6a` | `#4f46c8` | Botón principal, enlaces, paso actual |
| `--primary-hover` | `#20224e` | `#3f37a8` | Hover y texto sobre lavanda |
| `--primary-soft` | `#ececf4` | `#eceafb` | Activo del menú, pastilla pendiente |
| `--primary-border` | `#c9cae0` | `#dcd8f6` | Anillo de foco, bordes lavanda |
| `--accent` | `#2c2e6a` | `#4f46c8` | Foco visible |
| `--primary-tint` (nueva) | — | `#f5f3fd` | Fondo de resumen y aviso informativo |
| `--brand-ink` (nueva) | — | `#2c2e6a` | Fondo de la topbar |
| `--brand-ink-raised` (nueva) | — | `#34367a` | Selector de contexto y cuenta sobre la topbar |
| `--brand-ink-line` (nueva) | — | `#45478a` | Separadores y bordes sobre la topbar |
| `--brand-ink-muted` (nueva) | — | `#b9badf` | Migas y textos secundarios sobre la topbar |
| `--surface-sidebar` | `#faf8f3` | `#faf8f3` | Sin cambio: menú crema |
| `--success` / `-soft` | `#3f5147` / `#dfe5c8` | Sin cambio | Correcto, aceptada |
| `--warning` / `-soft` | `#6b4a00` / `#f8eac1` | Sin cambio | Por vencer, expirada |
| `--danger` / `-soft` | `#bd2525` / `#fff1f1` | Sin cambio | Error, cancelar |

- `--tone-peach` y `--tone-sand` dejan de usarse en pantallas. Se conservan definidas hasta terminar el barrido de la fase 7 y luego se eliminan.
- Contraste esperado: blanco sobre `#4f46c8` ≈ 6.9:1; `#3f37a8` sobre `#eceafb` > 7:1; `#b9badf` sobre `#2c2e6a` ≈ 6:1. Se mide en la fase 7.
- Preset PrimeNG (`app.config.ts`): la escala primaria pasa de índigo a violeta con 600 = `#4f46c8`. Propuesta: 50 `#f5f3fd`, 100 `#eceafb`, 200 `#dcd8f6`, 300 `#bdb6ef`, 400 `#8f86e0`, 500 `#6a61d4`, 600 `#4f46c8`, 700 `#3f37a8`, 800 `#322c85`, 900 `#26215f`, 950 `#171439`. `primary.color` = 600 y `highlight` = 100 con texto 700, para que la página activa del paginador y las filas seleccionadas sean suaves y no bloques sólidos.

## Tipografía

Sin cambios respecto al plan 3: Inter servida desde el proyecto, misma escala. G no introduce otra familia.

## Shell

### Topbar

- Fondo `--brand-ink` y altura de 64 px sin cambio.
- Orden, de izquierda a derecha: botón de menú, marca, separador vertical, migas. Después, empujados a la derecha: selector de contexto, notificaciones, cuenta.
- **Migas en la topbar**: `<app-breadcrumbs />` pasa de `app-shell.component.html` a la topbar. Mismo componente y misma lógica; cambian su ubicación y estilos (separador «/», último elemento en blanco 600, resto en `--brand-ink-muted`).
- En el shell, `.workspace` deja de tener dos filas y la página gana la altura de la barra eliminada.
- Móvil (< 1024 px): las migas muestran solo el último elemento; la marca conserva el icono.
- Selector de contexto: fondo `--brand-ink-raised`, borde `--brand-ink-line`, icono sobre salvia. Sin empresa: borde punteado y texto «Sin empresa todavía».
- Cuenta: pastilla `--brand-ink-raised` con avatar de iniciales en salvia. Las iniciales salen del nombre en sesión; si no hay nombre, se usa el icono actual.
- Menús desplegables (contexto, notificaciones, cuenta): siguen blancos; solo se ajustan los acentos a violeta.
- Foco visible sobre índigo: anillo blanco al 70 % para que se distinga del fondo.

### Marca

- G muestra el logo de Stockion. Hoy el frontend no tiene el archivo (solo `favicon.ico`), y el plan 1 asigna la marca al colega.
- Mientras tanto, la topbar usa el icono actual sobre un cuadro blanco de 34 px y el texto que el plan 1 decida. El cambio de marca no bloquea este plan.

### Menú lateral

- Fondo crema y borde derecho, igual que A.
- Activo: fondo `--primary-soft`, texto e icono `--primary-hover`, peso 600 y `box-shadow: inset 3px 0 0 var(--primary)`. Revierte el «índigo sólido» del ajuste del plan 3: con la topbar índigo, un segundo bloque sólido compite.
- Grupo con hijo activo: texto `--text` 600, sin fondo. Hijos con sangría de 28 px y altura de 40 px.
- Modo compacto: mismo criterio, solo icono.

## Componentes

Se definen una vez en `styles.css` para que todas las pantallas los hereden, siguiendo el criterio del plan 3 (estilos compartidos en global, no en cada componente).

### Botones

- Principal: fondo violeta, texto blanco, 46 px de alto, radio de 10 px, peso 600.
- Secundario: blanco con borde `--border-strong` y texto índigo.
- Fantasma: sin borde, texto índigo (Cancelar, Seguir explorando).
- Chico, para acciones dentro de tablas: 38 px.
- Botón de icono: 40 × 40 px, con `aria-label` y `title`.
- PrimeNG `p-button` hereda del preset; se revisan las variantes `secondary`, `text` y `danger`.

### Campos

- Grupo de campo: etiqueta de 13 px y 600; caja de 46 px con icono opcional a la izquierda; texto de ayuda de 12.5 px debajo.
- Foco: borde violeta y `box-shadow: 0 0 0 3px var(--primary-border)` usando `:focus-within` en la caja.
- Error: borde `--danger`, anillo `--danger-soft`, mensaje con icono y `aria-invalid` y `aria-describedby` en el control.
- Correcto: mensaje en `--success` con icono de palomita. Solo donde la validación ya existe (contraseña, nombre).
- Solo lectura: fondo crema y candado a la derecha.
- Select nativo: `appearance: none` y flecha propia dentro de la caja.
- Opcional: «(opcional)» en peso 400 junto a la etiqueta.
- PrimeNG `p-select`, `p-inputtext`, `p-inputnumber` y `p-datepicker`: mismo alto, radio y anillo de foco desde el preset y `styles.css`.

### Formularios largos

- Tarjeta con secciones, cada una con título de 15 px y descripción breve, separadas por una línea.
- Barra de acciones al pie, en crema y con borde superior: Cancelar a la izquierda de la acción principal, ambos alineados a la derecha.
- Se aplica en las pantallas de G (nueva invitación, corregir correo). Los demás formularios heredan campos y botones sin reorganizarse.

### Tablas

- Tarjeta única que contiene:
  - barra superior con filtros, búsqueda y selector;
  - tabla;
  - pie con el total y, si aplica, la paginación.
- Encabezado: fondo crema, 11.5 px, mayúsculas, `letter-spacing: 0.05em`, color `--text-muted`.
- Filas: relleno de 14 px, separador `#efeae0` y hover `#fcfbf8`. Cifras con `tabular-nums` (ya está).
- Primera columna con avatar de iniciales en lavanda o salvia cuando la fila es una persona.
- Columna de acciones alineada a la derecha.
- Se implementa en `styles/tables.css` para `p-table.app-table` y como clases equivalentes para tablas nativas (invitaciones).

### Filtros por estado

- Control segmentado: contenedor crema con borde; opción activa en blanco con sombra leve y contador en lavanda.
- Sustituye a las pastillas sueltas en Invitaciones y Empleados (`.status-tabs`).

### Pastillas de estado

- Pastilla con punto de color y texto: pendiente en lavanda, aceptada en salvia, expirada en mantequilla, cancelada en gris y error en rojo.
- El estado nunca depende solo del color: siempre lleva texto.

### Avisos

- Bloque con icono, título opcional en negrita y texto, en cuatro tonos: lavanda, salvia, mantequilla y rojo.
- Reemplazan a `aviso--salvia`, `aviso--arena` y `aviso--durazno` de las vistas de invitación.

## Pantallas

| Lámina | Pantalla real | Cambio |
| --- | --- | --- |
| G1 · Inicio | `features/home` | Hoy no existe el estado «sin empresa». Ahora: encabezado y tarjetas de sección con iconos en lavanda o salvia (como F2). El bloque «Tu panel ya está abierto» y el mosaico se implementan con el plan 1 |
| G2–G4 · Asistente | No existe | Lo construye el plan 1. Este plan deja listos los estilos de pasos, tarjeta de formulario y mosaico para que el plan 1 los use |
| G5 · Invitaciones | `features/equipo/invitaciones/invitaciones.*` | Plantilla reorganizada: filtros dentro de la tarjeta, sucursal y rol en una columna, avatar, estado con punto, vencimiento con reloj, acciones a la derecha y pie con total y regla de reenvío |
| G6 · Nueva invitación | `features/equipo/invitaciones/invitacion-nueva.*` | Formulario en dos secciones con barra de acciones; resumen en lavanda con avatar; enlace en tipografía monoespaciada con botón Copiar |
| G7 · Aceptar invitación | `features/invitacion/aceptar-invitacion.*` | Topbar mínima índigo sin sidebar; mosaico violeta y salvia; formulario en tarjeta; pasos con línea de progreso |
| G8 · Corregir correo | `features/equipo/invitaciones/corregir-correo.*` | Se mantiene como página. Tarjeta centrada con aspecto de diálogo: cabecera con avatar y cerrar (vuelve a la lista), cuerpo y pie con acciones |
| Resto (Negocios, Sucursales, Ventas, Catálogo, Productos, Proveedores, Compras, Precios, Equipo, Roles) | Varias | Heredan tokens, shell, campos, botones y tablas. Sin reorganizar su distribución |

## Qué aplica y qué no

| Área | Decisión |
| --- | --- |
| Tokens, preset PrimeNG y componentes globales | Aplica |
| Shell: topbar índigo, migas en la topbar, menú lateral | Aplica |
| Invitaciones, nueva invitación, corregir correo, aceptar invitación | Aplica, con cambios de plantilla |
| Inicio con contexto | Aplica: tarjetas y tonos |
| Inicio sin empresa y asistente G2–G4 | No aplica aquí; los hace el plan 1 con estos estilos |
| Pantallas de acceso (`auth`) | Solo hereda tokens. Mosaico y animación intactos; el durazno de su mosaico se conserva hasta decidir si también cambia |
| Logo y nombre de marca | No aplica; lo trae el plan 1 |
| Búsqueda en Invitaciones | Pendiente de decisión (ver abajo) |
| Backend, API, rutas y permisos | No aplica |
| Tema oscuro | No aplica; sigue desactivado |
| Plantillas de correo | No aplica |

## Archivos previstos

- `frontend/src/styles.css`, `frontend/src/styles/tables.css`.
- `frontend/src/app/app.config.ts` (preset).
- `frontend/src/app/layout/`: `shell`, `topbar`, `breadcrumbs`, `sidebar`.
- `frontend/src/app/features/home/home.component.*`.
- `frontend/src/app/features/equipo/invitaciones/`: `invitaciones.*`, `invitacion-nueva.*`, `corregir-correo.*`, `invitaciones-tema.css`.
- `frontend/src/app/features/invitacion/aceptar-invitacion.*`.
- `frontend/src/app/features/equipo/empleados.component.*`: solo el control segmentado.
- Archivos CSS de `features/` que todavía usen `--tone-peach` o `--tone-sand`, según el barrido.
- Pruebas `*.spec.ts` del shell, la topbar y las migas si cambian sus selectores.

## Fases

1. **Tokens y preset**: valores nuevos, variables de topbar y preset PrimeNG violeta. La app ya se verá con acciones violeta.
2. **Shell**: topbar índigo, migas dentro de la topbar, retiro de la segunda barra, menú lateral con activo lavanda, avatar de cuenta.
3. **Kit de formularios y botones**: grupo de campo, estados, select con flecha, botones, avisos y pastillas en `styles.css`; ajustes de PrimeNG.
4. **Tablas y filtros**: `tables.css`, control segmentado y clases de tabla nativa.
5. **Equipo**: plantillas de Invitaciones, Nueva invitación y Corregir correo; control segmentado en Empleados.
6. **Aceptar invitación e Inicio**: topbar mínima y mosaico en aceptar; tarjetas del inicio.
7. **Barrido y revisión**: retirar durazno y arena de pantallas; recorrer cada sección a 1440 px y 390 px junto al canvas; medir contraste.

Cada fase deja la app usable y puede revisarse por separado.

## Riesgos

- Mover las migas a la topbar reduce espacio en pantallas medianas. Con contexto, cuenta y migas largas («Equipo / Invitaciones / Nueva»), las migas deben recortarse con elipsis antes que empujar el selector.
- Pasar `--primary` a violeta recolorea toda la app de golpe: algún lugar que usaba `--primary` como color de texto oscuro puede quedar demasiado claro. El barrido de la fase 7 lo cubre.
- El cambio de highlight en PrimeNG afecta filas seleccionadas y la página activa del paginador en Productos, Compras y Catálogo.
- Las pruebas que buscan las migas dentro del shell fallarán al moverlas; se actualizan en la fase 2.
- Cruce con el plan 1 (marca, asistente) y con el plan 2 (vistas de invitación): acordar con el colega el orden de cambios en `styles.css` y en la topbar.
- Presupuesto de 4 kB por componente: las vistas de invitación ya están cerca del límite. Por eso los estilos nuevos van a global.

## Tareas y verificación futura

- [x] Revisar y aprobar este plan.
- [x] Resolver las decisiones pendientes (se tomaron las recomendadas).
- [x] Crear la rama de trabajo `thrs-vG` desde `thrs`.
- [x] Fase 1: tokens y preset.
- [x] Fase 2: shell.
- [x] Fase 3: kit de formularios y botones.
- [x] Fase 4: tablas y filtros.
- [x] Fase 5: pantallas de Equipo.
- [x] Fase 6: Aceptar invitación e Inicio.
- [ ] Fase 7: barrido y revisión visual.
- [x] `ng build` sin errores nuevos y `ng test` sin fallos nuevos frente a la línea base del plan 3 (130 pasan; 9 fallos previos en `productos.component.spec.ts` y `navigation.config.spec.ts`).
- [x] Capturas a 1440 px de G5, G6, G7, G8, Inicio, Productos y Compras, comparadas con el canvas.
- [x] Capturas a 390 px de la topbar, Invitaciones y Aceptar invitación.
- [ ] Contraste medido: texto ≥ 4.5:1; bordes de campo y foco ≥ 3:1.
- [ ] Navegación con teclado: foco visible en topbar índigo, menú, filtros, tabla y formularios.
- [ ] `graphify update .` después de modificar código.

## Criterios de aceptación

- La topbar es índigo y contiene las migas; ya no existe una segunda barra bajo ella.
- Botones principales, enlaces y paso actual en violeta; no quedan acciones en el índigo anterior fuera de la topbar.
- Menú lateral crema con activo en lavanda y barra violeta.
- Invitaciones, Nueva invitación, Corregir correo y Aceptar invitación coinciden con G5–G8 en estructura, colores y estados.
- Campos con icono, anillo de foco, ayuda y mensajes de error y correcto, en esas cuatro pantallas y en los formularios PrimeNG que heredan.
- Tablas PrimeNG con encabezado crema en mayúscula chica, hover suave y acciones a la derecha.
- Ninguna pantalla usa durazno ni arena, salvo las de acceso si así se decide.
- Ningún archivo de `backend/` cambia.

## Decisiones pendientes

- **Búsqueda en Invitaciones**: G5 la muestra. Opciones:
  - Omitirla (recomendado; fuera de alcance visual).
  - Filtrar en el cliente sobre la lista cargada. Es lógica nueva pequeña y requiere su prueba.
- **Contadores en el control segmentado**:
  - Mostrar solo el de la vista activa con los datos ya cargados (recomendado).
  - Mostrar todos, lo que puede requerir otra llamada o un cambio de API y queda fuera.
- **Corregir correo**:
  - Página con aspecto de diálogo (recomendado, sin lógica nueva).
  - Diálogo real sobre la lista, que cambia la navegación.
- **Pantallas de acceso**: conservar su mosaico con durazno (recomendado, para no tocar la capability 13) o pasarlo a violeta y salvia.
- **Paginador y filas seleccionadas de PrimeNG**: lavanda suave (recomendado, como G) o violeta sólido.
- **Nombre de rama**: `thrs-vG` u otro.

## Seguimiento

### Decisiones tomadas (2026-10-02)

- Búsqueda en Invitaciones: omitida.
- Contador del filtro segmentado: solo en la vista activa, con los datos ya cargados.
- Corregir correo: página con aspecto de diálogo; la ruta no cambia.
- Pantallas de acceso: sin cambios de mosaico; heredan los tokens (sus botones pasan a violeta).
- Paginador y filas seleccionadas de PrimeNG: lavanda suave.
- Marca: logo temporal en SVG (`frontend/public/stockion-logo.svg`), hexágono índigo con una caja isométrica en lavanda, blanco y salvia, sobre un cuadro blanco en la topbar. Se reemplaza cuando llegue el logo definitivo del plan 1.

### Implementado

- Tokens G en `styles.css` (violeta para acciones, variables `--brand-ink*` para la topbar, `--primary-tint`, `--border-soft`, `--focus-ring`, `--font-mono`) y preset PrimeNG con escala violeta, anillo de foco lavanda y campos con radio de 10 px.
- Kit global nuevo en `styles/componentes.css`: página, botones, campos con icono y estados, select con flecha propia, formulario por secciones, avisos, pastillas con punto, avatares, filtro segmentado, tabla nativa, pasos.
- Shell: topbar índigo con marca, migas dentro de la barra, contexto y cuenta (avatar con la inicial del correo); se retiró la segunda barra. Menú lateral con activo lavanda y barra violeta.
- Migas reescritas: ya no repiten niveles (Roles y permisos, Productos), anteponen el grupo del menú como texto y la sección como enlace en rutas planas (Catálogo / Marcas / Agregar marca; Proveedores / Agregar proveedor), recortan con elipsis y en móvil muestran solo el nivel actual.
- `.status-tabs` (Negocios, Sucursales, Empleados) con el aspecto segmentado; se quitaron sus copias locales. Tablas PrimeNG con encabezado crema en mayúscula chica, hover suave y página activa lavanda.
- Invitaciones, Nueva invitación, Corregir correo y Aceptar invitación según G5–G8. `invitaciones-tema.css` quedó reducido al contenedor y al enlace emitido.
- Inicio con tarjeta de contexto y tarjetas de sección enlazadas.
- Lógica nueva mínima, solo de presentación: iniciales y tono de avatar, texto del pie de la tabla, getters de error en Corregir correo e inicial de la cuenta.

### Pruebas

- `ng build` (Node 24, salida en carpeta temporal porque `dist/` pertenece a otro usuario): sin errores. Avisos de presupuesto de estilos previos; `topbar.component.css` queda en 6.5 kB (aviso, por debajo del límite de 9 kB).
- `ng test`: 136 pasan; fallan los mismos 9 de la línea base (`productos.component.spec.ts`, `navigation.config.spec.ts`). Se agregaron 6 casos de migas y se ajustaron las pruebas de topbar e inicio.
- Revisión visual con Chromium headless y API simulada (datos ficticios, sin tocar la base): Inicio, Invitaciones, Nueva invitación, Corregir correo, Aceptar invitación, Productos, Empleados y Agregar marca a 1440 px; Nueva invitación a 1024 px; Invitaciones a 390 px. Se corrigieron dos defectos encontrados: un enlace con estilo de botón sin texto visible y el hexágono del logo invisible sobre la topbar.

### Pendientes

- Fase 7 parcial: falta revisar Negocios, Sucursales, Proveedores, Compras, Precios y Ventas con datos reales.
- Contraste no medido con herramienta.
- Probar la navegación con teclado sobre la topbar índigo.
- `graphify update .` no se ejecutó: el comando no está instalado en este equipo.
- Asistente G2–G4 e Inicio sin empresa (G1): corresponden al plan 1.
