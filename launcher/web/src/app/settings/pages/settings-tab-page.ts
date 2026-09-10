import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  input,
  linkedSignal,
  signal,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { applyEach, disabled, form } from '@angular/forms/signals';
import { Subject, concatMap, tap } from 'rxjs';

import { MissionTable } from '@app/settings/components/mission-table';
import { SettingsRow } from '@app/settings/components/settings-row';
import { SettingsActions } from '@app/settings/settings-actions';
import { SettingsStore } from '@app/settings/settings-store';
import { slugOf } from '@app/settings/slug';
import { EmptyState } from '@app/ui/empty-state';
import { Notice } from '@app/ui/notice';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

/** Placed is a row and the section it was found in. */
interface Placed {
  readonly section: string;
  readonly field: Field;
}

/**
 * One row of the form the page binds to. locked travels in the model rather
 * than being looked up by index: the schema reads the row it is applied to, and
 * a row that moved would otherwise take another row's answer.
 */
interface Row {
  id: string;
  value: string;
  locked: boolean;
}

/**
 * One settings page: its own rows, then the rows of every section filed under
 * it, then the mission table if this is the page that owns it.
 *
 * The form is signal forms over the rows, so the bounds and the disabled state
 * a row carries are the form's rather than a check written twice. What a row
 * means still belongs to the launcher: a change goes there, and the rebuilt
 * model comes back on the stream.
 */
@Component({
  selector: 'app-settings-tab-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [EmptyState, MissionTable, Notice, SettingsRow],
  templateUrl: './settings-tab-page.html',
  styleUrl: './settings-tab-page.scss',
})
export class SettingsTabPage {
  readonly store = inject(SettingsStore);
  private readonly actions = inject(SettingsActions);

  readonly settingsTab = input('');
  readonly filter = linkedSignal<string, string>({
    source: this.settingsTab,
    computation: () => '',
  });

  readonly sections = computed(() => this.store.sections(this.settingsTab()));

  /** The rows on this page, in order, with the sections they belong to. */
  readonly rows = computed(() =>
    this.sections().flatMap((section) =>
      section.fields.map((field) => ({ section: section.title, field })),
    ),
  );

  readonly shown = computed(() => this.rows().filter(({ field }) => this.store.matches(field)));

  /** Which page of how many, so the player knows how much is left. */
  readonly stepLabel = computed(() => {
    const pages = this.store.topTabs();
    const here = pages.findIndex((tab) => slugOf(tab.title) === this.settingsTab());
    return here < 0 ? '' : `Section ${here + 1} of ${pages.length}`;
  });

  readonly title = computed(() => this.sections()[0]?.title ?? '');

  readonly missionTabs = computed(() => this.settingsTab() === 'missions');
  readonly intro = computed(() => this.sections()[0]?.intro ?? '');

  private readonly model = linkedSignal<Placed[], { rows: Row[] }>({
    source: () => this.rows(),
    computation: (placed) => ({
      rows: placed.map(({ field }) => ({
        id: field.id,
        value: this.store.value(field.id),
        locked: field.disabled,
      })),
    }),
  });

  /**
   * The form. Only `disabled` is a rule here: a Number's floor and ceiling are
   * on the input as min and max, and what a value means is form.Apply's on the
   * Go side, with the words the player reads. A second copy of that check here
   * would be a second place for it to be wrong.
   */
  readonly page = form(this.model, (path) => {
    applyEach(path.rows, (row) => {
      disabled(row.value, (context) => context.valueOf(row.locked));
    });
  });

  /** said is the last thing a button answered with: a path opened, a page to
      visit, a file saved. The launcher says the rest on the stream. */
  readonly said = signal('');

  readonly fired = new Subject<string>();

  constructor() {
    this.fired
      .pipe(
        concatMap((id) => this.actions.press(id)),
        tap((said) => this.said.set(said)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  answer(id: string, value: string): void {
    this.store.change(id, value);
  }

  value(id: string): string {
    return this.store.value(id);
  }

  indexOf(id: string): number {
    return this.rows().findIndex(({ field }) => field.id === id);
  }
}
