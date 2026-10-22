import '@angular/common/locales/global/es-MX';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { environment } from '../../../../environments/environment';
import { ContextoService } from '../../../contexto/contexto.service';
import { Invitaciones } from './invitaciones';

describe('Invitaciones', () => {
  const url = `${environment.apiUrl}/negocios/negocio-1/administracion/invitaciones`;
  const sucursal = signal({ id: 'suc-centro', nombre: 'Sucursal Centro' });
  let http: HttpTestingController;

  const invitaciones = (total: number) =>
    Array.from({ length: total }, (_, i) => ({ id: `inv-${i}`, nombre_empleado: `Empleado ${i}`, correo: `e${i}@x.mx`, estado: 'pendiente', expira_en: '2026-10-25T00:00:00Z', creado_en: '2026-10-22T00:00:00Z' }));

  it('solo pide las de la sucursal del contexto, recarga al cambiarla y pagina de 10 en 10', () => {
    TestBed.configureTestingModule({
      imports: [Invitaciones],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: ContextoService, useValue: { negocio: signal({ id: 'negocio-1', permisos: [] }), sucursal, permisos: signal(new Set()), puede: () => true } },
      ],
    });
    const fixture = TestBed.createComponent(Invitaciones);
    http = TestBed.inject(HttpTestingController);
    fixture.detectChanges();

    http.expectOne(`${url}?sin_aceptar=true&sucursal_id=suc-centro`).flush({ items: invitaciones(23), total: 23 });
    fixture.detectChanges();
    const page: HTMLElement = fixture.nativeElement;
    expect(page.querySelector('#filtro-sucursal')).toBeNull();
    expect(page.querySelectorAll('tbody tr').length).toBe(10);
    expect(page.querySelector('.tabla-paginador')?.textContent).toContain('1–10 de 23');

    page.querySelector<HTMLButtonElement>('[aria-label="Página siguiente"]')!.click();
    fixture.detectChanges();
    expect(page.querySelector('.tabla-paginador')?.textContent).toContain('11–20 de 23');

    sucursal.set({ id: 'suc-norte', nombre: 'Sucursal Norte' });
    fixture.detectChanges();
    http.expectOne(`${url}?sin_aceptar=true&sucursal_id=suc-norte`).flush({ items: invitaciones(2), total: 2 });
    fixture.detectChanges();
    expect(page.querySelector('.tabla-paginador')).toBeNull();
    expect(page.textContent).toContain('Sucursal Norte');
    http.verify();
  });
});
