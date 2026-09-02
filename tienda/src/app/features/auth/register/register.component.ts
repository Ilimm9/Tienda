import { Component } from '@angular/core';
import { AbstractControl, FormBuilder, ValidationErrors, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { finalize } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { AuthService } from '../auth.service';

function matchingPasswords(control: AbstractControl): ValidationErrors | null {
  const password = control.get('contrasena');
  const confirmation = control.get('confirmarContrasena');
  if (!password || !confirmation || !confirmation.value) {
    return null;
  }
  return password.value === confirmation.value ? null : { passwordsMismatch: true };
}

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  loading = false;
  error = '';
  readonly imageUrl = environment.loginHeaderImageUrl;
  readonly form = this.fb.group({
    nombreCompleto: ['', [Validators.required, Validators.minLength(3)]],
    correo: ['', [Validators.required, Validators.email]],
    telefono: [''],
    contrasena: ['', [Validators.required, Validators.minLength(8)]],
    confirmarContrasena: ['', Validators.required]
  }, { validators: matchingPasswords });

  constructor(private fb: FormBuilder, private auth: AuthService, private router: Router) {}

  submit(): void {
    this.error = '';
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const value = this.form.value;
    this.loading = true;
    this.auth.register({
      nombre_completo: value.nombreCompleto as string,
      correo: value.correo as string,
      telefono: (value.telefono as string) || '',
      contrasena: value.contrasena as string
    }).pipe(finalize(() => this.loading = false)).subscribe({
      next: () => this.router.navigate(['/login'], { queryParams: { registrado: '1' } }),
      error: response => this.error = response.error?.mensaje || 'No fue posible crear la cuenta.'
    });
  }
}
