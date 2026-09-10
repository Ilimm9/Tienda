import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';

import { environment } from '../../../environments/environment';
import { CatalogoComponent } from './catalogo.component';

describe('CatalogoComponent', () => {
  let component: CatalogoComponent;
  let http: HttpTestingController;

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
      ],
    });
    component = TestBed.createComponent(CatalogoComponent).componentInstance;
    http = TestBed.inject(HttpTestingController);
    http.expectOne(`${environment.apiUrl}/catalogo/marcas`).flush([]);
  });

  afterEach(() => http.verify());

  it('shows a save error immediately after a failed edit request', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.form.controls.nombre.setValue('Marca actualizada');

    component.save();

    http
      .expectOne({ method: 'PATCH', url: `${environment.apiUrl}/catalogo/marcas/marca-1` })
      .flush({ mensaje: 'La marca ya existe.' }, { status: 400, statusText: 'Bad Request' });

    expect(component.saving()).toBe(false);
    expect(component.formError()).toBe('La marca ya existe.');
  });

  it('shows a connection error when the browser cannot provide an API message', () => {
    component.openEdit({ id: 'marca-1', nombre: 'Marca original', activo: true });
    component.save();

    http
      .expectOne({ method: 'PATCH', url: `${environment.apiUrl}/catalogo/marcas/marca-1` })
      .error(new ProgressEvent('error'));

    expect(component.formError()).toBe(
      'No fue posible comunicarse con el servidor. Intenta de nuevo.',
    );
  });
});
