export interface LoginRequest {
  correo: string;
  contrasena: string;
  recordarme: boolean;
}

export interface LoginResponse {
  usuario: {
    id: string;
    correo: string;
  };
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
