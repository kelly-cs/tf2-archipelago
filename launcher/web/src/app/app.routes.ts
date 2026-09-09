import { Routes } from '@angular/router';

import { ROUTE, SEGMENT } from '@app/routing/app-routes';

/**
 * Every screen is lazy. The shell is not: it draws before anything else and
 * carries the state line the player watches while the first screen loads.
 *
 * The settings tab is a child route on its own page rather than five: which
 * pages exist comes from form.Model at run time, so the tab is a slug in the
 * URL and the page reads it with input().
 */
export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('@app/shell/shell').then((m) => m.Shell),
    children: [
      { path: '', pathMatch: 'full', redirectTo: SEGMENT.session },
      {
        path: ROUTE.session,
        loadComponent: () => import('@app/session/pages/session-page').then((m) => m.SessionPage),
      },
      {
        path: ROUTE.unlocks,
        loadComponent: () => import('@app/unlocks/pages/unlocks-page').then((m) => m.UnlocksPage),
      },
      {
        path: ROUTE.bots,
        loadComponent: () => import('@app/bots/pages/bots-page').then((m) => m.BotsPage),
      },
      {
        path: ROUTE.log,
        loadComponent: () => import('@app/log/pages/log-page').then((m) => m.LogPage),
      },
      {
        path: ROUTE.settings,
        loadComponent: () =>
          import('@app/settings/pages/settings-page').then((m) => m.SettingsPage),
        children: [
          {
            path: ROUTE.settingsTab,
            loadComponent: () =>
              import('@app/settings/pages/settings-tab-page').then((m) => m.SettingsTabPage),
          },
        ],
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
