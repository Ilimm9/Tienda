import { ZONA_HORARIA_MEXICO } from '../../shared/fecha-mexico.pipe';
import { PeriodoVentas } from './ventas.models';

const hora = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});
const diaMes = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  day: 'numeric',
  month: 'short',
});
const diaLargo = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  weekday: 'long',
  day: 'numeric',
  month: 'long',
});
const diaLargoConAnio = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  year: 'numeric',
});
const mesAnio = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  month: 'long',
  year: 'numeric',
});
const dia = new Intl.DateTimeFormat('es-MX', { timeZone: ZONA_HORARIA_MEXICO, day: 'numeric' });
const diaYMes = new Intl.DateTimeFormat('es-MX', {
  timeZone: ZONA_HORARIA_MEXICO,
  day: 'numeric',
  month: 'long',
});
const mes = new Intl.DateTimeFormat('es-MX', { timeZone: ZONA_HORARIA_MEXICO, month: 'long' });

const mayuscula = (texto: string) => texto.charAt(0).toUpperCase() + texto.slice(1);

/** «13:42» en hora de Ciudad de México. */
export function horaVenta(iso: string): string {
  return hora.format(new Date(iso));
}

/** «23 oct · 13:42», para la semana o el mes. */
export function fechaHoraCorta(iso: string): string {
  return `${diaMes.format(new Date(iso)).replace('.', '')} · ${horaVenta(iso)}`;
}

/** «Viernes 23 de octubre de 2026, 13:15». */
export function fechaHoraLarga(iso: string): string {
  return `${mayuscula(diaLargoConAnio.format(new Date(iso)).replace(',', ''))}, ${horaVenta(iso)}`;
}

/** Texto junto al selector: el día, la semana (del lunes al domingo) o el mes. `hasta` es exclusivo. */
export function etiquetaPeriodo(periodo: PeriodoVentas, desde?: string, hasta?: string): string {
  if (!desde || !hasta) return '';
  const inicio = new Date(desde);
  const fin = new Date(new Date(hasta).getTime() - 1);
  if (periodo === 'hoy') return mayuscula(diaLargo.format(inicio).replace(',', ''));
  if (periodo === 'mes') return mayuscula(mesAnio.format(inicio));
  const mismoMes = mes.format(inicio) === mes.format(fin);
  return mismoMes
    ? `Del ${dia.format(inicio)} al ${diaYMes.format(fin)}`
    : `Del ${diaYMes.format(inicio)} al ${diaYMes.format(fin)}`;
}
