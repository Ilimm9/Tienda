import { EstadoEmpleado } from './empleado.models';

/** Iniciales para avatares de Equipo: hasta dos palabras del nombre; sin nombre, la del respaldo. */
export function iniciales(nombre: string | null | undefined, respaldo = ''): string {
  const partes = (nombre ?? '').trim().split(/\s+/).filter(Boolean);
  if (partes.length) return partes.slice(0, 2).map((parte) => parte.charAt(0)).join('').toUpperCase();
  return respaldo.trim().charAt(0).toUpperCase() || '?';
}

export const ETIQUETA_ESTADO_EMPLEADO: Record<EstadoEmpleado, string> = {
  pendiente: 'Pendiente',
  activo: 'Activo',
  suspendido: 'Suspendido',
  terminado: 'Dado de baja',
};

/** Variante de `.pastilla` del kit para cada estado laboral. */
export const PASTILLA_ESTADO_EMPLEADO: Record<EstadoEmpleado, string> = {
  pendiente: 'pastilla--info',
  activo: 'pastilla--exito',
  suspendido: 'pastilla--atencion',
  terminado: 'pastilla--neutra',
};

/** Variante de `.avatar`: neutro para quien ya no opera (suspendido o de baja). */
export function avatarDeEstado(estado: EstadoEmpleado): string {
  return estado === 'suspendido' || estado === 'terminado' ? 'avatar--neutro' : '';
}
