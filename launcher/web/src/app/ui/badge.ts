import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type BadgeTone = 'neutral' | 'good' | 'warn' | 'bad' | 'info' | 'accent';

/** One word about one thing: a state, a source, a tier. */
@Component({
  selector: 'app-badge',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<span [class]="tone()"><ng-content /></span>`,
  styles: `
    :host {
      display: inline-flex;
    }

    span {
      display: inline-flex;
      align-items: center;
      gap: var(--gap-xs);
      font-size: var(--text-xs);
      font-weight: 600;
      padding: 2px 9px;
      border-radius: var(--radius-pill);
      border: 1px solid currentcolor;
      white-space: nowrap;
    }

    .neutral {
      color: var(--text-faint);
    }

    .good {
      color: var(--state-good);
    }

    .warn {
      color: var(--state-warn);
    }

    .bad {
      color: var(--state-bad);
    }

    .info {
      color: var(--state-info);
    }

    .accent {
      color: var(--accent);
    }
  `,
})
export class Badge {
  readonly tone = input<BadgeTone>('neutral');
}
