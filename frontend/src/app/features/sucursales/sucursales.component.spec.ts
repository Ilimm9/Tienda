import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';

import { FeedbackService } from '../../shared/feedback/feedback.service';
import { NegocioService } from '../negocios/negocio.service';
import { SucursalResumen } from './sucursal.models';
import { SucursalService } from './sucursal.service';
import { SucursalesComponent } from './sucursales.component';

describe('SucursalesComponent', () => {
  let fixture: ComponentFixture<SucursalesComponent>;
  const branch: SucursalResumen = {
    id: 'branch-1',
    negocio_id: 'business-1',
    codigo: 'SUC-001',
    nombre: 'Matriz',
    telefono: null,
    direccion_resumida: 'Mérida, Yucatán',
    es_principal: true,
    estado: 'activo',
    tipo_miembro: 'propietario',
    creado_en: '2026-09-10T00:00:00Z',
    actualizado_en: '2026-09-10T00:00:00Z',
  };
  const business = {
    id: 'business-1',
    nombre_comercial: 'Tienda Centro',
    tipo_miembro: 'propietario',
    estado: 'activo',
  };
  const branches = {
    listar: vi.fn(() => of({ items: [branch], total: 1 })),
    archivar: vi.fn(() => of(undefined)),
    restaurar: vi.fn(() => of({ ...branch, estado: 'activo' })),
  };
  const businesses = { obtener: vi.fn(() => of(business)) };
  const feedback = {
    confirmDanger: vi.fn(() => Promise.resolve(true)),
    success: vi.fn(),
    error: vi.fn(),
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    await TestBed.configureTestingModule({
      imports: [SucursalesComponent],
      providers: [
        { provide: SucursalService, useValue: branches },
        { provide: NegocioService, useValue: businesses },
        { provide: FeedbackService, useValue: feedback },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: {
              paramMap: { get: () => business.id },
              queryParamMap: { get: () => null },
            },
          },
        },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(SucursalesComponent);
    fixture.detectChanges();
  });

  it('carga cards activas y muestra estado principal', () => {
    expect(branches.listar).toHaveBeenCalledWith(business.id, 'activo', '');
    expect(fixture.nativeElement.textContent).toContain('Matriz');
    expect(fixture.nativeElement.textContent).toContain('Principal');
  });

  it('envía filtros de búsqueda y archivadas', () => {
    fixture.componentInstance.updateSearch('centro');
    fixture.componentInstance.search();
    fixture.componentInstance.setStatus('archivado');

    expect(branches.listar).toHaveBeenCalledWith(business.id, 'activo', 'centro');
    expect(branches.listar).toHaveBeenLastCalledWith(business.id, 'archivado', 'centro');
  });

  it('cancelar confirmación no ejecuta el archivado', async () => {
    feedback.confirmDanger.mockResolvedValueOnce(false);

    await fixture.componentInstance.archive(branch);

    expect(branches.archivar).not.toHaveBeenCalled();
  });

  it('archiva una vez, notifica y recarga', async () => {
    await fixture.componentInstance.archive(branch);

    expect(branches.archivar).toHaveBeenCalledOnce();
    expect(feedback.success).toHaveBeenCalledWith('Sucursal archivada');
    expect(branches.listar).toHaveBeenCalledTimes(2);
  });

  it('restaura sin confirmación y notifica', () => {
    fixture.componentInstance.restore({ ...branch, estado: 'archivado' });

    expect(feedback.confirmDanger).not.toHaveBeenCalled();
    expect(branches.restaurar).toHaveBeenCalledOnce();
    expect(feedback.success).toHaveBeenCalledWith('Sucursal restaurada');
  });
});
