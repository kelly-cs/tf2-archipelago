import { Injectable, computed, inject, linkedSignal } from '@angular/core';
import { Client, createClient } from '@connectrpc/connect';
import { Observable, from, map } from 'rxjs';

import { LauncherStore } from '@app/server/launcher-store';
import { LAUNCHER_TRANSPORT } from '@app/transport/connect-transport';
import { Field, Model, Tab } from '@gen/tf2ap/launcher/v1/form_pb';
import { SettingsService } from '@gen/tf2ap/launcher/v1/settings_pb';

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
  readonly tabs = computed(() => this.model()?.tabs ?? []);
  readonly topTabs = computed(() => this.tabs().filter((tab) => tab.under === ''));

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

  change(id: string, value: string): Observable<void> {
    this.edits.update((pending) => ({ ...pending, [id]: value }));
    return from(this.service.changeSetting({ change: { field: id, value } })).pipe(
      map(() => undefined),
    );
  }

  dispatch(id: string): Observable<void> {
    return from(this.service.dispatchAction({ id })).pipe(map(() => undefined));
  }

  openSettings(page: string): Observable<void> {
    return from(this.service.openSettings({ page })).pipe(map(() => undefined));
  }

  cancel(): Observable<void> {
    return from(this.service.cancelSettings({})).pipe(map(() => undefined));
  }

  save(restart: boolean): Observable<string> {
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
