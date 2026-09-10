import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';
import { FormField } from '@angular/forms/signals';
import type { FieldTree } from '@angular/forms/signals';

import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * A line that is never shown back, only replaced. The launcher blanks these
 * before the model leaves it, so an empty box means "unchanged" rather than
 * "empty", which is what the placeholder says.
 */
@Component({
  selector: 'app-password-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [FormField],
  template: `
    <div class="line">
      <input
        [id]="field().id"
        type="password"
        autocomplete="off"
        [formField]="control()"
        [attr.placeholder]="field().placeholder || 'unchanged'"
        (input)="typed.emit($any($event.target).value)"
      />
    </div>
  `,
  styleUrl: './control.scss',
})
export class PasswordRow {
  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly typed = output<string>();
}
