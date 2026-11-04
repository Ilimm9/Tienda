import { HttpClient } from '@angular/common/http';
import { computed, inject, Injectable, signal } from '@angular/core';
import { finalize, map, Observable, of, shareReplay, tap } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  AuthenticatedUser,
  LoginRequest,
  LoginResponse,
  MessageResponse,
  PasswordResetConfirmRequest,
  PasswordResetRequest,
  RegisterRequest,
  RegisterResponse,
  ResendVerificationRequest,
  SessionResponse,
  VerifyEmailRequest,
  VerifyEmailResponse,
} from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly url = `${environment.apiUrl}/auth`;
  readonly currentUser = signal<AuthenticatedUser | null>(null);
  // Comparte una sola consulta de sesión entre guards que se resuelven en la misma navegación.
  private sesionEnVuelo: Observable<AuthenticatedUser> | null = null;
  readonly isPlatformAdmin = computed(() =>
    this.currentUser()?.roles_plataforma?.includes('ADMIN_LECTURA') ?? false,
  );

  login(payload: LoginRequest): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(`${this.url}/login`, payload)
      .pipe(tap(({ usuario }) => this.currentUser.set(usuario)));
  }

  register(payload: RegisterRequest): Observable<RegisterResponse> {
    return this.http.post<RegisterResponse>(`${this.url}/register`, payload);
  }

  verifyEmail(payload: VerifyEmailRequest): Observable<VerifyEmailResponse> {
    return this.http
      .post<VerifyEmailResponse>(`${this.url}/verificar-correo`, payload)
      .pipe(tap(({ usuario }) => this.currentUser.set(usuario)));
  }

  resendVerification(payload: ResendVerificationRequest): Observable<RegisterResponse> {
    return this.http.post<RegisterResponse>(`${this.url}/reenviar-verificacion`, payload);
  }

  requestPasswordReset(payload: PasswordResetRequest): Observable<MessageResponse> {
    return this.http.post<MessageResponse>(`${this.url}/solicitar-recuperacion`, payload);
  }

  resetPassword(payload: PasswordResetConfirmRequest): Observable<MessageResponse> {
    return this.http.post<MessageResponse>(`${this.url}/restablecer-contrasena`, payload);
  }

  /** Consulta la sesión en la API y actualiza el usuario en memoria (sin sesión, lo limpia). */
  me(): Observable<SessionResponse> {
    return this.http.get<SessionResponse>(`${this.url}/me`).pipe(
      tap({
        next: ({ usuario }) => this.currentUser.set(usuario),
        error: () => this.currentUser.set(null),
      }),
    );
  }

  /**
   * Sesión para los guards: se consulta una vez y luego se reutiliza en cada navegación.
   * Si la sesión vence, la API responde 401 y el interceptor limpia el usuario y lleva al login.
   */
  asegurarSesion(): Observable<AuthenticatedUser> {
    const usuario = this.currentUser();
    if (usuario) return of(usuario);
    this.sesionEnVuelo ??= this.me().pipe(
      map(({ usuario }) => usuario),
      finalize(() => (this.sesionEnVuelo = null)),
      shareReplay({ bufferSize: 1, refCount: false }),
    );
    return this.sesionEnVuelo;
  }

  /** La sesión terminó fuera del flujo de logout (por ejemplo, expiró en el servidor). */
  sesionTerminada(): void {
    this.currentUser.set(null);
  }

  logout(): Observable<void> {
    return this.http
      .post<void>(`${this.url}/logout`, {})
      .pipe(finalize(() => this.currentUser.set(null)));
  }
}
