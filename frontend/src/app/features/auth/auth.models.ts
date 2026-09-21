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
  desafio_id: string;
  correo_enmascarado: string;
  reenviar_en_segundos: number;
}

export interface VerifyEmailRequest {
  desafio_id: string;
  codigo: string;
}

export interface VerifyEmailResponse extends LoginResponse {
  mensaje: string;
}

export interface ResendVerificationRequest {
  desafio_id: string;
}
