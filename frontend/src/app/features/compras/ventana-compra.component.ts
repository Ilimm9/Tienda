import { Component, ElementRef, effect, input, output, viewChild } from '@angular/core';

let siguienteId = 0;

/**
 * Ventana modal de Registrar compra sobre `<dialog>` nativo: el navegador aporta Esc,
 * el foco atrapado, el fondo inerte y el regreso del foco al cerrar.
 */
@Component({
  selector: 'app-ventana-compra',
  template: `
    <dialog #dialogo class="ventana" [attr.aria-labelledby]="tituloId" (close)="cerrar.emit()" (click)="alClic($event)">
      <section class="dialogo ventana__caja">
        <header class="dialogo__encabezado ventana__encabezado">
          <span class="ventana__icono" aria-hidden="true"><i class="pi" [class]="icono()"></i></span>
          <div class="ventana__titulos">
            <h1 [id]="tituloId">{{ titulo() }}</h1>
            @if (subtitulo()) {
              <p>{{ subtitulo() }}</p>
            }
          </div>
          <button type="button" class="ventana__cerrar" aria-label="Cerrar" (click)="dialogo.close()">
            <i class="pi pi-times" aria-hidden="true"></i>
          </button>
        </header>
        <div class="ventana__cuerpo"><ng-content /></div>
        <footer class="ventana__pie"><ng-content select="[pie]" /></footer>
      </section>
    </dialog>
  `,
  styles: `
    .ventana { width: min(520px, calc(100vw - 32px)); max-height: calc(100dvh - 64px); padding: 0; border: 0; background: transparent; overflow: visible; }
    .ventana::backdrop { background: rgb(23 37 84 / 45%); }
    .ventana__caja { display: flex; max-height: calc(100dvh - 64px); flex-direction: column; max-width: none; }
    .ventana__encabezado { align-items: center; border-bottom: 1px solid var(--border-soft); padding-bottom: 18px; }
    .ventana__icono { display: grid; width: 44px; height: 44px; flex: 0 0 44px; place-items: center; border-radius: 12px; color: var(--primary); background: var(--primary-soft); font-size: 18px; }
    .ventana__titulos { min-width: 0; flex: 1; }
    .ventana__cerrar { display: grid; width: 40px; height: 40px; flex: 0 0 40px; place-items: center; border: 0; border-radius: 10px; color: var(--text-muted); background: transparent; cursor: pointer; }
    .ventana__cerrar:hover { background: var(--surface-hover); }
    .ventana__cerrar:focus-visible { outline: none; box-shadow: var(--focus-ring); }
    .ventana__cuerpo { display: flex; flex-direction: column; gap: 16px; padding: 22px 24px; overflow-y: auto; }
    .ventana__pie { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 10px; padding: 16px 24px; border-top: 1px solid var(--border-soft); background: var(--surface-row-hover); }
  `,
})
export class VentanaCompraComponent {
  readonly abierta = input(false);
  readonly titulo = input.required<string>();
  readonly subtitulo = input('');
  readonly icono = input('pi-plus');
  readonly cerrar = output<void>();
  readonly tituloId = `ventana-compra-${siguienteId++}`;
  private readonly dialogo = viewChild.required<ElementRef<HTMLDialogElement>>('dialogo');

  constructor() {
    effect(() => {
      const dialogo = this.dialogo().nativeElement;
      if (this.abierta() && !dialogo.open) dialogo.showModal();
      else if (!this.abierta() && dialogo.open) dialogo.close();
    });
  }

  /** Un clic en el fondo cae sobre el propio `<dialog>`, fuera de su caja. */
  alClic(evento: MouseEvent): void {
    const dialogo = this.dialogo().nativeElement;
    if (evento.target === dialogo) dialogo.close();
  }
}
