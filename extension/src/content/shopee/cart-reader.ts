import { CartItem, Message } from "../../shared/types";
import { fetchCartViaApi } from "./api-client";
import { readVouchersFromDom } from "./voucher-reader";
import { CART_SELECTORS, PRODUCT_ID_SELECTORS, PRICE_SELECTORS, QTY_SELECTORS } from "./dom-selectors";

// Allow testing to set a mock reporter
let healthReporter = (msg: any) => {
  if (chrome && chrome.runtime && chrome.runtime.sendMessage) {
    chrome.runtime.sendMessage(msg).catch(() => {});
  }
};

export function setHealthReporter(fn: (msg: any) => void) {
  healthReporter = fn;
}

export function readCartFromDom(): CartItem[] | null {
  const items: CartItem[] = [];
  
  let elements: NodeListOf<Element> | null = null;
  for (const sel of CART_SELECTORS) {
    elements = document.querySelectorAll(sel);
    if (elements.length > 0) break;
  }
  
  if (!elements || elements.length === 0) return null;

  elements.forEach(el => {
    let productId = "";
    for (const sel of PRODUCT_ID_SELECTORS) {
      const node = el.querySelector(sel);
      if (node) {
        productId = (node as HTMLInputElement).value || (node as HTMLElement).dataset.itemid || node.textContent?.trim() || "";
        if (productId) break;
      }
    }

    let price = 0;
    for (const sel of PRICE_SELECTORS) {
      const node = el.querySelector(sel);
      if (node && node.textContent) {
        const match = node.textContent.match(/[\d\.]+/);
        if (match) {
          price = parseInt(match[0].replace(/\./g, ""), 10);
          break;
        }
      }
    }

    let qty = 0;
    for (const sel of QTY_SELECTORS) {
      const node = el.querySelector(sel);
      if (node) {
        qty = parseInt((node as HTMLInputElement).value, 10);
        if (!isNaN(qty)) break;
      }
    }

    if (productId) {
      items.push({ productId, price, qty });
    }
  });

  return items;
}

export async function readCart(): Promise<Message> {
  let items = await fetchCartViaApi();
  let source: "api" | "dom" = "api";
  
  if (items === null) {
    items = readCartFromDom();
    source = "dom";
  }
  
  if (items === null) {
    healthReporter({ type: "PARSER_HEALTH", platform: "shopee", broke: "cart", source });
    items = [];
  }
  
  const vouchers = readVouchersFromDom();
  return { type: "CART_READ", platform: "shopee", items, vouchers };
}
