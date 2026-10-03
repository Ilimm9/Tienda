import { provideHttpClient, withInterceptors, withXsrfConfiguration } from '@angular/common/http';
import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZoneChangeDetection } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter } from '@angular/router';
import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';

import { authInterceptor } from './features/auth/auth.interceptor';
import { routes } from './app.routes';

// Aura con la escala primaria en violeta (opción G): tablas, paginador, selects y foco coinciden con styles.css.
// El resaltado es lavanda suave para que la página activa y las filas elegidas no sean bloques sólidos.
const TemaTienda = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#f5f3fd',
      100: '#eceafb',
      200: '#dcd8f6',
      300: '#bdb6ef',
      400: '#8f86e0',
      500: '#6a61d4',
      600: '#4f46c8',
      700: '#3f37a8',
      800: '#322c85',
      900: '#26215f',
      950: '#171439',
    },
    focusRing: {
      width: '3px',
      style: 'solid',
      color: '{primary.200}',
      offset: '1px',
    },
    formField: {
      borderRadius: '10px',
      paddingX: '0.875rem',
      paddingY: '0.7rem',
    },
    colorScheme: {
      light: {
        primary: {
          color: '{primary.600}',
          contrastColor: '#ffffff',
          hoverColor: '{primary.700}',
          activeColor: '{primary.800}',
        },
        highlight: {
          background: '{primary.100}',
          focusBackground: '{primary.200}',
          color: '{primary.700}',
          focusColor: '{primary.800}',
        },
        formField: {
          borderColor: '#cfc8b8',
          hoverBorderColor: '#5b6478',
          focusBorderColor: '{primary.600}',
          color: '#172554',
          placeholderColor: '#8a90a0',
        },
      },
    },
  },
});

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes),
    provideHttpClient(
      withXsrfConfiguration({ cookieName: 'XSRF-TOKEN', headerName: 'X-XSRF-TOKEN' }),
      withInterceptors([authInterceptor]),
    ),
    provideAnimationsAsync(),
    providePrimeNG({
      license: 'eyJpZCI6IjljM2JjZWRmLWZmYjMtNDVhMC04NzNhLTRjMTRjNzUwOTYyYSIsInByb2R1Y3QiOiJwcmltZXVpIiwidGllciI6ImNvbW11bml0eSIsInR5cGUiOiJkZXYiLCJpYXQiOjE3ODkwNzIzNjYsImV4cCI6MTgyMDYwODM2Nn0.SlEtFpXk4FXnlfVrdQ4myvNICX4ix5crMKr29h6W2duqf9hMMb-31XJAlPyaAiAmsp2wkWLehCHb0YjETE8zAA',
      overlayAppendTo: 'body',
      theme: {
        preset: TemaTienda,
        options: {
          darkModeSelector: '[data-theme="dark"]',
        },
      },
    }),
  ],
};
