import '@angular/common/locales/global/es-MX';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { PreciosComponent } from './precios.component';
import { PriceProposal, PreciosService } from './precios.service';

function propuesta(id: string, compra: string, cambios: Partial<PriceProposal> = {}): PriceProposal {
  return {
    id, producto_negocio_id: `p-${id}`, producto: `Producto ${id}`, piezas_buenas: '12', costo_capturado: '25.0558', costo_anterior: '20',
    precio_anterior: '28', margen_sugerido: '0.3', precio_sugerido: '33', compra_id: compra, tipo_documento: 'FACTURA',
    folio_documento: `F-${compra}`, proveedor_id: 'prov-1', proveedor: 'Ricolino', fecha_documento: '2026-11-01', fecha_recepcion: '2026-11-03T00:00:00Z',
    total_compra: '564.67', codigo_barras: null, nota: '1 × Caja 12 · IEPS 8 % · IVA 16 %', impuestos_por_renglon: true, tasa_ieps: '0.08', tasa_iva: '0.16',
    producto_nuevo: false, ...cambios,
  };
}

describe('PreciosComponent', () => {
  const service = {
    pending: vi.fn(),
    authorize: vi.fn(() => of({})),
    authorizePurchase: vi.fn(() => of({})),
    history: vi.fn(),
    adjust: vi.fn(),
  };

  async function crear(items: PriceProposal[]) {
    service.pending.mockReturnValue(of({ items }));
    await TestBed.configureTestingModule({
      imports: [PreciosComponent],
      providers: [
        provideRouter([]),
        { provide: PreciosService, useValue: service },
        { provide: ContextoService, useValue: { puede: () => true, negocio: signal({ id: 'negocio-1' }), sucursal: signal({ id: 'sucursal-1' }) } },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(PreciosComponent);
    fixture.detectChanges();
    return fixture;
  }

  it('agrupa las propuestas por compra con sus impuestos', async () => {
    const fixture = await crear([propuesta('a', 'c1'), propuesta('b', 'c1', { tasa_ieps: '0', costo_anterior: null }), propuesta('c', 'c2')]);
    const component = fixture.componentInstance;
    expect(component.gruposVisibles().map((t) => t.renglones.length)).toEqual([2, 1]);
    expect(component.tickets()[0].impuestos).toBe('IEPS 8 % · IVA 16 %');
    expect(component.cambioCosto(component.renglones()[1]).texto).toBe('Primer costo');
    expect(component.resumenFiltro()).toBe('2 tickets · 3 productos');
    expect(fixture.nativeElement.querySelectorAll('.validar-ticket').length).toBe(2);
  });

  it('usa el precio sugerido del backend y redondea al peso al cambiar el margen', async () => {
    const component = (await crear([propuesta('a', 'c1', { costo_capturado: '10', precio_sugerido: '13' })])).componentInstance;
    const [renglon] = component.renglones();
    expect(renglon.precio).toBe(13);
    component.cambiarMargen(renglon, 10);
    expect(component.renglones()[0].precio).toBe(11);
  });

  it('respeta el precio manual y recalcula el margen y la ganancia', async () => {
    const component = (await crear([propuesta('a', 'c1', { costo_capturado: '20', precio_sugerido: '26' })])).componentInstance;
    component.cambiarPrecio(component.renglones()[0], 25);
    const renglon = component.renglones()[0];
    expect(renglon.precio).toBe(25);
    expect(renglon.margen).toBe(25);
    expect(component.ganancia(renglon)).toBe(5);
    expect(component.gananciaTotalRenglon(renglon)).toBe(60);
  });

  it('valida el ticket entero en una llamada con los precios de cada renglón', async () => {
    const component = (await crear([propuesta('a', 'c1'), propuesta('b', 'c1')])).componentInstance;
    component.validarTicket(component.gruposVisibles()[0]);
    expect(service.authorizePurchase).toHaveBeenCalledWith('negocio-1', 'c1', [
      { propuesta_id: 'a', costo: 25.0558, margen: 0.3, precio_venta: 33 },
      { propuesta_id: 'b', costo: 25.0558, margen: 0.3, precio_venta: 33 },
    ]);
    expect(component.renglones().length).toBe(0);
  });

  it('aplica el margen a los seleccionados', async () => {
    const component = (await crear([propuesta('a', 'c1', { costo_capturado: '20' }), propuesta('b', 'c2', { costo_capturado: '20' })])).componentInstance;
    component.alternarRenglon(component.renglones()[1], true);
    component.margenMasivo.set(10);
    component.aplicarMargen();
    expect(component.renglones().map((r) => r.precio)).toEqual([33, 22]);
  });

  it('muestra «Todo validado» sin pendientes', async () => {
    const fixture = await crear([]);
    expect(fixture.nativeElement.textContent).toContain('Todo validado');
  });
});
