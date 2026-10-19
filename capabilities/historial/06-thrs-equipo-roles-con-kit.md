# Plan 6: Equipo y Roles con el kit visual

## Estado y control

- Estado: `Verificada`; implementada el 2026-10-04, pendiente de aceptación final. Aprobada: aprobación explícita del usuario el 2026-10-04 con las opciones recomendadas (46 px, mejoras visuales incluidas, solo Equipo y Roles, sin componentes Angular).
- Fecha: 2026-10-04. Sin implementación: solo este plan y una propuesta visual.
- Archivo directo en `capabilities/thrs`, igual que los planes 1 a 5.
- Solo frontend: plantillas y estilos. Sin cambios en backend, contratos, rutas, permisos ni validaciones.
- Dependencias: kit del [plan 4](04-thrs-adaptacion-opcion-g.md), Tailwind de capability 05 y pantallas del [plan 5](05-thrs-permisos-frontend-formularios-equipo.md).
- Propuesta visual: canvas [Equipo y Roles con el kit](https://claude.ai/artifact/8KkscNyg7PNeWg1nheryMJ), seis láminas: Roles, Nuevo rol, Piezas del kit, Empleados, Detalle de empleado e Invitaciones.

## Objetivo

Que las diez plantillas de Equipo y Roles usen el kit global para sus piezas (botones, campos, pastillas, avisos, tablas) y Tailwind solo para acomodarlas (grid, flex, espacios, anchos, visibilidad por tamaño). Mismo aspecto, mucho menos marcado repetido y un solo lugar para cambiar cada pieza.

## Hallazgos

| Plantilla | Tamaño | Clases de más de 120 caracteres | Usos del kit |
| --- | --- | --- | --- |
| `empleado-form` | 21.5 kB | 13 | 0 |
| `invitaciones` | 17.6 kB | 16 | 0 |
| `invitacion-nueva` | 15.5 kB | 10 | 0 |
| `empleado-detalle` | 15.1 kB | 10 | 0 |
| `roles` | 12.1 kB | 5 | 0 |
| `corregir-correo` | 11.7 kB | 12 | 0 |
| `rol-form` | 11.5 kB | 10 | 0 |
| `empleados` | 8.8 kB | 7 | 0 |
| `empleado-sucursales` | 8.5 kB | 8 | 0 |
| `miembro-roles` | 6.7 kB | 5 | 0 |
| **Referencia:** `aceptar-invitacion` | 14.4 kB | **0** | **59** |

- Las diez plantillas suman 129 kB y no usan ni una clase del kit. El kit (`styles/componentes.css`, 892 líneas) se carga igual en todas.
- 96 atributos `class` largos; solo 51 son distintos. El botón principal está copiado 7 veces con 335 caracteres; el campo de texto, 5 veces con unos 320.
- `aceptar-invitacion` ya trabaja como se propone: kit para las piezas y cero clases largas. Es la prueba de que el enfoque funciona en este proyecto.
- El kit nació para invitaciones: sus pastillas se llaman `--pendiente`, `--aceptada`, `--expirada`, `--cancelada`. Faltan nombres por significado para estados de empleado, rol y membresía.
- Faltan en el kit cuatro piezas que hoy se arman a mano: encabezado de sección con icono, contador, tarjeta de opción con casilla y tarjeta de grupo de permisos.

## Regla de uso

**Kit** para todo lo que tiene aspecto propio: color, borde, radio, sombra, tipografía, estados (`hover`, `focus`, `disabled`, inválido).

**Tailwind** solo para acomodo: `grid`, `flex`, `gap-*`, `items-*`, `justify-*`, `p-*`, `m-*`, `w-*`, `max-w-*`, `min-w-0`, `truncate`, `sr-only` y prefijos `sm:`, `md:`, `lg:`.

**No permitido en estas plantillas** después del cambio: utilidades de color (`bg-*`, `text-*` de color, `border-*` de color), `rounded-*`, `shadow-*`, tamaños de fuente y pesos sueltos, `hover:` y `focus:`. Si hace falta, la pieza no existe en el kit: se agrega al kit.

Límite orientativo: ningún atributo `class` de más de 80 caracteres en estas diez plantillas.

## Ampliación del kit

Se agrega a `styles/componentes.css`. Nombres en español, como el resto.

| Pieza | Clases | Sustituye |
| --- | --- | --- |
| Pastillas por significado | `.pastilla--exito`, `--info`, `--atencion`, `--neutra`, `--peligro` | Seis combinaciones `[class.bg-*]` por pantalla. Las cuatro actuales quedan como alias |
| Encabezado de sección | `.seccion-encabezado`, `.seccion-icono`, `.seccion-icono--salvia` | El bloque icono + título + descripción, hoy repetido 14 veces |
| Contador | `.contador`, `.contador--salvia` | Las pastillas «5 de 22», «2» |
| Tarjeta de opción | `.opcion`, `.opcion--marcada` (vía `:has(:checked)`) | Casillas de rol, permiso y «marcar como principal» |
| Grupo de permisos | `.grupo-permisos`, `.grupo-permisos__encabezado` | Tarjeta de módulo en el formulario de rol |
| Botón de icono | Ya existe `.icono-accion`; se revisa tamaño 40 px y variante `--peligro` | Enlaces y botones de 40 × 40 armados a mano |
| Dato en ficha | `.dato`, `.dato__rotulo`, `.dato__valor` | Pares rótulo/valor del detalle de empleado |
| Tarjeta de enlace | `.enlace-tarjeta` | Accesos «Sucursales asignadas» e «Invitar» |

No se crean colores ni tokens. Cada pieza nueva usa las variables de `styles.css`.

Tamaños: el kit usa botón y campo de 46 px; las pantallas actuales usan 44 px. Se adopta el del kit (46 px) y `.boton--compacto` (38 px) dentro de tablas. Es el único cambio visible previsto, de 2 px.

## Correspondencia

| Hoy (utilidades) | Después (kit) |
| --- | --- |
| Botón principal, 335 caracteres | `boton` |
| Botón con borde | `boton boton--secundario` |
| Cancelar | `boton boton--fantasma` |
| Campo de texto | `campo` dentro de `campo-bloque` |
| Campo con icono | `campo-grupo`; `campo-grupo--error` |
| Select con flecha | `campo-grupo campo-grupo--select` |
| Campo de solo lectura | `campo-grupo campo-grupo--lectura` |
| Etiqueta y asterisco | `etiqueta` + `obligatorio` |
| Mensaje de error, ayuda, correcto | `error-campo`, `ayuda`, `correcto-campo` |
| Leyenda de obligatorios | `leyenda-obligatorios` |
| Formulario con secciones y pie | `formulario-g`, `__seccion`, `__titulo`, `__acciones` |
| Aviso rojo, verde, amarillo | `aviso aviso--error`, `--salvia`, `--atencion` |
| Estado con punto | `pastilla pastilla--exito` y variantes |
| Avatar con iniciales | `avatar`, `avatar--salvia`, `avatar--neutro`, `avatar--grande` |
| Tarjeta de tabla, barra y pie | `tabla-tarjeta`, `tabla-barra`, `tabla-pie` |
| Persona en tabla | `tabla-persona` |
| Filtro por estado | `segmentado`, `segmentado__opcion`, `segmentado__contador` |
| Sin resultados, cargando | `estado-vacio` |
| Encabezado de página | `pagina`, `encabezado-pagina` |

Lo que queda en Tailwind en cada plantilla: rejillas de dos columnas (`grid gap-4 sm:grid-cols-2`), alineaciones, anchos máximos y márgenes entre bloques.

## Mejoras incluidas en la propuesta

Sin funciones nuevas; solo orden y consistencia.

- **Roles**: una fila de acciones idéntica a la de Empleados; el rol de sistema destacado con candado; permisos del rol como barra de proporción «6 de 22» además del número.
- **Formulario de rol**: grupos de permisos con el mismo encabezado en todos los módulos; resumen fijo al pie con el total elegido.
- **Asignar roles**: cada rol como tarjeta de opción con descripción y número de permisos.
- **Empleados**: columna «Acceso» como pastilla, igual que «Estado»; acciones por fila con `icono-accion`.
- **Detalle de empleado**: ficha en pares rótulo/valor alineados; accesos como tarjetas de enlace; zona de estado separada con la acción destructiva al final.
- **Invitaciones**: vencimiento y estado en una sola lectura; acciones por fila agrupadas, peligro al final.
- **Formularios**: mismo orden de pie en todos (leyenda a la izquierda; Cancelar y acción principal a la derecha).

## Fases

1. **Kit**: agregar las piezas de «Ampliación del kit» y los alias de pastillas. Ninguna pantalla cambia todavía.
2. **Roles** (3 plantillas, las más recientes y pequeñas): sirve de piloto para validar la regla.
3. **Empleados**: lista, detalle, formulario y sucursales.
4. **Invitaciones**: lista, nueva y corregir correo.
5. **Limpieza**: retirar del kit lo que quede sin uso (`aviso--arena`, `boton--verde` si nadie los usa) y medir.

Cada fase deja la app usable y se revisa por separado con capturas antes y después.

## Qué no cambia

- Lógica de componentes, validaciones, permisos y textos.
- Identificadores, `aria-*`, `for`/`id` y orden de foco.
- Tablas PrimeNG: siguen con `p-table.app-table` y `styles/tables.css`.
- Pantallas fuera de Equipo y Roles (Catálogo, Productos, Compras, Precios, Negocios, Sucursales, acceso).
- `aceptar-invitacion`: ya usa el kit.

## Riesgos

- Diferencias de 1 a 2 px entre utilidades y kit (alto de campo, relleno). Mitigación: capturas comparadas por pantalla.
- `:has(:checked)` para la tarjeta de opción: compatible con los navegadores actuales; sin él la casilla sigue funcionando, solo sin resaltado.
- Tailwind corre sin Preflight: las piezas del kit no deben depender de un reinicio de estilos.
- Las pruebas que buscan clases concretas pueden fallar; se actualizan sus selectores.
- Trabajo de otra persona en las mismas plantillas: acordar el orden antes de la fase 3.

## Tareas

- [x] Aprobar el alcance y resolver las decisiones pendientes.
- [x] Fase 1: ampliar el kit.
- [x] Fase 2: Roles.
- [x] Fase 3: Empleados.
- [x] Fase 4: Invitaciones.
- [x] Fase 5: limpieza y medición.
- [x] `ng build` y `ng test` sin fallos nuevos frente a la línea base (156 pasan; 8 fallos previos en `productos.component.spec.ts`).
- [ ] Capturas antes y después a 1440 px y 390 px de las diez pantallas. Hecho: 10 pantallas a 1440 px y 2 a 390 px, solo «después».

## Criterios de aceptación

- Las diez plantillas no contienen utilidades de color, radio, sombra ni tipografía; solo acomodo.
- Ningún atributo `class` supera 80 caracteres en esas plantillas.
- El tamaño conjunto de las diez plantillas baja de 129 kB a 60 kB o menos.
- Cambiar el aspecto de un botón, un campo o una pastilla se hace en un solo lugar del kit y se refleja en las diez pantallas.
- Las capturas antes y después coinciden salvo las mejoras listadas y el alto de 46 px.
- Ninguna prueba nueva falla; ningún archivo de `backend/` cambia.

## Decisiones pendientes

- **Alto de botones y campos**: 46 px del kit (recomendado, una sola medida en toda la app) o bajar el kit a 44 px.
- **Mejoras visuales**: incluir las de «Mejoras incluidas» (recomendado) o hacer solo el cambio a kit sin tocar distribución.
- **Alcance**: solo Equipo y Roles (recomendado) o aprovechar para Negocios y Sucursales, que aún tienen CSS propio.
- **Piezas como componentes Angular** (`<app-boton>`, `<app-campo>`): fuera por ahora (recomendado) o plan aparte.

## Seguimiento

- 2026-10-04: plan redactado, propuesta visual publicada y, tras la aprobación del usuario, implementado.

### Resultado medido

| Medida | Antes | Después |
| --- | --- | --- |
| Diez plantillas | 129.0 kB | 79.1 kB |
| Atributos `class` de más de 80 caracteres | 96 (más de 120) | 0 |
| Atributos con utilidades de aspecto | 263 valores distintos | 0 |
| Enlaces `[class.bg-*]`, `[class.text-*]` | 49 | 0 |
| `styles/componentes.css` | 16.0 kB | 24.4 kB |
| Suma de plantillas y kit | 145.0 kB | 103.5 kB |

**La meta de 60 kB no se alcanzó**: quedó en 79 kB. Lo que resta es estructura y enlaces de Angular (formularios con `aria-*`, tablas, condicionales), no clases. Bajar más exige componentes Angular (`<app-campo>`), que quedaron fuera por decisión.

### Implementado

- Kit ampliado: pastillas por significado, `contador`, `boton--peligro`, `seccion-encabezado` y `seccion-icono`, `tarjeta-encabezado`, `tarjeta-pie`, `opcion`, `grupo-permisos`, `dato`, `enlace-tarjeta`, `dialogo`, `proporcion`, `estado-vacio--guiado`, `casilla`, `codigo`, `tenue`, `enlace-fila`, `acciones-fila`, `resumen-frase`. El error de campo también responde a `aria-invalid`.
- Diez plantillas reescritas con kit y Tailwind solo de acomodo. `presentacion.ts` centraliza la variante de pastilla y de avatar por estado.
- Mejoras: barra de proporción de permisos por rol; «Acceso» como pastilla y editar desde la fila en Empleados; ficha del empleado en pares rótulo/valor; «Dar de baja» como botón de peligro; un solo botón «Reenviar» que cambia de peso según el estado; mismo pie en todos los formularios.
- Limpieza: se retiraron del kit las reglas sin uso (`aviso--arena`, `campo-select`, `formulario-g__titulo`, `pastilla--pendiente`, `--expirada`, `--cancelada`, `segmentado__opcion--activa`, `seccion-icono--neutro`).

### Desviaciones

- Botones y campos pasan de 44 a 46 px, como se decidió.
- El formulario de empleado perdió el rótulo «Equipo · Empleados» sobre el título: las migas de la topbar ya lo dicen.
- Vista previa y resumen laterales muestran sus datos en pares rótulo/valor apilados en lugar de dos columnas.
- La columna «Acceso» de Empleados no muestra «Invitación enviada», como sugería la propuesta: la lista no trae ese dato.

### Pruebas

- `ng build` correcto; el paquete inicial bajó de 748.6 a 740.1 kB.
- `ng test`: 156 pasan; 8 fallan en `productos.component.spec.ts` (previos).
- Revisión visual con Chromium sin interfaz y datos reales en base temporal: 12 capturas en `capturas-plan6/`, sin errores de consola. Tres defectos encontrados y corregidos: pie del diálogo partido en dos filas, números de «Qué sigue» estirados y textos a 16 px dentro de opciones.
- Backend sin cambios.

### Pendientes

- Capturas «antes» para comparar lado a lado; no se tomaron.
- Los dos últimos ajustes (asterisco rojo en «Roles» y «Permisos», singular «1 permiso») no se volvieron a capturar.
- Contraste y teclado sin medir con herramienta.
- Negocios y Sucursales siguen con CSS propio (fuera de alcance).
