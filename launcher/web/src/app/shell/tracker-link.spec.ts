import { trackerLink } from './tracker-link';

describe('campaign tracker link', () => {
  it('loads the configured room and falls back to the tracker home', () => {
    expect(trackerLink('https://archipelago.gg/room/demo-room')).toBe(
      'https://m-this.github.io/tf2-archipelago/tracker/?room=demo-room',
    );
    expect(trackerLink('')).toBe('https://m-this.github.io/tf2-archipelago/tracker/');
    expect(trackerLink('not a room link')).toBe(
      'https://m-this.github.io/tf2-archipelago/tracker/',
    );
  });
});
