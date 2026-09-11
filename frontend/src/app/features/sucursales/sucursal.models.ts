export interface DireccionSucursal {
  id?: string;
  codigo_pais: string;
  estado?: string | null;
  municipio?: string | null;
  ciudad?: string | null;
  colonia?: string | null;
  codigo_postal?: string | null;
  calle?: string | null;
  numero_exterior?: string | null;
  numero_interior?: string | null;
  referencias?: string | null;
}

export interface SucursalResumen {
  id: string;
  negocio_id: string;
  codigo: string;
  nombre: string;
  telefono: string | null;
  direccion_resumida: string | null;
  es_principal: boolean;
  estado: 'activo' | 'archivado';
  tipo_miembro: 'propietario' | 'miembro';
  creado_en: string;
  actualizado_en: string;
}

export interface SucursalDetalle extends Omit<SucursalResumen, 'direccion_resumida'> {
  eliminado_en: string | null;
  direccion: DireccionSucursal | null;
}

export interface SucursalListResponse {
  items: SucursalResumen[];
  total: number;
}

export interface CrearSucursalPayload {
  codigo: string;
  nombre: string;
  telefono?: string;
  es_principal: boolean;
  direccion?: Omit<DireccionSucursal, 'id'>;
}

export interface ActualizarSucursalPayload {
  nombre: string;
  telefono: string | null;
  es_principal: boolean;
  direccion: Omit<DireccionSucursal, 'id'> | null;
}

export interface SucursalApiError {
  codigo?: string;
  mensaje?: string;
  campos?: Record<string, string>;
}
