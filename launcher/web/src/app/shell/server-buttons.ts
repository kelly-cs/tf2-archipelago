import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { Button } from '@app/ui/button';
import { ServerStatus } from '@gen/tf2ap/launcher/v1/launcher_pb';

/**
 * Start, Stop, Restart and Quit. One button says Start or Stop depending on
 * what the server is doing: two buttons where one is always wrong is how a
 * player presses Start on a running server.
 *
 * Every press goes through exhaustMap, so holding the button down sends one
 * call rather than one per click while the first is still in flight.
 */
@Component({
  selector: 'app-server-buttons',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button],
  template: `
    <app-button [tone]="running() ? 'halt' : 'go'" [disabled]="busy()" (press)="toggle.next()">
      {{ running() ? 'Stop server' : 'Start server' }}
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
  readonly busy = computed(
    () => this.store.busy() || this.store.status() === ServerStatus.STARTING,
  );

  readonly toggle = new Subject<void>();
  readonly restart = new Subject<void>();
  readonly quit = new Subject<void>();

  constructor() {
    this.toggle
      .pipe(
        exhaustMap(() => (this.running() ? this.commands.stop() : this.commands.start())),
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
