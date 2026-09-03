import { TestBed } from '@angular/core/testing';

import { HomeComponent } from './home.component';

describe('HomeComponent', () => {
  it('renders the dashboard summary inside the administrative shell', async () => {
    await TestBed.configureTestingModule({
      imports: [HomeComponent],
    }).compileComponents();
    const fixture = TestBed.createComponent(HomeComponent);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Bienvenido');
    expect(fixture.nativeElement.querySelectorAll('article')).toHaveLength(3);
  });
});
