# Capability 13: Acceso y registro Stockion

## Estado

Verificada.

## Aprobación

Aprobada explícitamente por el usuario el 2026-10-01 mediante la instrucción: “Implement the plan”.

La seguridad de contraseña y el diseño responsive del registro se aprobaron explícitamente el 2026-10-06 mediante la misma instrucción. Esta capability consolida las anteriores capabilities 14 y 15, eliminadas como documentos independientes.

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
- Reemplazar la marca del acceso por cajitas apiladas y reforzar la adaptación responsive de login y registro.
- Usar el logo `trace.svg` de `frontend/public` en la marca de login y registro.
- Añadir una entrada GSAP coordinada para marca, título y formulario, con énfasis animado en el logo público.
- Evaluar una apertura con el logo Stockion como protagonista antes de revelar el acceso.
- Reemplazar la apertura de logo protagonista por un mosaico inicial que se ensambla antes de mostrar el acceso.
- Eliminar la compuerta de introducción que demora el formulario tras ensamblarse el mosaico.
- Omitir el ensamblaje y la transición FLIP del mosaico en viewports de hasta 820 px, donde el mosaico ya está oculto, para presentar el acceso sin demora.
- Guiar la creación de contraseña en `/registro` con requisitos visibles y actualizados mientras se escribe.
- Exigir entre 8 y 72 caracteres, al menos una mayúscula, un número y un carácter especial, tanto en frontend como en backend.
- Mostrar inmediatamente si la confirmación coincide con la contraseña una vez que el usuario comienza a escribirla.
- Organizar el formulario de registro en dos columnas en escritorio y en una columna amplia bajo el breakpoint de 820 px.
- Mantener logo Stockion y enlace para iniciar sesión visibles, y centrar verticalmente el registro de escritorio sin ocultar controles en viewports bajos.
- Aplicar la paleta local del login a verificación de correo y al flujo completo de recuperación de contraseña.

## Reglas y contratos

- Las rutas públicas `/login` y `/registro` se conservan.
- No cambian los endpoints, payloads, validaciones ni redirecciones de autenticación.
- El shell respeta `prefers-reduced-motion` y mantiene navegación por teclado y foco visible.
- Los gráficos del mosaico son SVG locales, sin recursos externos.
- Los tonos de cada cuadro e ilustración se conservan, salvo la etiqueta, cuyo fondo usa el verde `#3F5147` por solicitud explícita; los controles y marca del acceso usan índigo `#2C2E6A` sin modificar el primario global.
- `confirmarContrasena` solo pertenece a la interfaz y nunca se envía al backend; un carácter especial es cualquier carácter que no sea letra, dígito ni espacio.
- La API rechaza contraseñas inseguras con `400`, código `CONTRASENA_INSEGURA` y detalle para el campo `contrasena`; los estados de la guía no revelan su contenido.
- En escritorio el contenido se compacta antes de permitir desplazamiento; el centrado vertical aplica desde 821 px y se sustituye por alineación superior bajo 761 px.
- Verificación y recuperación usan fondo crema, acciones índigo `#2C2E6A` y hover `#20224E` de forma local, sin modificar tokens globales ni contratos de autenticación.

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
- Login y registro muestran el logo Stockion de tres cajas apiladas y conservan proporción desde 320 px hasta escritorio, sin alterar el resto de la aplicación.
- El mosaico se conserva sobre 820 px; bajo ese ancho se oculta y los formularios pueden desplazarse verticalmente sin desbordamiento horizontal.
- La marca de login y registro usa el logo público `trace.svg` sin recolorearlo ni deformarlo.
- Marca, título y formulario se revelan con GSAP en paralelo; el logo se destaca con una entrada breve y una flotación sutil, anulada con movimiento reducido.
- Al abrir login o registro, los seis cuadros se ensamblan antes de revelar la marca, título y formulario; el logo público conserva su rebote y flotación posteriores.
- La compuerta de introducción se libera al terminar el último tile; formulario y título entran sin desenfoque y sin esperar la animación decorativa del logo.
- Bajo 820 px, login y registro se muestran directamente y su cambio no espera una animación de mosaico que no es visible.
- La guía indica por separado longitud, mayúscula, número y carácter especial; la confirmación informa inmediatamente si coincide.
- El formulario no invoca el registro mientras incumpla un requisito o las contraseñas difieran, y el backend rechaza la misma política para solicitudes directas.
- Desde 821 px, apellidos, contacto y contraseñas se organizan en pares; nombre, guía, errores y acción principal conservan el ancho necesario.
- Bajo 821 px, el registro usa una sola columna sin desbordamiento horizontal y, en escritorio estándar, marca, formulario y enlace de acceso se perciben centrados como un bloque.
- Verificación, solicitud de recuperación y restablecimiento conservan sus rutas, validaciones, temporizadores y estados, con acciones y focos visualmente coherentes con login.

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
- [x] Reemplazar la marca del acceso por el logo Stockion de cajas apiladas.
- [x] Reforzar la adaptación responsive de login y registro.
- [x] Mostrar el logo público `trace.svg` en login y registro.
- [x] Animar con GSAP la entrada del acceso y la flotación sutil del logo.
- [x] Evaluar la introducción GSAP del logo protagonista para compararla con el mosaico.
- [x] Animar el ensamblaje inicial del mosaico y revelar después el acceso.
- [x] Revelar el formulario inmediatamente tras el último tile.
- [x] Omitir la introducción y FLIP del mosaico cuando el breakpoint móvil lo oculta.
- [x] Definir y probar la política de contraseña del dominio cuenta.
- [x] Aplicar la validación reactiva, guía visual accesible y respuesta HTTP estable para contraseñas inseguras.
- [x] Ajustar la estructura semántica, retícula responsive y centrado vertical seguro del formulario de registro.
- [x] Ejecutar las verificaciones de frontend y backend, formato y actualización del grafo.
- [x] Ajustar la paleta local de verificación y recuperación al diseño de login.

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
- Logo Stockion y responsive 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Logo público Stockion 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado sin cambios de topología.
- Entrada GSAP Stockion 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado.
- Apertura de logo protagonista 2026-10-01: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido a 2381 nodos y 6586 relaciones.
- Apertura de mosaico ensamblado 2026-10-06: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify reconstruido a 2626 nodos y 7389 relaciones.
- Revelación inmediata de formulario 2026-10-06: `tsc --noEmit`, Prettier y `git diff --check` correctos; Graphify actualizado a 2626 nodos y 7389 relaciones.
- Acceso móvil inmediato 2026-10-06: `tsc --noEmit` y Prettier del shell/capability correctos; `graphify update .` actualizado a 2627 nodos y 7392 relaciones.
- Contraseña segura 2026-10-06: `GOCACHE=/private/tmp/tienda-go-cache go test ./internal/domain/cuenta ./internal/application/cuenta ./internal/interfaces/http/cuenta` y `tsc --noEmit` correctos; Prettier, `git diff --check` y `graphify update .` correctos (2639 nodos y 7416 relaciones).
- Registro responsive 2026-10-06: `tsc --noEmit`, Prettier, `git diff --check` y `graphify update .` correctos; el centrado vertical se verificó con fallback para alturas reducidas.
- Paleta de verificación y recuperación 2026-10-06: `tsc --noEmit`, Prettier y `git diff --check` correctos; `graphify update .` ejecutado sin cambios de topología.
- Pendiente conocido: el runner Angular requiere Node 22; el entorno principal usa Node 14. Las pruebas de componente se actualizaron y la compilación TypeScript fue correcta; las verificaciones previas con Node 22 no reportaron fallos.
