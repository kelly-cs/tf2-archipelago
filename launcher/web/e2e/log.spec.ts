import { expect, test } from './fixtures';

test.describe('the log screen', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/log');
  });

  test('shows what the server said, newest last', async ({ page }) => {
    await expect(page.getByText('Executing dedicated server config: server.cfg')).toBeVisible();
    await expect(page.getByText('bridge connected to archipelago.gg:38281 as Scout')).toBeVisible();
  });

  test('colours a line by what it looks like, and still says the words', async ({ page }) => {
    // Colour is never the only thing that says a line went wrong: the words are
    // there in a screenshot with the colour stripped.
    const warned = page.locator('li.warn', { hasText: 'Warning: sv_pure is not set' });
    await expect(warned).toBeVisible();
  });

  test('filtering narrows it and says so when nothing is left', async ({ page }) => {
    await page.getByRole('searchbox', { name: 'Filter the log' }).fill('sourcemod');
    await expect(page.getByText('[SM] Loaded plugin tf2_archipelago.smx')).toHaveCount(0);

    await page.getByRole('searchbox', { name: 'Filter the log' }).fill('nothing says this');
    await expect(page.getByText('No line matches that filter.')).toBeVisible();
  });

  test('a console command is sent and comes back on the stream', async ({ page }) => {
    await page.getByLabel('Server command').fill('sm_ap_status');
    await page.getByRole('button', { name: 'Send' }).click();

    await expect(page.getByText('rcon: sm_ap_status')).toBeVisible();
    // The box empties, because the next command is a new one.
    await expect(page.getByLabel('Server command')).toHaveValue('');
  });

  test('a new line arrives without a reload', async ({ page }) => {
    await page.getByRole('link', { name: 'Session', exact: true }).click();
    await page.getByRole('button', { name: 'Start', exact: true }).click();
    await page.getByRole('link', { name: 'Log', exact: true }).click();

    await expect(page.getByText('the server is up')).toBeVisible();
  });
});
