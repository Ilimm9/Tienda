import { Component, OnInit } from '@angular/core';
import { FormBuilder, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { finalize } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  loading = false;
  error = '';
  success = '';
  readonly imageUrl = environment.loginHeaderImageUrl;
  readonly form = this.fb.group({
    correo: ['', [Validators.required, Validators.email]],
    contrasena: ['', Validators.required],
    recordarme: [false]
  });

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    if (this.route.snapshot.queryParamMap.get('registrado') === '1') {
      this.success = 'Cuenta creada. Ahora inicia sesión.';
    }
  }

  submit(): void {
    this.error = '';
    this.success = '';

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading = true;
    this.auth.login(this.form.value as {
      correo: string;
      contrasena: string;
      recordarme: boolean;
    }).pipe(finalize(() => this.loading = false)).subscribe({
      next: () => this.router.navigate(['/inicio']),
      error: response => this.error = response.error?.mensaje || 'No fue posible iniciar sesión.'
    });
  }
}
