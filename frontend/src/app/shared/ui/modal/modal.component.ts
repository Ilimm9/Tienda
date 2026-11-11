import { A11yModule } from '@angular/cdk/a11y';
import {
  afterNextRender,
  Component,
  ElementRef,
  HostListener,
  inject,
  input,
  OnDestroy,
  output,
  viewChild,
} from '@angular/core';

import { FeedbackService } from '../../feedback/feedback.service';

export type ModalTamano = 'chico' | 'mediano' | 'grande' | 'amplio';

const ANCHOS: Record<ModalTamano, number> = { chico: 580, mediano: 680, grande: 940, amplio: 1180 };
let siguienteId = 0;

/**
 * Ventana modal de la app: título con subtítulo, cuerpo con scroll y pie con acciones.
 * En pantallas chicas se vuelve hoja inferior; con `completoEnMovil` ocupa toda la pantalla.
 * Cierra con Esc, con la X o al tocar el fondo; si `sucio` está activo, pregunta antes de
 * descartar lo escrito.
 *
 * El pie se proyecta con el atributo `modal-pie`. Los estilos viven en `styles/componentes.css`
 * (bloque «Ventana modal») porque el contenido proyectado también los usa.
 */
@Component({
  selector: 'app-modal',
  imports: [A11yModule],
  template: `
    <div class="modal" [class.modal--completo]="completoEnMovil()" role="presentation">
      <div class="modal__fondo" aria-hidden="true" (click)="solicitarCierre()"></div>
      <section
        #panel
        class="modal__panel"
        role="dialog"
        aria-modal="true"
        [attr.aria-labelledby]="idTitulo"
        [attr.aria-describedby]="subtitulo() ? idSubtitulo : null"
        [style.--modal-ancho.px]="ancho()"
        [attr.aria-busy]="ocupado() || null"
        cdkTrapFocus
      >
        <header class="modal__encabezado">
          <div class="modal__titulos">
            <h2 [id]="idTitulo">{{ titulo() }}</h2>
            @if (subtitulo()) {
              <p [id]="idSubtitulo">{{ subtitulo() }}</p>
            }
          </div>
          <button type="button" class="icono-accion modal__cerrar" aria-label="Cerrar" [disabled]="ocupado()" (click)="solicitarCierre()">
            <i class="pi pi-times" aria-hidden="true"></i>
          </button>
        </header>
        <div class="modal__cuerpo">
          <ng-content />
        </div>
        <footer class="modal__pie">
          <ng-content select="[modal-pie]" />
        </footer>
      </section>
    </div>
  `,
})
export class ModalComponent implements OnDestroy {
  readonly titulo = input.required<string>();
  readonly subtitulo = input<string | null>(null);
  readonly tamano = input<ModalTamano>('chico');
  /** Ventanas grandes (productos): en tablet casi toda la pantalla y en móvil, pantalla completa. */
  readonly completoEnMovil = input(false);
  /** Hay datos sin guardar: se confirma antes de cerrar. */
  readonly sucio = input(false);
  /** Mientras guarda no se puede cerrar. */
  readonly ocupado = input(false);
  readonly cerrar = output<void>();

  readonly idTitulo = `modal-titulo-${++siguienteId}`;
  readonly idSubtitulo = `modal-subtitulo-${siguienteId}`;

  private readonly feedback = inject(FeedbackService);
  private readonly panel = viewChild.required<ElementRef<HTMLElement>>('panel');
  private readonly focoPrevio = document.activeElement as HTMLElement | null;
  private confirmando = false;

  constructor() {
    afterNextRender(() => this.enfocarInicio());
  }

  ancho(): number {
    return ANCHOS[this.tamano()];
  }

  @HostListener('document:keydown.escape', ['$event'])
  alPresionarEscape(event: Event): void {
    // Un diálogo de SweetAlert abierto encima atiende su propio Esc.
    if (this.confirmando || document.body.classList.contains('swal2-shown')) return;
    event.preventDefault();
    void this.solicitarCierre();
  }

  async solicitarCierre(): Promise<void> {
    if (this.ocupado() || this.confirmando) return;
    if (this.sucio()) {
      this.confirmando = true;
      const descartar = await this.feedback.confirmDanger({
        titulo: '¿Descartar cambios?',
        descripcion: 'Lo que escribiste en esta ventana no se guardará.',
        textoConfirmar: 'Descartar',
        textoCancelar: 'Seguir editando',
      });
      this.confirmando = false;
      if (!descartar) return;
    }
    this.cerrar.emit();
  }

  /** Vuelve a poner el cursor en el primer campo (útil después de «Guardar y agregar otro»). */
  enfocarInicio(): void {
    const panel = this.panel().nativeElement;
    const campo = panel.querySelector<HTMLElement>(
      '.modal__cuerpo [autofocus], .modal__cuerpo input:not([type=hidden]):not(:disabled), .modal__cuerpo select:not(:disabled), .modal__cuerpo textarea:not(:disabled)',
    );
    (campo ?? panel.querySelector<HTMLElement>('.modal__cerrar'))?.focus();
  }

  ngOnDestroy(): void {
    this.focoPrevio?.focus?.();
  }
}
