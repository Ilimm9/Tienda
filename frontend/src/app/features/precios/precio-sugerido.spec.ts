import { margenDePrecio, precioSugerido } from './precio-sugerido';

describe('precioSugerido', () => {
  it('redondea hacia arriba al peso sin error de punto flotante', () => {
    expect(precioSugerido(10, 10)).toBe(11);
    expect(precioSugerido(25.0558, 30)).toBe(33);
    expect(precioSugerido(20, 30)).toBe(26);
  });

  it('recalcula el margen de un precio manual', () => {
    expect(margenDePrecio(20, 25)).toBe(25);
    expect(margenDePrecio(0, 25)).toBe(0);
  });
});
