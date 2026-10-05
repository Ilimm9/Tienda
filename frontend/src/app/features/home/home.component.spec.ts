import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { HomeComponent } from './home.component';

async function crear(otorgados: string[]) {
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
});
