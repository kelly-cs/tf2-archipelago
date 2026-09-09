import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { Badge } from '@app/ui/badge';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { Panel } from '@app/ui/panel';
import { SessionMission } from '@gen/tf2ap/launcher/v1/launcher_pb';

/**
 * The run's missions and what has happened to each. Played and cleared are not
 * the same thing: another world's !collect sends every check it still holds, so
 * the room can hold a mission's check that this server never played.
 */
@Component({
  selector: 'app-mission-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge, Button, EmptyState, Panel],
  template: `
    <app-panel heading="Missions">
      @if (missions().length === 0) {
        <app-empty-state>
          No missions yet. They arrive once the bridge is on the multiworld.
        </app-empty-state>
      } @else {
        <table>
          <thead>
            <tr>
              <th scope="col">Mission</th>
              <th scope="col">Map</th>
              <th scope="col">Waves</th>
              <th scope="col">State</th>
              <th scope="col"><span class="sr-only">Play</span></th>
            </tr>
          </thead>
          <tbody>
            @for (mission of missions(); track mission.popFile) {
              <tr [class.playing]="mission.popFile === playing()">
                <td>
                  {{ mission.name }}
                  @if (mission.loadout) {
                    <app-badge tone="info">{{ mission.loadout }}</app-badge>
                  }
                </td>
                <td>{{ mission.map }}</td>
                <td class="waves">{{ mission.waves }}</td>
                <td>
                  <app-badge [tone]="stateTone(mission)">{{ stateWord(mission) }}</app-badge>
                </td>
                <td>
                  <app-button
                    size="small"
                    [disabled]="!mission.unlocked || !running()"
                    (press)="choose.next(mission.popFile)"
                  >
                    Play next
                  </app-button>
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </app-panel>
  `,
  styleUrl: './mission-list.scss',
})
export class MissionList {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);

  readonly missions = computed(() => this.store.missions());
  readonly playing = computed(() => this.store.mission());
  readonly running = computed(() => this.store.running());

  readonly choose = new Subject<string>();

  constructor() {
    this.choose
      .pipe(
        exhaustMap((popFile) => this.commands.setMission(popFile)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  stateWord(mission: SessionMission): string {
    if (mission.played) {
      return 'played';
    }
    if (mission.cleared) {
      return 'cleared elsewhere';
    }
    return mission.unlocked ? 'unlocked' : 'locked';
  }

  stateTone(mission: SessionMission): 'good' | 'info' | 'neutral' {
    if (mission.played) {
      return 'good';
    }
    return mission.unlocked ? 'info' : 'neutral';
  }
}
