# Capability 10: Fechas de auditoría y presentación en Ciudad de México

## Estado

En implementación.

## Aprobación

Aprobada explícitamente por el usuario el 2026-09-23 mediante la instrucción: “Implement the plan”.

La presentación frontend se aprobó explícitamente el mismo día mediante la misma instrucción. Esta capability consolida la anterior capability 11, eliminada como documento independiente.

## Alcance

- Asignar desde Go, siempre en UTC, los campos de auditoría `creado_en` y `actualizado_en` ya existentes.
- Corregir los modelos de cuenta y catálogo que no estaban configurados para que GORM los administre.
- Sustituir los valores runtime de PostgreSQL (`now()`/`NOW()`) por instantes generados en Go.
- No modificar datos existentes, columnas ni agregar `actualizado_en` a entidades inmutables.
- Mostrar en Angular los instantes recibidos de la API con la zona fija `America/Mexico_City`.
- Reemplazar los usos visibles del pipe `date` sobre timestamps de auditoría, invitaciones y asignaciones por un pipe central.
- Mantener sin cambios la API, datos persistidos y campos que representan una fecha civil sin hora.

## Reglas y contratos

- PostgreSQL mantiene `timestamptz`; Go es la fuente de los valores insertados y actualizados.
- GORM usa una única función de tiempo UTC para sus tags `autoCreateTime` y `autoUpdateTime`.
- `creado_en` se asigna al crear y no se modifica después; `actualizado_en` se renueva en cada actualización de una entidad que lo tenga.
- Las tablas inmutables conservan solo su fecha de creación o evento actual: no se cambia su contrato.
- Los valores de la API continúan en UTC; la zona de Ciudad de México solo modifica su representación visual.
- El frontend no usa la zona configurada en el navegador para estos instantes, ni formatea como timestamp los valores que son solo fecha.
- `CST`, `CDT` y `CST6CDT` no se usan: la zona fija es el identificador IANA `America/Mexico_City`.

## Tareas y decisiones

- [x] Auditar modelos, migraciones y repositorios de fechas de auditoría.
- [x] Confirmar que Go es la fuente de tiempo y que no se repararán datos históricos.
- [x] Confirmar que las entidades inmutables no recibirán una nueva columna de actualización.
- [x] Configurar el reloj UTC de GORM y los tags de los modelos afectados.
- [x] Reemplazar las escrituras runtime de fecha generadas por PostgreSQL.
- [x] Probar el contrato de auditoría y ejecutar la suite backend.
- [x] Identificar las vistas que formatean timestamps con `date`.
- [x] Crear el pipe de formato de fecha para Ciudad de México y aplicarlo a las siete vistas afectadas.
- [x] Cubrir el pipe y ejecutar la verificación de TypeScript.
- [ ] Ejecutar la suite Angular con una versión compatible de Node.
- [x] Actualizar el grafo y registrar resultados.

## Criterios de aceptación

- Todo modelo persistido con `CreadoEn` o `ActualizadoEn` usa el tag automático correspondiente.
- Las nuevas escrituras de auditoría proceden de Go en UTC, incluidos los upserts y actualizaciones SQL runtime.
- Los modelos inmutables mantienen su esquema actual.
- Un mismo timestamp UTC se muestra en hora de Ciudad de México independientemente de la zona del navegador.
- Las siete vistas que muestran timestamps usan el pipe central y los campos de fecha civil no cambian.
- Las pruebas backend y TypeScript relevantes pasan, y la suite Angular queda pendiente hasta disponer de Node compatible.
- No se altera ningún registro histórico.

## Resultados

- `go test ./...`: correcto.
- Prueba estructural de modelos: todos los campos de auditoría persistidos usan el tag automático de GORM y las tablas inmutables conservan solo su marca de creación/evento.
- Verificación estática: no quedan campos de auditoría sin tag en `usuarios`, `perfiles_usuario` ni los modelos de catálogo; no quedan escrituras runtime de fecha que invoquen `NOW()`/`now()` en PostgreSQL.
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.spec.json`: correcto.
- La prueba del pipe cubre la conversión de `2026-09-23T18:30:00Z` a `2026-09-23 12:30` en Ciudad de México y valores ausentes.
- `frontend/node_modules/.bin/ng test --watch=false --browsers=ChromeHeadless`: pendiente; no se puede ejecutar con Node `v14.15.0`, anterior al mínimo requerido por Angular CLI (`v22.22.3`).
- `git diff --check`: correcto.
- `graphify update .`: correcto; el último recorrido reconstruyó el grafo con 2008 nodos y 5659 relaciones. Los cambios y renombrados de comunidades informados son advertencias de etiquetado, no de integridad.
