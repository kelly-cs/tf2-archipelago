import { buildView } from './model';
import { sampleSource } from './sample';

describe('disabled mission tickets', () => {
  it('shows every mission open', () => {
    const source = sampleSource('https://archipelago.gg');
    const disabled = buildView(
      {
        ...source,
        demoSlotData: { ...source.demoSlotData, mission_ticket_importance: 'disabled' },
      },
      1,
    );
    expect(disabled.missions.every((mission) => !mission.locked)).toBe(true);
  });
});
