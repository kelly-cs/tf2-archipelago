// Single source of truth for every route in the app.
//
// SEGMENT — individual URL segments, typed and reused everywhere.
// PARAM   — route parameter names; each must match the `input()` name on the
//           component that reads it (withComponentInputBinding).
// ROUTE   — path patterns for the Routes config, relative to their parent.
// appLink — absolute command arrays for [routerLink] / router.navigate.
// appUrl  — absolute URL strings for router.parseUrl / navigateByUrl.
//
// Links are absolute by design: they do not depend on where the component sits
// in the route tree, so no `../..` and every target is checked by the compiler.
// Extend these tables as pages arrive — never write a raw path in a component,
// a guard or a template.

export const SEGMENT = {
  session: 'session',
  unlocks: 'unlocks',
  bots: 'bots',
  log: 'log',
  settings: 'settings',
} as const;

// The settings pages come from form.Model at run time, so the tab is a slug of
// the tab title rather than a segment declared here.
export const PARAM = {
  settingsTab: 'settingsTab',
} as const;

export const ROUTE = {
  session: SEGMENT.session,
  unlocks: SEGMENT.unlocks,
  bots: SEGMENT.bots,
  log: SEGMENT.log,
  settings: SEGMENT.settings,
  settingsTab: `:${PARAM.settingsTab}`,
} as const;

export const appLink = {
  root: (): string[] => ['/'],
  session: (): string[] => ['/', SEGMENT.session],
  unlocks: (): string[] => ['/', SEGMENT.unlocks],
  bots: (): string[] => ['/', SEGMENT.bots],
  log: (): string[] => ['/', SEGMENT.log],
  settings: (): string[] => ['/', SEGMENT.settings],
  settingsTab: (tab: string): string[] => ['/', SEGMENT.settings, tab],
} as const;

export const appUrl = {
  root: (): string => '/',
  session: (): string => `/${SEGMENT.session}`,
  unlocks: (): string => `/${SEGMENT.unlocks}`,
  bots: (): string => `/${SEGMENT.bots}`,
  log: (): string => `/${SEGMENT.log}`,
  settings: (): string => `/${SEGMENT.settings}`,
  settingsTab: (tab: string): string => `/${SEGMENT.settings}/${tab}`,
} as const;
