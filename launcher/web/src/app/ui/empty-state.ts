import { ChangeDetectionStrategy, Component } from '@angular/core';

/**
 * What a list says when it has nothing in it. A blank area reads as a screen
 * that failed to load; a sentence reads as a screen with nothing to show.
 */
@Component({
  selector: 'app-empty-state',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<p><ng-content /></p>`,
  styles: `
    p {
      margin: 0;
      padding: var(--gap-xl);
      text-align: center;
      color: var(--text-faint);
      font-size: var(--text-md);
    }
  `,
})
export class EmptyState {}
