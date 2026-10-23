import { test as base, expect } from '@playwright/test';

/**
 * One fake launcher serves every test, so each one starts by putting it back to
 * how it began. That is also why the tests do not run in parallel: they share a
 * run the way two browser tabs on one launcher would.
 */
export const test = base.extend({
  page: async ({ page, baseURL }, use) => {
    await page.request.post(`${baseURL}/fake/reset`);
    await use(page);
  },
});

export { expect };
