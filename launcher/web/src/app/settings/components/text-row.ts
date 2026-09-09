import { ChangeDetectionStrategy, Component, computed, input, output } from '@angular/core';
import { FormField } from '@angular/forms/signals';
import type { FieldTree } from '@angular/forms/signals';

import { Button } from '@app/ui/button';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/**
 * A line: a folder, a server name, a room address.
 *
 * A Browse row gets a button beside it. The typed path stays editable: the
 * picker is a way to fill it in, not the only way.
 */
@Component({
  selector: 'app-text-row',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, FormField],
  template: `
    <div class="line">
      <input
        [id]="field().id"
        type="text"
        [formField]="control()"
        [attr.placeholder]="field().placeholder || null"
        (input)="typed.emit($any($event.target).value)"
      />
      @if (field().browse) {
        <app-button size="small" [disabled]="field().disabled" (press)="browse.emit()">
          Browse
        </app-button>
      }
    </div>
  `,
  styleUrl: './control.scss',
})
export class TextRow {
  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly typed = output<string>();
  readonly browse = output<void>();

  readonly deferred = computed(() => this.field().deferred);
}
