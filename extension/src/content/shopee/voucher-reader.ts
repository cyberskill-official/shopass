import { VoucherItem } from "../../shared/types";
import { VOUCHER_SELECTORS, VOUCHER_CODE_SELECTORS, VOUCHER_MIN_SPEND_SELECTORS } from "./dom-selectors";

export function readVouchersFromDom(): VoucherItem[] {
  const vouchers: VoucherItem[] = [];
  
  let elements: NodeListOf<Element> | null = null;
  for (const sel of VOUCHER_SELECTORS) {
    elements = document.querySelectorAll(sel);
    if (elements.length > 0) break;
  }
  if (!elements) return vouchers;

  elements.forEach(el => {
    let code = "";
    for (const sel of VOUCHER_CODE_SELECTORS) {
      const node = el.querySelector(sel);
      if (node) {
        code = node.textContent?.trim() || (node as HTMLElement).dataset.voucherCode || "";
        if (code) break;
      }
    }

    let minSpend = 0;
    for (const sel of VOUCHER_MIN_SPEND_SELECTORS) {
      const node = el.querySelector(sel);
      if (node && node.textContent) {
        const text = node.textContent;
        // Basic extraction for min spend numbers in VND
        const match = text.match(/[\d\.]+[kKđĐ]/);
        if (match) {
           let val = match[0].replace(/[kKđĐ\.]/g, "");
           minSpend = parseInt(val, 10);
           if (match[0].toLowerCase().includes('k')) {
             minSpend *= 1000;
           }
        }
        break;
      }
    }

    if (code) {
      vouchers.push({ code, minSpend });
    }
  });

  return vouchers;
}
