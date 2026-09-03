# Documentación de Tienda

Aplicación web para gestión de una tienda. Incluye registro, inicio de sesión y un panel administrativo responsive con navegación protegida.

## Tecnologías

### Frontend

- Angular 22.1 con componentes standalone y detección de cambios zoneless.
- TypeScript 6, RxJS 7, signals y formularios reactivos tipados.
- Angular Router con carga diferida para navegación.
- Controles HTML nativos y PrimeIcons 4.1.

### Backend

- Go 1.25.
- Gin para la API HTTP.
- GORM y PostgreSQL para persistencia.
- JWT en cookies para sesiones.
- bcrypt para proteger contraseñas.

## Requisitos y configuración

Se necesita Node.js 24 LTS, npm, Go 1.25 o compatible y PostgreSQL ejecutándose localmente. La base de datos predeterminada es `tienda`.

Configurar las credenciales de PostgreSQL en `backend/.env` . Las variables más importantes son:

| Variable                  | Uso                 | Valor local             |
| ------------------------- | ------------------- | ----------------------- |
| `APP_PORT`                | Puerto de la API    | `8080`                  |
| `DB_HOST`                 | Servidor PostgreSQL | `localhost`             |
| `DB_PORT`                 | Puerto PostgreSQL   | `5433`                  |
| `DB_USER` / `DB_PASSWORD` | Credenciales        | `postgres` / `postgres` |
| `DB_NAME`                 | Base de datos       | `tienda`                |
| `JWT_SECRET`              | Firma de sesión     | Cambiar en producción   |
| `JWT_EXPIRATION`          | Duración de sesión  | `24h`                   |
| `FRONTEND_URL`            | Origen CORS         | `http://localhost:4200` |

### Backend

Desde la raíz:

```bash
cd backend
go run ./cmd/api
```

### Frontend

```bash
cd frontend
nvm use
npm start
```

La aplicación queda disponible en `http://localhost:4200`.

## Rutas y flujo

| Ruta                   | Función                         |
| ---------------------- | ------------------------------- |
| `/login`               | Inicio de sesión                |
| `/registro`            | Registro de una cuenta          |
| `/inicio`              | Página inicial protegida        |
| `/negocios`            | Página base de negocios         |
| `/sucursales`          | Página base de sucursales       |
| `/ventas`              | Página base de ventas           |
| `/equipo/empleados`    | Página base de empleados        |
| `/equipo/invitaciones` | Página base de invitaciones     |
| `/roles-permisos`      | Página base de roles y permisos |

El usuario se registra, vuelve a `/login` y, después de autenticarse, entra a `/inicio`. `authGuard` valida la sesión antes de permitir el acceso a cualquier ruta del panel. Login y registro usan el outlet raíz; las rutas protegidas se muestran en el outlet anidado del shell administrativo.

## Arquitectura

El proyecto está dividido en dos aplicaciones:

```text
Tienda/
├── frontend/                       # Frontend Angular 22
│   ├── src/app/features/auth/      # Login, registro, servicio y guard
│   ├── src/app/features/           # Home y áreas funcionales lazy-loaded
│   ├── src/app/layout/             # Shell, topbar, sidebar y breadcrumbs
│   ├── src/app/shared/ui/          # Componentes visuales reutilizables
│   ├── src/app/app.routes.ts
│   ├── src/app/app.config.ts
│   ├── src/environments/           # Configuración por entorno
│   └── src/styles.css              # Estilos globales
├── tienda/                         # Frontend Angular 12 conservado como referencia
└── backend/                        # API Go
    ├── cmd/api/                    # Punto de entrada y rutas
    └── internal/
        ├── interfaces/http/        # Handlers HTTP
        ├── application/            # Casos de uso
        ├── domain/                 # Entidades del dominio
        ├── infrastructure/         # Repositorios GORM
        ├── database/               # Conexión y migraciones
        └── config/                 # Configuración
```

El frontend se organiza por funcionalidades y usa APIs standalone de Angular. El menú lateral es compacto o expandido en escritorio y funciona como drawer en pantallas menores de 1024 px. El tema respeta inicialmente la preferencia del sistema y conserva la selección manual. Los breadcrumbs se generan desde los metadatos de las rutas. El backend separa la entrada HTTP, la lógica de negocio, el dominio y la persistencia. El registro crea usuario y perfil dentro de una transacción; la contraseña se guarda únicamente como hash bcrypt.

## Estilos y diseño visual

Los estilos globales se definen en `frontend/src/styles.css`. Ahí se encuentran:

- `--text`: color principal del texto.
- `--background`: fondo general.
- `--primary`: botones y enlaces principales.
- `--accent`: color de foco de los campos.
- Tipografía base `Inter`, con fuentes del sistema como fallback.
- Import de PrimeIcons.

## Responsive

Todas las vistas deben ser responsive y funcionar en escritorio, tablet y teléfono. Los nuevos componentes deben usar layouts fluidos con Flexbox o CSS Grid, `width: 100%`, `max-width`, media queries y controles adecuados para pantallas táctiles. Se debe evitar el scroll horizontal y adaptar tamaños, espaciados y columnas.

La pantalla de registro usa dos columnas en escritorio. En pantallas menores a `700px` oculta la imagen decorativa y conserva el formulario centrado para aprovechar el espacio disponible.

## Pruebas y compilación

- Frontend: `npm run build`.
- Pruebas frontend: `npm test -- --watch=false`.
- Backend: `go test ./...`.

## Consideraciones para producción

- Cambiar `JWT_SECRET` por una llave segura.
- Usar credenciales seguras para PostgreSQL.
- Configurar CORS y HTTPS correctamente.
- Revisar `APP_ENV`, `FRONTEND_URL` y SSL de PostgreSQL antes de publicar.
