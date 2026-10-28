import {
  aCentavos,
  desgloseIVA,
  esCodigoEscaneado,
  margen,
  nivelMargen,
  resumir,
} from './venta-calculos';
import { LineaVenta } from './ventas.models';

const linea = (precioCentavos: number, cantidad: number): LineaVenta => ({
  productoNegocioId: `p-${precioCentavos}`,
  nombre: 'Producto',
  codigoBarras: null,
  precioCentavos,
  costoCentavos: null,
  existencia: 99,
  existenciaBaja: false,
  cantidad,
});

describe('cálculos de venta', () => {
  it('separa el IVA incluido igual que el backend', () => {
    expect(desgloseIVA(6400)).toEqual({ subtotal: 5517, iva: 883 });
    expect(desgloseIVA(1160)).toEqual({ subtotal: 1000, iva: 160 });
    expect(desgloseIVA(0)).toEqual({ subtotal: 0, iva: 0 });
  });

  it('resume productos, piezas y total en centavos', () => {
    expect(resumir([linea(1800, 2), linea(2800, 1)])).toEqual({
      productos: 2,
      piezas: 3,
      total: 6400,
      subtotal: 5517,
      iva: 883,
    });
    expect(aCentavos(0.07)).toBe(7);
    expect(aCentavos(18.5)).toBe(1850);
  });

  it('clasifica el margen para «Verificar costos»', () => {
    expect(margen(2000, 1500)).toBe(0.25);
    expect(margen(2000, null)).toBeNull();
    expect(nivelMargen(1000, 1200)).toBe('negativo');
    expect(nivelMargen(1000, 900)).toBe('bajo');
    expect(nivelMargen(1000, 800)).toBe('ok');
    expect(nivelMargen(1000, null)).toBeNull();
  });

  it('reconoce lecturas del lector', () => {
    expect(esCodigoEscaneado('7501055303038')).toBe(true);
    expect(esCodigoEscaneado('12345')).toBe(false);
    expect(esCodigoEscaneado('coca 600')).toBe(false);
  });
});
