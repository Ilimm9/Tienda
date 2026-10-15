import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { environment } from '../../../environments/environment';
import { CatalogoComponent } from './catalogo.component';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import Swal from 'sweetalert2/dist/sweetalert2.esm.js';

describe('CatalogoComponent', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;
  const router = { navigate: vi.fn().mockResolvedValue(true) };
  const negocioId = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [CatalogoComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { data: { section: 'marcas' } } },
        },
        {
          provide: ContextoService,
          useValue: { negocio: signal({ id: negocioId }), puede: () => true },
        },
        { provide: Router, useValue: router },
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`).flush([]);
    router.navigate.mockClear();
  });
  afterEach(() => http.verify());

  it('shows a save error immediately after a failed edit request', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.form.controls.nombre.setValue('Marca actualizada');

    component.save();

    http
      .expectOne({
        method: 'PATCH',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/marca-1`,
      })
      .flush({ mensaje: 'La marca ya existe.' }, { status: 400, statusText: 'Bad Request' });

    expect(component.saving()).toBe(false);
    expect(component.formError()).toBe('La marca ya existe.');
  });

  it('shows a connection error when the browser cannot provide an API message', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.save();

    http
      .expectOne({
        method: 'PATCH',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/marca-1`,
      })
      .error(new ProgressEvent('error'));

    expect(component.formError()).toBe(
      'No fue posible comunicarse con el servidor. Intenta de nuevo.',
    );
  });

  it('exposes unit types as PrimeNG select label/value options', () => {
    expect(component.unitTypeOptions).toEqual([
      { label: 'Peso', value: 'PESO' },
      { label: 'Volumen', value: 'VOLUMEN' },
      { label: 'Longitud', value: 'LONGITUD' },
      { label: 'Área', value: 'AREA' },
      { label: 'Cantidad', value: 'CANTIDAD' },
      { label: 'Empaque', value: 'EMPAQUE' },
    ]);
  });

  it('uses explicit singular titles instead of trimming the plural label', () => {
    const singularTitle = (section: string): string => {
      const subject = Object.create(CatalogoComponent.prototype) as CatalogoComponent;
      Object.defineProperty(subject, 'section', { value: section });
      return subject.singularTitle;
    };

    expect(singularTitle('marcas')).toBe('Marca');
    expect(singularTitle('categorias')).toBe('Categoría');
    expect(singularTitle('proveedores')).toBe('Proveedor');
    expect(singularTitle('unidades')).toBe('Unidad de medida');
  });

  it('maps parent categories to select options and excludes the edited category', () => {
    component.parents.set([
      { id: 'categoria-1', nombre: 'Bebidas', activo: true },
      { id: 'categoria-2', nombre: 'Refrescos', activo: true },
    ]);
    component.editingId = 'categoria-2';

    expect(component.parentCategoryOptions).toEqual([{ label: 'Bebidas', value: 'categoria-1' }]);
  });

  it('marks a dropped file as ready and clears it from the remove control', () => {
    const file = new File(['contenido'], 'marcas.xlsx', {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const preventDefault = vi.fn();
    component.importError.set('Error anterior');
    component.importResult.set({
      procesadas: 1,
      creadas: 1,
      omitidas: 0,
      invalidas: 0,
      errores: [],
    });

    component.onImportDrop({
      preventDefault,
      dataTransfer: { files: [file] },
    } as unknown as DragEvent);

    expect(preventDefault).toHaveBeenCalled();
    expect(component.importFile).toBe(file);
    expect(component.importError()).toBeNull();
    expect(component.importResult()).toBeNull();

    const stopPropagation = vi.fn();
    component.removeImportFile(
      { stopPropagation } as unknown as MouseEvent,
      document.createElement('input'),
    );

    expect(stopPropagation).toHaveBeenCalled();
    expect(component.importFile).toBeNull();
  });

  it('navigates brands to the dedicated pages', () => {
    component.navigateToCreate();
    component.navigateToImport();
    component.navigateToEdit({ id: 'marca-1', nombre: 'Marca', activo: true });

    expect(router.navigate).toHaveBeenNthCalledWith(1, ['/catalogo/marcas', 'nuevo']);
    expect(router.navigate).toHaveBeenNthCalledWith(2, ['/catalogo/marcas', 'importar']);
    expect(router.navigate).toHaveBeenNthCalledWith(3, ['/catalogo/marcas', 'marca-1', 'editar']);
  });

  it('shows the inserted-record summary and returns to the catalog after confirmation', async () => {
    const fire = vi.spyOn(Swal, 'fire').mockResolvedValue({ isConfirmed: true } as never);
    component.importFile = new File(['contenido'], 'marcas.xlsx');

    component.importCatalog();

    http
      .expectOne({
        method: 'POST',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/importar`,
      })
      .flush({ procesadas: 3, creadas: 3, omitidas: 0, invalidas: 0, errores: [] });

    await Promise.resolve();

    expect(fire).toHaveBeenCalledWith(
      expect.objectContaining({
        icon: 'success',
        title: 'Importación terminada',
        text: '3 registros insertados · 0 omitidos · 0 inválidos',
        confirmButtonText: 'Ver registros',
      }),
    );
    expect(router.navigate).toHaveBeenCalledWith(['/catalogo/marcas']);
  });

  it('uses a warning for a partial import and does not navigate after an API error', async () => {
    const fire = vi.spyOn(Swal, 'fire').mockResolvedValue({ isConfirmed: true } as never);
    component.importFile = new File(['contenido'], 'marcas.xlsx');

    component.importCatalog();
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/importar`)
      .flush({ procesadas: 2, creadas: 1, omitidas: 1, invalidas: 0, errores: [] });

    await Promise.resolve();
    expect(fire).toHaveBeenCalledWith(expect.objectContaining({ icon: 'warning' }));

    router.navigate.mockClear();
    fire.mockClear();
    component.importFile = new File(['contenido'], 'marcas.xlsx');
    component.importCatalog();
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/importar`)
      .flush({ mensaje: 'Archivo inválido' }, { status: 400, statusText: 'Bad Request' });

    expect(fire).not.toHaveBeenCalled();
    expect(router.navigate).not.toHaveBeenCalled();
  });
});

describe('CatalogoComponent al crear una categoría', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;
  const negocioId = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';

  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [CatalogoComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { data: { section: 'categorias', mode: 'create' } } },
        },
        {
          provide: ContextoService,
          useValue: { negocio: signal({ id: negocioId }), puede: () => true },
        },
        { provide: Router, useValue: { navigate: vi.fn().mockResolvedValue(true) } },
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('carga las categorías existentes como opciones de categoría padre', () => {
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/categorias`)
      .flush([
        { id: 'categoria-1', nombre: 'Bebidas', activo: true },
        { id: 'categoria-2', nombre: 'Refrescos', activo: true },
      ]);

    expect(component.parentCategoryOptions).toEqual([
      { label: 'Bebidas', value: 'categoria-1' },
      { label: 'Refrescos', value: 'categoria-2' },
    ]);
    expect(component.loading()).toBe(false);
  });
});

describe('Filtros de CatalogoComponent', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;
  const negocioId = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';

  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [CatalogoComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ActivatedRoute, useValue: { snapshot: { data: { section: 'marcas' } } } },
        {
          provide: ContextoService,
          useValue: { negocio: signal({ id: negocioId }), puede: () => true },
        },
        { provide: Router, useValue: { navigate: vi.fn().mockResolvedValue(true) } },
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`).flush([]);
  });

  afterEach(() => http.verify());

  it('busca sin distinguir acentos y filtra por estado', () => {
    component.items.set([
      { id: 'marca-1', nombre: 'Café del Sur', activo: true },
      { id: 'marca-2', nombre: 'Té Norte', activo: false },
    ]);

    component.setCatalogSearch('cafe');
    expect(component.filteredItems().map((item) => item.id)).toEqual(['marca-1']);

    component.setCatalogSearch('');
    component.setCatalogStatusFilter('inactive');
    expect(component.filteredItems().map((item) => item.id)).toEqual(['marca-2']);

    component.clearCatalogFilters();
    expect(component.filteredItems()).toHaveLength(2);
  });

  it('filtra categorías por padre y por categorías sin padre', () => {
    (component as any).section = 'categorias';
    component.items.set([
      { id: 'categoria-1', nombre: 'Bebidas', activo: true },
      { id: 'categoria-2', nombre: 'Refrescos', categoria_padre_id: 'categoria-1', activo: true },
    ]);

    component.setCategoryParentFilter('categoria-1');
    expect(component.filteredItems().map((item) => item.id)).toEqual(['categoria-2']);

    component.setCategoryParentFilter('__sin_categoria_padre__');
    expect(component.filteredItems().map((item) => item.id)).toEqual(['categoria-1']);
  });

  it('filtra unidades por tipo y reinicia la paginación', () => {
    (component as any).section = 'unidades';
    component.items.set([
      { id: 'unidad-1', nombre: 'Gramo', simbolo: 'g', tipo: 'PESO', factor_a_base: 1, permite_fraccion: true, decimales: 2, activo: true },
      { id: 'unidad-2', nombre: 'Caja', simbolo: 'cja', tipo: 'EMPAQUE', factor_a_base: 12, permite_fraccion: false, decimales: 0, activo: true },
    ]);
    component.catalogPageFirst.set(10);

    component.setUnitTypeFilter('EMPAQUE');

    expect(component.filteredItems().map((item) => item.id)).toEqual(['unidad-2']);
    expect(component.catalogPageFirst()).toBe(0);
  });

  it('busca proveedores por sus datos de contacto', () => {
    (component as any).section = 'proveedores';
    component.items.set([
      { id: 'proveedor-1', nombre: 'Comercializadora', rfc: 'ABC010101AAA', activo: true },
      { id: 'proveedor-2', nombre: 'Mayorista', email: 'ventas@ejemplo.mx', activo: true },
    ]);

    component.setCatalogSearch('VENTAS@EJEMPLO.MX');

    expect(component.filteredItems().map((item) => item.id)).toEqual(['proveedor-2']);
  });
});

describe('Eliminación segura de CatalogoComponent', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;
  let feedback: FeedbackService;
  const negocioId = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb';

  beforeEach(() => {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [CatalogoComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ActivatedRoute, useValue: { snapshot: { data: { section: 'marcas' } } } },
        {
          provide: ContextoService,
          useValue: { negocio: signal({ id: negocioId }), puede: () => true },
        },
        { provide: Router, useValue: { navigate: vi.fn().mockResolvedValue(true) } },
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
    feedback = TestBed.inject(FeedbackService);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`).flush([]);
  });

  afterEach(() => http.verify());

  it('confirma y elimina una marca, conservando la tabla actualizada', async () => {
    vi.spyOn(feedback, 'confirmDanger').mockResolvedValue(true);
    const success = vi.spyOn(feedback, 'success');

    const deletion = component.deleteCatalogItem({ id: 'marca-1', nombre: 'Marca', activo: true });
    await Promise.resolve();

    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/marca-1/uso`)
      .flush({ puede_eliminarse: true, vinculaciones: [] });
    await Promise.resolve();

    http
      .expectOne({
        method: 'DELETE',
        url: `${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/marca-1`,
      })
      .flush(null);
    http.expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas`).flush([]);
    await deletion;

    expect(success).toHaveBeenCalledWith('Marca eliminada');
    expect(component.deletingId()).toBeNull();
  });

  it('muestra la vinculación cuando la API rechaza el borrado', async () => {
    vi.spyOn(feedback, 'confirmDanger').mockResolvedValue(true);
    const warning = vi.spyOn(feedback, 'warning');
    const confirm = vi.spyOn(feedback, 'confirmDanger');

    const deletion = component.deleteCatalogItem({ id: 'marca-1', nombre: 'Marca', activo: true });
    await Promise.resolve();
    http
      .expectOne(`${environment.apiUrl}/negocios/${negocioId}/catalogo/marcas/marca-1/uso`)
      .flush({ puede_eliminarse: false, vinculaciones: ['productos'] });
    await deletion;

    expect(warning).toHaveBeenCalledWith('No se puede eliminar', 'Tiene vinculaciones con: productos.');
    expect(confirm).not.toHaveBeenCalled();
    expect(component.deletingId()).toBeNull();
  });
});
