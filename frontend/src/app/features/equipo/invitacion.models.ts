export type EstadoInvitacion = 'pendiente' | 'aceptada' | 'expirada' | 'cancelada';

export interface InvitacionResumen {
  id: string;
  negocio_id: string;
  empleado_id: string | null;
  nombre_empleado: string;
  correo: string;
  rol_predeterminado_id: string | null;
  estado: EstadoInvitacion;
  expira_en: string;
  aceptado_en: string | null;
  creado_en: string;
}

export interface InvitacionCreada {
  invitacion: InvitacionResumen;
  token: string;
}

export interface InvitacionPublica {
  correo: string;
  nombre_negocio: string;
  nombre_empleado: string;
  expira_en: string;
  requiere_cuenta: boolean;
}

export interface CrearInvitacionPayload {
  empleado_id: string;
  rol_predeterminado_id?: string | null;
}

export interface InvitacionListResponse {
  items: InvitacionResumen[];
  total: number;
}
