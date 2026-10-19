# Guía operativa del administrador global

## 1. Crear la cuenta

1. Registrar una cuenta normal desde `/registro`.
2. Verificar el correo e iniciar sesión al menos una vez.
3. No crear una tienda ni aceptar invitaciones para esa cuenta. El comando rechazará cuentas con membresías de negocio.

## 2. Asignar el perfil

Desde el entorno seguro donde está configurada `DATABASE_URL`:

```bash
cd backend
go run ./cmd/admin-plataforma grant \
  --email admin@ejemplo.com \
  --reason "Administración inicial de la plataforma"
```

La operación es idempotente. El comando valida que la cuenta esté activa, verificada y sin membresías, asigna `ADMIN_LECTURA` y guarda auditoría.

Para consultar administradores:

```bash
go run ./cmd/admin-plataforma list
```

Para retirar el perfil:

```bash
go run ./cmd/admin-plataforma revoke \
  --email admin@ejemplo.com \
  --reason "Rotación de responsabilidades"
```

## 3. Entrar al panel

El administrador inicia sesión normalmente. Si no tiene tiendas, el login lo dirige a `/administracion-plataforma`; no verá el asistente de alta inicial.

El panel contiene:

- Resumen de tiendas activas/archivadas, sucursales, cuentas y usuarios activos en 7/30 días.
- Directorio de tiendas con búsqueda, estado, sucursales, personal, cuentas vinculadas y última actividad.
- Detalle de cada tienda con usuarios, roles y sucursales asignadas.
- Tabla por sucursal con `personal asignado` y `con cuenta activa`.

## 4. Límites intencionales

- El perfil es de solo lectura.
- No puede crear tiendas, aceptar invitaciones, editar usuarios ni consumir las rutas tenant.
- La actividad se atribuye al negocio, no a una sucursal.
- No se muestran IP, agente del navegador ni contenido de solicitudes.

## 5. Recuperación

Si la cuenta pierde acceso, un operador con acceso al mismo entorno ejecuta `revoke` y después `grant` con un motivo nuevo. Las acciones quedan en `auditoria_plataforma`.
