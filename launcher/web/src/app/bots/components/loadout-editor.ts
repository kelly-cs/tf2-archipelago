import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, concatMap, tap } from 'rxjs';

import { SettingsStore } from '@app/settings/settings-store';
import { Button } from '@app/ui/button';
import { Chip } from '@app/ui/chip';
import { Field, Option } from '@gen/tf2ap/launcher/v1/form_pb';

/** One slot of the loadout being built, with the weapon it holds. */
interface Slot {
  readonly field: Field;
  readonly name: string;
  readonly held: string;
}

/**
 * The loadout builder: a class, a weapon per slot, a name, and the loadouts
 * already kept.
 *
 * The rows are form's: the class menu, one choice per slot, the name, and the
 * two menus that load and remove. This draws them as a builder rather than as
 * rows, because a loadout is looked at whole. Picking a class is a chip, a slot
 * is a card, and a saved loadout is a line with Load and Remove on it rather
 * than an entry in two dropdowns.
 */
@Component({
  selector: 'app-loadout-editor',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, Chip],
  templateUrl: './loadout-editor.html',
  styleUrl: './loadout-editor.scss',
})
export class LoadoutEditor {
  private readonly settings = inject(SettingsStore);

  readonly open = computed(() => this.settings.open());
  readonly feedback = signal('');

  /** The saved loadout Remove was pressed on once. A second press removes it. */
  readonly removing = signal('');

  readonly classField = computed(() => this.settings.field('loadout.class'));
  readonly nameField = computed(() => this.settings.field('loadout.name'));
  readonly loadField = computed(() => this.settings.field('loadout.load'));
  readonly removeField = computed(() => this.settings.field('loadout.remove'));

  readonly classes = computed(() => this.classField()?.options ?? []);
  readonly chosenClass = computed(() => this.value(this.classField()));

  readonly slots = computed<Slot[]>(() =>
    this.settings.fieldsMatching('loadout.slot.').map((field) => ({
      field,
      name: field.label.trim(),
      held: labelOf(field.options, this.settings.value(field.id)),
    })),
  );

  /** The loadouts kept so far: every option of Load but the one that does nothing. */
  readonly saved = computed<Option[]>(() => (this.loadField()?.options ?? []).slice(1));

  readonly fired = new Subject<string>();

  constructor() {
    this.fired
      .pipe(
        concatMap((id) => this.settings.dispatch(id)),
        tap((refusal) => this.feedback.set(refusal === '' ? 'Saved the loadout.' : refusal)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  pickClass(value: string): void {
    const field = this.classField();
    if (field !== undefined && value !== this.chosenClass()) {
      this.settings.change(field.id, value);
      this.feedback.set('');
    }
  }

  hold(slot: Slot, value: string): void {
    this.settings.change(slot.field.id, value);
  }

  name(value: string): void {
    const field = this.nameField();
    if (field !== undefined) {
      this.settings.change(field.id, value);
    }
  }

  save(): void {
    this.fired.next('loadout.save');
  }

  load(option: Option): void {
    const field = this.loadField();
    if (field !== undefined) {
      this.settings.change(field.id, option.value);
      this.feedback.set(`Loaded ${option.label} into the builder.`);
    }
  }

  remove(option: Option): void {
    const field = this.removeField();
    if (field === undefined) {
      return;
    }
    if (this.removing() !== option.value) {
      this.removing.set(option.value);
      return;
    }
    this.settings.change(field.id, option.value);
    this.removing.set('');
    this.feedback.set(`Removed ${option.label}. A seat still naming it plays stock.`);
  }

  value(field: Field | undefined): string {
    return field === undefined ? '' : this.settings.value(field.id);
  }
}

function labelOf(options: Option[], value: string): string {
  return options.find((option) => option.value === value)?.label ?? 'stock';
}
