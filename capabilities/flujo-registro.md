# Flujo de registro

Base vigente según código local. Actualizado: 2026-10-19.

## Flujo y reglas

1. Registrar nombres, primer apellido, segundo apellido opcional, correo y contraseña. El correo se normaliza a minúsculas; una cuenta ya registrada no se duplica.
2. Crear cuenta `pendiente_verificacion` y enviar OTP por correo. El registro responde `202` sin abrir sesión; verificar un código vigente activa la cuenta.
3. Iniciar sesión con correo verificado. La sesión usa cookie HttpOnly y protección CSRF; la recuperación cambia contraseña y revoca sesiones previas.
4. Sin negocio, el alta inicial guía para crear empresa y primera sucursal. La cuenta propietaria obtiene membresía; negocio y sucursal activos forman el contexto de operación.
5. Quien llega por invitación usa el correo de esa invitación, verifica su cuenta y la acepta para entrar al negocio y sucursal indicados. Ver [colaboradores](gestion-colaboradores.md).

La contraseña exige mayúscula, número y símbolo; longitud validada entre 8 y 72 bytes. OTP y credenciales de sesión se guardan como hashes; vigencia, intentos y reenvíos dependen de configuración. No confundir crear una cuenta con crear un negocio.

Sesión en el frontend (2026-11-03): los guards consultan `/auth/me` una sola vez y reutilizan el usuario en memoria al navegar; el backend sigue validando la sesión en cada petición. Un `401` de cualquier endpoint (salvo `/auth/*` e `/invitaciones/*`, que lo manejan sus pantallas) con sesión en memoria limpia el usuario y lleva al login. El contexto (`/contexto/opciones`) se guarda entre navegaciones y se revalida al volver a la pestaña solo si pasó más de un minuto; un cambio de contexto en otra pestaña se revalida de inmediato.

Sesión activa (2026-11-08): `invitadoGuard` protege `/login`, `/registro` y `/recuperar-contrasena`. Con sesión válida, en esta u otra pestaña, lleva a `/inicio` (o `/administracion-plataforma` para `ADMIN_LECTURA`) hasta cerrar sesión. `/restablecer-contrasena`, `/verificar-correo` e `/invitacion/:token` quedan abiertas porque llegan por enlaces de correo. Login y verificación navegan con `replaceUrl`, así que Atrás no regresa al formulario. Las pestañas se avisan por `BroadcastChannel('tienda.sesion')`: un cierre limpia usuario y contexto y lleva al login (salvo en pantallas públicas); un inicio saca de las pantallas de visitante. Sin `BroadcastChannel`, la pestaña lo detecta con el siguiente `401` o al navegar.

## Tablas principales

| Tabla | Uso |
| --- | --- |
| `usuarios` | Correo único, hash de contraseña, estado y verificación. |
| `perfil_usuarios` | Nombres, apellidos y teléfono; vínculo por `usuario_id`. |
| `desafio_autenticacions` | OTP de verificación o recuperación: propósito, hash, vencimiento, intentos y consumo. |
| `sesion_usuarios` | Sesiones, hashes de token/CSRF, vencimiento, actividad y revocación. |
| `negocios`, `membresias_negocio` | Empresa y acceso de la cuenta como propietario o miembro. |
| `sucursales`, `direcciones`, `configuraciones_negocio` | Establecimientos, ubicación y configuración del negocio. |

Los nombres poco habituales de las tablas de cuenta son los generados actualmente por GORM.

## Dónde cambiar y qué revisar

- Backend: [cuenta](../backend/internal/application/cuenta/), [negocio](../backend/internal/application/negocio/) y [rutas HTTP](../backend/cmd/api/main.go).
- Frontend: [autenticación](../frontend/src/app/features/auth/), [alta inicial](../frontend/src/app/features/alta-inicial/) y [contexto](../frontend/src/app/contexto/).
- Comprobar correo duplicado, OTP vencido/reutilizado, recuperación, revocación de sesión y alta inicial sin negocio; pendientes en [planes actuales](planes-actuales.md).
- Antecedentes: [seguridad](historial/06-preparacion-produccion-segura.md), [correo](historial/07-servicio-correo-transaccional.md), [acceso](historial/13-acceso-animado-stockion.md) y [alta inicial](historial/01-thrs-alta-inicial-empresa-sucursal.md).
