import { Component, computed, ElementRef, HostListener, input, model, signal, viewChild } from '@angular/core';
import { RouterLink } from '@angular/router';

import { MascotaFiguraComponent } from './mascota-figura.component';

/** Un mensaje que la mascota puede contar. `accion` y `ruta` van juntos y son opcionales. */
export interface ConsejoMascota {
  titulo: string;
  texto: string;
  accion?: string;
  ruta?: string;
}

let siguienteId = 0;

/**
 * Mascota con globo de diálogo. Al tocarla abre un panel blanco que recorre `consejos` en ciclo.
 * Recibe lo que dice (consejos, nombre, subtítulo, invitación), así que se reusa en cualquier sección.
 * `flotante` la fija abajo a la derecha; sin él se acomoda donde se coloque.
 */
@Component({
  selector: 'app-mascota',
  imports: [RouterLink, MascotaFiguraComponent],
  templateUrl: './mascota.component.html',
  styleUrl: './mascota.component.css',
  host: {
    class: 'mascota',
    '[class.mascota--flotante]': 'flotante()',
  },
})
export class MascotaComponent {
  readonly consejos = input.required<readonly ConsejoMascota[]>();
  readonly nombre = input('Stocki');
  readonly subtitulo = input('Te cuento trucos de la app');
  /** Globo pequeño que aparece junto a la mascota mientras el panel está cerrado. */
  readonly invitacion = input<string | null>('¿Te doy un tip?');
  readonly flotante = input(true);
  readonly abierto = model(false);

  readonly indice = signal(0);
  readonly consejo = computed(() => this.consejos()[this.indice() % Math.max(this.consejos().length, 1)] ?? null);
  readonly idTitulo = `mascota-titulo-${++siguienteId}`;

  private readonly boton = viewChild<ElementRef<HTMLButtonElement>>('boton');

  alternar(): void {
    this.abierto.update((a) => !a);
  }

  cerrar(): void {
    this.abierto.set(false);
    this.boton()?.nativeElement.focus();
  }

  mover(paso: number): void {
    const total = this.consejos().length;
    if (total) this.indice.update((i) => (i + paso + total) % total);
  }

  @HostListener('document:keydown.escape')
  alPresionarEscape(): void {
    if (this.abierto()) this.cerrar();
  }
}
