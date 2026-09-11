import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioDetalle } from '../negocios/negocio.models';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalDetalle } from './sucursal.models';
import { SucursalService } from './sucursal.service';
import { SucursalFormComponent } from './sucursal-form.component';

describe('SucursalFormComponent', () => {
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
  const branch: SucursalDetalle = {
    id: 'branch-1',
    negocio_id: business.id,
    codigo: 'SUC-001',
    nombre: 'Matriz',
    telefono: null,
    direccion: null,
    es_principal: true,
    estado: 'activo',
    tipo_miembro: 'propietario',
    creado_en: '2026-09-10T00:00:00Z',
    actualizado_en: '2026-09-10T00:00:00Z',
    eliminado_en: null,
  };
  const branches = {
    crear: vi.fn(() => of(branch)),
    obtener: vi.fn(() => of(branch)),
    actualizar: vi.fn(() => of(branch)),
  };
  const businesses = { obtener: vi.fn(() => of(business)) };
  const feedback = { success: vi.fn() };
  const router = { navigate: vi.fn(() => Promise.resolve(true)) };

  async function createComponent(sucursalId: string | null): Promise<SucursalFormComponent> {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [SucursalFormComponent],
      providers: [
        { provide: SucursalService, useValue: branches },
        { provide: NegocioService, useValue: businesses },
        { provide: FeedbackService, useValue: feedback },
        { provide: Router, useValue: router },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: {
                get: (key: string) => (key === 'negocioId' ? business.id : sucursalId),
              },
            },
          },
        },
      ],
    }).compileComponents();
    return TestBed.createComponent(SucursalFormComponent).componentInstance;
  }

  it('fuerza la primera sucursal como principal y normaliza el alta', async () => {
    const component = await createComponent(null);
    component.form.patchValue({ codigo: 'suc-002', nombre: '  Norte  ', telefono: ' 555 ' });

    component.submit();

    expect(component.form.controls.es_principal.disabled).toBe(true);
    expect(branches.crear).toHaveBeenCalledWith(business.id, {
      codigo: 'SUC-002',
      nombre: 'Norte',
      telefono: '555',
      es_principal: true,
    });
    expect(feedback.success).toHaveBeenCalledWith('Sucursal registrada');
  });

  it('bloquea visualmente el código durante edición y usa PATCH sin enviarlo', async () => {
    const component = await createComponent(branch.id);
    const fixture = TestBed.createComponent(SucursalFormComponent);
    fixture.detectChanges();

    expect(
      (fixture.nativeElement.querySelector('input[formcontrolname="codigo"]') as HTMLInputElement)
        .readOnly,
    ).toBe(true);

    component.form.controls.nombre.setValue('Matriz Centro');
    component.submit();
    expect(branches.actualizar).toHaveBeenCalledWith(
      business.id,
      branch.id,
      expect.not.objectContaining({ codigo: expect.anything() }),
    );
  });
});
