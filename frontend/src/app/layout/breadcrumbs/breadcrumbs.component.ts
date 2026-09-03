import { Component, DestroyRef, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRouteSnapshot, NavigationEnd, Router, RouterLink } from '@angular/router';
import { filter } from 'rxjs';

interface BreadcrumbItem {
  readonly label: string;
  readonly url: string;
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
    const breadcrumbs: BreadcrumbItem[] = [];
    let currentRoute: ActivatedRouteSnapshot | null = this.router.routerState.snapshot.root;
    let currentUrl = '';

    while (currentRoute?.firstChild) {
      currentRoute = currentRoute.firstChild;
      const routePath = currentRoute.url.map((segment) => segment.path).join('/');
      if (routePath) currentUrl += `/${routePath}`;

      const label = currentRoute.data['breadcrumb'];
      if (typeof label === 'string' && label) {
        breadcrumbs.push({ label, url: currentUrl || '/' });
      }
    }

    this.items.set(breadcrumbs);
  }
}
