import { expect, test } from './fixtures';

test.describe('the unlocks screen', () => {
  test('groups by kind and shows a buff held twice as a level', async ({ page }) => {
    await page.goto('/unlocks');
    await expect(page.getByText('Scattergun damage')).toBeVisible();
    await expect(page.getByText('level 2')).toBeVisible();
  });

  test('filters by kind and by name', async ({ page }) => {
    await page.goto('/unlocks');
    await page.getByRole('button', { name: 'Class', exact: true }).click();
    await expect(page.getByText('Scattergun damage')).toHaveCount(0);
    await expect(page.getByText('Soldier', { exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Class', exact: true }).click();
    await page.getByRole('searchbox', { name: 'Find an unlock' }).fill('nothing');
    await expect(page.getByText('No unlock matches that.')).toBeVisible();
  });
});

test.describe('the bots screen', () => {
  test('shows every seat RED holds, not just the named ones', async ({ page }) => {
    await page.goto('/bots');
    // Valve tunes every wave for six defenders, so a lineup that stopped at the
    // named seats would say RED holds fewer than it does.
    await expect(page.getByRole('row')).toHaveCount(7);
  });

  test('the switcher will not leave RED with nothing to draw from', async ({ page }) => {
    await page.goto('/bots');
    for (const name of [
      'scout',
      'soldier',
      'pyro',
      'demoman',
      'heavyweapons',
      'engineer',
      'medic',
      'sniper',
    ]) {
      await page.getByRole('button', { name, exact: true }).click();
    }
    // The ninth press is refused: a lineup the mod cannot draw from leaves the
    // seats empty.
    await page.getByRole('button', { name: 'spy', exact: true }).click();
    await expect(page.getByRole('button', { name: 'spy', exact: true })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });

  test('applying needs a running server', async ({ page }) => {
    await page.goto('/bots');
    await expect(page.getByRole('button', { name: 'Apply between waves' })).toBeDisabled();
    await expect(page.getByText('The server is not up.')).toBeVisible();
  });
});
