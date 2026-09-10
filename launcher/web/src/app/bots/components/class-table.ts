import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';

import { SettingsStore } from '@app/settings/settings-store';
import { Chip } from '@app/ui/chip';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/** One class: whether the mod may draw it, and what it carries when drawn. */
interface Mercenary {
  readonly key: string;
  readonly name: string;
  readonly allowed: Field;
  readonly carries: Field | undefined;
}

/**
 * The nine classes as a table: a chip that allows or forbids the draw, and
 * beside it the loadout the class carries when a seat did not say.
 *
 * Both columns are rows form declares, read off the model by id, so a class
 * the game adds later appears here without this file knowing about it, and a
 * class allowed here is the same answer the Team tab's seats draw from.
 */
@Component({
  selector: 'app-class-table',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Chip],
  templateUrl: './class-table.html',
  styleUrl: './class-table.scss',
})
export class ClassTable {
  private readonly settings = inject(SettingsStore);

  readonly open = computed(() => this.settings.open());

  readonly mercenaries = computed<Mercenary[]>(() => {
    const fields = this.settings.fieldsMatching('bots.class.');
    return fields
      .filter((field) => field.id.endsWith('.allowed'))
      .map((allowed) => ({
        key: allowed.id.slice('bots.class.'.length, -'.allowed'.length),
        name: allowed.label,
        allowed,
        carries: fields.find((one) => one.id === allowed.id.replace('.allowed', '.loadout')),
      }));
  });

  readonly allowedCount = computed(
    () => this.mercenaries().filter((one) => this.on(one.allowed)).length,
  );

  flip(mercenary: Mercenary): void {
    const allowed = this.on(mercenary.allowed);
    // Never all nine forbidden: a lineup the mod cannot draw from leaves the
    // seats empty, and an empty seat is a wave short of six defenders.
    if (allowed && this.allowedCount() === 1) {
      return;
    }
    this.settings.change(mercenary.allowed.id, allowed ? 'false' : 'true');
  }

  carry(mercenary: Mercenary, value: string): void {
    if (mercenary.carries !== undefined) {
      this.settings.change(mercenary.carries.id, value);
    }
  }

  value(field: Field | undefined): string {
    return field === undefined ? '' : this.settings.value(field.id);
  }

  on(field: Field): boolean {
    return this.value(field) === 'true';
  }
}
