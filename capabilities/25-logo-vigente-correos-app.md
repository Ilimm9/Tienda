# Capability 25: Logo vigente en correos y aplicación

- Estado: **Verificada** localmente el 2026-10-22; pendiente de aceptación final del usuario.
- Aprobación explícita: 2026-10-22, el usuario indicó «lo puedes implementar por favor, cualquier detalle me dices por favor» tras revisar la propuesta de esta capability.
- Fecha: 2026-10-22.
- Fuente de identidad indicada por el usuario: `frontend/public/trace.svg`.
- Coordinación: capability independiente de [24 · Corrección masiva de productos](historial/24-correccion-masiva-producto.md), actualmente trabajada por otro agente. No modificar su plan ni sus archivos de productos, importación o `backend/cmd/api/main.go`.

## Problema y objetivo

El correo de verificación mostrado por el usuario conserva el isotipo anterior. Los correos no leen los assets del frontend: incrustan su propio `backend/internal/infrastructure/correo/plantillas/logo.png`. Además, la pantalla de aceptar invitación todavía referencia `stockion-logo.svg`.

Unificar la identidad visible con `trace.svg`, sin rediseñar la aplicación ni cambiar los flujos de registro, invitación o recuperación.

## Revisión del estado actual

- `backend/internal/infrastructure/correo/renderizador.go`: incrusta `plantillas/logo.png` con `go:embed`; `Logo()` devuelve sus bytes. `NewRenderizador()` utiliza una base común para los cuatro correos.
- `backend/internal/infrastructure/correo/plantillas/base.html.tmpl`: muestra el logo como `cid:logo@stockion`, con nombre alternativo de la aplicación y tamaño de 183 × 40 px.
- `backend/internal/infrastructure/correo/smtp_transporte.go`: adjunta esos bytes como `logo.png`, tipo `image/png`, mediante CID.
- Los cuatro correos afectados son `verificacion`, `invitacion`, `recuperacion` y `contrasena_cambiada`; también sus reenvíos cuando corresponden.
- `frontend/src/app/features/invitacion/aceptar-invitacion.html:3`: única referencia ejecutable encontrada a `stockion-logo.svg`.
- Topbar, acceso, alta inicial y shell de plataforma ya usan `trace.svg`.
- `frontend/src/index.html:8`: el favicon ya apunta a `trace.svg`, pero declara `image/x-icon` en lugar de `image/svg+xml`.
- `frontend/public/favicon.svg`: conserva el mismo isotipo anterior; no se encontraron referencias activas a este archivo ni a `favicon.ico` en el código revisado. El ICO puede ser solicitado directamente por navegadores.
- Los documentos históricos conservan referencias al logo anterior como evidencia; no son usos ejecutables.

## Alcance y reglas

### 1. Correos transaccionales

1. Sustituir `plantillas/logo.png` por una composición derivada de `frontend/public/trace.svg`: isotipo vigente más nombre **Stockion** en blanco sobre transparencia, conservando su lectura en el encabezado índigo.
2. Conservar el lienzo PNG de 440 × 96 px y su proporción actual, sin deformar ni recolorear el isotipo. El margen del SVG no debe volverlo ilegible al reducirlo.
3. Mantener el PNG embebido por CID y `go:embed`. La conversión ocurre al preparar el asset, nunca durante el arranque, compilación o envío. El backend debe seguir compilando sin acceso a `frontend/`.
4. Aplicar el cambio desde el recurso compartido, cubriendo los cuatro tipos de correo sin duplicar logos por plantilla.
5. Mantener textos, asuntos, enlaces, OTP, vigencias, remitentes, transporte y versiones de texto plano. No rediseñar la paleta ni el contenido.
6. No agregar dependencias de ejecución ni un generador permanente. Registrar en esta capability el comando y herramienta de conversión usados para reproducir los derivados.

### 2. Aplicación y recursos anteriores

1. Cambiar la imagen de aceptar invitación a `trace.svg`, conservando dimensiones y nombre accesible del enlace de marca.
2. Confirmar por búsqueda que ningún consumidor activo referencia `stockion-logo.svg`; después retirar ese archivo y el `favicon.svg` obsoleto sin referencias. No alterar referencias históricas en documentos.
3. Regenerar `favicon.ico` a partir de `trace.svg`, conservando tamaños de 16, 32 y 48 px para solicitudes directas o automáticas a esa ruta.
4. Corregir únicamente el tipo del favicon declarado en `index.html` a `image/svg+xml`, conservando `trace.svg` como fuente.
5. Mantener `trace.svg` como original de identidad. PNG e ICO son derivados de distribución, sin modificaciones manuales al dibujo.

## Contratos y límites

- `Logo() ([]byte, error)`, `LogoCID = "logo@stockion"`, nombre `logo.png` y MIME `image/png` se conservan.
- `NewRenderizador`, `Render`, `SMTPTransporte` y los contratos de aplicación de correo mantienen sus firmas y comportamiento.
- Sin cambios de API, base de datos, permisos, cookies, tokens, URLs de invitación ni configuración SMTP.
- Una URL pública del frontend o un SVG remoto no sustituye el recurso CID de los correos.
- Los correos ya enviados contienen su propia imagen y no se actualizan. La revisión visual debe usar mensajes nuevos.
- Publicación, despliegue y envíos a destinatarios externos quedan fuera del alcance. Las verificaciones de envío usan SMTP local o Mailpit y datos de prueba.
- No modificar archivos del trabajo activo de capability 24. `main.go` es un consumidor revisado que no necesita cambios para esta corrección.

## Archivos previstos

- `backend/internal/infrastructure/correo/plantillas/logo.png`: reemplazo del asset común.
- `backend/internal/infrastructure/correo/renderizador_test.go`: ampliar la prueba existente para decodificar el PNG y comprobar dimensiones; conservar la cobertura de los cuatro correos. La identidad del dibujo se comprueba visualmente, no solo por firma PNG.
- `backend/internal/infrastructure/correo/smtp_transporte_test.go`: reutilizar la prueba existente de MIME y CID; ajustar únicamente si se detecta una carencia relacionada.
- `frontend/src/app/features/invitacion/aceptar-invitacion.html`: referencia al logo vigente.
- `frontend/src/index.html`: MIME del favicon SVG.
- `frontend/public/stockion-logo.svg` y `frontend/public/favicon.svg`: retiro de recursos obsoletos después de comprobar consumidores.
- `frontend/public/favicon.ico`: derivado actualizado de `trace.svg`.
- Esta capability y `PLAN_IMPLEMENTACION.md`: tareas, decisiones, resultados y estado. Actualizar `capabilities/planes-actuales.md` al entregar; las bases de registro y colaboradores solo requieren cambios si cambia alguna regla o flujo, lo cual no se propone aquí.

## Pruebas y criterios de aceptación

- [x] Los cuatro correos nuevos muestran el isotipo de `trace.svg` y el nombre Stockion, sin el hexágono anterior, deformación ni recortes.
- [x] El PNG se decodifica correctamente, mide 440 × 96 px y conserva transparencia.
- [x] La prueba de transporte conserva `multipart/related`, `multipart/alternative`, `image/png` y `Content-Id: <logo@stockion>`; no se requieren imágenes remotas.
- [x] Ejecutar desde `backend/`: `go test ./internal/infrastructure/correo ./internal/application/correo`.
- [x] Ejecutar desde `frontend/`: `npm run build`. Completado con salida temporal; ver resultado y limitación de permisos abajo.
- [x] Buscar referencias a `stockion-logo.svg` y `favicon.svg` en código/configuración activos y confirmar cero consumidores antes de retirarlos; confirmar `trace.svg` en las pantallas de marca revisadas.
- [x] Revisar en navegador la aceptación de invitación y el favicon actualizado.
- [x] Revisar mensajes nuevos de los cuatro tipos con SMTP local/Mailpit a 600 y 360 px: logo legible sobre encabezado índigo y contenido sin cambios. No usar cuentas reales para esta prueba.
- [x] Registrar pendientes de Gmail, Outlook y Apple Mail como no comprobados; Chromium/Mailpit no certifica su renderizado.
- [x] Tras modificar código, ejecutar `graphify update .` y registrar resultados sin descartar cambios ajenos.

## Tareas y autorización

- [x] Revisar grafo, consumidores de marca, renderizador, plantillas y transporte de correo.
- [x] Preparar capability independiente y enlazarla desde el índice general.
- [x] Obtener aprobación explícita del usuario; estado `Aprobada` registrado antes de comenzar implementación (2026-10-22).
- [x] Generar PNG e ICO desde `trace.svg` y registrar procedimiento.
- [x] Actualizar invitación, corregir MIME y retirar recursos anteriores.
- [x] Ejecutar pruebas y revisión visual; documentar resultados y pendientes.
- [x] Actualizar índices de estado y presentar la entrega para revisión.
- [ ] Recibir aceptación final explícita; no se declara cerrada ni se archiva mientras esta entrega siga en revisión del usuario.

## Decisiones y verificaciones registradas

- 2026-10-22: el usuario indicó `trace.svg` como logo vigente y pidió otra capability, independiente de la 24.
- 2026-10-22: se conserva el transporte PNG/CID existente; cambiar el dibujo compartido basta para los cuatro correos.
- 2026-10-22: `graphify query` ubicó `Logo()`, `Renderizador`, `SMTPTransporte` y el servicio de correo; la lectura dirigida confirmó los consumidores y referencias descritos.
- 2026-10-22: revisión estática realizada. Sin implementación, pruebas de aplicación, generación de imágenes, envíos de correo ni cambios Git de escritura en esta preparación documental.
- 2026-10-22: implementación aprobada explícitamente por el usuario; solo se modificaron los recursos y consumidores de marca previstos, la prueba del PNG y los índices. Trabajo de capability 24 preservado.

## Procedimiento de generación de recursos

Ejecutado desde la raíz con `rsvg-convert 2.62.3`, ImageMagick `7.1.2-31` y Liberation Sans Bold instalada en el sistema. Sin instalar dependencias ni crear un generador permanente. `-trim` elimina únicamente espacio transparente del original; el isotipo conserva proporción y colores. El recorte visible mide 823 × 890 px; el ICO se centra en un lienzo cuadrado de 890 px antes de reducirlo.

```bash
rsvg-convert frontend/public/trace.svg -o /tmp/tienda-cap25-trace.png
magick /tmp/tienda-cap25-trace.png -trim +repage /tmp/tienda-cap25-isotipo.png
magick /tmp/tienda-cap25-isotipo.png -resize 84x84 -gravity center -background none -extent 88x96 /tmp/tienda-cap25-marca.png
magick -size 440x96 xc:none /tmp/tienda-cap25-marca.png -gravity northwest -composite -font /usr/share/fonts/liberation/LiberationSans-Bold.ttf -pointsize 56 -fill white -annotate +108+16 Stockion backend/internal/infrastructure/correo/plantillas/logo.png
magick /tmp/tienda-cap25-isotipo.png -gravity center -background none -extent 890x890 -define icon:auto-resize=48,32,16 frontend/public/favicon.ico
```

`frontend/public/trace.svg` no se modificó. SHA-256 del original: `f943ebadf2031c8a6312ebc9a3667a104241cf424af74788a3d4ec192e03b2c9`.

## Resultados de verificación — 2026-10-22

- **Go:** `go test ./internal/infrastructure/correo ./internal/application/correo` correcto. La prueba existente ahora decodifica el PNG y comprueba 440 × 96 px y transparencia; se conservan las pruebas de los cuatro renderizados, escape de datos y MIME/CID del transporte.
- **Build Angular:** `npm run build` compiló, pero falló al reemplazar `frontend/dist/frontend/browser/media/primeicons-67GZ63CG.woff` con `EACCES` (archivo existente propiedad de `nobody`). Se repitió correctamente con Node `24.22.0`: `env PATH=/home/thrs/.local/share/mise/installs/node/lts/bin:$PATH npm run build -- --output-path /tmp/tienda-cap25-build`. No se cambiaron permisos ni se borró el `dist` compartido. Quedan avisos de presupuesto de bundle inicial (758.86 kB frente a advertencia de 500 kB) y estilos de componentes ajenos a este cambio de logo; no son errores de compilación.
- **Mailpit real local:** prueba temporal con overlay de Go, sin agregar archivos ejecutables al repositorio: `go test -overlay=/tmp/tienda-cap25-overlay.json ./internal/infrastructure/correo -run TestCap25Mailpit -count=1 -v`. Los cuatro casos pasaron y se capturaron cuatro mensajes nuevos para `cap25@stockion.invalid` en SMTP `127.0.0.1:1025`.
- **Recepción MIME:** se descargaron los cuatro mensajes desde Mailpit local y se decodificaron. Cada uno conserva HTML y texto plano, un único PNG con `Content-ID: <logo@stockion>` y bytes idénticos al nuevo asset del backend.
- **Chromium:** capturas de los cuatro correos recibidos a 600 y 360 px; todos cargan el PNG de 440 px y no presentan desbordamiento horizontal. Revisión visual de logo, proporción, colores, nombre y contenido correcta. Para visualizar en navegador fuera de Mailpit, únicamente el CID recibido se resolvió al mismo PNG mediante una URL de datos en la copia temporal.
- **Invitación y favicon:** build real servido localmente, consulta pública de invitación y sesión simuladas con datos de prueba; sin crear cuentas ni tocar base de datos. Capturas a 600/360 px, `trace.svg` cargado, enlace de marca accesible como Stockion y favicon declarado `image/svg+xml`. Los recursos retirados responden 404 y el ICO nuevo está disponible, con tamaños 48, 32 y 16 px verificados mediante ImageMagick.
- **Referencias:** no quedan consumidores de `stockion-logo.svg` ni `favicon.svg` en código/configuración activos revisados. Las referencias históricas documentales se preservan.
- **Grafo:** `graphify update .` terminó correctamente: 3170 nodos, 10635 relaciones y 130 comunidades. La extracción usó procesamiento secuencial por restricciones de procesos del entorno; sin API de pago.
- **Evidencias temporales:** [correos en escritorio](/tmp/tienda-cap25-qa/correos-escritorio.png), [correos en móvil](/tmp/tienda-cap25-qa/correos-movil.png) y [aceptar invitación en móvil](/tmp/tienda-cap25-qa/aceptar-invitacion-360.png). Datos y scripts temporales permanecen bajo `/tmp/tienda-cap25-*`; no forman parte de la aplicación.

## Pendientes y límites de la entrega

- Aceptación final del usuario; sin cierre automático.
- No se reinició el backend activo para evitar interrumpir el trabajo compartido de capability 24. Como `logo.png` se incrusta con `go:embed`, los correos enviados por ese proceso usarán el asset anterior hasta recompilarlo y reiniciarlo. Las pruebas anteriores usan un binario compilado con el nuevo recurso.
- Los correos enviados anteriormente conservan su imagen original; comprobar con un correo nuevo tras reiniciar.
- Gmail, Outlook, Apple Mail y modo oscuro en clientes reales no comprobados. No hubo envíos externos, cambios de configuración, despliegue ni operaciones Git de escritura.
- No se repitieron suites completas de otros dominios: el cambio reutiliza contratos y flujo de envío existentes; build y pruebas focalizadas están comprobados.
