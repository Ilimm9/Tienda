import { CatalogRecord, Categoria, UnidadMedida } from './catalogo.models';

export type SeccionCatalogo = 'marcas' | 'categorias' | 'proveedores' | 'unidades';
export type EstadoRenglon = 'nuevo' | 'existe' | 'repetido' | 'error';

export interface RenglonLista {
  numero: number;
  /** Texto original del renglón, para devolverlo a la lista si falla. */
  linea: string;
  nombre: string;
  detalle: string;
  estado: EstadoRenglon;
  /** Renglón para `POST …/lista` (sin `fila`); en categorías lleva la ruta completa. */
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
/** Mismos límites que las columnas del backend. */
const MAX_NOMBRE = 180;
const MAX_NOMBRE_UNIDAD = 100;
const MAX_SIMBOLO = 20;
const MAX_TELEFONO = 30;
const MAX_CORREO = 255;
const MAX_NIVELES = 10;

/**
 * Revisa una lista pegada (un registro por renglón) contra lo que ya existe,
 * con las mismas reglas que aplica el backend al guardar:
 * - Nombres sin mayúsculas, acentos ni espacios dobles.
 * - Unidades: chocan por nombre o por símbolo.
 * - Categorías: el nombre es único en el negocio, en cualquier nivel.
 */
export function revisarLista(
  seccion: SeccionCatalogo,
  texto: string,
  existentes: CatalogRecord[],
  rutaCategoria: (categoria: Categoria) => string[],
): RenglonLista[] {
  const lineas = texto
    .split(/\r?\n/)
    .map((linea) => linea.trim())
    .filter(Boolean);
  if (seccion === 'categorias') return revisarCategorias(lineas, existentes as Categoria[], rutaCategoria);

  const conocidos = new Map<string, string>();
  for (const item of existentes) {
    conocidos.set(normalizar(item.nombre), item.nombre);
    if (seccion === 'unidades') {
      const unidad = item as UnidadMedida;
      if (!conocidos.has(normalizar(unidad.simbolo))) conocidos.set(normalizar(unidad.simbolo), `${unidad.nombre} (${unidad.simbolo})`);
    }
  }
  const vistos = new Set<string>();

  return lineas.map((linea, indice) => {
    const renglon = interpretar(seccion, linea, indice + 1);
    if (renglon.estado === 'error') return renglon;
    const claves =
      seccion === 'unidades'
        ? [normalizar(renglon.nombre), normalizar(String(renglon.payload['simbolo']))]
        : [normalizar(renglon.nombre)];
    const existente = claves.map((clave) => conocidos.get(clave)).find(Boolean);
    if (existente) {
      return { ...renglon, estado: 'existe' as const, detalle: `Se omite: ya está como «${existente}»` };
    }
    if (claves.some((clave) => vistos.has(clave))) {
      return { ...renglon, estado: 'repetido' as const, detalle: 'Se omite: aparece antes en la lista' };
    }
    claves.forEach((clave) => vistos.add(clave));
    return renglon;
  });
}

function revisarCategorias(
  lineas: string[],
  existentes: Categoria[],
  rutaCategoria: (categoria: Categoria) => string[],
): RenglonLista[] {
  const porNombre = new Map<string, Categoria>();
  for (const categoria of existentes) {
    if (!porNombre.has(normalizar(categoria.nombre))) porNombre.set(normalizar(categoria.nombre), categoria);
  }
  const porId = new Map(existentes.map((categoria) => [categoria.id, categoria]));
  const ruta = (categoria: Categoria) => rutaCategoria(categoria).join(' › ');
  /** Nombres que la lista va a crear → clave de su padre (null = principal). */
  const nuevas = new Map<string, string | null>();

  return lineas.map((linea, indice) => {
    const renglon = interpretar('categorias', linea, indice + 1);
    if (renglon.estado === 'error') return renglon;
    const segmentos = [...(renglon.padres ?? []), renglon.nombre];
    const claves = segmentos.map(normalizar);
    const error = (detalle: string): RenglonLista => ({ ...renglon, estado: 'error', detalle });

    const repetido = segmentos.find((_, i) => claves.indexOf(claves[i]) !== i);
    if (repetido) return error(`«${repetido}» aparece dos veces en la ruta`);

    // Cada nombre intermedio que ya existe debe estar dentro del anterior.
    for (let i = 1; i < claves.length - 1; i++) {
      const existente = porNombre.get(claves[i]);
      if (existente) {
        const padre = existente.categoria_padre_id ? porId.get(existente.categoria_padre_id) : undefined;
        if (!padre || normalizar(padre.nombre) !== claves[i - 1]) {
          return error(`«${existente.nombre}» ya existe como «${ruta(existente)}»`);
        }
      } else if (nuevas.has(claves[i]) && nuevas.get(claves[i]) !== claves[i - 1]) {
        return error(`«${segmentos[i]}» ya va en otro lugar en un renglón anterior`);
      }
    }

    const ultima = claves[claves.length - 1];
    const existente = porNombre.get(ultima);
    if (existente) {
      return { ...renglon, estado: 'existe' as const, detalle: `Se omite: ya está como «${ruta(existente)}»` };
    }
    if (nuevas.has(ultima)) {
      return { ...renglon, estado: 'repetido' as const, detalle: 'Se omite: aparece antes en la lista' };
    }

    const faltantes = (renglon.padres ?? []).filter((_, i) => !porNombre.has(claves[i]) && !nuevas.has(claves[i]));
    claves.forEach((clave, i) => {
      if (!porNombre.has(clave) && !nuevas.has(clave)) nuevas.set(clave, i === 0 ? null : claves[i - 1]);
    });
    if (faltantes.length) renglon.detalle += ` · también crea «${faltantes.join('», «')}»`;
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
    if (partes.length > MAX_NIVELES) return error(nombre, `Máximo ${MAX_NIVELES} niveles`);
    if (partes.some((parte) => parte.length > MAX_NOMBRE)) return error(nombre.slice(0, 40), `Cada nombre: máximo ${MAX_NOMBRE} caracteres`);
    const padres = partes.slice(0, -1);
    return {
      ...base,
      nombre,
      padres,
      detalle: padres.length ? `Dentro de ${padres.join(' › ')}` : 'Categoría principal',
      payload: { ruta: partes },
    };
  }

  if (seccion === 'marcas') {
    if (linea.length > MAX_NOMBRE) return error(linea.slice(0, 40), `El nombre pasa de ${MAX_NOMBRE} caracteres`);
    return { ...base, nombre: linea, detalle: '', payload: { nombre: linea } };
  }

  const columnas = linea.split(',').map((columna) => columna.trim());
  const nombre = columnas[0] ?? '';
  if (!nombre) return error('', 'Falta el nombre');

  if (seccion === 'unidades') {
    const [, simbolo = '', tipoTexto = '', factorTexto = '', decimalesTexto = ''] = columnas;
    const tipo = TIPOS_UNIDAD[normalizar(tipoTexto)];
    const factor = Number(factorTexto.replace(/\s/g, ''));
    const decimales = decimalesTexto ? Number(decimalesTexto) : tipo === 'CANTIDAD' || tipo === 'EMPAQUE' ? 0 : 2;
    if (nombre.length > MAX_NOMBRE_UNIDAD) return error(nombre.slice(0, 40), `El nombre pasa de ${MAX_NOMBRE_UNIDAD} caracteres`);
    if (!simbolo) return error(nombre, 'Falta el símbolo');
    if (simbolo.length > MAX_SIMBOLO) return error(nombre, `El símbolo pasa de ${MAX_SIMBOLO} caracteres`);
    if (!tipo) return error(nombre, 'Tipo: peso, volumen, longitud, área, cantidad o empaque');
    if (!factorTexto || !(factor > 0)) return error(nombre, 'La equivalencia debe ser un número mayor a 0');
    if (!Number.isInteger(decimales) || decimales < 0 || decimales > 6) return error(nombre, 'Decimales: un entero de 0 a 6');
    return {
      ...base,
      nombre,
      detalle: `${simbolo} · ${ETIQUETA_TIPO[tipo]} · equivale a ${factor} de la base`,
      payload: { nombre, simbolo, tipo, factor_a_base: factor, decimales },
    };
  }

  // Proveedores
  const [, rfcTexto = '', telefono = '', correo = ''] = columnas;
  const rfc = rfcTexto.toUpperCase();
  if (nombre.length > MAX_NOMBRE) return error(nombre.slice(0, 40), `El nombre pasa de ${MAX_NOMBRE} caracteres`);
  if (rfc && !RFC.test(rfc)) return error(nombre, `RFC con formato inválido: ${rfcTexto}`);
  if (telefono.length > MAX_TELEFONO) return error(nombre, `El teléfono pasa de ${MAX_TELEFONO} caracteres`);
  if (correo && (!CORREO.test(correo) || correo.length > MAX_CORREO)) return error(nombre, `Correo inválido: ${correo}`);
  return {
    ...base,
    nombre,
    detalle: [rfc, telefono, correo].filter(Boolean).join(' · ') || 'Sin datos de contacto',
    payload: { nombre, rfc, telefono, email: correo },
  };
}
