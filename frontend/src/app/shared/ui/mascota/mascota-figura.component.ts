import { Component, input } from '@angular/core';

/**
 * Stocki, la mascota de Stockion: un robot armado con cajas de cartón y detalles morados.
 * Solo dibuja la figura; sirve sola para estados vacíos o avisos, o dentro de `app-mascota`.
 * Los colores son fijos (cartón y morado de marca) para que se reconozca en ambos temas.
 */
@Component({
  selector: 'app-mascota-figura',
  host: {
    '[style.width.px]': 'tamano()',
    '[style.height.px]': 'tamano() * 1.25',
  },
  styles: `
    :host {
      display: inline-block;
      flex-shrink: 0;
    }
    svg {
      display: block;
      width: 100%;
      height: 100%;
      overflow: visible;
      fill: none;
      stroke-width: 2.5;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
    .saludo {
      transform-box: fill-box;
      transform-origin: 0% 100%;
    }
    :host(.mascota-figura--saluda) .saludo {
      animation: saludar 1.6s ease-in-out 0.4s 2;
    }
    @keyframes saludar {
      0%, 100% { transform: rotate(0deg); }
      25% { transform: rotate(-14deg); }
      50% { transform: rotate(8deg); }
      75% { transform: rotate(-10deg); }
    }
    @media (prefers-reduced-motion: reduce) {
      :host(.mascota-figura--saluda) .saludo { animation: none; }
    }
  `,
  template: `
    <svg viewBox="0 0 120 150" role="img" [attr.aria-label]="etiqueta()" [attr.aria-hidden]="etiqueta() ? null : 'true'" focusable="false">
      <ellipse cx="60" cy="144" rx="30" ry="4" fill="#E9E3D4" />
      <path d="M60 22V11" stroke="#2C2E6A" />
      <circle cx="60" cy="8" r="5.5" fill="#4F46C8" stroke="#2C2E6A" />
      <rect x="20" y="22" width="80" height="56" rx="7" fill="#D9B27C" stroke="#2C2E6A" />
      <path d="M20 33h80" stroke="#B98A4E" stroke-width="2" />
      <rect x="51" y="22" width="18" height="11" fill="#7C74E0" />
      <rect x="31" y="39" width="58" height="30" rx="9" fill="#4F46C8" stroke="#2C2E6A" />
      <rect x="42" y="47" width="8" height="10" rx="4" fill="#fff" />
      <rect x="70" y="47" width="8" height="10" rx="4" fill="#fff" />
      <path d="M54 61q6 4 12 0" stroke="#fff" stroke-width="2.2" />
      <circle cx="25" cy="64" r="3" fill="#F0B5A8" />
      <circle cx="95" cy="64" r="3" fill="#F0B5A8" />
      <rect x="50" y="78" width="20" height="7" fill="#B98A4E" stroke="#2C2E6A" stroke-width="2" />
      <path d="M30 96q-14 4-15 20" stroke="#2C2E6A" stroke-width="9" />
      <path d="M30 96q-14 4-15 20" stroke="#C99D62" stroke-width="5" />
      <circle cx="15" cy="118" r="5.5" fill="#4F46C8" stroke="#2C2E6A" stroke-width="2" />
      <g class="saludo">
        <path d="M90 96q14-4 18-20" stroke="#2C2E6A" stroke-width="9" />
        <path d="M90 96q14-4 18-20" stroke="#C99D62" stroke-width="5" />
        <circle cx="108" cy="74" r="5.5" fill="#4F46C8" stroke="#2C2E6A" stroke-width="2" />
      </g>
      <rect x="28" y="85" width="64" height="46" rx="6" fill="#D9B27C" stroke="#2C2E6A" />
      <path d="M28 93l10 6M92 93l-10 6" stroke="#B98A4E" stroke-width="2" />
      <rect x="55" y="85" width="10" height="46" fill="#7C74E0" />
      <rect x="36" y="106" width="14" height="11" rx="2" fill="#fff" stroke="#B98A4E" stroke-width="1.5" />
      <path d="M39 110h8M39 113h5" stroke="#B98A4E" stroke-width="1.3" />
      <rect x="38" y="131" width="15" height="10" rx="2" fill="#B98A4E" stroke="#2C2E6A" stroke-width="2" />
      <rect x="67" y="131" width="15" height="10" rx="2" fill="#B98A4E" stroke="#2C2E6A" stroke-width="2" />
    </svg>
  `,
})
export class MascotaFiguraComponent {
  /** Ancho en px; el alto se ajusta a la proporción de la figura. */
  readonly tamano = input(96);
  /** Texto para lectores de pantalla. Vacío: la figura es decorativa. */
  readonly etiqueta = input('');
}
