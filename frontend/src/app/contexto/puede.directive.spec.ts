import { Component, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';

import { ContextoService } from './contexto.service';
import { PERMISOS } from './permisos';
import { PuedeDirective } from './puede.directive';

@Component({
  imports: [PuedeDirective],
  template: `<button *appPuede="P.rolGestionar">Nuevo rol</button>`,
})
class Anfitrion {
  readonly P = PERMISOS;
}

describe('PuedeDirective', () => {
  it('muestra y oculta el elemento cuando cambian los permisos', () => {
    const otorgados = signal<string[]>([]);
    TestBed.configureTestingModule({
      imports: [Anfitrion],
      providers: [
        { provide: ContextoService, useValue: { puede: (codigo: string) => otorgados().includes(codigo) } },
      ],
    });
    const fixture = TestBed.createComponent(Anfitrion);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('button')).toBeNull();

    otorgados.set(['roles.gestionar']);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('button')?.textContent).toBe('Nuevo rol');

    otorgados.set([]);
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });
});
