/** Redondeo a 4 decimales para quitar el error de punto flotante antes de subir al peso. */
export function round4(value: number): number {
  return Math.round(value * 10000) / 10000;
}

/** Misma regla que el backend: costo con margen, redondeado hacia arriba al peso. */
export function precioSugerido(costo: number, margenPorcentaje: number): number {
  return Math.ceil(round4(costo * (1 + margenPorcentaje / 100)));
}

/** Margen (en %) que deja un precio escrito a mano. */
export function margenDePrecio(costo: number, precio: number): number {
  return costo > 0 ? Math.round(((precio - costo) / costo) * 10000) / 100 : 0;
}
