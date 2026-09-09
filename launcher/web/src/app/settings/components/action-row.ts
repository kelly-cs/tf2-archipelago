import { ChangeDetectionStrategy, Component, computed, input, output, signal } from '@angular/core';

import { Button } from '@app/ui/button';
import { Field, Kind } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * A button. A Confirm is the same button that cannot be taken back, so it asks
 * first, in place: a dialog for "reset every setting" is a dialog the player
 * dismisses without reading.
 */
@Component({
  selector: 'app-action-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button],
  template: `
    @if (asking()) {
      <div class="asking">
        <span class="warning">{{ field().warning || 'This cannot be taken back.' }}</span>
        <app-button tone="danger" size="small" (press)="confirm()">Yes, do it</app-button>
        <app-button tone="ghost" size="small" (press)="asking.set(false)">Cancel</app-button>
      </div>
    } @else {
      <app-button
        [tone]="dangerous() ? 'danger' : 'secondary'"
        [disabled]="field().disabled"
        (press)="press()"
      >
        {{ field().label }}
      </app-button>
    }
  `,
  styles: `
    :host {
      display: block;
    }

    .asking {
      display: flex;
      align-items: center;
      gap: var(--gap-sm);
      flex-wrap: wrap;
      justify-content: flex-end;
      animation: rise-in var(--quick) var(--ease);
    }

    .warning {
      font-size: var(--text-sm);
      color: var(--state-warn);
    }
  `,
})
export class ActionRow {
  readonly field = input.required<Field>();
  readonly fired = output<string>();

  readonly asking = signal(false);
  readonly dangerous = computed(() => this.field().kind === Kind.CONFIRM);

  press(): void {
    if (this.dangerous()) {
      this.asking.set(true);
      return;
    }
    this.fired.emit(this.field().id);
  }

  confirm(): void {
    this.asking.set(false);
    this.fired.emit(this.field().id);
  }
}
