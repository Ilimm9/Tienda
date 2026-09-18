import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { AsignacionService } from './asignacion.service';

describe('AsignacionService', () => {
  let service: AsignacionService;
  let http: HttpTestingController;
  const negocioId = '11111111-1111-4111-8111-111111111111';
  const empleadoId = '22222222-2222-4222-8222-222222222222';
  const sucursalId = '33333333-3333-4333-8333-333333333333';
  const asignacionId = '44444444-4444-4444-8444-444444444444';
  const baseUrl = `${environment.apiUrl}/negocios/${negocioId}/administracion/empleados/${empleadoId}/sucursales`;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [AsignacionService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AsignacionService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  it('lista solo asignaciones activas por omisión', () => {
    service.listar(negocioId, empleadoId).subscribe();

    const request = http.expectOne(
      (candidate) => candidate.url === baseUrl && !candidate.params.has('estado'),
    );
    expect(request.request.method).toBe('GET');
    request.flush({ items: [], total: 0 });
  });

  it('incluye finalizadas cuando se solicita el historial', () => {
    service.listar(negocioId, empleadoId, true).subscribe();

    const request = http.expectOne(
      (candidate) => candidate.url === baseUrl && candidate.params.get('estado') === 'todos',
    );
    request.flush({ items: [], total: 0 });
  });

  it('asigna una sucursal indicando si es principal', () => {
    service.asignar(negocioId, empleadoId, { sucursal_id: sucursalId, es_principal: true }).subscribe();

    const request = http.expectOne(baseUrl);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ sucursal_id: sucursalId, es_principal: true });
    request.flush({ items: [], total: 0 });
  });

  it('promueve una asignación a principal', () => {
    service.establecerPrincipal(negocioId, empleadoId, asignacionId).subscribe();

    const request = http.expectOne(`${baseUrl}/${asignacionId}/principal`);
    expect(request.request.method).toBe('POST');
    request.flush({ items: [], total: 0 });
  });

  it('retira una asignación sin borrar el historial', () => {
    service.finalizar(negocioId, empleadoId, asignacionId).subscribe();

    const request = http.expectOne(`${baseUrl}/${asignacionId}`);
    expect(request.request.method).toBe('DELETE');
    request.flush(null);
  });
});
