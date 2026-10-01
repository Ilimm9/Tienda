import {
  AfterViewInit,
  Component,
  DestroyRef,
  ElementRef,
  Injector,
  QueryList,
  ViewChildren,
  afterNextRender,
  inject,
  signal,
} from '@angular/core';
import { NavigationEnd, Router, RouterLink, RouterOutlet } from '@angular/router';
import { filter } from 'rxjs';
import { gsap } from 'gsap';
import { Flip } from 'gsap/Flip';

import { AuthMode, AuthTransitionService } from '../auth-transition.service';

gsap.registerPlugin(Flip);

@Component({
  selector: 'app-auth-shell',
  imports: [RouterLink, RouterOutlet],
  templateUrl: './auth-shell.component.html',
  styleUrl: './auth-shell.component.css',
})
export class AuthShellComponent implements AfterViewInit {
  private readonly router = inject(Router);
  private readonly transition = inject(AuthTransitionService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly injector = inject(Injector);

  @ViewChildren('tile', { read: ElementRef }) private readonly tileElements!: QueryList<
    ElementRef<HTMLElement>
  >;

  readonly mode = signal<AuthMode>(this.modeFromUrl(this.router.url));
  private ready = false;
  private transitioning = false;

  constructor() {
    const requestSubscription = this.transition.requests$.subscribe((mode) =>
      this.switchMode(mode),
    );
    const navigationSubscription = this.router.events
      .pipe(filter((event): event is NavigationEnd => event instanceof NavigationEnd))
      .subscribe((event) => {
        const nextMode = this.modeFromUrl(event.urlAfterRedirects);
        if (!this.transitioning && nextMode !== this.mode()) this.mode.set(nextMode);
      });

    this.destroyRef.onDestroy(() => {
      requestSubscription.unsubscribe();
      navigationSubscription.unsubscribe();
    });
  }

  ngAfterViewInit(): void {
    this.ready = true;
  }

  private switchMode(mode: AuthMode): void {
    if (mode === this.mode() || this.transitioning) return;

    const tiles = this.tileElements.map(({ nativeElement }) => nativeElement);
    const shouldReduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const state = this.ready && !shouldReduceMotion ? Flip.getState(tiles) : undefined;
    const content = gsap.utils.toArray<HTMLElement>('.brand, .auth-shell__copy, .auth-shell__form');

    if (state) {
      gsap.set(content, { autoAlpha: 0, x: mode === 'login' ? -14 : 14 });
    }

    this.transitioning = true;
    void this.router.navigate([mode === 'login' ? '/login' : '/registro']).then((navigated) => {
      if (!navigated) {
        gsap.set(content, { clearProps: 'transform,opacity,visibility' });
        this.transitioning = false;
        return;
      }

      this.mode.set(mode);
      if (!state) {
        this.transitioning = false;
        return;
      }

      afterNextRender(
        () => {
          tiles.forEach((tile) => tile.style.setProperty('will-change', 'transform'));
          Flip.from(state, {
            absolute: true,
            scale: true,
            duration: 0.78,
            ease: 'power3.inOut',
            stagger: 0.055,
            onComplete: () => {
              tiles.forEach((tile) => tile.style.removeProperty('will-change'));
              gsap.to(content, {
                autoAlpha: 1,
                x: 0,
                duration: 0.32,
                ease: 'power3.out',
                clearProps: 'transform,opacity,visibility',
                onComplete: () => (this.transitioning = false),
              });
            },
          });
        },
        { injector: this.injector },
      );
    });
  }

  private modeFromUrl(url: string): AuthMode {
    return url.startsWith('/registro') ? 'registro' : 'login';
  }
}
