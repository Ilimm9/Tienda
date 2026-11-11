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
