import { HttpClient } from '@angular/common/http';
import { computed, DestroyRef, inject, Injectable, signal } from '@angular/core';
import { finalize, map, Observable, of, shareReplay, Subject, tap } from 'rxjs';

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

/** Aviso que una pestaña manda a las demás del mismo navegador cuando la sesión cambia. */
export type AvisoSesion = 'inicio' | 'cierre';

const canalSesion = 'tienda.sesion';

/** Primera pantalla tras tener sesión: las cuentas de plataforma no operan tiendas. */
export function destinoTrasSesion(usuario: AuthenticatedUser): string {
  return usuario.roles_plataforma?.includes('ADMIN_LECTURA') ? '/administracion-plataforma' : '/inicio';
}

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
  // La cookie de sesión es compartida por todas las pestañas; el canal solo avisa que cambió.
  private readonly canal = typeof BroadcastChannel === 'undefined' ? null : new BroadcastChannel(canalSesion);
  private readonly avisosRemotos = new Subject<AvisoSesion>();
  /** Inicios y cierres de sesión hechos en otra pestaña del mismo navegador. */
  readonly sesionEnOtraPestana = this.avisosRemotos.asObservable();

  constructor() {
    if (!this.canal) return;
    this.canal.onmessage = ({ data }: MessageEvent) => {
      if (data !== 'inicio' && data !== 'cierre') return;
      if (data === 'cierre') this.currentUser.set(null);
      this.avisosRemotos.next(data);
    };
    inject(DestroyRef).onDestroy(() => this.canal?.close());
  }

  login(payload: LoginRequest): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(`${this.url}/login`, payload)
      .pipe(tap(({ usuario }) => this.sesionIniciada(usuario)));
  }

  register(payload: RegisterRequest): Observable<RegisterResponse> {
    return this.http.post<RegisterResponse>(`${this.url}/register`, payload);
  }

  verifyEmail(payload: VerifyEmailRequest): Observable<VerifyEmailResponse> {
    return this.http
      .post<VerifyEmailResponse>(`${this.url}/verificar-correo`, payload)
      .pipe(tap(({ usuario }) => this.sesionIniciada(usuario)));
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
    // Se limpia antes de que quien se suscribe navegue: invitadoGuard debe ver la sesión cerrada.
    let avisado = false;
    const cerrar = () => {
      this.currentUser.set(null);
      if (avisado) return;
      avisado = true;
      this.avisar('cierre');
    };
    return this.http
      .post<void>(`${this.url}/logout`, {})
      .pipe(tap({ next: cerrar, error: cerrar }), finalize(cerrar));
  }

  private sesionIniciada(usuario: AuthenticatedUser): void {
    this.currentUser.set(usuario);
    this.avisar('inicio');
  }

  private avisar(aviso: AvisoSesion): void {
    try {
      this.canal?.postMessage(aviso);
    } catch {
      // Canal cerrado: las demás pestañas lo notarán en su siguiente consulta a la API.
    }
  }
}
