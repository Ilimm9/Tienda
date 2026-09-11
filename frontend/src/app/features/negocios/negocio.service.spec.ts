import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { NegocioService } from './negocio.service';

describe('NegocioService', () => {
  let service: NegocioService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [NegocioService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(NegocioService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lista negocios por estado', () => {
    service.listar('archivado').subscribe((response) => expect(response.total).toBe(0));

    const request = http.expectOne(
      (candidate) =>
        candidate.url === `${environment.apiUrl}/negocios` &&
        candidate.params.get('estado') === 'archivado',
    );
    expect(request.request.method).toBe('GET');
    request.flush({ items: [], total: 0 });
  });

  it('crea negocio con payload intacto', () => {
    const payload = {
      nombre_comercial: 'Tienda Centro',
      codigo_moneda: 'MXN',
      zona_horaria: 'America/Mexico_City',
    };

    service.crear(payload).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/negocios`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(payload);
    request.flush({});
  });

  it('usa endpoints de detalle, edición, archivado y restauración', () => {
    const id = '11111111-1111-4111-8111-111111111111';

    service.obtener(id).subscribe();
    expect(http.expectOne(`${environment.apiUrl}/negocios/${id}`).request.method).toBe('GET');

    service
      .actualizar(id, {
        nombre_comercial: 'Tienda Centro',
        razon_social: null,
        rfc: null,
        telefono: null,
        correo: null,
        codigo_moneda: 'MXN',
        zona_horaria: 'America/Mexico_City',
        direccion: null,
      })
      .subscribe();
    expect(http.expectOne(`${environment.apiUrl}/negocios/${id}`).request.method).toBe('PATCH');

    service.archivar(id).subscribe();
    expect(http.expectOne(`${environment.apiUrl}/negocios/${id}`).request.method).toBe('DELETE');

    service.restaurar(id).subscribe();
    const restore = http.expectOne(`${environment.apiUrl}/negocios/${id}/restaurar`);
    expect(restore.request.method).toBe('POST');
    expect(restore.request.body).toEqual({});
  });
});
