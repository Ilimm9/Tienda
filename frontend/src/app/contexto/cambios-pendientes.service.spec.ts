import { TestBed } from '@angular/core/testing';

import { CambiosPendientesService } from './cambios-pendientes.service';

describe('CambiosPendientesService', () => {
  let service: CambiosPendientesService;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [CambiosPendientesService] });
    service = TestBed.inject(CambiosPendientesService);
  });

  afterEach(() => TestBed.resetTestingModule());

  it('no reporta pendientes sin declaraciones registradas', () => {
    expect(service.hayPendientes()).toBe(false);
  });

  it('reporta pendientes cuando alguna declaración es verdadera', () => {
    service.registrar(() => false);
    service.registrar(() => true);

    expect(service.hayPendientes()).toBe(true);
  });

  it('deja de reportar después de dar de baja la declaración', () => {
    const baja = service.registrar(() => true);

    baja();

    expect(service.hayPendientes()).toBe(false);
  });
});
