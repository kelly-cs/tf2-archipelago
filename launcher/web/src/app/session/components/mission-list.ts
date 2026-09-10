import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { Tier } from '@app/ui/tier';
import { SessionMission } from '@gen/tf2ap/launcher/v1/launcher_pb';

type Column = 'name' | 'map' | 'tier' | 'source' | 'waves' | 'state';

/** One mission as the table draws it. */
interface Row {
  readonly mission: SessionMission;
  readonly state: string;
  readonly tone: string;
  readonly order: number;
}

// What a state means, and the order the table sorts them in: what you can play
// now first, what you have done last.
const states: Record<string, number> = { unlocked: 0, locked: 1, elsewhere: 2, played: 3 };

/**
 * The run's missions and what has happened to each.
 *
 * Played and cleared are not the same thing. Another world's !collect sends
 * every check it still holds, so the room can hold a mission's check that this
 * server never played, and a table saying only "cleared" would tell a player
 * they had finished a mission they had never loaded.
 */
@Component({
  selector: 'app-mission-list',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, EmptyState, Tier],
  templateUrl: './mission-list.html',
  styleUrl: './mission-list.scss',
})
export class MissionList {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);

  readonly sortBy = signal<Column>('state');
  readonly ascending = signal(true);

  readonly columns: { key: Column; label: string }[] = [
    { key: 'name', label: 'Mission' },
    { key: 'map', label: 'Map' },
    { key: 'tier', label: 'Tier' },
    { key: 'source', label: 'Archive' },
    { key: 'waves', label: 'Waves' },
    { key: 'state', label: 'State' },
  ];

  readonly playing = computed(() => this.store.mission());
  readonly running = computed(() => this.store.running());

  readonly rows = computed<Row[]>(() => {
    const rows = this.store.missions().map((mission) => describe(mission));
    const key = this.sortBy();
    const direction = this.ascending() ? 1 : -1;
    return rows.toSorted((left, right) => direction * compare(left, right, key));
  });

  /** What the multiworld has to do with this list, in one line. */
  readonly multiworldLine = computed(() => {
    const rows = this.rows();
    if (rows.length === 0) {
      return '';
    }
    const unlocked = rows.filter((row) => row.mission.unlocked).length;
    const played = rows.filter((row) => row.mission.played).length;
    return `${unlocked} of ${rows.length} unlocked, ${played} played`;
  });

  readonly switchHint = computed(() =>
    this.running()
      ? 'Play next loads the mission on the running server. Anyone on it is sent to the new map.'
      : 'Start the server to load a mission.',
  );

  readonly choose = new Subject<string>();

  constructor() {
    this.choose
      .pipe(
        exhaustMap((popFile) => this.commands.setMission(popFile)),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  sortOn(key: Column): void {
    if (this.sortBy() === key) {
      this.ascending.set(!this.ascending());
      return;
    }
    this.sortBy.set(key);
    this.ascending.set(true);
  }
}

function describe(mission: SessionMission): Row {
  if (mission.played) {
    return { mission, state: 'played', tone: 'good', order: states['played'] };
  }
  if (mission.cleared) {
    return { mission, state: 'cleared elsewhere', tone: 'info', order: states['elsewhere'] };
  }
  if (mission.unlocked) {
    return { mission, state: 'unlocked', tone: 'accent', order: states['unlocked'] };
  }
  return { mission, state: 'locked', tone: 'neutral', order: states['locked'] };
}

function compare(left: Row, right: Row, key: Column): number {
  if (key === 'state') {
    return left.order - right.order || left.mission.name.localeCompare(right.mission.name);
  }
  if (key === 'waves') {
    return left.mission.waves - right.mission.waves;
  }
  return left.mission[key].localeCompare(right.mission[key]);
}
