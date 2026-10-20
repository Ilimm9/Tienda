export interface Permiso {
  id: string;
  codigo: string;
  codigo_modulo: string;
  nombre: string;
  descripcion?: string | null;
  requiere?: string[];
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
  permisos?: string[];
  permisos_faltantes?: string[];
  permisos_detalle?: Permiso[];
}

export function completarDependencias(seleccionados: ReadonlySet<string>, catalogo: Permiso[]): Set<string> {
  const resultado = new Set(seleccionados);
  const porCodigo = new Map(catalogo.map(permiso => [permiso.codigo, permiso.id]));
  for (const id of resultado) {
    const permiso = catalogo.find(item => item.id === id);
    for (const codigo of permiso?.requiere ?? []) {
      const requerido = porCodigo.get(codigo);
      if (requerido) resultado.add(requerido);
    }
  }
  return resultado;
}

export interface RolDetalle extends Omit<RolResumen, 'total_permisos'> {
  permisos: string[];
}

export interface MiembroRoles {
  membresia_id: string;
  usuario_id: string;
  correo: string;
  /** Nombre del empleado vinculado o del perfil; vacío si no hay ninguno. */
  nombre: string;
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
  autorizar_dependencias_existentes?: boolean;
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
