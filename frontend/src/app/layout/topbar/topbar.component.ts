import { Component, HostListener, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { AuthService } from '../../features/auth/auth.service';
import { CambiosPendientesService } from '../../contexto/cambios-pendientes.service';
import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { LayoutStateService } from '../layout-state.service';
import { ThemeService } from '../theme.service';

@Component({
  selector: 'app-topbar',
  templateUrl: './topbar.component.html',
  styleUrl: './topbar.component.css',
})
export class TopbarComponent {
  readonly auth = inject(AuthService);
  readonly contexto = inject(ContextoService);
  readonly layout = inject(LayoutStateService);
  readonly theme = inject(ThemeService);
  private readonly feedback = inject(FeedbackService);
  private readonly cambiosPendientes = inject(CambiosPendientesService);
  private readonly router = inject(Router);

  readonly accountOpen = signal(false);
  readonly notificationsOpen = signal(false);
  readonly contextOpen = signal(false);

  toggleAccount(): void {
    this.notificationsOpen.set(false);
    this.contextOpen.set(false);
    this.accountOpen.update((open) => !open);
  }

  toggleNotifications(): void {
    this.accountOpen.set(false);
    this.contextOpen.set(false);
    this.notificationsOpen.update((open) => !open);
  }

  logout(): void {
    this.accountOpen.set(false);
    this.auth.logout().subscribe({
      next: () => void this.router.navigate(['/login']),
      error: () => void this.router.navigate(['/login']),
    });
    this.contexto.limpiar();
  }

  toggleContext(): void {
    this.accountOpen.set(false);
    this.notificationsOpen.set(false);
    this.contextOpen.update((open) => !open);
  }

  async cambiarNegocio(id: string): Promise<void> {
    if (id === this.contexto.negocio()?.id) {
      this.contextOpen.set(false);
      return;
    }
    if (!(await this.confirmarDescartarCambios())) return;
    this.contexto.seleccionarNegocio(id);
    this.contextOpen.set(false);
    const negocio = this.contexto.negocio();
    if (negocio) this.feedback.info('Negocio activo', negocio.nombre_comercial);
    // Se navega a inicio para no conservar vistas cargadas con el negocio anterior.
    void this.router.navigate(['/inicio']);
  }

  async cambiarSucursal(id: string): Promise<void> {
    if (id === this.contexto.sucursal()?.id) {
      this.contextOpen.set(false);
      return;
    }
    if (!(await this.confirmarDescartarCambios())) return;
    this.contexto.seleccionarSucursal(id);
    this.contextOpen.set(false);
  }

  /** Solicita confirmación una sola vez cuando algún componente declara trabajo sin guardar. */
  private async confirmarDescartarCambios(): Promise<boolean> {
    if (!this.cambiosPendientes.hayPendientes()) return true;
    return this.feedback.confirmDanger({
      titulo: 'Tienes cambios sin guardar',
      descripcion: 'Si cambias de contexto se perderán los datos que capturaste.',
      textoConfirmar: 'Sí, descartar',
      textoCancelar: 'Seguir editando',
    });
  }

  @HostListener('document:keydown.escape')
  closeMenus(): void {
    this.accountOpen.set(false);
    this.notificationsOpen.set(false);
    this.contextOpen.set(false);
  }
}
