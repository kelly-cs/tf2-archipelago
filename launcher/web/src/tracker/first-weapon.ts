import { Mercenary } from '@app/ui/tf2-art';

/** Stock weapon players recognize in each class's first AP unlock slot. */
export const firstWeapon: Readonly<Record<Mercenary, string>> = {
  Scout: 'Scattergun',
  Soldier: 'Rocket Launcher',
  Pyro: 'Flame Thrower',
  Demoman: 'Grenade Launcher',
  Heavy: 'Minigun',
  Engineer: 'Wrench',
  Medic: 'Medi Gun',
  Sniper: 'Sniper Rifle',
  Spy: 'Knife',
};
