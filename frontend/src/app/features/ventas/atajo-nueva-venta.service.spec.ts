import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';

import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { AtajoNuevaVentaService, formularioConDatos } from './atajo-nueva-venta.service';

describe('AtajoNuevaVentaService', () => {
  let router: { url: string; navigate: ReturnType<typeof vi.fn> };
  let puede: boolean;
  let pendientes: boolean;
  const feedback = { warning: vi.fn() };

  function servicio(): AtajoNuevaVentaService {
    router = { url: '/compras', navigate: vi.fn(() => Promise.resolve(true)) };
    TestBed.configureTestingModule({
      providers: [
        { provide: Router, useValue: router },
        { provide: ContextoService, useValue: { puede: () => puede } },
        { provide: CambiosPendientesService, useValue: { hayPendientes: () => pendientes } },
        { provide: FeedbackService, useValue: feedback },
      ],
    });
    return TestBed.inject(AtajoNuevaVentaService);
  }

  const f2 = (repeat = false) =>
    new KeyboardEvent('keydown', { key: 'F2', repeat, cancelable: true });

  beforeEach(() => {
    TestBed.resetTestingModule();
    puede = true;
    pendientes = false;
    feedback.warning.mockClear();
    document.body.innerHTML = '';
  });

  it('abre la venta desde cualquier pantalla', () => {
    const atajo = servicio();
    const evento = f2();
    atajo.alPresionarF2(evento);
    expect(evento.defaultPrevented).toBe(true);
    expect(router.navigate).toHaveBeenCalledWith(['/ventas/nueva']);
  });

  it('sin permiso, con la tecla sostenida o ya en la captura no navega', () => {
    const atajo = servicio();
    atajo.alPresionarF2(f2(true));
    router.url = '/ventas/nueva?x=1';
    atajo.alPresionarF2(f2());
    puede = false;
    router.url = '/compras';
    const evento = f2();
    atajo.alPresionarF2(evento);
    expect(evento.defaultPrevented).toBe(false);
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('con cambios pendientes avisa y no navega', () => {
    pendientes = true;
    const atajo = servicio();
    atajo.alPresionarF2(f2());
    expect(router.navigate).not.toHaveBeenCalled();
    expect(feedback.warning).toHaveBeenCalledWith(
      'Tienes cambios sin guardar',
      'Guarda o descarta los cambios antes de abrir una venta.',
    );
  });

  it('con el foco en un formulario con datos avisa; vacío sí navega', () => {
    document.body.innerHTML = '<form><input id="folio" /><input type="checkbox" id="iva" /></form>';
    const folio = document.getElementById('folio') as HTMLInputElement;
    folio.focus();
    const atajo = servicio();
    atajo.alPresionarF2(f2());
    expect(router.navigate).toHaveBeenCalledTimes(1);

    folio.value = 'A-123';
    atajo.alPresionarF2(f2());
    expect(router.navigate).toHaveBeenCalledTimes(1);
    expect(feedback.warning).toHaveBeenCalled();
  });

  it('formularioConDatos revisa texto y casillas, y solo dentro de un formulario', () => {
    document.body.innerHTML =
      '<form><textarea></textarea><input type="checkbox" /><button>Guardar</button></form><input id="suelto" value="x" />';
    const [area, casilla] = [
      document.querySelector('textarea')!,
      document.querySelector('input[type="checkbox"]') as HTMLInputElement,
    ];
    expect(formularioConDatos(area)).toBe(false);
    casilla.checked = true;
    expect(formularioConDatos(area)).toBe(true);
    expect(formularioConDatos(document.getElementById('suelto'))).toBe(false);
    expect(formularioConDatos(null)).toBe(false);
  });
});
