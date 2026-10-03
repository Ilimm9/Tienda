import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { ContextoService } from '../../../contexto/contexto.service';
import { FeedbackService } from '../../../shared/feedback/feedback.service';
import { FechaMexicoPipe } from '../../../shared/fecha-mexico.pipe';
import { RolService } from '../../roles-permisos/rol.service';
import { FiltroInvitaciones, InvitacionResumen } from '../invitacion.models';
import { InvitacionService } from '../invitacion.service';

type Vista = 'sin_aceptar' | 'todas' | 'aceptada' | 'cancelada';
type TonoAvatar = 'lavanda' | 'salvia' | 'neutro';

interface EnlaceEmitido {
  enlace: string;
  correo: string;
  correoEnviado: boolean;
}

const HORA_MS = 60 * 60 * 1000;

@Component({
  selector: 'app-invitaciones',
  imports: [FechaMexicoPipe, RouterLink],
  templateUrl: './invitaciones.html',
  styleUrls: ['./invitaciones-tema.css', './invitaciones.css'],
})
export class Invitaciones {
  private readonly invitacionService = inject(InvitacionService);
  private readonly rolService = inject(RolService);
  private readonly feedback = inject(FeedbackService);
  readonly contexto = inject(ContextoService);

  readonly invitaciones = signal<InvitacionResumen[]>([]);
  readonly misPermisos = signal<string[]>([]);
  readonly vista = signal<Vista>('sin_aceptar');
  readonly sucursalId = signal('');
  readonly loading = signal(true);
  readonly processingId = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly enlaceEmitido = signal<EnlaceEmitido | null>(null);

  readonly negocioId = computed(() => this.contexto.negocio()?.id ?? '');
  readonly sucursales = computed(() => this.contexto.negocio()?.sucursales ?? []);
  readonly puedeEnviar = computed(() => this.misPermisos().includes('equipo.invitaciones.enviar'));
  readonly puedeCorregir = computed(
    () => this.puedeEnviar() && this.misPermisos().includes('equipo.empleados.gestionar'),
  );

  readonly vistas: { valor: Vista; etiqueta: string }[] = [
    { valor: 'sin_aceptar', etiqueta: 'Sin aceptar' },
    { valor: 'aceptada', etiqueta: 'Aceptadas' },
    { valor: 'cancelada', etiqueta: 'Canceladas' },
    { valor: 'todas', etiqueta: 'Todas' },
  ];

  constructor() {
    this.load();
  }

  setVista(valor: Vista): void {
    if (valor === this.vista()) return;
    this.vista.set(valor);
    this.load();
  }

  setSucursal(valor: string): void {
    this.sucursalId.set(valor);
    this.load();
  }

  /** Solo la invitación más reciente de un empleado sin cuenta puede reemitirse. */
  puedeReemitir(invitacion: InvitacionResumen): boolean {
    return this.puedeEnviar() && invitacion.estado !== 'aceptada' && Boolean(invitacion.sucursal_id);
  }

  /** Iniciales del nombre del empleado; sin nombre, la primera letra del correo. */
  iniciales(invitacion: InvitacionResumen): string {
    const partes = (invitacion.nombre_empleado ?? '').trim().split(/\s+/).filter(Boolean);
    if (partes.length) return partes.slice(0, 2).map((parte) => parte.charAt(0)).join('').toUpperCase();
    return invitacion.correo.charAt(0).toUpperCase();
  }

  /** Lavanda para pendientes, salvia para aceptadas y neutro para las cerradas sin aceptar. */
  tonoAvatar(invitacion: InvitacionResumen): TonoAvatar {
    if (invitacion.estado === 'aceptada') return 'salvia';
    if (invitacion.estado === 'pendiente') return 'lavanda';
    return 'neutro';
  }

  /** Texto del pie de la tabla con el total de la vista activa. */
  resumenLista(): string {
    if (this.loading()) return 'Cargando…';
    const total = this.invitaciones().length;
    const vista = this.vista();
    const sufijo =
      vista === 'sin_aceptar'
        ? 'sin aceptar'
        : vista === 'aceptada'
          ? total === 1 ? 'aceptada' : 'aceptadas'
          : vista === 'cancelada'
            ? total === 1 ? 'cancelada' : 'canceladas'
            : 'en total';
    return `${total} ${total === 1 ? 'invitación' : 'invitaciones'} ${sufijo}`;
  }

  horasRestantes(invitacion: InvitacionResumen): number {
    return Math.max(0, Math.ceil((new Date(invitacion.expira_en).getTime() - Date.now()) / HORA_MS));
  }

  reenviar(invitacion: InvitacionResumen): void {
    if (this.processingId()) return;
    this.processingId.set(invitacion.id);
    this.enlaceEmitido.set(null);
    this.invitacionService.reenviar(this.negocioId(), invitacion.id).subscribe({
      next: (creada) => {
        this.processingId.set(null);
        this.enlaceEmitido.set({
          enlace: this.invitacionService.enlaceDeToken(creada.token),
          correo: creada.invitacion.correo,
          correoEnviado: creada.correo_enviado,
        });
        this.feedback.success('Invitación reenviada', 'El enlace anterior ya no funciona.');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error('No fue posible reenviar la invitación', this.mensaje(response));
        // Un conflicto significa que la lista quedó desactualizada: aceptada o ya reemplazada.
        if (response.status === 409) this.load();
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

  async cancelar(invitacion: InvitacionResumen): Promise<void> {
    const confirmado = await this.feedback.confirmDanger({
      titulo: `Cancelar invitación de ${invitacion.nombre_empleado || invitacion.correo}`,
      descripcion: 'El enlace dejará de funcionar de inmediato.',
      textoConfirmar: 'Sí, cancelar',
      textoCancelar: 'Conservar',
    });
    if (!confirmado) return;

    this.processingId.set(invitacion.id);
    this.invitacionService.cancelar(this.negocioId(), invitacion.id).subscribe({
      next: () => {
        this.processingId.set(null);
        this.feedback.success('Invitación cancelada');
        this.load();
      },
      error: (response: HttpErrorResponse) => {
        this.processingId.set(null);
        this.feedback.error('No fue posible cancelar la invitación', this.mensaje(response));
      },
    });
  }

  private filtro(): FiltroInvitaciones {
    const vista = this.vista();
    return {
      sinAceptar: vista === 'sin_aceptar',
      estado: vista === 'aceptada' || vista === 'cancelada' ? vista : '',
      sucursalId: this.sucursalId() || undefined,
    };
  }

  private load(): void {
    const negocioId = this.negocioId();
    if (!negocioId) {
      this.loading.set(false);
      this.error.set('Selecciona un negocio para administrar sus invitaciones.');
      return;
    }
    this.loading.set(true);
    this.error.set(null);
    forkJoin({
      invitaciones: this.invitacionService.listar(negocioId, this.filtro()),
      permisos: this.rolService.misPermisos(negocioId),
    }).subscribe({
      next: ({ invitaciones, permisos }) => {
        this.invitaciones.set(invitaciones.items);
        this.misPermisos.set(permisos.items);
        this.loading.set(false);
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        this.error.set(
          response.status === 403
            ? 'No tienes permiso para ver las invitaciones de este negocio.'
            : 'No fue posible cargar las invitaciones.',
        );
      },
    });
  }

  private mensaje(response: HttpErrorResponse): string {
    return (response.error as { mensaje?: string } | null)?.mensaje ?? 'Intenta nuevamente.';
  }
}
