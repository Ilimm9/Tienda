import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';

import { environment } from '../../../environments/environment';
import { AuthService } from './auth.service';
import { authInterceptor } from './auth.interceptor';

describe('authInterceptor', () => {
  const router = { navigate: vi.fn().mockResolvedValue(true) };
  let service: AuthService;
  let http: HttpTestingController;
  let client: HttpClient;

  beforeEach(() => {
    vi.clearAllMocks();
    TestBed.configureTestingModule({
      providers: [
        AuthService,
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
        { provide: Router, useValue: router },
      ],
    });
    service = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
    client = TestBed.inject(HttpClient);
  });

  afterEach(() => http.verify());

  it('enables credentials for API requests', () => {
    service.me().subscribe();

    const request = http.expectOne((candidate) => candidate.url.endsWith('/auth/me'));
    expect(request.request.withCredentials).toBe(true);
    request.flush({ usuario: { id: '1', correo: 'persona@ejemplo.com' } });
  });

  it('una sesión vencida limpia el usuario en memoria y lleva al login', () => {
    service.currentUser.set({ id: '1', correo: 'persona@ejemplo.com' } as never);
    client.get(`${environment.apiUrl}/negocios/x/catalogo/productos`).subscribe({ error: () => undefined });
    http.expectOne(`${environment.apiUrl}/negocios/x/catalogo/productos`).flush({}, { status: 401, statusText: 'Unauthorized' });

    expect(service.currentUser()).toBeNull();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });

  it('respeta los 401 que manejan el login, la sesión y la invitación pública', () => {
    service.currentUser.set({ id: '1', correo: 'persona@ejemplo.com' } as never);
    client.post(`${environment.apiUrl}/auth/login`, {}).subscribe({ error: () => undefined });
    http.expectOne(`${environment.apiUrl}/auth/login`).flush({}, { status: 401, statusText: 'Unauthorized' });
    client.post(`${environment.apiUrl}/invitaciones/token/aceptar`, {}).subscribe({ error: () => undefined });
    http.expectOne(`${environment.apiUrl}/invitaciones/token/aceptar`).flush({}, { status: 401, statusText: 'Unauthorized' });

    expect(service.currentUser()).not.toBeNull();
    expect(router.navigate).not.toHaveBeenCalled();
  });

  it('sin sesión iniciada un 401 no redirige', () => {
    client.get(`${environment.apiUrl}/negocios/x/sucursales`).subscribe({ error: () => undefined });
    http.expectOne(`${environment.apiUrl}/negocios/x/sucursales`).flush({}, { status: 401, statusText: 'Unauthorized' });
    expect(router.navigate).not.toHaveBeenCalled();
  });
});
