export interface ContextoSucursal {
  id: string;
  codigo: string;
  nombre: string;
  es_principal: boolean;
}

export interface ContextoNegocio {
  id: string;
  slug: string;
  nombre_comercial: string;
  tipo_miembro: 'propietario' | 'miembro';
  sucursales: ContextoSucursal[];
}

export interface ContextoOpcionesResponse {
  items: ContextoNegocio[];
  total: number;
}

export type EstadoContexto = 'cargando' | 'requiere_negocio' | 'listo' | 'sin_sucursal' | 'error';
