import { Categoria } from './catalogo.models';
import { normalizar, revisarLista } from './lista-pegada';

const ruta = (lista: Categoria[]) => (categoria: Categoria) => {
  const partes = [categoria.nombre];
  let padre = categoria.categoria_padre_id;
  while (padre) {
    const encontrada = lista.find((item) => item.id === padre);
    if (!encontrada) break;
    partes.unshift(encontrada.nombre);
    padre = encontrada.categoria_padre_id;
  }
  return partes;
};

describe('revisarLista', () => {
  it('normaliza acentos, mayúsculas y espacios', () => {
    expect(normalizar('  Peñafiel   Mineral ')).toBe('penafiel mineral');
  });

  it('marca marcas nuevas, existentes y repetidas', () => {
    const renglones = revisarLista('marcas', 'Coca-Cola\nPepsi\n\nbimbo\nPepsi', [
      { id: '1', nombre: 'Coca-Cola', activo: true },
      { id: '2', nombre: 'Bimbo', activo: true },
    ], ruta([]));

    expect(renglones.map((r) => [r.nombre, r.estado])).toEqual([
      ['Coca-Cola', 'existe'],
      ['Pepsi', 'nuevo'],
      ['bimbo', 'existe'],
      ['Pepsi', 'repetido'],
    ]);
  });

  it('entiende rutas de categorías y avisa los padres que se crearán', () => {
    const existentes: Categoria[] = [
      { id: 'b', nombre: 'Bebidas', activo: true, categoria_padre_id: null },
      { id: 'r', nombre: 'Refrescos', activo: true, categoria_padre_id: 'b' },
    ];
    const renglones = revisarLista('categorias', 'Bebidas > Refrescos\nBebidas > Aguas\nHigiene > Dental\nHigiene', existentes, ruta(existentes));

    expect(renglones[0].estado).toBe('existe');
    expect(renglones[1]).toMatchObject({ nombre: 'Aguas', estado: 'nuevo', padres: ['Bebidas'] });
    expect(renglones[2].detalle).toContain('también crea «Higiene»');
    expect(renglones[3].estado).toBe('repetido');
  });

  it('arma la ruta completa para la lista y respeta el nombre único de categoría', () => {
    const existentes: Categoria[] = [
      { id: 'b', nombre: 'Bebidas', activo: true, categoria_padre_id: null },
      { id: 'f', nombre: 'Frías', activo: true, categoria_padre_id: 'b' },
      { id: 'l', nombre: 'Lácteos', activo: true, categoria_padre_id: null },
    ];
    const renglones = revisarLista(
      'categorias',
      [
        'Lácteos > Frías',            // Frías ya existe (en Bebidas)
        'Lacteos > Frias > Leche',    // Frías está en otro lugar
        'Frías > Hielo',              // el primer nombre puede estar en cualquier nivel
        'Comida > Postres',
        'Dulces > Pasteles',          // Dulces se crea como principal
        'Botanas > Dulces > Chicles', // Dulces va en otro lugar en un renglón anterior
        'Uno > Dos > uno',
        'Postres',                    // ya se crea arriba
      ].join('\n'),
      existentes,
      ruta(existentes),
    );

    expect(renglones.map((r) => r.estado)).toEqual(['existe', 'error', 'nuevo', 'nuevo', 'nuevo', 'error', 'error', 'repetido']);
    expect(renglones[0].detalle).toBe('Se omite: ya está como «Bebidas › Frías»');
    expect(renglones[1].detalle).toBe('«Frías» ya existe como «Bebidas › Frías»');
    expect(renglones[2].payload).toEqual({ ruta: ['Frías', 'Hielo'] });
    expect(renglones[5].detalle).toContain('otro lugar');
    expect(renglones[6].detalle).toContain('dos veces');
  });

  it('compara unidades por nombre o símbolo', () => {
    const renglones = revisarLista(
      'unidades',
      'Kilo, KG, peso, 1000\nGramo, g, peso, 1\nGramito, G, peso, 1\ngramo, gr, peso, 1\nMetro, m, Área, 1',
      [{ id: 'k', nombre: 'Kilogramo', simbolo: 'kg', tipo: 'PESO', factor_a_base: 1000, permite_fraccion: true, decimales: 3, activo: true }],
      ruta([]),
    );

    expect(renglones.map((r) => r.estado)).toEqual(['existe', 'nuevo', 'repetido', 'repetido', 'nuevo']);
    expect(renglones[0].detalle).toBe('Se omite: ya está como «Kilogramo (kg)»');
    expect(renglones[4].payload).toMatchObject({ tipo: 'AREA', decimales: 2 });
  });

  it('respeta los límites de longitud del backend', () => {
    const [marca, simbolo] = [
      revisarLista('marcas', 'x'.repeat(181), [], ruta([]))[0],
      revisarLista('unidades', `Unidad, ${'s'.repeat(21)}, peso, 1`, [], ruta([]))[0],
    ];
    expect(marca.estado).toBe('error');
    expect(simbolo.detalle).toContain('símbolo');
  });

  it('valida unidades y arma su cuerpo', () => {
    const [caja, malo] = revisarLista('unidades', 'Caja, cja, cantidad, 12\nPaquete, paq, piezas, 6', [], ruta([]));

    expect(caja.payload).toEqual({ nombre: 'Caja', simbolo: 'cja', tipo: 'CANTIDAD', factor_a_base: 12, decimales: 0 });
    expect(malo.estado).toBe('error');
  });

  it('valida RFC y correo de proveedores', () => {
    const [bien, rfc, correo] = revisarLista(
      'proveedores',
      'Sabritas, PSM920311K52, 55 1234 5678, pedidos@pepsico.mx\nLala, XYZ, ,\nDel Valle, , , ventas@valle',
      [],
      ruta([]),
    );

    expect(bien.payload).toEqual({ nombre: 'Sabritas', rfc: 'PSM920311K52', telefono: '55 1234 5678', email: 'pedidos@pepsico.mx' });
    expect(rfc.estado).toBe('error');
    expect(correo.detalle).toContain('Correo inválido');
  });
});
