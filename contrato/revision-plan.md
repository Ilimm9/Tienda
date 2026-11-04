# Revisión del plan de cierre de Stockion

- **Fecha:** 2026-11-02.
- **Estado:** **cerrada el 2026-11-02.** Los pendientes se trasladaron a [preguntas](preguntas.md) y al [plan de desarrollo](plan-desarrolladores.md); ver [Cierre](#cierre). No aprueba capabilities ni acredita una entrega funcional.
- **Documentos revisados completos:** `plan-maestro.md`, `plan-desarrolladores.md`, `plan-lic.md`, `preguntas.md`, `servicios-externos.md` e `ideas-pentest.md`.
- **Contexto adicional:** `AGENTS.md`, `PLAN_IMPLEMENTACION.md`, bases de Compras y planes actuales, consulta de graphify y comprobaciones puntuales de código y documentación oficial.
- **Límite de la revisión:** no se ejecutaron suites, pruebas de seguridad, restauraciones ni pruebas en tienda. Los resultados de pruebas y vulnerabilidades del maestro siguen siendo evidencia histórica del 2026-10-30; no un diagnóstico técnico actualizado.

## Dictamen

El plan identifica bien los riesgos operativos y gran parte de lo prometido, pero todavía no demuestra que el alcance completo pueda entregarse dentro de cuatro meses con dos personas a 20 h semanales. Hace falta convertir bloques grandes en trabajo restante medible, resolver dependencias y asegurar disponibilidad de la clienta.

Las dos restricciones solicitadas quedan registradas como confirmadas: **20 h por persona por semana** y **máximo cuatro meses**. El calendario solapado y los presupuestos de capacidad son propuestas para revisión; no se presenta un calendario comprimido como prueba de viabilidad.

## Reasignación antes del arranque

El usuario aclaró que este plan aún no ha iniciado y solicitó trabajo paralelo por sección en cada fase. Confirmó almacenamiento y autenticación para thrs y enfoque de Ilian en ventas. El [plan de desarrollo](plan-desarrolladores.md) concentra ahora responsabilidad por módulos, archivos centrales y contratos; el detalle restante sigue como propuesta para revisar entre ambos.

| Fase | thrs | Ilian |
| --- | --- | --- |
| 1 | Infraestructura, almacenamiento, respaldos, autenticación, migraciones, catálogo/dinero, Compras/Precios, buzón y aprobación móvil | Ajuste de ventas, inventario, migración de Odoo y adelanto del núcleo sin conexión y prueba de hardware |
| 2 | Terminales y credenciales, reportes y Compras sin conexión | Núcleo local y servidor de sincronización, venta local, caja, cobros, devoluciones, clientes/envases, hardware y conteos sin conexión |
| 3 | Seguridad, aislamiento, carga del servidor, restauración y cierre administrativo/documental | E2E operativos, pruebas sin conexión, hardware y cierre de mostrador/inventario |

Cambios principales respecto del reparto anterior:

- E1-03b pasa de Ilian a thrs, incluidos endurecimiento de autenticación y coordinación de permisos centrales.
- E1-09/10 pasan a thrs para que todas las modificaciones de Compras tengan un único responsable.
- E1-12/14 pasan a thrs por pertenecer a Inicio administrativo, buzón y aprobación móvil. Ilian conserva la base PWA, su worker y manifest dentro de E2-01; desaparece la dependencia circular entre base PWA y aprobación móvil.
- E2-01/03a pasan a Ilian: mismo responsable para persistencia local, copia de catálogo, cola, envío y recepción de sincronización.
- E2-10 se divide en Compras sin conexión de thrs y conteos sin conexión de Ilian; el alcance se mantiene y no se duplica la estimación.
- E2-02 permanece con thrs como sección de identidad de terminal, credencial y PIN; Ilian consume su contrato sin modificar autenticación.
- F3-05 pasa a thrs para carga del servidor. Ilian conserva rendimiento local y hardware dentro de F3-04. F3-08 se divide por sección; los cierres tienen coordinador documental y cada corrección tiene responsable.
- Capturista se integra por thrs en permisos centrales; Ilian define y verifica necesidades del vendedor. No se agrega un rol nuevo por fuera del alcance ya descrito.

Contratos tempranos: dinero/impuestos, movimientos de inventario, buzón/auditoría, identidad de terminal, cola/transporte y consultas para reportes. Se distingue construir contra ejemplos acordados de completar la integración real. La entrega de contrato/modelo de ventas permite versionar migraciones antes de la aceptación fiscal integral y evita un ciclo entre E1-05, E1-03 y E1-06.

Los archivos centrales de arranque, modelos heredados de catálogo, permisos y navegación tienen integrador único: thrs. Ilian entrega sus registros y requisitos desde su módulo. Las migraciones nuevas y pruebas propias se mantienen por responsable, con nombres reservados. Revisión cruzada no equivale a editar módulos ajenos.

**Riesgo pendiente:** la carga cambia por momento: thrs concentra Fase 1 e Ilian Fase 2. Se permite adelantar módulos propios bajo sus capabilities aprobadas, pero todavía falta estimar horas por persona y semana. Se retiró la distribución anterior por paquetes y se conserva capacidad de 320 h por persona, con reserva propuesta de 64 h cada uno. Ninguna tarea de este plan se considera comenzada por la reasignación.

## Capacidad y plazo

| Concepto | Cálculo | Resultado |
| --- | --- | --- |
| Capacidad semanal del equipo | 2 personas × 20 h | 40 h |
| Horizonte conservador de planificación | 16 semanas | Menor que cuatro meses calendario |
| Capacidad por persona | 16 × 20 h | 320 h |
| Capacidad total | 16 × 40 h | 640 h |
| Reserva propuesta | 20 % de 640 h | 128 h |
| Trabajo planificable propuesto | 640 − 128 h | 512 h |
| Calendario anterior de desarrolladores | 33–42 semanas × 40 h | 1,320–1,680 h de capacidad prevista |
| Reducción frente a ese calendario | 1 − 640 / capacidad anterior | Aproximadamente 52–62 % |

La comparación anterior mide capacidad de calendario, no una estimación de esfuerzo ya validada. Los tamaños S/M/L no tienen horas restantes ni una velocidad observada. El maestro suponía dedicación principal y otro alcance; por eso sus 28–34 semanas tampoco son una base intercambiable sin más.

Las 640 h incluyen coordinación, investigación, revisión, pruebas manuales, atención al piloto, documentación y correcciones. No equivalen a 640 h de programación. No se descuentan automáticamente horas por usar agentes. Las ausencias reducen capacidad y deben registrarse al fijar fechas; no se añaden horas semanales por defecto.

## Ruta propuesta para respetar el máximo

El detalle está en [plan de desarrollo](plan-desarrolladores.md). La propuesta conserva las tres fases, solapa construcción y pruebas y mantiene las compuertas para retirar Odoo.

| Ventana | Resultado que debe estar probado |
| --- | --- |
| S1–2 | Respaldos y restauración antes de datos reales; hardware identificado; revisión legal inicial; contratos iniciales y estimación por bloque |
| S3–4 | Ventas ajustadas, después migraciones; dinero e impuestos definidos; prueba mínima de venta local y reintento en staging |
| S5–8 | Administración e inventario integrados, datos iniciales preparados y construcción de sincronización en paralelo |
| S9–12 | Cuatro semanas de compras reales; completar caja, cobros, envases, reportes y 13 escenarios sin conexión |
| S13–14 | Transición gradual y una semana completa en ambas sucursales sin Odoo, con conciliación y aceptación |
| S15–16 | Correcciones finales, recuperación, documentación y aceptación final |

La seguridad y las pruebas se realizan durante todas las ventanas. Si administración no está lista al terminar S8, el piloto de cuatro semanas ya amenaza el cierre; si el POS no supera sus pruebas al terminar S12, la transición amenaza S15–16. No se recortan esas evidencias para aparentar cumplimiento.

## Lo que no cuadra

1. **Había tres relatos de plazo y alcance.** La propuesta original citada era de cinco meses; el maestro proponía 28–34 semanas y una Etapa 3 de SaaS; desarrolladores y clienta usaban 33–42 semanas y revisión final. Se alinearon a máximo cuatro meses y revisión final, conservando E3 como inventario futuro. El alcance completo todavía requiere validación.

2. **La aceleración no tenía fundamento medido.** El propio plan advertía unas 35 semanas-persona al sumar tamaños máximos de Fase 1. Ese cálculo no es una estimación formal, pero contradice tratar 16 semanas de todo el proyecto como un compromiso ya sustentado. Los nuevos presupuestos son topes para detectar excesos, no horas estimadas artificialmente para que la suma cuadre.

3. **El trabajo paralelo tenía dependencias largas.** El reparto anterior concentraba migraciones, dinero, buzón, núcleo local, terminales, servidor de sincronización y reportes en thrs. La reasignación mueve el motor completo a Ilian y distingue contratos de aceptación integral. Permanecen dependencias reales de dinero, identidad y movimientos; deben probarse juntas, no darse por resueltas con ejemplos.

4. **Los maestros discrepaban en migraciones y ventas.** El maestro pedía E1-03 antes de E1-05, y desarrolladores invertía el orden para incluir ventas en la línea base. Se alineó la planeación: entrega de contrato/modelo estabilizado de ventas, línea base y después validación fiscal integral con E1-06. Esperar la aceptación final de ventas antes de migraciones formaría un ciclo con impuestos; el plan ahora distingue esos hitos. Integrar en producción sigue exigiendo respaldo, aprobación y prueba de esquema; la prioridad no autoriza ejecutar migraciones destructivas existentes.

5. **Los pilotos consumen calendario real.** Cuatro semanas de compras requieren que administración, catálogo y conteo estén listos al inicio de ese periodo. No basta marcar el piloto como «continuo» desde S1. La semana de ambas sucursales sin Odoo empieza después del despliegue gradual y sus pruebas.

6. **Legal y seguridad llegaban tarde.** F3-09 estaba al cierre, pero preguntas exigía revisión legal antes de datos reales. Se adelantó preparación legal y se hizo transversal la seguridad. F3 conserva revisión final, no el primer control.

7. **Hardware y durabilidad se descubrían demasiado tarde.** D-3 y D-8 condicionan impresión, cajón, báscula y PWA frente a escritorio. Se propone prueba con modelos reales desde S1–2. Sin esa evidencia, el plazo de E2-08 y la plataforma siguen abiertos.

8. **«Sin internet» prometía más de lo especificado.** Una PWA no evita pérdida por destrucción de equipo, almacenamiento lleno o eliminación de datos locales. Tampoco una referencia de tarjeta garantiza cobro bancario. Se aclararon límites y sigue pendiente ventana máxima, vencimiento y contingencia (D-4).

9. **Aceptar una venta ya cobrada no resuelve confianza.** «Nunca rechazar» convivía con credenciales revocadas y datos locales manipulables. Falta separar preservación de evidencia, escritura autorizada, cuarentena y recuperación administrativa (D-22). El servidor no debe convertir importes cobrados en otros para cuadrar el catálogo ni permitir acceso entre negocios.

10. **El cursor de sincronización no estaba definido.** Citar `pg_snapshot_xmin` no basta para evitar pérdida de cambios: devuelve un identificador de transacción, no el ID secuencial de una bitácora. Esta es una inferencia de diseño a partir de la [documentación de PostgreSQL](https://www.postgresql.org/docs/current/functions-info.html#FUNCTIONS-PG-SNAPSHOT). Hace falta algoritmo y prueba de confirmaciones en orden inverso, paginación y recuperación.

11. **R2 y S3 no se sustituyen íntegramente por configuración.** Las operaciones básicas pueden abstraerse, pero compatibilidad y retención difieren. Se corrigió la promesa: [API S3 de R2](https://developers.cloudflare.com/r2/api/s3/api/) y [bucket locks](https://developers.cloudflare.com/r2/buckets/bucket-locks/). Tampoco estaba definido cómo recuperar la llave de cifrado ni las fotos, además de la base.

12. **Hay promesas sin asignación completa.** Cambiar contraseña con sesión, cerrar sesiones y MFA aparecen en maestro §5.4, sin tarea equivalente completa en el plan de desarrolladores. FEFO, alertas por caducidad y envases con proveedor no tienen decisión de alcance cerrada. El inicio de vendedor aparece en el maestro pero no en E2-10 de desarrolladores. Cada punto necesita inclusión con responsable o exclusión explícita, sin asumir aprobación.

13. **Existen cambios de regla de negocio que necesitan capability.** Producto nuevo pendiente y rol Capturista afectan creación desde compras y permisos existentes. Las bases actuales documentan otro flujo; un resumen del contrato no autoriza modificarlo. Se comprobaron puntualmente `AutoMigrate`, `SKUInterno` sin `column` y existencias `float64`; no se comprobó toda la implementación de ventas, ausente como dominio `application/venta` en este checkout.

14. **Migrar datos exige conciliación, no solo importación.** Los 215 negativos, 128 productos sin código y 12 sin costo son cifras históricas del diagnóstico, pendientes de confirmar contra exportación vigente. Falta definir responsable del conteo, tratamiento de ventas entre conteo e importación, fecha de corte, entrada de cada sucursal y vuelta atrás si la transición falla.

15. **Márgenes y utilidad necesitan definiciones.** Costo, redondeo, impuestos, unidad base y presentación deben cuadrar entre compras, venta e inventario. Un reporte de margen bruto no equivale a utilidad neta cuando gastos y comisiones no están incluidos. Evitar una promesa comercial de utilidad real que dependa de módulos futuros.

16. **La revisión propia de seguridad no tiene esfuerzo dimensionado.** Proponer herramientas más ASVS nivel 2 completo puede superar el presupuesto de cierre. Falta versión, controles aplicables, responsables y evidencia. No equivale a pentest independiente ni se acredita con ejecutar un escáner.

17. **La carpeta aún no contiene contrato listo para firmar.** Precio, pagos, IVA, infraestructura posterior, propiedad intelectual, datos, garantía, soporte y cambios siguen abiertos. Esto describe trabajo contractual pendiente; no determina términos legales ni reemplaza revisión profesional.

18. **Los índices no contaban la misma historia de IDs.** `PLAN_IMPLEMENTACION.md` señalaba 35, mientras el maestro registra 35–37 usados en otras ramas y siguiente 38. En este checkout solo hay capabilities hasta 34. Se corrigió el índice para exigir confirmación y reserva con Ilian; 38 sigue como candidato, no como una disponibilidad global comprobada. Ausencia local no significa disponibilidad.

19. **Foto como evidencia y lectura automática estaban mezcladas.** Fase 1 incluye adjuntar foto del ticket. El extra es extraer datos automáticamente para prellenar compras. Se aclaró la versión para la Lic. para evitar cobrar o prometer dos veces la misma cosa.

20. **No había medidas completas de aceptación.** «Carga real», «sin errores graves» y «metas cumplidas» no definen concurrencia, volumen, percentil de latencia ni severidad. Los 13 escenarios sin conexión sí aportan una buena base, pero también faltan criterios medibles para el resto y agenda de revisión con la Lic.

## Decisiones que siguen abiertas

Detalle y responsables en [preguntas](preguntas.md). Resolver primero:

- **D-19 y D-20:** inicio, fecha límite y horas restantes por bloque; cómo preservar fecha y dedicación si la carga no cabe. No se ha aprobado quitar funciones comprometidas ni contratar apoyo.
- **D-1 y reparto:** revisar calendario y carga por persona con Ilian; preferencias del usuario ya incorporadas y secciones detalladas en el plan. No equivale a aprobación de implementación.
- **D-3, D-4, D-8 y D-9:** plataforma de caja, modelos reales, número de terminales, ventana sin conexión y comunicación entre cajas.
- **D-10 y D-11:** necesidad inmediata de CFDI, lotes, caducidad y FEFO. Si CFDI es condición de operación, dejarlo como extra cambia viabilidad y debe resolverse antes de prometer salida de Odoo.
- **Operación:** productos y unidades a granel, factores de caja, mayoreo, crédito, envases por cliente/proveedor, precios, redondeo y validación de impuestos con contador.
- **D-21 y D-22:** pérdida máxima aceptable, recuperación, retención, llaves, contingencia local y conciliación segura de operaciones pendientes.
- **Datos y pruebas:** responsable de conteos, disponibilidad para pilotos, criterios de aceptación y severidad, promesas sin tarea asignada.
- **Comerciales y legales:** precio, pagos, IVA, infraestructura posterior, quién firma, propiedad intelectual, datos, garantía, soporte, cambios, abogado y marca (D-12/D-13).

Planes de uso, recargas, cobros en línea, terminal integrada, WhatsApp y lectura automática pueden decidirse después: siguen fuera de estas tres fases. No deben bloquear la entrega actual. La necesidad de CFDI sí debe resolverse ahora con la Lic.

## Recomendación

Mantener máximo de cuatro meses y 20 h por persona como restricciones. Durante S1–2, estimar lo que falta, empezar por contratos críticos y probar equipo y venta local. Evitar trabajo de SaaS y extras. Entregar y validar semanalmente, con una sola tarea de construcción activa por persona y reserva visible.

Si el alcance completo supera la capacidad, hay tres caminos que deben explicitarse: simplificar detalles sin incumplir funciones pactadas, acordar recortes concretos con la Lic., o añadir apoyo con responsabilidades y costo definidos. Ninguno está aprobado por este análisis. Una distribución de horas que suma 640 no demuestra por sí sola que el trabajo cabe.

## Verificaciones de esta revisión

- Lectura completa de los seis documentos originales y contraste de decisiones, fases y dependencias.
- Comprobación aritmética de capacidad, reserva y presupuestos por persona.
- Reasignación comprobada: cada tarea tiene un responsable o coordinador explícito; E2-10 y F3-08 conservan alcance mediante desglose por sección. Autenticación y Compras quedan con thrs; núcleo local y sincronización cliente/servidor, con Ilian. Se conservaron los códigos originales mediante sus desgloses y se verificaron enlaces y espacios tras los cambios.
- Consulta de graphify y lecturas puntuales; resultados técnicos históricos preservados como tales.
- Revisión de enlaces locales de los siete documentos resultantes y del índice, y revisión de espacios y conflictos documentales.
- Sin cambios de código, aprobación de capabilities, operaciones de escritura Git, contratación ni despliegue. La viabilidad funcional, el esfuerzo restante y la aceptación siguen pendientes.

## Cierre

Esta revisión no se actualiza más. El seguimiento vive en [preguntas](preguntas.md) (decisiones) y en [plan de desarrollo](plan-desarrolladores.md) (tareas). Las respuestas de thrs del 2026-11-02 quedan pendientes de aprobación de Ilian.

| # | Punto | Estado al cierre | Dónde sigue |
| --- | --- | --- | --- |
| 1 | Relatos de plazo y alcance | **Resuelto.** Cuatro meses, inicio propuesto el 10 de noviembre de 2026 y límite el 10 de marzo de 2027 | Preguntas D-18/D-19 |
| 2 | Aceleración sin medir | Abierto. Capacidad propuesta de 20–30 h por persona (640–960 h); sin aceleración supuesta por agentes | D-7, D-20 |
| 3 | Dependencias del paralelo | Mitigado con contratos por sección; integración real pendiente | Tabla de contratos del plan |
| 4 | Orden ventas y migraciones | **Resuelto.** Hitos A y B de E1-05 | E1-05 desglosado |
| 5 | Pilotos consumen calendario | **Resuelto** en el orden sugerido (piloto S9–12) | Orden sugerido |
| 6 | Legal y seguridad tardíos | **Resuelto.** Legal antes de datos reales; seguridad transversal | F3-09, F3-01 |
| 7 | Hardware tardío | Mitigado con prueba en S1–2; faltan modelos reales | E2-08.1; Parte 2 de preguntas |
| 8 | Promesa «sin internet» | Límites aclarados a la Lic., incluida tarjeta. Falta ventana máxima | D-4 |
| 9 | Confianza en ventas ya cobradas | Abierto | D-22 |
| 10 | Cursor de sincronización | Abierto; prueba exigida y algoritmo por acordar con la bitácora | E2-01.4 |
| 11 | R2 frente a S3 | **Resuelto** en E1-16; llaves y recuperación de fotos pendientes | D-21 |
| 12 | Promesas sin asignación | FEFO en E1-11.9; envases con proveedor en E2-07 si la Lic. lo confirma; inicio de vendedor en E2-03b; MFA en F4-03. Cambiar contraseña y cerrar sesiones siguen sin tarea | Fila «sesiones, MFA…» de preguntas |
| 13 | Reglas que requieren capability | Normal: cada cambio entra por capability aprobada | Capabilities de cada tarea |
| 14 | Migración con conciliación | Desglosado en E1-13; fecha de corte y vuelta atrás pendientes | E1-13; fila de movimientos en preguntas |
| 15 | Margen frente a utilidad | Abierto; `plan-lic.md` no promete utilidad neta | Nueva fila en preguntas |
| 16 | Esfuerzo de revisión de seguridad | Abierto | Nueva fila en preguntas |
| 17 | Contrato listo para firmar | Parcial: precio, pagos, infraestructura y soporte con respuesta propuesta. IVA, garantía, cambios, propiedad intelectual, firma y abogado abiertos | Preguntas, sección comercial y legal |
| 18 | IDs de capability | Abierto; 38 sigue como candidato | Nueva fila en preguntas |
| 19 | Foto frente a lectura automática | **Resuelto** en `plan-lic.md` | — |
| 20 | Medidas de aceptación | Abierto | Fila de medidas objetivas en preguntas |

**Riesgo nuevo detectado al cerrar:** D-10 se decidió como «no se ofrece CFDI por ahora». Si hoy la Lic. factura desde Odoo, apagar Odoo le quitaría esa función. Hay que confirmar cómo factura hoy antes de E2-11; quedó como fila en preguntas.

