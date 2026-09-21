import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { AuthService } from './auth.service';

describe('AuthService', () => {
  let service: AuthService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [AuthService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('sends the login payload to the existing endpoint', () => {
    const payload = { correo: 'persona@ejemplo.com', contrasena: 'secreto', recordarme: false };

    service.login(payload).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/auth/login`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(payload);
    request.flush({ usuario: { id: '1', correo: payload.correo } });
    expect(service.currentUser()?.correo).toBe(payload.correo);
  });

  it('keeps registration, verification, session and logout endpoint contracts', () => {
    const registration = {
      nombre_completo: 'Juan Pérez',
      correo: 'juan@ejemplo.com',
      telefono: '',
      contrasena: '12345678',
    };

    service.register(registration).subscribe();
    const registerRequest = http.expectOne(`${environment.apiUrl}/auth/register`);
    expect(registerRequest.request.body).toEqual(registration);
    registerRequest.flush({
      mensaje: 'Código enviado',
      desafio_id: 'challenge-1',
      correo_enmascarado: 'j***@ejemplo.com',
      reenviar_en_segundos: 60,
    });

    service.verifyEmail({ desafio_id: 'challenge-1', codigo: '123456' }).subscribe();
    const verifyRequest = http.expectOne(`${environment.apiUrl}/auth/verificar-correo`);
    expect(verifyRequest.request.body).toEqual({ desafio_id: 'challenge-1', codigo: '123456' });
    verifyRequest.flush({
      mensaje: 'Correo verificado',
      usuario: { id: '1', correo: registration.correo },
    });
    expect(service.currentUser()?.correo).toBe(registration.correo);

    service.resendVerification({ desafio_id: 'challenge-1' }).subscribe();
    const resendRequest = http.expectOne(`${environment.apiUrl}/auth/reenviar-verificacion`);
    expect(resendRequest.request.body).toEqual({ desafio_id: 'challenge-1' });
    resendRequest.flush({
      mensaje: 'Código reenviado',
      desafio_id: 'challenge-2',
      correo_enmascarado: 'j***@ejemplo.com',
      reenviar_en_segundos: 60,
    });

    service.me().subscribe();
    const meRequest = http.expectOne(`${environment.apiUrl}/auth/me`);
    expect(meRequest.request.method).toBe('GET');
    meRequest.flush({
      autenticado: true,
      usuario: { id: '1', correo: registration.correo },
    });
    expect(service.currentUser()?.correo).toBe(registration.correo);

    service.logout().subscribe();
    const logoutRequest = http.expectOne(`${environment.apiUrl}/auth/logout`);
    expect(logoutRequest.request.method).toBe('POST');
    expect(logoutRequest.request.body).toEqual({});
    logoutRequest.flush(null);
    expect(service.currentUser()).toBeNull();
  });
});
