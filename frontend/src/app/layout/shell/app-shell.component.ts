import { Component, DestroyRef, HostListener, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs';

import { BreadcrumbsComponent } from '../breadcrumbs/breadcrumbs.component';
import { LayoutStateService } from '../layout-state.service';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { TopbarComponent } from '../topbar/topbar.component';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, BreadcrumbsComponent, SidebarComponent, TopbarComponent],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.css',
})
export class AppShellComponent {
  readonly layout = inject(LayoutStateService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  constructor() {
    this.router.events
      .pipe(
        filter((event): event is NavigationEnd => event instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => this.layout.closeMobileMenu());
  }

  @HostListener('document:keydown.escape')
  closeMobileMenu(): void {
    this.layout.closeMobileMenu();
  }
}
