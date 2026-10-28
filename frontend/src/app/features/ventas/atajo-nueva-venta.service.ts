import { inject, Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';

/** F2 abre una venta nueva desde cualquier pantalla, sin perder trabajo sin guardar. */
@Injectable({ providedIn: 'root' })
export class AtajoNuevaVentaService {
  private readonly router = inject(Router);
  private readonly contexto = inject(ContextoService);
  private readonly cambios = inject(CambiosPendientesService);
  private readonly feedback = inject(FeedbackService);

  alPresionarF2(evento: KeyboardEvent): void {
    if (evento.repeat || !this.contexto.puede(PERMISOS.ventaRegistrar)) return;
    evento.preventDefault();
    // En la captura, la propia pantalla enfoca el campo.
    if (this.router.url.split(/[?#]/, 1)[0] === '/ventas/nueva') return;
    if (this.cambios.hayPendientes() || formularioConDatos(document.activeElement)) {
      this.feedback.warning(
        'Tienes cambios sin guardar',
        'Guarda o descarta los cambios antes de abrir una venta.',
      );
      return;
    }
    void this.router.navigate(['/ventas/nueva']);
  }
}

/**
 * El foco está en un formulario con datos capturados: algún campo tiene texto o una opción marcada.
 * Es conservador a propósito: ante la duda, no se navega y no se pierde lo capturado.
 */
export function formularioConDatos(elemento: Element | null): boolean {
  const formulario = elemento?.closest('form');
  if (!formulario) return false;
  return Array.from(formulario.elements).some((campo) => {
    if (campo instanceof HTMLInputElement) {
      if (campo.type === 'checkbox' || campo.type === 'radio')
        return campo.checked !== campo.defaultChecked;
      if (['button', 'submit', 'reset', 'hidden'].includes(campo.type)) return false;
      return campo.value.trim() !== campo.defaultValue.trim();
    }
    if (campo instanceof HTMLTextAreaElement)
      return campo.value.trim() !== campo.defaultValue.trim();
    return false;
  });
}
