import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { SucursalService } from './sucursal.service';

describe('SucursalService', () => {
  let service: SucursalService;
  let http: HttpTestingController;
  const negocioId = '11111111-1111-4111-8111-111111111111';
  const sucursalId = '22222222-2222-4222-8222-222222222222';
  const baseUrl = `${environment.apiUrl}/negocios/${negocioId}/administracion/sucursales`;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [SucursalService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(SucursalService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lista con estado y búsqueda normalizada', () => {
    service.listar(negocioId, 'archivado', '  centro  ').subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === baseUrl &&
        candidate.params.get('estado') === 'archivado' &&
        candidate.params.get('buscar') === 'centro',
    );
    expect(request.request.method).toBe('GET');
    request.flush({ items: [], total: 0 });
  });

  it('usa el CRUD administrativo sin modificar el endpoint legacy', () => {
    const createPayload = { codigo: 'SUC-001', nombre: 'Matriz', es_principal: true };
    service.crear(negocioId, createPayload).subscribe();
    const create = http.expectOne(baseUrl);
    expect(create.request.method).toBe('POST');
    expect(create.request.body).toEqual(createPayload);
    create.flush({});

    service.obtener(negocioId, sucursalId).subscribe();
    const detail = http.expectOne(`${baseUrl}/${sucursalId}`);
    expect(detail.request.method).toBe('GET');
    detail.flush({});

    const updatePayload = {
      nombre: 'Matriz Centro',
      telefono: null,
      es_principal: true,
      direccion: null,
    };
    service.actualizar(negocioId, sucursalId, updatePayload).subscribe();
    const update = http.expectOne(`${baseUrl}/${sucursalId}`);
    expect(update.request.method).toBe('PATCH');
    expect(update.request.body).toEqual(updatePayload);
    update.flush({});

    service.archivar(negocioId, sucursalId).subscribe();
    const archive = http.expectOne(`${baseUrl}/${sucursalId}`);
    expect(archive.request.method).toBe('DELETE');
    archive.flush(null);

    service.restaurar(negocioId, sucursalId).subscribe();
    const restore = http.expectOne(`${baseUrl}/${sucursalId}/restaurar`);
    expect(restore.request.method).toBe('POST');
    expect(restore.request.body).toEqual({});
    restore.flush({});
  });
});
