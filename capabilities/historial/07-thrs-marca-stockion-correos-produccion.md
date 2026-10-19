# Plan 7: Marca Stockion, correos y envío a producción

## Estado y control

- Estado: `Aprobada`; aprobación del usuario el 2026-10-05 con alcance reducido: **solo código**. Sin Docker ni archivos de configuración; sin tocar identificadores técnicos, producción ni el trabajo de Ilu. El envío a `production` espera su indicación.
- Fecha: 2026-10-05.
- Archivo directo en `capabilities/thrs`, igual que los planes 1 a 6.
- Dependencias: correo transaccional (capability 07 general), sistema visual de la opción G (plan 4), producción en Lightsail (capability 06).
- Toca Git y una rama compartida: la parte D exige confirmación final explícita, según `AGENTS.md`.

## Objetivo

Que todo lo que ve una persona diga **Stockion** y se vea como Stockion: logo, pestaña del navegador y los cuatro correos. Después, dejar la rama `production` con estos cambios, compilada y probada.

## Hallazgos

### Logo

- `frontend/public/stockion-logo.svg`: logo **temporal** del plan 4 (hexágono índigo con caja). Lo usan la topbar, el asistente y la aceptación de invitación.
- `backend/internal/infrastructure/correo/plantillas/logo.png` (440 × 96 px): icono azul de tienda con la palabra **«Tienda» dibujada dentro de la imagen**. Por eso el correo dice «Tienda» aunque el texto ya diga Stockion.
- `frontend/public/favicon.ico`: icono genérico anterior a la marca.
- La rama `origin/ilu` tiene un commit sin integrar (`13f410d favicon`) que ya cambia el título a «Stockion», borra `favicon.ico` y agrega `trace.svg` como icono. Hay que decidir si se integra o se rehace.

### Correos

Cuatro correos, cada uno en HTML y texto, sobre una base común:

| Plantilla | Asunto actual | Cuándo se envía |
| --- | --- | --- |
| `verificacion` | Tu código de verificación de {app} | Registro y reenvío de código |
| `invitacion` | {negocio} te invitó a {app} | Invitar o reenviar invitación |
| `recuperacion` | Restablece tu contraseña de {app} | Solicitar recuperación |
| `contrasena_cambiada` | Tu contraseña de {app} cambió | Tras cambiar contraseña |

- `base.html.tmpl` usa la paleta anterior: fondo gris azulado `#f8fafc`, texto `#111c32`, bordes `#e2e8f0`.
- Botones y enlaces en azul `#2864e9` / `#1d4fc5` (6 apariciones). La app ya no usa ese azul desde el plan 4.
- El nombre visible sale de `MAIL_FROM_NAME`. En producción ya vale `Stockion`; **el valor por omisión en código sigue siendo `Tienda`** (`internal/config/config.go:116`), igual que en `.env.example` y `docker-compose.yml`.
- El logo viaja embebido por CID (`logo@tienda`); el mecanismo funciona y se conserva.

### «Tienda» visible

| Dónde | Qué dice |
| --- | --- |
| `frontend/src/index.html` | `<title>Tienda</title>` (pestaña del navegador) |
| `logo.png` de correo | Palabra «Tienda» en la imagen |
| `config.go`, `.env.example`, `docker-compose.yml` | `MAIL_FROM_NAME` por omisión `Tienda` |
| `internal/database/database.go:147` | Sucursal de desarrollo «Tienda prueba» |
| `AGENTS.md`, `DOCUMENTACION.md`, `frontend/README.md`, `PLAN_IMPLEMENTACION.md` | Títulos «proyecto Tienda», «Documentación de Tienda» |
| Pruebas de correo y configuración | `NombreApp: "Tienda"`, asuntos esperados «…de Tienda» |

Dentro de la aplicación no queda ningún texto «Tienda» en pantallas: topbar, acceso y asistente ya dicen Stockion.

### «tienda» técnico (no lo ve nadie)

| Identificador | Apariciones | Efecto de renombrarlo |
| --- | --- | --- |
| Módulo Go `tienda/backend` | 169 | Cambia cada `import`; conflictos seguros con otras ramas |
| Cookie `tienda_session` | 16 | **Cierra la sesión de todos** al desplegar |
| Claves de navegador `tienda.contexto.*`, `tienda.verification.*`, `tienda.invitacion.*`, `tienda.sidebar.*`, `tienda.alta_inicial.*` | 7 claves | Se pierde negocio y sucursal elegidos; una verificación o invitación en curso se interrumpe |
| Dominio `tienda.mergemakers.com` | 12 | Es infraestructura (DNS, túnel, correos ya enviados); fuera de código |
| Clases CSS `tienda-alerta*` | 30 | Ninguno visible |
| CID `logo@tienda`, preset `TemaTienda` | 2 | Ninguno visible |
| Carpeta del repositorio, base de datos local `tienda` | — | Rompe rutas y entornos locales |

### Producción

- `thrs` está **10 commits por delante** de `origin/production` y 0 por detrás: se puede integrar sin conflictos (avance directo).
- `compose.production.yaml` ya se llama `stockion` y usa `MAIL_FROM_NAME` por omisión `Stockion`.
- El despliegue al servidor lo hace el usuario con `pasos-update-server.md`; este plan no despliega.
- Al arrancar la versión nueva en producción correrán, por primera vez, las migraciones de los planes 5 y 6:
  - apellidos del perfil (`apellidos` completo pasa a `primer_apellido`);
  - índice único de correo por empleado (se omite y avisa si hay duplicados);
  - permisos `precios.ver` y `precios.autorizar`;
  - **filtro de sucursales por asignación**: un miembro activo sin asignación deja de ver sucursales.
- Pruebas que fallan hoy:
  - 8 en `productos.component.spec.ts` (anteriores a estos planes).
  - `TestProductionAcceptsStrongOTPSecret` en `internal/config`: espera 5 intentos de OTP y lee 10. `Load()` toma el `backend/.env` local; hay que confirmar si es solo por el entorno local o un defecto real antes de enviar a producción.

## Alcance

### A · Logo

1. Logo definitivo en SVG como única fuente. Si el usuario no entrega archivo, se pule el actual (ver decisiones).
2. Derivados:
   - `frontend/public/stockion-logo.svg` (isotipo, el que ya consumen las pantallas).
   - `frontend/public/favicon.svg` más `favicon.ico` de respaldo; `index.html` los referencia.
   - `backend/.../plantillas/logo.png`: isotipo + palabra «Stockion», 440 × 96 px (220 × 48 a doble densidad), fondo transparente, legible sobre blanco.
3. `alt` del logo en correos y pantallas: «Stockion».

### B · Correos con la identidad de la opción G

Misma estructura de tablas (compatible con Gmail, Outlook y Apple Mail); cambia el aspecto.

| Elemento | Hoy | Propuesto |
| --- | --- | --- |
| Fondo exterior | `#f8fafc` | Crema `#faf8f3` |
| Tarjeta | Blanca, borde `#e2e8f0`, radio 16 | Blanca, borde `#e6e0d2`, radio 16 |
| Encabezado | Logo suelto | Franja índigo `#2c2e6a` con logo claro, como la topbar |
| Texto | `#111c32` / `#475569` | `#172554` / `#5b6478` |
| Botón | Azul `#2864e9` | Violeta `#4f46c8`, texto blanco, radio 10, alto 46 |
| Enlaces | `#1d4fc5` | Violeta `#3f37a8` |
| Código OTP | Texto grande | Bloque lavanda `#eceafb` con borde `#dcd8f6`, cifras espaciadas |
| Avisos | Texto plano | Bloque salvia o mantequilla, como en la app |
| Pie | Gris | `#5b6478` sobre crema, con «Stockion» y año |

- Tipografía: Arial/Helvetica como hoy; los clientes de correo no cargan Inter de forma fiable.
- Texto previo (preheader) oculto por correo, para la vista de la bandeja.
- Cada plantilla mantiene sus variables y su versión de texto plano; no cambian contratos ni el servicio de envío.
- Contenido por correo:
  - **Verificación**: saludo, código en bloque, vigencia en minutos, «si no fuiste tú».
  - **Invitación**: quién invita, a qué empresa y sucursal si el dato está disponible, botón, vigencia en horas, enlace copiable.
  - **Recuperación**: botón, vigencia, enlace copiable, «si no lo pediste».
  - **Contraseña cambiada**: confirmación y qué hacer si no fue la persona.
- Sin colores ni tokens nuevos: son los de `frontend/src/styles.css`.

### C · «Tienda» pasa a «Stockion» en lo visible

- `index.html`: `<title>Stockion</title>`.
- `MAIL_FROM_NAME` por omisión `Stockion` en `config.go`, `.env.example` y `docker-compose.yml`.
- Sucursal de desarrollo «Tienda prueba» pasa a «Sucursal de prueba» (solo afecta bases locales nuevas).
- Títulos de `AGENTS.md`, `DOCUMENTACION.md`, `frontend/README.md` y `PLAN_IMPLEMENTACION.md`.
- Pruebas de correo y configuración actualizadas al nombre nuevo.
- Identificadores internos sin efecto: CID `logo@stockion`, preset `TemaStockion`.

**No se renombra** (recomendado): módulo Go, cookie de sesión, claves de navegador, dominio, clases `tienda-alerta`, carpeta ni base local. Ver decisiones.

**No se toca**: nombres de ejemplo en pruebas como «Tienda Centro» o «Tienda Ana». Son negocios ficticios de tipo tienda, no la marca. Tampoco los documentos históricos en `capabilities/`.

### D · Verificar y enviar a `production`

1. `go vet ./...` y `go test ./...`. Aclarar primero `TestProductionAcceptsStrongOTPSecret`.
2. `ng test` y `ng build` de producción con Node 24.
3. `docker compose -f compose.production.yaml build` en local, para confirmar que las imágenes de producción compilan.
4. Recorrido en base temporal (`DB_NAME=tienda_e2e`) con Mailpit: disparar los cuatro correos y revisar HTML, texto plano y logo. Capturas a 600 px y 360 px.
5. Revisar pestaña del navegador y favicon.
6. Reporte previo para producción (solo lectura): miembros activos sin asignación y correos de empleado duplicados. El usuario lo ejecuta en el servidor o autoriza hacerlo.
7. **Con confirmación explícita del usuario**: commit en `thrs`, integrar en `production` por avance directo y `git push`. Sin `--force`. Sin atribución a Claude en commits.
8. El despliegue en Lightsail queda para el usuario (`pasos-update-server.md`).

## Archivos previstos

- Frontend: `public/stockion-logo.svg`, `public/favicon.svg`, `public/favicon.ico`, `src/index.html`, `src/app/app.config.ts`.
- Backend: `infrastructure/correo/plantillas/*` (base, 4 HTML, 4 texto, `logo.png`), `infrastructure/correo/renderizador.go`, `config/config.go`, `database/database.go`, pruebas de `correo` y `config`.
- Raíz: `.env.example`, `docker-compose.yml`, `AGENTS.md`, `DOCUMENTACION.md`, `frontend/README.md`, `PLAN_IMPLEMENTACION.md`.

## Riesgos

- Outlook de escritorio ignora radios y algunos rellenos: el diseño debe verse correcto también sin ellos.
- Un logo PNG con transparencia sobre la franja índigo necesita versión clara; se prueba en modo oscuro de Gmail y Apple Mail.
- Chocar con el commit `favicon` de `origin/ilu` (toca `index.html`, `favicon.ico` y `package-lock.json`).
- Enviar a `production` arrastra los planes 1, 5 y 6, todavía «pendientes de aceptación final», y sus migraciones.
- Miembros sin asignación perderán acceso a sucursales al desplegar.

## Tareas

- [ ] Aprobar el alcance y resolver las decisiones pendientes.
- [ ] A: logo definitivo y derivados.
- [ ] B: base y cuatro correos.
- [ ] C: nombre visible, valores por omisión y documentos.
- [ ] D: verificación completa y reporte previo.
- [ ] D: commit, integración en `production` y `push`, **solo con confirmación explícita**.

## Criterios de aceptación

- La pestaña del navegador dice «Stockion» y muestra el icono nuevo.
- Los cuatro correos muestran el logo de Stockion, sin la palabra «Tienda» en imagen, asunto ni cuerpo, y usan índigo, violeta y crema.
- Un entorno nuevo sin `MAIL_FROM_NAME` envía correos como «Stockion».
- `grep -rn "Tienda"` sobre código ejecutable solo devuelve nombres de ejemplo en pruebas.
- Sesiones abiertas, negocio y sucursal seleccionados sobreviven a la actualización.
- `go test`, `ng test` y `ng build` sin fallos nuevos; las imágenes de producción compilan.
- `production` contiene los cambios y coincide con `thrs`.

## Decisiones pendientes

- **Logo**: el usuario entrega el archivo definitivo (recomendado) o se pule el temporal actual y se dibuja su versión con texto.
- **Favicon**: integrar el commit de `origin/ilu` con `trace.svg` o usar el isotipo de Stockion (recomendado, una sola marca).
- **Identificadores técnicos** (`tienda_session`, claves `tienda.*`, módulo Go): dejarlos (recomendado; nadie los ve y cambiarlos cierra sesiones y genera conflictos) o renombrarlos también.
- **Dominio `tienda.mergemakers.com`**: fuera de alcance (recomendado) o planear un dominio nuevo aparte.
- **Nombres de ejemplo en pruebas** («Tienda Centro»): dejarlos (recomendado) o cambiarlos.
- **Qué entra a `production`**: todo `thrs` (planes 1, 5, 6 y 7) o solo el plan 7. Lo segundo exige separar commits.
- **Fallos previos** (8 pruebas de Productos y la de configuración OTP): corregirlos antes de enviar (recomendado para la de configuración) o enviarlos como están.

## Seguimiento

### Decisiones del usuario (2026-10-05)

- Identificadores técnicos y producción: no se tocan.
- Trabajo de Ilu (`origin/ilu`, commit `favicon`): no se integra por ahora.
- Logo: pulir el temporal.
- Solo código: sin Docker ni archivos de configuración.
- A `production` va esta versión completa de `thrs`, cuando el usuario lo indique.

### Implementado (2026-10-05)

- **Logo**: `frontend/public/stockion-logo.svg` redibujado (caras separadas, sin trazos finos, cinta violeta); `favicon.svg` y `favicon.ico` de 16, 32 y 48 px generados del mismo isotipo; `logo.png` de correo con isotipo sobre cuadro blanco y la palabra «Stockion» en blanco, 440 × 96 px con transparencia.
- **Pestaña**: `<title>Stockion</title>`, icono SVG con respaldo `.ico` y `theme-color` índigo.
- **Correos**: base con franja índigo, fondo crema, tarjeta con borde arena y pie crema; botones violeta; código OTP en bloque lavanda; vigencia de la invitación en aviso mantequilla; alerta roja en «contraseña cambiada»; texto previo para la bandeja en los cuatro.
- **Nombre**: `MAIL_FROM_NAME` por omisión `Stockion` en `config.go`; sucursal de desarrollo «Sucursal de prueba»; CID `logo@stockion`; preset `TemaStockion`; pruebas de correo actualizadas.

### No tocado, por alcance

- `.env.example` y `docker-compose.yml` siguen con `MAIL_FROM_NAME` por omisión `Tienda` (configuración y Docker). Producción ya define `Stockion`.
- Títulos de `AGENTS.md`, `DOCUMENTACION.md`, `frontend/README.md` (documentos, no código).
- Cookie, claves de navegador, módulo Go, dominio y clases `tienda-alerta`.
- Compilación de imágenes de producción (Docker).

### Pruebas

- `go vet` limpio. `go test ./...`: pasa todo salvo `TestProductionAcceptsStrongOTPSecret`.
  - Causa confirmada: no es el entorno local. El commit `a827745` (2026-09-30) subió el valor por omisión de `OTP_MAX_ATTEMPTS` de 5 a 10 en `config.go` y la prueba sigue esperando 5. Es anterior a estos planes; decidir cuál de los dos valores es el correcto antes de producción.
- `ng build` correcto; `ng test`: 176 pasan, 8 fallan en `productos.component.spec.ts` (previos).
- Recorrido real en base temporal con Mailpit: los cuatro correos se enviaron con remitente y asunto «Stockion», logo embebido, y sin la palabra «Tienda» en HTML ni en texto plano. Capturas en `capturas-plan7/` (600 px y 360 px).

### Pendientes

- Revisión en clientes reales (Gmail, Outlook, Apple Mail) y en modo oscuro: solo se vio en Chromium.
- Parte D: reporte previo en producción, commit, integración en `production` y `push`, a la espera de la indicación del usuario.
