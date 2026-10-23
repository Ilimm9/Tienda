import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

import { ContextoService } from '../../contexto/contexto.service';
import { AgregarCatalogoComponent } from './agregar-catalogo.component';
import { ProductosService } from './productos.service';

describe('AgregarCatalogoComponent', () => {
  const productos = [
    { id: 'p-1', nombre: 'Jabón', presentacion: null, sku: 'S1', codigo_barras: '750', categoria: 'Limpieza', variantes_faltantes: 0 },
    { id: 'p-2', nombre: 'Playera', presentacion: null, sku: null, codigo_barras: null, categoria: 'Ropa', variantes_faltantes: 2 },
    { id: 'p-3', nombre: 'Cloro', presentacion: '1 L', sku: null, codigo_barras: null, categoria: 'Limpieza', variantes_faltantes: 0 },
  ];
  const service = {
    listNotInBranch: vi.fn(() => of(productos)),
    addToBranch: vi.fn(() => of({ agregados: 2 })),
  };
  const router = { navigate: vi.fn().mockResolvedValue(true) };
  let fixture: ComponentFixture<AgregarCatalogoComponent>;
  let component: AgregarCatalogoComponent;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(Swal, 'fire').mockResolvedValue({ isConfirmed: true } as never);
    TestBed.configureTestingModule({
      imports: [AgregarCatalogoComponent],
      providers: [
        { provide: ProductosService, useValue: service },
        { provide: ContextoService, useValue: { negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-2', nombre: 'SUR' }) } },
        { provide: Router, useValue: router },
      ],
    });
    fixture = TestBed.createComponent(AgregarCatalogoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  afterEach(() => vi.restoreAllMocks());

  it('carga los productos que no están en la sucursal activa', () => {
    expect(service.listNotInBranch).toHaveBeenCalledWith('negocio-1', 'sucursal-2');
    expect(fixture.nativeElement.querySelectorAll('tbody tr')).toHaveLength(3);
    expect(fixture.nativeElement.textContent).toContain('2 variantes');
  });

  it('filtra sin acentos y elige solo los visibles', () => {
    component.busqueda.set('limpieza');
    component.alternarVisibles();
    expect([...component.seleccion()].sort()).toEqual(['p-1', 'p-3']);

    component.busqueda.set('jabon');
    expect(component.visibles().map((item) => item.id)).toEqual(['p-1']);
  });

  it('agrega los elegidos y vuelve a la lista', () => {
    component.alternar('p-1');
    component.alternar('p-2');
    component.agregarElegidos();

    expect(service.addToBranch).toHaveBeenCalledWith('negocio-1', 'sucursal-2', ['p-1', 'p-2']);
    expect(router.navigate).toHaveBeenCalledWith(['/catalogo/productos']);
  });

  it('agrega todo el catálogo tras confirmar', async () => {
    component.agregarTodos();
    await Promise.resolve();
    await Promise.resolve();

    expect(service.addToBranch).toHaveBeenCalledWith('negocio-1', 'sucursal-2', 'todos');
  });

  it('no envía nada sin productos elegidos', () => {
    component.agregarElegidos();
    expect(service.addToBranch).not.toHaveBeenCalled();
  });
});
