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
import { Subject, concatMap, debounceTime, groupBy, mergeMap, tap } from 'rxjs';

import { MissionTable } from '@app/settings/components/mission-table';
import { SettingsRow } from '@app/settings/components/settings-row';
import { SettingsActions } from '@app/settings/settings-actions';
import { SettingsStore } from '@app/settings/settings-store';
import { EmptyState } from '@app/ui/empty-state';
import { Notice } from '@app/ui/notice';
import { SearchBox } from '@app/ui/search-box';
import { Field } from '@gen/tf2ap/launcher/v1/form_pb';

// How long to wait after the last keystroke before telling the launcher. Every
// change rebuilds the model, so a call per character would be a model per
// character.
const settleMs = 250;

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
  imports: [EmptyState, MissionTable, Notice, SearchBox, SettingsRow],
  templateUrl: './settings-tab-page.html',
  styleUrl: './settings-tab-page.scss',
})
export class SettingsTabPage {
  private readonly store = inject(SettingsStore);
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

  readonly shown = computed(() => {
    const needle = this.filter().trim().toLowerCase();
    if (needle === '') {
      return this.rows();
    }
    return this.rows().filter(
      ({ field }) =>
        field.label.toLowerCase().includes(needle) || field.help.toLowerCase().includes(needle),
    );
  });

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

  readonly typed = new Subject<{ id: string; value: string }>();
  readonly fired = new Subject<string>();

  constructor() {
    // Grouped by row: two rows edited in the same breath both land, and two
    // edits of the same row send only the last. concatMap inside the group so
    // one row's answers reach the launcher in the order they were given.
    this.typed
      .pipe(
        groupBy((change) => change.id),
        mergeMap((ofOneRow) =>
          ofOneRow.pipe(
            debounceTime(settleMs),
            concatMap((change) => this.store.change(change.id, change.value)),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.fired
      .pipe(
        concatMap((id) => this.actions.press(id)),
        tap((said) => this.said.set(said)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  value(id: string): string {
    return this.store.value(id);
  }

  indexOf(id: string): number {
    return this.rows().findIndex(({ field }) => field.id === id);
  }
}
