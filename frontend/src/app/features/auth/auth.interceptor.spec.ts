import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AuthService } from './auth.service';
import { authInterceptor } from './auth.interceptor';

describe('authInterceptor', () => {
  it('enables credentials for API requests', () => {
    TestBed.configureTestingModule({
      providers: [
        AuthService,
        provideHttpClient(withInterceptors([authInterceptor])),
        provideHttpClientTesting(),
      ],
    });
    const service = TestBed.inject(AuthService);
    const http = TestBed.inject(HttpTestingController);

    service.me().subscribe();

    const request = http.expectOne((candidate) => candidate.url.endsWith('/auth/me'));
    expect(request.request.withCredentials).toBe(true);
    request.flush({ usuario: { id: '1', correo: 'persona@ejemplo.com' } });
    http.verify();
  });
});
