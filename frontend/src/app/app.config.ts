import { provideHttpClient, withInterceptors, withXsrfConfiguration } from '@angular/common/http';
import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZoneChangeDetection } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter } from '@angular/router';
import Aura from '@primeuix/themes/aura';
import { providePrimeNG } from 'primeng/config';

import { authInterceptor } from './features/auth/auth.interceptor';
import { routes } from './app.routes';

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
        preset: Aura,
        options: {
          darkModeSelector: '[data-theme="dark"]',
        },
      },
    }),
  ],
};
