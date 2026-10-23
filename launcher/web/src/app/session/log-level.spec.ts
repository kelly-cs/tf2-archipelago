import { describe, expect, it } from 'vitest';

import { sourceOf } from '@app/session/log-level';

/**
 * The source is a fact the launcher set, so this only checks the mapping and
 * the fallback. Nothing guesses severity from the words any more: a line saying
 * "error" may be srcds reporting a missing sound, and painting it red on the
 * strength of a substring told the player something the launcher did not know.
 */
describe('who said a log line', () => {
  it('names the four the launcher writes', () => {
    expect(sourceOf('srcds')).toBe('srcds');
    expect(sourceOf('rcon')).toBe('rcon');
    expect(sourceOf('launcher')).toBe('launcher');
    expect(sourceOf('install')).toBe('install');
  });

  it('falls back rather than inventing a colour for a source it does not know', () => {
    expect(sourceOf('bridge')).toBe('other');
    expect(sourceOf('')).toBe('other');
  });
});
