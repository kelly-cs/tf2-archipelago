import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';

import { LauncherStore } from '@app/server/launcher-store';
import { Badge } from '@app/ui/badge';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { Panel } from '@app/ui/panel';
import { SearchBox } from '@app/ui/search-box';

/** One unlock as a row, with the same buff held twice shown as a level. */
interface Row {
  readonly kind: string;
  readonly name: string;
  readonly level: number;
}

/**
 * Everything the run has handed this slot. The bridge serves it by key, the
 * launcher turns the keys into the names a person uses, and this only groups
 * and filters them.
 */
@Component({
  selector: 'app-unlocks-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Badge, Button, EmptyState, Panel, SearchBox],
  template: `
    <app-panel heading="Everything unlocked">
      <div panelAside class="count">{{ rows().length }} of {{ all().length }}</div>

      <div class="bar">
        <app-search-box label="Find an unlock" [(text)]="filter" />
        @for (kind of kinds(); track kind) {
          <app-button
            size="small"
            [tone]="kind === chosenKind() ? 'primary' : 'ghost'"
            [pressed]="kind === chosenKind()"
            (press)="chosenKind.set(kind === chosenKind() ? '' : kind)"
          >
            {{ kind }}
          </app-button>
        }
      </div>

      @if (rows().length === 0) {
        <app-empty-state>
          @if (all().length === 0) {
            Nothing yet. Unlocks arrive as the multiworld sends them.
          } @else {
            No unlock matches that.
          }
        </app-empty-state>
      } @else {
        <ul>
          @for (row of rows(); track row.kind + row.name) {
            <li>
              <app-badge tone="accent">{{ row.kind }}</app-badge>
              <span class="name">{{ row.name }}</span>
              @if (row.level > 1) {
                <app-badge tone="good">level {{ row.level }}</app-badge>
              }
            </li>
          }
        </ul>
      }
    </app-panel>
  `,
  styleUrl: './unlocks-page.scss',
})
export class UnlocksPage {
  private readonly store = inject(LauncherStore);

  readonly filter = signal('');
  readonly chosenKind = signal('');

  readonly all = computed<Row[]>(() =>
    this.store.unlocks().map((unlock) => ({
      kind: unlock.kind,
      name: unlock.name,
      level: unlock.level,
    })),
  );

  readonly kinds = computed(() => [...new Set(this.all().map((row) => row.kind))]);

  readonly rows = computed(() => {
    const needle = this.filter().trim().toLowerCase();
    const kind = this.chosenKind();
    return this.all().filter(
      (row) =>
        (kind === '' || row.kind === kind) &&
        (needle === '' || row.name.toLowerCase().includes(needle)),
    );
  });
}
