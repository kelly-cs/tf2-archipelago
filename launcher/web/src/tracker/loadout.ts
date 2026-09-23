import { mercenaries } from '@app/ui/tf2-art';

import { ClassView, TrackerSource } from './types';

const weaponSlots = ['Primary', 'Secondary', 'Melee'] as const;

export function loadoutView(
  source: TrackerSource,
  owned: ReadonlyMap<string, number>,
  sharedSlotCount: number,
  classWeaponSlots: boolean,
  anyOrder: boolean,
): { classes: readonly ClassView[]; slotCount: number; slotTotal: number } {
  const classes = mercenaries.map((name) => {
    const unlocked = owned.has(`Class: ${name}`);
    const order = source.classLoadouts.get(name) ?? weaponSlots;
    // The class's first slot comes free with the class, so the rest is what its
    // own item can earn: the order the export gives, less that first one.
    const earned = order.length - 1;
    const copies = Math.min(earned, owned.get(`Progressive Weapon Slot: ${name}`) ?? 0);
    // A seed that names its slots opens exactly the ones it handed out, in no
    // particular order; a progressive one opens them in the class's order.
    const byName = order.filter(
      (slot, index) => (index === 0 && unlocked) || owned.has(`${name} ${slot} Slot`),
    );
    const unlockedSlots = classWeaponSlots
      ? new Set(
          anyOrder
            ? byName
            : order.filter((_, index) => (index === 0 ? unlocked : index <= copies)),
        )
      : new Set(order.slice(0, sharedSlotCount));
    const slots = order.map((slot) => ({ name: slot, unlocked: unlockedSlots.has(slot) }));
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
