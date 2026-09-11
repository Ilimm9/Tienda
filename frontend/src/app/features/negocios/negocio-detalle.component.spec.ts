import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from './negocio.models';
import { NegocioService } from './negocio.service';
import { NegocioDetalleComponent } from './negocio-detalle.component';

describe('NegocioDetalleComponent', () => {
  const business: NegocioDetalle = {
    id: 'business-1',
    slug: 'tienda-centro',
    nombre_comercial: 'Tienda Centro',
    razon_social: null,
    rfc: null,
    telefono: null,
    correo: null,
    codigo_moneda: 'MXN',
    zona_horaria: 'America/Mexico_City',
    tipo_miembro: 'propietario',
    estado: 'activo',
    tiene_sucursales: false,
    total_sucursales: 0,
    creado_en: '2026-09-10T00:00:00Z',
    actualizado_en: '2026-09-10T00:00:00Z',
    archivado_en: null,
    direccion: null,
  };
  const restored = { ...business, estado: 'activo' as const };
  const service = {
    obtener: vi.fn(() => of(business)),
    archivar: vi.fn(() => of(undefined)),
    restaurar: vi.fn(() => of(restored)),
  };
  const feedback = {
    confirmDanger: vi.fn(() => Promise.resolve(true)),
    success: vi.fn(),
    error: vi.fn(),
  };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [NegocioDetalleComponent],
      providers: [
        { provide: NegocioService, useValue: service },
        { provide: FeedbackService, useValue: feedback },
        { provide: Router, useValue: router },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: () => business.id } } },
        },
      ],
    }).compileComponents();
  });

  it('cancela el archivado sin ejecutar petición', async () => {
    feedback.confirmDanger.mockResolvedValueOnce(false);
    const component = TestBed.createComponent(NegocioDetalleComponent).componentInstance;

    await component.archive();

    expect(service.archivar).not.toHaveBeenCalled();
  });

  it('confirma, archiva y notifica antes de volver al listado', async () => {
    const component = TestBed.createComponent(NegocioDetalleComponent).componentInstance;

    await component.archive();

    expect(service.archivar).toHaveBeenCalledWith(business.id);
    expect(feedback.success).toHaveBeenCalledWith('Negocio archivado');
    expect(router.navigate).toHaveBeenCalledWith(['/negocios']);
  });

  it('restaura sin confirmación y actualiza el detalle', () => {
    const component = TestBed.createComponent(NegocioDetalleComponent).componentInstance;

    component.restore();

    expect(feedback.confirmDanger).not.toHaveBeenCalled();
    expect(service.restaurar).toHaveBeenCalledWith(business.id);
    expect(component.business()).toEqual(restored);
    expect(feedback.success).toHaveBeenCalledWith('Negocio restaurado');
  });
});
