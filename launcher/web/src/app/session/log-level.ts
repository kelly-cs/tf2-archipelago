/**
 * What colour a log line's source tag gets.
 *
 * The source is a fact the launcher set, not a guess: srcds said it, or rcon
 * did, or the launcher, or the installer. Colouring by that is worth doing
 * because it answers "who is telling me this", which is the first question
 * anybody reading a wall of server output has.
 *
 * Nothing here guesses severity from the words. A line saying "error" may be
 * srcds reporting a missing sound; a line saying nothing of the sort may be the
 * install giving up. Painting one red on the strength of a substring tells the
 * player something the launcher does not know.
 */
export type LogSource = 'srcds' | 'rcon' | 'launcher' | 'install' | 'other';

const known: Record<string, LogSource> = {
  srcds: 'srcds',
  rcon: 'rcon',
  launcher: 'launcher',
  install: 'install',
};

export function sourceOf(source: string): LogSource {
  return known[source] ?? 'other';
}
