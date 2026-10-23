import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, concatMap, tap } from 'rxjs';

import { SettingsStore } from '@app/settings/settings-store';
import { Button } from '@app/ui/button';
import { Chip } from '@app/ui/chip';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/** One name in the pool: the row that holds it, and whether it is drawn. */
interface BotName {
  readonly id: string;
  readonly name: string;
  readonly field: Field;
  readonly mine: boolean;
}

/**
 * The names the bots draw from: the ones this launcher ships, the ones you
 * added, and the box that adds another.
 *
 * Every name is a row form declares, so the shipped list can grow in a later
 * release and appear here without this file knowing about it. A name added is
 * a press rather than a change, because the refusals live beside the settings:
 * too long for the game to keep, a comma the Compose list would split on, or a
 * name the pool already holds.
 *
 * The mod reads the pool when a map starts, so a change here reaches the bots
 * on the next mission rather than the next wave.
 */
@Component({
  selector: 'app-name-pool',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, Chip],
  templateUrl: './name-pool.html',
  styleUrl: './name-pool.scss',
})
export class NamePool {
  private readonly settings = inject(SettingsStore);

  readonly open = computed(() => this.settings.open());
  readonly feedback = signal('');

  readonly typed = computed(() => this.settings.field('bots.name_new'));

  readonly mine = computed(() => this.namesUnder('bots.name.added.', true));
  readonly shipped = computed(() => this.namesUnder('bots.name.shipped.', false));

  /** How many names the bots can actually draw, which is the whole point of
      the page: a pool of one is six bots with the same name. */
  readonly drawn = computed(
    () =>
      this.mine().filter((one) => this.on(one.field)).length +
      this.shipped().filter((one) => this.on(one.field)).length,
  );

  readonly added = new Subject<void>();

  constructor() {
    this.added
      .pipe(
        concatMap(() => this.settings.dispatch('bots.name_add')),
        tap((said) => this.feedback.set(said)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  flip(name: BotName): void {
    this.settings.change(name.field.id, this.on(name.field) ? 'false' : 'true');
    this.feedback.set(
      this.on(name.field)
        ? `${name.name} is out of the pool.`
        : `${name.name} is back in the pool.`,
    );
  }

  type(value: string): void {
    this.settings.change('bots.name_new', value);
  }

  value(field: Field | undefined): string {
    return field === undefined ? '' : this.settings.value(field.id);
  }

  on(field: Field): boolean {
    return this.value(field) === 'true';
  }

  private namesUnder(prefix: string, mine: boolean): BotName[] {
    return this.settings
      .fieldsMatching(prefix)
      .map((field) => ({ id: field.id, name: field.label, field, mine }));
  }
}
