import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { environment } from '../../../environments/environment';
import { ProductRow } from './product.models';
import { ProductosComponent } from './productos.component';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

describe('ProductosComponent', () => {
  let component: ProductosComponent;
  let http: HttpTestingController;
  let fixture: ComponentFixture<ProductosComponent>;
  const sucursal = signal({ id: 'sucursal-1' });
  const router = { navigate: vi.fn().mockResolvedValue(true) };
  const negocioId = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';

  beforeEach(() => {
    sucursal.set({ id: 'sucursal-1' });
    TestBed.configureTestingModule({
      imports: [ProductosComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ContextoService,
          useValue: {
            negocio: signal({ id: negocioId }),
            sucursal,
            puede: () => true,
          },
        },
        { provide: Router, useValue: router },
      ],
    });
    fixture = TestBed.createComponent(ProductosComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
    http = TestBed.inject(HttpTestingController);

    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`)
      .flush({ items: [], total: 0 });
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`)
      .flush([]);

    component.openCreateDialog();
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/categorias`)
      .flush([{ id: 'categoria-bebidas', nombre: 'Bebidas' }]);
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`)
      .flush([{ id: 'marca-coca-cola', nombre: 'Coca Cola' }]);
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`)
      .flush([]);
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/sucursales`)
      .flush([{ id: 'sucursal-1', nombre: 'Tienda prueba' }]);
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/unidades-medida`)
      .flush([{ id: 'unidad-ml', nombre: 'Mililitro', simbolo: 'ml' }]);
    router.navigate.mockClear();
  });

  afterEach(() => http.verify());

  it('descarta respuestas de la sucursal anterior y limpia al perder contexto', () => {
    component.loadProducts();
    const anterior = http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);
    sucursal.set({ id: 'sucursal-2' });
    fixture.detectChanges();
    const actual = http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-2`);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);
    expect(component.products()).toEqual([]);
    actual.flush({ items: [{ id: 'actual', nombre: 'Sucursal actual', stock: 2 }], total: 1 });
    anterior.flush({ items: [{ id: 'anterior', nombre: 'Sucursal anterior', stock: 99 }], total: 1 });
    expect(component.products().map((p) => p.id)).toEqual(['actual']);
    sucursal.set({ id: '' });
    fixture.detectChanges();
    expect(component.products()).toEqual([]);
    expect(component.loading()).toBe(false);
  });

  it('muestra la lista guardada al volver y la refresca en segundo plano', () => {
    const url = `${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`;
    component.loadProducts();
    http.expectOne(url).flush({ items: [{ id: 'guardado', nombre: 'Guardado', stock: 1, variantes: [] }], total: 1 });
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);

    const otra = TestBed.createComponent(ProductosComponent);
    otra.detectChanges();
    expect(otra.componentInstance.loading()).toBe(false);
    expect(otra.componentInstance.products().map((p) => p.id)).toEqual(['guardado']);
    const refresco = http.expectOne(url);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);
    refresco.flush({ items: [{ id: 'fresco', nombre: 'Fresco', stock: 2, variantes: [] }], total: 1 });
    expect(otra.componentInstance.products().map((p) => p.id)).toEqual(['fresco']);
  });

  it('conserva la lista guardada si falla el refresco en segundo plano', () => {
    const url = `${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`;
    component.loadProducts();
    http.expectOne(url).flush({ items: [{ id: 'guardado', nombre: 'Guardado', stock: 1, variantes: [] }], total: 1 });
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);

    component.loadProducts();
    http.expectOne(url).flush('error', { status: 500, statusText: 'Server Error' });
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`).flush([]);
    expect(component.products().map((p) => p.id)).toEqual(['guardado']);
    expect(component.error()).toBeNull();
    expect(component.loading()).toBe(false);
  });

  it('no recarga cuando el contexto se revalida con la misma sucursal', () => {
    sucursal.set({ id: 'sucursal-1' });
    fixture.detectChanges();
    http.expectNone(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`);
  });

  it('fills compatible fields and keeps them editable', () => {
    expect(component.loading()).toBe(false);

    component.productForm.controls.codigo_barras.setValue('7501055303038');
    component.lookupProduct();

    http
      .expectOne(
        `${environment.apiUrl}/negocios/${negocioId}/catalogo/productos/consulta-codigo/7501055303038`,
      )
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
    expect(controls.sku_interno.value).toBe('');
    expect(controls.precio_venta.value).toBe(18.5);
    expect(controls.contenido.value).toBe(600);
    expect(controls.marca_id.value).toBe('marca-coca-cola');
    expect(controls.categoria_id.value).toBe('categoria-bebidas');
    expect(component.previewImageURL()).toBe('https://cdn.example.com/coca-cola.jpg');
  });

  it('offers and creates a brand without losing the product form', () => {
    const controls = component.productForm.controls;
    controls.nombre.setValue('Refresco de naranja');
    controls.marca_id.setValue('marca-coca-cola');

    component.onBrandSelected('__new_brand__');

    expect(component.brandOptions().at(-1)).toEqual({
      id: '__new_brand__',
      nombre: 'Crear nueva…',
    });
    expect(component.brandCreationVisible()).toBe(true);
    expect(controls.marca_id.value).toBe('');
    expect(controls.nombre.value).toBe('Refresco de naranja');

    component.brandForm.controls.nombre.setValue('Naranjax');
    component.createBrand();

    http
      .expectOne({
        method: 'POST',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`,
      })
      .flush(null);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`).flush([
      { id: 'marca-coca-cola', nombre: 'Coca Cola' },
      { id: 'marca-naranjax', nombre: 'Naranjax' },
    ]);

    expect(controls.marca_id.value).toBe('marca-naranjax');
    expect(component.brandCreationVisible()).toBe(false);
  });

  it('validates, cancels, and reports errors while creating a brand', () => {
    component.onBrandSelected('__new_brand__');
    component.createBrand();

    expect(component.brandForm.controls.nombre.touched).toBe(true);
    expect(component.brandCreating()).toBe(false);

    component.cancelBrandCreation();
    expect(component.brandCreationVisible()).toBe(false);

    component.onBrandSelected('__new_brand__');
    component.brandForm.controls.nombre.setValue('Duplicada');
    component.createBrand();
    http
      .expectOne({
        method: 'POST',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`,
      })
      .flush({ mensaje: 'La marca ya existe.' }, { status: 400, statusText: 'Bad Request' });

    expect(component.brandCreationVisible()).toBe(true);
    expect(component.brandCreationError()).toBe('La marca ya existe.');
  });

  it('preserves manually edited fields and leaves unmatched catalogs empty', () => {
    const controls = component.productForm.controls;
    controls.nombre.setValue('Nombre escrito por el usuario');
    controls.nombre.markAsDirty();
    controls.codigo_barras.setValue('7501234567890');
    component.lookupProduct();

    http
      .expectOne(
        `${environment.apiUrl}/negocios/${negocioId}/catalogo/productos/consulta-codigo/7501234567890`,
      )
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
    expect(controls.sku_interno.value).toBe('');
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

  it('keeps the base price editable when variants are enabled', () => {
    const controls = component.productForm.controls;
    controls.tiene_variantes.setValue(true);
    component.onVariantsChanged();

    expect(component.variantControls.length).toBe(0);
    expect(controls.sku_interno.disabled).toBe(true);
    expect(controls.precio_venta.enabled).toBe(true);
    expect(controls.stock_inicial.disabled).toBe(true);
  });

  it('only enables saving an edited product after a real change', () => {
    component.editingProduct.set({ id: 'producto-1' } as ProductRow);
    component.productForm.markAsPristine();

    expect(component.canSubmitProduct()).toBe(false);

    component.productForm.controls.nombre.setValue('Nombre actualizado');
    component.productForm.controls.nombre.markAsDirty();

    expect(component.canSubmitProduct()).toBe(true);
  });

  it('leaves a simple product conversion empty when variants are enabled', () => {
    component.editingProduct.set({
      id: 'producto-1',
      nombre: 'Café',
      presentacion: '500 g',
      precio: 85,
      stock: 4,
      variantes: [],
    } as unknown as ProductRow);
    component.productForm.controls.tiene_variantes.setValue(true);
    component.onVariantsChanged();

    expect(component.variantControls.length).toBe(0);
    expect(component.variantDimensions()).toEqual([]);
    expect(component.productForm.controls.codigo_barras.value).toBe('');
    expect(component.canSubmitProduct()).toBe(false);
  });

  it('keeps stock editable for variants generated after stock editing is enabled', () => {
    component.editingProduct.set({ id: 'familia-1', variantes: [{}] } as unknown as ProductRow);
    component.productForm.controls.tiene_variantes.setValue(true);
    component.stockEditing.set(true);
    component.variantDimensions.set([{ nombre: 'Talla', valores: ['M'], nuevoValor: '' }]);

    component.generateCombinations();

    expect(component.variantControls.length).toBe(1);
    expect(component.variantControls.at(0).controls.stock_inicial.enabled).toBe(true);
  });

  it('does not allow saving variants with duplicate combinations', () => {
    component.productForm.controls.tiene_variantes.setValue(true);
    component.variantDimensions.set([{ nombre: 'Color', valores: ['Rojo'], nuevoValor: '' }]);
    component.generateCombinations();
    component.generateCombinations();

    expect(component.variantControls.length).toBe(1);
    expect(component.canSubmitProduct()).toBe(true);

    const duplicate = component['createVariantForm']();
    duplicate.controls.atributos.setValue('Color=Rojo');
    component.variantControls.push(duplicate);

    expect(component.canSubmitProduct()).toBe(false);
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
      productos_simples_creados: 0,
      productos_base_creados: 0,
      productos_base_reutilizados: 0,
      variantes_creadas: 0,
      costos_faltantes: 0,
      catalogos_nuevos: [],
      errores: [],
      advertencias: [],
    });

    component.onImportDrop({
      preventDefault: vi.fn(),
      dataTransfer: { files: [firstFile] },
    } as unknown as DragEvent);
    component.onImportDrop({
      preventDefault: vi.fn(),
      dataTransfer: { files: [replacementFile] },
    } as unknown as DragEvent);

    expect(component.importFile).toBe(replacementFile);
    expect(component.importPreview()).toBeNull();

    component.removeImportFile(
      { stopPropagation: vi.fn() } as unknown as MouseEvent,
      document.createElement('input'),
    );

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
    expect(router.navigate).toHaveBeenNthCalledWith(3, [
      '/catalogo/productos',
      'producto-1',
      'editar',
    ]);
  });

  it('returns to the product table after confirming a completed import summary', async () => {
    const fire = vi.spyOn(Swal, 'fire').mockResolvedValue({ isConfirmed: true } as never);

    (component as any).pollProductImport('importacion-1');

    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos/importaciones/importacion-1`)
      .flush({
        id: 'importacion-1',
        estado: 'completada',
        etapa: 'Importación terminada',
        porcentaje: 100,
        resultado: {
          procesadas: 2,
          creadas: 2,
          omitidas: 0,
          invalidas: 0,
          productos_simples_creados: 2,
          productos_base_creados: 0,
          productos_base_reutilizados: 0,
          variantes_creadas: 0,
          costos_faltantes: 0,
          errores: [],
          advertencias: [],
        },
      });
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/productos?sucursal_id=sucursal-1`)
      .flush({ items: [], total: 0 });
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/proveedores`)
      .flush([]);

    await Promise.resolve();

    expect(fire).toHaveBeenCalledWith(
      expect.objectContaining({
        title: 'Importación terminada',
        confirmButtonText: 'Ver registros',
      }),
    );
    expect(router.navigate).toHaveBeenCalledWith(['/catalogo/productos']);
  });

  it('filters a product base by attributes and identifiers of its variants', () => {
    component.products.set([
      {
        id: 'producto-1',
        nombre: 'Base de michelada',
        imagen_url: null,
        sku: null,
        precio: 0,
        stock: 0,
        categoria: 'Botanas',
        categoria_id: 'categoria-1',
        marca: null,
        marca_id: null,
        proveedor: 'Distribuidora Norte',
        proveedor_id: 'proveedor-1',
        descripcion: null,
        presentacion: '1 litro',
        contenido: null,
        unidad_contenido: null,
        unidad_medida: 'Unidades',
        unidad_medida_id: 'unidad-1',
        codigo_barras: null,
        inventario: [],
        estado: 'En stock',
        variantes: [
          {
            id: 'variante-1',
            sku: null,
            precio: 150,
            stock: 4,
            codigo_barras: '7503030374392',
            atributos: [{ nombre: 'Sabor', valor: 'Mango' }],
          },
        ],
      },
    ]);

    component.setProductSearch('mángo');
    expect(component.filteredProducts()).toHaveLength(1);

    component.setProductSearch('7503030374392');
    expect(component.filteredProducts()).toHaveLength(1);

    component.setProductSearch('piña');
    expect(component.filteredProducts()).toHaveLength(0);

    component.setProductSearch('');
    component.setProductProviderFilter('proveedor-1');
    expect(component.filteredProducts()).toHaveLength(1);

    component.setProductProviderFilter('__sin_proveedor__');
    expect(component.filteredProducts()).toHaveLength(0);
  });

  it('keeps optional product filters empty until a value is selected', () => {
    expect(component.productCategoryFilter()).toBeNull();
    expect(component.productBrandFilter()).toBeNull();
    expect(component.productProviderFilter()).toBeNull();
    expect(component.productStatusFilter()).toBeNull();

    component.setProductCategoryFilter('categoria-1');
    component.setProductBrandFilter('marca-1');
    component.setProductStatusFilter('En stock');

    expect(component.productCategoryFilter()).toBe('categoria-1');
    expect(component.productBrandFilter()).toBe('marca-1');
    expect(component.productStatusFilter()).toBe('En stock');

    component.clearProductFilters();

    expect(component.productCategoryFilter()).toBeNull();
    expect(component.productBrandFilter()).toBeNull();
    expect(component.productStatusFilter()).toBeNull();
  });
});
