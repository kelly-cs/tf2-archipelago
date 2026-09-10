import { ChangeDetectionStrategy, Component, input } from '@angular/core';

import { Badge } from '@app/ui/badge';

/**
 * The frame every settings row sits in: its label, its help, and the reason it
 * cannot be used where that is the case.
 *
 * A row the player cannot use is shown and disabled rather than hidden. Hiding
 * it leaves them looking for a setting the documentation says exists.
 */
@Component({
  selector: 'app-field-shell',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge],
  template: `
    <div class="row" [class.off]="disabled()">
      <div class="words">
        @if (label()) {
          <label [attr.for]="controlId()">{{ label() }}</label>
        }
        @if (help()) {
          <p>{{ help() }}</p>
        }
        @if (disabled() && reason()) {
          <app-badge tone="warn">{{ reason() }}</app-badge>
        }
      </div>
      <div class="control"><ng-content /></div>
    </div>
  `,
  styles: `
    :host {
      display: block;
    }

    .row {
      display: grid;
      grid-template-columns: minmax(0, 1fr) minmax(0, 320px);
      gap: var(--gap) var(--gap-lg);
      align-items: start;
      padding: var(--gap) 0;
      border-bottom: 1px solid var(--surface-hover);
    }

    @media (max-width: 700px) {
      .row {
        grid-template-columns: 1fr;
      }
    }

    .off {
      opacity: 0.6;
    }

    label {
      font-weight: 600;
    }

    p {
      margin: 2px 0 0;
      font-size: var(--text-sm);
      color: var(--text-dim);
    }

    .control {
      display: flex;
      justify-content: flex-end;
    }
  `,
})
export class FieldShell {
  readonly label = input('');
  readonly help = input('');
  readonly reason = input('');
  readonly disabled = input(false);
  readonly controlId = input('');
}
