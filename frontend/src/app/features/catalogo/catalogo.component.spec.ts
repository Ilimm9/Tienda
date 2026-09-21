import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { environment } from '../../../environments/environment';
import { CatalogoComponent } from './catalogo.component';

describe('CatalogoComponent', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;
  const router = { navigate: vi.fn().mockResolvedValue(true) };

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
          useValue: { negocio: signal({ id: environment.defaultBusinessId }) },
        },
        { provide: Router, useValue: router },
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
    http.expectOne(`${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/marcas`).flush([]);
    router.navigate.mockClear();
  });

  afterEach(() => http.verify());

  it('shows a save error immediately after a failed edit request', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.form.controls.nombre.setValue('Marca actualizada');

    component.save();

    http
      .expectOne({ method: 'PATCH', url: `${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/marcas/marca-1` })
      .flush({ mensaje: 'La marca ya existe.' }, { status: 400, statusText: 'Bad Request' });

    expect(component.saving()).toBe(false);
    expect(component.formError()).toBe('La marca ya existe.');
  });

  it('shows a connection error when the browser cannot provide an API message', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.save();

    http
      .expectOne({ method: 'PATCH', url: `${environment.apiUrl}/negocios/${environment.defaultBusinessId}/catalogo/marcas/marca-1` })
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

  it('maps parent categories to select options and excludes the edited category', () => {
    component.parents.set([
      { id: 'categoria-1', nombre: 'Bebidas', activo: true },
      { id: 'categoria-2', nombre: 'Refrescos', activo: true },
    ]);
    component.editingId = 'categoria-2';

    expect(component.parentCategoryOptions).toEqual([
      { label: 'Bebidas', value: 'categoria-1' },
    ]);
  });

  it('marks a dropped file as ready and clears it from the remove control', () => {
    const file = new File(['contenido'], 'marcas.xlsx', {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    });
    const preventDefault = vi.fn();
    component.importError.set('Error anterior');
    component.importResult.set({ procesadas: 1, creadas: 1, omitidas: 0, invalidas: 0, errores: [] });

    component.onImportDrop({ preventDefault, dataTransfer: { files: [file] } } as unknown as DragEvent);

    expect(preventDefault).toHaveBeenCalled();
    expect(component.importFile).toBe(file);
    expect(component.importError()).toBeNull();
    expect(component.importResult()).toBeNull();

    const stopPropagation = vi.fn();
    component.removeImportFile({ stopPropagation } as unknown as MouseEvent, document.createElement('input'));

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
});
