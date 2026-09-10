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

  /**
   * Who this server is, in one line: what it is doing, which room it answers
   * to, and which slot it plays. Separated by middots because each part is a
   * different kind of fact and a comma would read as a list of one thing.
   */
  readonly words = computed(() => {
    if (this.store.lost()) {
      return 'Launcher unreachable';
    }
    if (!this.store.connected()) {
      return 'Connecting';
    }
    const parts = [statusWord(this.store.status())];
    parts.push(this.store.room() || 'no room connected');
    if (this.store.slot()) {
      parts.push(`slot ${this.store.slot()}`);
    }
    return parts.join(' \u00b7 ');
  });
}

function statusWord(status: ServerStatus): string {
  switch (status) {
    case ServerStatus.RUNNING:
      return 'Running';
    case ServerStatus.STARTING:
      return 'Starting';
    case ServerStatus.STOPPED:
      return 'Stopped';
    default:
      return 'Unknown';
  }
}
