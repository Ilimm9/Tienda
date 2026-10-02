export type EstadoInvitacion = 'pendiente' | 'aceptada' | 'expirada' | 'cancelada';

export interface InvitacionResumen {
  id: string;
  negocio_id: string;
  empleado_id: string | null;
  nombre_empleado: string;
  correo: string;
  sucursal_id: string | null;
  nombre_sucursal: string;
  rol_predeterminado_id: string | null;
  nombre_rol: string;
  estado: EstadoInvitacion;
  expira_en: string;
  aceptado_en: string | null;
  creado_en: string;
}

export interface InvitacionCreada {
  invitacion: InvitacionResumen;
  token: string;
  correo_enviado: boolean;
}

/** Lo que ve quien abre el enlace: nunca incluye correo completo ni nombre del empleado. */
export interface InvitacionPublica {
  correo_enmascarado: string;
  nombre_negocio: string;
  nombre_sucursal: string;
  nombre_rol: string;
  expira_en: string;
  requiere_cuenta: boolean;
}

export interface CrearInvitacionPayload {
  empleado_id: string;
  sucursal_id: string;
  rol_predeterminado_id: string;
}

export interface FiltroInvitaciones {
  estado?: EstadoInvitacion | '';
  sucursalId?: string;
  sinAceptar?: boolean;
}

/** El correo no viaja: el servidor lo toma de la invitación. */
export interface RegistroInvitacionPayload {
  nombres: string;
  apellidos: string;
  telefono: string;
  contrasena: string;
}

export interface DesafioRegistro {
  desafio_id: string;
  correo_enmascarado: string;
  reenviar_en_segundos: number;
}

export interface InvitacionAceptada {
  aceptada: boolean;
  negocio_id: string;
  sucursal_id: string | null;
}

export interface InvitacionListResponse {
  items: InvitacionResumen[];
  total: number;
}
