import { digitosTelefono, formatoTelefono, prefijoSucursal, siguienteNumero } from './numero-empleado';

describe('numero-empleado', () => {
  it('arma el prefijo con las iniciales que distinguen a la sucursal', () => {
    expect(prefijoSucursal('Sucursal Centro')).toBe('SC');
    expect(prefijoSucursal('Sucursal Norte')).toBe('SN');
    expect(prefijoSucursal('Matriz de Querétaro')).toBe('SMQ');
    expect(prefijoSucursal('Sucursal')).toBe('SUC');
  });

  it('propone el consecutivo siguiente sin importar huecos ni mayúsculas', () => {
    expect(siguienteNumero('Sucursal Centro', [])).toBe('SC-001');
    expect(siguienteNumero('Sucursal Centro', ['SC-001', 'sc-007', 'SN-020', null, 'SC-ABC'])).toBe('SC-008');
    expect(siguienteNumero('Sucursal Centro', ['SC-999'])).toBe('SC-1000');
  });

  it('deja el teléfono en 10 dígitos sin lada', () => {
    expect(digitosTelefono('+52 55 1234 5678')).toBe('5512345678');
    expect(digitosTelefono('55-1234-5678-99')).toBe('5512345678');
    expect(formatoTelefono('551234')).toBe('55 1234');
    expect(formatoTelefono('5512345678')).toBe('55 1234 5678');
    expect(formatoTelefono('')).toBe('');
  });
});
