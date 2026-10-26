import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { CuentaPlataformaResumen } from './plataforma.models';
import { PlataformaService } from './plataforma.service';

@Component({
  selector: 'app-plataforma-cuentas',
  imports: [DatePipe, DecimalPipe, FormsModule],
  templateUrl: './plataforma-cuentas.component.html',
  styleUrl: './plataforma.css',
})
export class PlataformaCuentasComponent implements OnInit {
  private readonly plataforma = inject(PlataformaService);
  readonly items = signal<CuentaPlataformaResumen[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 20;
  readonly loading = signal(true);
  readonly error = signal('');
  query = '';
  estado = '';

  ngOnInit(): void { this.load(); }

  buscar(): void { this.page.set(1); this.load(); }
  anterior(): void { if (this.page() > 1) { this.page.update((value) => value - 1); this.load(); } }
  siguiente(): void { if (this.page() * this.pageSize < this.total()) { this.page.update((value) => value + 1); this.load(); } }

  private load(): void {
    this.loading.set(true);
    this.error.set('');
    this.plataforma.cuentas(this.query, this.estado, this.page(), this.pageSize).subscribe({
      next: (result) => { this.items.set(result.items ?? []); this.total.set(result.total); this.loading.set(false); },
      error: () => { this.error.set('No fue posible cargar las cuentas.'); this.loading.set(false); },
    });
  }
}
