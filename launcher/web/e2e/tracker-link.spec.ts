import { expect, test } from './fixtures';

test('opens the configured room in the public campaign tracker', async ({ page }) => {
  await page.goto('/session');
  await expect(page.getByRole('link', { name: 'Campaign tracker' })).toHaveAttribute(
    'href',
    'https://m-this.github.io/tf2-archipelago/tracker/?room=demo-room',
  );
});
