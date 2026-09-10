import { Injectable } from '@angular/core';
import { fromBinary } from '@bufbuild/protobuf';
import {
  Observable,
  catchError,
  distinctUntilChanged,
  fromEvent,
  map,
  of,
  retry,
  startWith,
  switchMap,
  timer,
} from 'rxjs';
import { webSocket } from 'rxjs/webSocket';

import { StreamMessage, StreamMessageSchema } from '@gen/tf2ap/launcher/v1/stream_pb';

// The launcher pushes and this app only listens, so the live state is a
// WebSocket rather than a request anybody holds open. The first frame is the
// whole snapshot; after it a log line arrives on its own and the whole state
// arrives whenever anything else moved.

// A launcher on this machine either answers or is gone. These bound the wait so
// a tab left open on a launcher that quit stops asking instead of retrying for
// the rest of the evening.
const retriesMax = 8;
const backoffMinMs = 250;
const backoffMaxMs = 5_000;

/** Lost is the stream having given up: the launcher is not answering. */
export interface Lost {
  readonly lost: true;
}

export type Frame = StreamMessage | Lost;

export function isLost(frame: Frame): frame is Lost {
  return 'lost' in frame;
}

@Injectable({ providedIn: 'root' })
export class LauncherStream {
  /**
   * frames is the live stream, reconnecting with capped backoff and giving up
   * after retriesMax tries. It never errors: giving up arrives as a Lost frame,
   * because the shell has to say so rather than fail.
   */
  frames(): Observable<Frame> {
    return visible().pipe(
      switchMap((showing) => (showing ? this.socket() : of<Frame>())),
      catchError(() => of<Frame>({ lost: true })),
    );
  }

  /**
   * The socket itself. Opening it is what asks the launcher for the state, so
   * every time the tab comes back the first frame is a whole snapshot and
   * nothing has to be caught up on.
   */
  private socket(): Observable<Frame> {
    return webSocket<StreamMessage>({
      url: streamURL(),
      binaryType: 'arraybuffer',
      deserializer: (event: MessageEvent<ArrayBuffer>) =>
        fromBinary(StreamMessageSchema, new Uint8Array(event.data)),
    }).pipe(
      retry({
        count: retriesMax,
        delay: (_, attempt) => timer(Math.min(backoffMaxMs, backoffMinMs * 2 ** (attempt - 1))),
        resetOnSuccess: true,
      }),
      map((message): Frame => message),
    );
  }
}

/**
 * Whether the tab is in front of the player.
 *
 * A launcher tab is left open for a whole evening, usually behind the game. A
 * hidden tab is told nothing: the socket is closed, the launcher drops the
 * listener, and no frame is decoded for a screen nobody is looking at. Coming
 * back opens it again, and the first frame is the whole state, so there is
 * nothing to replay and nothing was missed.
 */
function visible(): Observable<boolean> {
  return fromEvent(document, 'visibilitychange').pipe(
    map(() => document.visibilityState === 'visible'),
    startWith(document.visibilityState === 'visible'),
    distinctUntilChanged(),
  );
}

function streamURL(): string {
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${scheme}//${window.location.host}/ws`;
}
