import { describe, expect, it } from 'vitest';

import { levelOf, levelOfLine } from '@app/log/log-level';

describe('reading a level off a log line', () => {
  it('finds the words that mean something broke', () => {
    expect(levelOf('L 09/09/2026 - 21:11:02: [SM] Plugin failed to load')).toBe('error');
    expect(levelOf('cannot open tf/cfg/server.cfg')).toBe('error');
    expect(levelOf('Segfault in engine.so')).toBe('error');
  });

  it('finds the words that mean something may break', () => {
    expect(levelOf('Warning: sv_pure is not set')).toBe('warn');
    expect(levelOf('retrying the download')).toBe('warn');
  });

  it('leaves an ordinary line alone', () => {
    expect(levelOf('Wave 3 of 6 begins')).toBe('plain');
  });

  // The launcher's own lines are worth telling from the server's: they are the
  // ones that say what the launcher is about to do to the player's machine.
  it('marks the launcher speaking, but never over a real error', () => {
    expect(levelOfLine('launcher', 'installing SourceMod')).toBe('note');
    expect(levelOfLine('launcher', 'cannot reach archipelago.gg')).toBe('error');
    expect(levelOfLine('srcds', 'Wave 3 of 6 begins')).toBe('plain');
  });
});
