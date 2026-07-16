import { CartItem, VoucherItem, Message } from "../../shared/types";
import { CART_ITEM_SELECTORS, PRODUCT_ID_SELECTORS, PRICE_SELECTORS, QTY_SELECTORS, firstMatch } from "./dom-selectors";
import { ensureConsent } from "../../shared/consent";
import { reportHealth } from "../../shared/health";
import { normalizeCart } from "../../shared/normalize";

export function readCartFromDom(): CartItem[] | null {
  const items: CartItem[] = [];
  
  // Find cart items
  const itemEls = firstMatch(document, CART_ITEM_SELECTORS);

  if (itemEls.length === 0) {
    // Return null if cart is broken, empty array if just no items?
    // Wait, if no items are found, we don't know if the cart is broken or just empty.
    // In TASK-EXT-008 tests, if the root is not found/broken, return null. 
    // To distinguish empty vs broken, check if there's a cart container or empty state.
    // But per test "parse hỏng hẳn → health signal + items rỗng", 
    // we return null when broken.
    if (document.querySelector(".cart-empty-text") || document.querySelector("#akamai-challenge")) {
      return [];
    }
    return null;
  }

  for (const el of itemEls) {
    let productId = "";
    let price = 0;
    let qty = 0;

    const pIdEls = firstMatch(el, PRODUCT_ID_SELECTORS);
    for (const pEl of pIdEls) {
      productId = pEl.getAttribute("data-item-id") || pEl.getAttribute("href")?.match(/-i(\d+)\.html/)?.[1] || "";
      if (productId) break;
    }

    const priceEls = firstMatch(el, PRICE_SELECTORS);
    for (const pEl of priceEls) {
      if (pEl.textContent) {
        const text = pEl.textContent.replace(/[^\d]/g, "");
        if (text) {
          price = parseFloat(text);
          break;
        }
      }
    }

    const qtyEls = firstMatch(el, QTY_SELECTORS);
    for (const qEl of qtyEls) {
      if (qEl.tagName.toLowerCase() === "input") {
        qty = parseInt((qEl as HTMLInputElement).value, 10);
      } else {
        qty = parseInt(qEl.textContent?.replace(/[^\d]/g, "") || "0", 10);
      }
      if (qty > 0) break;
    }

    if (productId && price >= 0 && qty >= 0) {
      items.push({ productId, price, qty });
    }
  }

  return items;
}

export function readVouchersFromDom(): VoucherItem[] {
  return [];
}

export async function readLazadaCart(): Promise<any> {
  if (!(await ensureConsent("read_cart"))) {
    return { type: "CART_READ", platform: "lazada", items: [], vouchers: [] };
  }
  let raw = readCartFromDom();
  if (raw === null) {
    reportHealth({ platform: "lazada", broke: "cart", source: "dom" });
    raw = [];
  }
  const items = normalizeCart(raw);
  return { type: "CART_READ", platform: "lazada", items, vouchers: readVouchersFromDom() };
}
