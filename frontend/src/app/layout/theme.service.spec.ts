import { TestBed } from '@angular/core/testing';

import { ThemeService } from './theme.service';

describe('ThemeService', () => {
  beforeEach(() => vi.stubGlobal('localStorage', storage()));

  afterEach(() => {
    TestBed.resetTestingModule();
    document.documentElement.removeAttribute('data-theme');
    document.documentElement.style.removeProperty('color-scheme');
    vi.unstubAllGlobals();
  });

  it('uses the system theme initially and persists a manual change', () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => mediaQuery(true)),
    );
    const service = TestBed.inject(ThemeService);

    expect(service.theme()).toBe('dark');
    expect(document.documentElement.dataset['theme']).toBe('dark');

    service.toggle();

    expect(service.theme()).toBe('light');
    expect(localStorage.getItem('tienda.theme')).toBe('light');
  });

  it('prefers a saved theme over the operating system', () => {
    localStorage.setItem('tienda.theme', 'dark');
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => mediaQuery(false)),
    );

    expect(TestBed.inject(ThemeService).theme()).toBe('dark');
  });
});

function mediaQuery(matches: boolean): MediaQueryList {
  return {
    matches,
    media: '',
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  };
}

function storage(): Storage {
  const values = new Map<string, string>();
  return {
    get length() {
      return values.size;
    },
    clear: () => values.clear(),
    getItem: (key) => values.get(key) ?? null,
    key: (index) => [...values.keys()][index] ?? null,
    removeItem: (key) => void values.delete(key),
    setItem: (key, value) => void values.set(key, value),
  };
}
