# Frontend de Tienda

Frontend Angular 22 para registro, inicio de sesión y página inicial protegida.

## Requisitos

- Node.js 24 LTS mediante NVM.
- Backend disponible en `http://localhost:8080`.

## Desarrollo

```bash
nvm use
npm install
npm start
```

Abrir `http://localhost:4200`.

## Verificación

```bash
npm test -- --watch=false
npm run build
```

La configuración de desarrollo consume `http://localhost:8080/api/v1`. La compilación de producción usa `/api/v1`.
