import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { SeccionOcultable, VistasService } from '../../contexto/vistas.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { VistasComponent } from './vistas.component';

describe('VistasComponent', () => {
  let fixture: ComponentFixture<VistasComponent>;
  let ocultas: ReturnType<typeof signal<ReadonlySet<SeccionOcultable>>>;
  let guardar: ReturnType<typeof vi.fn>;
  let success: ReturnType<typeof vi.fn>;

  function crear(iniciales: SeccionOcultable[] = []): HTMLElement {
    ocultas = signal<ReadonlySet<SeccionOcultable>>(new Set(iniciales));
    guardar = vi.fn((lista: SeccionOcultable[]) => of({ secciones_ocultas: [...lista].sort(), secciones_ocultables: ['categorias', 'marcas'] }));
    success = vi.fn();
    TestBed.configureTestingModule({
      imports: [VistasComponent],
      providers: [
        { provide: VistasService, useValue: { asegurar: () => of(undefined), ocultas, guardar } },
        { provide: ContextoService, useValue: { negocio: signal({ id: 'n1', nombre_comercial: 'Empresa THRS' }) } },
        { provide: FeedbackService, useValue: { success } },
      ],
    });
    fixture = TestBed.createComponent(VistasComponent);
    fixture.detectChanges();
    return fixture.nativeElement as HTMLElement;
  }

  afterEach(() => TestBed.resetTestingModule());

  const interruptor = (root: HTMLElement, nombre: string) =>
    root.querySelector<HTMLButtonElement>(`button[role="switch"][aria-label="Mostrar ${nombre}"]`)!;
  const boton = (root: HTMLElement, texto: string) =>
    [...root.querySelectorAll<HTMLButtonElement>('button')].find((item) => item.textContent?.trim() === texto)!;

  it('empieza con lo guardado y deja fijas Productos y Unidades', () => {
    const root = crear(['marcas']);
    expect(root.textContent).toContain('Empresa THRS');
    expect(interruptor(root, 'Marcas').getAttribute('aria-checked')).toBe('false');
    expect(interruptor(root, 'Categorías').getAttribute('aria-checked')).toBe('true');
    expect(interruptor(root, 'Productos').disabled).toBe(true);
    expect(interruptor(root, 'Unidades de medida').disabled).toBe(true);
    expect(boton(root, 'Guardar cambios').disabled).toBe(true);
  });

  it('oculta Categorías, avisa y guarda para todo el equipo', () => {
    const root = crear();
    interruptor(root, 'Categorías').click();
    fixture.detectChanges();
    expect(interruptor(root, 'Categorías').getAttribute('aria-checked')).toBe('false');
    expect(root.textContent).toContain('Categorías quedará oculta para todo el equipo');
    expect(root.textContent).toContain('1 cambio sin guardar.');
    expect(root.textContent).toContain('Campo, filtro y columna «Categoría».');
    boton(root, 'Guardar cambios').click();
    fixture.detectChanges();
    expect(guardar).toHaveBeenCalledWith(['categorias']);
    expect(success).toHaveBeenCalled();
    expect(root.textContent).toContain('Sin cambios pendientes.');
  });

  it('descartar vuelve a lo guardado', () => {
    const root = crear(['categorias']);
    interruptor(root, 'Categorías').click();
    interruptor(root, 'Marcas').click();
    fixture.detectChanges();
    expect(root.textContent).toContain('2 cambios sin guardar.');
    boton(root, 'Descartar').click();
    fixture.detectChanges();
    expect(interruptor(root, 'Categorías').getAttribute('aria-checked')).toBe('false');
    expect(interruptor(root, 'Marcas').getAttribute('aria-checked')).toBe('true');
    expect(guardar).not.toHaveBeenCalled();
  });

  it('muestra el error de la API sin perder la selección', () => {
    const root = crear();
    guardar.mockReturnValueOnce(throwError(() => ({ error: { mensaje: 'Solo el propietario puede cambiar las vistas.' } })));
    interruptor(root, 'Marcas').click();
    fixture.detectChanges();
    boton(root, 'Guardar cambios').click();
    fixture.detectChanges();
    expect(root.querySelector('[role="alert"]')?.textContent).toContain('Solo el propietario');
    expect(interruptor(root, 'Marcas').getAttribute('aria-checked')).toBe('false');
  });
});
