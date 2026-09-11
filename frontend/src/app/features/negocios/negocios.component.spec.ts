import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioResumen } from './negocio.models';
import { NegocioService } from './negocio.service';
import { NegociosComponent } from './negocios.component';

describe('NegociosComponent', () => {
  let fixture: ComponentFixture<NegociosComponent>;
  const service = {
    listar: vi.fn(() => of({ items: [], total: 0 })),
    archivar: vi.fn(() => of(undefined)),
    restaurar: vi.fn(() => of({})),
  };
  const feedback = {
    confirmDanger: vi.fn(() => Promise.resolve(true)),
    success: vi.fn(),
    error: vi.fn(),
  };
  const business: NegocioResumen = {
    id: 'business-1',
    slug: 'tienda-centro',
    nombre_comercial: 'Tienda Centro',
    rfc: null,
    tipo_miembro: 'propietario',
    estado: 'activo',
    tiene_sucursales: false,
    total_sucursales: 0,
    creado_en: '2026-09-10T00:00:00Z',
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [NegociosComponent],
      providers: [
        { provide: NegocioService, useValue: service },
        { provide: FeedbackService, useValue: feedback },
        provideRouter([]),
      ],
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

  it('no archiva cuando la confirmación se cancela', async () => {
    feedback.confirmDanger.mockResolvedValueOnce(false);

    await fixture.componentInstance.archive(business);

    expect(service.archivar).not.toHaveBeenCalled();
  });

  it('archiva una sola vez y notifica el resultado', async () => {
    await fixture.componentInstance.archive(business);

    expect(service.archivar).toHaveBeenCalledOnce();
    expect(service.archivar).toHaveBeenCalledWith(business.id);
    expect(feedback.success).toHaveBeenCalledWith('Negocio archivado');
  });

  it('muestra el fallo de restauración como notificación', () => {
    service.restaurar.mockReturnValueOnce(
      throwError(() => ({ error: { mensaje: 'No autorizado' } })),
    );

    fixture.componentInstance.restore({ ...business, estado: 'archivado' });

    expect(feedback.error).toHaveBeenCalledWith(
      'No fue posible restaurar el negocio',
      'No autorizado',
    );
    expect(fixture.componentInstance.error()).toBeNull();
  });
});
