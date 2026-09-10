import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

export type ChipTone = 'accent' | 'allow';

/**
 * A pill that is on or off: a filter, a class the mod may draw.
 *
 * Not a Button. A button does something; a chip says what is being shown or
 * allowed, and the pressed state is the whole point of it. Struck through when
 * off so the state survives a screenshot with the colour stripped.
 */
@Component({
  selector: 'app-chip',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <button
      type="button"
      [class]="tone()"
      [class.on]="pressed()"
      [attr.aria-pressed]="pressed()"
      [attr.title]="hint() || null"
      (click)="press.emit()"
    >
      <ng-content />
    </button>
  `,
  styles: `
    :host {
      display: inline-flex;
    }

    button {
      font-family: var(--font-body);
      font-size: var(--text-sm);
      font-weight: 600;
      padding: 5px 12px;
      border-radius: var(--radius-pill);
      border: 1px solid var(--line-soft);
      background: transparent;
      color: var(--text-mute);
      cursor: pointer;
      transition:
        background var(--quick) var(--ease-out),
        border-color var(--quick) var(--ease-out),
        color var(--quick) var(--ease-out);
    }

    .allow:not(.on) {
      text-decoration: line-through;
    }

    .accent.on {
      background: var(--accent-tint-strong);
      border-color: var(--accent);
      color: var(--accent);
    }

    .allow.on {
      background: var(--good-tint);
      border-color: var(--state-good);
      color: var(--text);
    }

    @media (hover: hover) and (pointer: fine) {
      button:hover {
        color: var(--text);
      }
    }
  `,
})
export class Chip {
  readonly tone = input<ChipTone>('accent');
  readonly pressed = input(false);
  readonly hint = input('');
  readonly press = output<void>();
}
