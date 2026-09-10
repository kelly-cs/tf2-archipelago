import { ChangeDetectionStrategy, Component } from '@angular/core';

import { JoinPanel } from '@app/session/components/join-panel';
import { MissionList } from '@app/session/components/mission-list';
import { WhatYouCanPlay } from '@app/session/components/what-you-can-play';
import { YourBotTeam } from '@app/session/components/your-bot-team';

/**
 * The screen a player watches while they play: how to get in, what the run has
 * handed them, who is on their team, and what is left to play.
 */
@Component({
  selector: 'app-session-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [JoinPanel, MissionList, WhatYouCanPlay, YourBotTeam],
  template: `
    <div class="cards">
      <app-join-panel />
      <app-what-you-can-play />
      <app-your-bot-team />
    </div>
    <app-mission-list />
  `,
  styles: `
    :host {
      display: flex;
      flex-direction: column;
      gap: var(--gap-lg);
    }

    .cards {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: var(--gap-lg);
    }
  `,
})
export class SessionPage {}
