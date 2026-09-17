import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { RolService } from './rol.service';

describe('RolService', () => {
  let service: RolService;
  let http: HttpTestingController;
  const negocioId = '11111111-1111-4111-8111-111111111111';
  const rolId = '22222222-2222-4222-8222-222222222222';
  const membresiaId = '33333333-3333-4333-8333-333333333333';
  const baseUrl = `${environment.apiUrl}/negocios/${negocioId}/administracion`;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [RolService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(RolService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  it('consulta el catálogo global de permisos del negocio', () => {
    service.permisos(negocioId).subscribe();

    const request = http.expectOne(`${baseUrl}/permisos`);
    expect(request.request.method).toBe('GET');
    request.flush({ items: [], total: 0 });
  });

  it('consulta los permisos efectivos de la sesión', () => {
    service.misPermisos(negocioId).subscribe();

    const request = http.expectOne(`${baseUrl}/mis-permisos`);
    expect(request.request.method).toBe('GET');
    request.flush({ items: ['roles.ver'], total: 1 });
  });

  it('lista solo roles activos por omisión', () => {
    service.listar(negocioId).subscribe();

    const request = http.expectOne(
      (candidate) => candidate.url === `${baseUrl}/roles` && !candidate.params.has('estado'),
    );
    request.flush({ items: [], total: 0 });
  });

  it('incluye roles inactivos cuando se solicita', () => {
    service.listar(negocioId, true).subscribe();

    const request = http.expectOne(
      (candidate) => candidate.url === `${baseUrl}/roles` && candidate.params.get('estado') === 'todos',
    );
    request.flush({ items: [], total: 0 });
  });

  it('crea, actualiza y elimina roles sobre el CRUD administrativo', () => {
    service.crear(negocioId, { codigo: 'CAJERO', nombre: 'Cajero', permisos: [] }).subscribe();
    const creacion = http.expectOne(`${baseUrl}/roles`);
    expect(creacion.request.method).toBe('POST');
    expect(creacion.request.body.codigo).toBe('CAJERO');
    creacion.flush({});

    service.actualizar(negocioId, rolId, { nombre: 'Cajero principal' }).subscribe();
    const actualizacion = http.expectOne(`${baseUrl}/roles/${rolId}`);
    expect(actualizacion.request.method).toBe('PATCH');
    actualizacion.flush({});

    service.eliminar(negocioId, rolId).subscribe();
    const eliminacion = http.expectOne(`${baseUrl}/roles/${rolId}`);
    expect(eliminacion.request.method).toBe('DELETE');
    eliminacion.flush(null);
  });

  it('reemplaza los roles de una membresía con PUT', () => {
    service.asignarRoles(negocioId, membresiaId, [rolId]).subscribe();

    const request = http.expectOne(`${baseUrl}/miembros/${membresiaId}/roles`);
    expect(request.request.method).toBe('PUT');
    expect(request.request.body).toEqual({ roles: [rolId] });
    request.flush(null);
  });
});
