# Capability 13: Acceso animado Stockion

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-01 mediante la instrucción: “Implement the plan”.

## Alcance

- Unificar `/login` y `/registro` bajo un shell visual persistente.
- Presentar la marca Stockion, copy editorial, formularios existentes y un mosaico de inventario inspirado en la referencia.
- Animar el reordenamiento del mosaico con GSAP y FLIP al cambiar entre ambos modos.
- En móvil, ocultar el mosaico y priorizar el formulario.
- Refinar el acceso para usar dos mitades iguales en escritorio, Inter en toda la composición, acentos de interfaz azules y mosaico sin etiquetas inferiores.
- Mantener el acceso de escritorio dentro de un viewport, sin scroll vertical, y compactar el registro a alturas reducidas.
- Ampliar el mosaico de escritorio y suavizar la transición FLIP entre login y registro.
- Sincronizar el cambio de titular, formulario y mosaico en el mismo ciclo de renderizado.
- Ocultar el contenido de destino hasta que FLIP inicie para evitar un cambio de título visible previo al movimiento.
- Revelar el contenido únicamente al terminar FLIP y usar índigo `#2C2E6A` para las acciones del acceso.
- Revelar marca, título y formulario como un solo grupo tras completar la animación del mosaico.
- Reflejar las posiciones del mosaico en registro sin invertir las ilustraciones SVG.
- Reordenar el mosaico de registro con progresión visual: tiles pequeños arriba, inventario amplio abajo.
- Organizar el mosaico de registro con café arriba, columna decorativa derecha y código/despensa alineados antes de la caja inferior.
- Reordenar el registro como una escalera descendente de derecha a izquierda y reemplazar los iconos de café, latas y planta.
- Ajustar el orden final de los seis tiles de registro con espacio negativo en la esquina inferior derecha.
- Reubicar en login la caja al nivel inferior y subir bolsa y código de barras al nivel central, sin cambiar tamaños.
- Convertir la etiqueta de login a un bloque vertical junto a la bolsa, preservando su área de dos celdas.
- Reordenar exclusivamente el mosaico de registro para igualar la composición de referencia, sin modificar login.
- Extender verticalmente el cuadro del carrito en registro hasta alinearlo con la base de la caja.
- Reducir la etiqueta de registro a la misma área vertical que la de login y orientar su ilustración.

## Reglas y contratos

- Las rutas públicas `/login` y `/registro` se conservan.
- No cambian los endpoints, payloads, validaciones ni redirecciones de autenticación.
- El shell respeta `prefers-reduced-motion` y mantiene navegación por teclado y foco visible.
- Los gráficos del mosaico son SVG locales, sin recursos externos.
- Los tonos de cada cuadro e ilustración se conservan, salvo la etiqueta, cuyo fondo usa el verde `#3F5147` por solicitud explícita; los controles y marca del acceso usan índigo `#2C2E6A` sin modificar el primario global.

## Pruebas y criterios de aceptación

- Cada URL muestra su formulario y distribución correspondiente.
- Alternar login/registro reordena los tiles sin desmontar el mosaico.
- Login y registro conservan sus contratos de API y manejo de errores.
- La experiencia móvil no muestra el mosaico y sigue permitiendo completar ambos formularios.
- Las preferencias de movimiento reducido desactivan la animación.
- En escritorio, contenido y mosaico ocupan el 50% del ancho; ningún cuadro muestra texto decorativo.
- Login y registro no generan scroll en escritorio; móvil conserva desplazamiento para teclado virtual y accesibilidad.
- El mosaico aprovecha su mitad sin superar el viewport y su animación conserva preferencia de movimiento reducido.
- El copy, formulario y tiles inician su transición a la vez, sin un frame visual previo de cambio de layout.
- Al solicitar un cambio de vista, el copy y formulario permanecen ocultos hasta que inicia el desplazamiento del mosaico.
- El copy y formulario nuevos aparecen solamente tras finalizar el desplazamiento completo del mosaico.
- La marca sigue la misma secuencia de revelación y no permanece visible durante el movimiento.
- En registro, los tiles ocupan la composición horizontalmente reflejada; sus gráficos permanecen en orientación normal.
- La caja de inventario ocupa la franja inferior del mosaico de registro.
- En registro, el código de barras queda en formato vertical a la izquierda y la despensa ocupa el bloque central.
- Registro reserva espacios negativos entre niveles de la escalera; muestra ticket, bolsa de súper y carrito como iconos SVG locales.
- El registro usa seis tiles: ticket superior central, botella derecha, etiqueta izquierda, bolsa central, carrito derecho inferior y caja inferior izquierda; no añade bloque decorativo extra.
- En login, la caja ocupa la franja inferior derecha; bolsa y código de barras ocupan la franja central.
- En login, la etiqueta ocupa una columna por dos filas y comparte su borde izquierdo con la bolsa; su gráfico se orienta verticalmente sin afectar el mosaico de registro.
- En registro, el carrito ocupa dos columnas por tres filas y su borde inferior se alinea con el de la caja.
- En registro, la etiqueta ocupa una columna por dos filas junto a la bolsa y usa la misma orientación vertical que en login.

## Tareas y resultados

- [x] Añadir GSAP y el shell persistente de autenticación.
- [x] Reestructurar rutas y formularios para usar el shell.
- [x] Implementar el mosaico, layouts y accesibilidad.
- [x] Cubrir la jerarquía de rutas y conservar las pruebas de formularios existentes.
- [x] Ejecutar verificaciones y actualizar el grafo.
- [x] Igualar proporciones, tipografía y acentos del acceso; retirar etiquetas de los tiles.
- [x] Limitar la vista de escritorio a `100dvh` y compactar el registro en alturas reducidas.
- [x] Ampliar el mosaico y ajustar el ritmo, escala y composición de la transición FLIP.
- [x] Sincronizar la actualización del modo con FLIP y eliminar el retraso de contenido.
- [x] Evitar el parpadeo del contenido durante el renderizado de la ruta destino.
- [x] Encadenar la revelación del contenido al final de FLIP y aplicar acciones índigo locales.
- [x] Integrar la marca al grupo de revelación posterior a FLIP.
- [x] Reflejar las posiciones de los tiles para la vista de registro.
- [x] Mover la caja de inventario al bloque inferior de registro y reordenar los tiles por escala visual.
- [x] Ajustar la composición de registro para alinear visualmente café, código, despensa, caja y planta.
- [x] Crear composición escalonada y sustituir café, latas y planta por ticket, bolsa y carrito.
- [x] Reordenar los seis tiles para igualar la referencia y conservar el espacio libre inferior derecho.
- [x] Intercambiar las posiciones de caja, bolsa y código de barras en login preservando sus dimensiones.
- [x] Reorientar verticalmente la etiqueta de login junto a la bolsa, manteniendo su área total.
- [x] Aplicar el verde `#3F5147` al fondo de la etiqueta.
- [x] Reubicar ticket, bolsa y carrito en el mosaico de registro según la referencia.
- [x] Extender el cuadro del carrito de registro una fila hacia abajo.
- [x] Compactar y orientar verticalmente la etiqueta de registro.

## Verificación

- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- `ng test --watch=false` ejecutado con Node 22: bundle de pruebas generado y runner Vitest iniciado sin fallos reportados.
- `ng build` ejecutado con Node 22: compilación iniciada sin errores reportados.
- `git diff --check`: correcto.
- `graphify update .`: correcto; grafo actualizado a 2374 nodos y 6570 relaciones.
- Refinamiento visual 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; se confirmó la eliminación de las cinco etiquetas del mosaico y se actualizó Graphify sin cambios de topología.
- Ajuste sin scroll 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Refinamiento de mosaico 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido con 2374 nodos y 6570 relaciones.
- Sincronización de transición 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido con 2374 nodos y 6570 relaciones.
- Eliminación de parpadeo 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido con 2374 nodos y 6570 relaciones.
- Secuencia final e índigo 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido con 2374 nodos y 6570 relaciones.
- Revelación unificada 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Mosaico en espejo 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Progresión del mosaico 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Composición ordenada de registro 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Escalera e iconografía de compra 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Orden final de registro 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Reordenamiento de login 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Etiqueta vertical de login 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Color de etiqueta 2026-10-01: Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Composición final de registro 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Carrito extendido en registro 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Etiqueta compacta en registro 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Pendiente conocido: el ejecutable `npm test` del entorno usa Node 14 y Angular CLI exige Node 22. La verificación se ejecutó invocando la CLI con Node 22 instalado localmente.
