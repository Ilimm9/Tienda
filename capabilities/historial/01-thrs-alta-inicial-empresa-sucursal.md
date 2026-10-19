# Plan 1: Alta inicial de empresa y sucursal, y formularios con el kit

## Estado y control

- Estado: `Verificada`; aprobada e implementada el 2026-10-05, pendiente de aceptación final. Sustituye al borrador del 2026-10-01 (copia en `.respaldo-2026-10-04/`).
- Archivo directo en `capabilities/thrs`. Solo frontend.
- Dependencias: contexto y permisos del plan 5, kit del plan 6, láminas G1 a G4.

## Contexto

Una cuenta nueva entra hoy a `/inicio`, `contextoGuard` la manda a «Elige un negocio» con la lista vacía y nadie le dice qué hacer. El plan 1 (`capabilities/thrs/01-alta-inicial-empresa-sucursal.md`) lo describía pero sigue en `Borrador`. El usuario pide implementarlo ahora y, de paso, pasar los formularios de Negocio y Sucursal al kit global (como se hizo con Equipo y Roles en el plan 6).

Decisiones del usuario (2026-10-05):
- El asistente **se abre solo tras el login y se puede saltar**. Cambia la decisión anterior de «solo aviso».
- Pide **lo mínimo**: empresa `nombre_comercial`; sucursal `nombre` y `codigo`. El resto va plegado en «Agregar más datos».
- El plan de Catálogo, Proveedores, Compras y Precios queda **en pausa** (no se toca en este trabajo).

Solo frontend. Backend ya tiene todo: `POST /negocios` crea membresía y rol propietario; `POST …/sucursales` marca la primera como principal; `403 SOLO_PROPIETARIOS` impide a un invitado crear empresa.

## Parte A · Asistente de alta inicial

### Cuándo aparece

Se calcula en el frontend con lo que ya trae `ContextoService` (`frontend/src/app/contexto/contexto.service.ts`); no hace falta endpoint nuevo.

| Situación | Destino |
| --- | --- |
| Sin negocios y `puedeCrearNegocio()` | Asistente, paso «Tu empresa» |
| Negocio propio activo sin sucursales (`sin_sucursal` y `puede(sucursalCrear)`) | Asistente, paso «Tu primera sucursal» |
| Contexto `listo` | Inicio normal, sin aviso |
| Miembro invitado sin sucursal asignada | Inicio con aviso «Solicita una asignación al administrador»; nunca el asistente |
| Varios negocios sin elegir | `/seleccionar-negocio`, como hoy |

Nuevo `estadoAlta` computado en `ContextoService`: `'requiere_empresa' | 'requiere_sucursal' | 'sin_asignacion' | 'listo'`.

### Navegación

- Ruta nueva `/configuracion-inicial`, **fuera del shell** y con `authGuard`, igual que `invitacion/:token`: pantalla enfocada, sin menú lateral.
- Guard nuevo `altaInicialGuard` en `contexto/` para la ruta `inicio` (sustituye ahí a `contextoGuard`):
  - `requiere_empresa` o `requiere_sucursal` y no se saltó en esta sesión: redirige a `/configuracion-inicial`.
  - Saltado: deja entrar a Inicio aunque no haya contexto.
  - Resto: comportamiento actual de `contextoGuard`.
- «Hacerlo después» guarda una marca en `sessionStorage` (no `localStorage`): en el siguiente inicio de sesión vuelve a ofrecerse.
- El asistente se protege a sí mismo: con contexto `listo` redirige a `/inicio`; un invitado ve «Sin acceso».
- Login y OTP no cambian: siguen yendo a `/inicio` y el guard decide. El retorno de invitación conserva prioridad (ya ocurre antes).

### Diseño base: opción G del canvas

Referencia obligatoria: láminas **G1 a G4** de [Tienda — Alta inicial e invitaciones](https://claude.ai/artifact/MKcds35194SagmaVT5sb84) (última fila del canvas). Se respetan distribución, textos y colores; los valores ya existen como tokens en `styles.css`.

Común a G2, G3 y G4:
- **Barra mínima índigo** de 64 px, sin menú lateral: logo, «Stockion», separador y rótulo «Configuración inicial». A la derecha «Salir al panel» (no aparece en G4).
- **Dos mitades**. Izquierda: columna de 460 px con pasos, título de 46 px (`-0.05em`), descripción y tarjeta blanca de formulario. Derecha: **mosaico** de 6 × 6 celdas (salvia `#DFE5C8`, lavanda `#ECEAFB` / `#DCD8F6`, violeta `#4F46C8`, verde `#3F5147`) que funciona como **resumen en vivo**. El mosaico se oculta por debajo de 820 px.
- **Pasos** «1 Empresa — 2 Sucursal — 3 Listo»: actual en violeta, hecho en verde con palomita, pendiente con borde. Ya existen en el kit (`pasos`, `paso--activo`, `paso--hecho`, `paso__linea--hecha`).

### Pantalla

Archivo nuevo `features/alta-inicial/alta-inicial.ts` + `.html` (sin sufijo `.component`, convención de los planes 2 y 5).

- Reutiliza del kit: `pasos`, `formulario-g` / `tarjeta`, `campo-grupo` (campo con icono), `etiqueta`, `obligatorio`, `ayuda`, `error-campo`, `correcto-campo`, `boton`, `boton--secundario`, `boton--fantasma`, `aviso`, `aviso--salvia`, `pastilla`, `seccion-icono`.
- Piezas nuevas en el kit: `mosaico` y `mosaico__celda` con variantes de tono, `barra-minima` (hoy vive dentro de `aceptar-invitacion.css`; se sube al kit para compartirla), `titulo-asistente`.
- Tres pasos: **Empresa → Sucursal → Listo**.
  1. **G2 · «Tu empresa»**. Texto: «Con el nombre comercial basta para empezar. Lo demás se completa después en Empresa.»
     - *Nombre comercial* (obligatorio, icono de edificio). Ayuda: «Así aparecerá en tickets, invitaciones y reportes.»
     - *Moneda* y *Zona horaria* en una fila, como selects con sus valores por omisión (`MXN · Peso mexicano`, `Ciudad de México (UTC−6)`), usando las mismas opciones del formulario de negocio.
     - Plegado **«Agregar más datos»** (decisión del usuario; sustituye al aviso «No los pedimos ahora» de la lámina): razón social, RFC, teléfono y correo.
     - Botón de ancho completo «Guardar y continuar →». Debajo: «¿Prefieres hacerlo luego? **Seguir explorando**».
     - Mosaico: celda grande salvia «TU EMPRESA» con el nombre escrito en vivo y «MXN · Ciudad de México»; celda violeta «1/2»; celda lavanda «PRIMERA SUCURSAL · Sigue después de guardar»; celda blanca «Serás propietaria/o · Con acceso a todas las sucursales activas de esta empresa.»
  2. **G3 · «Tu primera sucursal»**. Texto: «Es el lugar donde vendes y guardas mercancía. Podrás agregar más cuando quieras.»
     - *Nombre de la sucursal* (icono de tienda) y *Código* (icono `#`), ambos obligatorios. Con valor válido muestran «✓ Correcto»; con error, borde rojo y mensaje: «Usa letras, números y guiones, sin espacios. Ejemplo: CEN-01.»
     - Plegado «Agregar más datos»: teléfono y dirección.
     - Aviso salvia con estrella: «Al ser la primera sucursal activa, queda como principal automáticamente.»
     - Botón «Crear sucursal →». Debajo: «Tu empresa ya quedó guardada. **Terminar después**».
     - Mosaico: «✓ EMPRESA GUARDADA» con el nombre; celda lavanda grande «PRIMERA SUCURSAL» con el nombre en vivo, pastilla «Principal» y el código (o «Código por confirmar»); celda verde «2/2».
     - **Sin botón «Atrás»** (la lámina lo muestra): la empresa ya está guardada y volver obligaría a editarla desde el asistente. Se corrige después en Negocios.
  3. **G4 · «Ya puedes operar»**. Texto: «Dejamos seleccionadas tu empresa y tu sucursal. Las verás en la barra superior en todo el panel.»
     - Lista de dos filas: *Empresa* con pastilla «Propietaria/o»; *Sucursal principal* «Nombre · CÓDIGO» con pastilla «Activa».
     - Botones apilados: «Ir a Inicio →» (principal) y «Invitar a mi equipo» (secundario, a `/equipo/invitaciones/nueva`).
     - Mosaico decorativo con palomita grande sobre violeta.
- Género: la lámina dice «Serás propietaria» / «Propietaria». Se usa una forma neutra: «Tendrás el rol de propietario» y pastilla «Propietario», igual que el resto de la app.
- Usa los servicios existentes: `NegocioService.crear`, `SucursalService.crear`. Mismas validaciones que los formularios completos (se copian los `Validators`, no se inventan).
- Tras cada alta: `contexto.recargar()`; al terminar selecciona empresa y sucursal creadas y navega a `/inicio`.
- Reanudar: si ya existe la empresa, entra directo al paso 2; no la duplica.
- Antes de proponer empresa nueva consulta `NegocioService.listar('archivado')`: si la persona tiene una propia archivada, ofrece «Restaurar» junto a «Crear otra».
- Doble clic bloqueado mientras guarda; error del servidor bajo su campo (`campos`) y foco al primer inválido con `enfocarPrimerInvalido` (`shared/formularios/formulario.ts`).
- `403 SOLO_PROPIETARIOS`: mensaje claro y salida a Inicio.

### Inicio (lámina G1)

`features/home/home.component.*`, mientras `estadoAlta` no sea `listo`:

- Encabezado «Inicio» con la fecha del día a la derecha.
- **Bloque de configuración** en tarjeta blanca de radio 18 px, junto a un mosaico de 4 × 4:
  - Pastilla con barra de avance: «Configuración inicial · 0 de 2 pasos» (o «1 de 2» si ya existe la empresa).
  - Título de 40 px: «Tu panel ya está abierto. Falta tu empresa.» (con empresa guardada: «Falta tu primera sucursal.»).
  - Texto: «Registra tu empresa y una primera sucursal para vender, comprar y llevar inventario. Mientras tanto puedes recorrer cualquier sección.»
  - Lista de dos pasos: «1 Tu empresa — Solo el nombre comercial. RFC y datos fiscales pueden esperar.» y «2 Tu primera sucursal — Nombre y código. Queda como principal.», cada uno con su estado («Siguiente», «Pendiente», «Hecho»).
  - Botones: «Configurar empresa →» (o «Crear sucursal →») y «Seguir explorando» (fantasma, baja a las tarjetas).
- Sección **«Lo que podrás hacer»**: tres tarjetas (Productos, Inventario y compras, Equipo) con el texto de la lámina y «Ver sección →».
- Variante para invitado sin sucursal: mismo bloque sin pasos ni botón, con «Solicita una asignación al administrador».
- Con `estadoAlta = listo` Inicio queda como hoy.

Barra superior (`layout/topbar`): sin empresa, el selector de contexto muestra «Sin empresa todavía» con borde punteado, como en G1. Se verifica si ya está implementado (plan 4 lo describe) y se completa si falta.

`contexto/seleccionar-negocio.component.ts`: el enlace del estado vacío apunta al asistente en vez de `/negocios`.

## Parte B · Formularios de Negocio y Sucursal con el kit

Misma regla del plan 6: **kit para el aspecto, Tailwind solo para acomodo**; ninguna clase de más de 80 caracteres.

- `features/negocios/negocio-form.component.html` (6.0 kB, 17 campos) y `features/sucursales/sucursal-form.component.html` (5.8 kB, 14 campos).
- Estructura `formulario-g` con secciones (`seccion-encabezado` + `seccion-icono`):
  - Negocio: *Datos generales*, *Contacto*, *Dirección*, *Operación* (moneda y zona).
  - Sucursal: *Datos de la sucursal*, *Dirección*; «Es principal» como `opcion opcion--compacta`.
- Campos con `campo-bloque`, `etiqueta`, asterisco `obligatorio`, `campo`/`campo-grupo`, `error-campo`, `aria-required`, `aria-invalid`, `aria-describedby`.
- Pie igual al de Equipo: `leyenda-obligatorios`, Cancelar (`boton--fantasma`) y guardar (`boton`).
- Se retiran `negocio-form.component.css` y `sucursal-form.component.css`. `negocio-shared.css` y `sucursal-shared.css` se conservan: los usan lista y detalle, que quedan fuera.
- Lógica intacta: mismos `formControlName`, validadores, `submit()` y servicios. Solo se añade foco al primer error.
- Si falta alguna pieza se agrega a `styles/componentes.css`; nada de utilidades de color en las plantillas.

## Archivos

- Nuevos: `features/alta-inicial/alta-inicial.{ts,html}`, `contexto/alta-inicial.guard.ts` y sus `*.spec.ts`.
- Modificados: `app.routes.ts`, `contexto/contexto.service.ts`, `contexto/seleccionar-negocio.component.ts`, `features/home/home.component.{ts,html,css}`, `layout/topbar` (solo si falta el estado «Sin empresa todavía»), `features/invitacion/aceptar-invitacion.*` (solo para tomar la barra mínima del kit), los dos formularios (`.html`, `.ts` mínimo), `styles/componentes.css`.
- Eliminados: los dos `.css` de formulario.
- Documentos: `capabilities/thrs/01-…md`, `PLAN_IMPLEMENTACION.md`.
- `backend/`: sin cambios.

## Fuera de alcance

- Lista y detalle de Negocios y Sucursales (siguen con su CSS).
- Catálogo, Proveedores, Compras y Precios (plan en pausa; son trabajo de otra persona).
- Logo definitivo, tema oscuro, cambios de API.

## Verificación

1. `ng build` y `ng test` con Node 24: sin fallos nuevos frente a la línea base (156 pasan; 8 previos en `productos.component.spec.ts`). Actualizar `negocio-form.component.spec.ts`, `sucursal-form.component.spec.ts` y `home.component.spec.ts`.
2. Pruebas nuevas: `estadoAlta` en cada situación de la tabla; `altaInicialGuard` (redirige, respeta «saltar», no afecta a invitados); asistente (valida, crea empresa, reanuda en paso 2, no duplica).
3. Recorrido real en base temporal (`DB_NAME=tienda_e2e`, receta guardada en memoria), sin tocar la base local:
   - Cuenta nueva: login → asistente → empresa → sucursal → Inicio con menú completo de propietario.
   - «Hacerlo después»: entra a Inicio con aviso; el botón retoma en el paso correcto.
   - Alta interrumpida tras crear la empresa: al volver entra en el paso 2.
   - Empleado invitado: nunca ve el asistente.
   - Editar negocio y sucursal con los formularios nuevos: guarda, valida y muestra errores del servidor.
4. Capturas con Chromium sin interfaz a 1440 px y 390 px: tres pasos del asistente, Inicio con aviso, y los dos formularios con errores. **Comparar una por una contra las láminas G1 a G4** del canvas y anotar cualquier diferencia.
5. Medir: tamaño de las dos plantillas antes y después, y confirmar 0 utilidades de aspecto.
6. Git solo lectura; sin `add` ni `commit`.

## Seguimiento

### Implementado (2026-10-05)

- `ContextoService`: `estadoAlta`, `altaPospuesta`, `posponerAlta()`; la marca vive en `sessionStorage` y se limpia al cerrar sesión.
- `contexto/alta-inicial.guard.ts`: `altaInicialGuard` (ruta `inicio`) y `asistenteGuard` (ruta `/configuracion-inicial`, fuera del shell).
- `features/alta-inicial/alta-inicial.{ts,html}`: tres pasos según G2 a G4, mosaico con resumen en vivo, «Agregar más datos» plegado, restaurar empresa archivada, reanudación en el paso 2.
- Inicio según G1: bloque de configuración con avance, lista de pasos y «Lo que podrás hacer»; variante para invitado sin sucursal.
- Topbar: «Sin empresa todavía» con borde punteado, enlazado al asistente (no existía).
- Formularios de Negocio y Sucursal con el kit; se retiraron sus dos archivos CSS.
- Kit: `barra-minima`, `asistente`, `titulo-asistente`, `mosaico`, `plegable`, `alta-aviso`, `alta-paso`, `pastilla--solida`.

### Desviaciones

- La barra mínima se agregó al kit, pero `aceptar-invitacion` conserva la suya: no se tocó para no arriesgar una pantalla ya aceptada.
- G3 sin botón «Atrás», como estaba previsto.
- G4: el segundo botón dice «Registrar a mi equipo» y lleva a nuevo empleado; invitar exige un empleado registrado y la lámina llevaba a una lista vacía.
- Textos en forma neutra: «Tendrás el rol de propietario», pastilla «Propietario».
- El selector de moneda corta su texto en la columna angosta («MXN · Peso m…»).
- Cinco celdas del mosaico superan 80 caracteres en `class` (clases del kit más `grid-area`); ninguna usa utilidades de aspecto.

### Medición de los formularios

| Archivo | HTML antes | HTML después | CSS antes | CSS después |
| --- | --- | --- | --- | --- |
| `negocio-form` | 6.0 kB | 12.6 kB | 2.2 kB | 0 |
| `sucursal-form` | 5.8 kB | 10.9 kB | 3.8 kB | 0 |

Las plantillas **crecieron**: antes cada campo era `<label><span>…</span><input></label>` con estilos por selector de etiqueta; ahora lleva `id`, `for`, `aria-invalid`, `aria-describedby` y su error bajo el campo. Se ganó consistencia y accesibilidad, no tamaño.

### Pruebas

- `ng build` correcto. `ng test`: 176 pasan; 8 fallan en `productos.component.spec.ts` (previos). 20 pruebas nuevas: `estadoAlta`, guards, asistente e Inicio.
- Recorrido real con Chromium sin interfaz en base temporal vacía, sin errores de consola:
  - Cuenta nueva: `/inicio` abre el asistente; empresa, sucursal y cierre; Inicio queda con contexto «Abarrotes La Esperanza · CEN-01 · Centro» y menú completo de propietario.
  - Recarga a mitad del alta: vuelve al paso 2 sin duplicar la empresa (una sola fila en `negocios`).
  - «Salir al panel»: Inicio con aviso y topbar «Sin empresa todavía»; volver a `/inicio` ya no redirige.
  - Con el alta completa, `/configuracion-inicial` redirige a Inicio.
  - Formularios de negocio y sucursal: cargan datos reales, validan, enfocan el primer error; el código de sucursal queda de solo lectura al editar.
- 15 capturas en `capturas-plan1/`, comparadas con G1 a G4.
- `backend/` sin cambios.

### Pendientes

- No se probó en navegador: restaurar una empresa archivada, el `403 SOLO_PROPIETARIOS` y la variante de invitado sin sucursal (cubiertas por pruebas unitarias).
- Capturas móviles tomadas pero no revisadas una por una.
- Lista y detalle de Negocios y Sucursales siguen con CSS propio.
- Contraste y teclado sin medir con herramienta.
