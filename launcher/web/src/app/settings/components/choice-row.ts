import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { FormField } from '@angular/forms/signals';
import type { FieldTree } from '@angular/forms/signals';

import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * One of a fixed set. The value saved and the line the player reads are
 * separate: the settings file holds "progression" and the player is asked
 * "Required for progression".
 */
@Component({
  selector: 'app-choice-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField],
  template: `
    <div class="line">
      <select
        [id]="field().id"
        [formField]="control()"
        (change)="typed.emit($any($event.target).value)"
      >
        @for (option of field().options; track option.value) {
          <option [value]="option.value">{{ option.label }}</option>
        }
      </select>
    </div>
  `,
  styleUrl: './control.scss',
})
export class ChoiceRow {
  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly typed = output<string>();
}
