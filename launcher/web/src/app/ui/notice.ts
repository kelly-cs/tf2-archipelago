import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type NoticeTone = 'info' | 'warn' | 'bad';

/** The launcher saying something: a save refused, a restart needed, a notice. */
@Component({
  selector: 'app-notice',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div [class]="tone()" role="status">
      <ng-content />
    </div>
  `,
  styles: `
    :host {
      display: block;
    }

    div {
      display: flex;
      align-items: center;
      justify-content: space-between;
      flex-wrap: wrap;
      gap: var(--gap);
      border: 1px solid currentcolor;
      border-radius: var(--radius);
      padding: 10px 14px;
      font-size: var(--text-md);
      opacity: 1;
      transform: none;
      transition:
        opacity var(--settle) var(--ease-out),
        transform var(--settle) var(--ease-out);

      /* A transition rather than a keyframe: the launcher replaces this as fast
         as it has something to say, and a keyframe restarts from zero when it
         is interrupted. */
      @starting-style {
        opacity: 0;
        transform: translateY(4px);
      }
    }

    .info {
      color: var(--state-info);
      background: color-mix(in srgb, var(--state-info) 8%, transparent);
    }

    .warn {
      color: var(--accent);
      background: color-mix(in srgb, var(--accent) 8%, transparent);
    }

    .bad {
      color: var(--state-warn);
      background: color-mix(in srgb, var(--state-bad) 12%, transparent);
    }
  `,
})
export class Notice {
  readonly tone = input<NoticeTone>('info');
}
