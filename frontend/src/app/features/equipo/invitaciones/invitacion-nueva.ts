import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../../contexto/contexto.service';
import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { RolResumen } from '../../roles-permisos/rol.models';
import { RolService } from '../../roles-permisos/rol.service';
import { EmpleadoResumen } from '../empleado.models';
import { EmpleadoService } from '../empleado.service';
import { InvitacionService } from '../invitacion.service';

// El rol propietario nunca se delega por invitación; el servidor también lo rechaza.
const CODIGO_ROL_PROPIETARIO = 'PROPIETARIO';

interface EnlaceEmitido {
  enlace: string;
  correo: string;
  correoEnviado: boolean;
}

@Component({
  selector: 'app-invitacion-nueva',
  imports: [FormsModule, RouterLink],
  templateUrl: './invitacion-nueva.html',
  styleUrls: ['./invitaciones-tema.css', './invitacion-nueva.css'],
})
export class InvitacionNueva {
  private readonly invitacionService = inject(InvitacionService);
  private readonly empleadoService = inject(EmpleadoService);
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  readonly contexto = inject(ContextoService);

  readonly invitables = signal<EmpleadoResumen[]>([]);
  readonly roles = signal<RolResumen[]>([]);
  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly enlaceEmitido = signal<EnlaceEmitido | null>(null);

  // Permite llegar con el empleado o la sucursal ya elegidos desde sus pantallas.
  readonly empleadoId = signal(this.route.snapshot.queryParamMap.get('empleado') ?? '');
  readonly sucursalId = signal(
    this.route.snapshot.queryParamMap.get('sucursal') ?? this.contexto.sucursal()?.id ?? '',
  );
  readonly rolId = signal('');

  readonly negocio = computed(() => this.contexto.negocio());
  readonly sucursales = computed(() => this.negocio()?.sucursales ?? []);
  readonly empleado = computed(
    () => this.invitables().find((item) => item.id === this.empleadoId()) ?? null,
  );
  readonly sucursal = computed(
    () => this.sucursales().find((item) => item.id === this.sucursalId()) ?? null,
  );
  readonly rol = computed(() => this.roles().find((item) => item.id === this.rolId()) ?? null);
  readonly completa = computed(() => Boolean(this.empleado() && this.sucursal() && this.rol()));

  constructor() {
    this.load();
  }

  emitir(): void {
    const negocio = this.negocio();
    if (!negocio || !this.completa() || this.saving()) return;
    this.saving.set(true);
    this.enlaceEmitido.set(null);
    this.invitacionService
      .crear(negocio.id, {
        empleado_id: this.empleadoId(),
        sucursal_id: this.sucursalId(),
        rol_predeterminado_id: this.rolId(),
      })
      .subscribe({
        next: (creada) => {
          this.saving.set(false);
          this.enlaceEmitido.set({
            enlace: this.invitacionService.enlaceDeToken(creada.token),
            correo: creada.invitacion.correo,
            correoEnviado: creada.correo_enviado,
          });
          if (creada.correo_enviado) {
            this.feedback.success('Invitación emitida', `Enviamos el enlace a ${creada.invitacion.correo}.`);
          } else {
            this.feedback.warning('No pudimos enviar el correo', 'Copia el enlace y compártelo.');
          }
        },
        error: (response: HttpErrorResponse) => {
          this.saving.set(false);
          this.feedback.error(
            'No fue posible emitir la invitación',
            (response.error as { mensaje?: string } | null)?.mensaje ?? 'Intenta nuevamente.',
          );
        },
      });
  }

  async copiar(enlace: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(enlace);
      this.feedback.success('Enlace copiado');
    } catch {
      this.feedback.warning('Copia el enlace manualmente');
    }
  }

  private load(): void {
    const negocio = this.negocio();
    if (!negocio) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para invitar empleados.');
      return;
    }
    forkJoin({
      empleados: this.empleadoService.listar(negocio.id),
      roles: this.rolService.listar(negocio.id),
    }).subscribe({
      next: ({ empleados, roles }) => {
        // Solo tiene sentido invitar a quien ya está registrado, tiene correo y aún no tiene cuenta.
        this.invitables.set(
          empleados.items.filter(
            (empleado) =>
              !empleado.tiene_cuenta &&
              Boolean(empleado.correo) &&
              empleado.estado !== 'suspendido' &&
              empleado.estado !== 'terminado',
          ),
        );
        this.roles.set(
          roles.items.filter(
            (rol) => rol.activo && rol.codigo.toUpperCase() !== CODIGO_ROL_PROPIETARIO,
          ),
        );
        if (!this.empleado()) this.empleadoId.set('');
        if (!this.sucursal()) this.sucursalId.set('');
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para invitar empleados en este negocio.'
            : 'No fue posible cargar empleados y roles.',
        );
      },
    });
  }
}
