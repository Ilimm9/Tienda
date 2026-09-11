import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from './negocio.models';
import { NegocioService } from './negocio.service';
import { NegocioFormComponent } from './negocio-form.component';

describe('NegocioFormComponent', () => {
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
  const service = {
    crear: vi.fn(() => of(business)),
    actualizar: vi.fn(() => of(business)),
    obtener: vi.fn(() => of(business)),
  };
  const feedback = { success: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };

  async function configure(negocioId: string | null): Promise<void> {
    await TestBed.configureTestingModule({
      imports: [NegocioFormComponent],
      providers: [
        { provide: NegocioService, useValue: service },
        { provide: FeedbackService, useValue: feedback },
        { provide: Router, useValue: router },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: { get: () => negocioId } } },
        },
      ],
    }).compileComponents();
  }

  beforeEach(() => vi.clearAllMocks());

  it('notifica una creación exitosa y navega al detalle', async () => {
    await configure(null);
    const component = TestBed.createComponent(NegocioFormComponent).componentInstance;
    component.form.patchValue({ nombre_comercial: 'Tienda Centro' });

    component.submit();

    expect(service.crear).toHaveBeenCalledOnce();
    expect(feedback.success).toHaveBeenCalledWith('Negocio registrado');
    expect(router.navigate).toHaveBeenCalledWith(['/negocios', business.id]);
  });

  it('notifica una edición exitosa sin cambiar el contrato', async () => {
    await configure(business.id);
    const component = TestBed.createComponent(NegocioFormComponent).componentInstance;

    component.submit();

    expect(service.actualizar).toHaveBeenCalledWith(
      business.id,
      expect.objectContaining({ nombre_comercial: business.nombre_comercial }),
    );
    expect(feedback.success).toHaveBeenCalledWith('Cambios guardados');
  });
});
