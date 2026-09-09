import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { LauncherStore } from '@app/server/launcher-store';
import { ServerButtons } from '@app/shell/server-buttons';
import { StatusLine } from '@app/shell/status-line';
import { Logo } from '@app/ui/logo';
import { Notice } from '@app/ui/notice';
import { appLink } from '@app/routing/app-routes';

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
  readonly notice = computed(() => this.store.notice());

  readonly tabs = [
    { label: 'Session', link: appLink.session() },
    { label: 'Unlocks', link: appLink.unlocks() },
    { label: 'Bots', link: appLink.bots() },
    { label: 'Log', link: appLink.log() },
    { label: 'Settings', link: appLink.settings() },
  ];
}
