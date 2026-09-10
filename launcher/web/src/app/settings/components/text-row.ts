import { ChangeDetectionStrategy, Component, computed, input, output, signal } from '@angular/core';
import { FormField } from '@angular/forms/signals';
import type { FieldTree } from '@angular/forms/signals';

import { FolderPicker } from '@app/settings/components/folder-picker';
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
  imports: [Button, FolderPicker, FormField],
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
        <app-button size="small" [disabled]="field().disabled" (press)="picking.set(!picking())">
          Browse
        </app-button>
      }
    </div>
    @if (picking()) {
      <app-folder-picker
        [start]="value()"
        (chosen)="pick($event)"
        (dismissed)="picking.set(false)"
      />
    }
  `,
  styleUrl: './control.scss',
})
export class TextRow {
  readonly field = input.required<Field>();
  readonly control = input.required<FieldTree<string>>();
  readonly value = input('');
  readonly typed = output<string>();

  readonly picking = signal(false);
  readonly deferred = computed(() => this.field().deferred);

  pick(path: string): void {
    this.picking.set(false);
    this.typed.emit(path);
  }
}
