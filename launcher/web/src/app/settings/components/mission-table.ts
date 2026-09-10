import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  output,
  signal,
} from '@angular/core';

import { SettingsStore } from '@app/settings/settings-store';
import { Badge } from '@app/ui/badge';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { SearchBox } from '@app/ui/search-box';
import { MissionPoolRow } from '@gen/tf2ap/launcher/v1/launcher_pb';

type Column = 'in_pool' | 'source' | 'map' | 'name' | 'waves' | 'compatibility' | 'mods';

/** One drawn row: the pool data, plus the tick that decides whether it plays. */
interface Row {
  readonly pool: MissionPoolRow;
  readonly on: boolean;
  readonly usable: boolean;
  readonly reason: string;
}

/**
 * The mission pool: twenty-six rows and more with the community packs on, so it
 * is a table with columns that sort and a search over it rather than a list.
 *
 * Compatibility is the launcher's word, not a guess made here. A row that is
 * not Ready still ticks: a mission the seed cannot draw yet is a mission the
 * player may be about to make drawable by ticking a pack.
 */
@Component({
  selector: 'app-mission-table',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge, Button, EmptyState, SearchBox],
  templateUrl: './mission-table.html',
  styleUrl: './mission-table.scss',
})
export class MissionTable {
  readonly store = inject(SettingsStore);

  readonly typed = output<{ id: string; value: string }>();

  readonly filter = signal('');
  readonly sortBy = signal<Column>('map');
  readonly ascending = signal(true);

  readonly columns: { key: Column; label: string }[] = [
    { key: 'in_pool', label: 'In pool' },
    { key: 'source', label: 'Archive' },
    { key: 'map', label: 'Map' },
    { key: 'name', label: 'Mission' },
    { key: 'waves', label: 'Waves' },
    { key: 'compatibility', label: 'Compatibility' },
    { key: 'mods', label: 'Server mod' },
  ];

  private readonly all = computed<Row[]>(() =>
    this.store.missionPool().map((pool) => ({
      pool,
      on: this.store.value(pool.field) === 'true',
      usable: pool.compatibility === 'Ready',
      reason: pool.compatibility,
    })),
  );

  readonly rows = computed(() => {
    const needle = this.filter().trim().toLowerCase();
    const key = this.sortBy();
    const direction = this.ascending() ? 1 : -1;
    return this.all()
      .filter((row) => needle === '' || matches(row.pool, needle))
      .toSorted((left: Row, right: Row) => direction * compare(left, right, key));
  });

  readonly chosen = computed(() => this.all().filter((row) => row.on).length);

  sortOn(key: Column): void {
    if (this.sortBy() === key) {
      this.ascending.set(!this.ascending());
      return;
    }
    this.sortBy.set(key);
    this.ascending.set(true);
  }

  /** Ticks every row on screen, not every row there is: the search is part of
      what the player meant by "all". */
  setShown(on: boolean): void {
    for (const row of this.rows()) {
      if (row.on !== on) {
        this.typed.emit({ id: row.pool.field, value: on ? 'true' : 'false' });
      }
    }
  }

  flip(row: Row): void {
    this.typed.emit({ id: row.pool.field, value: row.on ? 'false' : 'true' });
  }

  tone(row: Row): 'good' | 'warn' {
    return row.usable ? 'good' : 'warn';
  }
}

function matches(pool: MissionPoolRow, needle: string): boolean {
  return (
    pool.name.toLowerCase().includes(needle) ||
    pool.map.toLowerCase().includes(needle) ||
    pool.source.toLowerCase().includes(needle) ||
    pool.compatibility.toLowerCase().includes(needle)
  );
}

/**
 * How two rows compare on one column.
 *
 * Three kinds of column, and text order is wrong for two of them. The tick is
 * what the player came to sort by: seeing the pool together is the question the
 * screen answers. And waves read "1-10", which sorts before "1-6" as text.
 */
function compare(left: Row, right: Row, key: Column): number {
  if (key === 'in_pool') {
    return Number(left.on) - Number(right.on) || left.pool.name.localeCompare(right.pool.name);
  }
  if (key === 'waves') {
    return wavesOf(left.pool.waves) - wavesOf(right.pool.waves);
  }
  return left.pool[key].localeCompare(right.pool[key]);
}

/** The wave count out of "1-6". The launcher writes the range; the number at
    the end of it is what a mission is long or short by. */
function wavesOf(waves: string): number {
  return Number(/\d+$/.exec(waves)?.[0] ?? 0);
}
