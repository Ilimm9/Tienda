import { CatalogRecord, Categoria } from './catalogo.models';

export type SeccionCatalogo = 'marcas' | 'categorias' | 'proveedores' | 'unidades';
export type EstadoRenglon = 'nuevo' | 'existe' | 'repetido' | 'error';

export interface RenglonLista {
  numero: number;
  /** Texto original del renglón, para devolverlo a la lista si falla. */
  linea: string;
  nombre: string;
  detalle: string;
  estado: EstadoRenglon;
  /** Cuerpo listo para `crear`; en categorías el padre se resuelve al guardar. */
  payload: Record<string, unknown>;
  /** Solo categorías: nombres de los ancestros, del más alto al inmediato. */
  padres?: string[];
}

/** Comparación de nombres sin mayúsculas, acentos ni espacios dobles. */
export function normalizar(valor: string | null | undefined): string {
  return (valor ?? '')
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLocaleLowerCase('es')
    .replace(/\s+/g, ' ')
    .trim();
}

const TIPOS_UNIDAD: Record<string, string> = {
  peso: 'PESO',
  volumen: 'VOLUMEN',
  longitud: 'LONGITUD',
  area: 'AREA',
  cantidad: 'CANTIDAD',
  empaque: 'EMPAQUE',
};
const ETIQUETA_TIPO: Record<string, string> = {
  PESO: 'Peso',
  VOLUMEN: 'Volumen',
  LONGITUD: 'Longitud',
  AREA: 'Área',
  CANTIDAD: 'Cantidad',
  EMPAQUE: 'Empaque',
};
const RFC = /^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{3}$/;
const CORREO = /^[^@\s]+@[^@\s]+\.[^@\s]+$/;

/**
 * Revisa una lista pegada (un registro por renglón) contra lo que ya existe.
 * Marca nuevos, existentes, repetidos dentro de la misma lista y renglones con error.
 */
export function revisarLista(
  seccion: SeccionCatalogo,
  texto: string,
  existentes: CatalogRecord[],
  rutaCategoria: (categoria: Categoria) => string[],
): RenglonLista[] {
  const conocidos = new Map<string, string>();
  for (const item of existentes) {
    const clave =
      seccion === 'categorias'
        ? rutaCategoria(item as Categoria).map(normalizar).join('>')
        : normalizar(item.nombre);
    conocidos.set(clave, seccion === 'categorias' ? rutaCategoria(item as Categoria).join(' › ') : item.nombre);
  }
  const vistos = new Set<string>();

  return texto
    .split(/\r?\n/)
    .map((linea) => linea.trim())
    .filter(Boolean)
    .map((linea, indice) => {
      const renglon = interpretar(seccion, linea, indice + 1);
      if (renglon.estado === 'error') return renglon;
      const clave =
        seccion === 'categorias'
          ? [...(renglon.padres ?? []), renglon.nombre].map(normalizar).join('>')
          : normalizar(renglon.nombre);
      const existente = conocidos.get(clave);
      if (existente) {
        return { ...renglon, estado: 'existe' as const, detalle: `Se omite: ya está como «${existente}»` };
      }
      if (vistos.has(clave)) {
        return { ...renglon, estado: 'repetido' as const, detalle: 'Se omite: aparece antes en la lista' };
      }
      vistos.add(clave);
      // Un ancestro que todavía no existe se crea al guardar; se avisa en el detalle.
      if (seccion === 'categorias' && renglon.padres?.length) {
        const faltantes = renglon.padres.filter((_, i) => {
          const claveAncestro = renglon.padres!.slice(0, i + 1).map(normalizar).join('>');
          return !conocidos.has(claveAncestro) && !vistos.has(claveAncestro);
        });
        renglon.padres.forEach((_, i) => vistos.add(renglon.padres!.slice(0, i + 1).map(normalizar).join('>')));
        if (faltantes.length) renglon.detalle += ` · también crea «${faltantes.join('», «')}»`;
      }
      return renglon;
    });
}

function interpretar(seccion: SeccionCatalogo, linea: string, numero: number): RenglonLista {
  const base = { numero, linea, estado: 'nuevo' as EstadoRenglon };
  const error = (nombre: string, detalle: string): RenglonLista => ({
    ...base,
    nombre: nombre || '(sin nombre)',
    detalle,
    estado: 'error',
    payload: {},
  });

  if (seccion === 'categorias') {
    const partes = linea.split('>').map((parte) => parte.trim());
    const nombre = partes[partes.length - 1];
    if (partes.some((parte) => !parte)) return error(nombre, 'Falta un nombre antes o después de «>»');
    const padres = partes.slice(0, -1);
    return {
      ...base,
      nombre,
      padres,
      detalle: padres.length ? `Dentro de ${padres.join(' › ')}` : 'Categoría principal',
      payload: { nombre },
    };
  }

  const columnas = linea.split(',').map((columna) => columna.trim());
  const nombre = columnas[0] ?? '';
  if (!nombre) return error('', 'Falta el nombre');
  if (nombre.length > 180) return error(nombre.slice(0, 40), 'El nombre pasa de 180 caracteres');

  if (seccion === 'unidades') {
    const [, simbolo = '', tipoTexto = '', factorTexto = '', decimalesTexto = ''] = columnas;
    const tipo = TIPOS_UNIDAD[normalizar(tipoTexto)];
    const factor = Number(factorTexto.replace(/\s/g, ''));
    const decimales = decimalesTexto ? Number(decimalesTexto) : tipo === 'CANTIDAD' || tipo === 'EMPAQUE' ? 0 : 2;
    if (!simbolo) return error(nombre, 'Falta el símbolo');
    if (!tipo) return error(nombre, 'Tipo: peso, volumen, longitud, área, cantidad o empaque');
    if (!(factor > 0)) return error(nombre, 'La equivalencia debe ser un número mayor a 0');
    if (!Number.isInteger(decimales) || decimales < 0 || decimales > 6) return error(nombre, 'Decimales: un entero de 0 a 6');
    return {
      ...base,
      nombre,
      detalle: `${simbolo} · ${ETIQUETA_TIPO[tipo]} · equivale a ${factor} de la base`,
      payload: { nombre, simbolo, tipo, factor_a_base: factor, decimales },
    };
  }

  if (seccion === 'proveedores') {
    const [, rfcTexto = '', telefono = '', correo = ''] = columnas;
    const rfc = rfcTexto.toUpperCase();
    if (rfc && !RFC.test(rfc)) return error(nombre, `RFC con formato inválido: ${rfcTexto}`);
    if (correo && !CORREO.test(correo)) return error(nombre, `Correo inválido: ${correo}`);
    return {
      ...base,
      nombre,
      detalle: [rfc, telefono, correo].filter(Boolean).join(' · ') || 'Sin datos de contacto',
      payload: { nombre, rfc: rfc || null, telefono: telefono || null, email: correo || null },
    };
  }

  return { ...base, nombre: linea, detalle: '', payload: { nombre: linea } };
}
