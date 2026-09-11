import { Injectable } from '@angular/core';
import { toast } from 'ngx-sonner';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

export interface DangerConfirmationOptions {
  titulo: string;
  descripcion: string;
  textoConfirmar: string;
  textoCancelar?: string;
}

@Injectable({ providedIn: 'root' })
export class FeedbackService {
  private readonly dangerDialog = Swal.mixin({
    buttonsStyling: false,
    reverseButtons: true,
    allowOutsideClick: false,
    allowEscapeKey: true,
    returnFocus: true,
    heightAuto: false,
    customClass: {
      container: 'tienda-alerta-container',
      popup: 'tienda-alerta',
      icon: 'tienda-alerta__icono',
      title: 'tienda-alerta__titulo',
      htmlContainer: 'tienda-alerta__descripcion',
      actions: 'tienda-alerta__acciones',
      confirmButton: 'tienda-alerta__confirmar',
      cancelButton: 'tienda-alerta__cancelar',
    },
  });

  success(titulo: string, descripcion?: string): void {
    toast.success(titulo, { description: descripcion });
  }

  error(titulo: string, descripcion?: string): void {
    toast.error(titulo, { description: descripcion, important: true });
  }

  info(titulo: string, descripcion?: string): void {
    toast.info(titulo, { description: descripcion });
  }

  warning(titulo: string, descripcion?: string): void {
    toast.warning(titulo, { description: descripcion });
  }

  async confirmDanger(options: DangerConfirmationOptions): Promise<boolean> {
    const result = await this.dangerDialog.fire({
      icon: 'warning',
      titleText: options.titulo,
      text: options.descripcion,
      showCancelButton: true,
      focusCancel: true,
      confirmButtonText: options.textoConfirmar,
      cancelButtonText: options.textoCancelar ?? 'Cancelar',
    });

    return result.isConfirmed;
  }
}
