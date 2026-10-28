import '@angular/common/locales/global/es-MX';
import { HttpErrorResponse } from '@angular/common/http';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { TicketVentaComponent } from './ticket-venta.component';
import { VentaRegistrada } from './ventas.models';
import { VentasService } from './ventas.service';

HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
  this.setAttribute('open', '');
};
HTMLDialogElement.prototype.close ??= function (this: HTMLDialogElement) {
  this.removeAttribute('open');
};

const ticket: VentaRegistrada = {
  id: 'v230',
  folio: 'V-000230',
  numero: 230,
  sucursal_id: 's1',
  usuario_id: 'u1',
  subtotal: 24.14,
  iva: 3.86,
  total: 28,
  recibido: 50,
  cambio: 22,
  piezas: 1,
  forma_pago: 'EFECTIVO',
  estado: 'PAGADA',
  creado_en: '2026-10-23T19:15:00Z',
  atendio: 'Laura Martínez',
  cancelable: true,
  lineas: [
    {
      id: 'l1',
      producto_negocio_id: 'p1',
      nombre: 'Leche Lala entera 1 L',
      cantidad: 1,
      precio_unitario: 28,
      importe: 28,
    },
  ],
};

describe('TicketVentaComponent', () => {
  let service: { obtener: ReturnType<typeof vi.fn>; cancelar: ReturnType<typeof vi.fn> };
  let permisos: Set<string>;
  const feedback = { success: vi.fn() };

  async function crear(
    venta: VentaRegistrada = ticket,
  ): Promise<ComponentFixture<TicketVentaComponent>> {
    service = {
      obtener: vi.fn(() => of(venta)),
      cancelar: vi.fn(() =>
        of({
          ...venta,
          estado: 'CANCELADA',
          cancelable: false,
          cancelada_por: 'Ana Gerente',
          cancelada_en: '2026-10-23T20:00:00Z',
          motivo_cancelacion: 'Cobro duplicado',
        }),
      ),
    };
    await TestBed.configureTestingModule({
      imports: [TicketVentaComponent],
      providers: [
        provideRouter([]),
        {
          provide: ActivatedRoute,
          useValue: { paramMap: of(convertToParamMap({ ventaId: venta.id })) },
        },
        { provide: VentasService, useValue: service },
        { provide: FeedbackService, useValue: feedback },
        {
          provide: ContextoService,
          useValue: {
            puede: (p: string) => permisos.has(p),
            negocio: signal({
              id: 'n1',
              nombre_comercial: 'Tienda',
              sucursales: [{ id: 's1', nombre: 'Reforma' }],
            }),
            sucursal: signal({ id: 's1', nombre: 'Reforma' }),
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(TicketVentaComponent);
    fixture.detectChanges();
    return fixture;
  }

  const boton = (fixture: ComponentFixture<TicketVentaComponent>, textoBoton: string) =>
    Array.from(fixture.nativeElement.querySelectorAll('.ticket-venta__acciones button')).find(
      (b) => (b as HTMLElement).textContent?.trim() === textoBoton,
    ) as HTMLButtonElement | undefined;

  beforeEach(() => {
    TestBed.resetTestingModule();
    permisos = new Set([PERMISOS.ventaVer, PERMISOS.ventaCancelar]);
    feedback.success.mockClear();
  });

  it('muestra encabezado, líneas y resumen sin caja', async () => {
    const fixture = await crear();
    const texto = (fixture.nativeElement as HTMLElement).textContent ?? '';
    expect(service.obtener).toHaveBeenCalledWith('n1', 'v230');
    expect(texto).toContain('Ticket V-000230');
    expect(texto).toContain('Viernes 23 de octubre de 2026, 13:15 · Atendió Laura Martínez');
    expect(texto).not.toContain('Caja');
    expect(fixture.nativeElement.querySelector('.tabla-ticket tbody')?.textContent).toContain(
      'Leche Lala entera 1 L',
    );
    expect(fixture.nativeElement.querySelector('.ticket-venta__resumen')?.textContent).toContain(
      '$22.00',
    );
  });

  it('sin ventas.cancelar no aparece el botón; fuera de plazo queda deshabilitado', async () => {
    permisos = new Set([PERMISOS.ventaVer]);
    let fixture = await crear();
    expect(boton(fixture, 'Cancelar venta')).toBeUndefined();

    TestBed.resetTestingModule();
    permisos = new Set([PERMISOS.ventaVer, PERMISOS.ventaCancelar]);
    fixture = await crear({ ...ticket, cancelable: false });
    expect(boton(fixture, 'Cancelar venta')?.disabled).toBe(true);
  });

  it('cancelar exige motivo, «Otro» exige explicación, y al confirmar tacha el ticket', async () => {
    const fixture = await crear();
    boton(fixture, 'Cancelar venta')!.click();
    fixture.detectChanges();
    const dialogo = fixture.nativeElement.querySelector('dialog.cancelar') as HTMLDialogElement;
    expect(dialogo.open).toBe(true);
    expect(dialogo.textContent).toContain('Se regresarán 1 pieza al inventario');
    const confirmar = dialogo.querySelector('button[type="submit"]') as HTMLButtonElement;
    expect(confirmar.disabled).toBe(true);

    const radios = Array.from(
      dialogo.querySelectorAll('input[type="radio"]'),
    ) as HTMLInputElement[];
    radios[3].click();
    fixture.detectChanges();
    expect(confirmar.disabled).toBe(true);
    const detalle = dialogo.querySelector('textarea') as HTMLTextAreaElement;
    detalle.value = 'Cliente cambió de opinión';
    detalle.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(confirmar.disabled).toBe(false);

    radios[1].click();
    fixture.detectChanges();
    confirmar.click();
    fixture.detectChanges();
    expect(service.cancelar).toHaveBeenCalledWith(
      'n1',
      'v230',
      'COBRO_DUPLICADO',
      'Cliente cambió de opinión',
    );
    expect(dialogo.open).toBe(false);
    expect(feedback.success).toHaveBeenCalledWith(
      'Venta V-000230 cancelada',
      'Las piezas regresaron al inventario.',
    );
    const resumen = fixture.nativeElement.querySelector('.ticket-venta__resumen') as HTMLElement;
    expect(resumen.textContent).toContain('Cancelada por Ana Gerente');
    expect(resumen.textContent).toContain('Cobro duplicado');
    expect(resumen.querySelector('.ticket-venta__total--cancelado')).not.toBeNull();
    expect(boton(fixture, 'Cancelar venta')).toBeUndefined();
  });

  it('si ya pasó el día, avisa y deshabilita la cancelación', async () => {
    const fixture = await crear();
    service.cancelar.mockReturnValueOnce(
      throwError(
        () => new HttpErrorResponse({ status: 409, error: { codigo: 'VENTA_FUERA_DE_PLAZO' } }),
      ),
    );
    boton(fixture, 'Cancelar venta')!.click();
    fixture.detectChanges();
    const dialogo = fixture.nativeElement.querySelector('dialog.cancelar') as HTMLDialogElement;
    (dialogo.querySelectorAll('input[type="radio"]')[0] as HTMLInputElement).click();
    fixture.detectChanges();
    (dialogo.querySelector('button[type="submit"]') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(dialogo.querySelector('[role="alert"]')?.textContent).toContain(
      'Solo se pueden cancelar ventas del mismo día',
    );
    expect(boton(fixture, 'Cancelar venta')?.disabled).toBe(true);
  });

  it('reimprimir abre la impresión con el ticket', async () => {
    const fixture = await crear();
    const imprimir = vi.spyOn(window, 'print').mockImplementation(() => undefined);
    boton(fixture, 'Reimprimir')!.click();
    expect(imprimir).toHaveBeenCalled();
    imprimir.mockRestore();
    expect(fixture.nativeElement.querySelector('.ticket-imprimible')?.textContent).toContain(
      '1 × Leche Lala entera 1 L',
    );
  });

  it('un ticket ajeno o inexistente muestra el aviso', async () => {
    const fixture = await crear();
    service.obtener.mockReturnValue(throwError(() => new HttpErrorResponse({ status: 404 })));
    (fixture.componentInstance as unknown as { cargar(n: string, v: string): void }).cargar(
      'n1',
      'otra',
    );
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('No encontramos este ticket');
  });
});
