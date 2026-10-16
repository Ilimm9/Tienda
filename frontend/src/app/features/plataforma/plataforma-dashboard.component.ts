import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { PlataformaResumen } from './plataforma.models';
import { PlataformaService } from './plataforma.service';

@Component({
  selector: 'app-plataforma-dashboard',
  imports: [DatePipe, DecimalPipe, RouterLink],
  templateUrl: './plataforma-dashboard.component.html',
  styleUrl: './plataforma.css',
})
export class PlataformaDashboardComponent implements OnInit {
  private readonly plataforma = inject(PlataformaService);
  readonly resumen = signal<PlataformaResumen | null>(null);
  readonly loading = signal(true);
  readonly error = signal('');

  ngOnInit(): void {
    this.plataforma.resumen().subscribe({
      next: (value) => { this.resumen.set(value); this.loading.set(false); },
      error: () => { this.error.set('No fue posible cargar las métricas de plataforma.'); this.loading.set(false); },
    });
  }
}
