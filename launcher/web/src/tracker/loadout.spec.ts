import { buildView } from './model';
import { sampleSource } from './sample';

describe('tracker loadouts', () => {
  const source = sampleSource('https://archipelago.gg');
  const view = buildView(source, 1);

  it('shows each class slot in the order that class actually earns it', () => {
    const slots = (name: string) =>
      view.classes
        .find((classView) => classView.name === name)
        ?.slots.map((slot) => `${slot.name}:${slot.unlocked}`);

    expect(slots('Scout')).toEqual(['Primary:true', 'Secondary:true', 'Melee:true']);
    expect(slots('Soldier')).toEqual(['Primary:true', 'Secondary:true', 'Melee:false']);
    expect(slots('Engineer')).toEqual(['Primary:true', 'Secondary:false', 'Melee:true']);
    expect(slots('Medic')).toEqual(['Primary:true', 'Secondary:true', 'Melee:true']);
    expect(slots('Spy')).toEqual(['Primary:false', 'Secondary:true', 'Melee:false']);
  });

  it('keeps the shared three-slot view for rooms using the original option', () => {
    const shared = buildView(
      {
        ...source,
        demoSlotData: { ...source.demoSlotData, class_weapon_slots: false },
      },
      1,
    );

    expect(shared.classWeaponSlots).toBe(false);
    expect(shared.slotCount).toBe(1);
    expect(shared.slotTotal).toBe(3);
    expect(shared.classes.find((entry) => entry.name === 'Medic')?.slots).toEqual([
      { name: 'Primary', unlocked: true },
      { name: 'Secondary', unlocked: false },
      { name: 'Melee', unlocked: false },
    ]);
  });
});
