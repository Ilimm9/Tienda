import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { InvitacionService } from './invitacion.service';

describe('InvitacionService', () => {
  let service: InvitacionService;
  let http: HttpTestingController;
  const negocioId = '11111111-1111-4111-8111-111111111111';
  const empleadoId = '22222222-2222-4222-8222-222222222222';
  const invitacionId = '33333333-3333-4333-8333-333333333333';
  const baseUrl = `${environment.apiUrl}/negocios/${negocioId}/administracion/invitaciones`;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [InvitacionService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(InvitacionService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  it('crea una invitación con el empleado y el rol inicial', () => {
    service.crear(negocioId, { empleado_id: empleadoId, rol_predeterminado_id: null }).subscribe();

    const request = http.expectOne(baseUrl);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ empleado_id: empleadoId, rol_predeterminado_id: null });
    request.flush({ invitacion: {}, token: 'abc' });
  });

  it('filtra el listado por estado', () => {
    service.listar(negocioId, 'pendiente').subscribe();

    const request = http.expectOne(
      (candidate) => candidate.url === baseUrl && candidate.params.get('estado') === 'pendiente',
    );
    request.flush({ items: [], total: 0 });
  });

  it('cancela una invitación pendiente', () => {
    service.cancelar(negocioId, invitacionId).subscribe();

    const request = http.expectOne(`${baseUrl}/${invitacionId}`);
    expect(request.request.method).toBe('DELETE');
    request.flush(null);
  });

  it('consulta el enlace público sin pasar por el negocio', () => {
    service.consultar('token-publico').subscribe();

    const request = http.expectOne(`${environment.apiUrl}/invitaciones/token-publico`);
    expect(request.request.method).toBe('GET');
    request.flush({});
  });

  it('acepta la invitación con el token del enlace', () => {
    service.aceptar('token-publico').subscribe();

    const request = http.expectOne(`${environment.apiUrl}/invitaciones/token-publico/aceptar`);
    expect(request.request.method).toBe('POST');
    request.flush(null);
  });

  it('arma el enlace copiable sobre el origen actual', () => {
    expect(service.enlaceDeToken('abc')).toBe(`${window.location.origin}/invitacion/abc`);
  });
});
