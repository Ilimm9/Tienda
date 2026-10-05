import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { HomeComponent } from './home.component';

async function crear(otorgados: string[], verNegocios = true, alta = 'listo') {
  await TestBed.configureTestingModule({
    imports: [HomeComponent],
    providers: [
      provideRouter([]),
      {
        provide: ContextoService,
        useValue: {
          negocio: signal(null),
          sucursal: signal(null),
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
  afterEach(() => TestBed.resetTestingModule());

  it('renders the dashboard summary inside the administrative shell', async () => {
    const fixture = await crear(['sucursales.ver', 'equipo.empleados.ver']);

    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Bienvenido');
    const tarjetas = fixture.nativeElement.querySelectorAll('a.resumen-tarjeta');
    expect(tarjetas).toHaveLength(3);
    expect(tarjetas[2].getAttribute('href')).toBe('/equipo/empleados');
  });

  it('muestra solo las secciones que el rol permite', async () => {
    const fixture = await crear([]);

    const tarjetas = fixture.nativeElement.querySelectorAll('a.resumen-tarjeta');
    expect(tarjetas).toHaveLength(1);
    expect(tarjetas[0].getAttribute('href')).toBe('/negocios');
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
