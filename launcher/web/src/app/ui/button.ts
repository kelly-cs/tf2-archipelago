import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

export type ButtonTone = 'primary' | 'secondary' | 'ghost' | 'danger';
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
        background var(--quick) var(--ease),
        border-color var(--quick) var(--ease),
        color var(--quick) var(--ease);
    }

    /* The press itself, not a hover: the only thing that moves is the thing
       under the finger, and only while it is under it. */
    button:active:not(:disabled) {
      transform: translateY(1px);
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

    .primary {
      background: var(--accent);
      color: var(--surface-page);
      font-weight: 700;
    }

    .primary:hover:not(:disabled) {
      background: var(--accent-bright);
    }

    .secondary {
      background: var(--surface-raised);
      border-color: var(--line-strong);
      color: var(--text);
    }

    .secondary:hover:not(:disabled) {
      background: var(--surface-hover);
    }

    .ghost {
      background: transparent;
      border-color: var(--line-strong);
      color: var(--text-dim);
    }

    .ghost:hover:not(:disabled) {
      color: var(--text);
      background: var(--surface-raised);
    }

    .danger {
      background: transparent;
      border-color: var(--state-bad);
      color: var(--state-warn);
    }

    .danger:hover:not(:disabled) {
      background: var(--state-bad);
      color: var(--text);
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
