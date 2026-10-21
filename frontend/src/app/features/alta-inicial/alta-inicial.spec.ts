import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { EstadoAlta } from '../../contexto/contexto.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalService } from '../sucursales/sucursal.service';
import { AltaInicial } from './alta-inicial';

describe('AltaInicial', () => {
  const negocios = {
    crear: vi.fn(),
    restaurar: vi.fn(),
    listar: vi.fn((): unknown => of({ items: [], total: 0 })),
  };
  const sucursales = { crear: vi.fn() };
  let alta: ReturnType<typeof signal<EstadoAlta>>;
  let contexto: Record<string, unknown>;

  function crear(
    estado: EstadoAlta,
    negocio: { id: string; nombre_comercial: string } | null = null,
  ) {
    alta = signal<EstadoAlta>(estado);
    contexto = {
      estadoAlta: alta,
      negocio: signal(negocio),
      puede: () => true,
      recargar: vi.fn(() => of('listo')),
      seleccionarNegocio: vi.fn(),
      seleccionarSucursal: vi.fn(),
      posponerAlta: vi.fn(),
    };
    TestBed.configureTestingModule({
      imports: [AltaInicial],
      providers: [
        provideRouter([]),
        { provide: NegocioService, useValue: negocios },
        { provide: SucursalService, useValue: sucursales },
        { provide: ContextoService, useValue: contexto },
      ],
    });
    const fixture = TestBed.createComponent(AltaInicial);
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => vi.clearAllMocks());
  afterEach(() => TestBed.resetTestingModule());

  it('no registra la empresa sin nombre comercial', () => {
    const fixture = crear('requiere_empresa');
    fixture.componentInstance.guardarEmpresa();

    expect(negocios.crear).not.toHaveBeenCalled();
    expect(fixture.componentInstance.invalidoEmpresa('nombre_comercial')).toBe(true);
  });

  it('usa menus PrimeNG para moneda y zona horaria', () => {
    const fixture = crear('requiere_empresa');
    const selects = fixture.nativeElement.querySelectorAll('p-select');

    expect(selects).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('#codigo_moneda select, #zona_horaria select')).toHaveLength(0);
  });

  it('avanza a la sucursal sin guardar todavía la empresa', () => {
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;
    componente.empresaForm.patchValue({ nombre_comercial: '  Abarrotes La Esperanza ' });

    componente.guardarEmpresa();

    expect(negocios.crear).not.toHaveBeenCalled();
    expect(contexto['recargar']).not.toHaveBeenCalled();
    expect(componente.paso()).toBe('sucursal');
  });

  it('conserva ambos formularios al volver a la empresa', () => {
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;
    componente.empresaForm.patchValue({ nombre_comercial: 'Abarrotes', codigo_moneda: 'USD' });
    componente.guardarEmpresa();
    componente.sucursalForm.patchValue({ nombre: 'Centro', codigo: 'CEN-01' });

    componente.irPaso('empresa');

    expect(componente.paso()).toBe('empresa');
    expect(componente.empresaForm.getRawValue().nombre_comercial).toBe('Abarrotes');
    expect(componente.empresaForm.getRawValue().codigo_moneda).toBe('USD');
    expect(componente.sucursalForm.getRawValue().nombre).toBe('Centro');
  });

  it('usa las tres cards del lateral para navegar los pasos', () => {
    const fixture = crear('requiere_empresa');
    const cards = fixture.nativeElement.querySelectorAll('.vista-previa__card');

    expect(cards).toHaveLength(3);
    expect(cards[0].getAttribute('data-relativo')).toBe('0');
    expect(cards[1].getAttribute('data-relativo')).toBe('1');
    expect(cards[2].getAttribute('data-relativo')).toBe('2');
    expect(fixture.nativeElement.querySelectorAll('.pi-arrow-right')).toHaveLength(0);
    (cards[1] as HTMLButtonElement).click();
    fixture.detectChanges();

    expect(fixture.componentInstance.paso()).toBe('sucursal');
    expect(cards[0].getAttribute('data-relativo')).toBe('-1');
    expect(cards[1].getAttribute('data-relativo')).toBe('0');
    expect(cards[2].getAttribute('data-relativo')).toBe('1');
    expect((cards[2] as HTMLButtonElement).disabled).toBe(false);
  });

  it('recorre las tres cards y deja las dos anteriores arriba al finalizar', () => {
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;
    const cards = fixture.nativeElement.querySelectorAll('.vista-previa__card');

    componente.paso.set('listo');
    fixture.detectChanges();

    expect(cards[0].getAttribute('data-relativo')).toBe('-2');
    expect(cards[1].getAttribute('data-relativo')).toBe('-1');
    expect(cards[2].getAttribute('data-relativo')).toBe('0');
  });

  it('mantiene la ilustración de tienda durante todos los pasos', () => {
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;

    expect(
      fixture.nativeElement.querySelector('.vista-previa__ilustracion--tienda'),
    ).not.toBeNull();
    expect(
      fixture.nativeElement.querySelector('.vista-previa__ilustracion--inventario'),
    ).toBeNull();

    componente.paso.set('listo');
    fixture.detectChanges();

    expect(
      fixture.nativeElement.querySelector('.vista-previa__ilustracion--tienda'),
    ).not.toBeNull();
    expect(
      fixture.nativeElement.querySelector('.vista-previa__ilustracion--inventario'),
    ).toBeNull();
  });

  it('mantiene los puntos intermedios de la transición sobre la circunferencia', () => {
    const componente = crear('requiere_empresa').componentInstance;
    const puntoOrbital = (
      componente as unknown as {
        puntoOrbital: (
          angulo: number,
          geometria: { x: number; y: number; radio: number },
        ) => {
          x: number;
          y: number;
        };
      }
    ).puntoOrbital;

    [90, 112.5, 135, 157.5, 180, 202.5, 225, 247.5, 270].forEach((angulo) => {
      const punto = puntoOrbital(angulo, { x: 100, y: 100, radio: 50 });
      expect(Math.hypot(punto.x - 100, punto.y - 100)).toBeCloseTo(50, 5);
    });
  });

  it('crea empresa y sucursal juntas al confirmar el segundo paso', () => {
    negocios.crear.mockReturnValue(of({ id: 'n1', nombre_comercial: 'Abarrotes' }));
    sucursales.crear.mockReturnValue(of({ id: 's1', nombre: 'Centro', codigo: 'CEN-01' }));
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;
    componente.empresaForm.patchValue({ nombre_comercial: 'Abarrotes' });
    componente.guardarEmpresa();
    componente.sucursalForm.patchValue({ nombre: 'Centro', codigo: 'cen-01' });

    componente.crearSucursal();

    expect(negocios.crear).toHaveBeenCalledOnce();
    expect(sucursales.crear).toHaveBeenCalledWith('n1', {
      codigo: 'CEN-01',
      nombre: 'Centro',
      es_principal: true,
    });
    expect(componente.paso()).toBe('listo');
  });

  it('no reabre pasos después de terminar el alta', () => {
    sucursales.crear.mockReturnValue(of({ id: 's1', nombre: 'Centro', codigo: 'CEN-01' }));
    const fixture = crear('requiere_sucursal', { id: 'n1', nombre_comercial: 'Abarrotes' });
    const componente = fixture.componentInstance;
    componente.sucursalForm.patchValue({ nombre: 'Centro', codigo: 'CEN-01' });
    componente.crearSucursal();

    componente.irPaso('empresa');

    expect(componente.paso()).toBe('listo');
  });

  it('retoma en la sucursal si la empresa ya existe y no la duplica', () => {
    sucursales.crear.mockReturnValue(of({ id: 's1', nombre: 'Centro', codigo: 'CEN-01' }));
    const fixture = crear('requiere_sucursal', { id: 'n1', nombre_comercial: 'Abarrotes' });
    const componente = fixture.componentInstance;

    expect(componente.paso()).toBe('sucursal');
    expect(negocios.listar).not.toHaveBeenCalled();

    componente.sucursalForm.patchValue({ nombre: 'Centro', codigo: 'cen-01' });
    componente.crearSucursal();

    expect(negocios.crear).not.toHaveBeenCalled();
    expect(sucursales.crear).toHaveBeenCalledWith('n1', {
      codigo: 'CEN-01',
      nombre: 'Centro',
      es_principal: true,
    });
    expect(contexto['seleccionarSucursal']).toHaveBeenCalledWith('s1');
    expect(componente.paso()).toBe('listo');
  });

  it('rechaza un código con espacios antes de enviar', () => {
    const fixture = crear('requiere_sucursal', { id: 'n1', nombre_comercial: 'Abarrotes' });
    const componente = fixture.componentInstance;
    componente.sucursalForm.patchValue({ nombre: 'Centro', codigo: 'CEN 01' });

    componente.crearSucursal();

    expect(sucursales.crear).not.toHaveBeenCalled();
    expect(componente.invalidoSucursal('codigo')).toBe(true);
  });

  it('muestra el error del servidor junto a su campo y permite reintentar', () => {
    sucursales.crear.mockReturnValue(
      throwError(() => ({
        error: {
          mensaje: 'Ya existe una sucursal con ese código.',
          campos: { codigo: 'ya existe' },
        },
      })),
    );
    const fixture = crear('requiere_sucursal', { id: 'n1', nombre_comercial: 'Abarrotes' });
    const componente = fixture.componentInstance;
    componente.sucursalForm.patchValue({ nombre: 'Centro' });

    componente.crearSucursal();

    expect(componente.error()).toBe('Ya existe una sucursal con ese código.');
    expect(componente.invalidoSucursal('codigo')).toBe(true);
    expect(componente.guardando()).toBe(false);
    expect(componente.paso()).toBe('sucursal');
  });

  it('ofrece restaurar una empresa propia archivada', () => {
    negocios.listar.mockReturnValueOnce(
      of({
        items: [
          { id: 'a1', nombre_comercial: 'Vieja', tipo_miembro: 'propietario' },
          { id: 'a2', nombre_comercial: 'Ajena', tipo_miembro: 'miembro' },
        ],
        total: 2,
      }),
    );
    const fixture = crear('requiere_empresa');

    expect(fixture.componentInstance.archivadas().map((item) => item.id)).toEqual(['a1']);
    expect(fixture.nativeElement.textContent).toContain('Restaurar Vieja');
  });

  it('«hacerlo después» pospone y vuelve a Inicio', () => {
    const fixture = crear('requiere_empresa');
    const navegar = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);

    fixture.componentInstance.posponer();

    expect(contexto['posponerAlta']).toHaveBeenCalledOnce();
    expect(navegar).toHaveBeenCalledWith(['/inicio']);
  });
});
