import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { ComprasService } from '../compras/compras.service';
import { EmpleadoService } from '../equipo/empleado.service';
import { ProductosService } from '../productos/productos.service';
import { HomeComponent, saludoPorHora } from './home.component';

const lista = (total: number) => of({ items: [], total });
const productos = { listByBusiness: vi.fn(() => lista(0)) };
const empleados = { listar: vi.fn(() => lista(0)) };
const compras = { list: vi.fn(() => lista(0)) };

async function crear(otorgados: string[], verNegocios = true, alta = 'listo', tipo: 'propietario' | 'miembro' | null = null) {
  await TestBed.configureTestingModule({
    imports: [HomeComponent],
    providers: [
      provideRouter([]),
      { provide: ProductosService, useValue: productos },
      { provide: EmpleadoService, useValue: empleados },
      { provide: ComprasService, useValue: compras },
      {
        provide: ContextoService,
        useValue: {
          negocio: signal(tipo ? { id: 'n1', nombre_comercial: 'Empresa THRS', tipo_miembro: tipo } : null),
          sucursal: signal(tipo ? { id: 's1', nombre: 'Sucursal Reforma', codigo: 'SUC-001' } : null),
          puede: (codigo: string) => otorgados.includes(codigo),
          puedeVerNegocios: signal(verNegocios),
          estadoAlta: signal(alta),
        },
      },
    ],
  }).compileComponents();
  const fixture = TestBed.createComponent(HomeComponent);
  fixture.detectChanges();
  return fixture;
}

describe('HomeComponent', () => {
  beforeEach(() => vi.clearAllMocks());
  afterEach(() => TestBed.resetTestingModule());

  it('saluda según la hora', () => {
    expect(saludoPorHora(4)).toBe('Buenas noches');
    expect(saludoPorHora(5)).toBe('Buenos días');
    expect(saludoPorHora(12)).toBe('Buenas tardes');
    expect(saludoPorHora(19)).toBe('Buenas noches');
  });

  it('da la bienvenida con la sucursal activa, sin accesos ni tarjeta de tips', async () => {
    const fixture = await crear([], true, 'listo', 'miembro');
    const texto = fixture.nativeElement.textContent as string;

    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Empresa THRS');
    expect(texto).toContain('Hoy estás en Sucursal Reforma.');
    expect(texto).toContain('Miembro');
    expect(fixture.nativeElement.querySelector('.resumen-tarjeta')).toBeNull();
    expect(fixture.nativeElement.querySelector('.consejo')).toBeNull();
  });

  it('muestra a la mascota con los tips de la app', async () => {
    const fixture = await crear([], true, 'listo', 'miembro');
    const boton: HTMLButtonElement = fixture.nativeElement.querySelector('app-mascota .mascota__boton');

    expect(boton.getAttribute('aria-label')).toBe('Abrir tips de Stocki');
    boton.click();
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('app-mascota h2').textContent).toContain('Sube tu Excel tal cual');
  });

  it('a un miembro no le muestra pasos ni consulta las listas', async () => {
    const fixture = await crear([], false, 'listo', 'miembro');

    expect(fixture.nativeElement.querySelector('.guia')).toBeNull();
    expect(productos.listByBusiness).not.toHaveBeenCalled();
  });

  it('calcula los pasos del propietario con datos reales', async () => {
    empleados.listar.mockReturnValueOnce(lista(2));
    const fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    const guia = fixture.nativeElement.querySelector('.guia');

    const detalle = () => guia.querySelector('.guia__detalle a').getAttribute('href');
    const tarjetas = guia.querySelectorAll('button.etapa');

    expect(productos.listByBusiness).toHaveBeenCalledWith('n1', 's1');
    expect(guia.textContent).toContain('3 de 5');
    expect(fixture.nativeElement.textContent).toContain('Te faltan 2 pasos');
    expect(guia.querySelectorAll('.etapa--hecho')).toHaveLength(3);
    // Al entrar se elige el primer pendiente.
    expect(tarjetas[2].classList).toContain('etapa--siguiente');
    expect(tarjetas[2].getAttribute('aria-pressed')).toBe('true');
    expect(detalle()).toBe('/productos/importar');

    expect(guia.querySelector('.guia__rotulo').textContent).toContain('Paso 3 de 5 · Sigue este');
    expect(guia.querySelectorAll('.guia__puntos li')).toHaveLength(3);

    tarjetas[3].click();
    fixture.detectChanges();
    expect(detalle()).toBe('/equipo/empleados');
    expect(guia.querySelector('.guia__puntos-titulo').textContent).toContain('Para tener en cuenta');
    // «Ver siguiente paso» avanza al de la derecha.
    guia.querySelector('.guia__acciones button').click();
    fixture.detectChanges();
    expect(detalle()).toBe('/compras/nueva');
    expect(guia.querySelector('.guia__acciones button')).toBeNull();
  });

  it('oculta los pasos si todo está hecho o si una consulta falla', async () => {
    productos.listByBusiness.mockReturnValueOnce(lista(1));
    empleados.listar.mockReturnValueOnce(lista(1));
    compras.list.mockReturnValueOnce(lista(1));
    let fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.guia')).toBeNull();

    TestBed.resetTestingModule();
    compras.list.mockReturnValueOnce(throwError(() => new Error('falla')));
    fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.guia')).toBeNull();
  });

  it('ofrece el asistente cuando falta la empresa', async () => {
    const fixture = await crear([], true, 'requiere_empresa');
    const texto = fixture.nativeElement.textContent as string;

    expect(texto).toContain('Falta tu empresa.');
    expect(texto).toContain('0 de 2 pasos');
    expect(fixture.nativeElement.querySelector('a.boton').getAttribute('href')).toBe('/configuracion-inicial');
  });

  it('retoma en la sucursal cuando la empresa ya existe', async () => {
    const fixture = await crear([], true, 'requiere_sucursal');
    const texto = fixture.nativeElement.textContent as string;

    expect(texto).toContain('Falta tu primera sucursal.');
    expect(texto).toContain('1 de 2 pasos');
    expect(fixture.nativeElement.querySelector('a.boton').textContent).toContain('Crear sucursal');
  });

  it('a un invitado sin sucursal le pide solicitar una asignación, sin asistente', async () => {
    const fixture = await crear([], false, 'sin_asignacion');

    expect(fixture.nativeElement.textContent).toContain('Solicita una asignación al administrador');
    expect(fixture.nativeElement.querySelector('a[href="/configuracion-inicial"]')).toBeNull();
  });
});
