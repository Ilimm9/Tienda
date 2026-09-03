import { HttpClient } from '@angular/common/http';
import { inject, Injectable, signal } from '@angular/core';
import { finalize, Observable, tap } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  AuthenticatedUser,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  SessionResponse,
} from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly url = `${environment.apiUrl}/auth`;
  readonly currentUser = signal<AuthenticatedUser | null>(null);

  login(payload: LoginRequest): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(`${this.url}/login`, payload)
      .pipe(tap(({ usuario }) => this.currentUser.set(usuario)));
  }

  register(payload: RegisterRequest): Observable<RegisterResponse> {
    return this.http.post<RegisterResponse>(`${this.url}/register`, payload);
  }

  me(): Observable<SessionResponse> {
    return this.http
      .get<SessionResponse>(`${this.url}/me`)
      .pipe(tap(({ usuario }) => this.currentUser.set(usuario)));
  }

  logout(): Observable<void> {
    return this.http
      .post<void>(`${this.url}/logout`, {})
      .pipe(finalize(() => this.currentUser.set(null)));
  }
}
