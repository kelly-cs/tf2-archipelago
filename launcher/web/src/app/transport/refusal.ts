import { ConnectError } from '@connectrpc/connect';
import { Observable, OperatorFunction, catchError } from 'rxjs';

/**
 * orRefusal turns a rejected call into what the launcher said, on the same
 * stream. A refusal is an answer, not a failure: the calls behind a screen
 * share one subscription, and an error that reached it ended it, after which
 * nothing pressed or typed got through and nobody was told.
 */
export function orRefusal<T>(handle: (refusal: string) => Observable<T>): OperatorFunction<T, T> {
  // eslint-disable-next-line no-restricted-syntax -- A caught error is whatever was thrown.
  return catchError((error: unknown) => handle(refusalOf(error)));
}

// eslint-disable-next-line no-restricted-syntax -- A caught error is whatever was thrown.
function refusalOf(error: unknown): string {
  if (error instanceof ConnectError) {
    return error.rawMessage;
  }
  return error instanceof Error ? error.message : 'The launcher did not answer.';
}
