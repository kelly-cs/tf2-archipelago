import { ScrollingModule, CdkVirtualScrollViewport } from '@angular/cdk/scrolling';
import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  linkedSignal,
  signal,
  viewChild,
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap, tap } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { levelOfLine } from '@app/log/log-level';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { SearchBox } from '@app/ui/search-box';

/** One line, ready to draw: the clock, where it came from, and what it is. */
interface Row {
  readonly key: string;
  readonly at: string;
  readonly source: string;
  readonly text: string;
  readonly level: string;
}

/**
 * Twenty thousand lines, drawn a screenful at a time. The whole log is never in
 * the DOM: a virtual viewport keeps the rows it can see, which is what makes a
 * tab left open all evening cost nothing.
 */
@Component({
  selector: 'app-log-page',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, EmptyState, ScrollingModule, SearchBox],
  templateUrl: './log-page.html',
  styleUrl: './log-page.scss',
})
export class LogPage {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);
  private readonly viewport = viewChild(CdkVirtualScrollViewport);

  readonly filter = signal('');
  readonly follow = signal(true);
  readonly command = signal('');
  readonly copied = signal(false);

  readonly rows = computed<Row[]>(() => {
    const needle = this.filter().trim().toLowerCase();
    return this.store
      .logs()
      .filter((line) => needle === '' || line.text.toLowerCase().includes(needle))
      .map((line, index) => ({
        key: `${index}:${line.text}`,
        at: clock(line.at?.seconds ?? 0n),
        source: line.source,
        text: line.text,
        level: levelOfLine(line.source, line.text),
      }));
  });

  /** The newest row, tracked so following can jump to it as it changes. */
  readonly newest = linkedSignal(() => {
    const last = this.rows().length - 1;
    if (this.follow() && last >= 0) {
      this.viewport()?.scrollToIndex(last, 'auto');
    }
    return last;
  });

  readonly send = new Subject<void>();
  readonly copy = new Subject<void>();

  constructor() {
    this.send
      .pipe(
        exhaustMap(() => this.commands.sendRcon(this.command())),
        tap(() => this.command.set('')),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  trackRow(_: number, row: Row): string {
    return row.key;
  }

  onScroll(index: number): void {
    if (index + 1 < this.rows().length - 1) {
      this.follow.set(false);
    }
  }

  copyAll(): void {
    void navigator.clipboard.writeText(
      this.rows()
        .map((row) => row.text)
        .join('\n'),
    );
    this.copied.set(true);
  }
}

function clock(seconds: bigint): string {
  return new Date(Number(seconds) * 1000).toLocaleTimeString([], { hour12: false });
}
