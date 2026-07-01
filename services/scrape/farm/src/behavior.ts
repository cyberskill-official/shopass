import { Page } from 'playwright';

export async function humanize(page: Page): Promise<void> {
  // Wait a random amount of time (dwell)
  const dwellTime = Math.floor(Math.random() * 500) + 500; // 500ms - 1000ms
  await page.waitForTimeout(dwellTime);

  // Mouse path: Start from a random position
  const startX = Math.floor(Math.random() * 800) + 100;
  const startY = Math.floor(Math.random() * 600) + 100;
  await page.mouse.move(startX, startY);

  // Move to somewhere else
  const endX = Math.floor(Math.random() * 800) + 100;
  const endY = Math.floor(Math.random() * 600) + 100;
  await page.mouse.move(endX, endY, { steps: 10 });

  // Scroll down a bit
  await page.mouse.wheel(0, Math.floor(Math.random() * 500) + 200);
  
  // Wait a bit more before extracting
  await page.waitForTimeout(Math.floor(Math.random() * 300) + 200);
}
