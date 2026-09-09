import { ChangeDetectionStrategy, Component } from '@angular/core';

import { JoinPanel } from '@app/session/components/join-panel';
import { MissionList } from '@app/session/components/mission-list';
import { RunGlance } from '@app/session/components/run-glance';

/** The screen a player watches while they play: how to get in, and what is left. */
@Component({
  selector: 'app-session-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [JoinPanel, MissionList, RunGlance],
  template: `
    <div class="top">
      <app-join-panel />
      <app-run-glance />
    </div>
    <app-mission-list />
  `,
  styles: `
    :host {
      display: flex;
      flex-direction: column;
      gap: var(--gap-lg);
      animation: fade-in var(--settle) var(--ease);
    }

    .top {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
      gap: var(--gap-lg);
    }
  `,
})
export class SessionPage {}
