import { etiquetaPeriodo, fechaHoraCorta, fechaHoraLarga, horaVenta } from './venta-fechas';

describe('fechas de ventas', () => {
  // 2026-10-23 13:15 en Ciudad de México (UTC−6).
  const venta = '2026-10-23T19:15:00Z';

  it('formatea horas y fechas en Ciudad de México', () => {
    expect(horaVenta(venta)).toBe('13:15');
    expect(fechaHoraCorta(venta)).toBe('23 oct · 13:15');
    expect(fechaHoraLarga(venta)).toBe('Viernes 23 de octubre de 2026, 13:15');
  });

  it('describe el periodo elegido', () => {
    expect(etiquetaPeriodo('hoy', '2026-10-23T06:00:00Z', '2026-10-24T06:00:00Z')).toBe(
      'Viernes 23 de octubre',
    );
    expect(etiquetaPeriodo('semana', '2026-10-19T06:00:00Z', '2026-10-26T06:00:00Z')).toBe(
      'Del 19 al 25 de octubre',
    );
    expect(etiquetaPeriodo('semana', '2026-10-26T06:00:00Z', '2026-11-02T06:00:00Z')).toBe(
      'Del 26 de octubre al 1 de noviembre',
    );
    expect(etiquetaPeriodo('mes', '2026-10-01T06:00:00Z', '2026-11-01T06:00:00Z')).toBe(
      'Octubre de 2026',
    );
    expect(etiquetaPeriodo('hoy')).toBe('');
  });
});
