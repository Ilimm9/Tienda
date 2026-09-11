import { TestBed } from '@angular/core/testing';
import { toast } from 'ngx-sonner';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

import { FeedbackService } from './feedback.service';

describe('FeedbackService', () => {
  const fire = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(Swal, 'mixin').mockReturnValue({ fire } as never);
    TestBed.configureTestingModule({});
  });

  it('delegates semantic notifications to Sonner', () => {
    const success = vi.spyOn(toast, 'success').mockReturnValue('success');
    const error = vi.spyOn(toast, 'error').mockReturnValue('error');
    const info = vi.spyOn(toast, 'info').mockReturnValue('info');
    const warning = vi.spyOn(toast, 'warning').mockReturnValue('warning');
    const service = TestBed.inject(FeedbackService);

    expect(Swal.mixin).toHaveBeenCalledWith(
      expect.objectContaining({
        buttonsStyling: false,
        allowOutsideClick: false,
        returnFocus: true,
      }),
    );

    service.success('Guardado', 'Los cambios están listos.');
    service.error('No fue posible guardar', 'Intenta nuevamente.');
    service.info('Información');
    service.warning('Atención');

    expect(success).toHaveBeenCalledWith('Guardado', {
      description: 'Los cambios están listos.',
    });
    expect(error).toHaveBeenCalledWith('No fue posible guardar', {
      description: 'Intenta nuevamente.',
      important: true,
    });
    expect(info).toHaveBeenCalledWith('Información', { description: undefined });
    expect(warning).toHaveBeenCalledWith('Atención', { description: undefined });
  });

  it.each([
    [{ isConfirmed: true }, true],
    [{ isConfirmed: false, dismiss: 'cancel' }, false],
    [{ isConfirmed: false, dismiss: 'esc' }, false],
  ])('maps the dialog result to a boolean', async (dialogResult, expected) => {
    fire.mockResolvedValue(dialogResult);
    const service = TestBed.inject(FeedbackService);

    await expect(
      service.confirmDanger({
        titulo: 'Archivar negocio',
        descripcion: 'Los datos se conservarán.',
        textoConfirmar: 'Sí, archivar',
      }),
    ).resolves.toBe(expected);

    expect(fire).toHaveBeenCalledWith(
      expect.objectContaining({
        titleText: 'Archivar negocio',
        text: 'Los datos se conservarán.',
        confirmButtonText: 'Sí, archivar',
        cancelButtonText: 'Cancelar',
        focusCancel: true,
      }),
    );
  });
});
