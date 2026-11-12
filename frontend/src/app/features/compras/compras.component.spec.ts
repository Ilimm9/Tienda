import '@angular/common/locales/global/es-MX';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { ActivatedRoute, convertToParamMap, Router } from '@angular/router';
import { Select } from 'primeng/select';
import { of, Subject } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { ComprasComponent } from './compras.component';
import { ComprasService } from './compras.service';

// jsdom no implementa la ventana modal nativa.
HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
  this.setAttribute('open', '');
};
HTMLDialogElement.prototype.close ??= function (this: HTMLDialogElement) {
  this.removeAttribute('open');
  this.dispatchEvent(new Event('close'));
};

describe('ComprasComponent', () => {
  const service = {
    providers: vi.fn(() => of([])),
    branches: vi.fn(() => of([{ id: 'sucursal-1', nombre: 'Matriz' }])),
    units: vi.fn(() => of([{ id: 'caja-1', nombre: 'Caja', tipo: 'EMPAQUE', factor_a_base: 12 }])),
    list: vi.fn(() => of({ items: [], total: 0 })),
    products: vi.fn(() => of([])),
    create: vi.fn(),
    createUnit: vi.fn(),
    createProvider: vi.fn(() => of(undefined)),
  };

  async function createFixture(): Promise<ComponentFixture<ComprasComponent>> {
    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
        { provide: ActivatedRoute, useValue: { snapshot: { data: { mode: 'create' } } } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('muestra la fecha de recepción como fecha civil en la tabla', async () => {
    service.list.mockReturnValueOnce(of({
      items: [{ id: 'compra-1', fecha_recepcion: '2026-10-21T00:00:00Z', tipo_documento: 'TICKET', folio_documento: 'A1', forma_pago: 'EFECTIVO', total: 100, importe_pagado: 100 }],
      total: 1,
    }) as never);
    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
        { provide: ActivatedRoute, useValue: { snapshot: { data: {} } } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();

    const celda = fixture.nativeElement.querySelector('.compras-tabla td.compras-num') as HTMLElement;
    expect(celda.textContent?.trim()).toBe('21 oct 2026');
  });

  it('abre el detalle de la compra que llega en ?compra=', async () => {
    service.list.mockReturnValueOnce(of({
      items: [{ id: 'compra-9', fecha_recepcion: '2026-10-21T00:00:00Z', tipo_documento: 'TICKET', folio_documento: 'A9', forma_pago: 'EFECTIVO', total: 100, importe_pagado: 100 }],
      total: 1,
    }) as never);
    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [
        { provide: ComprasService, useValue: { ...service, view: vi.fn(() => new Subject()) } },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
        { provide: ActivatedRoute, useValue: { snapshot: { data: {}, queryParamMap: convertToParamMap({ compra: 'compra-9' }) } } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();
    expect(fixture.componentInstance.accion()).toMatchObject({ tipo: 'ver', compra: { id: 'compra-9' } });
  });

  it('lista las compras de la sucursal activa y recarga al cambiarla', async () => {
    const sucursal = signal({ id: 'sucursal-1' });
    await TestBed.configureTestingModule({
      imports: [ComprasComponent],
      providers: [
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'negocio-1' }), sucursal } },
        { provide: ActivatedRoute, useValue: { snapshot: { data: {} } } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasComponent);
    fixture.detectChanges();
    expect(service.list).toHaveBeenLastCalledWith('negocio-1', 'sucursal-1');
    const llamadas = service.list.mock.calls.length;

    sucursal.set({ id: 'sucursal-2' });
    fixture.detectChanges();
    expect(service.list).toHaveBeenLastCalledWith('negocio-1', 'sucursal-2');
    expect(service.list).toHaveBeenCalledTimes(llamadas + 1);
  });

  it('acumula el resumen global de todos los productos y muestra cero al iniciar', async () => {
    const component = (await createFixture()).componentInstance;
    expect(component.total()).toBe(0);

    component.details.at(0).patchValue({ unidad_medida_id: 'caja-1', cantidad_empaques: 2, costo_empaque: 100, descuento: 10, tasa_iva: 0 });
    component.addDetail();
    component.details.at(1).patchValue({ unidad_medida_id: 'caja-1', cantidad_empaques: 3, costo_empaque: 50, descuento: 0, tasa_iva: 0 });

    expect(component.subtotal()).toBe(340);
    expect(component.iva()).toBe(0);
    expect(component.total()).toBe(340);
  });

  it('calcula IEPS e IVA por producto con su desglose por tasa', async () => {
    const component = (await createFixture()).componentInstance;
    component.details.at(0).patchValue({ unidad_medida_id: 'caja-1', cantidad_empaques: 1, costo_empaque: 240, aplica_ieps: true, tasa_ieps: 8, tasa_iva: 16 });
    component.addDetail();
    component.details.at(1).patchValue({ unidad_medida_id: 'caja-1', cantidad_empaques: 1, costo_empaque: 264, aplica_ieps: false, tasa_iva: 0 });

    expect(component.ieps()).toBe(19.2);
    expect(component.iva()).toBe(41.47);
    expect(component.total()).toBe(564.67);
    expect(component.costoPieza(0)).toBe(25.0558);
    expect(component.costoPieza(1)).toBe(22);
    expect(component.desgloseIeps()).toEqual([{ tasa: 8, base: 240, productos: 1, monto: 19.2 }]);
    expect(component.desgloseIva().map((tasa) => tasa.tasa)).toEqual([16, 0]);
  });

  it('ajusta el IEPS del ticket repartiendo la diferencia y lo manda al guardar', async () => {
    service.create.mockReturnValueOnce(of({}) as never);
    const component = (await createFixture()).componentInstance;
    component.form.patchValue({ proveedor_id: 'proveedor-1', folio_documento: 'A-1' });
    component.details.at(0).patchValue({ producto_negocio_id: 'producto-1', unidad_medida_id: 'caja-1', cantidad_empaques: 1, costo_empaque: 240, aplica_ieps: true, tasa_ieps: 8, tasa_iva: 16 });
    component.alternarAjusteIeps();
    expect(component.form.controls.ieps_ajustado.value).toBe(19.2);
    component.form.controls.ieps_ajustado.setValue(20);

    expect(component.ieps()).toBe(20);
    expect(component.diferenciaIeps()).toBe(0.8);

    component.submit();
    const request = service.create.mock.calls.at(-1)?.[1] as unknown as { ieps_ajustado: number; incluye_iva: boolean; detalles: { tasa_ieps: number; tasa_iva: number }[] };
    expect(request.ieps_ajustado).toBe(20);
    expect(request.incluye_iva).toBe(false);
    expect(request.detalles[0]).toMatchObject({ tasa_ieps: 0.08, tasa_iva: 0.16 });
    service.create.mockClear();
  });

  it('llena IEPS e IVA con lo guardado en el producto', async () => {
    const component = (await createFixture()).componentInstance;
    component.products.set([
      { id: 'con', nombre: 'Gomilocas', sku: null, codigo_barras: null, del_proveedor: true, tasa_ieps: 0.265, tasa_iva: 0 },
      { id: 'sin', nombre: 'Arroz', sku: null, codigo_barras: null, del_proveedor: true, tasa_ieps: null, tasa_iva: null },
    ]);
    component.selectProduct(0, 'con');
    expect(component.details.at(0).value).toMatchObject({ aplica_ieps: true, tasa_ieps: 26.5, tasa_iva: 0 });
    component.selectProduct(0, 'sin');
    expect(component.details.at(0).value).toMatchObject({ aplica_ieps: false, tasa_iva: 16 });
  });

  it('contrae los productos anteriores al agregar uno y permite alternarlos', async () => {
    const fixture = await createFixture();
    const component = fixture.componentInstance;
    component.details.at(0).patchValue({ cantidad_empaques: 2, costo_empaque: 100, descuento: 0 });

    component.addDetail();
    fixture.detectChanges();

    expect(component.estaContraido(0)).toBe(true);
    expect(component.estaContraido(1)).toBe(false);
    const toggles = fixture.nativeElement.querySelectorAll('.purchase-detail-toggle');
    expect(toggles[0].getAttribute('aria-expanded')).toBe('false');
    expect(toggles[0].textContent).toContain('2 empaques');
    expect(toggles[0].textContent).toContain('200.00');
    expect(fixture.nativeElement.querySelectorAll('.purchase-detail-body')).toHaveLength(1);

    toggles[0].click();
    fixture.detectChanges();
    expect(component.estaContraido(0)).toBe(false);
    expect(toggles[0].getAttribute('aria-expanded')).toBe('true');
  });

  it('conserva el estado de los demás productos al quitar uno', async () => {
    const component = (await createFixture()).componentInstance;
    component.addDetail();
    component.addDetail();
    component.alternarDetalle(2);

    component.removeDetail(0);

    expect(component.details.length).toBe(2);
    expect(component.estaContraido(0)).toBe(true);
    expect(component.estaContraido(1)).toBe(true);
    expect(component.contraidos().size).toBe(2);
  });

  it('abre los productos incompletos al intentar guardar', async () => {
    const component = (await createFixture()).componentInstance;
    component.addDetail();
    expect(component.estaContraido(0)).toBe(true);

    component.submit();

    expect(component.estaContraido(0)).toBe(false);
    expect(service.create).not.toHaveBeenCalled();
  });

  it('captura el producto nuevo en una ventana y lo deja en el renglón, sin formularios incrustados', async () => {
    const fixture = await createFixture();
    const component = fixture.componentInstance;

    component.abrirProducto(0);
    fixture.detectChanges();
    const ventana: HTMLDialogElement = fixture.nativeElement.querySelectorAll('dialog')[1];
    expect(ventana.open).toBe(true);
    expect(ventana.textContent).toContain('Nuevo producto');
    expect(ventana.querySelectorAll('input')).toHaveLength(2);
    expect(fixture.nativeElement.querySelector('.purchase-product-form')).toBeNull();

    component.productForm.setValue({ nombre: ' Galletas ', codigo_barras: '' });
    component.usarProducto();
    fixture.detectChanges();

    expect(component.isNewProduct(0)).toBe(true);
    expect(component.details.at(0).get('nombre_producto_nuevo')?.value).toBe('Galletas');
    expect(component.ventana()).toBeNull();
    expect(ventana.open).toBe(false);
    expect(fixture.nativeElement.querySelector('.producto-nuevo')?.textContent).toContain('Galletas');

    component.quitarProductoNuevo(0);
    expect(component.isNewProduct(0)).toBe(false);
    expect(component.details.at(0).get('nombre_producto_nuevo')?.value).toBe('');
  });

  it('cierra la ventana con Esc o clic en el fondo', async () => {
    const fixture = await createFixture();
    const component = fixture.componentInstance;
    component.abrirProveedor();
    fixture.detectChanges();
    const ventana: HTMLDialogElement = fixture.nativeElement.querySelector('dialog');
    expect(ventana.open).toBe(true);

    ventana.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    fixture.detectChanges();

    expect(ventana.open).toBe(false);
    expect(component.ventana()).toBeNull();
  });

  it('preselecciona Pieza y muestra las piezas que entran', async () => {
    service.units.mockReturnValueOnce(of([
      { id: 'caja-1', nombre: 'Caja 12', tipo: 'EMPAQUE', factor_a_base: 12 },
      { id: 'pieza-1', nombre: 'Pieza', tipo: 'EMPAQUE', factor_a_base: 1 },
      { id: 'kilo-1', nombre: 'Kilo', tipo: 'CANTIDAD', factor_a_base: 1 },
    ]) as never);
    const fixture = await createFixture();
    const component = fixture.componentInstance;

    expect(component.details.at(0).get('unidad_medida_id')?.value).toBe('pieza-1');
    expect(component.unitOptions().map((unit) => unit.etiqueta)).toEqual(['Caja 12 (12 piezas)', 'Pieza (1 pieza)']);
    const empaques: HTMLInputElement = fixture.nativeElement.querySelector('[formcontrolname="cantidad_empaques"]');
    empaques.value = '8';
    empaques.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.purchase-pieces')?.textContent).toContain('8 piezas');

    component.addDetail();
    expect(component.details.at(1).get('unidad_medida_id')?.value).toBe('pieza-1');
    component.details.at(1).patchValue({ unidad_medida_id: 'caja-1', cantidad_empaques: 2 });
    expect(component.pieces(1)).toBe(24);
  });

  it('pide número de cuenta solo para transferencia o tarjeta y lo limpia al cambiar', async () => {
    const fixture = await createFixture();
    const component = fixture.componentInstance;
    const cuenta = () => fixture.nativeElement.querySelector('[formcontrolname="numero_cuenta"]');

    expect(cuenta()).toBeNull();
    component.form.controls.forma_pago.setValue('TRANSFERENCIA');
    fixture.detectChanges();
    expect(cuenta()).not.toBeNull();

    component.form.controls.numero_cuenta.setValue('012345678901234567');
    component.form.controls.forma_pago.setValue('EFECTIVO');
    fixture.detectChanges();
    expect(cuenta()).toBeNull();
    expect(component.form.controls.numero_cuenta.value).toBe('');
  });

  it('requiere solo el nombre al crear un producto nuevo; el código de barras es opcional', async () => {
    const component = (await createFixture()).componentInstance;

    component.selectProduct(0, '__new_product__');

    expect(component.details.at(0).get('nombre_producto_nuevo')?.invalid).toBe(true);
    expect(component.details.at(0).get('codigo_barras_nuevo')?.valid).toBe(true);

    component.details.at(0).patchValue({ nombre_producto_nuevo: 'Galletas' });

    expect(component.details.at(0).get('nombre_producto_nuevo')?.valid).toBe(true);
    expect(component.details.at(0).get('codigo_barras_nuevo')?.valid).toBe(true);

    component.details.at(0).patchValue({ codigo_barras_nuevo: '7501234567890' });
    expect(component.details.at(0).get('codigo_barras_nuevo')?.valid).toBe(true);
  });

  it('solo habilita el alta rápida de proveedor después de escribir su nombre', async () => {
    const component = (await createFixture()).componentInstance;

    component.abrirProveedor();
    expect(component.ventana()).toBe('proveedor');
    expect(component.providerForm.invalid).toBe(true);

    component.providerForm.controls.nombre.setValue('Distribuidora Norte');
    expect(component.providerForm.valid).toBe(true);
  });

  it('crea una unidad de empaque sin código y la selecciona en el detalle', async () => {
    const component = (await createFixture()).componentInstance;
    service.createUnit.mockReturnValue(of(undefined));
    service.units.mockReturnValue(
      of([
        { id: 'caja-1', nombre: 'Caja', tipo: 'EMPAQUE', factor_a_base: 12 },
        { id: 'paquete-24', nombre: 'Paquete', tipo: 'EMPAQUE', factor_a_base: 24 },
      ]),
    );

    component.startUnit(0);
    component.unitForm.patchValue({ nombre: 'Paquete', simbolo: 'paq', factor_a_base: 24 });
    component.createUnit();

    expect(service.createUnit).toHaveBeenCalledWith('negocio-1', {
      nombre: 'Paquete',
      simbolo: 'paq',
      tipo: 'EMPAQUE',
      factor_a_base: 24,
      decimales: 0,
    });
    expect(component.details.at(0).get('unidad_medida_id')?.value).toBe('paquete-24');
    expect(component.unitDetailIndex()).toBeNull();
  });
  it('agrupa todos los productos con los del proveedor primero y avisa si se ligará al guardar', async () => {
    service.products.mockReturnValueOnce(of([
      { id: 'pn-1', nombre: 'Arroz', sku: null, codigo_barras: null, del_proveedor: true },
      { id: 'pn-2', nombre: 'Zanahoria', sku: 'Z1', codigo_barras: '750', del_proveedor: false },
    ]) as never);
    const fixture = await createFixture();
    const component = fixture.componentInstance;
    component.form.controls.proveedor_id.setValue('proveedor-1');
    component.loadProducts();

    const groups = component.productOptions();
    expect(groups.map((group) => group.label)).toEqual(['De este proveedor', 'Otros productos']);
    expect(groups[0].items.map((item) => item.id)).toEqual(['pn-1']);
    expect(groups[1].items.map((item) => item.id)).toEqual(['pn-2']);

    component.details.at(0).patchValue({ producto_negocio_id: 'pn-1' });
    expect(component.seLigaraAlProveedor(0)).toBe(false);
    component.details.at(0).patchValue({ producto_negocio_id: 'pn-2' });
    expect(component.seLigaraAlProveedor(0)).toBe(true);
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('Se ligará a este proveedor al guardar.');
  });

  it('omite los grupos vacíos', async () => {
    const component = (await createFixture()).componentInstance;
    expect(component.productOptions()).toEqual([]);
  });
  it('usa scroll virtual en el selector de productos para catálogos grandes', async () => {
    const fixture = await createFixture();
    const select = fixture.debugElement
      .queryAll(By.directive(Select))
      .find((item) => item.attributes['formControlName'] === 'producto_negocio_id')?.componentInstance as Select;
    expect(select.virtualScroll()).toBe(true);
    expect(select.virtualScrollItemSize()).toBe(38);
  });
});
