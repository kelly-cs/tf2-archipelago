import { Transport } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';

// One transport for the whole app. The launcher serves the app and answers the
// RPCs on the same loopback origin, so there is no base URL to configure and
// nothing to authenticate: the four passwords never cross a network.
export function launcherTransport(): Transport {
  return createConnectTransport({ baseUrl: '/' });
}
