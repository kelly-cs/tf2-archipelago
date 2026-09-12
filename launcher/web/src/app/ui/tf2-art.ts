/** TF2 artwork shared by the launcher and the public campaign tracker. */
export const mercenaries = [
  'Scout',
  'Soldier',
  'Pyro',
  'Demoman',
  'Heavy',
  'Engineer',
  'Medic',
  'Sniper',
  'Spy',
] as const;

export type Mercenary = (typeof mercenaries)[number];

export const mercenaryIcons: Readonly<Record<Mercenary, string>> = {
  Scout: 'https://wiki.teamfortress.com/w/images/b/b4/Class_scoutred.png',
  Soldier: 'https://wiki.teamfortress.com/w/images/2/2f/Class_soldierred.png',
  Pyro: 'https://wiki.teamfortress.com/w/images/1/17/Class_pyrored.png',
  Demoman: 'https://wiki.teamfortress.com/w/images/e/e9/Class_demored.png',
  Heavy: 'https://wiki.teamfortress.com/w/images/c/c1/Class_heavyred.png',
  Engineer: 'https://wiki.teamfortress.com/w/images/3/36/Class_engired.png',
  Medic: 'https://wiki.teamfortress.com/w/images/b/bb/Class_medicred.png',
  Sniper: 'https://wiki.teamfortress.com/w/images/8/81/Class_sniperred.png',
  Spy: 'https://wiki.teamfortress.com/w/images/6/64/Class_spyred.png',
};

export const grapplingHookIcon =
  'https://wiki.teamfortress.com/w/images/thumb/e/ea/Item_icon_Grappling_Hook.png/128px-Item_icon_Grappling_Hook.png';
