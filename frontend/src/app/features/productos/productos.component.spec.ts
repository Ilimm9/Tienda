import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { environment } from '../../../environments/environment';
import { ProductosComponent } from './productos.component';

describe('ProductosComponent', () => {
  let component: ProductosComponent;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [ProductosComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    component = TestBed.createComponent(ProductosComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);

    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/productos`)
      .flush({ items: [], total: 0 });

    component.openCreateDialog();
    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/categorias`)
      .flush([{ id: 'categoria-bebidas', nombre: 'Bebidas' }]);
    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/marcas`)
      .flush([{ id: 'marca-coca-cola', nombre: 'Coca Cola' }]);
    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/sucursales`)
      .flush([{ id: 'sucursal-1', nombre: 'Tienda prueba' }]);
    http.expectOne(`${environment.apiUrl}/catalogo/unidades-medida`)
      .flush([{ id: 'unidad-ml', nombre: 'Mililitro', codigo: 'ml', simbolo: 'ml' }]);
  });

  afterEach(() => http.verify());

  it('fills compatible fields and keeps them editable', () => {
    component.productForm.controls.codigo_barras.setValue('7501055303038');
    component.lookupProduct();

    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/productos/consulta-codigo/7501055303038`)
      .flush({
        codigo_barras: '7501055303038',
        nombre: 'Coca-Cola 600 ml',
        descripcion: 'Refresco de cola',
        marca: 'cocá cola',
        categoria: 'Bebidas',
        precio_sugerido: 18.5,
        contenido: 600,
        unidad_contenido: 'ml',
        imagen_url: 'https://cdn.example.com/coca-cola.jpg',
        fuentes: ['PrecioCheck'],
      });

    const controls = component.productForm.controls;
    expect(controls.nombre.value).toBe('Coca-Cola 600 ml');
    expect(controls.sku_interno.value).toBe('7501055303038');
    expect(controls.precio_venta.value).toBe(18.5);
    expect(controls.contenido.value).toBe(600);
    expect(controls.marca_id.value).toBe('marca-coca-cola');
    expect(controls.categoria_id.value).toBe('categoria-bebidas');
    expect(component.previewImageURL()).toBe('https://cdn.example.com/coca-cola.jpg');
  });

  it('preserves manually edited fields and leaves unmatched catalogs empty', () => {
    const controls = component.productForm.controls;
    controls.nombre.setValue('Nombre escrito por el usuario');
    controls.nombre.markAsDirty();
    controls.codigo_barras.setValue('7501234567890');
    component.lookupProduct();

    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/productos/consulta-codigo/7501234567890`)
      .flush({
        codigo_barras: '7501234567890',
        nombre: 'Nombre de PrecioCheck',
        descripcion: null,
        marca: 'Marca inexistente',
        categoria: 'Categoría inexistente',
        precio_sugerido: null,
        contenido: null,
        unidad_contenido: null,
        imagen_url: null,
        fuentes: ['UPCitemdb'],
      });

    expect(controls.nombre.value).toBe('Nombre escrito por el usuario');
    expect(controls.sku_interno.value).toBe('7501234567890');
    expect(controls.marca_id.value).toBe('');
    expect(controls.categoria_id.value).toBe('');
    expect(component.catalogLookupWarnings()).toHaveLength(2);
  });
});
