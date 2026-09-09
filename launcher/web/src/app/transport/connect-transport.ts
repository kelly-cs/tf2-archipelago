import { InjectionToken } from '@angular/core';
import { Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

/**
 * One transport for the whole app. The launcher serves this page and answers
 * the RPCs on the same loopback origin, so there is no base URL to configure
 * and nothing to authenticate: the four passwords never cross a network.
 *
 * A token rather than a service so a test can hand a component a transport
 * that answers from a fixture without a server behind it.
 */
export const LAUNCHER_TRANSPORT = new InjectionToken<Transport>('launcher transport', {
  providedIn: 'root',
  factory: () => createConnectTransport({ baseUrl: '/' }),
});
