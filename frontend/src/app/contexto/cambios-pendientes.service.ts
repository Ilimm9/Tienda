import { Injectable } from '@angular/core';

/** Un componente declara aquí si tiene trabajo sin guardar. */
export type DeclaracionCambiosPendientes = () => boolean;

/**
 * Contrato explícito de cambios sin guardar.
 *
 * El cambio de contexto no se bloquea por patrones de URL: cada componente con un formulario o
 * una importación en curso se registra y se da de baja al destruirse.
 */
@Injectable({ providedIn: 'root' })
export class CambiosPendientesService {
  private readonly declaraciones = new Set<DeclaracionCambiosPendientes>();

  registrar(declaracion: DeclaracionCambiosPendientes): () => void {
    this.declaraciones.add(declaracion);
    return () => this.declaraciones.delete(declaracion);
  }

  hayPendientes(): boolean {
    for (const declaracion of this.declaraciones) {
      if (declaracion()) return true;
    }
    return false;
  }
}
