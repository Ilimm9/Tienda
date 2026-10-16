import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { PlataformaService } from './plataforma.service';

describe('PlataformaService', () => {
  let service: PlataformaService;
  let http: HttpTestingController;
  const base = `${environment.apiUrl}/plataforma`;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [PlataformaService, provideHttpClient(), provideHttpClientTesting()] });
    service = TestBed.inject(PlataformaService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => { http.verify(); TestBed.resetTestingModule(); });

  it('consulta el resumen global', () => {
    service.resumen().subscribe();
    const request = http.expectOne(`${base}/resumen`);
    expect(request.request.method).toBe('GET');
    request.flush({ actividad: [] });
  });

  it('pagina y filtra negocios', () => {
    service.negocios('Centro', 'activo', 2, 20).subscribe();
    const request = http.expectOne((candidate) => candidate.url === `${base}/negocios`);
    expect(request.request.params.get('q')).toBe('Centro');
    expect(request.request.params.get('estado')).toBe('activo');
    expect(request.request.params.get('page')).toBe('2');
    request.flush({ items: [], total: 0, page: 2, page_size: 20 });
  });

  it('consulta detalle y usuarios sin rutas tenant', () => {
    const id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
    service.negocio(id).subscribe();
    http.expectOne(`${base}/negocios/${id}`).flush({});
    service.usuarios(id).subscribe();
    http.expectOne(`${base}/negocios/${id}/usuarios`).flush({ items: [], total: 0 });
  });
});
