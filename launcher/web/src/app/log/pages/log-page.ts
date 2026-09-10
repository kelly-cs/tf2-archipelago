import { ScrollingModule, CdkVirtualScrollViewport } from '@angular/cdk/scrolling';
import {
  ChangeDetectionStrategy,
  Component,
  computed,
  inject,
  signal,
  viewChild,
} from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap, filter, map, switchMap, tap, timer } from 'rxjs';

import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { levelOfLine } from '@app/log/log-level';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { SearchBox } from '@app/ui/search-box';

// How close to the bottom still counts as being at it. A few pixels of slack,
// because a wheel rarely lands exactly on the end.
const atBottomPx = 24;

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

    // Following is a side effect on the viewport, so it runs through the stream
    // of row counts rather than inside a computed. It waits a frame because the
    // viewport has to measure the new rows before it can scroll to them.
    toObservable(this.rows)
      .pipe(
        filter(() => this.follow()),
        switchMap((rows) => timer(0).pipe(map(() => rows.length - 1))),
        tap((last) => {
          if (last >= 0) {
            this.viewport()?.scrollTo({ bottom: 0 });
          }
        }),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  trackRow(_: number, row: Row): string {
    return row.key;
  }

  /**
   * Scrolling away from the bottom stops following, and scrolling back starts
   * again. Measured from the bottom rather than by index: a log short enough to
   * fit reports the first index and no scrolling has happened, which used to
   * turn following off the moment the page opened.
   */
  onScroll(): void {
    const viewport = this.viewport();
    if (viewport === undefined) {
      return;
    }
    this.follow.set(viewport.measureScrollOffset('bottom') <= atBottomPx);
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
