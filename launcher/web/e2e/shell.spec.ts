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
    await expect(page.getByText(/^Stopped · room archipelago\.gg:38281/)).toBeVisible();
  });

  test('every section is one click away and survives a reload', async ({ page }) => {
    await page.goto('/');
    for (const [name, path] of [
      ['Unlocks', '/unlocks'],
      ['Bots', '/bots'],
      // Settings lands on its first page: a settings screen with nothing on
      // it is a screen the player has to guess at.
      ['Settings', '/settings/player-options'],
      ['Play', '/session'],
    ] as const) {
      await page
        .getByRole('navigation', { name: 'Sections' })
        .getByRole('link', { name, exact: true })
        .click();
      await expect(page).toHaveURL(new RegExp(`${path}$`));
    }

    // A refresh on a client-side route is the thing a file server gets wrong.
    await page.goto('/unlocks');
    await expect(page.getByRole('heading', { name: 'Everything unlocked' })).toBeVisible();
  });

  test('Start moves the state, and the change arrives without a reload', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('button', { name: 'Start server' })).toBeVisible();
    // Join is there before the server is, disabled: the one button the player
    // came for is where they will look for it.
    await expect(page.getByRole('button', { name: 'Join', exact: true })).toBeDisabled();
    await expect(page.getByRole('link', { name: 'Star on GitHub' })).toHaveAttribute(
      'href',
      'https://github.com/m-this/tf2-archipelago',
    );

    await page.getByRole('button', { name: 'Start server' }).click();

    // No reload anywhere: the button sent an RPC and the stream pushed the
    // answer back into the state line.
    await expect(page.getByRole('button', { name: 'Stop server' })).toBeVisible();
    await expect(page.getByText(/^Running/)).toBeVisible();

    await page.getByRole('button', { name: 'Stop server' }).click();
    await expect(page.getByRole('button', { name: 'Start server' })).toBeVisible();
  });
});
