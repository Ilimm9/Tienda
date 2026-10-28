import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { VentasComponent } from './ventas.component';
import { ListadoVentas, VentaRegistrada } from './ventas.models';
import { VentasService } from './ventas.service';

const venta = (folio: string, extra: Partial<VentaRegistrada>): VentaRegistrada =>
  ({
    id: folio,
    folio,
    numero: 1,
    sucursal_id: 's1',
    usuario_id: 'u1',
    subtotal: 0,
    iva: 0,
    total: 28,
    recibido: 28,
    cambio: 0,
    piezas: 1,
    forma_pago: 'EFECTIVO',
    estado: 'PAGADA',
    creado_en: '2026-10-23T19:15:00Z',
    atendio: 'Laura Martínez',
    cancelable: true,
    ...extra,
  }) as VentaRegistrada;

const listado: ListadoVentas = {
  items: [
    venta('V-000231', { estado: 'CANCELADA', total: 64, piezas: 3 }),
    venta('V-000230', { forma_pago: 'TARJETA' }),
    venta('V-000229', { atendio: 'Paola Reyes', total: 161, piezas: 6 }),
  ],
  total: 3,
  resumen: { vendido: 189, tickets: 2, ticket_promedio: 94.5, efectivo: 161, cancelados: 1 },
  desde: '2026-10-23T06:00:00Z',
  hasta: '2026-10-24T06:00:00Z',
};

describe('VentasComponent', () => {
  let service: { listar: ReturnType<typeof vi.fn> };

  async function crear(): Promise<ComponentFixture<VentasComponent>> {
    service = { listar: vi.fn(() => of(listado)) };
    await TestBed.configureTestingModule({
      imports: [VentasComponent],
      providers: [
        provideRouter([]),
        { provide: VentasService, useValue: service },
        {
          provide: ContextoService,
          useValue: {
            puede: () => true,
            negocio: signal({ id: 'n1' }),
            sucursal: signal({ id: 's1', nombre: 'Sucursal Reforma' }),
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(VentasComponent);
    fixture.detectChanges();
    return fixture;
  }

  const texto = (fixture: ComponentFixture<VentasComponent>) =>
    (fixture.nativeElement as HTMLElement).textContent ?? '';
  const folios = (fixture: ComponentFixture<VentasComponent>) =>
    Array.from(fixture.nativeElement.querySelectorAll('tbody .folio')).map((celda) =>
      (celda as HTMLElement).textContent?.trim(),
    );

  beforeEach(() => TestBed.resetTestingModule());

  it('muestra la sucursal, el día y el resumen del periodo', async () => {
    const fixture = await crear();
    expect(service.listar).toHaveBeenCalledWith('n1', 's1', 'hoy');
    expect(texto(fixture)).toContain('Sucursal Reforma');
    expect(texto(fixture)).toContain('Viernes 23 de octubre');
    const metricas = Array.from(fixture.nativeElement.querySelectorAll('.metrica strong')).map(
      (m) => (m as HTMLElement).textContent?.trim(),
    );
    expect(metricas).toEqual(['$189.00', '2', '$94.50', '$161.00']);
    expect(fixture.nativeElement.querySelector('.ventas__pie')?.textContent).toContain(
      '3 tickets hoy',
    );
    expect(fixture.nativeElement.querySelector('.ventas__pie')?.textContent).toContain(
      '1 cancelado',
    );
  });

  it('la tabla muestra hora, cajero, pago y tacha las canceladas', async () => {
    const fixture = await crear();
    const fila = fixture.nativeElement.querySelector('tbody tr') as HTMLElement;
    expect(fila.classList).toContain('fila--cancelada');
    expect(fila.textContent).toContain('13:15');
    expect(fila.textContent).toContain('Laura Martínez');
    expect(fila.textContent).toContain('3 piezas');
    expect(fila.textContent).toContain('Cancelada');
    expect(fila.querySelector('a')?.getAttribute('href')).toBe('/ventas/V-000231');
  });

  it('filtra por folio o cajero sin acentos y por forma de pago', async () => {
    const fixture = await crear();
    const buscar = fixture.nativeElement.querySelector('input[type="search"]') as HTMLInputElement;
    buscar.value = 'paola';
    buscar.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(folios(fixture)).toEqual(['V-000229']);
    buscar.value = 'martinez';
    buscar.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(folios(fixture)).toEqual(['V-000231', 'V-000230']);
    const forma = fixture.nativeElement.querySelector('select') as HTMLSelectElement;
    forma.value = 'TARJETA';
    forma.dispatchEvent(new Event('change'));
    fixture.detectChanges();
    expect(folios(fixture)).toEqual(['V-000230']);
  });

  it('cambiar el periodo vuelve a consultar', async () => {
    const fixture = await crear();
    (fixture.nativeElement.querySelectorAll('.segmentado__opcion')[1] as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(service.listar).toHaveBeenLastCalledWith('n1', 's1', 'semana');
    expect(texto(fixture)).toContain('Fecha');
  });
});
