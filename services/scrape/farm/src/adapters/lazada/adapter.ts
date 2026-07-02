import { BrowserContext, Page } from 'playwright';
import { ScrapeJob, PriceSnapshot, detectChallenge, ChallengedError } from '../../render';
import { humanize } from '../../behavior';
import { lazadaSelectors } from './selectors';
import { extractLazada } from './extract';
import { PageView } from '../types';

export const PLATFORM_LAZADA = 3; 

// Detect Akamai specific challenge (if different from generic)
async function detectAkamaiChallenge(page: Page): Promise<boolean> {
  const generic = await detectChallenge(page);
  if (generic) return true;

  const content = await page.content();
  const lower = content.toLowerCase();
  
  if (lower.includes('akamai') && (lower.includes('access denied') || lower.includes('you don\'t have permission'))) {
    return true;
  }
  return false;
}

export class LazadaAdapter {
  platformID() { return PLATFORM_LAZADA; }

  async fetch(ctx: BrowserContext, job: ScrapeJob): Promise<PriceSnapshot> {
    const page = await ctx.newPage();
    try {
      await page.goto(job.url, { waitUntil: 'domcontentloaded' });
      if (await detectAkamaiChallenge(page)) {
        throw new ChallengedError('akamai sensor/verify');
      }
      
      await humanize(page);
      
      try {
        await page.waitForSelector(lazadaSelectors.readyAnchor, { timeout: 10000 });
      } catch (err) {
      }

      const stateStr = await page.evaluate(() => {
        const scripts = Array.from(document.querySelectorAll('script'));
        for (const s of scripts) {
          const html = s.innerHTML;
          if (html.includes('window.pageData') || html.includes('__moduleData__') || html.includes('pdpTrackingData')) {
            return html;
          }
        }
        return null;
      });

      const priceText = await page.evaluate((sel) => {
        const el = document.querySelector<HTMLElement>(sel);
        return el ? el.innerText : null;
      }, lazadaSelectors.priceText);

      const listPriceText = await page.evaluate((sel) => {
        const el = document.querySelector<HTMLElement>(sel);
        return el ? el.innerText : null;
      }, lazadaSelectors.listPriceText);

      const flashExists = await page.evaluate((sel) => {
        return !!document.querySelector(sel);
      }, lazadaSelectors.flashBadge);

      const liveView: PageView = {
        embeddedJSON: (sel: string) => stateStr,
        text: (sel: string) => {
          if (sel === lazadaSelectors.priceText) return priceText;
          if (sel === lazadaSelectors.listPriceText) return listPriceText;
          return null;
        },
        exists: (sel: string) => {
          if (sel === lazadaSelectors.flashBadge) return flashExists;
          return false;
        }
      };

      const raw = extractLazada(liveView);

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
