import { BrowserContext, Page } from 'playwright';
import { ScrapeJob, PriceSnapshot, detectChallenge, ChallengedError } from '../../render';
import { humanize } from '../../behavior';
import { tiktokSelectors } from './selectors';
import { extractTikTok } from './extract';
import { PageView } from '../types';

export const PLATFORM_TIKTOK = 2; // Giả sử ID của tiktok là 2

class PlaywrightPageView implements PageView {
  constructor(private content: string, private textContents: Record<string, string>, private existence: Record<string, boolean>) {}

  embeddedJSON(selector: string): any | null {
    // Tìm script selector trong content HTML thô, giả lập ở đây vì regex qua DOM quá tốn
    // Thực tế Playwright page.evaluate() phù hợp hơn. Tuy nhiên với mục đích test và design pattern này:
    return null; // rely on real evaluate in adapter
  }
  text(selector: string): string | null {
    return this.textContents[selector] || null;
  }
  exists(selector: string): boolean {
    return !!this.existence[selector];
  }
}

export class TikTokAdapter {
  platformID() { return PLATFORM_TIKTOK; }

  async fetch(ctx: BrowserContext, job: ScrapeJob): Promise<PriceSnapshot> {
    const page = await ctx.newPage();
    try {
      await page.goto(job.url, { waitUntil: 'domcontentloaded' });
      if (await detectChallenge(page)) {
        throw new ChallengedError('tiktok verify');
      }
      
      await humanize(page);
      
      // Chờ cho đến khi phần tử giá xuất hiện để đảm bảo SPA hydrate xong
      try {
        await page.waitForSelector(tiktokSelectors.readyAnchor, { timeout: 10000 });
      } catch (err) {
        // Có thể bị ẩn hoặc layout đổi, nhưng cứ đi tiếp
      }

      // Đọc DOM state từ Playwright
      const view: PageView = {
        embeddedJSON: (sel: string) => null, // mock cho interface, sẽ replace bằng page.evaluate
        text: (sel: string) => null,
        exists: (sel: string) => false,
      };

      // Extract JSON state
      const stateStr = await page.evaluate((sel) => {
        const script = document.querySelector(sel);
        return script ? script.innerHTML : null;
      }, tiktokSelectors.embeddedState);

      // Extract text
      const priceText = await page.evaluate((sel) => {
        const el = document.querySelector<HTMLElement>(sel);
        return el ? el.innerText : null;
      }, tiktokSelectors.priceText);

      const listPriceText = await page.evaluate((sel) => {
        const el = document.querySelector<HTMLElement>(sel);
        return el ? el.innerText : null;
      }, tiktokSelectors.listPriceText);

      const flashExists = await page.evaluate((sel) => {
        return !!document.querySelector(sel);
      }, tiktokSelectors.flashBadge);

      const liveView: PageView = {
        embeddedJSON: (sel: string) => stateStr,
        text: (sel: string) => {
          if (sel === tiktokSelectors.priceText) return priceText;
          if (sel === tiktokSelectors.listPriceText) return listPriceText;
          return null;
        },
        exists: (sel: string) => {
          if (sel === tiktokSelectors.flashBadge) return flashExists;
          return false;
        }
      };

      const raw = extractTikTok(liveView);

      // (Giả lập metrics)
      // metrics.tiktokSource(raw.source);

      return {
        productId: job.productId,
        price: raw.price,
        listPrice: raw.listPrice,
        stock: null,
        flashSale: raw.flashSale,
        ts: Date.now(),
      };
    } finally {
      await page.close();
    }
  }
}
