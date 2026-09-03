import { Component, inject } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-section-placeholder',
  templateUrl: './section-placeholder.component.html',
  styleUrl: './section-placeholder.component.css',
})
export class SectionPlaceholderComponent {
  private readonly route = inject(ActivatedRoute);

  readonly title = this.route.snapshot.data['title'] as string;
  readonly description = this.route.snapshot.data['description'] as string;
  readonly icon = this.route.snapshot.data['icon'] as string;
}
