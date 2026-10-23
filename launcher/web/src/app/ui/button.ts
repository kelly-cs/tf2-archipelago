import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

export type ButtonTone = 'primary' | 'go' | 'halt' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'small' | 'medium' | 'large';

/**
 * Every press in the app. One component so a button cannot be half a button:
 * the disabled state, the focus ring and the press feedback are decided here
 * and nowhere else.
 */
@Component({
  selector: 'app-button',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <button
      [type]="type()"
      [class]="tone() + ' ' + size()"
      [disabled]="disabled()"
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
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: var(--gap-sm);
      width: 100%;
      font-family: var(--font-body);
      font-weight: 600;
      border-radius: var(--radius);
      border: 1px solid transparent;
      cursor: pointer;
      transition:
        transform var(--press) var(--ease-out),
        background var(--quick) var(--ease-out),
        border-color var(--quick) var(--ease-out),
        color var(--quick) var(--ease-out);
    }

    /* The press. scale rather than a nudge, because scale takes the label and
       the icon with it: the whole control answers, which is what makes the
       launcher feel like it heard. */
    button:active:not(:disabled) {
      transform: scale(0.97);
    }

    button:disabled {
      cursor: not-allowed;
      opacity: 0.55;
    }

    .small {
      font-size: var(--text-sm);
      padding: 5px 10px;
    }

    .medium {
      font-size: var(--text-md);
      padding: 8px 14px;
    }

    .large {
      font-size: var(--text-lg);
      padding: 11px 16px;
    }

    /* Hover only where there is a pointer. A tap on a touch screen fires
       hover too, and leaves the control looking pressed after the finger
       has gone. */
    @media (hover: hover) and (pointer: fine) {
      .primary:hover:not(:disabled) {
        background: var(--accent-bright);
      }

      .go:hover:not(:disabled),
      .halt:hover:not(:disabled) {
        filter: brightness(1.12);
      }

      .secondary:hover:not(:disabled) {
        background: var(--surface-hover);
      }

      .ghost:hover:not(:disabled) {
        color: var(--text);
        background: var(--surface-raised);
      }

      .danger:hover:not(:disabled) {
        background: var(--state-bad);
        color: var(--text);
      }
    }

    .primary {
      background: var(--accent);
      color: var(--surface-page);
      font-weight: 700;
    }

    /* Starting and stopping the server are the two presses with consequences,
       so they are the two that are not gold: green to go, red to stop. */
    .go {
      background: var(--state-good);
      color: var(--surface-page);
      font-weight: 700;
    }

    .halt {
      background: var(--state-stop);
      color: var(--surface-page);
      font-weight: 700;
    }

    .secondary {
      background: var(--surface-raised);
      border-color: var(--line-strong);
      color: var(--text);
    }

    .ghost {
      background: transparent;
      border-color: var(--line-strong);
      color: var(--text-dim);
    }

    .danger {
      background: transparent;
      border-color: var(--state-bad);
      color: var(--state-warn);
    }
  `,
})
export class Button {
  readonly tone = input<ButtonTone>('secondary');
  readonly size = input<ButtonSize>('medium');
  readonly type = input<'button' | 'submit'>('button');
  readonly disabled = input(false);
  readonly hint = input('');

  /** pressed drives aria-pressed for a button that is a toggle. */
  readonly pressed = input<boolean | undefined>(undefined);

  readonly press = output<void>();
}
