import { LineaVenta, NivelMargen, ProductoVendible } from './ventas.models';

/** Convierte pesos de la API (máximo dos decimales) a centavos enteros. */
export function aCentavos(pesos: number): number {
  return Math.round(pesos * 100);
}

export function aPesos(centavos: number): number {
  return centavos / 100;
}

/** Misma regla que el backend: subtotal = round(total ÷ 1.16) con IVA incluido. */
export function desgloseIVA(totalCentavos: number): { subtotal: number; iva: number } {
  const subtotal = Math.floor((totalCentavos * 100 + 58) / 116);
  return { subtotal, iva: totalCentavos - subtotal };
}

export interface ResumenVenta {
  productos: number;
  piezas: number;
  subtotal: number;
  iva: number;
  total: number;
}

export function resumir(lineas: readonly LineaVenta[]): ResumenVenta {
  const total = lineas.reduce((suma, linea) => suma + linea.precioCentavos * linea.cantidad, 0);
  const piezas = lineas.reduce((suma, linea) => suma + linea.cantidad, 0);
  return { productos: lineas.length, piezas, total, ...desgloseIVA(total) };
}

/** Margen (precio − costo) ÷ precio; null sin costo o sin precio. */
export function margen(precioCentavos: number, costoCentavos: number | null): number | null {
  if (costoCentavos === null || precioCentavos <= 0) return null;
  return (precioCentavos - costoCentavos) / precioCentavos;
}

export function nivelMargen(
  precioCentavos: number,
  costoCentavos: number | null,
): NivelMargen | null {
  if (costoCentavos === null || precioCentavos <= 0) return null;
  if (costoCentavos > precioCentavos) return 'negativo';
  if ((precioCentavos - costoCentavos) * 100 < precioCentavos * 20) return 'bajo';
  return 'ok';
}

export function lineaDesdeProducto(producto: ProductoVendible): LineaVenta {
  return {
    productoNegocioId: producto.producto_negocio_id,
    nombre: producto.nombre,
    codigoBarras: producto.codigo_barras,
    precioCentavos: aCentavos(producto.precio),
    costoCentavos: producto.costo === undefined ? null : aCentavos(producto.costo),
    existencia: producto.existencia,
    existenciaBaja: producto.existencia_baja,
    cantidad: 1,
  };
}

/** Un texto de solo dígitos con 6 o más se trata como lectura del lector de códigos. */
export function esCodigoEscaneado(texto: string): boolean {
  return /^\d{6,}$/.test(texto);
}
