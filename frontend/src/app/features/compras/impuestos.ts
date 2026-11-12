/** Tasas que ofrece el formulario, en porcentaje. Las mismas que siembra el backend. */
export const TASAS_IEPS = [8, 25, 26.5, 30, 53, 160];
export const TASAS_IVA = [16, 0];

export interface RenglonImpuesto {
  subtotal: number;
  piezas: number;
  /** Porcentaje; 0 = sin IEPS. */
  tasaIeps: number;
  /** Porcentaje: 16 o 0. */
  tasaIva: number;
}
export interface RenglonCalculado extends RenglonImpuesto {
  ieps: number;
  iva: number;
  /** Costo por pieza buena con impuestos, como `CostoFinal` del backend. */
  costoPieza: number;
}
export interface Desglose {
  tasa: number;
  base: number;
  productos: number;
  monto: number;
}

export const money = (value: number): number => Math.round(value * 100) / 100;
export const money4 = (value: number): number => Math.round(value * 10000) / 10000;

/**
 * Misma regla que `CalcularImpuestosRenglon` del backend: IEPS por renglón, el ajuste del ticket
 * se reparte entre los renglones con IEPS (el último absorbe los centavos) e IVA sobre subtotal + IEPS.
 */
export function calcularImpuestos(renglones: RenglonImpuesto[], iepsAjustado: number | null = null): RenglonCalculado[] {
  const filas = renglones.map((r) => ({ ...r, ieps: money(r.subtotal * (r.tasaIeps / 100)), iva: 0, costoPieza: 0 }));
  const conIeps = filas.filter((r) => r.tasaIeps > 0);
  const base = conIeps.reduce((total, r) => total + r.subtotal, 0);
  if (iepsAjustado !== null && Number.isFinite(iepsAjustado) && iepsAjustado >= 0 && base > 0) {
    const calculado = filas.reduce((total, r) => total + r.ieps, 0);
    const diferencia = money(money(iepsAjustado) - money(calculado));
    let repartido = 0;
    conIeps.forEach((r, i) => {
      const parte = i === conIeps.length - 1 ? money(diferencia - repartido) : money((diferencia * r.subtotal) / base);
      repartido = money(repartido + parte);
      r.ieps = money(r.ieps + parte);
    });
  }
  for (const r of filas) {
    r.iva = money((r.subtotal + r.ieps) * (r.tasaIva / 100));
    r.costoPieza = r.piezas > 0 ? money4((r.subtotal + r.ieps + r.iva) / r.piezas) : 0;
  }
  return filas;
}

/** Agrupa por tasa para el resumen: «8 % sobre $588.00 · 2 productos $47.04». */
export function desglosar(filas: RenglonCalculado[], tipo: 'ieps' | 'iva'): Desglose[] {
  const grupos = new Map<number, Desglose>();
  for (const fila of filas) {
    const tasa = tipo === 'ieps' ? fila.tasaIeps : fila.tasaIva;
    if (tipo === 'ieps' && tasa <= 0) continue;
    const grupo = grupos.get(tasa) ?? { tasa, base: 0, productos: 0, monto: 0 };
    grupo.base = money(grupo.base + (tipo === 'ieps' ? fila.subtotal : fila.subtotal + fila.ieps));
    grupo.productos += 1;
    grupo.monto = money(grupo.monto + (tipo === 'ieps' ? fila.ieps : fila.iva));
    grupos.set(tasa, grupo);
  }
  return [...grupos.values()].sort((a, b) => b.tasa - a.tasa);
}
