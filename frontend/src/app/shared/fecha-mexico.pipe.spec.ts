import '@angular/common/locales/global/es-MX';

import { FechaMexicoPipe } from './fecha-mexico.pipe';

describe('FechaMexicoPipe', () => {
  const pipe = new FechaMexicoPipe();

  it('muestra un instante UTC en la hora de Ciudad de México', () => {
    expect(pipe.transform('2026-09-23T18:30:00Z', 'yyyy-MM-dd HH:mm')).toBe('2026-09-23 12:30');
  });

  it('preserva valores ausentes', () => {
    expect(pipe.transform(null)).toBeNull();
  });
});
