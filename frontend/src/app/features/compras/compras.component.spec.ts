import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { ComprasComponent } from './compras.component';
import { ComprasService } from './compras.service';

describe('ComprasComponent', () => {
  const service = {
    providers: vi.fn(() => of([])),
    branches: vi.fn(() => of([{ id: 'sucursal-1', nombre: 'Matriz' }])),
    units: vi.fn(() => of([{ id: 'caja-1', nombre: 'Caja', tipo: 'EMPAQUE', factor_a_base: 12 }])),
    list: vi.fn(() => of({ items: [], total: 0 })),
    products: vi.fn(() => of([])),
    create: vi.fn(),
    createUnit: vi.fn(),
  };

  async function createFixture(): Promise<ComponentFixture<ComprasComponent>> {
    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
        { provide: ActivatedRoute, useValue: { snapshot: { data: { mode: 'create' } } } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('acumula el resumen global de todos los productos y muestra cero al iniciar', async () => {
    const component = (await createFixture()).componentInstance;
    expect(component.total()).toBe(0);

    component.details.at(0).patchValue({ cantidad_empaques: 2, costo_empaque: 100, descuento: 10 });
    component.addDetail();
    component.details.at(1).patchValue({ cantidad_empaques: 3, costo_empaque: 50, descuento: 0 });
    component.form.controls.incluye_iva.setValue(true);

    expect(component.subtotal()).toBe(340);
    expect(component.iva()).toBe(54.4);
    expect(component.total()).toBe(394.4);

    component.form.controls.incluye_ieps.setValue(true);
    expect(component.ieps()).toBe(27.2);

    component.form.controls.ieps.setValue(18);
    expect(component.iva()).toBe(57.28);
    expect(component.total()).toBe(415.28);
  });

  it('muestra el panel de nuevo producto con el formato de nueva unidad', async () => {
    const fixture = await createFixture();
    const component = fixture.componentInstance;

    component.selectProduct(0, '__new_product__');
    fixture.detectChanges();

    const productPanel = fixture.nativeElement.querySelector('.purchase-product-form');
    expect(productPanel?.querySelector('h3')?.textContent).toContain('Nuevo producto');
    expect(productPanel?.querySelectorAll('input')).toHaveLength(2);
  });

  it('requiere nombre y código de barras al crear un producto nuevo', async () => {
    const component = (await createFixture()).componentInstance;

    component.selectProduct(0, '__new_product__');

    expect(component.details.at(0).get('nombre_producto_nuevo')?.invalid).toBe(true);
    expect(component.details.at(0).get('codigo_barras_nuevo')?.invalid).toBe(true);

    component.details.at(0).patchValue({
      nombre_producto_nuevo: 'Galletas',
      codigo_barras_nuevo: '7501234567890',
    });

    expect(component.details.at(0).get('nombre_producto_nuevo')?.valid).toBe(true);
    expect(component.details.at(0).get('codigo_barras_nuevo')?.valid).toBe(true);
  });
});
