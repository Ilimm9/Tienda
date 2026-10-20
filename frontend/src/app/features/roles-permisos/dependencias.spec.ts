import { completarDependencias, Permiso } from './rol.models';

describe('dependencias de permisos', () => {
  const catalogo: Permiso[] = [
    { id: 'autorizar', codigo: 'precios.autorizar', codigo_modulo: 'precios', nombre: 'Autorizar', requiere: ['precios.ver'] },
    { id: 'ver', codigo: 'precios.ver', codigo_modulo: 'precios', nombre: 'Ver', requiere: [] },
    { id: 'asignar', codigo: 'equipo.asignaciones.editar', codigo_modulo: 'equipo', nombre: 'Asignar', requiere: ['equipo.asignaciones.ver'] },
    { id: 'asignaciones', codigo: 'equipo.asignaciones.ver', codigo_modulo: 'equipo', nombre: 'Ver asignaciones', requiere: ['equipo.empleados.ver'] },
    { id: 'empleados', codigo: 'equipo.empleados.ver', codigo_modulo: 'equipo', nombre: 'Ver empleados', requiere: [] },
  ];

  it('agrega lectura y dependencias transitivas sin modificar selección original', () => {
    const seleccion = new Set(['autorizar', 'asignar']);
    expect([...completarDependencias(seleccion, catalogo)].sort()).toEqual(['asignaciones', 'asignar', 'autorizar', 'empleados', 'ver']);
    expect([...seleccion]).toEqual(['autorizar', 'asignar']);
  });

  it('no agrega acciones al seleccionar lectura ni duplica dependencias', () => {
    expect([...completarDependencias(new Set(['ver']), catalogo)]).toEqual(['ver']);
    expect(completarDependencias(new Set(['autorizar', 'ver']), catalogo).size).toBe(2);
  });
});
