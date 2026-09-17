import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../environments/environment';
import { ContextoNegocio } from './contexto.models';
import { ContextoService } from './contexto.service';

const negocioKey = 'tienda.contexto.negocio_id';
const sucursalKey = 'tienda.contexto.sucursal_id';
const opcionesUrl = `${environment.apiUrl}/contexto/opciones`;

const negocioA = '11111111-1111-4111-8111-111111111111';
const negocioB = '33333333-3333-4333-8333-333333333333';
const sucursalPrincipal = '22222222-2222-4222-8222-222222222222';
const sucursalSecundaria = '44444444-4444-4444-8444-444444444444';

function negocio(id: string, sucursales: ContextoNegocio['sucursales']): ContextoNegocio {
  return { id, slug: `negocio-${id.slice(0, 4)}`, nombre_comercial: `Negocio ${id.slice(0, 4)}`, tipo_miembro: 'propietario', sucursales };
}

const sucursales = [
  { id: sucursalPrincipal, codigo: 'SUC-001', nombre: 'Matriz', es_principal: true },
  { id: sucursalSecundaria, codigo: 'SUC-002', nombre: 'Norte', es_principal: false },
];

describe('ContextoService', () => {
  let service: ContextoService;
  let http: HttpTestingController;

  beforeEach(() => {
    vi.stubGlobal('localStorage', storage());
    TestBed.configureTestingModule({
      providers: [ContextoService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ContextoService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    TestBed.resetTestingModule();
    vi.unstubAllGlobals();
  });

  function responder(items: ContextoNegocio[]): void {
    http.expectOne(opcionesUrl).flush({ items, total: items.length });
  }

  it('selecciona automáticamente cuando solo existe un negocio y usa la sucursal principal', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales)]);

    expect(service.estado()).toBe('listo');
    expect(service.negocio()?.id).toBe(negocioA);
    expect(service.sucursal()?.id).toBe(sucursalPrincipal);
    expect(localStorage.getItem(negocioKey)).toBe(negocioA);
    expect(localStorage.getItem(sucursalKey)).toBe(sucursalPrincipal);
  });

  it('restaura un par guardado válido', () => {
    localStorage.setItem(negocioKey, negocioB);
    localStorage.setItem(sucursalKey, sucursalSecundaria);

    service.inicializar().subscribe();
    responder([negocio(negocioA, []), negocio(negocioB, sucursales)]);

    expect(service.negocio()?.id).toBe(negocioB);
    expect(service.sucursal()?.id).toBe(sucursalSecundaria);
  });

  it('pide elegir negocio cuando hay varios y ninguno guardado es válido', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales), negocio(negocioB, sucursales)]);

    expect(service.estado()).toBe('requiere_negocio');
    expect(service.negocio()).toBeNull();
    expect(localStorage.getItem(negocioKey)).toBeNull();
  });

  it('descarta un UUID malformado guardado sin mostrar error', () => {
    localStorage.setItem(negocioKey, 'no-es-uuid');

    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales), negocio(negocioB, sucursales)]);

    expect(service.estado()).toBe('requiere_negocio');
  });

  it('descarta una sucursal ajena y cae en la principal', () => {
    localStorage.setItem(negocioKey, negocioA);
    localStorage.setItem(sucursalKey, '55555555-5555-4555-8555-555555555555');

    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales)]);

    expect(service.sucursal()?.id).toBe(sucursalPrincipal);
    expect(localStorage.getItem(sucursalKey)).toBe(sucursalPrincipal);
  });

  it('mantiene el negocio y marca sin_sucursal cuando no hay sucursales activas', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, [])]);

    expect(service.estado()).toBe('sin_sucursal');
    expect(service.negocio()?.id).toBe(negocioA);
    expect(service.sucursal()).toBeNull();
    expect(localStorage.getItem(sucursalKey)).toBeNull();
  });

  it('una cuenta sin negocios limpia el contexto', () => {
    localStorage.setItem(negocioKey, negocioA);

    service.inicializar().subscribe();
    responder([]);

    expect(service.estado()).toBe('requiere_negocio');
    expect(localStorage.getItem(negocioKey)).toBeNull();
  });

  it('cambiar de negocio limpia la sucursal anterior y toma la principal del nuevo', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales), negocio(negocioB, [])]);
    service.seleccionarNegocio(negocioA);
    expect(service.sucursal()?.id).toBe(sucursalPrincipal);

    service.seleccionarNegocio(negocioB);

    expect(service.negocio()?.id).toBe(negocioB);
    expect(service.sucursal()).toBeNull();
    expect(service.estado()).toBe('sin_sucursal');
    expect(localStorage.getItem(sucursalKey)).toBeNull();
  });

  it('ignora una sucursal que no pertenece al negocio activo', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales)]);

    service.seleccionarSucursal('55555555-5555-4555-8555-555555555555');

    expect(service.sucursal()?.id).toBe(sucursalPrincipal);
  });

  it('deduplica inicializaciones concurrentes en una sola petición', () => {
    service.inicializar().subscribe();
    service.inicializar().subscribe();

    responder([negocio(negocioA, sucursales)]);

    expect(service.estado()).toBe('listo');
  });

  it('asegurarInicializado no vuelve a pedir opciones una vez resuelto', () => {
    service.asegurarInicializado().subscribe();
    responder([negocio(negocioA, sucursales)]);

    service.asegurarInicializado().subscribe();

    http.expectNone(opcionesUrl);
    expect(service.inicializado()).toBe(true);
  });

  it('recargar fuerza una revalidación contra la API', () => {
    service.asegurarInicializado().subscribe();
    responder([negocio(negocioA, sucursales)]);

    service.recargar().subscribe();
    responder([negocio(negocioA, [])]);

    expect(service.estado()).toBe('sin_sucursal');
  });

  it('un error de red deja estado recuperable sin respaldo fijo', () => {
    service.inicializar().subscribe();
    http.expectOne(opcionesUrl).error(new ProgressEvent('error'));

    expect(service.estado()).toBe('error');
    expect(service.negocio()).toBeNull();
    expect(localStorage.getItem(negocioKey)).toBeNull();
  });

  it('limpiar borra ambas claves al cerrar sesión', () => {
    service.inicializar().subscribe();
    responder([negocio(negocioA, sucursales)]);

    service.limpiar();

    expect(localStorage.getItem(negocioKey)).toBeNull();
    expect(localStorage.getItem(sucursalKey)).toBeNull();
    expect(service.inicializado()).toBe(false);
    expect(service.estado()).toBe('requiere_negocio');
  });

  it('un cambio de otra pestaña revalida antes de adoptar el contexto', () => {
    service.asegurarInicializado().subscribe();
    responder([negocio(negocioA, sucursales), negocio(negocioB, sucursales)]);

    window.dispatchEvent(new StorageEvent('storage', { key: negocioKey, newValue: negocioB }));

    responder([negocio(negocioB, sucursales)]);
    expect(service.negocio()?.id).toBe(negocioB);
  });
});

function storage(): Storage {
  const values = new Map<string, string>();
  return {
    get length() {
      return values.size;
    },
    clear: () => values.clear(),
    getItem: (key) => values.get(key) ?? null,
    key: (index) => [...values.keys()][index] ?? null,
    removeItem: (key) => void values.delete(key),
    setItem: (key, value) => void values.set(key, value),
  };
}
