export interface AsignacionResumen {
  id: string;
  negocio_id: string;
  empleado_id: string;
  sucursal_id: string;
  codigo_sucursal: string;
  nombre_sucursal: string;
  es_principal: boolean;
  activo: boolean;
  asignado_en: string;
  finalizado_en: string | null;
}

export interface AsignacionListResponse {
  items: AsignacionResumen[];
  total: number;
}

export interface AsignarSucursalPayload {
  sucursal_id: string;
  es_principal: boolean;
}
