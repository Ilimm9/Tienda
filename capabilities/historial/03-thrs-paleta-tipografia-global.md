# Plan 3: Paleta clara y tipografía global

Resumen al 2026-10-04. Versión completa anterior en `.respaldo-2026-10-04/`.

## Estado

- `Aprobada` e implementada el 2026-10-02. Solo estilos del frontend.
- **Sus valores de color fueron sustituidos por el [plan 4](04-thrs-adaptacion-opcion-g.md)** (opción G). Siguen vigentes la tipografía, la base crema y el tema oscuro desactivado.

## Vigente hoy

- **Inter** variable servida desde el proyecto (`@fontsource-variable/inter`), familia en `--font-sans`. Antes se declaraba sin cargarse.
- Escala: título de página 30 px / 700; sección 18 px / 600; cuerpo y tablas 14 px; etiquetas 13 a 14 px / 600; rótulos 11 a 12 px / 600.
- Cifras tabulares en tablas. `lang="es"` en `index.html`.
- Base crema: `--background` `#faf8f3`, `--text` `#172554`, `--text-muted` `#5b6478`, `--border` `#e6e0d2`, `--border-strong` `#cfc8b8`.
- Estados: `--success` `#3f5147` / `#dfe5c8`; `--warning` `#6b4a00` / `#f8eac1`; `--danger` `#bd2525` / `#fff1f1`.
- Tema oscuro desactivado: `ThemeService` aplica siempre claro; sin botón de tema.
- Preset PrimeNG derivado de Aura (la escala primaria la fija el plan 4).
- Menú lateral crema, sin «Acceso seguro».
- Colores fijos de diálogos, alertas y sombras pasados a variables.

## Sustituido por el plan 4

- `--primary` índigo `#2c2e6a` como color de acción: ahora violeta `#4f46c8`; el índigo queda para la topbar.
- Activo del menú en índigo sólido: ahora lavanda con barra violeta.
- Pastillas sueltas de filtro: ahora control segmentado.
- Tonos durazno y arena: retirados de pantallas.

## Otras ramas

- `thrs-vB`: variante B (marino, esmeralda, Sora en títulos), solo para comparar. No es la elegida.

## Pendiente

- [ ] Revisión visual de Inicio, Negocios, Sucursales, Proveedores, Compras, Precios y Ventas, y todo en móvil.
- [ ] Medir contraste: texto 4.5:1; bordes de campo y foco 3:1.
- [ ] Login: la fila «Contraseña / ¿Olvidaste tu contraseña?» queda pegada al campo.
