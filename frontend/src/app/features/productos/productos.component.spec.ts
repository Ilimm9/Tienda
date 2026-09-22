import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { environment } from '../../../environments/environment';
import { ProductRow } from './product.models';
import { ProductosComponent } from './productos.component';

describe('ProductosComponent', () => {
  let component: ProductosComponent;
  let http: HttpTestingController;
  const router = { navigate: vi.fn().mockResolvedValue(true) };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [ProductosComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ContextoService,
          useValue: {
            negocio: signal({ id: environment.defaultBusinessId }),
            sucursal: signal(null),
          },
        },
        { provide: Router, useValue: router },
      ],
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
    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/unidades-medida`)
      .flush([{ id: 'unidad-ml', nombre: 'Mililitro', codigo: 'ml', simbolo: 'ml' }]);
    router.navigate.mockClear();
  });

  afterEach(() => http.verify());

  it('fills compatible fields and keeps them editable', () => {
    expect(component.loading()).toBe(false);

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

  it('enables automatic SKU generation for a manual product', () => {
    const controls = component.productForm.controls;
    controls.sku_interno.setValue('SKU-MANUAL');
    controls.generar_sku_interno.setValue(true);
    component.onGenerateSKUChanged();

    expect(controls.sku_interno.disabled).toBe(true);
    expect(controls.sku_interno.value).toBe('');

    controls.generar_sku_interno.setValue(false);
    component.onGenerateSKUChanged();
    expect(controls.sku_interno.enabled).toBe(true);
  });

  it('adds variant rows and disables simple-product commercial fields', () => {
    const controls = component.productForm.controls;
    controls.tiene_variantes.setValue(true);
    component.onVariantsChanged();

    expect(component.variantControls.length).toBe(1);
    expect(controls.sku_interno.disabled).toBe(true);
    expect(controls.precio_venta.disabled).toBe(true);
    expect(controls.stock_inicial.disabled).toBe(true);
  });

  it('replaces a dropped import file and clears its preview when removed', () => {
    const firstFile = new File(['primero'], 'productos-inicial.xlsx', {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const replacementFile = new File(['segundo'], 'productos-actualizado.xlsx', {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    component.importPreview.set({
      procesadas: 1,
      creadas: 0,
      skus_generados: 0,
      insertables: 1,
      omitidas: 0,
      invalidas: 0,
      productos_base_creados: 0,
      productos_base_reutilizados: 0,
      variantes_creadas: 0,
      errores: [],
      advertencias: [],
    });

    component.onImportDrop({ preventDefault: vi.fn(), dataTransfer: { files: [firstFile] } } as unknown as DragEvent);
    component.onImportDrop({ preventDefault: vi.fn(), dataTransfer: { files: [replacementFile] } } as unknown as DragEvent);

    expect(component.importFile).toBe(replacementFile);
    expect(component.importPreview()).toBeNull();

    component.removeImportFile({ stopPropagation: vi.fn() } as unknown as MouseEvent, document.createElement('input'));

    expect(component.importFile).toBeNull();
    expect(component.importResult()).toBeNull();
    expect(component.importPreview()).toBeNull();
  });

  it('navigates to the dedicated product pages instead of opening dialogs', () => {
    component.navigateToCreate();
    component.navigateToImport();
    component.navigateToEdit({ id: 'producto-1' } as ProductRow);

    expect(router.navigate).toHaveBeenNthCalledWith(1, ['/catalogo/productos/nuevo']);
    expect(router.navigate).toHaveBeenNthCalledWith(2, ['/catalogo/productos/importar']);
    expect(router.navigate).toHaveBeenNthCalledWith(3, ['/catalogo/productos', 'producto-1', 'editar']);
  });
});
