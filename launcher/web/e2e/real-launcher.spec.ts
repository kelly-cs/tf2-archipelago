import type { Page } from '@playwright/test';

import { expect, test } from '@playwright/test';

/**
 * openSettings makes sure the screen is open without insisting it was closed.
 *
 * A real launcher keeps its draft between these tests the way it keeps it
 * between two browser tabs, so the button is there only the first time.
 */
async function openSettings(page: Page): Promise<void> {
  await page.goto('/settings');

  // The page opens the draft itself once the first frame says it is closed.
  await expect(page.getByRole('navigation', { name: 'Settings pages' })).toBeVisible();
}

const at = process.env['TF2AP_REAL'] ?? '';

test.describe('the real launcher', () => {
  test.skip(at === '', 'set TF2AP_REAL to the address of a running launcher');
  test.use({ baseURL: at });

  test('serves the interface it embeds', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/session$/);
    await expect(page.getByRole('img', { name: 'Mann vs Archipelago' })).toBeVisible();

    // The state line comes off the real WebSocket, decoded from real proto.
    await expect(page.getByText(/^Stopped/)).toBeVisible();
    await expect(page.getByRole('button', { name: 'Start server' })).toBeEnabled();
  });

  test('draws the real form model, every page of it', async ({ page }) => {
    await openSettings(page);

    const pages = page.getByRole('navigation', { name: 'Settings pages' });
    for (const name of ['Player options', 'Rewards', 'Missions', 'Game server', 'Bots']) {
      await expect(pages.getByRole('link', { name, exact: true })).toBeVisible();
    }

    await pages.getByRole('link', { name: 'Missions', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Mission pool' })).toBeVisible();
    expect(await page.locator('td.name').count()).toBeGreaterThan(20);
  });

  test('reads a folder for a Browse row', async ({ page }) => {
    await openSettings(page);
    await page
      .getByRole('navigation', { name: 'Settings pages' })
      .getByRole('link', { name: 'Player options', exact: true })
      .click();
    await page.getByRole('button', { name: 'Browse' }).first().click();

    const picker = page.getByRole('dialog', { name: 'Choose a folder' });
    await expect(picker).toBeVisible();
    await expect(picker.getByRole('button', { name: 'Up' })).toBeEnabled();
  });
});
