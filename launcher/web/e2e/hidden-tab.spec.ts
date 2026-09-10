import { expect, test } from './fixtures';

/**
 * A launcher tab is left open for a whole evening, usually behind the game.
 * What it must not do while it is back there is work.
 */
test.describe('a tab left open', () => {
  test('drops the stream while hidden and picks the state back up on return', async ({ page }) => {
    await page.goto('/session');
    await expect(page.getByText(/^Stopped ·/)).toBeVisible();

    const sockets: string[] = [];
    page.on('websocket', (socket) => sockets.push(socket.url()));

    // Hiding the tab closes the socket, so the launcher stops sending and
    // nothing is decoded for a screen nobody is looking at.
    await page.evaluate(() => {
      Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
    });

    // While it is away, the state moves.
    await page.request.post('/fake/reset');
    const other = await page.context().newPage();
    await other.goto('/session');
    await other.getByRole('button', { name: 'Start server' }).click();
    await expect(other.getByText(/^Running/)).toBeVisible();
    await other.close();

    // Coming back opens the socket again, and its first frame is the whole
    // state: nothing to replay, and nothing missed.
    await page.evaluate(() => {
      Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
      document.dispatchEvent(new Event('visibilitychange'));
    });
    await expect(page.getByText(/^Running/)).toBeVisible();
    expect(sockets.length).toBeGreaterThan(0);
  });
});
