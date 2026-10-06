import {
  AfterViewInit,
  Component,
  DestroyRef,
  ElementRef,
  Injector,
  QueryList,
  ViewChild,
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

  @ViewChild('brandLogo', { read: ElementRef })
  private readonly brandLogo!: ElementRef<HTMLElement>;

  @ViewChild('mosaic', { read: ElementRef })
  private readonly mosaic!: ElementRef<HTMLElement>;

  readonly mode = signal<AuthMode>(this.modeFromUrl(this.router.url));
  readonly introActive = signal(true);
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
    if (this.shouldReduceMotion() || this.isCompactViewport()) {
      this.introActive.set(false);
      return;
    }

    this.playInitialMosaicAssembly();
  }

  private switchMode(mode: AuthMode): void {
    if (mode === this.mode() || this.transitioning) return;

    const tiles = this.tileElements.map(({ nativeElement }) => nativeElement);
    const shouldReduceMotion = this.shouldReduceMotion();
    const shouldAnimateMosaic = this.ready && !shouldReduceMotion && !this.isCompactViewport();
    const state = shouldAnimateMosaic ? Flip.getState(tiles) : undefined;
    const content = gsap.utils.toArray<HTMLElement>('.brand, .auth-shell__copy, .auth-shell__form');

    if (state) {
      this.stopLogoMotion();
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
              this.revealContent(content, () => (this.transitioning = false));
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

  private shouldReduceMotion(): boolean {
    return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;
  }

  private isCompactViewport(): boolean {
    return window.matchMedia?.('(max-width: 820px)').matches ?? false;
  }

  private playInitialMosaicAssembly(): void {
    const content = gsap.utils.toArray<HTMLElement>('.brand, .auth-shell__copy, .auth-shell__form');
    const mosaic = this.mosaic.nativeElement;
    const tiles = this.tileElements.map(({ nativeElement }) => nativeElement);
    const offsets = [
      [-28, -20],
      [24, -18],
      [30, 16],
      [-24, 18],
      [18, 24],
      [-18, -12],
    ];

    this.transitioning = true;
    gsap.set(content, { autoAlpha: 0 });
    gsap.set(mosaic, { autoAlpha: 1 });
    tiles.forEach((tile, index) => {
      const [x, y] = offsets[index];
      gsap.set(tile, {
        autoAlpha: 0,
        scale: 0.72,
        x,
        y,
        transformOrigin: '50% 50%',
        willChange: 'transform,opacity',
      });
    });

    const timeline = gsap.timeline({
      onComplete: () => {
        tiles.forEach((tile) => tile.style.removeProperty('will-change'));
        this.startLogoFloat();
        this.transitioning = false;
      },
    });

    timeline.to(tiles, {
      autoAlpha: 1,
      scale: 1,
      x: 0,
      y: 0,
      duration: 0.48,
      ease: 'back.out(1.35)',
      stagger: 0.065,
    });

    timeline.call(() => this.introActive.set(false), [], 0.805);
    this.addContentReveal(timeline, content, 0.805);
  }

  private revealContent(
    content = gsap.utils.toArray<HTMLElement>('.brand, .auth-shell__copy, .auth-shell__form'),
    onComplete?: () => void,
  ): void {
    const logo = this.brandLogo.nativeElement;
    gsap.killTweensOf(content);
    gsap.killTweensOf(logo);

    const timeline = gsap.timeline({
      onComplete: () => {
        this.startLogoFloat();
        onComplete?.();
      },
    });

    this.addContentReveal(timeline, content, 0);
  }

  private addContentReveal(
    timeline: gsap.core.Timeline,
    content: HTMLElement[],
    position: number,
  ): void {
    const logo = this.brandLogo.nativeElement;
    const copy = content.find((element) => element.classList.contains('auth-shell__copy'));
    const form = content.find((element) => element.classList.contains('auth-shell__form'));
    const direction = this.mode() === 'login' ? -14 : 14;

    timeline
      .set(logo, { transformOrigin: '50% 50%' }, position)
      .fromTo(
        content,
        { autoAlpha: 0, x: direction },
        { autoAlpha: 1, x: 0, duration: 0.24, ease: 'power2.out' },
        position,
      )
      .fromTo(
        logo,
        { scale: 0.72, y: -12, rotation: -10 },
        { scale: 1, y: 0, rotation: 0, duration: 0.42, ease: 'back.out(1.7)' },
        position,
      );

    if (copy) {
      timeline.fromTo(
        copy,
        { y: 8 },
        { y: 0, duration: 0.25, ease: 'power2.out' },
        position + 0.01,
      );
    }

    if (form) {
      timeline.fromTo(
        form,
        { y: 10 },
        { y: 0, duration: 0.25, ease: 'power2.out' },
        position + 0.02,
      );
    }
  }

  private startLogoFloat(): void {
    gsap.to(this.brandLogo.nativeElement, {
      y: -2,
      rotation: 1.25,
      duration: 1.7,
      delay: 0.35,
      ease: 'sine.inOut',
      repeat: -1,
      yoyo: true,
    });
  }

  private stopLogoMotion(): void {
    const logo = this.brandLogo.nativeElement;
    gsap.killTweensOf(logo);
    gsap.set(logo, { clearProps: 'transform' });
  }
}
