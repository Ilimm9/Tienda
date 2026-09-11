export interface DireccionNegocio {
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

export interface NegocioResumen {
  id: string;
  slug: string;
  nombre_comercial: string;
  rfc: string | null;
  tipo_miembro: 'propietario' | 'miembro';
  estado: 'activo' | 'archivado';
  tiene_sucursales: boolean;
  total_sucursales: number;
  creado_en: string;
}

export interface NegocioDetalle extends NegocioResumen {
  razon_social: string | null;
  telefono: string | null;
  correo: string | null;
  codigo_moneda: string;
  zona_horaria: string;
  actualizado_en: string;
  archivado_en: string | null;
  direccion: DireccionNegocio | null;
}

export interface NegocioListResponse {
  items: NegocioResumen[];
  total: number;
}

export interface CrearNegocioPayload {
  nombre_comercial: string;
  razon_social?: string;
  rfc?: string;
  telefono?: string;
  correo?: string;
  codigo_moneda: string;
  zona_horaria: string;
  direccion?: Omit<DireccionNegocio, 'id'>;
}

export interface ActualizarNegocioPayload {
  nombre_comercial: string;
  razon_social: string | null;
  rfc: string | null;
  telefono: string | null;
  correo: string | null;
  codigo_moneda: string;
  zona_horaria: string;
  direccion: Omit<DireccionNegocio, 'id'> | null;
}

export interface ApiErrorResponse {
  codigo?: string;
  mensaje?: string;
  campos?: Record<string, string>;
}
