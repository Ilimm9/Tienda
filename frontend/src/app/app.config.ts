import { provideHttpClient, withInterceptors, withXsrfConfiguration } from '@angular/common/http';
import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZoneChangeDetection } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter } from '@angular/router';
import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';

import { authInterceptor } from './features/auth/auth.interceptor';
import { routes } from './app.routes';

// Aura con la escala primaria en índigo: tablas, paginador, selects y foco coinciden con styles.css.
const TemaTienda = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#f3f3f9',
      100: '#ececf4',
      200: '#c9cae0',
      300: '#a8aad0',
      400: '#7e80b5',
      500: '#5a5d9c',
      600: '#414486',
      700: '#2c2e6a',
      800: '#20224e',
      900: '#181a3c',
      950: '#0f1027',
    },
    colorScheme: {
      light: {
        primary: {
          color: '{primary.700}',
          contrastColor: '#ffffff',
          hoverColor: '{primary.800}',
          activeColor: '{primary.900}',
        },
        highlight: {
          background: '{primary.700}',
          focusBackground: '{primary.800}',
          color: '#ffffff',
          focusColor: '#ffffff',
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
