/**
 * What a line is, read off the line. srcds and the bridge write plain text with
 * no level in it, so this is the launcher reading the words rather than a field
 * anybody set: a guess that is right often enough to colour by, and never load
 * bearing.
 */
export type LogLevel = 'error' | 'warn' | 'note' | 'plain';

const errorWords = /\b(error|failed|failure|cannot|refused|fatal|crash|segfault|panic)\b/i;
const warnWords = /\b(warn|warning|missing|retry|retrying|timeout|timed out|deprecated)\b/i;

export function levelOf(text: string): LogLevel {
  if (errorWords.test(text)) {
    return 'error';
  }
  if (warnWords.test(text)) {
    return 'warn';
  }
  return 'plain';
}

/** The launcher's own lines are worth telling apart from the server's. */
export function levelOfLine(source: string, text: string): LogLevel {
  const level = levelOf(text);
  return level === 'plain' && source === 'launcher' ? 'note' : level;
}
