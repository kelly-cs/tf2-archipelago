import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { Badge } from '@app/ui/badge';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { Notice } from '@app/ui/notice';
import { Panel } from '@app/ui/panel';

/** The nine classes, in the order the game lists them. */
const classes = [
  'scout',
  'soldier',
  'pyro',
  'demoman',
  'heavyweapons',
  'engineer',
  'medic',
  'sniper',
  'spy',
] as const;

/**
 * The Bot Switcher: which classes the mod may draw for RED's empty seats, and
 * what it drew.
 *
 * Applying is one console command, taken between waves. A wave in progress is
 * left alone: swapping a defender mid-wave is how a run is lost to the
 * interface rather than to the robots.
 */
@Component({
  selector: 'app-bots-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge, Button, EmptyState, Notice, Panel],
  template: `
    <app-panel heading="Your bot team">
      <app-badge panelAside tone="neutral">{{ seats().length }} seats</app-badge>
      @if (seats().length === 0) {
        <app-empty-state>RED holds nobody yet. Start the server.</app-empty-state>
      } @else {
        <table>
          <thead>
            <tr>
              <th scope="col">Seat</th>
              <th scope="col">Class</th>
              <th scope="col">Weapons</th>
            </tr>
          </thead>
          <tbody>
            @for (seat of seats(); track seat.number) {
              <tr>
                <td class="seat">{{ seat.number }}</td>
                <td>{{ seat.class }}</td>
                <td class="weapons">{{ seat.weapons || 'stock' }}</td>
              </tr>
            }
          </tbody>
        </table>
      }
      @if (drawn()) {
        <p class="drawn">{{ drawn() }}</p>
      }
    </app-panel>

    <app-panel heading="Bot Switcher">
      <p class="lead">Classes the mod may draw for a seat you have not named.</p>
      <div class="chips">
        @for (name of classNames; track name) {
          <app-button
            size="small"
            [tone]="allowed().has(name) ? 'primary' : 'ghost'"
            [pressed]="allowed().has(name)"
            (press)="flip(name)"
          >
            {{ name }}
          </app-button>
        }
      </div>
      @if (!running()) {
        <app-notice tone="info">
          The server is not up. The lineup applies the next time it starts.
        </app-notice>
      }
      <div class="apply">
        <app-button tone="primary" [disabled]="!running()" (press)="apply.next()">
          Apply between waves
        </app-button>
      </div>
    </app-panel>
  `,
  styleUrl: './bots-page.scss',
})
export class BotsPage {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);

  readonly classNames = classes;
  readonly seats = computed(() => this.store.bots());
  readonly drawn = computed(() => this.store.drawnBots());
  readonly running = computed(() => this.store.running());

  readonly allowed = signal(new Set<string>(classes));

  readonly apply = new Subject<void>();

  constructor() {
    this.apply
      .pipe(
        exhaustMap(() => this.commands.sendRcon(`sm_ap_bots ${[...this.allowed()].join(',')}`)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  flip(name: string): void {
    const next = new Set(this.allowed());
    if (next.has(name)) {
      next.delete(name);
    } else {
      next.add(name);
    }
    // Never all nine off: a lineup the mod cannot draw from leaves RED empty.
    if (next.size > 0) {
      this.allowed.set(next);
    }
  }
}
