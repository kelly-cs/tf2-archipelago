import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { Subject, exhaustMap, tap } from 'rxjs';

import { SettingsStore, slugOf } from '@app/settings/settings-store';
import { appLink } from '@app/routing/app-routes';
import { Button } from '@app/ui/button';
import { SearchBox } from '@app/ui/search-box';
import { EmptyState } from '@app/ui/empty-state';
import { Notice } from '@app/ui/notice';

/**
 * The settings screen: the pages down the side, the rows in the middle, and
 * Save at the bottom.
 *
 * Which pages exist is decided at run time, so they come off the model rather
 * than a list here. Opening the screen is a call: the launcher makes the draft,
 * and closing it without saving throws the draft away.
 */
@Component({
  selector: 'app-settings-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, EmptyState, Notice, RouterLink, RouterLinkActive, RouterOutlet, SearchBox],
  templateUrl: './settings-page.html',
  styleUrl: './settings-page.scss',
})
export class SettingsPage {
  readonly store = inject(SettingsStore);
  private readonly router = inject(Router);

  readonly known = computed(() => this.store.known());
  readonly dirty = computed(() => this.store.dirty());
  readonly search = this.store.search;
  readonly open = computed(() => this.store.open());
  readonly restartNeeded = computed(() => this.store.restartNeeded());

  /** The pages, numbered. Which ones exist is decided at run time, so the
      number is the position in the model rather than anything declared here. */
  readonly pages = computed(() =>
    this.store.topTabs().map((tab, index) => ({
      title: tab.title,
      slug: slugOf(tab.title),
      number: String(index + 1).padStart(2, '0'),
      link: appLink.settingsTab(slugOf(tab.title)),
    })),
  );

  /** Where the player is, and what is either side of it. */
  readonly here = computed(() => this.pages().findIndex((page) => page.slug === this.store.page()));
  readonly previous = computed(() => this.pages()[this.here() - 1]);
  readonly next = computed(() => this.pages()[this.here() + 1]);
  readonly nextLabel = computed(() => {
    const next = this.next();
    return next === undefined ? 'Last section' : `Next: ${next.title} →`;
  });

  /** What the footer says about the draft: nothing to save, something to save,
      or something to save that the running server will not read until it is
      restarted. */
  readonly dirtyLabel = computed(() => {
    if (!this.store.dirty()) {
      return 'All saved';
    }
    return this.restartNeeded() ? 'Unsaved, needs a restart' : 'Unsaved changes';
  });

  /** What the launcher refused the last save for, empty when it took it. A
      signal, not a field: a computed over a field never runs again. */
  readonly refusal = signal('');

  /** What it said when it took one. */
  readonly saved = signal('');

  readonly openScreen = new Subject<void>();
  readonly save = new Subject<boolean>();
  readonly cancel = new Subject<void>();

  constructor() {
    this.openScreen
      .pipe(
        exhaustMap(() => this.store.openSettings('')),
        tap(() => this.goToFirstPage()),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.save
      .pipe(
        exhaustMap((restart) => this.store.save(restart)),
        tap((refusal) => {
          this.refusal.set(refusal);
          this.saved.set(refusal === '' ? 'Saved.' : '');
        }),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.cancel
      .pipe(
        exhaustMap(() => this.store.cancel()),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  private goToFirstPage(): void {
    const first = this.pages()[0];
    if (first !== undefined) {
      void this.router.navigate(first.link);
    }
  }
}
