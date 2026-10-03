import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';
import { of } from 'rxjs';

import { AuthService } from '../../features/auth/auth.service';
import { LayoutStateService } from '../layout-state.service';
import { TopbarComponent } from './topbar.component';

describe('TopbarComponent', () => {
  it('shows the current account and logs out through the global menu', async () => {
    const auth = {
      currentUser: signal({ id: '1', correo: 'persona@ejemplo.com' }),
      logout: vi.fn(() => of(undefined)),
    };
    const layout = {
      isMobile: signal(false),
      mobileMenuOpen: signal(false),
      sidebarCollapsed: signal(false),
      toggleNavigation: vi.fn(),
    };

    await TestBed.configureTestingModule({
      imports: [TopbarComponent],
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: auth },
        { provide: LayoutStateService, useValue: layout },
      ],
    }).compileComponents();
    const router = TestBed.inject(Router);
    vi.spyOn(router, 'navigate').mockResolvedValue(true);
    const fixture = TestBed.createComponent(TopbarComponent);
    fixture.componentInstance.toggleAccount();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('persona@ejemplo.com');
    // La inicial del correo sustituye al icono genérico de la cuenta.
    expect(fixture.nativeElement.querySelector('.account-avatar').textContent.trim()).toBe('P');
    // Las migas viven dentro de la topbar.
    expect(fixture.nativeElement.querySelector('app-breadcrumbs')).not.toBeNull();

    fixture.componentInstance.logout();
    expect(auth.logout).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });
});
