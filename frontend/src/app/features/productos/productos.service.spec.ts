import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { ProductosService } from './productos.service';

describe('ProductosService', () => {
  let service: ProductosService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [ProductosService, provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ProductosService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('looks up a PrecioCheck image through the backend', () => {
    const businessId = 'negocio-1';
    const barcode = '7501055303038';

    service.lookupProduct(businessId, barcode).subscribe((result) => {
      expect(result.imagen_url).toBe('https://cdn.example.com/producto.jpg');
    });

    const request = http.expectOne(
      `${environment.apiUrl}/negocios/${businessId}/catalogo/productos/consulta-codigo/${barcode}`,
    );
    expect(request.request.method).toBe('GET');
    request.flush({ codigo_barras: barcode, imagen_url: 'https://cdn.example.com/producto.jpg' });
  });

  it('keeps barcode and image optional in the create request', () => {
    const payload = {
      nombre: 'Producto manual',
      sku_interno: 'PROD-001',
      marca_id: null,
      categoria_id: 'categoria-1',
      sucursal_id: 'sucursal-1',
      descripcion: null,
      contenido: null,
      unidad_contenido: null,
      presentacion: null,
      precio_venta: 10,
      stock_inicial: 0,
      codigo_barras: null,
      imagen_url: null,
    };

    service.create('negocio-1', payload).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/negocios/negocio-1/catalogo/productos`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual(payload);
    request.flush(null);
  });

  it('updates commercial product data without inventory fields', () => {
    const payload = {
      nombre: 'Producto actualizado',
      sku_interno: 'PROD-001',
      marca_id: null,
      categoria_id: 'categoria-1',
      descripcion: 'Descripción',
      contenido: 600,
      unidad_contenido: 'ml',
      unidad_medida_id: 'unidad-ml',
      presentacion: 'Botella',
      precio_venta: 18.5,
      codigo_barras: '7501055303038',
      imagen_url: 'https://cdn.example.com/producto.jpg',
    };

    service.update('negocio-1', 'producto-1', payload).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/negocios/negocio-1/catalogo/productos/producto-1`);
    expect(request.request.method).toBe('PATCH');
    expect(request.request.body).toEqual(payload);
    request.flush(null);
  });

  it('deactivates a product within its business', () => {
    service.deactivate('negocio-1', 'producto-1').subscribe();

    const request = http.expectOne(`${environment.apiUrl}/negocios/negocio-1/catalogo/productos/producto-1`);
    expect(request.request.method).toBe('DELETE');
    request.flush(null);
  });

  it('uploads the spreadsheet and selected branch for a product import', () => {
    const file = new File(['spreadsheet'], 'productos.xlsx', { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
    service.importProducts('negocio-1', 'sucursal-1', file).subscribe();

    const request = http.expectOne(`${environment.apiUrl}/negocios/negocio-1/catalogo/productos/importar`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body.get('sucursal_id')).toBe('sucursal-1');
    expect(request.request.body.get('archivo')).toBe(file);
    request.flush({ procesadas: 1, creadas: 1, omitidas: 0, invalidas: 0, errores: [], advertencias: [] });
  });

  it('validates the spreadsheet before importing it', () => {
    const file = new File(['spreadsheet'], 'productos.xlsx', { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' });
    service.previewProductImport('negocio-1', 'sucursal-1', file).subscribe((result) => {
      expect(result.insertables).toBe(2);
    });

    const request = http.expectOne(`${environment.apiUrl}/negocios/negocio-1/catalogo/productos/validar-importacion`);
    expect(request.request.method).toBe('POST');
    expect(request.request.body.get('sucursal_id')).toBe('sucursal-1');
    request.flush({ procesadas: 3, insertables: 2, creadas: 0, omitidas: 0, invalidas: 1, errores: [], advertencias: [] });
  });
});
