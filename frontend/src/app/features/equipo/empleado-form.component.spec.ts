import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter, Router } from '@angular/router';

import { environment } from '../../../environments/environment';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { EmpleadoFormComponent } from './empleado-form.component';
import { CREAR_NUEVO } from './presentacion';

describe('EmpleadoFormComponent (alta en 3 pasos)', () => {
  const negocioId = 'negocio-1';
  const base = `${environment.apiUrl}/negocios/${negocioId}/administracion`;
  const centro = { id: 'suc-centro', nombre: 'Sucursal Centro', codigo: 'SUC-001' };
  let fixture: ComponentFixture<EmpleadoFormComponent>;
  let component: EmpleadoFormComponent;
  let http: HttpTestingController;

  const empleado = (correo: string | null) => ({ id: 'emp-1', nombre: 'Ana', primer_apellido: 'Ruiz', correo, puesto_id: 'p1', puesto: 'Cajero', numero_empleado: 'SC-003' });

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [EmpleadoFormComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: convertToParamMap({}) } } },
        {
          provide: ContextoService,
          useValue: { negocio: signal({ id: negocioId, sucursales: [centro] }), sucursal: signal(centro), puede: () => true },
        },
      ],
    });
    fixture = TestBed.createComponent(EmpleadoFormComponent);
    component = fixture.componentInstance;
    http = TestBed.inject(HttpTestingController);
    http.expectOne(`${base}/puestos`).flush({ items: [{ id: 'p1', nombre: 'Cajero', rol_sugerido_id: 'rol-cajero', activo: true }], total: 1 });
    http.expectOne(`${base}/empleados`).flush({ items: [{ numero_empleado: 'SC-002' }, { numero_empleado: 'SN-009' }], total: 2 });
    http.expectOne(`${base}/roles?delegables=true`).flush({ items: [{ id: 'rol-cajero', codigo: 'CAJERO', nombre: 'Cajero', activo: true }], total: 1 });
    fixture.detectChanges();
  });
  afterEach(() => http.verify());

  function registrar(correo: string | null): void {
    component.form.patchValue({ nombre: 'Ana', primer_apellido: 'Ruiz', segundo_nombre: 'María', correo: correo ?? '', puesto_id: 'p1' });
    component.form.controls.telefono.setValue('+52 55 1234 5678');
    component.formatearTelefono();
    component.generarNumero();
    component.submit();
    const crear = http.expectOne({ method: 'POST', url: `${base}/empleados` });
    expect(crear.request.body).toMatchObject({ puesto_id: 'p1', telefono: '5512345678', numero_empleado: 'SC-003', segundo_nombre: 'María' });
    crear.flush(empleado(correo));
    fixture.detectChanges();
  }

  it('crea, asigna la sucursal del contexto e invita con el rol sugerido por el puesto', () => {
    registrar('ana@correo.mx');
    expect(component.paso()).toBe(2);

    component.guardarSucursales();
    const asignar = http.expectOne({ method: 'POST', url: `${base}/empleados/emp-1/sucursales` });
    expect(asignar.request.body).toEqual({ sucursal_id: centro.id, es_principal: true });
    asignar.flush({ items: [], total: 0 });
    http.expectOne(`${base}/empleados/emp-1/sucursales`).flush({
      items: [{ id: 'a1', sucursal_id: centro.id, nombre_sucursal: centro.nombre, es_principal: true }],
      total: 1,
    });
    fixture.detectChanges();
    expect(component.paso()).toBe(3);
    expect(component.rolId()).toBe('rol-cajero');

    component.enviarInvitacion();
    const invitar = http.expectOne({ method: 'POST', url: `${base}/invitaciones` });
    expect(invitar.request.body).toEqual({ empleado_id: 'emp-1', sucursal_id: centro.id, rol_predeterminado_id: 'rol-cajero' });
    invitar.flush({ invitacion: { correo: 'ana@correo.mx' }, token: 't', correo_enviado: true });
    fixture.detectChanges();
    expect(component.paso()).toBe(4);
  });

  it('si se omite la sucursal bloquea la invitación y permite volver a los pasos 2 y 1', () => {
    registrar('ana@correo.mx');
    component.irPaso(3);
    fixture.detectChanges();
    expect(component.bloqueoInvitacion()).toBe('sucursal');
    expect(fixture.nativeElement.textContent).toContain('Volver a sucursal');

    component.irPaso(2);
    expect(component.paso()).toBe(2);
    component.irPaso(1);
    fixture.detectChanges();
    expect(component.paso()).toBe(1);

    // Volver al paso 1 actualiza al mismo empleado en vez de crear otro.
    component.submit();
    http.expectOne({ method: 'PATCH', url: `${base}/empleados/emp-1` }).flush(empleado('ana@correo.mx'));
    expect(component.paso()).toBe(2);
  });

  it('ofrece crear un puesto y, sin datos capturados, va directo a Puestos', async () => {
    const navegar = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
    const select: HTMLSelectElement = fixture.nativeElement.querySelector('#puesto_id');
    expect(select.options[select.options.length - 1].textContent).toContain('Crear nuevo puesto');

    component.form.controls.puesto_id.setValue(CREAR_NUEVO);
    await component.alElegirPuesto();

    expect(navegar).toHaveBeenCalledWith(['/equipo/puestos']);
    expect(component.form.controls.puesto_id.value).toBe('');
  });

  it('con datos capturados pide confirmar antes de ir a Puestos', async () => {
    const navegar = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
    const confirmar = vi.spyOn(TestBed.inject(FeedbackService), 'confirmDanger').mockResolvedValue(false);
    component.form.controls.nombre.setValue('Ana');
    component.form.controls.nombre.markAsDirty();
    component.form.controls.puesto_id.setValue('p1');
    await component.alElegirPuesto();

    component.form.controls.puesto_id.setValue(CREAR_NUEVO);
    await component.alElegirPuesto();

    expect(confirmar).toHaveBeenCalled();
    expect(navegar).not.toHaveBeenCalled();
    expect(component.form.controls.puesto_id.value).toBe('p1');
  });

  it('el rol «Crear nuevo» lleva a Roles y permisos', () => {
    const navegar = vi.spyOn(TestBed.inject(Router), 'navigate').mockResolvedValue(true);
    component.elegirRol(CREAR_NUEVO);
    expect(navegar).toHaveBeenCalledWith(['/roles-permisos/nuevo']);
    component.elegirRol('rol-cajero');
    expect(component.rolId()).toBe('rol-cajero');
  });

  it('sin correo pide volver a datos', () => {
    registrar(null);
    component.irPaso(3);
    expect(component.bloqueoInvitacion()).toBe('correo');
  });

  it('rechaza un teléfono que no tenga 10 dígitos', () => {
    component.form.controls.telefono.setValue('55 1234');
    expect(component.form.controls.telefono.invalid).toBe(true);
  });
});
