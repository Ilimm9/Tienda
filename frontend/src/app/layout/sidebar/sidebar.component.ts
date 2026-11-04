import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';

import { LayoutStateService } from '../layout-state.service';
import { ContextoService } from '../../contexto/contexto.service';
import { VistasService } from '../../contexto/vistas.service';
import { filtrarNavegacion, NAVIGATION_ITEMS, NavigationItem } from '../navigation.config';

@Component({
  selector: 'app-sidebar',
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './sidebar.component.html',
  styleUrl: './sidebar.component.css',
})
export class SidebarComponent {
  readonly layout = inject(LayoutStateService);
  private readonly router = inject(Router);
  private readonly expandedGroups = signal(new Set<string>());

  private readonly contexto = inject(ContextoService);
  private readonly vistas = inject(VistasService);

  /**
   * El menú refleja los permisos y las vistas del negocio activo; cambia al cambiar de negocio,
   * recargar el contexto o guardar Configuración > Vistas.
   */
  readonly navigationItems = computed(() =>
    filtrarNavegacion(
      NAVIGATION_ITEMS,
      (item) =>
        this.contexto.puede(item.permiso) &&
        (item.acceso !== 'ver-negocios' || this.contexto.puedeVerNegocios()) &&
        (!item.soloPropietario || this.contexto.negocio()?.tipo_miembro === 'propietario') &&
        this.vistas.visible(item.vista),
    ),
  );

  isGroupExpanded(item: NavigationItem): boolean {
    return this.isGroupActive(item) || this.expandedGroups().has(item.label);
  }

  isGroupActive(item: NavigationItem): boolean {
    return item.children?.some((child) => this.isRouteActive(child.route)) ?? false;
  }

  toggleGroup(item: NavigationItem): void {
    if (this.layout.sidebarCollapsed() && !this.layout.isMobile()) {
      this.layout.toggleNavigation();
    }

    this.expandedGroups.update((current) => {
      const next = new Set(current);
      next.has(item.label) ? next.delete(item.label) : next.add(item.label);
      return next;
    });
  }

  closeAfterNavigation(): void {
    this.layout.closeMobileMenu();
  }

  private isRouteActive(route?: string): boolean {
    if (!route) return false;
    const currentPath = this.router.url.split(/[?#]/, 1)[0];
    return currentPath === route || currentPath.startsWith(`${route}/`);
  }
}
