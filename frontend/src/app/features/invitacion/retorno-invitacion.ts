const RETORNO_KEY = 'tienda.invitacion.retorno';
// Solo se acepta la ruta interna de una invitación: nunca una redirección externa.
const RUTA_INVITACION = /^\/invitacion\/[A-Za-z0-9_-]{20,200}$/;

/** Recuerda a qué invitación volver después de iniciar sesión o verificar el correo. */
export function guardarRetornoInvitacion(token: string): void {
  const ruta = `/invitacion/${token}`;
  if (!RUTA_INVITACION.test(ruta)) return;
  try {
    sessionStorage.setItem(RETORNO_KEY, ruta);
  } catch {
    // Sin sessionStorage el invitado vuelve a abrir su enlace; el flujo sigue siendo válido.
  }
}

/** Devuelve el retorno pendiente una sola vez; cualquier valor alterado se descarta. */
export function consumirRetornoInvitacion(): string | null {
  try {
    const ruta = sessionStorage.getItem(RETORNO_KEY);
    sessionStorage.removeItem(RETORNO_KEY);
    return ruta && RUTA_INVITACION.test(ruta) ? ruta : null;
  } catch {
    return null;
  }
}

export function olvidarRetornoInvitacion(): void {
  try {
    sessionStorage.removeItem(RETORNO_KEY);
  } catch {
    // Nada que limpiar.
  }
}
