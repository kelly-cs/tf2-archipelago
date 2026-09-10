import { Provider } from '@angular/core';

import { SECTION_RENDERERS, SectionRenderer } from '@app/settings/section-renderers';

const sections: SectionRenderer[] = [
  {
    key: 'bots/team',
    load: () => import('@app/bots/components/bot-lineup').then((m) => m.BotLineup),
  },
  {
    key: 'bots/classes',
    load: () => import('@app/bots/components/class-table').then((m) => m.ClassTable),
  },
  {
    key: 'bots/loadouts',
    load: () => import('@app/bots/components/loadout-editor').then((m) => m.LoadoutEditor),
  },
];

/** provideBotSections is the Bots domain drawing its own settings sections. */
export function provideBotSections(): Provider {
  return { provide: SECTION_RENDERERS, useValue: sections };
}
