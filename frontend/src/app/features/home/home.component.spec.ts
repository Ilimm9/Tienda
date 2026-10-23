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

  it('da la bienvenida con la sucursal activa y los accesos permitidos', async () => {
    const fixture = await crear(['catalogo.ver', 'compras.ver', 'sucursales.ver', 'equipo.empleados.ver'], true, 'listo', 'miembro');
    const texto = fixture.nativeElement.textContent as string;

    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Qué gusto verte de nuevo.');
    expect(texto).toContain('Sucursal Reforma');
    expect(texto).toContain('Miembro');
    const rutas = [...fixture.nativeElement.querySelectorAll('a.resumen-tarjeta')].map((a: Element) => a.getAttribute('href'));
    expect(rutas).toEqual(['/productos', '/compras', '/sucursales', '/equipo/empleados', '/negocios']);
  });

  it('muestra solo las secciones que el rol permite', async () => {
    const fixture = await crear([]);

    const tarjetas = fixture.nativeElement.querySelectorAll('a.resumen-tarjeta');
    expect(tarjetas).toHaveLength(1);
    expect(tarjetas[0].getAttribute('href')).toBe('/negocios');
  });

  it('a un miembro no le muestra pasos ni consulta las listas', async () => {
    const fixture = await crear([], false, 'listo', 'miembro');

    expect(fixture.nativeElement.querySelector('.pasos')).toBeNull();
    expect(productos.listByBusiness).not.toHaveBeenCalled();
  });

  it('calcula los pasos del propietario con datos reales', async () => {
    empleados.listar.mockReturnValueOnce(lista(2));
    const fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    const pasos = fixture.nativeElement.querySelector('.pasos');

    expect(productos.listByBusiness).toHaveBeenCalledWith('n1', 's1');
    expect(pasos.textContent).toContain('3 de 5 listos');
    expect(pasos.querySelector('a[href="/productos/importar"]')).not.toBeNull();
    expect(pasos.querySelector('a[href="/equipo/empleados/nuevo"]')).toBeNull();
    expect(pasos.querySelector('a[href="/compras/nueva"]')).not.toBeNull();
  });

  it('oculta los pasos si todo está hecho o si una consulta falla', async () => {
    productos.listByBusiness.mockReturnValueOnce(lista(1));
    empleados.listar.mockReturnValueOnce(lista(1));
    compras.list.mockReturnValueOnce(lista(1));
    let fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.pasos')).toBeNull();

    TestBed.resetTestingModule();
    compras.list.mockReturnValueOnce(throwError(() => new Error('falla')));
    fixture = await crear([], true, 'listo', 'propietario');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.pasos')).toBeNull();
  });

  it('recorre los consejos en ciclo', async () => {
    const fixture = await crear([]);
    const titulo = () => fixture.nativeElement.querySelector('.consejo h2').textContent;
    const [anterior, siguiente] = fixture.nativeElement.querySelectorAll('.consejo__boton');

    expect(titulo()).toContain('Carga productos sin preparar catálogos');
    anterior.click();
    fixture.detectChanges();
    expect(titulo()).toContain('Cada sucursal tiene su precio');
    siguiente.click();
    siguiente.click();
    fixture.detectChanges();
    expect(titulo()).toContain('El código de barras es opcional');
  });

  it('oculta Negocios a quien solo fue invitado', async () => {
    const fixture = await crear([], false);

    expect(fixture.nativeElement.querySelectorAll('a.resumen-tarjeta')).toHaveLength(0);
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
