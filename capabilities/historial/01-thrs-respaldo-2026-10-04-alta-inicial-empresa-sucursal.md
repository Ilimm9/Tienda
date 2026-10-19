# Plan 1: Alta inicial de empresa y sucursal

## Estado y control

- Estado: `Borrador`; pendiente revisión y aprobación explícita.
- Fecha: 2026-10-01. Alcance actual: documentación, sin implementación.
- Ubicación directa en `capabilities/thrs`, solicitada por el usuario; sin carpetas adicionales.
- Este borrador no sustituye la capability aprobada requerida antes de implementar.
- Dependencias: identidad y contexto de capability 01; seguridad de 06; acceso visual de 13.

## Objetivo

Después de autenticarse, orientar al propietario hasta disponer de mínimo una empresa y una sucursal activa. Permitir navegar sin completar alta; cada sección explica los datos que necesita antes de operar. Empresa corresponde a `negocio` en código y DB; no crear una entidad paralela.

## Hallazgos de revisión

- Revisión transversal de los 36 archivos Markdown del repositorio, incluidos índice, fases, seguridad, correo, despliegue y diseño visual.
- Referencia principal: [base.MD](base-modelo-datos.md), especialmente identidad, negocios, sucursales y membresías. Es una propuesta de esquema, no evidencia de tablas físicas en producción.
- Modelos y migraciones actuales ya contemplan las tablas necesarias; no se consultó ni modificó una DB.
- `contextoGuard` acepta `sin_sucursal`; varias rutas del panel no tienen ese guard.
- Login y verificación redirigen a `/inicio`; hoy no existe asistente de alta inicial.
- Documentación antigua menciona JWT y catálogos globales; código y capability 06 ya usan sesiones opacas y aislamiento por negocio.
- Las asignaciones existentes todavía no restringen sucursales del empleado; ese cambio pertenece al [plan 2](02-thrs-respaldo-2026-10-04-invitaciones-empleados-sucursal.md).

## Enfoque recomendado

Aviso visible en Inicio y asistente voluntario con estado calculado desde datos vigentes. Evita depender de un indicador `primer_login` que podría quedar desactualizado.

- Alternativa descartada por el usuario: asistente obligatorio que impide navegar hasta completar alta.
- Alternativa: tabla de progreso del asistente; añade sincronización y migraciones sin necesidad inicial.
- Decisión explícita del usuario: solo aviso, navegación libre sin completar alta. Mínimo operativo propuesto: 1 empresa y 1 sucursal, sin limitar altas posteriores.

## Flujo y reglas

1. Validar cuenta, correo verificado y sesión mediante contratos actuales.
2. Dar prioridad al retorno de una invitación válida; nunca pedir crear empresa al empleado durante su aceptación.
3. Consultar estado de alta y opciones autorizadas desde backend.
4. Sin membresías previas: mostrar aviso en `/inicio` y CTA «Configurar empresa» hacia `/configuracion-inicial`, paso «Tu empresa».
5. Guardar empresa con configuración, membresía propietaria y rol propietario mediante flujo existente.
6. Continuar con «Tu primera sucursal»: nombre y código; primera activa queda principal automáticamente.
7. Recargar contexto, seleccionar empresa/sucursal recién creadas y abrir `/inicio`.
8. Con empresa propia incompleta: CTA retoma paso sucursal; no duplicar empresa ni forzar redirección.
9. Con contexto válido existente: entrar directamente; ocultar aviso y no repetir asistente.
10. Con varias empresas: permitir elegir una operable; una empresa incompleta no bloquea las demás.
11. Miembro sin sucursales disponibles: mostrar «Solicita una asignación al administrador»; no darle permisos de alta.
12. Membresía suspendida/revocada o negocio archivado: informar acceso no disponible; no tratarlo como cuenta nueva.

Campos mínimos: empresa `nombre_comercial`; sucursal `nombre` y `codigo`. Moneda `MXN` y zona `America/Mexico_City` como valores iniciales. RFC, razón social, contactos y direcciones siguen opcionales; no exigir logo ni datos fiscales para empezar.

## Navegación, operaciones y recuperación

- Permitir rutas del panel y acceso directo por URL según permisos; falta de alta no oculta ni inhabilita menú general.
- Inicio muestra aviso persistente mientras falta estructura, con CTA y opción de seguir explorando. Asistente permite salir al panel.
- Cada sección sin contexto muestra estado vacío orientado a configurar empresa/sucursal; no dispara consultas con IDs nulos ni inventados.
- Deshabilitar solo acciones que necesitan contexto faltante, con explicación visible; secciones que no lo necesitan siguen funcionando.
- En API conservar validaciones de negocio, permisos y sucursal necesarias para cada operación; no agregar bloqueo global por alta incompleta.
- Negocio y sucursal de cada operación se validan en servidor; nunca usar IDs de ejemplo ni confiar en `localStorage`.
- Si se archiva la última sucursal, se recupera estado incompleto; propietario puede crear/restaurar y miembro solicita ayuda.
- Permitir recuperación administrativa autorizada; conservar datos y reglas actuales de archivado.
- Error de red muestra reintento y conserva formulario; no redirigir en bucle ni asumir que lista vacía significa fallo.
- Guardar por paso usando transacciones existentes. Tras respuesta perdida, reconciliar altas existentes antes de repetir un POST; no reenviar automáticamente.
- Bloquear doble clic mientras se guarda. Avisar antes de perder cambios no guardados y reanudar registros ya persistidos.

## Tablas y contratos

| Tabla existente | Uso |
| --- | --- |
| `usuarios`, `perfiles_usuario`, `sesiones_usuario` | Identidad, perfil y sesión; no guardar empresa fija en usuario. |
| `negocios` | Empresa activa, creador, moneda y zona horaria. |
| `configuraciones_negocio` | Configuración 1:1 creada con valores predeterminados. |
| `membresias_negocio` | Acceso vigente y vínculo propietario con empresa. |
| `sucursales`, `direcciones` | Primera ubicación activa y dirección opcional. |
| `roles`, `permisos`, `permisos_rol`, `roles_membresia` | Permisos efectivos, incluido rol propietario. |

**Tablas nuevas recomendadas: ninguna.** Estado se obtiene de membresías, negocios y sucursales vigentes. No crear `empresas`, `onboarding` ni tablas comerciales futuras de `base.MD` para este flujo.

- Mantener `GET /api/v1/contexto/opciones` y contratos actuales de alta.
- Proponer `GET /api/v1/contexto/alta-inicial`, autenticado: `estado`, `negocio_id`, `sucursal_id`, `puede_completar`, `ruta_sugerida`; esta ruta orienta, no fuerza redirección.
- Estados: `requiere_empresa`, `requiere_sucursal`, `listo`, `sin_asignacion`, `acceso_no_disponible`; distinguir fallos HTTP de estos estados funcionales.
- Altas: `POST /api/v1/negocios` y `POST /api/v1/negocios/:negocioId/administracion/sucursales`.
- Error uniforme existente: `codigo`, `mensaje`, `campos`; `401` sin sesión y `403` sin permiso. Una operación sin contexto requerido devuelve error controlado, nunca fallo interno; mantener contratos actuales donde ya existe validación.
- Garantizar unicidad de membresía y código de sucursal; conservar índice parcial de principal activa.

## Diseño visual propuesto

Referencia prioritaria: login implementado; imagen de marca aportada queda como referencia secundaria y provisional.

| Uso | Color de referencia |
| --- | --- |
| Acciones y paso activo | Índigo `#2C2E6A`; hover `#20224E`. |
| Fondo general | Crema `#FAF8F3`; formularios blancos. |
| Texto principal | Azul profundo `#172554`. |
| Pasos completados y acentos | Verde `#3F5147`, salvia `#DFE5C8` y `#CFD9BD`. |
| Detalles decorativos | Durazno `#F4D3C1`, arena `#F5E6C5` y amarillo suave `#F8EAC1`. |

- Pantalla dedicada: título claro, progreso «Empresa / Sucursal / Listo», formulario y resumen lateral.
- Dar vida con bloques suaves de color, iconos de empresa/sucursal y jerarquía clara; evitar llenar toda la interfaz de colores.
- Inter, tarjetas con radios y sombras discretas, estados vacíos orientados a completar alta.
- Reutilizar tokens semánticos; sustituir paleta clara global y revisar shell, Inicio y pantallas de empresa/sucursal. Otras secciones heredan tokens; rediseñar sus layouts queda fuera.
- Mantener login como referencia, sin rehacer mosaico ni animación. Marca aportada no se instala como logo definitivo.
- Solo diseñar y aceptar modalidad clara ahora; modo oscuro queda pendiente, sin ampliar este plan a su rediseño.
- Contraste mínimo 4.5:1 para texto normal; color acompañado de texto/icono, foco visible y movimiento reducido.
- En móvil, pasos y formulario en una columna; permitir scroll natural y teclado virtual.

## Alcance técnico futuro y compatibilidad

- Backend: dominio `negocio` en sus cuatro capas; composición en `cmd/api/main.go` para consulta de estado. Sin middleware de bloqueo global.
- Frontend: contexto, guards, rutas, shell, Inicio, empresa/sucursales y tokens de `styles.css`.
- Autenticación solo para resolver destino después de login/OTP; conservar contratos de sesión y recuperación.
- Estas áreas quedan expresamente incluidas para implementación futura; no mover ni reorganizar otros dominios.
- Mantener navegación administrativa libre de fase 3; revisar guards y secciones que hoy redirigen por falta de contexto para permitir exploración con estado vacío.
- Paleta nueva es alcance adicional frente a capability 05, aún en revisión; no depende de instalar Tailwind.

## Tareas y verificación futura

- [x] Confirmar aviso sin bloqueo de navegación; usuario eligió esta opción explícitamente.
- [ ] Confirmar mínimos y propuesta visual.
- [ ] Aprobar formalmente alcance antes de implementar.
- [ ] Implementar resolución de estado, aviso, estados vacíos y asistente voluntario con componentes existentes.
- [ ] Aplicar paleta clara y revisar superficies que heredan tokens.
- [ ] Probar cuenta nueva, alta interrumpida, recarga, URL directa, fallo de red y doble clic.
- [ ] Probar varias empresas, última sucursal archivada, permisos insuficientes y empleado invitado.
- [ ] Probar navegación libre sin alta, acciones sin contexto y API directa con referencias a otro negocio.
- [ ] Ejecutar `go test ./...`, pruebas Angular y build con Node compatible; registrar fallos previos por separado.
- [ ] Revisar móvil/escritorio, teclado, contraste y actualizar Graphify si se modifica código y existe grafo.

## Criterios de aceptación

- Cuenta nueva puede navegar entre secciones sin completar alta y recibe aviso claro para registrar empresa/sucursal.
- Alta existente se retoma; contexto completo entra directamente y empleado invitado no crea empresa propia.
- Acciones dependientes de empresa/sucursal no se ejecutan sin contexto válido; API valida acceso, recuperación autorizada y logout siguen disponibles.
- No aparecen IDs ficticios ni duplicados por doble clic; alta fallida no deja relaciones incompletas dentro de cada transacción.
- Pantallas incluidas muestran identidad del login, legibilidad y diseño responsive en claro.

## Evidencia y decisiones pendientes

- Revisión documental y contraste con modelos, migraciones, rutas y estilos realizados el 2026-10-01.
- Solo se redactó plan; pruebas funcionales y visuales no ejecutadas porque no hay implementación.
- Decisión confirmada: «Solo aviso: permitir navegar sin completar alta». Mínimo operativo propuesto: 1 + 1.
- Aprobar borrador no equivale a aceptación final ni autoriza Git, publicación o despliegue.
