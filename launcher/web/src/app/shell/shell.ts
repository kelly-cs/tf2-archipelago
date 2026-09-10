import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { filter, switchMap, tap, timer } from 'rxjs';

import { LauncherStore } from '@app/server/launcher-store';
import { ServerButtons } from '@app/shell/server-buttons';
import { StatusLine } from '@app/shell/status-line';
import { Logo } from '@app/ui/logo';
import { Notice } from '@app/ui/notice';
import { appLink } from '@app/routing/app-routes';

// How long the launcher's last word stays on screen. Long enough to read a
// sentence, short enough that it is gone before it becomes furniture.
const noticeForMs = 8_000;

/**
 * The frame every screen sits in: the mark and the state at the top, the
 * sections below it, the screen under that.
 *
 * The tabs are links rather than buttons, so a refresh and the browser's own
 * Back both land where the player was.
 */
@Component({
  selector: 'app-shell',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Logo, Notice, RouterLink, RouterLinkActive, RouterOutlet, ServerButtons, StatusLine],
  templateUrl: './shell.html',
  styleUrl: './shell.scss',
})
export class Shell {
  private readonly store = inject(LauncherStore);

  readonly title = computed(() => this.store.title() || 'Mann vs Archipelago');
  readonly lost = computed(() => this.store.lost());

  /**
   * The last thing the launcher said, while it is still worth saying.
   *
   * Driven by the sequence number rather than the words: the launcher says
   * "settings saved" every time it saves, and a screen watching the text alone
   * shows the second save nothing at all. The number changing is the event.
   */
  readonly notice = signal('');

  constructor() {
    toObservable(this.store.noticeSeq)
      .pipe(
        filter((seq) => seq > 0n),
        tap(() => this.notice.set(this.store.notice())),
        switchMap(() => timer(noticeForMs)),
        tap(() => this.notice.set('')),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  /** What the sections are called. The URL keeps its own words: /session and
      /log are what a bug report pastes, and renaming a route renames a link
      somebody already has. */
  readonly tabs = [
    { label: 'Play', link: appLink.session() },
    { label: 'Unlocks', link: appLink.unlocks() },
    { label: 'Bots', link: appLink.bots() },
    { label: 'Console', link: appLink.log() },
    { label: 'Settings', link: appLink.settings() },
  ];
}
