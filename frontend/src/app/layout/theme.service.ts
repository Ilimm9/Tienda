import { DOCUMENT, isPlatformBrowser } from '@angular/common';
import { inject, Injectable, PLATFORM_ID, signal } from '@angular/core';

export type AppTheme = 'light' | 'dark';

const THEME_STORAGE_KEY = 'tienda.theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly document = inject(DOCUMENT);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly storage = typeof localStorage === 'undefined' ? undefined : localStorage;
  private mediaQuery?: MediaQueryList;
  private hasManualPreference = false;

  readonly theme = signal<AppTheme>('light');

  constructor() {
    if (!isPlatformBrowser(this.platformId)) return;

    const savedTheme = this.storage?.getItem(THEME_STORAGE_KEY);
    this.hasManualPreference = savedTheme === 'light' || savedTheme === 'dark';
    this.mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)');
    this.applyTheme(this.hasManualPreference ? (savedTheme as AppTheme) : this.systemTheme());
    this.mediaQuery?.addEventListener('change', this.handleSystemThemeChange);
  }

  toggle(): void {
    const nextTheme: AppTheme = this.theme() === 'light' ? 'dark' : 'light';
    this.hasManualPreference = true;
    this.storage?.setItem(THEME_STORAGE_KEY, nextTheme);
    this.applyTheme(nextTheme);
  }

  private systemTheme(): AppTheme {
    return this.mediaQuery?.matches ? 'dark' : 'light';
  }

  private applyTheme(theme: AppTheme): void {
    this.theme.set(theme);
    this.document.documentElement.dataset['theme'] = theme;
    this.document.documentElement.style.colorScheme = theme;
  }

  private readonly handleSystemThemeChange = (): void => {
    if (!this.hasManualPreference) this.applyTheme(this.systemTheme());
  };
}
