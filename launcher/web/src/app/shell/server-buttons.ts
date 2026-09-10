import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap, filter, tap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { Button } from '@app/ui/button';
import { ServerStatus } from '@gen/tf2ap/launcher/v1/launcher_pb';

/**
 * Join, Start, Stop, Restart and Quit. Join is first and yellow: once the
 * server is up it is the one button the player came for, and the Play page
 * is a click further than the header. One button says Start or Stop depending on
 * what the server is doing: two buttons where one is always wrong is how a
 * player presses Start on a running server.
 *
 * Stop is never disabled. Starting is when it is needed most: an install that
 * is a few gigabytes in, or a server waiting on Steam for an address it will
 * not get, is stopped by this button and by nothing else.
 *
 * Every press goes through exhaustMap, so holding the button down sends one
 * call rather than one per click while the first is still in flight.
 */
@Component({
  selector: 'app-server-buttons',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button],
  template: `
    <app-button
      tone="primary"
      [disabled]="!joinable()"
      [hint]="joinable() ? 'Open TF2 and connect to this server' : 'Start the server to join it'"
      (press)="join.next()"
    >
      Join
    </app-button>
    <app-button [tone]="halting() ? 'halt' : 'go'" (press)="toggle.next()">
      {{ halting() ? 'Stop server' : 'Start server' }}
    </app-button>
    <app-button tone="ghost" [disabled]="!running()" (press)="restart.next()">Restart</app-button>
    <app-button tone="ghost" hint="Close the launcher" (press)="quit.next()">Quit</app-button>
  `,
  styles: `
    :host {
      display: flex;
      gap: var(--gap-sm);
    }
  `,
})
export class ServerButtons {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);

  readonly running = computed(() => this.store.running());
  readonly joinable = computed(() => this.store.joinUrl() !== '');

  /** halting is the button meaning Stop: the server is up, or on its way up. */
  readonly halting = computed(
    () =>
      this.store.running() || this.store.busy() || this.store.status() === ServerStatus.STARTING,
  );

  readonly join = new Subject<void>();
  readonly toggle = new Subject<void>();
  readonly restart = new Subject<void>();
  readonly quit = new Subject<void>();

  constructor() {
    // A steam:// link: the browser hands it to Steam, which connects the game.
    this.join
      .pipe(
        filter(() => this.joinable()),
        tap(() => (window.location.href = this.store.joinUrl())),
        takeUntilDestroyed(),
      )
      .subscribe();
    this.toggle
      .pipe(
        exhaustMap(() => (this.halting() ? this.commands.stop() : this.commands.start())),
        takeUntilDestroyed(),
      )
      .subscribe();
    this.restart
      .pipe(
        exhaustMap(() => this.commands.restart()),
        takeUntilDestroyed(),
      )
      .subscribe();
    this.quit
      .pipe(
        exhaustMap(() => this.commands.quit()),
        takeUntilDestroyed(),
      )
      .subscribe();
  }
}
