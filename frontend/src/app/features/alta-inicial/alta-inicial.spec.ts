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

  function crear(estado: EstadoAlta, negocio: { id: string; nombre_comercial: string } | null = null) {
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

  it('registra la empresa con lo mínimo y pasa a la sucursal', () => {
    negocios.crear.mockReturnValue(of({ id: 'n1', nombre_comercial: 'Abarrotes La Esperanza' }));
    const fixture = crear('requiere_empresa');
    const componente = fixture.componentInstance;
    componente.empresaForm.patchValue({ nombre_comercial: '  Abarrotes La Esperanza ' });
    alta.set('requiere_sucursal');

    componente.guardarEmpresa();

    expect(negocios.crear).toHaveBeenCalledWith({
      nombre_comercial: 'Abarrotes La Esperanza',
      codigo_moneda: 'MXN',
      zona_horaria: 'America/Mexico_City',
    });
    expect(contexto['recargar']).toHaveBeenCalledOnce();
    expect(contexto['seleccionarNegocio']).toHaveBeenCalledWith('n1');
    expect(componente.paso()).toBe('sucursal');
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
    expect(sucursales.crear).toHaveBeenCalledWith('n1', { codigo: 'CEN-01', nombre: 'Centro', es_principal: true });
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
      throwError(() => ({ error: { mensaje: 'Ya existe una sucursal con ese código.', campos: { codigo: 'ya existe' } } })),
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
