import { BrowserContext, Page } from 'playwright';
import { humanize } from './behavior';

export class ChallengedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ChallengedError';
  }
}

export interface ScrapeJob {
  url: string;
  productId: string;
}

export interface PriceSelectors {
  price: string;
  listPrice?: string;
  stock?: string;
  flash?: string;
}

export interface PriceSnapshot {
  productId: string;
  price: number;
  listPrice: number | null;
  stock: number | null;
  flashSale: boolean;
  ts: number;
}

export async function detectChallenge(page: Page): Promise<boolean> {
  const content = await page.content();
  const lower = content.toLowerCase();
  
  if (lower.includes('captcha') || lower.includes('verify you are human') || lower.includes('verify_slider')) {
    return true;
  }
  
  const title = await page.title();
  if (title.toLowerCase().includes('verify')) {
    return true;
  }

  return false;
}

async function extractDom(page: Page, sel: PriceSelectors): Promise<{ priceStr: string, listPriceStr?: string, stockStr?: string, flashStr?: string }> {
  let priceStr = '';
  let listPriceStr = '';
  let stockStr = '';
  let flashStr = '';

  try {
    const el = await page.$(sel.price);
    if (el) {
      priceStr = await el.innerText() || '';
    }
    
    if (sel.listPrice) {
      const lel = await page.$(sel.listPrice);
      if (lel) listPriceStr = await lel.innerText() || '';
    }
    
    if (sel.stock) {
      const stockEl = await page.$(sel.stock);
      if (stockEl) stockStr = await stockEl.innerText() || '';
    }
    
    if (sel.flash) {
      const flashEl = await page.$(sel.flash);
      if (flashEl) flashStr = await flashEl.innerText() || '';
    }
  } catch (e) {
    // ignore elements not found
  }

  return { priceStr, listPriceStr, stockStr, flashStr };
}

function parseNumber(str: string): number | null {
  if (!str) return null;
  const cleaned = str.replace(/[^\d]/g, '');
  if (!cleaned) return null;
  return parseInt(cleaned, 10);
}

export function toSnapshot(productId: string, raw: { priceStr: string, listPriceStr?: string, stockStr?: string, flashStr?: string }): PriceSnapshot {
  const price = parseNumber(raw.priceStr) || 0;
  const listPrice = parseNumber(raw.listPriceStr || '');
  const stock = parseNumber(raw.stockStr || '');
  
  return {
    productId,
    price,
    listPrice,
    stock,
    flashSale: !!raw.flashStr,
    ts: Date.now(),
  };
}

export async function renderPrice(
  ctx: BrowserContext, job: ScrapeJob, sel: PriceSelectors,
): Promise<PriceSnapshot> {
  const page = await ctx.newPage();
  try {
    await page.goto(job.url, { waitUntil: 'domcontentloaded', timeout: 30000 });
    if (await detectChallenge(page)) {
      throw new ChallengedError('captcha/slider/verify');
    }
    
    await humanize(page);
    
    const raw = await extractDom(page, sel);
    
    // We should emit metrics here in a real scenario
    return toSnapshot(job.productId, raw);
  } finally {
    await page.close();
  }
}
