import { expect, test } from './fixtures';

test.describe('the session screen', () => {
  test('shows the connect line and what the run has counted', async ({ page }) => {
    await page.goto('/session');

    await expect(page.getByText('connect 127.0.0.1:27015; password ""')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Join with Steam' })).toBeEnabled();

    await expect(page.getByRole('heading', { name: 'The run' })).toBeVisible();
    await expect(page.getByText('PROOF-OF-A-SEED')).toBeVisible();
  });

  test('tells played apart from cleared elsewhere', async ({ page }) => {
    await page.goto('/session');
    const played = page.getByRole('row', { name: /Doe's Doom/ });
    await expect(played.getByText('played')).toBeVisible();

    // Another world's !collect sends every check it still holds, so the room can
    // hold a mission's check that this server never played. The two words are
    // the whole reason the column exists.
    const elsewhere = page.getByRole('row', { name: /Mean Machines/ });
    await expect(elsewhere.getByText('cleared elsewhere')).toBeVisible();
  });

  test('choosing the next mission needs a running server', async ({ page }) => {
    await page.goto('/session');
    const row = page.getByRole('row', { name: /Caliginous Caper/ });
    await expect(row.getByRole('button', { name: 'Play next' })).toBeDisabled();

    await page.getByRole('button', { name: 'Start', exact: true }).click();
    await expect(row.getByRole('button', { name: 'Play next' })).toBeEnabled();
    await row.getByRole('button', { name: 'Play next' }).click();

    await page.getByRole('link', { name: 'Log', exact: true }).click();
    await expect(page.getByText('next mission is mvm_coaltown_advanced')).toBeVisible();
  });
});
