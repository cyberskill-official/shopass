import { CartItem, VoucherItem } from "../../shared/types";
import { CART_ITEM_SELECTORS, PRODUCT_ID_SELECTORS, PRICE_SELECTORS, QTY_SELECTORS } from "./dom-selectors";
import { ensureConsent } from "../../shared/consent";
import { reportHealth } from "../../shared/health";
import { normalizeCart } from "../../shared/normalize";

export function readCartFromDom(): CartItem[] | null {
  const items: CartItem[] = [];
  
  // Find cart items
  let itemEls = null;
  for (const sel of CART_ITEM_SELECTORS) {
    const els = document.querySelectorAll(sel);
    if (els.length > 0) {
      itemEls = els;
      break;
    }
  }

  if (!itemEls) {
    return null; // Cart root not found or no items
  }

  for (const el of itemEls) {
    let productId = "";
    let price = 0;
    let qty = 0;

    for (const sel of PRODUCT_ID_SELECTORS) {
      const pEl = el.querySelector(sel);
      if (pEl) {
        productId = pEl.getAttribute("data-product-id") || pEl.getAttribute("href")?.match(/\/product\/([^\/?]+)/)?.[1] || "";
        if (productId) break;
      }
    }

    for (const sel of PRICE_SELECTORS) {
      const pEl = el.querySelector(sel);
      if (pEl && pEl.textContent) {
        const text = pEl.textContent.replace(/[^\d]/g, "");
        if (text) {
          price = parseFloat(text);
          break;
        }
      }
    }

    for (const sel of QTY_SELECTORS) {
      const qEl = el.querySelector(sel);
      if (qEl) {
        if (qEl.tagName.toLowerCase() === "input") {
          qty = parseInt((qEl as HTMLInputElement).value, 10);
        } else {
          qty = parseInt(qEl.textContent?.replace(/[^\d]/g, "") || "0", 10);
        }
        if (qty > 0) break;
      }
    }

    if (productId && price >= 0 && qty >= 0) {
      items.push({ productId, price, qty });
    } else {
      console.log("SKIPPED ITEM:", { productId, price, qty });
    }
  }

  return items;
}

export function readVouchersFromDom(): VoucherItem[] {
  return [];
}


export async function readTiktokCart(): Promise<any> {
  if (!(await ensureConsent("read_cart"))) {
    return { type: "CART_READ", platform: "tiktok", items: [], vouchers: [] };
  }
  let raw = readCartFromDom();
  if (raw === null) {
    reportHealth({ platform: "tiktok", broke: "cart", source: "dom" });
    raw = [];
  }
  const items = normalizeCart(raw);
  return { type: "CART_READ", platform: "tiktok", items, vouchers: readVouchersFromDom() };
}
