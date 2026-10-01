import { formatDate } from '@angular/common';
import { Pipe, PipeTransform } from '@angular/core';

export const ZONA_HORARIA_MEXICO = 'America/Mexico_City';

/** Formatea instantes de la API en la hora civil de Ciudad de México. */
@Pipe({
  name: 'fechaMexico',
  standalone: true,
})
export class FechaMexicoPipe implements PipeTransform {
  transform(value: string | number | Date | null | undefined, format = 'medium'): string | null {
    if (value === null || value === undefined || value === '') return null;
    return formatDate(value, format, 'es-MX', ZONA_HORARIA_MEXICO);
  }
}
