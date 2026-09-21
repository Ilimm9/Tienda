# Capability 05: Preparación segura para producción en AWS Lightsail

## Estado

En implementación (fases 1 y 2).

## Aprobación

Las fases 1 y 2 de [PLAN_APLICACION.md](PLAN_APLICACION.md) fueron aprobadas explícitamente por el usuario el 2026-09-20 mediante la instrucción: “puedes comenzar con la fase 1 y fase 2”. Las fases 3 a 7, la infraestructura y el despliegue continúan pendientes de aprobación.

## Plan inmediato de aplicación

El diseño detallado y ordenado de los cambios de código vive en [PLAN_APLICACION.md](PLAN_APLICACION.md). Ese plan prioriza rutas y aislamiento multiempresa, reemplaza JWT por sesiones opacas revocables y después incorpora verificación de correo mediante OTP y Amazon SES. Lightsail queda fuera de esas primeras fases.

## Propósito

Preparar Tienda para un primer despliegue público en una instancia Linux de AWS Lightsail con controles preventivos, detectivos y de recuperación. El objetivo no es prometer una aplicación “imposible de hackear”, sino reducir la superficie de ataque, limitar el impacto de una intrusión, detectar abuso y poder recuperar el servicio y los datos.

La capability también funciona como guía reutilizable para evaluar cualquier aplicación web antes de publicarla.

## Decisión arquitectónica propuesta

Para el primer lanzamiento se propone una arquitectura de costo contenido en una sola instancia Lightsail, con Docker Compose y el frontend y API bajo el mismo origen. La base de datos puede iniciar en la misma instancia sólo si permanece inaccesible desde Internet y existe respaldo cifrado fuera de la instancia, probado mediante restauración.

```text
Internet
   |
   |  HTTPS 443 (HTTP 80 sólo para redirección y renovación)
   v
Firewall Lightsail + firewall del host
   |
   v
Proxy inverso con TLS, límites y cabeceras
   |-------------------|
   v                   v
Frontend estático      API Go (red privada Docker)
                           |
                           v
                       PostgreSQL (red privada, sin puerto público)
                           |
                           v
                  Respaldo cifrado fuera de la instancia
```

Esta topología no ofrece alta disponibilidad: la instancia, su zona y su volumen siguen siendo puntos únicos de falla. Cuando las ventas dependan continuamente del sistema o el costo de una interrupción supere el ahorro de una sola VM, se deberá separar la base de datos, añadir redundancia y colocar protección administrada frente a tráfico abusivo o ataques volumétricos.

## Supuestos y objetivos operativos

- Producción usará un dominio propio, HTTPS obligatorio y un origen único para navegador y API.
- Sólo `80/tcp` y `443/tcp` serán públicos. `22/tcp` se limitará a IPs administrativas autorizadas o se abrirá temporalmente. PostgreSQL, API, Docker y herramientas administrativas no serán públicos.
- La cuenta AWS tendrá MFA, usuarios individuales y permisos mínimos; no se usará la cuenta root para operación diaria.
- Objetivo inicial de recuperación: RPO máximo de 1 hora y RTO máximo de 4 horas. Si Tienda procesa ventas reales, estos objetivos deberán reevaluarse antes de considerar el servicio crítico.
- Los ambientes local, pruebas y producción tendrán datos, credenciales y configuración separados.
- Un lanzamiento sólo será permitido si todos los controles P0 están cerrados y verificados.

## Modelo de amenazas mínimo

### Activos a proteger

- Credenciales, sesiones y datos personales de usuarios.
- Datos de negocios, sucursales, empleados, catálogo, inventario y futuras ventas.
- Claves JWT, contraseñas de base de datos, llaves de proveedores y credenciales AWS.
- Integridad de los roles y del aislamiento entre negocios.
- Disponibilidad del servicio, respaldos, logs y capacidad de recuperación.

### Adversarios y fallas consideradas

- Bots de Internet que buscan puertos, credenciales débiles y software vulnerable.
- Atacantes anónimos contra registro, login, invitaciones, cargas y endpoints públicos.
- Usuario autenticado que intenta actuar sobre otro negocio o elevar privilegios.
- Robo de cookie, secreto, llave SSH o credencial AWS.
- Dependencia, imagen de contenedor o pipeline comprometido.
- Consumo intencional o accidental de CPU, memoria, disco, conexiones o APIs de terceros.
- Error operativo, migración defectuosa, corrupción o eliminación de datos.

## Método y límites de la revisión actual

Revisión estática realizada el 2026-09-19 sobre la configuración y el código presentes en el repositorio. Se inspeccionaron rutas HTTP, autenticación, autorización, Docker, Nginx, configuración, secretos, dependencias declaradas y pruebas existentes. No se realizó pentest, DAST contra una instancia, escaneo de imágenes, análisis de historial Git, prueba de restauración ni revisión de una cuenta AWS real.

Los hallazgos describen evidencia visible hoy. La ausencia de un hallazgo no demuestra ausencia de vulnerabilidades.

## Controles positivos existentes que deben conservarse

- Contraseñas almacenadas con bcrypt y comparación con mensaje general para credenciales inválidas.
- Cookie de sesión `HttpOnly`; el atributo `Secure` se activa cuando `APP_ENV=production`.
- JWT restringido explícitamente a HS256 y con expiración requerida.
- CORS compara un origen configurado y no usa comodín con credenciales.
- El servicio backend no publica un puerto al host en el Compose actual.
- Las cargas de catálogo observadas limitan el cuerpo a 5 MiB.
- Clientes HTTP externos tienen timeout de 7 segundos y exigen URLs HTTPS para imágenes.
- Varias operaciones de negocio, roles, sucursales, empleados e invitaciones aplican permisos en la capa de aplicación.
- Los tokens de invitación se generan aleatoriamente y se persisten como hash.
- El contenedor del backend se ejecuta como usuario sin privilegios.
- `.env` y `backend/.env` no están rastreados por Git en el estado revisado.

## Hallazgos específicos de Tienda

### P0 — bloquean cualquier salida a Internet

| ID | Falla observada | Evidencia actual | Riesgo | Resultado requerido |
| --- | --- | --- | --- | --- |
| P0-01 | Endpoints administrativos globales de marcas, categorías y unidades están registrados directamente en el router público. Incluyen creación, edición e importación. | `backend/cmd/api/main.go`, rutas `/api/v1/catalogo/*` registradas fuera de un grupo con `RequireAuth`. | Un usuario anónimo puede alterar o cargar catálogos globales. | Inventariar todas las rutas y exigir autenticación, ámbito y permiso explícito; sólo salud y flujos deliberadamente públicos quedan exentos. |
| P0-02 | Las escrituras de productos y proveedores validan membresía activa, pero no un permiso RBAC específico. | Grupo `negocioActual` usa `RequireAuth` y `RequireNegocioActivo`; `product_handler.go` conserva comentarios de validación de permiso pendiente. `RequierePermiso` existe, pero no está conectado en `main.go`. | Cualquier miembro activo puede modificar datos aunque su rol sea de lectura. | Definir permisos de catálogo y aplicarlos en servidor a cada lectura/escritura sensible; probar denegación por rol y por otro negocio. |
| P0-03 | No existe terminación TLS en la configuración versionada; Nginx sólo escucha en el puerto 80. | `frontend/nginx.conf` y `frontend/Dockerfile`. | Credenciales y sesiones pueden viajar sin cifrado; la cookie segura no funcionará correctamente sobre HTTP. | Dominio, certificado válido, renovación automática, redirección HTTP→HTTPS, TLS moderno y HSTS activado después de validar HTTPS. |
| P0-04 | No hay una defensa CSRF explícita para la sesión basada en cookie ni `SameSite` configurado deliberadamente. | `auth_handler.go` usa `SetCookie`; no se encontró middleware/token CSRF ni validación de `Origin` para mutaciones. | Un sitio malicioso puede intentar inducir acciones con la sesión de la víctima; no se debe depender del comportamiento predeterminado del navegador. | Elegir y probar una estrategia: cookie `SameSite` explícita más validación estricta de origen y/o token CSRF para toda operación con efecto. |
| P0-05 | La configuración puede arrancar con secretos y credenciales predecibles si faltan variables. | `config.go` tiene fallback de JWT y base de datos; Compose tiene defaults `postgres`; `APP_ENV` puede quedar fuera de producción. | Una omisión de despliegue puede dejar firma de sesiones o base de datos comprometibles y activar comportamiento de desarrollo. | En producción, fallar al arrancar si falta, es débil o es placeholder cualquier secreto/configuración obligatoria; separar configuración local de producción. |
| P0-06 | Compose publica PostgreSQL en el host por defecto. | `docker-compose.yml`, `${DB_PUBLIC_PORT:-5433}:5432`. | Una regla de firewall accidental puede exponer la base de datos a Internet. | El manifiesto de producción no publicará PostgreSQL, API ni pgAdmin; la DB sólo aceptará conexiones de su red privada. |
| P0-07 | No existen respaldos externos ni evidencia de una restauración ensayada. | No se encontró configuración o runbook de backup/restore. | Una falla de disco, borrado, ransomware o error de migración puede destruir los datos del negocio. | Respaldo cifrado fuera de la instancia que cumpla RPO/RTO, retención definida, alertas y simulacro de restauración documentado. Los snapshots complementan, no sustituyen, el respaldo de datos. |

### P1 — necesarios para un lanzamiento responsable

| ID | Falla observada | Riesgo | Resultado requerido |
| --- | --- | --- | --- |
| P1-01 | Login, registro, consulta de invitación y endpoints costosos no tienen límites globales por IP/identidad. El bloqueo por cuenta no cubre abuso distribuido y permite bloquear a una víctima conocida. | Fuerza bruta, enumeración, spam, denegación de servicio y consumo de APIs de terceros. | Límites por IP y cuenta, ventanas y ráfagas definidas, respuesta `429`, espera creciente y alertas; límites adicionales en proxy. |
| P1-02 | La sesión es un JWT autocontenido y logout sólo borra la cookie. Deshabilitar una cuenta, cambiar contraseña o revocar acceso no invalida tokens ya emitidos. | Una sesión robada conserva acceso hasta expirar. | Sesiones revocables del lado servidor o versión de sesión verificada en cada solicitud; rotación, revocación por usuario/dispositivo y expiración absoluta/inactiva. |
| P1-03 | “Recordarme” cambia la duración de la cookie a 30 días, pero el JWT mantiene la expiración configurada, actualmente 24 horas. | Contrato inconsistente y falsa expectativa de persistencia; decisiones de duración no probadas. | Definir duraciones corta/larga coherentes y pruebas de expiración, renovación y revocación. |
| P1-04 | Login distingue “cuenta no disponible” de “credenciales inválidas” y registro confirma si un correo existe. | Enumeración de cuentas y de estados de bloqueo/deshabilitación. | Respuestas externas uniformes donde corresponda; detalle sólo en logs seguros y flujos de recuperación controlados. |
| P1-05 | El registro activa inmediatamente la cuenta y no hay verificación de correo, recuperación de contraseña ni MFA para cuentas privilegiadas. | Suplantación de correo y recuperación operacional insegura; mayor impacto del robo de contraseña. | Verificación de correo, recuperación con token de un solo uso, notificación de eventos sensibles y MFA al menos para administradores/propietarios. |
| P1-06 | La contraseña sólo exige mínimo de 8 caracteres y no establece máximo. | Contraseñas débiles y comportamiento problemático con entradas excesivas o el límite efectivo de bcrypt. | Política compatible con frases largas, máximo explícito acorde al algoritmo, lista de contraseñas comprometidas cuando sea viable y reautenticación para acciones críticas. |
| P1-07 | Se usa `gin.Default()`/`router.Run()` sin `http.Server` con `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` y límites generales de cuerpo/cabeceras. | Conexiones lentas o cuerpos grandes pueden agotar goroutines, memoria y sockets. | Timeouts y límites en proxy y API, cierre ordenado y pruebas de solicitudes lentas/grandes. |
| P1-08 | Nginx no configura cabeceras defensivas ni límites de tasa, conexiones o cuerpo. | Clickjacking, sniffing, abuso automatizado y mayor impacto de XSS. | Política CSP compatible, `X-Content-Type-Options`, `Referrer-Policy`, protección de framing, política de permisos, límites y ocultación de versión; validar que no rompa Angular. |
| P1-09 | Las migraciones se ejecutan automáticamente al iniciar la API y no existe proceso de despliegue/rollback versionado. | Un despliegue puede dejar esquema y binario incompatibles o ampliar una caída. | Migraciones versionadas, compatibles hacia atrás, ejecutadas como paso controlado después de backup; estrategia de rollback que no dependa de revertir datos destructivamente. |
| P1-10 | No existe un manifiesto de producción separado y endurecido. | Se pueden desplegar puertos, perfiles, valores y herramientas pensados para desarrollo. | Compose de producción con exposición mínima, usuarios no root, filesystem de sólo lectura cuando aplique, capacidades eliminadas, `no-new-privileges`, límites de CPU/memoria/PIDs y healthchecks reales. |
| P1-11 | No hay observabilidad de seguridad: request ID, logs estructurados, auditoría durable, métricas de aplicación ni alertas. | Ataques, errores y degradación pueden pasar inadvertidos; no habrá evidencia útil para responder. | Logs sin secretos/PII innecesaria, auditoría de login y cambios privilegiados, métricas, rotación/retención y alertas accionables. |
| P1-12 | No existe protección demostrable contra llenado de disco por logs, imágenes, temporales, DB o capas Docker. | La instancia puede caer y corromper operaciones aunque CPU y memoria parezcan sanas. | Cuotas/rotación, limpieza segura, alertas de disco e inodos y reserva operativa definida. |

### P2 — endurecimiento y madurez posterior al primer lanzamiento

| ID | Observación | Resultado requerido |
| --- | --- | --- |
| P2-01 | Imágenes base usan tags y pgAdmin usa `latest`; no hay escaneo de imágenes ni SBOM. | Fijar versiones y, para producción reproducible, digests; escanear imagen y dependencias antes de desplegar; documentar excepciones. |
| P2-02 | El frontend Nginx usa la imagen oficial sin endurecimiento explícito y normalmente corre como root para enlazar el puerto 80 dentro del contenedor. | Evaluar imagen unprivileged o capacidades mínimas y demostrar que ningún contenedor privilegiado monta el socket Docker. |
| P2-03 | Los `.env` locales revisados tienen modo `0644`, duplican secretos y contienen contraseñas locales predecibles. No están rastreados en Git, lo cual es positivo. | Permisos mínimos, una sola fuente por ambiente, rotación antes de producción y gestor/proceso de secretos que evite incorporarlos a imágenes, logs o respaldos. |
| P2-04 | Los clientes externos tienen timeout, pero crean solicitudes con `context.Background()` y decodifican respuestas sin límite visible. | Propagar cancelación del request, limitar respuesta, controlar redirecciones y aislar fallas con métricas/circuit breaker si el volumen lo exige. |
| P2-05 | La imagen remota de Unsplash crea una dependencia y tráfico de terceros desde pantallas de acceso/invitación. | Hospedar activos críticos bajo control propio y definir `Referrer-Policy`; evitar que URLs sensibles se filtren a terceros o logs. |
| P2-06 | La salud pública sólo responde que el proceso vive; no existe separación clara entre liveness, readiness y dependencias. | Liveness barata sin secretos; readiness interna que compruebe dependencias esenciales sin revelar topología; retirar instancias no listas. |
| P2-07 | No hay política de retención/borrado de datos personales, términos operativos ni inventario de terceros. | Definir minimización, conservación, eliminación, exportación, responsables y obligaciones aplicables antes de almacenar datos reales. |

## Alcance de implementación futura

Después de la aprobación explícita, el trabajo deberá dividirse y ejecutarse en este orden. Si una fase altera contratos de otra capability activa, se detendrá esa parte para revisión.

### Fase 0 — decisiones y línea base

- Clasificar datos y operaciones críticas.
- Confirmar dominio, región, volumen esperado, presupuesto, responsables y canal de alertas.
- Congelar objetivos RPO/RTO y política de mantenimiento.
- Crear inventario de endpoints con exposición, autenticación, ámbito de negocio, permiso, costo y límite de entrada.
- Registrar una línea base de pruebas, dependencias e imágenes sin corregir silenciosamente fallas.

### Fase 1 — cerrar autorización y autenticación P0

- Proteger rutas globales de catálogo y eliminar cualquier mutación anónima.
- Definir permisos de productos, proveedores, importación y catálogos; aplicar autorización en servidor.
- Probar aislamiento horizontal entre negocios y sucursales en repositorio, aplicación y HTTP.
- Adoptar sesión revocable, cookie con `Secure`, `HttpOnly`, `SameSite` y alcance mínimos.
- Incorporar CSRF, límites de login/registro y contratos externos no enumerables.
- Añadir verificación de correo, recuperación y controles reforzados para cuentas privilegiadas.

### Fase 2 — configuración, secretos y cadena de suministro

- Crear configuración de producción que falle cerrada y jamás use fallbacks de desarrollo.
- Generar secretos únicos de alta entropía, rotables y separados por ambiente.
- Definir dónde se almacenan, quién los puede leer, cómo se rotan y cómo se revocan.
- Asegurar que los secretos no estén en Git, imagen, argumentos de build, logs, frontend o artefactos.
- Ejecutar SCA para Go y npm, escaneo de secretos, SAST, escaneo de imágenes y SBOM.
- Fijar toolchains, paquetes e imágenes; definir cadencia y SLA de parches por severidad.

### Fase 3 — endurecimiento de la aplicación y del proxy

- Configurar timeouts, límites de cabeceras/cuerpo, concurrencia, tasa y conexiones.
- Añadir cabeceras defensivas y CSP incremental con modo de reporte antes de bloquear si es necesario.
- Normalizar errores sin filtrar consultas, rutas internas, secretos o stack traces.
- Añadir request/correlation ID, logs JSON, auditoría de eventos sensibles y redacción de datos.
- Definir timeouts de DB, pool de conexiones, cancelación de contexto y apagado ordenado.
- Mantener cargas con allowlist de formato, validación por contenido, nombres generados, límites por archivo/lote y procesamiento fuera de rutas públicas cuando crezca el volumen.

### Fase 4 — infraestructura Lightsail

- Crear cuenta/identidades AWS seguras: MFA root y administradores, sin access keys root, usuarios individuales y mínimo privilegio.
- Crear instancia Linux LTS soportada, IP estática, DNS y sincronización horaria.
- Configurar firewall Lightsail IPv4 e IPv6 y firewall del host. Abrir únicamente 80/443; restringir 22 por origen.
- Usar SSH por llave, deshabilitar contraseña y root remoto, limitar intentos y documentar acceso de emergencia.
- Aplicar actualizaciones de seguridad del sistema y Docker con ventana y política de reinicio.
- Desplegar proxy TLS con renovación automática y prueba de expiración.
- No publicar DB, API, pgAdmin, Docker socket ni puertos de métricas.
- Endurecer contenedores y establecer límites/rotación. Ejecutar como usuario no root siempre que sea viable.

### Fase 5 — datos, migraciones y recuperación

- Separar credenciales DB de administrador, migrador, runtime y backup; runtime con mínimo privilegio.
- Ejecutar migraciones como paso controlado, idempotente y observable.
- Cifrar respaldos en tránsito y reposo; almacenarlos fuera de la instancia y con credenciales limitadas.
- Usar respaldo de DB coherente con RPO de 1 hora, más snapshot automático diario como defensa adicional.
- Definir retención operativa inicial: 24 copias horarias, 14 diarias y 3 mensuales, sujeta a costo y obligaciones de datos.
- Alertar si el backup falla o envejece y hacer restauración completa trimestral en un ambiente aislado.
- Conservar al menos una copia que no desaparezca al eliminar la instancia; los snapshots automáticos de Lightsail se eliminan con el recurso de origen.

### Fase 6 — observabilidad, respuesta y operación

- Métricas: disponibilidad, latencia, tasa de errores, login fallido, `429`, CPU, memoria, burst capacity, red, disco, inodos, contenedores, conexiones DB, backup y expiración TLS.
- Alertas con umbral, destinatario, severidad y acción; probar el canal antes del lanzamiento.
- Logs con rotación y retención; protegerlos de manipulación y nunca registrar contraseñas, cookies, JWT, tokens de invitación ni cuerpos sensibles.
- Runbooks para servicio caído, CPU/memoria/disco altos, DB inaccesible, certificado próximo a vencer, secreto filtrado, cuenta comprometida y restauración.
- Procedimiento de incidente: contener, preservar evidencia, rotar, restaurar, comunicar, analizar causa y registrar acciones.
- Presupuesto y alertas de gasto/transferencia para detectar abuso y evitar sorpresas.

### Fase 7 — despliegue y puerta de producción

- Construir artefactos una vez y promover el mismo digest; no compilar manualmente en producción.
- Respaldar antes de cambios de esquema, comprobar espacio y ejecutar smoke tests.
- Desplegar con versión identificable y estrategia de rollback del binario/configuración.
- Validar desde fuera de AWS: DNS, cadena TLS, redirecciones, cabeceras, endpoints públicos, puertos y flujos críticos.
- Observar métricas y logs durante una ventana definida; revertir ante criterios objetivos.
- Registrar versión, migraciones, operador, hora, resultado y pendientes.

## Contratos de seguridad

### Red y transporte

- Todo tráfico de usuario se sirve por HTTPS; HTTP sólo redirige a HTTPS.
- La API confía en cabeceras de proxy únicamente cuando provienen del proxy controlado.
- Ningún puerto de datos o administración responde públicamente.
- TLS y certificados se renuevan automáticamente y generan alerta antes de expirar.

### Identidad, sesión y autorización

- La denegación es el valor por defecto. Toda ruta declara si es pública y por qué.
- Autenticación, membresía de negocio y permiso son verificaciones independientes.
- Cada acceso a recurso multi-tenant restringe la consulta por el `negocio_id` autorizado; no basta con ocultar botones o validar UUID.
- Cambiar contraseña, deshabilitar usuario, revocar membresía o cerrar sesiones invalida el acceso activo dentro del tiempo definido.
- Cookies de sesión nunca son accesibles desde JavaScript ni viajan por HTTP.
- Acciones destructivas o privilegiadas requieren autorización específica y, según riesgo, reautenticación/MFA.

### Entrada, salida y consumo de recursos

- Cada endpoint tiene límites de tamaño, tiempo, frecuencia y cardinalidad acordes a su costo.
- El servidor valida tipo, formato, longitud, rango y pertenencia; el frontend no es una frontera de seguridad.
- Las cargas no se ejecutan ni se sirven con tipo peligroso y se procesan con límites de memoria/filas.
- Errores al cliente son estables y no contienen detalles internos; logs correlacionados conservan el diagnóstico permitido.

### Datos y secretos

- Secretos de producción no tienen valor por defecto y su ausencia impide arrancar.
- Cada servicio usa la credencial mínima; desarrollo nunca comparte secretos o datos con producción.
- Respaldos cifrados, restaurables y externos cumplen RPO/RTO y retención.
- Ningún log, métrica, error, artefacto o respaldo expone credenciales o tokens en claro.

### Operación

- Sólo artefactos que pasan la puerta de seguridad pueden desplegarse.
- Todo cambio de producción es identificable, reversible en código/configuración y acompañado por migración compatible.
- Alertas tienen dueño y runbook; una alerta sin respuesta definida no cuenta como control.

## Pruebas y verificaciones requeridas

### Aplicación

- Matriz automática para cada endpoint: anónimo, autenticado sin membresía, miembro sin permiso, rol autorizado, negocio ajeno, recurso inexistente y cuenta revocada.
- Pruebas específicas para IDOR/BOLA en todos los identificadores de negocio, sucursal, empleado, producto, proveedor, rol e invitación.
- Pruebas CSRF positivas y negativas, atributos de cookie, expiración, revocación, rotación y logout.
- Pruebas de rate limit con IP, usuario y ráfaga; validación de `429` y recuperación de ventana.
- Pruebas de límites de JSON, multipart, XLSX, cabeceras, conexiones lentas y respuestas externas grandes.
- SAST, SCA Go/npm, escaneo de secretos y DAST autenticado/no autenticado en staging.

### Infraestructura y contenedores

- Escaneo externo de puertos IPv4/IPv6: sólo 80/443, y 22 únicamente desde origen autorizado.
- Verificación TLS, redirección, HSTS y cabeceras con herramientas independientes.
- Escaneo de imágenes sin vulnerabilidades críticas/altas explotables sin excepción documentada.
- Comprobación de usuario, capacidades, mounts, filesystem, límites, healthchecks y ausencia del socket Docker.
- Reinicios de host y contenedores no activan configuración de desarrollo ni pierden datos.

### Resiliencia y recuperación

- Prueba de carga con objetivo acordado y margen; observar CPU sostenida, burst capacity, memoria, DB y latencia.
- Simular caída/reinicio de API y DB, disco casi lleno, proveedor externo lento y certificado próximo a vencer.
- Restaurar en limpio desde respaldo externo y medir RPO/RTO reales.
- Ejecutar rollback de una versión sin corromper el esquema.

### Seguridad operacional

- Verificar MFA, ausencia de llaves root, permisos AWS mínimos y contactos de recuperación.
- Disparar cada alerta crítica y confirmar recepción y runbook.
- Revisar que logs y errores no contengan secretos, tokens, contraseñas ni datos personales innecesarios.
- Simular robo de sesión/secreto y completar revocación y rotación.

## Puerta de salida a producción

El despliegue público se autoriza únicamente cuando:

- Todos los hallazgos P0 están cerrados con prueba automática o evidencia operativa.
- Los P1 están cerrados; cualquier excepción requiere riesgo, compensación, responsable y fecha de vencimiento aprobados explícitamente.
- La matriz de autorización y aislamiento multi-tenant pasa completa.
- El escaneo de dependencias, secretos e imágenes no tiene críticos ni altos explotables sin excepción aprobada.
- HTTPS, renovación, cabeceras y superficie de puertos se validaron desde Internet.
- Existe respaldo externo reciente y una restauración completa cumple RPO ≤ 1 hora y RTO ≤ 4 horas.
- Monitoreo, alertas, rotación de logs, presupuesto y runbooks fueron probados.
- Existe rollback ensayado y la versión desplegada es trazable a un artefacto inmutable.
- Un segundo revisor confirma configuración, permisos y checklist; quien despliega no es el único que valida.

## Checklist reutilizable para cualquier aplicación web

1. **Entender el riesgo:** datos, usuarios, dinero, operaciones críticas, amenazas, RPO/RTO y responsables.
2. **Reducir exposición:** inventario de rutas/puertos, negar por defecto, HTTPS, administración privada y mínimo privilegio.
3. **Asegurar identidad:** MFA administrativa, contraseñas robustas, recuperación segura, sesiones revocables y anti-enumeración.
4. **Asegurar autorización:** controles en servidor, permiso por acción, ámbito tenant y pruebas IDOR/BOLA.
5. **Controlar entradas y recursos:** validación, límites, uploads seguros, timeouts, rate limits y protección de dependencias externas.
6. **Proteger navegador y API:** CSRF, CORS exacto, CSP, HSTS, cookies seguras y errores no reveladores.
7. **Proteger datos y secretos:** cifrado, separación de ambientes, rotación, credenciales mínimas, retención y borrado.
8. **Asegurar suministro:** versiones fijadas, revisión, SAST/SCA, secret scan, imágenes, SBOM y parches continuos.
9. **Endurecer runtime:** SO actualizado, firewall, SSH por llave, procesos no root, límites, aislamiento y sin puertos auxiliares.
10. **Prepararse para fallar:** backups externos, restauración probada, migraciones seguras, rollback y capacidad suficiente.
11. **Detectar y responder:** logs útiles y redactados, métricas, alertas con dueño, runbooks e incident response.
12. **Operar continuamente:** escaneos y restauraciones periódicas, revisión de accesos, caducidad de excepciones y reevaluación tras cada cambio importante.

## Fuera del alcance aprobado actualmente

- Crear o configurar la cuenta, instancia, DNS, certificado, firewall o recursos AWS.
- Implementar las fases 3 a 7, modificar Nginx para producción, desplegar o publicar servicios.
- Comprar dominio, servicios, WAF, CDN, base administrada o herramientas de seguridad.
- Ejecutar pentest contra terceros o producción.
- Declarar cumplimiento legal o certificación formal.
- Garantizar resistencia a DDoS volumétrico o alta disponibilidad con una sola instancia.

## Verificaciones ejecutadas durante esta propuesta

| Verificación | Resultado 2026-09-19 |
| --- | --- |
| `go test ./...` | Correcta en todos los paquetes del backend. |
| `go vet ./...` | Correcta. |
| Inventario estático de rutas y uso de middleware | Detectó rutas administrativas públicas y ausencia de conexión de `RequierePermiso`. |
| Revisión de archivos `.env` sin mostrar valores | No están rastreados; modo `0644`; existen contraseñas locales predecibles que no deben reutilizarse. |
| `npm audit --omit=dev --json` | No concluyente: el entorno usa Node `v14.21.3`/npm `6.14.18` y npm falló al interpretar el lock de Angular 22. Debe repetirse con Node 24/npm 11. |

## Pendientes de decisión antes de aprobar

- Confirmar que el objetivo inicial es una sola instancia Lightsail y aceptar explícitamente su punto único de falla.
- Confirmar RPO de 1 hora y RTO de 4 horas, o elegir objetivos distintos con su costo asociado.
- Elegir dominio, región, responsables de operación y canal de alertas.
- Decidir si PostgreSQL inicia en la instancia con backup externo o si se usa una base administrada desde el comienzo.
- Confirmar la aprobación de implementación de cada fase de [PLAN_APLICACION.md](PLAN_APLICACION.md). Se recomienda aprobarlas por separado para no mezclar aplicación, datos e infraestructura en una sola autorización.

## Decisiones y seguimiento

- 2026-09-19: se creó la propuesta por solicitud del usuario; permanece `En revisión` y no autoriza implementación.
- 2026-09-19: se recomienda priorizar correcciones de autorización antes de crear una instancia pública.
- 2026-09-19: se adopta defensa en profundidad: prevenir, limitar impacto, detectar y recuperar.
- 2026-09-19: el usuario confirmó Angular y API bajo el mismo origen, catálogos por negocio y documentación del plan de aplicación; la implementación continúa pendiente de aprobación.
- 2026-09-20: el usuario aprobó explícitamente implementar las fases 1 y 2; no autorizó las fases posteriores ni el despliegue.
- 2026-09-20: se implementaron y verificaron mediante pruebas automatizadas las fases 1 y 2. La base legacy del puerto `5433` permanece bloqueada por datos ambiguos y debe reconstruirse o migrarse explícitamente.

## Registro de implementación de las fases 1 y 2

### Fase 1 — autorización y aislamiento

- [x] Se retiraron las rutas globales de catálogo y se anidaron bajo `/api/v1/negocios/:negocioId`.
- [x] Las lecturas exigen sesión, negocio activo y `catalogo.ver`; las mutaciones exigen además `catalogo.gestionar`.
- [x] Marcas, categorías, unidades, productos, categorías de producto y códigos de producto tienen propietario `negocio_id`.
- [x] Repositorios, importaciones y validaciones filtran por negocio y rechazan referencias de otro negocio.
- [x] PostgreSQL respalda el aislamiento con `NOT NULL`, unicidad por negocio y claves foráneas compuestas.
- [x] La migración se niega a adivinar si encuentra productos o catálogos compartidos entre negocios.
- [x] Se identificaron dos PostgreSQL locales. La base de Compose del puerto `5434` tenía un negocio y catálogos vacíos; la base usada por `backend/.env` en `5433` tenía tres negocios y catálogos globales sin propietario.
- [x] La migración fue probada desde cero y ejecutada dos veces para comprobar idempotencia.
- [ ] Resolver la base legacy de `5433`: reconstruirla o asignar explícitamente sus catálogos antes de ejecutar la migración sobre ella.

### Fase 2 — sesión y CSRF

- [x] Se eliminó la emisión y validación JWT y también la dependencia Go correspondiente.
- [x] Se creó `sesiones_usuario` con token y CSRF almacenados sólo como SHA-256, expiración, actividad, revocación, IP y agente limitado.
- [x] Se configuraron 24 horas, 30 días con `recordarme`, 7 días de inactividad y actualización de actividad cada 5 minutos.
- [x] Producción usa cookie `__Host-tienda_session` con `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/` y sin `Domain`.
- [x] Logout revoca la fila antes de borrar cookies; dos sesiones del mismo usuario permanecen independientes.
- [x] Angular usa `XSRF-TOKEN`/`X-XSRF-TOKEN`; el backend valida token ligado a sesión y origen exacto en métodos con efecto.
- [x] Las duraciones inválidas impiden arrancar y una tarea periódica depura sesiones inactivas después de la retención configurada.
- [x] Las cuentas activas anteriores a OTP reciben una marca de verificación una sola vez. Hasta aprobar la fase 3, el registro conserva su contrato anterior y crea la cuenta verificada; la fase 3 cambiará el alta a pendiente y activación por OTP.

### Verificaciones ejecutadas el 2026-09-20

| Verificación | Resultado |
| --- | --- |
| `go test ./...` con caché temporal | Correcta en todos los paquetes. |
| `go vet ./...` con caché temporal | Correcta en todos los paquetes. |
| Integración `TestSecurityPhaseCatalogTenancyMigrationAndIsolation` contra `tienda_security_test` | Correcta; migración idempotente, nombres iguales entre negocios, actualización cruzada rechazada y revocación inmediata comprobada. |
| `npm test -- --watch=false` con Node 24 en contenedor | 32 archivos y 108 pruebas correctas. |
| `docker compose build frontend` | Correcta; conserva advertencias previas de presupuesto de bundle/CSS, sin error de compilación. |
| `docker compose build backend` | Correcta; el primer intento tuvo un fallo transitorio de DNS y el reintento construyó el binario e imagen. |
| `git diff --check` | Correcta. |

### Pendientes conocidos

- La base local legacy del puerto `5433` no ha sido migrada: la protección detuvo el proceso para no asignar catálogos al negocio equivocado.
- La verificación real por OTP y Amazon SES pertenece a las fases 3 y 4 y no fue iniciada.
- Cambio/recuperación de contraseña todavía no existe como flujo; `SessionService.RevokeAll` quedó disponible para conectarlo cuando se apruebe esa fase.
- Las advertencias de presupuesto del frontend no bloquean estas fases, pero deben atenderse antes de la puerta final de producción.
- No se desplegó ni se modificó infraestructura externa.

## Referencias de diseño consultadas

- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [OWASP HTTP Headers Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html)
- [OWASP Docker Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [AWS Lightsail: firewall y mapeo de puertos](https://docs.aws.amazon.com/lightsail/latest/userguide/understanding-firewall-and-port-mappings-in-amazon-lightsail.html)
- [AWS Lightsail: snapshots automáticos](https://docs.aws.amazon.com/lightsail/latest/userguide/amazon-lightsail-configuring-automatic-snapshots.html)
- [AWS Lightsail: métricas de salud](https://docs.aws.amazon.com/lightsail/latest/userguide/understanding-instance-health-metrics-in-amazon-lightsail.html)
- [AWS: prácticas para el usuario root](https://docs.aws.amazon.com/IAM/latest/UserGuide/root-user-best-practices.html)
