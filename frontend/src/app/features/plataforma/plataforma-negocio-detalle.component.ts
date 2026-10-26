import { DatePipe, DecimalPipe } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { NegocioPlataformaDetalle, UsuarioPlataformaResumen } from './plataforma.models';
import { PlataformaService } from './plataforma.service';

@Component({
  selector: 'app-plataforma-negocio-detalle',
  imports: [DatePipe, DecimalPipe, RouterLink],
  templateUrl: './plataforma-negocio-detalle.component.html',
  styleUrl: './plataforma.css',
})
export class PlataformaNegocioDetalleComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly plataforma = inject(PlataformaService);
  readonly negocio = signal<NegocioPlataformaDetalle | null>(null);
  readonly usuarios = signal<UsuarioPlataformaResumen[]>([]);
  readonly loading = signal(true);
  readonly error = signal('');

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('negocioId');
    if (!id) { this.error.set('Identificador de tienda no disponible.'); this.loading.set(false); return; }
    forkJoin({ negocio: this.plataforma.negocio(id), usuarios: this.plataforma.usuarios(id) }).subscribe({
      next: ({ negocio, usuarios }) => { this.negocio.set(negocio); this.usuarios.set(usuarios.items ?? []); this.loading.set(false); },
      error: () => { this.error.set('No fue posible cargar el detalle de la tienda.'); this.loading.set(false); },
    });
  }
}
