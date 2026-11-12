import { calcularImpuestos, desglosar } from './impuestos';

describe('calcularImpuestos', () => {
  it('cada producto lleva su IEPS e IVA', () => {
    const [gomilocas, leche] = calcularImpuestos([
      { subtotal: 240, piezas: 12, tasaIeps: 8, tasaIva: 16 },
      { subtotal: 264, piezas: 12, tasaIeps: 0, tasaIva: 0 },
    ]);
    expect(gomilocas.ieps).toBe(19.2);
    expect(gomilocas.iva).toBe(41.47);
    expect(gomilocas.costoPieza).toBe(25.0558);
    expect(leche.costoPieza).toBe(22);
  });

  it('reparte el ajuste del ticket solo entre productos con IEPS y cuadra', () => {
    const filas = calcularImpuestos(
      [
        { subtotal: 100, piezas: 1, tasaIeps: 8, tasaIva: 16 },
        { subtotal: 500, piezas: 1, tasaIeps: 0, tasaIva: 16 },
        { subtotal: 200, piezas: 1, tasaIeps: 8, tasaIva: 0 },
      ],
      25.01,
    );
    expect(filas.map((f) => f.ieps)).toEqual([8.34, 0, 16.67]);
    expect(desglosar(filas, 'ieps')).toEqual([{ tasa: 8, base: 300, productos: 2, monto: 25.01 }]);
    expect(desglosar(filas, 'iva').map((d) => d.tasa)).toEqual([16, 0]);
  });
});
