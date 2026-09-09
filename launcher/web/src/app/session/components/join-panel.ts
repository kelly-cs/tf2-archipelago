import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap, from, tap, timer } from 'rxjs';

import { LauncherStore } from '@app/server/launcher-store';
import { Button } from '@app/ui/button';
import { Panel } from '@app/ui/panel';

// How long the Copy button says it copied before going back to saying Copy.
const copiedForMs = 1600;

/**
 * How the player gets into the server: the connect line, a Steam link, and a
 * copy for the console. Steam is a link the browser follows, not a call: the
 * launcher would otherwise be launching a game on behalf of whatever machine
 * happens to be looking at the page.
 */
@Component({
  selector: 'app-join-panel',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, Panel],
  template: `
    <app-panel heading="Join the server">
      <div class="line">{{ join() || 'The server is not up yet.' }}</div>
      <div class="buttons">
        <app-button tone="primary" size="large" [disabled]="!ready()" (press)="steam.next()">
          Join with Steam
        </app-button>
        <app-button size="large" [disabled]="!join()" (press)="copy.next()">
          {{ copied() ? 'Copied' : 'Copy the line' }}
        </app-button>
      </div>
      <p>Or open the TF2 console and paste the line above.</p>
      @if (store.itemServer()) {
        <div class="note">{{ store.itemServer() }}</div>
      }
    </app-panel>
  `,
  styles: `
    .line {
      font-family: var(--font-mono);
      font-size: 19px;
      word-break: break-all;
      background: var(--surface-void);
      border: 1px solid var(--line);
      border-radius: var(--radius);
      padding: 12px 14px;
    }

    .buttons {
      display: flex;
      gap: var(--gap-sm);
      flex-wrap: wrap;
    }

    .buttons app-button {
      flex: 1;
      min-width: 140px;
    }

    p {
      margin: 0;
      color: var(--text-dim);
    }

    .note {
      font-size: var(--text-sm);
      color: var(--text-faint);
    }
  `,
})
export class JoinPanel {
  readonly store = inject(LauncherStore);

  readonly join = computed(() => this.store.join());
  readonly ready = computed(() => this.store.joinUrl() !== '');
  readonly copied = signal(false);

  readonly steam = new Subject<void>();
  readonly copy = new Subject<void>();

  constructor() {
    this.steam
      .pipe(
        tap(() => (window.location.href = this.store.joinUrl())),
        takeUntilDestroyed(),
      )
      .subscribe();

    this.copy
      .pipe(
        exhaustMap(() =>
          from(navigator.clipboard.writeText(this.join())).pipe(
            tap(() => this.copied.set(true)),
            exhaustMap(() => timer(copiedForMs)),
            tap(() => this.copied.set(false)),
          ),
        ),
        takeUntilDestroyed(),
      )
      .subscribe();
  }
}
