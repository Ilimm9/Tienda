const CONECTORES = new Set(['sucursal', 'de', 'del', 'la', 'las', 'los', 'el', 'y']);

/** «Sucursal Centro» → `SC`: una S más las iniciales (hasta 3) de las palabras que la distinguen. */
export function prefijoSucursal(nombre: string): string {
  const palabras = nombre
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toUpperCase()
    .split(/[^A-Z0-9]+/)
    .filter((palabra) => palabra && !CONECTORES.has(palabra.toLowerCase()));
  return palabras.length ? 'S' + palabras.slice(0, 3).map((palabra) => palabra[0]).join('') : 'SUC';
}

/** Siguiente `PREFIJO-NNN` libre según los números ya usados en el negocio. */
export function siguienteNumero(nombreSucursal: string, existentes: readonly (string | null)[]): string {
  const prefijo = prefijoSucursal(nombreSucursal);
  const patron = new RegExp(`^${prefijo}-(\\d+)$`, 'i');
  const mayor = existentes.reduce((max, numero) => {
    const coincidencia = patron.exec(numero ?? '');
    return coincidencia ? Math.max(max, Number(coincidencia[1])) : max;
  }, 0);
  return `${prefijo}-${String(mayor + 1).padStart(3, '0')}`;
}

/** Deja solo los 10 dígitos del teléfono; quita la lada 52 de valores anteriores con 12 dígitos. */
export function digitosTelefono(valor: string | null | undefined): string {
  const digitos = (valor ?? '').replace(/\D/g, '');
  return (digitos.length === 12 && digitos.startsWith('52') ? digitos.slice(2) : digitos).slice(0, 10);
}

/** `5512345678` → `55 1234 5678`, también mientras se escribe. */
export function formatoTelefono(valor: string | null | undefined): string {
  const digitos = digitosTelefono(valor);
  return [digitos.slice(0, 2), digitos.slice(2, 6), digitos.slice(6)].filter(Boolean).join(' ');
}
