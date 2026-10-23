import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type DotTone = 'good' | 'warn' | 'bad' | 'idle';

/**
 * The dot beside a state. It pulses only while something is still happening,
 * because a dot that always pulses says nothing.
 */
@Component({
  selector: 'app-status-dot',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<span [class]="tone()" [class.pulsing]="pulsing()" aria-hidden="true"></span>`,
  styles: `
    :host {
      display: inline-flex;
    }

    span {
      width: 9px;
      height: 9px;
      border-radius: 50%;
      background: var(--text-disabled);
    }

    .pulsing {
      animation: pulse-dot 1s ease-in-out infinite;
    }

    .good {
      background: var(--state-good);
    }

    .warn {
      background: var(--accent);
    }

    .bad {
      background: var(--state-bad);
    }
  `,
})
export class StatusDot {
  readonly tone = input<DotTone>('idle');
  readonly pulsing = input(false);
}
