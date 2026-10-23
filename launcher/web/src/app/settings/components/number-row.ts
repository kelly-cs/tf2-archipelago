import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { FormField } from '@angular/forms/signals';
import type { FieldTree } from '@angular/forms/signals';

import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * A whole number with a floor and a ceiling. Both are resolved by the launcher:
 * the mission count's ceiling is however many missions the chosen tier leaves,
 * so they are read off the row rather than declared here.
 *
 * They are shown beside the box rather than set as min and max on it. Signal
 * forms owns those attributes and its min/max rules are for a number field,
 * while every value here crosses the wire as text. The refusal a player reads
 * for an out-of-range answer is form.Apply's, on the Go side, which is the one
 * place that decides: a second check here could only ever disagree with it.
 */
@Component({
  selector: 'app-number-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField],
  template: `
    <div class="line">
      <span class="bounds">{{ field().low }} to {{ field().high }}</span>
      <input
        [id]="field().id"
        type="number"
        [formField]="control()"
        (input)="typed.emit($any($event.target).value)"
      />
    </div>
  `,
  styleUrl: './control.scss',
})
export class NumberRow {
  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly typed = output<string>();
}
