/**
 * offscreen.ts — tài liệu offscreen nhận HTML thô, parse bằng DOM thật, trả kết quả.
 * MUST NOT tự fetch trang sàn (DEC-EXT-22).
 */
import type { ParseDomRequest, ParseDomResult } from "../shared/types";

chrome.runtime.onMessage.addListener(
  (msg: ParseDomRequest, _sender, sendResponse) => {
    if (msg.target !== "offscreen" || msg.type !== "PARSE_DOM") return;

    try {
      const result = parseDom(msg.html, msg.platform);
      sendResponse(result);
    } catch {
      sendResponse({ type: "PARSE_DOM_RESULT", items: [] } satisfies ParseDomResult);
    }
  }
);

function parseDom(html: string, platform: string): ParseDomResult {
  const parser = new DOMParser();
  const doc = parser.parseFromString(html, "text/html");

  const items: ParseDomResult["items"] = [];

  if (platform === "shopee") {
    const cartItems = doc.querySelectorAll("[data-product-id]");
    cartItems.forEach((el) => {
      const productId = el.getAttribute("data-product-id") ?? "";
      const priceEl = el.querySelector("[data-price]");
      const qtyEl = el.querySelector("[data-qty]");
      items.push({
        productId,
        price: Number(priceEl?.getAttribute("data-price") ?? 0),
        qty: Number(qtyEl?.getAttribute("data-qty") ?? 1),
      });
    });
  }

  return { type: "PARSE_DOM_RESULT", items };
}

export { parseDom };
