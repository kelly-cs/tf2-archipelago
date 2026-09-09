import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import type { FieldTree } from '@angular/forms/signals';

import { ActionRow } from '@app/settings/components/action-row';
import { ChoiceRow } from '@app/settings/components/choice-row';
import { FieldShell } from '@app/settings/components/field-shell';
import { NumberRow } from '@app/settings/components/number-row';
import { PasswordRow } from '@app/settings/components/password-row';
import { TextRow } from '@app/settings/components/text-row';
import { ToggleRow } from '@app/settings/components/toggle-row';
import { Field, Kind } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * One row, drawn by its kind.
 *
 * The switch is exhaustive on purpose and there is no default: a Kind added to
 * form and to the contract but not here has to fail loudly rather than draw as
 * nothing. TestProtoKindsMatchFormKinds guards the other half of that.
 */
@Component({
  selector: 'app-settings-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [ActionRow, ChoiceRow, FieldShell, NumberRow, PasswordRow, TextRow, ToggleRow],
  template: `
    @if (isButton()) {
      <app-field-shell
        [label]="field().label"
        [help]="field().help"
        [reason]="field().reason"
        [disabled]="field().disabled"
      >
        <app-action-row [field]="field()" (fired)="fired.emit($event)" />
      </app-field-shell>
    } @else {
      <app-field-shell
        [label]="field().label"
        [help]="field().help"
        [reason]="field().reason"
        [disabled]="field().disabled"
        [controlId]="field().id"
      >
        @switch (field().kind) {
          @case (Kind.TEXT) {
            <app-text-row
              [field]="field()"
              [control]="control()"
              [value]="value()"
              (typed)="typed.emit($event)"
            />
          }
          @case (Kind.PASSWORD) {
            <app-password-row
              [field]="field()"
              [control]="control()"
              (typed)="typed.emit($event)"
            />
          }
          @case (Kind.NUMBER) {
            <app-number-row [field]="field()" [control]="control()" (typed)="typed.emit($event)" />
          }
          @case (Kind.CHOICE) {
            <app-choice-row [field]="field()" [control]="control()" (typed)="typed.emit($event)" />
          }
          @case (Kind.TOGGLE) {
            <app-toggle-row [field]="field()" [value]="value()" (typed)="typed.emit($event)" />
          }
        }
      </app-field-shell>
    }
  `,
})
export class SettingsRow {
  readonly Kind = Kind;

  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly value = input.required<string>();

  readonly typed = output<string>();
  readonly fired = output<string>();

  readonly isButton = computed(
    () => this.field().kind === Kind.ACTION || this.field().kind === Kind.CONFIRM,
  );
}
