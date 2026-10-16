import { Component, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { AuthService } from '../auth/auth.service';

@Component({
  selector: 'app-plataforma-shell',
  imports: [RouterLink, RouterLinkActive, RouterOutlet],
  templateUrl: './plataforma-shell.component.html',
  styleUrl: './plataforma.css',
})
export class PlataformaShellComponent {
  readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  logout(): void {
    this.auth.logout().subscribe({
      next: () => void this.router.navigate(['/login']),
      error: () => void this.router.navigate(['/login']),
    });
  }
}
