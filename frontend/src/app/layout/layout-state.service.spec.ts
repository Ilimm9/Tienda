import { TestBed } from '@angular/core/testing';

import { LayoutStateService } from './layout-state.service';

describe('LayoutStateService', () => {
  beforeEach(() => vi.stubGlobal('localStorage', storage()));

  afterEach(() => {
    TestBed.resetTestingModule();
    vi.unstubAllGlobals();
  });

  it('persists the compact desktop sidebar preference', () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => mediaQuery(false)),
    );
    const service = TestBed.inject(LayoutStateService);

    service.toggleNavigation();

    expect(service.sidebarCollapsed()).toBe(true);
    expect(localStorage.getItem('tienda.sidebar.collapsed')).toBe('true');
  });

  it('opens an overlay menu on mobile without changing the desktop preference', () => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => mediaQuery(true)),
    );
    const service = TestBed.inject(LayoutStateService);

    service.toggleNavigation();

    expect(service.mobileMenuOpen()).toBe(true);
    expect(service.sidebarCollapsed()).toBe(false);
    expect(localStorage.getItem('tienda.sidebar.collapsed')).toBeNull();
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
