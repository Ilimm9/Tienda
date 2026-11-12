import '@angular/common/locales/global/es-MX';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { ComprasListaComponent } from './compras-lista.component';
import { PurchaseRow } from './compras.models';
import { ComprasService } from './compras.service';

HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
  this.setAttribute('open', '');
};
HTMLDialogElement.prototype.close ??= function (this: HTMLDialogElement) {
  this.removeAttribute('open');
};

const compra = (id: string, cambios: Partial<PurchaseRow> = {}): PurchaseRow => ({
  id, proveedor_id: 'prov-1', proveedor: 'Ricolino', sucursal_id: 'sucursal-1', tipo_documento: 'FACTURA', folio_documento: `F-${id}`,
  fecha_recepcion: '2026-11-03T00:00:00Z', forma_pago: 'EFECTIVO', importe_pagado: 100, total: 100, estado: 'ACTIVA', productos: 2, piezas: 24,
  costos_pendientes: 0, costos_validados: 0, costos_cancelados: 0, ...cambios,
});

describe('ComprasListaComponent', () => {
  const service = { exportar: vi.fn(() => of(new Blob())) };

  async function crear(compras: PurchaseRow[], puede: (permiso: string) => boolean = () => true) {
    await TestBed.configureTestingModule({
      imports: [ComprasListaComponent],
      providers: [
        provideRouter([]),
        { provide: ComprasService, useValue: service },
        { provide: ContextoService, useValue: { puede, negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(ComprasListaComponent);
    fixture.componentRef.setInput('compras', compras);
    fixture.detectChanges();
    return fixture;
  }

  it('filtra por folio, proveedor y fecha de recepción', async () => {
    const component = (await crear([
      compra('1'),
      compra('2', { proveedor_id: 'prov-2', proveedor: 'Sabritas', fecha_recepcion: '2026-11-08T00:00:00Z' }),
    ])).componentInstance;
    component.q.set('sabri');
    expect(component.filtradas().map((c) => c.id)).toEqual(['2']);
    component.q.set('');
    component.desde.set('2026-11-05');
    expect(component.filtradas().map((c) => c.id)).toEqual(['2']);
    component.proveedor.set('prov-1');
    expect(component.resumen()).toBe('0 de 2 compras');
  });

  it('muestra el estado de los costos, el aviso y las acciones por estado', async () => {
    const fixture = await crear([compra('1', { costos_pendientes: 3 }), compra('2', { estado: 'SUSPENDIDA', costos_cancelados: 2 })]);
    const component = fixture.componentInstance;
    expect(component.costos(component.compras()[0]).texto).toBe('3 por validar');
    expect(component.costos(component.compras()[1]).texto).toBe('Cancelados');
    const texto = fixture.nativeElement.textContent as string;
    expect(texto).toContain('3 costos nuevos');
    expect(fixture.nativeElement.querySelector('[aria-label="Reactivar Factura F-2"]')).not.toBeNull();
    expect(fixture.nativeElement.querySelector('[aria-label="Editar Factura F-2"]')).toBeNull();
    expect(fixture.nativeElement.querySelector('.compras-tabla__suspendida')).not.toBeNull();
  });

  it('sin permisos oculta editar y suspender y muestra la nota con candado', async () => {
    const fixture = await crear([compra('1')], (permiso) => !['compras.editar', 'compras.suspender'].includes(permiso));
    expect(fixture.nativeElement.querySelector('[aria-label="Editar Factura F-1"]')).toBeNull();
    expect(fixture.nativeElement.querySelector('.compras-candado')).not.toBeNull();
  });

  it('descarga las compras seleccionadas', async () => {
    const component = (await crear([compra('1'), compra('2')])).componentInstance;
    URL.createObjectURL ??= () => 'blob:prueba';
    URL.revokeObjectURL ??= () => undefined;
    component.alternar(component.compras()[1], true);
    component.abrirExcel();
    expect(component.excel().alcance).toBe('seleccion');
    expect(component.resumenExcel()).toMatchObject({ compras: 1, renglones: 2, importe: 100 });
    component.descargarExcel();
    expect(service.exportar).toHaveBeenCalledWith('negocio-1', { sucursal_id: 'sucursal-1', comprobacion: true, compra_id: ['2'], incluir_suspendidas: false });
  });
});
