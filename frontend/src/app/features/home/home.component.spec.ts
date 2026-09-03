import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { HomeComponent } from './home.component';

describe('HomeComponent', () => {
  it('logs out before returning to login', async () => {
    const auth = { logout: vi.fn(() => of(undefined)) };
    const router = { navigate: vi.fn(() => Promise.resolve(true)) };
    await TestBed.configureTestingModule({
      imports: [HomeComponent],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
      ],
    }).compileComponents();
    const component = TestBed.createComponent(HomeComponent).componentInstance;

    component.logout();

    expect(auth.logout).toHaveBeenCalledOnce();
    expect(router.navigate).toHaveBeenCalledWith(['/login']);
  });
});
