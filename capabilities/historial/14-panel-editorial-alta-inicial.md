# Capability 14: Panel editorial de alta inicial

## Estado

En implementación.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-07 mediante la instrucción: “Implement the plan”.

## Alcance

- Reemplazar el mosaico decorativo derecho de `/configuracion-inicial` por una vista previa editorial útil.
- Mantener los tres estados actuales: empresa, sucursal y listo.
- Reflejar en vivo nombre, moneda y zona horaria de empresa, y nombre y código de sucursal.
- Mantener el formulario, APIs, rutas, validaciones y persistencia sin cambios.
- Ocultar la vista previa bajo 820 px, como ocurre actualmente con el lateral del asistente.
- Incorporar una fachada de tienda en Empresa y Sucursal, y una ilustración de inventario en Listo.
- Reproducir en los tres estados la composición editorial aprobada: tarjeta principal frontal, ilustración grande a la derecha y tarjeta de siguiente paso inferior.
- Ampliar las ilustraciones para que ocupen el fondo del panel y animar la entrada de cada nuevo paso con GSAP.

## Reglas y contratos

- La previsualización usa HTML, CSS, iconos locales de PrimeIcons e ilustraciones PNG locales en `frontend/public`.
- Las ilustraciones son decorativas, se marcan con `alt=""` y `aria-hidden="true"`; el contenido operativo permanece en HTML.
- El flujo sigue teniendo dos capturas y un cierre; `Listo` no se convierte en un paso adicional.
- Los valores no capturados usan textos de reserva claros: `Tu empresa` y `Tu sucursal`.
- El panel conserva exactamente la mitad derecha en escritorio, sin introducir llamadas HTTP ni estado persistente adicional.
- Colores: fondo cálido, índigo `#2C2E6A`, salvia y tarjetas blancas, respetando contraste y `prefers-reduced-motion`.

## Criterios de aceptación

- Empresa muestra una tarjeta viva con nombre, moneda y zona horaria, más el siguiente paso de sucursal.
- Sucursal conecta visualmente empresa guardada y sucursal principal con nombre y código en vivo.
- Listo resume empresa, sucursal e inventario mediante tres tarjetas conectadas y checks de completado.
- No existe mosaico de celdas en ninguno de los tres estados.
- En móvil no hay desbordamiento horizontal y el panel derecho permanece oculto.
- El asistente conserva sus formularios y el flujo de guardado actual.
- Las ilustraciones no generan scroll horizontal ni se superponen al contenido legible.
- Los tres estados usan la misma jerarquía visual; no se muestra un diagrama de nodos o conexiones.
- La ilustración ocupa el fondo recortado del lateral y las tarjetas se muestran por encima con mayor tamaño.
- Al cambiar de paso, GSAP anima halo, ilustración, tarjeta principal y tarjeta inferior; se omite en móvil y con movimiento reducido.

## Tareas y resultados

- [x] Diseñar y aprobar la dirección visual de referencia.
- [x] Documentar alcance, reglas y criterios de aceptación.
- [x] Reemplazar los tres mosaicos por las composiciones editoriales.
- [x] Ajustar la adaptación responsive y accesibilidad visual.
- [x] Ejecutar validaciones y actualizar el grafo.
- [x] Incorporar fachada de tienda e ilustración de inventario como recursos locales decorativos.
- [x] Ejecutar verificaciones posteriores a la integración de ilustraciones.
- [x] Sustituir la composición conectada por la composición de referencia en los tres estados.
- [x] Ejecutar verificaciones posteriores al ajuste de composición.
- [x] Ampliar la composición de fondo y animar la transición entre pasos con GSAP.
- [x] Ejecutar verificaciones posteriores a la animación.

## Verificación

- `frontend/node_modules/.bin/prettier --write frontend/src/app/features/alta-inicial/alta-inicial.html frontend/src/styles/componentes.css capabilities/14-panel-editorial-alta-inicial/CAPABILITY.md PLAN_IMPLEMENTACION.md`: correcto.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2672 nodos, 7532 relaciones y 142 comunidades.
- Pendiente conocido: el runner Angular requiere Node 22 y el entorno actual usa Node 14; la comprobación estática de TypeScript fue satisfactoria.
- Integración de ilustraciones 2026-10-07: se añadieron PNG RGBA locales para tienda e inventario; `tsc --noEmit`, Prettier, `git diff --check` y `graphify update .` correctos.
- Composición fiel a la referencia 2026-10-07: Empresa, Sucursal y Listo adoptaron tarjeta principal, ilustración protagonista y tarjeta inferior; `tsc --noEmit`, Prettier, `git diff --check` y `graphify update .` correctos.
- Fondo protagonista y transición GSAP 2026-10-07: ilustraciones ampliadas, tarjetas reforzadas y entrada coordinada entre pasos; `tsc --noEmit`, Prettier, `git diff --check` y `graphify update .` correctos (2677 nodos y 7543 relaciones).
