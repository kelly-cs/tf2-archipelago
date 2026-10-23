import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';

import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * Yes or no. Some rows read better with a different word for each answer: "left
 * out" beside an unticked mission, not an unticked "in the pool". hintOff is
 * that word, and empty means the same word either way.
 */
@Component({
  selector: 'app-toggle-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <label class="tick">
      <input
        [id]="field().id"
        type="checkbox"
        [checked]="on()"
        [disabled]="field().disabled"
        (change)="typed.emit($any($event.target).checked ? 'true' : 'false')"
      />
      <span class="hint">{{ word() }}</span>
    </label>
  `,
  styleUrl: './control.scss',
})
export class ToggleRow {
  readonly field = input.required<Field>();
  readonly value = input.required<string>();
  readonly typed = output<string>();

  readonly on = computed(() => this.value() === 'true');
  readonly word = computed(() => {
    const field = this.field();
    return this.on() ? field.hint : field.hintOff || field.hint;
  });
}
