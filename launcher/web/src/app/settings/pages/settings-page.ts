import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { Subject, exhaustMap, tap } from 'rxjs';

import { SettingsStore, slugOf } from '@app/settings/settings-store';
import { appLink } from '@app/routing/app-routes';
import { Button } from '@app/ui/button';
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
  imports: [Button, EmptyState, Notice, RouterLink, RouterLinkActive, RouterOutlet],
  templateUrl: './settings-page.html',
  styleUrl: './settings-page.scss',
})
export class SettingsPage {
  private readonly store = inject(SettingsStore);
  private readonly router = inject(Router);

  readonly known = computed(() => this.store.known());
  readonly open = computed(() => this.store.open());
  readonly restartNeeded = computed(() => this.store.restartNeeded());

  readonly pages = computed(() =>
    this.store.topTabs().map((tab) => ({
      title: tab.title,
      link: appLink.settingsTab(slugOf(tab.title)),
    })),
  );

  /** What the launcher refused the last save for, empty when it took it. A
      signal, not a field: a computed over a field never runs again. */
  readonly refusal = signal('');

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
        tap((refusal) => this.refusal.set(refusal)),
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
