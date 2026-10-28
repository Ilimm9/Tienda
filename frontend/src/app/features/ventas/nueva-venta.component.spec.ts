import '@angular/common/locales/global/es-MX';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { Observable, of, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { PERMISOS } from '../../contexto/permisos';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NuevaVentaComponent, recibidoEnCentavos } from './nueva-venta.component';
import { ProductoVendible, ResultadoBusqueda, VentaRegistrada } from './ventas.models';
import { VentasService } from './ventas.service';

const coca: ProductoVendible = {
  producto_negocio_id: 'coca',
  nombre: 'Coca-Cola 600 ml',
  codigo_barras: '7501055303038',
  precio: 18,
  existencia: 2,
  agotado: false,
  existencia_baja: false,
  costo: 12,
};
const leche: ProductoVendible = {
  ...coca,
  producto_negocio_id: 'leche',
  nombre: 'Leche Lala entera 1 L',
  codigo_barras: '7501020565119',
  precio: 28,
  costo: 30,
};
const galletas: ProductoVendible = {
  ...coca,
  producto_negocio_id: 'galletas',
  nombre: 'Galletas Marías',
  codigo_barras: '123456',
  existencia: 0,
  agotado: true,
};

// jsdom no implementa la ventana modal nativa.
HTMLDialogElement.prototype.showModal ??= function (this: HTMLDialogElement) {
  this.setAttribute('open', '');
};
HTMLDialogElement.prototype.close ??= function (this: HTMLDialogElement) {
  this.removeAttribute('open');
};

describe('recibidoEnCentavos', () => {
  it('convierte el texto capturado sin flotantes', () => {
    expect(recibidoEnCentavos('100')).toBe(10000);
    expect(recibidoEnCentavos('64.5')).toBe(6450);
    expect(recibidoEnCentavos('$1,000.07')).toBe(100007);
    expect(recibidoEnCentavos('')).toBeNull();
    expect(recibidoEnCentavos('10.999')).toBeNull();
    expect(recibidoEnCentavos('-5')).toBeNull();
  });
});

describe('NuevaVentaComponent', () => {
  let registrar: ReturnType<typeof vi.fn>;
  let catalogo: Record<string, ResultadoBusqueda>;
  let borradores: Record<string, unknown>;
  let permisos: Set<string>;
  const feedback = { conAccion: vi.fn(), success: vi.fn(), confirmDanger: vi.fn(async () => true) };

  async function crear(): Promise<ComponentFixture<NuevaVentaComponent>> {
    const service = {
      buscar: vi.fn((_n: string, _s: string, q: string) =>
        of(catalogo[q] ?? { items: [], por_codigo: /^\d{6,}$/.test(q) }),
      ),
      leerBorrador: vi.fn((n: string, s: string) => (borradores[`${n}.${s}`] as never) ?? []),
      guardarBorrador: vi.fn(
        (n: string, s: string, lineas: unknown) => (borradores[`${n}.${s}`] = lineas),
      ),
      registrar,
    };
    await TestBed.configureTestingModule({
      imports: [NuevaVentaComponent],
      providers: [
        provideRouter([]),
        { provide: VentasService, useValue: service },
        { provide: FeedbackService, useValue: feedback },
        {
          provide: ContextoService,
          useValue: {
            puede: (permiso: string) => permisos.has(permiso),
            negocio: signal({ id: 'negocio-1' }),
            sucursal: signal({ id: 'sucursal-1', nombre: 'Sucursal Reforma' }),
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(NuevaVentaComponent);
    fixture.detectChanges();
    return fixture;
  }

  function escanear(fixture: ComponentFixture<NuevaVentaComponent>, texto: string): void {
    const campo = fixture.nativeElement.querySelector('input[role="combobox"]') as HTMLInputElement;
    campo.value = texto;
    campo.dispatchEvent(new Event('input'));
    campo.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
    fixture.detectChanges();
  }

  const filas = (fixture: ComponentFixture<NuevaVentaComponent>) =>
    Array.from(fixture.nativeElement.querySelectorAll('tbody tr')) as HTMLElement[];
  const texto = (fixture: ComponentFixture<NuevaVentaComponent>) =>
    (fixture.nativeElement as HTMLElement).textContent ?? '';

  beforeEach(() => {
    TestBed.resetTestingModule();
    catalogo = {
      '7501055303038': { items: [coca], por_codigo: true },
      '7501020565119': { items: [leche], por_codigo: true },
      '123456': { items: [galletas], por_codigo: true },
    };
    borradores = {};
    permisos = new Set([PERMISOS.ventaRegistrar]);
    feedback.conAccion.mockClear();
    feedback.success.mockClear();
    registrar = vi.fn((): Observable<VentaRegistrada> =>
      of({
        id: 'v1',
        folio: 'V-000233',
        total: 64,
        subtotal: 55.17,
        iva: 8.83,
        recibido: 100,
        cambio: 36,
        piezas: 3,
        creado_en: '2026-10-28T18:00:00Z',
        lineas: [
          {
            id: 'l1',
            producto_negocio_id: 'coca',
            nombre: 'Coca-Cola 600 ml',
            cantidad: 2,
            precio_unitario: 18,
            importe: 36,
          },
        ],
      } as VentaRegistrada),
    );
  });

  function abrirCobro(fixture: ComponentFixture<NuevaVentaComponent>): HTMLDialogElement {
    (fixture.nativeElement.querySelector('.resumen-venta .boton') as HTMLButtonElement).click();
    fixture.detectChanges();
    return fixture.nativeElement.querySelector('dialog.cobro') as HTMLDialogElement;
  }
  const confirmar = (dialogo: HTMLDialogElement) =>
    dialogo.querySelector('button[type="submit"]') as HTMLButtonElement;

  it('F2 regresa el foco al campo de captura', async () => {
    const fixture = await crear();
    const campo = fixture.nativeElement.querySelector('input[role="combobox"]') as HTMLInputElement;
    campo.blur();
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'F2' }));
    expect(document.activeElement).toBe(campo);
  });

  it('«Terminar venta» se habilita solo con productos', async () => {
    const fixture = await crear();
    const terminar = fixture.nativeElement.querySelector(
      '.resumen-venta .boton',
    ) as HTMLButtonElement;
    expect(terminar.disabled).toBe(true);
    escanear(fixture, '7501055303038');
    expect(terminar.disabled).toBe(false);
  });

  it('el cobro en efectivo bloquea con «Falta» y confirma con el cambio', async () => {
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501020565119');
    const dialogo = abrirCobro(fixture);
    expect(dialogo.open).toBe(true);
    expect(dialogo.textContent).toContain('$64.00');
    expect(dialogo.textContent).toContain('3 piezas');
    expect(dialogo.querySelector('.cobro__saldo--falta')?.textContent).toContain('$64.00');
    expect(confirmar(dialogo).disabled).toBe(true);
    expect(
      Array.from(dialogo.querySelectorAll('.cobro__opcion:disabled')).map((b) =>
        b.textContent?.trim(),
      ),
    ).toEqual(['Tarjeta', 'Transferencia']);
    expect(
      Array.from(dialogo.querySelectorAll('.cobro__rapidos button')).map((b) =>
        b.textContent?.trim(),
      ),
    ).toEqual(['Exacto', '$100', '$200', '$500']);

    const recibido = dialogo.querySelector('.cobro__campo input') as HTMLInputElement;
    expect(document.activeElement).toBe(recibido);
    recibido.value = '50';
    recibido.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(dialogo.querySelector('.cobro__saldo--falta')?.textContent).toContain('$14.00');
    expect(confirmar(dialogo).disabled).toBe(true);

    (dialogo.querySelectorAll('.cobro__rapidos button')[1] as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(dialogo.querySelector('.cobro__saldo--cambio')?.textContent).toContain('$36.00');
    confirmar(dialogo).click();
    fixture.detectChanges();

    expect(registrar).toHaveBeenCalledWith('negocio-1', {
      sucursal_id: 'sucursal-1',
      forma_pago: 'EFECTIVO',
      recibido: '100.00',
      lineas: [
        { producto_negocio_id: 'coca', cantidad: 2 },
        { producto_negocio_id: 'leche', cantidad: 1 },
      ],
    });
    expect(borradores['negocio-1.sucursal-1']).toEqual([]);
    expect(dialogo.open).toBe(false);

    const terminada = fixture.nativeElement.querySelector('.venta-terminada') as HTMLElement;
    expect(terminada.textContent).toContain('Venta V-000233 cobrada');
    expect(terminada.textContent).toContain('$64.00');
    expect(terminada.querySelector('.venta-terminada__cambio')?.textContent).toContain('$36.00');
    expect(terminada.textContent).toContain('3 piezas');
    expect(terminada.textContent).toContain('Las existencias de la sucursal ya se descontaron.');
    expect(fixture.nativeElement.querySelector('.ticket-imprimible')?.textContent).toContain(
      '2 × Coca-Cola 600 ml',
    );

    const imprimir = vi.spyOn(window, 'print').mockImplementation(() => undefined);
    (terminada.querySelector('.boton--secundario') as HTMLButtonElement).click();
    expect(imprimir).toHaveBeenCalled();
    imprimir.mockRestore();

    (
      terminada.querySelector(
        '.venta-terminada__acciones .boton:not(.boton--secundario)',
      ) as HTMLButtonElement
    ).click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.venta-terminada')).toBeNull();
    expect(filas(fixture).length).toBe(0);
    expect(fixture.nativeElement.querySelector('input[role="combobox"]')).not.toBeNull();
  });

  it('si otro cajero vendió la última pieza, avisa el producto y conserva la lista', async () => {
    registrar = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              codigo: 'EXISTENCIA_INSUFICIENTE',
              mensaje:
                'Ya no hay existencias suficientes de Coca-Cola 600 ml: quedan 1 y pediste 2',
              producto: 'Coca-Cola 600 ml',
              disponible: 1,
            },
          }),
      ),
    );
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501055303038');
    const dialogo = abrirCobro(fixture);
    (dialogo.querySelector('.cobro__rapidos button') as HTMLButtonElement).click();
    fixture.detectChanges();
    confirmar(dialogo).click();
    fixture.detectChanges();
    expect(dialogo.querySelector('[role="alert"]')?.textContent).toContain('Coca-Cola 600 ml');
    expect(dialogo.open).toBe(true);
    expect(filas(fixture)[0].querySelector('output')?.textContent?.trim()).toBe('2');

    (dialogo.querySelector('.cobro__pie .boton-enlace') as HTMLButtonElement).click();
    fixture.detectChanges();
    const [menos, mas] = Array.from(
      filas(fixture)[0].querySelectorAll('.cantidad button'),
    ) as HTMLButtonElement[];
    expect(mas.disabled).toBe(true);
    menos.click();
    fixture.detectChanges();
    expect(filas(fixture)[0].querySelector('output')?.textContent?.trim()).toBe('1');
  });

  it('muestra la sucursal en el encabezado y enfoca la captura', async () => {
    const fixture = await crear();
    expect(texto(fixture)).toContain('Ventas · Sucursal Reforma');
    expect(document.activeElement).toBe(
      fixture.nativeElement.querySelector('input[role="combobox"]'),
    );
  });

  it('un código escaneado dos veces queda en una línea con cantidad 2 y resaltada', async () => {
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501055303038');
    expect(filas(fixture).length).toBe(1);
    expect(filas(fixture)[0].querySelector('output')?.textContent?.trim()).toBe('2');
    expect(filas(fixture)[0].classList).toContain('fila--ultima');
    expect(
      (fixture.nativeElement.querySelector('input[role="combobox"]') as HTMLInputElement).value,
    ).toBe('');
  });

  it('un código que no existe avisa y no cambia la lista', async () => {
    const fixture = await crear();
    escanear(fixture, '7500000000000');
    expect(texto(fixture)).toContain('No encontramos el código «7500000000000»');
    expect(filas(fixture).length).toBe(0);
  });

  it('un producto agotado no se agrega', async () => {
    const fixture = await crear();
    escanear(fixture, '123456');
    expect(texto(fixture)).toContain('Galletas Marías no tiene existencias');
    expect(filas(fixture).length).toBe(0);
  });

  it('el botón + se desactiva al llegar a la existencia y escanear de más avisa', async () => {
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    const mas = filas(fixture)[0].querySelectorAll('.cantidad button')[1] as HTMLButtonElement;
    expect(mas.disabled).toBe(false);
    mas.click();
    fixture.detectChanges();
    expect(
      (filas(fixture)[0].querySelectorAll('.cantidad button')[1] as HTMLButtonElement).disabled,
    ).toBe(true);
    escanear(fixture, '7501055303038');
    expect(texto(fixture)).toContain('Ya agregaste todas las existencias');
    expect(filas(fixture)[0].querySelector('output')?.textContent?.trim()).toBe('2');
  });

  it('calcula subtotal, IVA y total de $64.00', async () => {
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501020565119');
    const resumen =
      (fixture.nativeElement.querySelector('.resumen-venta') as HTMLElement).textContent ?? '';
    expect(resumen).toContain('$55.17');
    expect(resumen).toContain('$8.83');
    expect(resumen).toContain('$64.00');
    expect(resumen).toContain('3 piezas');
  });

  it('quitar una línea ofrece deshacer y la regresa a su lugar', async () => {
    const fixture = await crear();
    escanear(fixture, '7501055303038');
    escanear(fixture, '7501020565119');
    (filas(fixture)[0].querySelector('.quitar') as HTMLButtonElement).click();
    fixture.detectChanges();
    expect(filas(fixture).length).toBe(1);
    const [, etiqueta, deshacer] = feedback.conAccion.mock.calls[0] as unknown as [
      string,
      string,
      () => void,
    ];
    expect(etiqueta).toBe('Deshacer');
    deshacer();
    fixture.detectChanges();
    expect(
      filas(fixture).map((fila) => fila.querySelector('.producto__nombre')?.textContent),
    ).toEqual(['Coca-Cola 600 ml', 'Leche Lala entera 1 L']);
  });

  it('el texto escrito muestra coincidencias y Enter agrega la resaltada', async () => {
    vi.useFakeTimers();
    catalogo['coca'] = { items: [galletas, coca], por_codigo: false };
    const fixture = await crear();
    const campo = fixture.nativeElement.querySelector('input[role="combobox"]') as HTMLInputElement;
    campo.value = 'coca';
    campo.dispatchEvent(new Event('input'));
    vi.advanceTimersByTime(250);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelectorAll('[role="option"]').length).toBe(2);
    expect(
      fixture.nativeElement.querySelector('.captura__opcion--resaltada')?.textContent,
    ).toContain('Coca-Cola');
    campo.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
    fixture.detectChanges();
    expect(filas(fixture).length).toBe(1);
    vi.useRealTimers();
  });

  it('Escape limpia el campo', async () => {
    const fixture = await crear();
    const campo = fixture.nativeElement.querySelector('input[role="combobox"]') as HTMLInputElement;
    campo.value = 'leche';
    campo.dispatchEvent(new Event('input'));
    campo.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
    expect(campo.value).toBe('');
  });

  it('«Verificar costos» solo aparece con permiso y alerta ventas bajo costo', async () => {
    let fixture = await crear();
    expect(fixture.nativeElement.querySelector('.interruptor')).toBeNull();

    TestBed.resetTestingModule();
    permisos = new Set([PERMISOS.ventaRegistrar, PERMISOS.ventaVerCostos]);
    fixture = await crear();
    escanear(fixture, '7501020565119');
    (fixture.nativeElement.querySelector('.interruptor input') as HTMLInputElement).click();
    fixture.detectChanges();
    expect(texto(fixture)).toContain('por debajo de su costo');
    expect(filas(fixture)[0].querySelector('.margen--negativo')).not.toBeNull();
  });

  it('recupera la lista guardada al recargar y la guarda al cambiar', async () => {
    borradores['negocio-1.sucursal-1'] = [
      {
        productoNegocioId: 'coca',
        nombre: 'Coca-Cola 600 ml',
        codigoBarras: null,
        precioCentavos: 1800,
        costoCentavos: null,
        existencia: 5,
        existenciaBaja: true,
        cantidad: 3,
      },
    ];
    const fixture = await crear();
    expect(filas(fixture).length).toBe(1);
    expect(filas(fixture)[0].querySelector('.existencia--baja')).not.toBeNull();
    escanear(fixture, '7501020565119');
    expect((borradores['negocio-1.sucursal-1'] as unknown[]).length).toBe(2);
  });
});
