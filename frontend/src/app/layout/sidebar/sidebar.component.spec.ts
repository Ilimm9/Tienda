import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { SidebarComponent } from './sidebar.component';

async function crear(otorgados: string[]) {
  await TestBed.configureTestingModule({
    imports: [SidebarComponent],
    providers: [
      provideRouter([]),
      {
        provide: ContextoService,
        useValue: {
          puede: (codigo?: string) => !codigo || otorgados.includes(codigo),
          puedeVerNegocios: signal(false),
        },
      },
    ],
  }).compileComponents();
  const fixture = TestBed.createComponent(SidebarComponent);
  fixture.detectChanges();
  return fixture;
}

describe('SidebarComponent', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('muestra «Nueva venta» con F2 arriba del menú a quien puede vender', async () => {
    const fixture = await crear(['ventas.registrar']);
    const nav = fixture.nativeElement.querySelector('nav') as HTMLElement;
    const boton = nav.firstElementChild as HTMLAnchorElement;
    expect(boton.classList).toContain('nueva-venta');
    expect(boton.getAttribute('href')).toBe('/ventas/nueva');
    expect(boton.getAttribute('aria-keyshortcuts')).toBe('F2');
    expect(boton.textContent).toContain('F2');
  });

  it('sin permiso de vender no muestra el botón', async () => {
    const fixture = await crear([]);
    expect(fixture.nativeElement.querySelector('.nueva-venta')).toBeNull();
  });
});
