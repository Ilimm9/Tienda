# Plan 3: Paleta clara y tipografía global

## Estado y control

- Estado: `Aprobada`; aprobación explícita del usuario el 2026-10-02.
- Fecha: 2026-10-02. Implementado el mismo día; ver «Seguimiento».
- Archivo directo en `capabilities/thrs`, sin carpetas adicionales, igual que los planes 1 y 2.
- Solo estilos del frontend. Sin cambios en backend, contratos, rutas ni lógica de negocio.
- Dependencias: opción A del diseño, [plan 2](02-thrs-respaldo-2026-10-04-invitaciones-empleados-sucursal.md) ya implementado, capability 13 (acceso). Se cruza con capability 05 (Tailwind, en revisión) y con la sección «Diseño visual» del [plan 1](01-thrs-respaldo-2026-10-04-alta-inicial-empresa-sucursal.md).

## Objetivo

Que toda la aplicación comparta la paleta crema e índigo del acceso y de las vistas de invitación, con una tipografía que sí se cargue. El tema oscuro queda desactivado mientras se estabiliza el claro.

## Hallazgos

- **La tipografía actual no es Inter.** `styles.css` declara `Inter`, pero ningún archivo la carga: no hay `<link>`, `@font-face` ni paquete de fuente. El navegador cae a la fuente del sistema, distinta en cada equipo. Ese es el origen de la letra que hoy no convence.
- Los colores globales viven en 26 variables de `:root` en `styles.css`. Casi todos los componentes ya las consumen: cambiar esas variables recolorea la mayor parte de la app sin tocar componentes.
- 25 archivos CSS tienen colores fijos. La mayoría tiene uno solo (blanco sobre botón primario). Los que concentran colores fijos: `styles.css` (diálogos y alertas), las pantallas de `auth` y `topbar`.
- La topbar y las vistas de invitación ya usan la paleta nueva mediante variables locales (plan 2). Con la paleta global esas variables sobran.
- PrimeNG usa el preset `Aura` sin personalizar: tablas, paginador, diálogos y selects conservan su azul propio aunque cambien las variables de la app.
- El tema oscuro se activa solo si el sistema operativo lo pide (`ThemeService`), y existe un botón en la topbar. `index.html` declara `lang="en"` y título «Tienda».
- El presupuesto de estilos por componente es 4 kB (aviso) y 9 kB (error): conviene mover estilos compartidos a global en vez de engordar componentes.

## Enfoque recomendado

Sustituir los valores de las variables globales, cargar Inter desde el propio proyecto y ajustar a mano solo shell, PrimeNG y los pocos archivos con colores fijos.

- Alternativa: esperar a capability 05 y hacerlo con Tailwind. Retrasa el cambio visual y mezcla dos riesgos en una entrega.
- Alternativa: mantener variables locales por componente, como en el plan 2. No escala y deja la app con dos paletas.

## Paleta

Los nombres de variable se conservan; cambian los valores del tema claro.

| Variable | Hoy | Propuesto | Uso |
| --- | --- | --- | --- |
| `--text` | `#171b25` | `#172554` | Texto principal |
| `--text-muted` | `#6f7785` | `#5b6478` | Texto secundario |
| `--text-subtle` | `#9299a6` | `#6f7785` | Marcadores y rótulos; el actual no alcanza contraste 4.5:1 |
| `--background` | `#f8fafc` | `#faf8f3` | Fondo general (crema) |
| `--surface-raised` | `#ffffff` | `#ffffff` | Tarjetas, formularios, topbar |
| `--surface-sidebar` | `#f7f8fa` | `#ffffff` | Menú lateral |
| `--surface-hover` | `#edf1f6` | `#f3efe6` | Hover y filas |
| `--border` | `#e0e4ea` | `#e6e0d2` | Bordes |
| `--border-strong` | `#cbd1da` | `#cfc8b8` | Bordes de campos |
| `--primary` | `#2864e9` | `#2c2e6a` | Acciones (índigo) |
| `--primary-hover` | `#1d4fc5` | `#20224e` | Hover de acciones |
| `--primary-soft` | `#edf3ff` | `#ececf4` | Elemento activo, fondos suaves |
| `--primary-border` | `#cad9fb` | `#c9cae0` | Borde de elemento activo |
| `--accent` | `#0284c7` | `#2c2e6a` | Foco visible |
| `--success` / `-soft` | `#00966d` / `#e8f8f2` | `#3f5147` / `#dfe5c8` | Completado, aceptado |
| `--warning` / `-soft` | `#d97904` / `#fff4dd` | `#6b4a00` / `#f8eac1` | Pendiente, aviso |
| `--danger` / `-soft` | `#c43d4b` / `#fff0f1` | `#bd2525` / `#fff1f1` | Error, acción destructiva |

- Variables nuevas para los bloques suaves de la opción A: `--tone-sage` `#dfe5c8`, `--tone-sage-strong` `#cfd9bd`, `--tone-peach` `#f4d3c1`, `--tone-sand` `#f5e6c5`, `--tone-butter` `#f8eac1`, cada una con su color de texto legible.
- Sombras: mismo desplazamiento, tinte índigo en vez de azul gris.
- `--warning` pasa de naranja a ocre oscuro: como color de texto debe leerse sobre su fondo suave.

## Tipografía

- Familia: **Inter**, la misma del diseño propuesto, cargada de verdad.
- Carga recomendada: paquete `@fontsource-variable/inter` importado en `styles.css`. Un solo archivo variable cubre los pesos 400 a 700, se sirve desde el propio dominio y no depende de Google Fonts ni de políticas de contenido externas. Agrega una dependencia de desarrollo de fuente; sin código.
- Alternativa: `<link>` a Google Fonts en `index.html`. Más simple, pero depende de un tercero y de la red del cliente.
- Escala y pesos:

| Uso | Tamaño | Peso | Tracking |
| --- | --- | --- | --- |
| Título de página | 30 px | 700 | -0.035em |
| Título de sección o tarjeta | 18 px | 600 | -0.02em |
| Cuerpo y tablas | 14 px | 400 | 0 |
| Etiquetas de campo y botones | 13–14 px | 600 | 0 |
| Rótulos en mayúsculas | 11–12 px | 600 | 0.06em |

- Cifras tabulares (`font-variant-numeric: tabular-nums`) en tablas de precios, inventario y compras, para que las columnas alineen.
- PrimeNG hereda la familia desde `body`; confirmar en tablas, selects y diálogos.
- `index.html`: `lang="es"` y título con el nombre de la marca cuando el colega del plan 1 lo defina.

## Menú lateral

- Fondo blanco con borde derecho crema; el contenido queda sobre crema, igual que en la opción A.
- Elemento activo: fondo `--primary-soft`, texto e icono índigo, peso 600. Sustituye al bloque azul sólido actual.
- Grupo con hijo activo (Equipo, Catálogo): solo texto índigo y peso 600; el fondo suave queda en el hijo.
- Subelementos sin icono propio o con icono más discreto, sangría de 30 px, altura mínima 44 px.
- Rótulo «Administración»: 11 px, 600, color `--text-muted`.
- **Quitar la tarjeta «Acceso seguro»** del pie: markup y estilos `.security-*`.
- Corregir de paso: Invitaciones no se marca activo en `/equipo/invitaciones/nueva` ni en `/corregir-correo` porque el enlace exige coincidencia exacta. Es un atributo de plantilla, sin lógica nueva.
- Modo compacto (72 px): mismo criterio de activo, solo icono.

## Topbar

- Ya está en crema e índigo. Al subir la paleta a global se eliminan sus variables locales (`:host-context…`) para que no existan dos fuentes de color.
- Selector de contexto: icono sobre `--primary-soft`, borde `--border`, hover con `--primary-border`.
- Icono de sucursal dentro del selector: pasa de verde brillante a `--success` nuevo.
- Icono y nombre de marca: sin cambio hasta que el colega suba el suyo.

## Tema oscuro desactivado

- `ThemeService` aplica siempre `light`: ignora la preferencia del sistema y cualquier valor guardado en `localStorage`.
- Se oculta el botón de luna/sol de la topbar.
- El bloque `:root[data-theme='dark']` de `styles.css` se conserva sin uso, con un comentario que explique por qué. Reactivarlo será otro plan, con su propia paleta oscura.
- Es el único cambio fuera de CSS y plantillas: unas líneas en `theme.service.ts` y su prueba.

## PrimeNG

- Definir un preset derivado de `Aura` con la escala primaria en índigo y superficies crema, para que tablas, paginador, selects, checkboxes y foco coincidan con la app.
- Revisar `styles/tables.css` y los diálogos de producto en `styles.css`, que hoy fijan grises azulados.
- SweetAlert2 y Sonner ya leen las variables globales; solo verificar. El botón de confirmar destructivo (`#b42332`) pasa a `--danger`.

## Qué aplica y qué no

| Área | Decisión |
| --- | --- |
| Variables globales, tipografía, `body`, enlaces, foco | Aplica |
| Shell: topbar, menú lateral, migas | Aplica |
| Preset PrimeNG, tablas, diálogos, alertas | Aplica |
| Pantallas de acceso (`auth`): login, registro, verificación, recuperar y restablecer | Aplica solo limpieza: sustituir colores fijos por variables. El aspecto no cambia; mosaico y animación intactos |
| Vistas de invitación (plan 2) | Aplica solo limpieza: `invitaciones-tema.css` pasa a apuntar a las variables globales |
| Inicio, Negocios, Sucursales, Equipo, Roles, Catálogo, Productos, Proveedores, Compras, Precios, Ventas | Heredan color y letra. Solo se corrigen colores fijos y contrastes; no se rediseña su distribución |
| Tema oscuro | No aplica; desactivado |
| Logo, icono y nombre de marca | No aplica; lo trae el plan 1 |
| Tailwind (capability 05) | No aplica; este plan no lo instala ni lo bloquea |
| Backend, API, rutas, permisos | No aplica |
| Plantillas de correo | No aplica; siguen con su estilo actual |

## Archivos previstos

- `frontend/src/styles.css`, `frontend/src/styles/tables.css`, `frontend/src/styles/uploads.css`.
- `frontend/src/index.html`, `frontend/package.json` y lockfile (fuente).
- `frontend/src/app/app.config.ts` (preset PrimeNG).
- `frontend/src/app/layout/`: `theme.service.ts` y su prueba, `topbar`, `sidebar`, `breadcrumbs`, `shell`.
- Archivos CSS de `features/` con colores fijos, según el barrido de la fase 4.

## Fases

1. **Base**: cargar Inter, sustituir valores de variables, tonos nuevos, `lang="es"`.
2. **Shell**: menú lateral (incluye quitar «Acceso seguro»), topbar sin variables locales, migas.
3. **Tema oscuro**: forzar claro y ocultar el botón.
4. **Componentes de terceros y barrido**: preset PrimeNG, tablas, diálogos, alertas; reemplazar colores fijos por variables en `features/`.
5. **Revisión**: recorrer cada sección en escritorio y móvil, medir contraste, ajustar.

Cada fase deja la app usable; pueden entregarse por separado.

## Riesgos

- Estados que hoy se distinguen por color (verde brillante, naranja) quedan más sobrios: cada estado debe conservar texto o icono, no solo color.
- Colores fijos no detectados pueden quedar azules sobre crema; el barrido de la fase 4 y la revisión visual los cubren.
- Inter es más ancha que algunas fuentes del sistema: revisar cortes de texto en tablas y en el selector de contexto.
- Usuarios con preferencia oscura guardada verán claro sin aviso.
- Cruce con el plan 1, que también prevé sustituir la paleta: acordar con el colega quién toca `styles.css` primero para evitar conflictos.

## Tareas y verificación futura

- [x] Confirmar paleta, forma de cargar la fuente y aspecto del menú lateral.
- [x] Aprobar formalmente el alcance antes de implementar.
- [x] Fase 1: fuente y variables.
- [x] Fase 2: shell y retiro de «Acceso seguro».
- [x] Fase 3: tema oscuro desactivado.
- [x] Fase 4: PrimeNG y barrido de colores fijos.
- [ ] Fase 5: revisión visual por sección, escritorio y móvil.
- [ ] Verificar contraste mínimo 4.5:1 en texto y 3:1 en bordes de campos y foco.
- [x] Confirmar en la pestaña de red que Inter se descarga y en estilos calculados que se aplica.
- [x] Ejecutar `ng build` y `ng test`; registrar por separado los fallos previos (`productos.component.spec.ts`, `navigation.config.spec.ts`).
- [x] Actualizar pruebas de `ThemeService` y de `navigation.config` si cambian.

## Criterios de aceptación

- Toda pantalla del panel muestra fondo crema, acciones índigo y texto azul profundo; no quedan botones ni enlaces en el azul anterior.
- La letra es Inter en cualquier equipo, incluidos tablas y diálogos de PrimeNG.
- El menú lateral no muestra «Acceso seguro» y marca el elemento activo con fondo suave e índigo.
- No existe forma de activar el tema oscuro desde la interfaz ni por preferencia del sistema.
- Las pantallas de acceso y de invitación se ven igual que antes del cambio.
- Ningún archivo de `backend/` cambia.

## Decisiones pendientes

- Carga de la fuente: paquete propio (recomendado) o Google Fonts.
- Menú lateral blanco (recomendado, como la opción A) o crema.
- Elemento activo con fondo suave (recomendado) o índigo sólido como en las vistas de invitación.
- Orden respecto al plan 1 del colega para tocar `styles.css`.

## Seguimiento

Decisiones del usuario, 2026-10-02: fuente propia; menú lateral crema como la opción A; elemento activo con fondo suave; el usuario aplica este cambio antes que el plan 1 del colega.

### Implementado

- Inter variable servida desde el proyecto (`@fontsource-variable/inter`), familia en `--font-sans`, cifras tabulares en tablas.
- Variables globales con la paleta crema e índigo y tonos `--tone-*`; `lang="es"`.
- Menú lateral crema, activo con fondo suave, sin «Acceso seguro»; los subelementos se marcan activos también en sus subpantallas.
- Topbar sin variables locales ni botón de tema. `ThemeService` aplica siempre claro.
- Preset PrimeNG derivado de Aura con primario índigo.
- Colores fijos de diálogos, alertas y sombras pasados a variables; `invitaciones-tema.css` apunta a las variables globales.
- Desviación menor: buscador y filtro por estado (`.search-field`, `.status-tabs`) se añadieron a `styles.css` porque Empleados los usaba sin tener estilos y se veían como controles nativos.

### Pruebas

- `ng build` de producción: sin errores; la fuente se empaqueta en `media/`. Avisos de presupuesto de estilos, la mayoría previos.
- `ng test`: 130 pasan, 9 fallan en `productos.component.spec.ts` y `navigation.config.spec.ts`, igual que antes del cambio.
- Revisión visual en claro, 1440 px: Empleados, Productos, Roles y permisos, Invitaciones, Login. Backend sin cambios.

### Pendientes

- Fase 5 incompleta: falta recorrer Inicio, Negocios, Sucursales, Proveedores, Compras, Precios y Ventas, y todo en móvil.
- Contraste no medido con herramienta.
- Login: con Inter cargada, la fila «Contraseña / ¿Olvidaste tu contraseña?» queda pegada al campo; no se tocó para respetar el alcance de acceso.
- Avisos de presupuesto de 4 kB en `aceptar-invitacion.css` e `invitaciones-tema.css`.

### Ajuste del 2026-10-02: acercar la app a la opción A

Petición del usuario tras ver el resultado: el índigo de la propuesta pesa más. Cambia la decisión previa de «activo con fondo suave».

- Menú lateral, filtros por estado y página seleccionada del paginador: índigo sólido con texto blanco.
- Filtros por estado como pastillas con borde, igual que en la lista de invitaciones.
- Títulos de página más grandes y apretados (700, -0.045em); rótulo superior discreto, sin mayúsculas.
- Tarjetas planas (sin sombra) y encabezado de tabla blanco con rótulos pequeños en gris.
- Icono del selector de contexto sobre salvia, como toque cálido del mosaico.
- Verificado: `ng build` sin errores; `ng test` 130 pasan, 9 fallos previos sin cambio; capturas de Inicio y Empleados.

### Variante B en la rama `thrs-vB` (2026-10-02)

Solicitada por el usuario para que el colega compare. Parte de `thrs` y solo cambia estilos del frontend.

- Paleta del logo: marino `#0b1f33`, esmeralda `#046c5e` / `#12b88a`, naranja `#f0923a`, fondo `#f4f6f3`.
- Títulos en Sora (`@fontsource-variable/sora`); cuerpo en Inter.
- Menú lateral marino con icono activo esmeralda; marca, contexto y cuenta de la topbar en hexágonos.
- Filtros y paginador activos en marino; botones primarios en esmeralda; preset PrimeNG en esmeralda.
- Vistas de invitación: resumen como «ticket» marino; aceptación sobre fondo marino con tarjeta blanca.
- Acceso: mismos mosaico y animación, con los colores de la variante.
- Verificado: `ng build` sin errores; `ng test` 130 pasan, 9 fallos previos sin cambio; capturas de Empleados, Nueva invitación y Aceptar invitación.
- No replica la estructura completa de las láminas B (riel del asistente, tarjetas de opción, contadores): eso exige cambiar plantillas.
