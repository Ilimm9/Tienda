/** Catálogo de puestos del negocio (`puestos_negocio`), común a todas sus sucursales. */
export interface Puesto {
  id: string;
  negocio_id: string;
  nombre: string;
  rol_sugerido_id: string | null;
  activo: boolean;
  creado_en: string;
  actualizado_en: string;
}

export interface CrearPuestoPayload {
  nombre: string;
  rol_sugerido_id?: string | null;
}

export interface ActualizarPuestoPayload {
  nombre?: string;
  rol_sugerido_id?: string | null;
  activo?: boolean;
}

export interface PuestoListResponse {
  items: Puesto[];
  total: number;
}
