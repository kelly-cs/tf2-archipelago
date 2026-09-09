import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';

import { LauncherStore } from '@app/server/launcher-store';
import { StatusDot } from '@app/ui/status-dot';
import { ServerStatus } from '@gen/tf2ap/launcher/v1/launcher_pb';

/**
 * The line under the title: what the server is doing, and what room and
 * mission it is doing it for. Starting pulses because there is nothing to
 * report but that it is still happening.
 */
@Component({
  selector: 'app-status-line',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [StatusDot],
  template: `
    <app-status-dot [tone]="tone()" [pulsing]="starting()" />
    <span>{{ words() }}</span>
  `,
  styles: `
    :host {
      display: flex;
      align-items: center;
      gap: 7px;
      margin-top: var(--gap-xs);
      font-size: var(--text-sm);
      color: var(--text-dim);
    }
  `,
})
export class StatusLine {
  private readonly store = inject(LauncherStore);

  readonly starting = computed(() => this.store.status() === ServerStatus.STARTING);

  readonly tone = computed(() => {
    if (this.store.lost()) {
      return 'bad' as const;
    }
    switch (this.store.status()) {
      case ServerStatus.RUNNING:
        return 'good' as const;
      case ServerStatus.STARTING:
        return 'warn' as const;
      default:
        return 'idle' as const;
    }
  });

  readonly words = computed(() => {
    if (this.store.lost()) {
      return 'launcher unreachable';
    }
    if (!this.store.connected()) {
      return 'connecting';
    }
    const state = statusWord(this.store.status());
    const room = this.store.room();
    return room === '' ? state : `${state}, ${room}`;
  });
}

function statusWord(status: ServerStatus): string {
  switch (status) {
    case ServerStatus.RUNNING:
      return 'running';
    case ServerStatus.STARTING:
      return 'starting';
    case ServerStatus.STOPPED:
      return 'stopped';
    default:
      return 'unknown';
  }
}
