export type EstadoEmpleado = 'pendiente' | 'activo' | 'suspendido' | 'terminado';

export interface EmpleadoResumen {
  id: string;
  negocio_id: string;
  numero_empleado: string | null;
  nombre_completo: string;
  correo: string | null;
  telefono: string | null;
  puesto: string | null;
  estado: EstadoEmpleado;
  tiene_cuenta: boolean;
  creado_en: string;
  actualizado_en: string;
}

export interface EmpleadoDetalle {
  id: string;
  negocio_id: string;
  membresia_id: string | null;
  numero_empleado: string | null;
  nombre: string;
  segundo_nombre: string | null;
  primer_apellido: string;
  segundo_apellido: string | null;
  nombre_completo: string;
  correo: string | null;
  telefono: string | null;
  puesto: string | null;
  estado: EstadoEmpleado;
  contratado_en: string | null;
  terminado_en: string | null;
  tiene_cuenta: boolean;
  creado_en: string;
  actualizado_en: string;
}

export interface CrearEmpleadoPayload {
  numero_empleado?: string | null;
  nombre: string;
  segundo_nombre?: string | null;
  primer_apellido: string;
  segundo_apellido?: string | null;
  correo?: string | null;
  telefono?: string | null;
  puesto?: string | null;
  contratado_en?: string | null;
}

export interface ActualizarEmpleadoPayload {
  numero_empleado?: string | null;
  nombre?: string;
  segundo_nombre?: string | null;
  primer_apellido?: string;
  segundo_apellido?: string | null;
  correo?: string | null;
  telefono?: string | null;
  puesto?: string | null;
  estado?: EstadoEmpleado;
  contratado_en?: string | null;
  terminado_en?: string | null;
}

export interface EmpleadoListResponse {
  items: EmpleadoResumen[];
  total: number;
}

export interface EmpleadoApiError {
  codigo: string;
  mensaje: string;
  campos?: Record<string, string>;
}
