export interface Permiso {
  id: string;
  codigo: string;
  codigo_modulo: string;
  nombre: string;
  descripcion?: string | null;
}

export interface RolResumen {
  id: string;
  negocio_id: string;
  codigo: string;
  nombre: string;
  descripcion: string | null;
  es_rol_sistema: boolean;
  activo: boolean;
  total_permisos: number;
  total_miembros: number;
  creado_en: string;
  actualizado_en: string;
}

export interface RolDetalle extends Omit<RolResumen, 'total_permisos'> {
  permisos: string[];
}

export interface MiembroRoles {
  membresia_id: string;
  usuario_id: string;
  correo: string;
  tipo_miembro: 'propietario' | 'miembro';
  estado: string;
  roles: string[];
}

export interface ListaRespuesta<T> {
  items: T[];
  total: number;
}

export interface CrearRolPayload {
  codigo: string;
  nombre: string;
  descripcion?: string | null;
  permisos: string[];
}

export interface ActualizarRolPayload {
  nombre?: string;
  descripcion?: string | null;
  activo?: boolean;
  permisos?: string[];
}

export interface RolApiError {
  codigo: string;
  mensaje: string;
  campos?: Record<string, string>;
}
