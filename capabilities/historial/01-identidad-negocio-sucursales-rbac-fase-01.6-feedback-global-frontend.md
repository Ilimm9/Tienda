# Fase 1.6: Feedback global del frontend

## Estado de fase

`Cerrada`

Propuesta elaborada y autorizada explícitamente para implementación el 2026-09-10. Implementación y verificación técnica terminadas el 2026-09-10; resultado aceptado por el usuario el 2026-09-11.

## Objetivo y alcance

Unificar las notificaciones transitorias y las confirmaciones del frontend antes de construir sucursales. `ngx-sonner` mostrará notificaciones y `sweetalert2` solicitará confirmaciones. Ambas bibliotecas usarán configuración, estilos y textos centralizados.

Dependencias ya instaladas y comprobadas en `frontend/package.json`:

- `ngx-sonner ^3.1.0`.
- `sweetalert2 ^11.26.25`.

Incluye:

- Montar una sola instancia global de `NgxSonnerToaster` en la raíz de la aplicación.
- Crear un servicio compartido que encapsule `ngx-sonner` y `sweetalert2`.
- Definir estilos globales profesionales compatibles con los temas claro y oscuro existentes.
- Sustituir los `window.confirm` actuales del dominio negocio.
- Aplicar notificaciones a autenticación y a mutaciones de negocio ya implementadas.
- Establecer reglas obligatorias para las pantallas frontend de fases posteriores.
- Agregar pruebas unitarias, de integración de componentes y revisión visual responsive.

No incluye:

- Cambios de backend, base de datos o contratos HTTP.
- Cambios funcionales a negocios, autenticación o navegación.
- Edición de producto, catálogo, proveedores o inventario, porque pertenecen al trabajo del colaborador responsable.
- Confirmaciones para acciones reversibles que no tengan consecuencias relevantes.
- Reemplazar mensajes persistentes de validación o errores de carga por notificaciones efímeras.

## Organización propuesta

```text
frontend/src/app/
├── app.component.ts
├── app.component.html
└── shared/
    └── feedback/
        ├── feedback.service.ts
        └── feedback.service.spec.ts

frontend/src/styles.css
```

Reglas:

- Componentes y features consumirán `FeedbackService`; no importarán directamente `toast` ni `Swal`.
- Solo `AppComponent` montará `<ngx-sonner-toaster />`.
- Configuración visual global permanecerá en `styles.css`; no se duplicará en CSS de componentes.
- `shared/feedback` se permite porque contiene una capacidad transversal concreta y nombrada, no funciona como carpeta comodín.
- Tipos de opciones propios del servicio permanecerán junto al servicio mientras no exista otro consumidor.

## Contrato propuesto de `FeedbackService`

El servicio expondrá una API pequeña y estable:

```text
success(titulo, descripcion?)
error(titulo, descripcion?)
info(titulo, descripcion?)
warning(titulo, descripcion?)
confirmDanger(opciones) -> Promise<boolean>
```

`confirmDanger` recibirá como mínimo:

```text
titulo
descripcion
textoConfirmar
textoCancelar opcional, predeterminado "Cancelar"
```

Reglas del contrato:

- Los componentes reciben `boolean`; no conocen `SweetAlertResult`.
- Confirmaciones destructivas no se cierran haciendo clic fuera del diálogo.
- Tecla `Escape` cancela y el foco regresa al control que abrió el diálogo.
- Se usará texto escapado (`titleText` y `text`), nunca HTML construido con datos del usuario.
- Errores mostrados al usuario no expondrán cuerpos HTTP, stack traces ni detalles internos.
- Textos estarán en español, serán breves y describirán resultado o acción siguiente.

## Configuración global propuesta de notificaciones

`NgxSonnerToaster` usará:

- Posición superior derecha.
- Tema enlazado a `ThemeService`, reaccionando a `light` y `dark`.
- Colores semánticos enriquecidos.
- Botón de cierre visible.
- Duración predeterminada de 4 segundos.
- Máximo de 4 notificaciones visibles.
- Separación de 24 px en escritorio y adaptación propia de la biblioteca a 16 px en móvil y áreas seguras.
- Una sola región global para evitar notificaciones duplicadas al navegar.

El estilo de Sonner se ajustará mediante sus variables CSS oficiales `--ngx-sonner-*` y selectores `data-sonner-*`, reutilizando colores, bordes, radios, tipografía y sombras ya definidos por Tienda.

## Configuración global propuesta de confirmaciones

SweetAlert2 se configurará desde el servicio mediante un `mixin` único:

- `buttonsStyling: false` para usar botones del diseño Tienda.
- Botón de cancelar visible y enfocado inicialmente en acciones destructivas.
- Botones en orden seguro: cancelar antes de confirmar.
- Cierre por clic exterior deshabilitado en acciones destructivas.
- Cierre con `Escape` habilitado como cancelación.
- Restauración del foco habilitada.
- Clases globales con prefijo `tienda-alerta` para popup, título, texto, acciones y botones.

`styles.css` importará la hoja base de SweetAlert2 y agregará la capa visual del proyecto:

- Fondo, texto, borde y sombra desde variables de tema existentes.
- Botón destructivo rojo y botón cancelar neutro.
- Estados `hover`, `disabled` y `focus-visible` con contraste suficiente.
- Controles con altura táctil mínima de 44 px.
- Diálogo responsive; botones apilados y de ancho completo en pantallas pequeñas.
- Overlay y `z-index` coherentes con menús, diálogos y notificaciones existentes.
- Respeto a `prefers-reduced-motion`.

## Reglas de uso

### Usar notificación

- Éxito después de una mutación confirmada por API.
- Error de una acción puntual que el usuario puede volver a intentar.
- Aviso breve que no requiere una decisión inmediata.

### Mantener mensaje dentro de la pantalla

- Validación de campos antes de enviar formulario.
- Errores asociados a un campo específico.
- Error de carga inicial que impide usar toda la vista.
- Información que debe permanecer visible para poder corregirla.

### Usar confirmación

- Archivar negocio u otra acción destructiva o de impacto relevante.
- No usar SweetAlert2 como mensaje informativo ni para anunciar éxito.
- Restaurar negocio no requiere confirmación porque es reversible; mostrará notificación de éxito.
- No mostrar simultáneamente toast y mensaje inline para el mismo error.
- Quedan prohibidos `window.alert` y `window.confirm` en código nuevo.

## Flujos actuales que se migrarán

### Autenticación

- Registro correcto: navegar a login y mostrar `Cuenta creada` mediante notificación global; retirar el mensaje duplicado basado en `?registrado=1`.
- Login correcto: navegar sin notificación de éxito para evitar ruido en una acción habitual.
- Logout: navegar siempre a login como hoy, sin notificación de éxito. Si falla, no afirmar que el cierre remoto ocurrió.
- Validaciones y errores de credenciales permanecerán inline porque requieren atención en el formulario.

### Negocios

- Crear: `Negocio registrado` después de respuesta correcta.
- Editar: `Cambios guardados` después de respuesta correcta.
- Archivar: reemplazar `window.confirm` por `confirmDanger`; notificar `Negocio archivado` solo después del éxito.
- Restaurar: ejecutar sin confirmación y notificar `Negocio restaurado` después del éxito.
- Errores de carga completos y validación permanecerán inline.
- Fallos de archivar o restaurar usarán notificación de error sin duplicar un mensaje inline.

## Aplicación obligatoria en fases posteriores

- Toda mutación frontend nueva debe definir su resultado exitoso, error persistente y, si corresponde, confirmación.
- Features usarán exclusivamente `FeedbackService` para notificaciones y confirmaciones.
- Componentes no agregarán estilos globales propios para Sonner o SweetAlert2.
- Acciones reversibles no pedirán confirmación por defecto.
- Acciones destructivas deben usar lenguaje específico; evitar textos genéricos como `¿Está seguro?`.
- Producto y catálogo adoptarán estas reglas únicamente cuando su responsable coordine el cambio; fase 1.6 no modificará sus archivos.

## Estrategia de implementación

1. Crear `FeedbackService` y sus pruebas unitarias.
2. Montar y configurar el toaster global enlazado al tema actual.
3. Importar base de SweetAlert2 y agregar estilos globales de Tienda para ambas bibliotecas.
4. Migrar flujos de autenticación definidos en esta fase.
5. Migrar creación, edición, archivado y restauración de negocios.
6. Eliminar `window.confirm` únicamente de los archivos en alcance.
7. Ejecutar pruebas, build, revisión de accesibilidad y revisión visual responsive en ambos temas.
8. Registrar comandos, resultados y pendientes en esta sección.

## Pruebas de fase 1.6

- `FeedbackService` delega correctamente `success`, `error`, `info` y `warning` a Sonner.
- `confirmDanger` devuelve `true` al confirmar y `false` al cancelar o cerrar con `Escape`.
- La aplicación renderiza exactamente un toaster global.
- El toaster cambia entre tema claro y oscuro junto con `ThemeService`.
- Cancelar archivado no ejecuta petición HTTP ni cambia navegación.
- Confirmar archivado ejecuta una sola petición; éxito y error muestran feedback correcto.
- Crear, editar y restaurar negocio muestran una sola notificación después de éxito.
- Registro conserva validaciones inline y muestra éxito una sola vez; login conserva sus validaciones y navega sin toast de éxito.
- Logout conserva navegación actual y no comunica éxito remoto cuando la API falla.
- Búsqueda estática confirma ausencia de nuevos `window.alert`, `window.confirm` e imports directos productivos de las bibliotecas fuera del servicio y raíz autorizados.
- Revisión visual en 390 px, 768 px y 1440 px, temas claro y oscuro.
- Revisión con teclado cubre foco inicial, tabulación, cancelación con `Escape` y retorno del foco.
- `npm test -- --watch=false` y `npm run build` pasan sin errores nuevos.
- Diff confirma cero modificaciones a producto, catálogo, proveedores e inventario.

## Criterios de aceptación de fase 1.6

- [x] Dependencias instaladas fueron identificadas y sus versiones registradas.
- [x] Estructura, contrato y reglas de uso fueron revisados por el usuario.
- [x] Estilo global profesional fue aprobado por el usuario.
- [x] Fase recibió aprobación explícita para implementación.
- [x] Existe un único toaster global sincronizado con el tema.
- [x] Notificaciones y confirmaciones pasan por `FeedbackService`.
- [x] Flujos actuales de autenticación y negocio fueron migrados sin cambiar contratos HTTP.
- [x] No quedan `window.confirm` en archivos de negocio.
- [x] Producto y catálogo permanecen intactos.
- [x] Pruebas, build, accesibilidad y revisión responsive pasan.
- [x] Resultados de verificación fueron registrados.
- [x] Usuario acepta resultado final.

## Verificación técnica de fase 1.6

Fecha: 2026-09-10.

- `npm test -- --watch=false`: correcto, 19 archivos y 40 pruebas.
- `npm run build`: correcto.
- Build conserva tres avisos CSS preexistentes: dos presupuestos de `negocio-shared.css` y uno de `productos.component.css`.
- Import directo del build ESM de SweetAlert2 eliminó el aviso nuevo de CommonJS; declaración local conserva tipado mediante el paquete oficial.
- Búsqueda estática: ningún `window.alert` o `window.confirm` permanece en `src/app`.
- Búsqueda estática: imports productivos de Sonner y SweetAlert2 limitados a raíz y `FeedbackService`.
- Revisión visual: confirmación a 1440 px y 390 px en tema oscuro; toast real a 768 px en tema claro.
- Confirmación visual: foco inicial en `Cancelar`, botones apilados en móvil y sin desbordamiento horizontal.
- Prueba visual de registro interceptó la respuesta HTTP dentro del navegador; no escribió usuarios ni datos en backend.
- Diff revisado: ningún archivo de producto, catálogo, proveedores o inventario fue modificado.

## Imprevistos resueltos de fase 1.6

- jsdom no implementa `window.matchMedia`, requerido por `ngx-sonner`; se agregó stub limitado a la prueba de `AppComponent`.
- El import principal de SweetAlert2 generaba advertencia CommonJS en Angular. Se cambió al artefacto ESM oficial y se agregó declaración TypeScript local.
- La primera prueba visual apuntó a `/auth/registro` en vez de `/auth/register`; CORS detuvo la petición antes del backend. Se corrigió el interceptor y la verificación final se ejecutó sin persistencia.

## Decisiones resueltas de fase 1.6

- Contrato, configuración global e implementación aprobados explícitamente el 2026-09-10.
- Login y logout no mostrarán notificación de éxito; registro y mutaciones de negocio sí.

---
