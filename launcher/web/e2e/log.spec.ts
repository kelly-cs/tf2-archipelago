import { expect, test } from './fixtures';

test.describe('the Console screen', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/session');
  });

  test('shows what the server said, and who said it', async ({ page }) => {
    await expect(page.getByText('Executing dedicated server config: server.cfg')).toBeVisible();

    // The source is a fact the launcher set, so it is what gets the colour. The
    // words are not read for severity: a line saying "error" may be srcds
    // reporting a missing sound.
    const fromSrcds = page.locator('.source.srcds').first();
    await expect(fromSrcds).toHaveText('[srcds]');
    await expect(page.locator('.source.launcher').first()).toHaveText('[launcher]');
  });

  test('follows the newest line from the moment it opens', async ({ page }) => {
    // A log short enough to fit reports the first index and no scrolling has
    // happened, which used to switch following off as the page opened.
    await expect(page.getByLabel('Follow the newest lines')).toBeChecked();
  });

  test('filtering narrows it and says so when nothing is left', async ({ page }) => {
    await page.getByRole('searchbox', { name: 'Filter the log' }).fill('sourcemod');
    await expect(page.getByText('[SM] Loaded plugin tf2_archipelago.smx')).toHaveCount(0);

    await page.getByRole('searchbox', { name: 'Filter the log' }).fill('nothing says this');
    await expect(page.getByText('No line matches that filter.')).toBeVisible();
  });

  test('a console command is sent, and the box empties for the next one', async ({ page }) => {
    await page.getByLabel('RCON command').fill('sm_ap_status');
    await page.getByRole('button', { name: 'Send' }).click();

    await expect(page.getByText('rcon: sm_ap_status')).toBeVisible();
    await expect(page.getByLabel('RCON command')).toHaveValue('');
  });

  test('the up arrow walks what was typed before', async ({ page }) => {
    const box = page.getByLabel('RCON command');
    for (const command of ['sm_ap_status', 'sm_ap_resync']) {
      await box.fill(command);
      await page.getByRole('button', { name: 'Send' }).click();
      await expect(box).toHaveValue('');
    }

    await box.press('ArrowUp');
    await expect(box).toHaveValue('sm_ap_resync');
    await box.press('ArrowUp');
    await expect(box).toHaveValue('sm_ap_status');
    await box.press('ArrowDown');
    await expect(box).toHaveValue('sm_ap_resync');
  });

  test("clearing the view keeps the launcher's own log", async ({ page }) => {
    await expect(page.getByText('Server is hibernating')).toBeVisible();
    await page.getByRole('button', { name: 'Clear this view' }).click();

    await expect(page.getByText('Server is hibernating')).toHaveCount(0);
    await expect(page.getByText(/The launcher still has every line/)).toBeVisible();
  });

  test('a new line arrives without a reload', async ({ page }) => {
    await page.getByRole('link', { name: 'Play', exact: true }).click();
    await page.getByRole('button', { name: 'Start server' }).click();

    await expect(page.getByText('the server is up')).toBeVisible();
  });
});
