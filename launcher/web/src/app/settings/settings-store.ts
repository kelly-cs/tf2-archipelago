import { Injectable, computed, inject, linkedSignal, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Client, createClient } from '@connectrpc/connect';
import {
  EMPTY,
  Observable,
  Subject,
  concat,
  concatMap,
  debounceTime,
  defer,
  from,
  groupBy,
  last,
  map,
  mergeMap,
  of,
  switchMap,
} from 'rxjs';

import { LauncherStore } from '@app/server/launcher-store';
import { LAUNCHER_TRANSPORT } from '@app/transport/connect-transport';
import { Field, Model, Tab } from '@gen/tf2ap/launcher/v1/form_pb';
import { SettingsService } from '@gen/tf2ap/launcher/v1/settings_pb';

// How long to wait after the last keystroke before telling the launcher. Every
// change rebuilds the model, so a call per character would be a model per
// character.
const settleMs = 250;

/**
 * The settings screen, over the launcher's draft.
 *
 * The launcher owns the answers: every change goes to it and the rebuilt model
 * arrives on the stream, because a change can move another row's bounds and add
 * or remove rows outright. So this holds no settings of its own.
 *
 * What it does hold is the keystrokes not yet acknowledged. Without them a text
 * row would take the server's value back mid-word and put the caret with it.
 * An edit is dropped the moment the server agrees with it.
 */
@Injectable({ providedIn: 'root' })
export class SettingsStore {
  private readonly launcher = inject(LauncherStore);
  private readonly service: Client<typeof SettingsService> = createClient(
    SettingsService,
    inject(LAUNCHER_TRANSPORT),
  );

  readonly model = computed(() => this.launcher.screen());

  /** known is false until the first frame. Before it, the screen is not closed:
      it is not yet known, and drawing it as closed flashes a button that is
      about to vanish. */
  readonly known = computed(() => this.launcher.connected());
  readonly open = computed(() => this.model() !== undefined);
  readonly page = computed(() => this.launcher.screenPage());
  readonly missionPool = computed(() => this.launcher.missionPool());
  readonly restartNeeded = computed(() => this.launcher.restartNeeded());

  /** tabs are the top-level pages; a tab with `under` is a section of one. */
  /** Every row answered since the screen opened. Not the same as the pending
      edits: an answer the launcher has taken is still an answer Save has to
      write, because the launcher holds it in a draft rather than in the file. */
  /*
   * A plain signal, cleared where the clearing means something: when the screen
   * is opened, and when a save has written what was in it. It used to hang off
   * whether the screen was open, which looked equivalent and was not: the
   * launcher drops its draft and the browser reopens it in the same breath, and
   * a computation that is only re-run when something reads it never saw the
   * moment in between.
   */
  private readonly touched = signal(new Set<string>());

  /** dirty is the draft holding an answer the launcher has not written yet.
      The launcher clears its draft on Save, so a screen with rows on it and no
      pending edits is one where Save would write what is already there. */
  readonly dirty = computed(() => this.touched().size > 0);

  /**
   * What the player is looking for, across every page rather than the one they
   * are on. A setting you cannot name the page of is exactly the setting you
   * search for, so a search that only looked at the current page would miss
   * every time it mattered.
   */
  readonly search = signal('');

  readonly tabs = computed(() => this.model()?.tabs ?? []);

  /** matches is whether a row answers the search. Label and help both, because
      half of them are found by what they do rather than what they are called. */
  matches(field: Field): boolean {
    const needle = this.search().trim().toLowerCase();
    return (
      needle === '' ||
      field.label.toLowerCase().includes(needle) ||
      field.help.toLowerCase().includes(needle)
    );
  }

  /** pagesWithMatches is which pages have anything the search found. */
  readonly pagesWithMatches = computed(() => {
    const found = new Set<string>();
    for (const tab of this.tabs()) {
      const page = tab.under === '' ? tab.title : tab.under;
      if (tab.fields.some((field) => this.matches(field))) {
        found.add(page);
      }
    }
    return found;
  });
  readonly topTabs = computed(() => this.tabs().filter((tab) => tab.under === ''));

  private readonly typed = new Subject<{ id: string; value: string }>();

  constructor() {
    // Grouped by row: two rows answered in the same breath both land, and two
    // answers to the same row send only the last. concatMap inside the group,
    // so one row's answers reach the launcher in the order they were given.
    this.typed
      .pipe(
        groupBy((change) => change.id),
        mergeMap((ofOneRow) =>
          ofOneRow.pipe(
            debounceTime(settleMs),
            concatMap((change) => this.send(change.id, change.value)),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  private readonly edits = linkedSignal<Model | undefined, Record<string, string>>({
    source: () => this.model(),
    computation: (model, previous) => {
      const pending: Record<string, string> = {};
      for (const [id, typed] of Object.entries(previous?.value ?? {})) {
        if (serverValue(model, id) !== typed) {
          pending[id] = typed;
        }
      }
      return pending;
    },
  });

  /** value is what the row shows: what was typed if the launcher has not caught
      up, and the launcher's own answer once it has. */
  value(id: string): string {
    return this.edits()[id] ?? serverValue(this.model(), id) ?? '';
  }

  /** field finds one row by id, wherever it is. The Bots screen picks a handful
      out of the model rather than drawing a whole page of it. */
  field(id: string): Field | undefined {
    for (const tab of this.tabs()) {
      for (const candidate of tab.fields) {
        if (candidate.id === id) {
          return candidate;
        }
      }
    }
    return undefined;
  }

  /** fieldsMatching is every row whose id starts with a prefix, in model order:
      the seats are bots.seat.0.class, bots.seat.0.loadout, and so on. */
  fieldsMatching(prefix: string): Field[] {
    return this.tabs().flatMap((tab) => tab.fields.filter((one) => one.id.startsWith(prefix)));
  }

  tab(slug: string): Tab | undefined {
    return this.tabs().find((candidate) => slugOf(candidate.title) === slug);
  }

  /** sections are the tab itself plus every tab filed under it. */
  sections(slug: string): Tab[] {
    const parent = this.tab(slug);
    if (parent === undefined) {
      return [];
    }
    return [parent, ...this.tabs().filter((candidate) => candidate.under === parent.title)];
  }

  /**
   * change records the answer and tells the launcher about it a moment later.
   *
   * The wait is because every change rebuilds the model: a call per keystroke
   * would be a model per keystroke. The edit is recorded here and now though,
   * not when it is sent, so the row keeps showing what was typed and Save can
   * see what has not gone yet.
   */
  change(id: string, value: string): void {
    this.edits.update((pending) => ({ ...pending, [id]: value }));
    this.touched.update((seen) => new Set(seen).add(id));
    this.typed.next({ id, value });
  }

  /**
   * send tells the launcher one answer, once.
   *
   * Two paths reach here: the wait after the last keystroke, and Save hurrying
   * what has not gone yet. Without a guard they both send the same answer and
   * the launcher rebuilds the model twice for one keystroke.
   *
   * The guard is the launcher's own value, not a note of what was last sent.
   * A note goes stale: type 42, then 43, then 42 again, and a note would say 42
   * had already gone while the launcher was holding 43. Asking what the
   * launcher has cannot be wrong about it.
   */
  private send(id: string, value: string): Observable<void> {
    return defer(() => {
      if (serverValue(this.model(), id) === value) {
        return EMPTY;
      }
      return from(this.service.changeSetting({ change: { field: id, value } }));
    }).pipe(map(() => undefined));
  }

  dispatch(id: string): Observable<void> {
    return from(this.service.dispatchAction({ id })).pipe(map(() => undefined));
  }

  openSettings(page: string): Observable<void> {
    this.touched.set(new Set());
    return from(this.service.openSettings({ page })).pipe(map(() => undefined));
  }

  cancel(): Observable<void> {
    return from(this.service.cancelSettings({})).pipe(map(() => undefined));
  }

  /**
   * Save writes the file, and answers with a refusal or nothing.
   *
   * Every answer still in flight goes first. A row is sent a quarter of a
   * second after the last keystroke, and pressing Save inside that quarter
   * second used to write the value from before the word was finished: the
   * player watched their own typing be discarded by the button meant to keep
   * it.
   */
  save(restart: boolean): Observable<string> {
    return concat(
      this.flush(),
      defer(() => this.write(restart)),
    ).pipe(
      last(),
      // The launcher drops its draft once the file is written, which in a
      // window meant the dialog closed. Here the settings are a tab the player
      // is standing on, so a save that emptied the screen would throw them out
      // of it. Reopening gives them the saved values back in place.
      switchMap((refusal) =>
        refusal === '' ? this.reopen().pipe(map(() => refusal)) : of(refusal),
      ),
    );
  }

  private reopen(): Observable<void> {
    return this.openSettings(this.launcher.screenPage());
  }

  /**
   * flush sends every answer that has not gone yet.
   *
   * Deferred, and so is the save after it. A Connect call is a Promise, which
   * starts the moment it is built rather than the moment it is subscribed to:
   * without this the file was written before the flush had sent anything, which
   * is the whole thing this exists to prevent.
   */
  private flush(): Observable<never> {
    return defer(() => from(Object.entries(this.edits()))).pipe(
      concatMap(([id, value]) => this.send(id, value)),
      concatMap(() => EMPTY),
    );
  }

  private write(restart: boolean): Observable<string> {
    return from(this.service.saveSettings({ restart })).pipe(
      map((answer) => (answer.saved ? '' : answer.refusal)),
    );
  }
}

/** slugOf turns a tab title into the segment that names it in the URL. Which
    tabs exist is decided at run time, so the URL cannot hold a declared name. */
export function slugOf(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

export function fieldsOf(tab: Tab | undefined): Field[] {
  return tab?.fields ?? [];
}

function serverValue(model: Model | undefined, id: string): string | undefined {
  for (const tab of model?.tabs ?? []) {
    for (const field of tab.fields) {
      if (field.id === id) {
        return field.value;
      }
    }
  }
  return undefined;
}
