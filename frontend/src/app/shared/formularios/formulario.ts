import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

/** Lleva el foco al primer campo marcado como inválido dentro de `host`; ayuda a quien no ve el error. */
export function enfocarPrimerInvalido(host: HTMLElement): void {
  // Los errores se pintan en el siguiente ciclo de detección: se espera a que `aria-invalid` exista.
  queueMicrotask(() =>
    setTimeout(() => host.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus()),
  );
}

/** Rechaza un valor formado solo por espacios en un campo obligatorio. */
export const sinSoloEspacios: ValidatorFn = (control: AbstractControl): ValidationErrors | null =>
  typeof control.value === 'string' && control.value.length > 0 && control.value.trim() === ''
    ? { soloEspacios: true }
    : null;

/** Fecha `AAAA-MM-DD` que no puede ser posterior a hoy. */
export const fechaNoFutura: ValidatorFn = (control: AbstractControl): ValidationErrors | null => {
  const valor = control.value as string | null;
  if (!valor) return null;
  const hoy = new Date();
  const limite = `${hoy.getFullYear()}-${String(hoy.getMonth() + 1).padStart(2, '0')}-${String(hoy.getDate()).padStart(2, '0')}`;
  return valor > limite ? { fechaFutura: true } : null;
};
