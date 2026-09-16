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
  it('opens exactly the slots a named-slot seed handed out, in no order', () => {
    const view = buildView(
      {
        ...source,
        demoSlotData: { ...source.demoSlotData, class_weapon_slots_any_order: true },
        itemNames: new Map([
          ...source.itemNames,
          [9_900_001, 'Scout Melee Slot'],
          [9_900_002, 'Class: Scout'],
        ]),
        received: [{ player: 1, items: [[9_900_002], [9_900_001]] }],
      },
      1,
    );

    // The Scout's melee, found without his secondary: progressive could not
    // have produced this state at all.
    expect(
      view.classes.find((one) => one.name === 'Scout')?.slots.map((s) => `${s.name}:${s.unlocked}`),
    ).toEqual(['Primary:true', 'Secondary:false', 'Melee:true']);
  });
});
