import { Component, HostListener, inject, signal } from '@angular/core';
import { Router } from '@angular/router';

import { AuthService } from '../../features/auth/auth.service';
import { ContextoService } from '../../contexto/contexto.service';
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

  cambiarNegocio(id: string): void {
    this.contexto.seleccionarNegocio(id);
    void this.router.navigate(['/inicio']);
  }

  cambiarSucursal(id: string): void {
    this.contexto.seleccionarSucursal(id);
    this.contextOpen.set(false);
  }

  @HostListener('document:keydown.escape')
  closeMenus(): void {
    this.accountOpen.set(false);
    this.notificationsOpen.set(false);
    this.contextOpen.set(false);
  }
}
