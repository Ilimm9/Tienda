import { DOCUMENT, isPlatformBrowser } from '@angular/common';
import { inject, Injectable, PLATFORM_ID, signal } from '@angular/core';

export type AppTheme = 'light' | 'dark';

/**
 * El tema oscuro está desactivado mientras se estabiliza la paleta clara: siempre se aplica
 * `light`, sin leer la preferencia del sistema ni valores guardados antes.
 */
@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly document = inject(DOCUMENT);
  private readonly platformId = inject(PLATFORM_ID);

  readonly theme = signal<AppTheme>('light');

  constructor() {
    if (!isPlatformBrowser(this.platformId)) return;
    this.document.documentElement.dataset['theme'] = 'light';
    this.document.documentElement.style.colorScheme = 'light';
  }
}
