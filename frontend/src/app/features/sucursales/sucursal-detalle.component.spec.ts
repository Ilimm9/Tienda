import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalDetalle } from './sucursal.models';
import { SucursalService } from './sucursal.service';
import { SucursalDetalleComponent } from './sucursal-detalle.component';

describe('SucursalDetalleComponent', () => {
  const branch: SucursalDetalle = {
    id: 'branch-1',
    negocio_id: 'business-1',
    codigo: 'SUC-001',
    nombre: 'Matriz',
    telefono: null,
    direccion: null,
    es_principal: false,
    estado: 'activo',
    tipo_miembro: 'propietario',
    creado_en: '2026-09-10T00:00:00Z',
    actualizado_en: '2026-09-10T00:00:00Z',
    eliminado_en: null,
  };
  const restored = { ...branch, estado: 'activo' as const };
  const branches = {
    obtener: vi.fn(() => of(branch)),
    archivar: vi.fn(() => of(undefined)),
    restaurar: vi.fn(() => of(restored)),
  };
  const businesses = {
    obtener: vi.fn(() =>
      of({ id: branch.negocio_id, nombre_comercial: 'Tienda', estado: 'activo' }),
    ),
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
      imports: [SucursalDetalleComponent],
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
                get: (key: string) => (key === 'negocioId' ? branch.negocio_id : branch.id),
              },
            },
          },
        },
      ],
    }).compileComponents();
  });

  it('cancelar confirmación no ejecuta HTTP', async () => {
    feedback.confirmDanger.mockResolvedValueOnce(false);
    const component = TestBed.createComponent(SucursalDetalleComponent).componentInstance;

    await component.archive();

    expect(branches.archivar).not.toHaveBeenCalled();
  });

  it('archiva una vez, notifica y vuelve al filtro archivado', async () => {
    const component = TestBed.createComponent(SucursalDetalleComponent).componentInstance;

    await component.archive();

    expect(branches.archivar).toHaveBeenCalledOnce();
    expect(feedback.success).toHaveBeenCalledWith('Sucursal archivada');
    expect(router.navigate).toHaveBeenCalledWith(['/negocios', branch.negocio_id, 'sucursales'], {
      queryParams: { estado: 'archivado' },
    });
  });

  it('restaura sin confirmar y actualiza el detalle', () => {
    const component = TestBed.createComponent(SucursalDetalleComponent).componentInstance;

    component.restore();

    expect(feedback.confirmDanger).not.toHaveBeenCalled();
    expect(branches.restaurar).toHaveBeenCalledWith(branch.negocio_id, branch.id);
    expect(component.branch()).toEqual(restored);
    expect(feedback.success).toHaveBeenCalledWith('Sucursal restaurada');
  });
});
