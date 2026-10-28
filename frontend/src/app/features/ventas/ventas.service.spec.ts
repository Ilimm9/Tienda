import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { VentasService } from './ventas.service';

describe('VentasService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    localStorage.clear();
  });

  it('registra la venta en el negocio activo', () => {
    const service = TestBed.inject(VentasService);
    const http = TestBed.inject(HttpTestingController);
    const request = {
      sucursal_id: 's1',
      forma_pago: 'EFECTIVO' as const,
      recibido: '100.00',
      lineas: [{ producto_negocio_id: 'p1', cantidad: 2 }],
    };
    service.registrar('n1', request).subscribe();
    const peticion = http.expectOne('/api/v1/negocios/n1/ventas');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual(request);
    peticion.flush({});
    http.verify();
  });

  it('guarda el borrador por sucursal, lo borra vacío y descarta datos dañados', () => {
    const service = TestBed.inject(VentasService);
    const linea = {
      productoNegocioId: 'p1',
      nombre: 'Coca',
      codigoBarras: null,
      precioCentavos: 1800,
      costoCentavos: null,
      existencia: 3,
      existenciaBaja: false,
      cantidad: 1,
    };
    service.guardarBorrador('n1', 's1', [linea]);
    expect(service.leerBorrador('n1', 's1')).toEqual([linea]);
    expect(service.leerBorrador('n1', 's2')).toEqual([]);
    service.guardarBorrador('n1', 's1', []);
    expect(localStorage.getItem('tienda.venta.borrador.n1.s1')).toBeNull();
    localStorage.setItem('tienda.venta.borrador.n1.s1', '{dañado');
    expect(service.leerBorrador('n1', 's1')).toEqual([]);
  });
});
