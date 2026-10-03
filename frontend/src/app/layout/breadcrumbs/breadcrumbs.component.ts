import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRouteSnapshot, NavigationEnd, Router, RouterLink } from '@angular/router';
import { filter } from 'rxjs';

import { NAVIGATION_ITEMS, NavigationItem } from '../navigation.config';

export interface BreadcrumbItem {
  readonly label: string;
  /** Sin URL, el nivel se muestra como texto: un grupo del menú no tiene página propia. */
  readonly url: string | null;
}

interface NavigationMatch {
  readonly item: NavigationItem;
  readonly group: NavigationItem | null;
}

@Component({
  selector: 'app-breadcrumbs',
  imports: [RouterLink],
  templateUrl: './breadcrumbs.component.html',
  styleUrl: './breadcrumbs.component.css',
})
export class BreadcrumbsComponent {
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  readonly items = signal<readonly BreadcrumbItem[]>([]);
  /** En pantallas angostas solo cabe el nivel actual. */
  readonly current = computed(() => this.items().at(-1) ?? null);

  constructor() {
    this.updateBreadcrumbs();
    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => this.updateBreadcrumbs());
  }

  private updateBreadcrumbs(): void {
    const fromRoutes: BreadcrumbItem[] = [];
    let currentRoute: ActivatedRouteSnapshot | null = this.router.routerState.snapshot.root;
    let currentUrl = '';

    while (currentRoute?.firstChild) {
      currentRoute = currentRoute.firstChild;
      const routePath = currentRoute.url.map((segment) => segment.path).join('/');
      if (routePath) currentUrl += `/${routePath}`;

      const label = currentRoute.routeConfig?.data?.['breadcrumb'];
      if (typeof label !== 'string' || !label) continue;

      const url = currentUrl || '/';
      // Un padre y su ruta vacía comparten URL (Roles y permisos, Productos): queda un solo nivel.
      if (fromRoutes.at(-1)?.url === url) {
        fromRoutes[fromRoutes.length - 1] = { label, url };
      } else {
        fromRoutes.push({ label, url });
      }
    }

    this.items.set(this.withNavigation(fromRoutes, currentUrl || '/'));
  }

  /**
   * Antepone el grupo y la sección del menú cuando las rutas no los declaran.
   * Así «Agregar marca» queda como «Catálogo / Marcas / Agregar marca».
   */
  private withNavigation(fromRoutes: BreadcrumbItem[], url: string): BreadcrumbItem[] {
    const match = this.findNavigation(url);
    if (!match?.item.route) return fromRoutes;

    const sectionUrl = match.item.route;
    const descendants = fromRoutes.filter((item) => item.url !== null && this.isDescendant(item.url, sectionUrl));
    const trail: BreadcrumbItem[] = [];
    if (match.group) trail.push({ label: match.group.label, url: null });
    trail.push({ label: match.item.label, url: sectionUrl });
    return [...trail, ...descendants];
  }

  private findNavigation(url: string): NavigationMatch | null {
    let best: NavigationMatch | null = null;
    const consider = (item: NavigationItem, group: NavigationItem | null) => {
      if (!item.route || !this.isWithin(url, item.route)) return;
      if (!best || item.route.length > (best.item.route?.length ?? 0)) best = { item, group };
    };

    for (const item of NAVIGATION_ITEMS) {
      consider(item, null);
      for (const child of item.children ?? []) consider(child, item);
    }
    return best;
  }

  private isWithin(url: string, route: string): boolean {
    return url === route || url.startsWith(`${route}/`);
  }

  private isDescendant(url: string, route: string): boolean {
    return url.startsWith(`${route}/`);
  }
}
