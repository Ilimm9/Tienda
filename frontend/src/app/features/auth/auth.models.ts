export interface LoginRequest {
  correo: string;
  contrasena: string;
  recordarme: boolean;
}

export interface AuthenticatedUser {
  id: string;
  correo: string;
}

export interface LoginResponse {
  usuario: AuthenticatedUser;
}

export interface SessionResponse {
  autenticado: true;
  usuario: AuthenticatedUser;
}

export interface RegisterRequest {
  nombre_completo: string;
  correo: string;
  telefono: string;
  contrasena: string;
}

export interface RegisterResponse {
  mensaje: string;
}
