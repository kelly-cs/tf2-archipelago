import { expect, test } from './fixtures';

// The shell is what every screen sits in, so a break here breaks all of them.
test.describe('the shell', () => {
  test('opens on the session screen with the launcher state read off the stream', async ({
    page,
  }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/session$/);

    await expect(page.getByRole('img', { name: 'Mann vs Archipelago' })).toBeVisible();
    await expect(page.getByText('Mann vs Archipelago (fake)')).toBeVisible();

    // The state line comes from the first WebSocket frame. Seeing the room in
    // it proves the frame arrived, decoded, and reached a signal.
    await expect(page.getByText(/stopped, room archipelago\.gg:38281/)).toBeVisible();
  });

  test('every section is one click away and survives a reload', async ({ page }) => {
    await page.goto('/');
    for (const [name, path] of [
      ['Unlocks', '/unlocks'],
      ['Bots', '/bots'],
      ['Log', '/log'],
      ['Settings', '/settings'],
      ['Session', '/session'],
    ] as const) {
      await page.getByRole('link', { name, exact: true }).click();
      await expect(page).toHaveURL(new RegExp(`${path}$`));
    }

    // A refresh on a client-side route is the thing a file server gets wrong.
    await page.goto('/unlocks');
    await expect(page.getByRole('heading', { name: 'Everything unlocked' })).toBeVisible();
  });

  test('Start moves the state, and the change arrives without a reload', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Start', exact: true }).click();

    // No reload anywhere: the button sent an RPC and the stream pushed the
    // answer back into the state line.
    await expect(page.getByRole('button', { name: 'Stop', exact: true })).toBeVisible();
    await expect(page.getByText(/^running,/)).toBeVisible();

    await page.getByRole('button', { name: 'Stop', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Start', exact: true })).toBeVisible();
  });
});
