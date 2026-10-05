import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnDestroy, computed, inject, signal } from '@angular/core';
import { NonNullableFormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';

import { ContextoService } from '../../contexto/contexto.service';
import { FeedbackService } from '../../shared/feedback/feedback.service';
import { FechaMexicoPipe } from '../../shared/fecha-mexico.pipe';
import { AuthService } from '../auth/auth.service';
import { DesafioRegistro, InvitacionAceptada, InvitacionPublica } from '../equipo/invitacion.models';
import { InvitacionService } from '../equipo/invitacion.service';
import { guardarRetornoInvitacion, olvidarRetornoInvitacion } from './retorno-invitacion';

type Paso = 'cuenta' | 'verificacion' | 'acceso';

const DESAFIO_KEY = 'tienda.invitacion.desafio';
const HORA_MS = 60 * 60 * 1000;

@Component({
  selector: 'app-aceptar-invitacion',
  imports: [FechaMexicoPipe, ReactiveFormsModule],
  templateUrl: './aceptar-invitacion.html',
  styleUrl: './aceptar-invitacion.css',
})
export class AceptarInvitacion implements OnDestroy {
  private readonly invitacionService = inject(InvitacionService);
  private readonly auth = inject(AuthService);
  private readonly contexto = inject(ContextoService);
  private readonly feedback = inject(FeedbackService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly formBuilder = inject(NonNullableFormBuilder);
  private timer?: ReturnType<typeof setInterval>;
  private desafioId = '';

  readonly token = this.route.snapshot.paramMap.get('token') ?? '';
  readonly invitacion = signal<InvitacionPublica | null>(null);
  readonly paso = signal<Paso>('cuenta');
  readonly sesionActiva = signal(false);
  readonly loading = signal(true);
  readonly processing = signal(false);
  readonly resending = signal(false);
  /** Error que impide continuar con este enlace: vencido, cancelado, ya aceptado o inexistente. */
  readonly bloqueo = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly correoDistinto = signal(false);
  readonly passwordVisible = signal(false);
  readonly secondsUntilResend = signal(0);

  readonly horasRestantes = computed(() => {
    const invitacion = this.invitacion();
    if (!invitacion) return 0;
    return Math.max(0, Math.ceil((new Date(invitacion.expira_en).getTime() - Date.now()) / HORA_MS));
  });

  readonly cuentaForm = this.formBuilder.group({
    nombres: ['', [Validators.required, Validators.maxLength(120)]],
    apellidos: ['', [Validators.required, Validators.maxLength(120)]],
    telefono: ['', Validators.maxLength(30)],
    contrasena: ['', [Validators.required, Validators.minLength(8), Validators.maxLength(72)]],
    confirmacion: ['', Validators.required],
  });
  readonly codigoForm = this.formBuilder.group({
    codigo: ['', [Validators.required, Validators.pattern(/^\d{6}$/)]],
  });

  constructor() {
    this.load();
  }

  ngOnDestroy(): void {
    if (this.timer) clearInterval(this.timer);
  }

  get contrasenasDistintas(): boolean {
    const { contrasena, confirmacion } = this.cuentaForm.getRawValue();
    return contrasena !== confirmacion;
  }

  togglePassword(): void {
    this.passwordVisible.update((visible) => !visible);
  }

  registrar(): void {
    this.error.set(null);
    if (this.cuentaForm.invalid || this.contrasenasDistintas) {
      this.cuentaForm.markAllAsTouched();
      return;
    }
    if (this.processing()) return;
    this.processing.set(true);
    const { nombres, apellidos, telefono, contrasena } = this.cuentaForm.getRawValue();
    this.invitacionService
      .registrar(this.token, {
        nombres: nombres.trim(),
        apellidos: apellidos.trim(),
        telefono: telefono.trim(),
        contrasena,
      })
      .subscribe({
        next: (desafio) => {
          this.processing.set(false);
          this.iniciarVerificacion(desafio);
        },
        error: (response: HttpErrorResponse) => {
          this.processing.set(false);
          const cuerpo = response.error as (Partial<DesafioRegistro> & { mensaje?: string }) | null;
          if (response.status === 503 && cuerpo?.desafio_id) {
            // La cuenta quedó pendiente: se pasa a verificación para poder pedir otro código.
            this.iniciarVerificacion(cuerpo as DesafioRegistro);
            this.error.set('No pudimos entregar el primer código. Espera el contador y solicita otro.');
            return;
          }
          if (response.status === 409 && (response.error as { codigo?: string })?.codigo === 'CUENTA_EXISTENTE') {
            this.invitacion.update((actual) => (actual ? { ...actual, requiere_cuenta: false } : actual));
            return;
          }
          if (this.esBloqueo(response)) return;
          this.error.set(cuerpo?.mensaje ?? 'No fue posible crear la cuenta. Intenta nuevamente.');
        },
      });
  }

  verificar(): void {
    this.error.set(null);
    if (this.codigoForm.invalid) {
      this.codigoForm.markAllAsTouched();
      return;
    }
    if (this.processing()) return;
    this.processing.set(true);
    this.auth
      .verifyEmail({ desafio_id: this.desafioId, codigo: this.codigoForm.controls.codigo.value })
      .subscribe({
        next: () => {
          this.olvidarDesafio();
          this.sesionActiva.set(true);
          this.confirmarAcceso();
        },
        error: (response: HttpErrorResponse) => {
          this.processing.set(false);
          this.error.set(response.error?.mensaje || 'No fue posible verificar el código.');
        },
      });
  }

  reenviarCodigo(): void {
    if (!this.desafioId || this.secondsUntilResend() > 0 || this.resending()) return;
    this.error.set(null);
    this.resending.set(true);
    this.auth.resendVerification({ desafio_id: this.desafioId }).subscribe({
      next: (response) => {
        this.resending.set(false);
        this.guardarDesafio(response.desafio_id);
        this.codigoForm.reset();
        this.startCountdown(response.reenviar_en_segundos);
        this.feedback.success('Código reenviado', 'Revisa nuevamente tu correo.');
      },
      error: (response: HttpErrorResponse) => {
        this.resending.set(false);
        this.error.set(response.error?.mensaje || 'No fue posible reenviar el código.');
        if (response.status === 429) this.startCountdown(60);
      },
    });
  }

  /** Confirma el acceso con la sesión actual; registro y aceptación son etapas separadas. */
  aceptar(): void {
    if (this.processing()) return;
    this.error.set(null);
    this.correoDistinto.set(false);
    this.processing.set(true);
    this.confirmarAcceso();
  }

  iniciarSesion(): void {
    guardarRetornoInvitacion(this.token);
    void this.router.navigate(['/login']);
  }

  /** Cierra la sesión de otro correo y lleva al login conservando el retorno a esta invitación. */
  cambiarSesion(): void {
    this.processing.set(true);
    this.auth.logout().subscribe({
      next: () => this.trasCerrarSesion(),
      error: () => this.trasCerrarSesion(),
    });
  }

  private trasCerrarSesion(): void {
    this.contexto.limpiar();
    this.iniciarSesion();
  }

  private confirmarAcceso(): void {
    this.paso.set('acceso');
    this.invitacionService.aceptar(this.token).subscribe({
      next: (aceptada) => this.entrar(aceptada),
      error: (response: HttpErrorResponse) => {
        this.processing.set(false);
        if (response.status === 401) {
          this.sesionActiva.set(false);
          this.error.set('Tu sesión terminó. Inicia sesión con el correo invitado.');
          return;
        }
        if ((response.error as { codigo?: string } | null)?.codigo === 'CORREO_NO_COINCIDE') {
          this.correoDistinto.set(true);
          return;
        }
        if (this.esBloqueo(response)) return;
        this.error.set(
          (response.error as { mensaje?: string } | null)?.mensaje ??
            'No fue posible confirmar tu acceso. Intenta nuevamente.',
        );
      },
    });
  }

  /** Recarga el contexto y selecciona explícitamente la empresa y sucursal de la invitación. */
  private entrar(aceptada: InvitacionAceptada): void {
    olvidarRetornoInvitacion();
    this.feedback.success('Invitación aceptada', 'Ya puedes operar en tu sucursal.');
    const abrirInicio = () => {
      this.contexto.seleccionarNegocio(aceptada.negocio_id);
      if (aceptada.sucursal_id) this.contexto.seleccionarSucursal(aceptada.sucursal_id);
      void this.router.navigate(['/inicio']);
    };
    this.contexto.recargar().subscribe({ next: abrirInicio, error: abrirInicio });
  }

  private esBloqueo(response: HttpErrorResponse): boolean {
    const mensaje = (response.error as { mensaje?: string } | null)?.mensaje;
    if (response.status === 404) this.bloqueo.set('Este enlace de invitación no existe.');
    else if (response.status === 410)
      this.bloqueo.set('Este enlace venció o fue cancelado. Pide a tu administrador que lo reenvíe.');
    else if (response.status === 409)
      this.bloqueo.set(mensaje ?? 'Esta invitación ya no puede aceptarse.');
    else if (response.status === 429)
      this.error.set('Demasiados intentos. Espera un momento e inténtalo de nuevo.');
    else return false;
    return true;
  }

  private iniciarVerificacion(desafio: DesafioRegistro): void {
    this.guardarDesafio(desafio.desafio_id);
    this.paso.set('verificacion');
    this.startCountdown(desafio.reenviar_en_segundos);
  }

  private guardarDesafio(desafioId: string): void {
    this.desafioId = desafioId;
    try {
      // Solo el desafío y un rastro del enlace; el token completo nunca se guarda aquí.
      sessionStorage.setItem(DESAFIO_KEY, JSON.stringify({ enlace: this.token.slice(-12), desafioId }));
    } catch {
      // Sin sessionStorage una recarga obliga a registrar de nuevo; el flujo sigue siendo válido.
    }
  }

  private olvidarDesafio(): void {
    this.desafioId = '';
    try {
      sessionStorage.removeItem(DESAFIO_KEY);
    } catch {
      // Nada que limpiar.
    }
  }

  /** Tras una recarga, retoma la verificación pendiente de este mismo enlace. */
  private desafioGuardado(): string {
    try {
      const guardado = JSON.parse(sessionStorage.getItem(DESAFIO_KEY) ?? 'null') as
        | { enlace?: string; desafioId?: string }
        | null;
      return guardado?.enlace === this.token.slice(-12) ? (guardado.desafioId ?? '') : '';
    } catch {
      return '';
    }
  }

  private startCountdown(seconds: number): void {
    if (this.timer) clearInterval(this.timer);
    this.secondsUntilResend.set(Math.max(0, seconds));
    this.timer = setInterval(() => {
      this.secondsUntilResend.update((current) => Math.max(0, current - 1));
      if (this.secondsUntilResend() === 0 && this.timer) {
        clearInterval(this.timer);
        this.timer = undefined;
      }
    }, 1000);
  }

  private load(): void {
    if (!this.token) {
      this.loading.set(false);
      this.bloqueo.set('El enlace de invitación no es válido.');
      return;
    }
    this.invitacionService.consultar(this.token).subscribe({
      next: (invitacion) => {
        this.invitacion.set(invitacion);
        const desafio = this.desafioGuardado();
        if (invitacion.requiere_cuenta && desafio) {
          this.desafioId = desafio;
          this.paso.set('verificacion');
        }
        // La sesión solo importa cuando la cuenta ya existe; un 401 aquí es un estado normal.
        this.auth
          .me()
          .pipe(
            map(() => true),
            catchError(() => of(false)),
          )
          .subscribe((activa) => {
            this.sesionActiva.set(activa);
            this.loading.set(false);
          });
      },
      error: (response: HttpErrorResponse) => {
        this.loading.set(false);
        if (!this.esBloqueo(response)) this.bloqueo.set('No fue posible validar la invitación.');
        else if (response.status === 429) this.bloqueo.set(this.error());
      },
    });
  }
}
