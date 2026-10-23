import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { appLink } from '@app/routing/app-routes';
import { LauncherStore } from '@app/server/launcher-store';
import { Panel } from '@app/ui/panel';

/**
 * Who is holding RED's other seats. Read-only here: changing it is the Bots
 * screen, a click away, because a seat swapped by accident from the screen you
 * watch while playing is a wave lost to the interface.
 */
@Component({
  selector: 'app-your-bot-team',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Panel, RouterLink],
  template: `
    <app-panel heading="Your bot team">
      <a panelAside [routerLink]="bots">Change &rarr;</a>

      <div class="seats">
        @for (seat of seats(); track seat.number) {
          <div class="seat">
            <span class="number">{{ seat.number }}</span>
            <span class="class" [class.drawn]="drawn(seat.class)">{{ seat.class }}</span>
            <span class="weapons">{{ seat.weapons || 'stock' }}</span>
          </div>
        }
      </div>

      <div class="line">{{ teamSizeLine() }}</div>
    </app-panel>
  `,
  styleUrl: './your-bot-team.scss',
})
export class YourBotTeam {
  private readonly store = inject(LauncherStore);

  readonly bots = appLink.bots();
  readonly seats = computed(() => this.store.bots());

  /** Valve tunes every wave for six defenders, so the line counts the humans
      in: a team of four with two players is a full one. */
  readonly teamSizeLine = computed(() => {
    const seats = this.seats().length;
    if (seats === 0) {
      return 'RED holds nobody yet.';
    }
    const drawn = this.store.drawnBots();
    const named = this.seats().filter((seat) => !this.drawn(seat.class)).length;
    const filled = `RED fills to ${seats}, humans included. ${named} named, ${seats - named} left to the mod.`;
    return drawn === '' ? filled : `${filled} ${drawn}`;
  });

  /** A seat the lineup does not name is one the mod draws for. */
  drawn(className: string): boolean {
    return className === 'the mod picks';
  }
}
