import { Component, DestroyRef, HostListener, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs';

import { AtajoNuevaVentaService } from '../../features/ventas/atajo-nueva-venta.service';
import { LayoutStateService } from '../layout-state.service';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { TopbarComponent } from '../topbar/topbar.component';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, SidebarComponent, TopbarComponent],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.css',
})
export class AppShellComponent {
  readonly layout = inject(LayoutStateService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly atajoVenta = inject(AtajoNuevaVentaService);

  constructor() {
    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => this.layout.closeMobileMenu());
  }

  @HostListener('document:keydown.f2', ['$event'])
  abrirNuevaVenta(evento: Event): void {
    this.atajoVenta.alPresionarF2(evento as KeyboardEvent);
  }

  @HostListener('document:keydown.escape')
  closeMobileMenu(): void {
    this.layout.closeMobileMenu();
  }
}
