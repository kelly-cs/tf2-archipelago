import { mercenaries } from '@app/ui/tf2-art';

import { ClassView, TrackerSource } from './types';

const weaponSlots = ['Primary', 'Secondary', 'Melee'] as const;

export function loadoutView(
  source: TrackerSource,
  owned: ReadonlyMap<string, number>,
  sharedSlotCount: number,
  classWeaponSlots: boolean,
): { classes: readonly ClassView[]; slotCount: number; slotTotal: number } {
  const classes = mercenaries.map((name) => {
    const unlocked = owned.has(`Class: ${name}`);
    const order = source.classLoadouts.get(name) ?? weaponSlots;
    const copies = Math.min(2, owned.get(`Progressive Weapon Slot: ${name}`) ?? 0);
    const unlockedSlots = classWeaponSlots
      ? new Set(order.filter((_, index) => (index === 0 ? unlocked : index <= copies)))
      : new Set(weaponSlots.slice(0, sharedSlotCount));
    const slots = weaponSlots.map((slot) => ({ name: slot, unlocked: unlockedSlots.has(slot) }));
    return {
      name,
      unlocked,
      slots,
      slotCount: slots.filter((slot) => slot.unlocked).length,
    };
  });
  return {
    classes,
    slotCount: classWeaponSlots
      ? classes.reduce((total, classView) => total + classView.slotCount, 0)
      : sharedSlotCount,
    slotTotal: classWeaponSlots ? mercenaries.length * weaponSlots.length : weaponSlots.length,
  };
}
