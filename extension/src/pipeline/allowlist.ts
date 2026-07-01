/**
 * allowlist.ts — ALLOWLIST (whitelist trường) per DEC-EXT-14.
 * CHỈ trường được liệt kê tường minh mới đi tiếp; mọi trường khác bị loại.
 * KHÔNG dùng denylist ("gửi mọi thứ trừ...") — mặc định an toàn.
 */
import type { OutboundItem, OutboundVoucher } from "./schema";

export const ALLOWED_ITEM_FIELDS = ["productId", "price", "qty"] as const;
export const ALLOWED_VOUCHER_FIELDS = ["code", "minSpend", "discountText"] as const;

/**
 * pickItem — chỉ lấy trường trong ALLOWED_ITEM_FIELDS, loại mọi thứ khác.
 */
export function pickItem(raw: Record<string, unknown>): OutboundItem {
  return pick(raw, ALLOWED_ITEM_FIELDS) as unknown as OutboundItem;
}

/**
 * pickVoucher — chỉ lấy trường trong ALLOWED_VOUCHER_FIELDS.
 */
export function pickVoucher(raw: Record<string, unknown>): OutboundVoucher {
  return pick(raw, ALLOWED_VOUCHER_FIELDS) as unknown as OutboundVoucher;
}

function pick<T extends string>(
  o: Record<string, unknown>,
  keys: readonly T[]
): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const k of keys) {
    if (k in o) out[k] = o[k];
  }
  return out;
}
