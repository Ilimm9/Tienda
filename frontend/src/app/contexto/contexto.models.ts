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
  /** Códigos efectivos de la membresía. Solo ocultan o deshabilitan en la interfaz; la API autoriza. */
  permisos: string[];
  sucursales: ContextoSucursal[];
}

export interface ContextoOpcionesResponse {
  items: ContextoNegocio[];
  total: number;
}

export type EstadoContexto = 'cargando' | 'requiere_negocio' | 'listo' | 'sin_sucursal' | 'error';

/** Qué le falta a la cuenta para operar: decide si se ofrece el asistente de alta inicial. */
export type EstadoAlta = 'requiere_empresa' | 'requiere_sucursal' | 'sin_asignacion' | 'listo';
