import { expect, test, type Page } from '@playwright/test';

/**
 * The pictures in the README and the book, taken against the fake launcher so
 * the same run draws the same image every time: no machine's fonts, no
 * player's home directory, no state left over from an evening of playing.
 *
 * Run with `make web-captures`. Skipped otherwise: a screenshot that differs by
 * a pixel is not a test failure, and a suite that rewrites committed files on
 * every run is a suite nobody can trust.
 *
 * The tracker pictures come from `dist/tracker` served on TF2AP_TRACKER_URL,
 * which `make web-captures` starts. Without it those tests skip.
 */
const taking = process.env['TF2AP_CAPTURE'] === '1';
const trackerURL = process.env['TF2AP_TRACKER_URL'];
const images = '../../docs/images/';

/**
 * A page-long picture without the sticky footer landing in the middle of it:
 * grow the viewport to the document, then take a plain screenshot.
 */
async function whole(page: Page, name: string): Promise<void> {
  await page.waitForTimeout(600);
  const height = await page.evaluate(() => document.documentElement.scrollHeight);
  await page.setViewportSize({ width: 1280, height: Math.min(Math.max(height, 860), 4000) });
  await page.waitForTimeout(300);
  await page.screenshot({ path: `${images}${name}.png` });
  await page.setViewportSize({ width: 1280, height: 860 });
}

async function settingsPage(page: Page, slug: string): Promise<void> {
  await page.goto(`/settings/${slug}`);
  await expect(page.getByRole('navigation', { name: 'Settings pages' })).toBeVisible();
}

test.describe('the pictures', () => {
  test.skip(!taking, 'set TF2AP_CAPTURE=1 to redraw them');
  test.use({ viewport: { width: 1280, height: 860 } });

  test('the play screen, stopped', async ({ page }) => {
    await page.goto('/session');
    await expect(page.getByRole('heading', { name: 'Join the server' })).toBeVisible();
    await whole(page, 'launcher-play-stopped');
  });

  test('the play screen, running', async ({ page }) => {
    await page.goto('/session');
    await page.getByRole('button', { name: 'Start server' }).click();
    await expect(page.getByText(/^Running/)).toBeVisible();
    await expect(page.getByText('bridge connected to archipelago.gg:38281 as Scout')).toBeVisible();
    await whole(page, 'launcher-session');
    await page.screenshot({ path: `${images}launcher-log.png` });
  });

  test('the unlocks screen', async ({ page }) => {
    await page.goto('/unlocks');
    await expect(page.getByText(/unlocks so far/)).toBeVisible();
    await whole(page, 'launcher-unlocks');
  });

  test('the bots screen', async ({ page }) => {
    await page.goto('/bots');
    const open = page.getByRole('button', { name: 'Open the lineup' });
    if (await open.isVisible().catch(() => false)) {
      await open.click();
    }
    await expect(page.getByRole('heading', { name: 'Bot switcher' })).toBeVisible();
    await whole(page, 'launcher-bots');
  });

  for (const slug of [
    'player-options',
    'rewards',
    'balancing',
    'missions',
    'archipelago-room',
    'game-server',
    'networking',
  ]) {
    test(`the settings, ${slug}`, async ({ page }) => {
      await settingsPage(page, slug);
      await whole(page, `launcher-settings-${slug}`);
    });
  }

  test('the settings, on the mission pool', async ({ page }) => {
    await settingsPage(page, 'missions');
    await expect(page.getByRole('heading', { name: 'Mission pool' })).toBeVisible();
    await page.screenshot({ path: `${images}launcher-settings.png` });
  });

  for (const section of ['team', 'classes', 'names', 'looks', 'loadouts']) {
    test(`the settings, bots, ${section}`, async ({ page }) => {
      await settingsPage(page, `bots/${section}`);
      await whole(page, `launcher-settings-bots-${section}`);
    });
  }
});

test.describe('the tracker pictures', () => {
  test.skip(!taking || !trackerURL, 'set TF2AP_CAPTURE=1 and TF2AP_TRACKER_URL to redraw them');
  test.use({ viewport: { width: 1280, height: 860 } });

  test('the front page', async ({ page }) => {
    await page.goto(trackerURL!);
    await expect(page.getByRole('button', { name: 'View sample run' })).toBeVisible();
    await page.screenshot({ path: `${images}tracker-home.png` });
  });

  test('a run', async ({ page }) => {
    await page.goto(trackerURL!);
    await page.getByRole('button', { name: 'View sample run' }).click();
    await expect(page.getByRole('heading', { name: 'Mission board' })).toBeVisible();
    await whole(page, 'tracker-run');
    await page.locator('.class-card').first().click();
    await expect(page.locator('dialog.buff-dialog')).toBeVisible();
    await page.waitForTimeout(500);
    await page.screenshot({ path: `${images}tracker-class.png` });
  });
});
