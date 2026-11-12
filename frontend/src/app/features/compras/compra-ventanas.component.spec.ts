import '@angular/common/locales/global/es-MX';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { CompraVentanasComponent } from './compra-ventanas.component';
import { PurchaseRow, PurchaseView, PurchaseViewRow } from './compras.models';
import { ComprasService } from './compras.service';

HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
  this.setAttribute('open', '');
};
HTMLDialogElement.prototype.close ??= function (this: HTMLDialogElement) {
  this.removeAttribute('open');
};

const fila = (id: string, cambios: Partial<PurchaseViewRow> = {}): PurchaseViewRow => ({
  id, producto_negocio_id: `p-${id}`, unidad_medida_id: 'u-1', producto: `Producto ${id}`, unidad: 'Caja 12', cantidad_empaques: 2,
  piezas_por_empaque: 12, total_piezas: 24, piezas_buenas: 24, piezas_danadas: 0, piezas_faltantes: 0, costo_empaque: 120, descuento: 0,
  subtotal: 240, ieps: 19.2, iva: 41.47, tasa_ieps: 0.08, tasa_iva: 0.16, impuestos_por_renglon: true, numero_lote: null,
  fecha_caducidad: null, costo_final: 12.53, estado_costo: 'PENDIENTE', ...cambios,
});
const vista = (cambios: Partial<PurchaseView> = {}): PurchaseView => ({
  id: 'c-1', proveedor_id: 'prov-1', proveedor: 'Ricolino', sucursal_id: 's-1', tipo_documento: 'TICKET', folio_documento: 'T-77',
  fecha_documento: '2026-11-03T00:00:00Z', fecha_recepcion: '2026-11-03T00:00:00Z', forma_pago: 'EFECTIVO', importe_pagado: 601.34,
  total: 601.34, subtotal: 480, ieps: 38.4, iva: 82.94, estado: 'ACTIVA', renglones: [fila('1'), fila('2', { estado_costo: 'AUTORIZADA' })],
  ...cambios,
});

describe('CompraVentanasComponent', () => {
  const service = {
    view: vi.fn(() => of(vista())),
    edit: vi.fn(() => of({})),
    suspend: vi.fn(() => of(undefined)),
    reactivate: vi.fn(() => of(undefined)),
    exportar: vi.fn(() => of(new Blob())),
  };

  async function abrir(tipo: 'ver' | 'editar' | 'suspender' | 'reactivar', datos = vista()) {
    service.view.mockReturnValue(of(datos));
    await TestBed.configureTestingModule({
      imports: [CompraVentanasComponent],
      providers: [
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'n-1' }), sucursal: signal({ id: 's-1' }) } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(CompraVentanasComponent);
    fixture.componentRef.setInput('accion', { tipo, compra: { id: datos.id } as PurchaseRow });
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => Object.values(service).forEach((f) => f.mockClear()));

  it('muestra el detalle con renglones, impuestos y estado de costo', async () => {
    const fixture = await abrir('ver');
    const texto = fixture.nativeElement.textContent as string;
    expect(service.view).toHaveBeenCalledWith('n-1', 'c-1');
    expect(fixture.componentInstance.titulo()).toBe('Ticket T-77');
    expect(texto).toContain('Producto 1');
    expect(texto).toContain('Por validar');
    expect(texto).toContain('Validado');
  });

  it('con un costo validado bloquea productos y solo manda el documento', async () => {
    const fixture = await abrir('editar');
    const component = fixture.componentInstance;
    expect(component.bloqueada()).toBe(true);
    expect(fixture.nativeElement.textContent).toContain('no se pueden cambiar');
    component.cambiarEncabezado('folio_documento', 'T-78');
    component.guardarEdicion();
    const [, , request] = service.edit.mock.calls[0] as unknown as [string, string, { folio_documento: string; detalles?: unknown }];
    expect(request.folio_documento).toBe('T-78');
    expect(request.detalles).toBeUndefined();
  });

  it('sin costos validados manda los renglones con tasas en fracción', async () => {
    const fixture = await abrir('editar', vista({ renglones: [fila('1'), fila('2')] }));
    const component = fixture.componentInstance;
    component.cambiarRenglon(0, { empaques: 3, ieps: 0 });
    component.guardarEdicion();
    const [, , request] = service.edit.mock.calls[0] as unknown as [string, string, { detalles: { id: string; cantidad_empaques: number; tasa_ieps: number; tasa_iva: number }[] }];
    expect(request.detalles[0]).toMatchObject({ id: '1', cantidad_empaques: 3, tasa_ieps: 0, tasa_iva: 0.16 });
    expect(request.detalles[1]).toMatchObject({ id: '2', tasa_ieps: 0.08 });
  });

  it('suspender exige motivo y avisa del cambio', async () => {
    const fixture = await abrir('suspender');
    const component = fixture.componentInstance;
    const cambios: string[] = [];
    component.cambio.subscribe((m) => cambios.push(m));
    expect(fixture.nativeElement.textContent).toContain('48 piezas');
    component.confirmarSuspension();
    expect(service.suspend).not.toHaveBeenCalled();
    component.motivo.set('Proveedor recogió');
    component.confirmarSuspension();
    expect(service.suspend).toHaveBeenCalledWith('n-1', 'c-1', 'Proveedor recogió');
    expect(cambios).toEqual(['Compra suspendida.']);
    expect(component.tipo()).toBeNull();
  });

  it('reactivar cuenta los costos cancelados que regresan', async () => {
    const fixture = await abrir('reactivar', vista({ estado: 'SUSPENDIDA', renglones: [fila('1', { estado_costo: 'CANCELADA' })] }));
    expect(fixture.nativeElement.textContent).toContain('1 costo cancelado regresa');
    fixture.componentInstance.confirmarReactivacion();
    expect(service.reactivate).toHaveBeenCalledWith('n-1', 'c-1', '');
  });
});
