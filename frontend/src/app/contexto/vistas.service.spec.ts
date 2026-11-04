import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../environments/environment';
import { ContextoService } from './contexto.service';
import { VistasService } from './vistas.service';

const negocioA = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
const negocioB = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';
const url = (id: string) => `${environment.apiUrl}/negocios/${id}/configuracion/vistas`;

describe('VistasService', () => {
  let http: HttpTestingController;
  const negocio = signal<{ id: string } | null>(null);

  beforeEach(() => {
    negocio.set(null);
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), { provide: ContextoService, useValue: { negocio } }],
    });
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
  });

  function crear(): VistasService {
    const service = TestBed.inject(VistasService);
    TestBed.tick();
    return service;
  }

  it('sin negocio muestra todo y no consulta la API', () => {
    const service = crear();
    expect(service.visible('categorias')).toBe(true);
    expect(service.visible(undefined)).toBe(true);
  });

  it('carga las vistas del negocio activo y las vuelve a pedir al cambiar de negocio', () => {
    negocio.set({ id: negocioA });
    const service = crear();
    http.expectOne(url(negocioA)).flush({ secciones_ocultas: ['categorias'], secciones_ocultables: ['categorias', 'marcas'] });
    expect(service.visible('categorias')).toBe(false);
    expect(service.visible('marcas')).toBe(true);

    negocio.set({ id: negocioB });
    TestBed.tick();
    // Mientras llega la respuesta del otro negocio no se aplican las vistas del anterior.
    expect(service.visible('categorias')).toBe(true);
    http.expectOne(url(negocioB)).flush({ secciones_ocultas: ['marcas'], secciones_ocultables: ['categorias', 'marcas'] });
    expect(service.visible('marcas')).toBe(false);
    expect(service.visible('categorias')).toBe(true);
  });

  it('si la API falla, muestra todo', () => {
    negocio.set({ id: negocioA });
    const service = crear();
    http.expectOne(url(negocioA)).flush({ mensaje: 'error' }, { status: 500, statusText: 'Error' });
    expect(service.visible('categorias')).toBe(true);
  });

  it('guarda con PUT y aplica la respuesta sin recargar', () => {
    negocio.set({ id: negocioA });
    const service = crear();
    http.expectOne(url(negocioA)).flush({ secciones_ocultas: [], secciones_ocultables: ['categorias', 'marcas'] });
    let guardadas: string[] = [];
    service.guardar(['categorias', 'marcas']).subscribe((vistas) => (guardadas = vistas.secciones_ocultas));
    const peticion = http.expectOne(url(negocioA));
    expect(peticion.request.method).toBe('PUT');
    expect(peticion.request.body).toEqual({ secciones_ocultas: ['categorias', 'marcas'] });
    peticion.flush({ secciones_ocultas: ['categorias', 'marcas'], secciones_ocultables: ['categorias', 'marcas'] });
    expect(guardadas).toEqual(['categorias', 'marcas']);
    expect(service.visible('categorias')).toBe(false);
    expect(service.visible('marcas')).toBe(false);
  });

  it('asegurar espera la respuesta en curso sin repetir la petición', () => {
    negocio.set({ id: negocioA });
    const service = crear();
    let listo = false;
    service.asegurar().subscribe(() => (listo = true));
    expect(listo).toBe(false);
    http.expectOne(url(negocioA)).flush({ secciones_ocultas: ['marcas'], secciones_ocultables: ['categorias', 'marcas'] });
    expect(listo).toBe(true);
    service.asegurar().subscribe();
    http.expectNone(url(negocioA));
  });
});
