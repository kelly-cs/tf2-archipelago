import { Injectable } from '@angular/core';
import { fromBinary } from '@bufbuild/protobuf';
import { Observable, catchError, map, of, retry, timer } from 'rxjs';
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
      catchError(() => of<Frame>({ lost: true })),
    );
  }
}

function streamURL(): string {
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${scheme}//${window.location.host}/ws`;
}
