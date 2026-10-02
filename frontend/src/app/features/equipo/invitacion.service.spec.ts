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

  const sucursalId = '44444444-4444-4444-8444-444444444444';
  const rolId = '55555555-5555-4555-8555-555555555555';

  it('crea una invitación con empleado, sucursal y rol', () => {
    const payload = { empleado_id: empleadoId, sucursal_id: sucursalId, rol_predeterminado_id: rolId };
    service.crear(negocioId, payload).subscribe();

    const request = http.expectOne(baseUrl);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(payload);
    request.flush({ invitacion: {}, token: 'abc' });
  });

  it('filtra el listado por estado y sucursal', () => {
    service.listar(negocioId, { estado: 'pendiente', sucursalId }).subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === baseUrl &&
        candidate.params.get('estado') === 'pendiente' &&
        candidate.params.get('sucursal_id') === sucursalId,
    );
    request.flush({ items: [], total: 0 });
  });

  it('pide las no aceptadas sin mandar estado', () => {
    service.listar(negocioId, { sinAceptar: true, estado: 'aceptada' }).subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === baseUrl &&
        candidate.params.get('sin_aceptar') === 'true' &&
        !candidate.params.has('estado'),
    );
    request.flush({ items: [], total: 0 });
  });

  it('reenvía conservando al destinatario cuando no hay correo nuevo', () => {
    service.reenviar(negocioId, invitacionId).subscribe();

    const request = http.expectOne(`${baseUrl}/${invitacionId}/reenviar`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({});
    request.flush({ invitacion: {}, token: 'abc' });
  });

  it('reenvía corrigiendo el correo', () => {
    service.reenviar(negocioId, invitacionId, 'nuevo@tienda.mx').subscribe();

    const request = http.expectOne(`${baseUrl}/${invitacionId}/reenviar`);
    expect(request.request.body).toEqual({ correo: 'nuevo@tienda.mx' });
    request.flush({ invitacion: {}, token: 'abc' });
  });

  it('registra la cuenta del invitado sin enviar correo', () => {
    const payload = { nombres: 'Ana', apellidos: 'Ruiz', telefono: '', contrasena: 'contrasena-segura' };
    service.registrar('token-publico', payload).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/invitaciones/token-publico/registro`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(payload);
    request.flush({ desafio_id: 'd', correo_enmascarado: 'a***@tienda.mx', reenviar_en_segundos: 60 });
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
    request.flush({ aceptada: true, negocio_id: negocioId, sucursal_id: null });
  });

  it('arma el enlace copiable sobre el origen actual', () => {
    expect(service.enlaceDeToken('abc')).toBe(`${window.location.origin}/invitacion/abc`);
  });
});
