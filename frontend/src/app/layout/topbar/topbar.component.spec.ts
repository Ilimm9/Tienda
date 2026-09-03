import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';

import { AuthService } from '../../features/auth/auth.service';
import { LayoutStateService } from '../layout-state.service';
import { ThemeService } from '../theme.service';
import { TopbarComponent } from './topbar.component';

describe('TopbarComponent', () => {
  it('shows the current account and logs out through the global menu', async () => {
    const auth = {
      currentUser: signal({ id: '1', correo: 'persona@ejemplo.com' }),
      logout: vi.fn(() => of(undefined)),
    };
    const router = { navigate: vi.fn(() => Promise.resolve(true)) };
    const layout = {
      isMobile: signal(false),
      mobileMenuOpen: signal(false),
      sidebarCollapsed: signal(false),
      toggleNavigation: vi.fn(),
    };
    const theme = { theme: signal('light'), toggle: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [TopbarComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        { provide: LayoutStateService, useValue: layout },
        { provide: ThemeService, useValue: theme },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(TopbarComponent);
    fixture.componentInstance.toggleAccount();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('persona@ejemplo.com');

    fixture.componentInstance.logout();
    expect(auth.logout).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });
});
