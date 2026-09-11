import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { NegocioService } from './negocio.service';
import { NegociosComponent } from './negocios.component';

describe('NegociosComponent', () => {
  let fixture: ComponentFixture<NegociosComponent>;
  const service = {
    listar: vi.fn(() => of({ items: [], total: 0 })),
    archivar: vi.fn(() => of(undefined)),
    restaurar: vi.fn(() => of({})),
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [NegociosComponent],
      providers: [{ provide: NegocioService, useValue: service }, provideRouter([])],
    }).compileComponents();
    fixture = TestBed.createComponent(NegociosComponent);
    fixture.detectChanges();
  });

  it('muestra estado vacío y carga activos inicialmente', () => {
    expect(service.listar).toHaveBeenCalledWith('activo');
    expect(fixture.nativeElement.textContent).toContain('Aún no tienes negocios');
  });

  it('cambia al filtro archivados', () => {
    fixture.componentInstance.setStatus('archivado');
    fixture.detectChanges();

    expect(service.listar).toHaveBeenLastCalledWith('archivado');
    expect(fixture.nativeElement.textContent).toContain('No tienes negocios archivados');
  });
});
