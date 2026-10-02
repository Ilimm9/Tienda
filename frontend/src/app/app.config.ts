import { provideHttpClient, withInterceptors, withXsrfConfiguration } from '@angular/common/http';
import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZoneChangeDetection } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter } from '@angular/router';
import { definePreset } from '@primeuix/themes';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';

import { authInterceptor } from './features/auth/auth.interceptor';
import { routes } from './app.routes';

// Aura con la escala primaria en esmeralda (variante B): tablas, paginador, selects y foco coinciden con styles.css.
const TemaTienda = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#ecf8f4',
      100: '#ddf5ec',
      200: '#b5e8d6',
      300: '#7fd6ba',
      400: '#3fc29e',
      500: '#12b88a',
      600: '#0a8f72',
      700: '#046c5e',
      800: '#03574c',
      900: '#02463d',
      950: '#012b26',
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
