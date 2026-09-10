import { expect, test } from './fixtures';

/**
 * The pictures in the README and the book, taken against the fake launcher so
 * the same run draws the same image every time: no machine's fonts, no
 * player's home directory, no state left over from an evening of playing.
 *
 * Run with `make web-captures`. Skipped otherwise: a screenshot that differs by
 * a pixel is not a test failure, and a suite that rewrites committed files on
 * every run is a suite nobody can trust.
 */
const taking = process.env['TF2AP_CAPTURE'] === '1';

test.describe('the pictures', () => {
  test.skip(!taking, 'set TF2AP_CAPTURE=1 to redraw them');
  test.use({ viewport: { width: 1280, height: 860 } });

  test('the session screen', async ({ page }) => {
    await page.goto('/session');
    await page.getByRole('button', { name: 'Start', exact: true }).click();
    await expect(page.getByText(/^running,/)).toBeVisible();
    await page.screenshot({ path: '../../docs/images/launcher-session.png' });
  });

  test('the log', async ({ page }) => {
    await page.goto('/log');
    await expect(page.getByText('bridge connected to archipelago.gg:38281 as Scout')).toBeVisible();
    await page.screenshot({ path: '../../docs/images/launcher-log.png' });
  });

  test('the settings, on the mission pool', async ({ page }) => {
    await page.goto('/settings');
    await page.getByRole('button', { name: 'Open the settings' }).click();
    await page
      .getByRole('navigation', { name: 'Settings pages' })
      .getByRole('link', { name: 'Missions', exact: true })
      .click();
    await expect(page.getByRole('heading', { name: 'Mission pool' })).toBeVisible();
    await page.screenshot({ path: '../../docs/images/launcher-settings.png' });
  });
});
