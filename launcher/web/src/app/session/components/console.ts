import { CdkVirtualScrollViewport, ScrollingModule } from '@angular/cdk/scrolling';
import {
  ChangeDetectionStrategy,
  Component,
  DestroyRef,
  computed,
  inject,
  signal,
  viewChild,
} from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { Subject, exhaustMap, filter, map, switchMap, tap, timer } from 'rxjs';

import { LogSource, sourceOf } from '@app/session/log-level';
import { LauncherCommands } from '@app/server/launcher-commands';
import { LauncherStore } from '@app/server/launcher-store';
import { SettingsActions } from '@app/settings/settings-actions';
import { LogLine } from '@gen/tf2ap/launcher/v1/launcher_pb';
import { Button } from '@app/ui/button';
import { EmptyState } from '@app/ui/empty-state';
import { SearchBox } from '@app/ui/search-box';

// How close to the bottom still counts as being at it. A few pixels of slack,
// because a wheel rarely lands exactly on the end.
const atBottomPx = 24;

// A line is two rows tall at most in the common case; the viewport needs one
// number and this is the one that keeps the scrollbar honest.
const rowHeightPx = 18;

// What the player typed before, kept across a reload because a launcher is
// restarted more often than a browser is.
const historyKey = 'tf2ap.rcon.history.v1';
const historyMax = 100;

/** One line, ready to draw: the clock, who said it, and what it was. */
interface Row {
  readonly key: string;
  readonly at: string;
  readonly source: string;
  readonly kind: LogSource;
  readonly text: string;
}

/**
 * Everything the server and the launcher have said, under the missions on the
 * Play screen: the log is read while playing, beside the state it explains,
 * not on a page of its own.
 *
 * Twenty thousand lines, drawn a screenful at a time: the whole log is never in
 * the DOM, which is what makes a tab left open all evening cost nothing.
 */
@Component({
  selector: 'app-console',
  changeDetection: ChangeDetectionStrategy.OnPush,
  imports: [Button, EmptyState, ScrollingModule, SearchBox],
  templateUrl: './console.html',
  styleUrl: './console.scss',
})
export class Console {
  private readonly store = inject(LauncherStore);
  private readonly commands = inject(LauncherCommands);
  private readonly actions = inject(SettingsActions);
  private readonly viewport = viewChild(CdkVirtualScrollViewport);
  private readonly destroyRef = inject(DestroyRef);

  readonly rowHeight = rowHeightPx;

  readonly filterText = signal('');
  readonly follow = signal(true);
  readonly command = signal('');
  readonly said = signal('');

  /**
   * cleared is the clock of the last line on screen when the player asked for
   * a clean view; the launcher keeps its log, this only stops showing what was
   * there before. A time rather than a count: the log is a ring of twenty
   * thousand lines, and a count into a ring blanked the whole view for good
   * once the ring turned.
   */
  private readonly cleared = signal(-1n);

  private history: string[] = readHistory();
  private walking = -1;
  private draft = '';

  readonly rows = computed<Row[]>(() => {
    const needle = this.filterText().trim().toLowerCase();
    const logs = this.store.logs();
    const first = logs.findIndex((line) => stampOf(line) > this.cleared());
    return (first < 0 ? [] : logs.slice(first))
      .filter((line) => needle === '' || line.text.toLowerCase().includes(needle))
      .map((line, index) => ({
        key: `${index}:${line.text}`,
        at: clock(line.at?.seconds ?? 0n),
        source: line.source,
        kind: sourceOf(line.source),
        text: line.text,
      }));
  });

  readonly send = new Subject<void>();

  constructor() {
    this.send
      .pipe(
        filter(() => this.command().trim() !== ''),
        exhaustMap(() => {
          const typed = this.command().trim();
          this.remember(typed);
          return this.commands.sendRcon(typed);
        }),
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
        switchMap((rows) => timer(0).pipe(map(() => rows.length))),
        tap((rows) => {
          if (rows > 0) {
            this.viewport()?.scrollTo({ bottom: 0 });
          }
        }),
        takeUntilDestroyed(),
      )
      .subscribe();
  }

  /**
   * Scrolling away from the bottom stops following, and scrolling back starts
   * again. Measured from the bottom rather than by index: a log short enough to
   * fit reports the first index and no scrolling has happened, which used to
   * turn following off the moment the page opened.
   */
  trackRow(_: number, row: Row): string {
    return row.key;
  }

  onScroll(): void {
    const viewport = this.viewport();
    if (viewport === undefined) {
      return;
    }
    this.follow.set(viewport.measureScrollOffset('bottom') <= atBottomPx);
  }

  /** Up and down walk what was typed before, keeping the half-written line to
      come back to. */
  onKey(event: KeyboardEvent): void {
    if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') {
      return;
    }
    event.preventDefault();
    if (this.walking < 0) {
      this.draft = this.command();
    }
    const step = event.key === 'ArrowUp' ? 1 : -1;
    this.walking = Math.min(this.history.length - 1, Math.max(-1, this.walking + step));
    this.command.set(this.walking < 0 ? this.draft : this.history[this.walking]);
  }

  clearView(): void {
    const last = this.store.logs().at(-1);
    this.cleared.set(last === undefined ? -1n : stampOf(last));
    this.said.set(
      'This view is clear. The launcher still has every line, and the bundle carries them.',
    );
  }

  saveBundle(): void {
    this.actions
      .press('server.debug_bundle')
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => this.said.set(bundleHint));
  }

  private remember(command: string): void {
    this.walking = -1;
    this.history = [command, ...this.history.filter((one) => one !== command)].slice(0, historyMax);
    try {
      localStorage.setItem(historyKey, JSON.stringify(this.history));
    } catch {
      // A browser with no storage still gets a history for this session.
    }
  }
}

const bundleHint =
  'Saved the debug bundle to your downloads: the log above, your settings with the passwords ' +
  'stripped, and the player file. Send that one file when you report a problem.';

function readHistory(): string[] {
  try {
    // eslint-disable-next-line no-restricted-syntax -- JSON.parse answers with anything the browser stored, including whatever a previous version wrote.
    const kept: unknown = JSON.parse(localStorage.getItem(historyKey) ?? '[]');
    return Array.isArray(kept) ? kept.filter((one): one is string => typeof one === 'string') : [];
  } catch {
    return [];
  }
}

function stampOf(line: LogLine): bigint {
  return (line.at?.seconds ?? 0n) * 1_000_000_000n + BigInt(line.at?.nanos ?? 0);
}

function clock(seconds: bigint): string {
  return new Date(Number(seconds) * 1000).toLocaleTimeString([], { hour12: false });
}
