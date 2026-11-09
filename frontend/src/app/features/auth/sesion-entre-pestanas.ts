import { inject } from '@angular/core';
import { Router } from '@angular/router';

import { ContextoService } from '../../contexto/contexto.service';
import { AuthService, destinoTrasSesion } from './auth.service';

// Pantallas protegidas por invitadoGuard: si otra pestaña inicia sesión, esta pasa a la app.
const rutasDeInvitado = ['/login', '/registro', '/recuperar-contrasena'];
// Pantallas que no requieren sesión: un cierre en otra pestaña no las interrumpe.
const rutasPublicas = [...rutasDeInvitado, '/restablecer-contrasena', '/verificar-correo', '/invitacion/'];

function rutaActual(router: Router): string {
  return router.url.split(/[?#]/)[0];
}

/** Mantiene todas las pestañas del navegador alineadas con la sesión (cookie compartida). */
export function sincronizarSesionEntrePestanas(): void {
  const auth = inject(AuthService);
  const router = inject(Router);
  const contexto = inject(ContextoService);

  auth.sesionEnOtraPestana.subscribe((aviso) => {
    const ruta = rutaActual(router);
    if (aviso === 'cierre') {
      contexto.limpiar();
      if (!rutasPublicas.some((publica) => ruta === publica || ruta.startsWith(publica))) {
        void router.navigate(['/login'], { replaceUrl: true });
      }
      return;
    }
    if (!rutasDeInvitado.includes(ruta)) return;
    auth.asegurarSesion().subscribe({
      next: (usuario) => void router.navigateByUrl(destinoTrasSesion(usuario), { replaceUrl: true }),
      // Sin sesión válida la pantalla de acceso sigue disponible.
      error: () => undefined,
    });
  });
}
