import { defaultHost, parseSource } from '@tracker/source-address';

const trackerHome = 'https://m-this.github.io/tf2-archipelago/tracker/';

/** A room page contains the ID that the game's host:port address does not. */
export function trackerLink(roomPage: string): string {
  const url = new URL(trackerHome);
  if (roomPage.trim() === '') return url.href;
  try {
    const source = parseSource(roomPage);
    if (source.kind === 'unknown') return url.href;
    url.searchParams.set(source.kind, source.id);
    if (source.host !== defaultHost) url.searchParams.set('host', source.host);
  } catch {
    // An incomplete optional setting still leaves the tracker home accessible.
  }
  return url.href;
}
