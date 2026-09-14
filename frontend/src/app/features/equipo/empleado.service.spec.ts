import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { EmpleadoService } from './empleado.service';

describe('EmpleadoService', () => {
  let service: EmpleadoService;
  let http: HttpTestingController;
  const negocioId = '11111111-1111-4111-8111-111111111111';
  const empleadoId = '22222222-2222-4222-8222-222222222222';
  const baseUrl = `${environment.apiUrl}/negocios/${negocioId}/administracion/empleados`;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [EmpleadoService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(EmpleadoService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  it('lista sin filtros cuando no se indican', () => {
    service.listar(negocioId).subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === baseUrl && !candidate.params.has('estado') && !candidate.params.has('buscar'),
    );
    expect(request.request.method).toBe('GET');
    request.flush({ items: [], total: 0 });
  });

  it('envía estado y búsqueda normalizada', () => {
    service.listar(negocioId, 'activo', '  ana  ').subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === baseUrl &&
        candidate.params.get('estado') === 'activo' &&
        candidate.params.get('buscar') === 'ana',
    );
    request.flush({ items: [], total: 0 });
  });

  it('registra un empleado sin exigir cuenta', () => {
    service.crear(negocioId, { nombre: 'Ana', primer_apellido: 'López' }).subscribe();

    const request = http.expectOne(baseUrl);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ nombre: 'Ana', primer_apellido: 'López' });
    request.flush({});
  });

  it('actualiza con PATCH parcial', () => {
    service.actualizar(negocioId, empleadoId, { estado: 'terminado' }).subscribe();

    const request = http.expectOne(`${baseUrl}/${empleadoId}`);
    expect(request.request.method).toBe('PATCH');
    expect(request.request.body).toEqual({ estado: 'terminado' });
    request.flush({});
  });

  it('obtiene el detalle por identificador', () => {
    service.obtener(negocioId, empleadoId).subscribe();

    const request = http.expectOne(`${baseUrl}/${empleadoId}`);
    expect(request.request.method).toBe('GET');
    request.flush({});
  });
});
