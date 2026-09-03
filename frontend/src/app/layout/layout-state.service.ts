import { isPlatformBrowser } from '@angular/common';
import { inject, Injectable, PLATFORM_ID, signal } from '@angular/core';

const SIDEBAR_STORAGE_KEY = 'tienda.sidebar.collapsed';
const MOBILE_QUERY = '(max-width: 1023px)';

@Injectable({ providedIn: 'root' })
export class LayoutStateService {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly storage = typeof localStorage === 'undefined' ? undefined : localStorage;
  private mediaQuery?: MediaQueryList;

  readonly sidebarCollapsed = signal(false);
  readonly mobileMenuOpen = signal(false);
  readonly isMobile = signal(false);

  constructor() {
    if (!isPlatformBrowser(this.platformId)) return;

    this.sidebarCollapsed.set(this.storage?.getItem(SIDEBAR_STORAGE_KEY) === 'true');
    this.mediaQuery = window.matchMedia?.(MOBILE_QUERY);
    if (!this.mediaQuery) return;
    this.isMobile.set(this.mediaQuery.matches);
    this.mediaQuery.addEventListener('change', this.handleBreakpointChange);
  }

  toggleNavigation(): void {
    if (this.isMobile()) {
      this.mobileMenuOpen.update((open) => !open);
      return;
    }

    this.sidebarCollapsed.update((collapsed) => !collapsed);
    this.storage?.setItem(SIDEBAR_STORAGE_KEY, String(this.sidebarCollapsed()));
  }

  closeMobileMenu(): void {
    this.mobileMenuOpen.set(false);
  }

  private readonly handleBreakpointChange = (event: MediaQueryListEvent): void => {
    this.isMobile.set(event.matches);
    if (!event.matches) this.mobileMenuOpen.set(false);
  };
}
