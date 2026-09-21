# Capability 04: Tailwind y sistema visual

## Estado

En revisión.

## Aprobación

Pendiente de aprobación explícita del usuario.

## Objetivo

Integrar Tailwind CSS 4 en el frontend Angular como capa de utilidades, conservar la identidad visual existente y comprobar el enfoque mediante una mejora incremental del shell y la pantalla de inicio.

## Alcance

- Instalar Tailwind CSS mediante la integración oficial para Angular y npm.
- Configurar PostCSS y cargar Tailwind desde la hoja global del frontend.
- Mantener las variables CSS actuales como fuente de verdad para colores y sombras.
- Exponer tokens semánticos de la aplicación a utilities de Tailwind, incluidos texto, fondos, superficies, bordes, colores de estado y sombras.
- Hacer que la variante `dark:` responda al selector existente `[data-theme="dark"]`.
- Conservar `ThemeService`, la preferencia guardada en `tienda.theme` y el seguimiento del tema del sistema cuando no existe preferencia manual.
- Mantener compatible el selector oscuro configurado para PrimeNG.
- Aplicar Tailwind de forma piloto al shell compartido, sus elementos de navegación y la pantalla de inicio para reducir CSS repetido y mejorar consistencia responsive, estados de foco y espaciado.
- Mantener CSS convencional donde Tailwind no sea buena opción: overrides de PrimeNG, SweetAlert2, Sonner, tablas, carga de archivos y reglas complejas de terceros.
- Eliminar solo reglas CSS que queden demostrablemente sin uso dentro de los archivos incluidos en este alcance.

## Fuera de alcance

- Rediseñar la marca o sustituir la paleta vigente.
- Cambiar contratos HTTP, lógica de negocio, rutas, permisos o comportamiento funcional.
- Migrar en masa las vistas de autenticación, catálogo, productos, negocios, sucursales, equipo o roles y permisos.
- Sustituir PrimeNG, PrimeIcons, SweetAlert2 o Sonner.
- Forzar que todo estilo se exprese con utilities o usar `@apply` como reemplazo general del CSS actual.
- Modificar archivos de dominio que formen parte de trabajo activo no coordinado.

## Reglas y contratos

### Integración

- Se usará Tailwind CSS 4 y su plugin oficial de PostCSS, compatibles con el Angular actual del proyecto.
- Dependencias y lockfile permanecerán dentro de `frontend/`.
- La configuración preferirá el modelo CSS-first de Tailwind 4; no se creará `tailwind.config.js` sin necesidad demostrada.
- Se revisará el efecto de Preflight antes de conservarlo. Si altera controles o componentes existentes, se importarán solo las capas necesarias o se documentará una corrección acotada.

### Paleta y temas

- Los valores de `:root` y `:root[data-theme='dark']` seguirán definiendo la paleta efectiva.
- Tailwind expondrá nombres semánticos, no colores de marca duplicados ni valores hexadecimales repetidos en templates.
- Una utility como `bg-app-background`, `bg-app-surface`, `text-app`, `text-app-muted`, `border-app` o `text-app-primary` resolverá al token CSS correspondiente.
- El cambio de tema actualizará inmediatamente CSS existente, utilities Tailwind y componentes PrimeNG mediante el mismo atributo `data-theme`.
- La variante `dark:` usará `[data-theme='dark']`; no se añadirá un segundo estado `.dark`.
- Estados `primary`, `danger`, `success` y `warning` conservarán contraste legible en ambos temas.

### Convivencia y migración

- La adopción será incremental. Templates migrados pueden combinar clases de componente y utilities mientras exista una razón clara.
- Estilos encapsulados o de terceros que requieran selectores complejos continuarán en CSS.
- No se construirán nombres de clase Tailwind mediante concatenación dinámica que impida su detección durante el build.
- Estados condicionales usarán clases completas y detectables por Tailwind.
- Responsive y foco visible no pueden degradarse frente al estado actual.

## Plan de trabajo

- [ ] Instalar dependencias y registrar configuración PostCSS.
- [ ] Importar Tailwind y validar orden de capas con estilos globales actuales.
- [ ] Mapear tokens semánticos mediante `@theme inline`.
- [ ] Registrar variante oscura basada en `data-theme`.
- [ ] Migrar y refinar shell, topbar, sidebar y breadcrumbs.
- [ ] Migrar y refinar pantalla de inicio y placeholder compartido.
- [ ] Revisar apariencia en claro, oscuro, escritorio y móvil.
- [ ] Ejecutar pruebas y build.
- [ ] Registrar resultados, pendientes y decisiones finales.

## Pruebas

- Pruebas unitarias existentes de `ThemeService`, layout y navegación.
- Prueba de que tema manual claro/oscuro sigue persistiendo en `localStorage`.
- Prueba de que tema del sistema sigue aplicándose sin preferencia manual.
- Build de producción para confirmar generación y purga de utilities.
- Suite completa del frontend.
- Revisión visual en anchos móvil y escritorio, con ambos temas.
- Revisión visual de PrimeNG, diálogos, tablas, cargas y mensajes para detectar regresiones de Preflight u orden CSS.
- Comprobación de foco por teclado y contraste en navegación, botones y tarjetas tocadas.

## Criterios de aceptación

- Tailwind compila dentro del flujo Angular sin pasos manuales adicionales.
- Build y pruebas del frontend terminan correctamente con Node 24 definido por `.nvmrc`.
- Paleta clara y oscura conserva los tokens actuales como fuente de verdad.
- Toggle de tema controla estilos propios, utilities Tailwind y PrimeNG sin estados divergentes.
- Shell e inicio muestran composición consistente en móvil y escritorio, sin cambiar comportamiento funcional.
- No aparecen regresiones visibles en componentes no migrados.
- CSS eliminado corresponde solo a reglas reemplazadas y sin uso dentro del alcance.
- Resultados y pendientes quedan registrados en esta capability.

## Riesgos y mitigaciones

- Preflight puede alterar controles y bibliotecas: validar componentes existentes y limitar capas importadas si hace falta.
- Convivencia temporal puede producir especificidad inesperada: definir orden de capas y evitar `!important` salvo compatibilidad documentada.
- Utilities con valores directos pueden fragmentar la paleta: usar tokens semánticos para color y sombra.
- Una migración total aumentaría riesgo y conflictos: limitar esta capability a base visual, shell e inicio.
- Tailwind CSS 4 requiere navegadores modernos; se validará que el soporte objetivo sea compatible antes de cerrar.

## Decisiones y seguimiento

- 2026-09-18: se propone Tailwind CSS 4 por ser la integración documentada actualmente para Angular.
- 2026-09-18: se conserva `data-theme` porque ya coordina CSS propio, `ThemeService` y PrimeNG.
- 2026-09-18: se propone adopción incremental, no reescritura completa de casi 5,000 líneas CSS.
- 2026-09-18: implementación detenida hasta aprobación explícita.

