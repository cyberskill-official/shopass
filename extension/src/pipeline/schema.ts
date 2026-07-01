export interface OutboundItem {
  productId: string;
  price: number;
  qty: number;
}

export interface OutboundVoucher {
  code: string;
  minSpend?: number;
  discountText?: string;
}

export interface OutboundPayload {
  platform: "shopee" | "tiktok" | "lazada";
  items: OutboundItem[];
  vouchers: OutboundVoucher[];
  // KHÔNG trường nào khác. Mọi mở rộng phải sửa schema + allowlist + review.
}

const VALID_PLATFORMS = ["shopee", "tiktok", "lazada"];
const ID_RE = /^[A-Za-z0-9._-]{1,64}$/;

/**
 * validatePayload — fail-closed validator (DEC-EXT-17).
 * Payload không khớp schema bị từ chối, KHÔNG gửi 'cứ gửi đại'.
 */
export function validatePayload(p: unknown): p is OutboundPayload {
  if (!p || typeof p !== "object") return false;
  const o = p as OutboundPayload;
  if (!VALID_PLATFORMS.includes(o.platform)) return false;
  if (!Array.isArray(o.items) || !Array.isArray(o.vouchers)) return false;
  return o.items.every(
    (it) =>
      typeof it.productId === "string" &&
      ID_RE.test(it.productId) &&
      Number.isInteger(it.price) &&
      it.price >= 0 &&
      it.price < 1e12 &&
      Number.isInteger(it.qty) &&
      it.qty >= 0 &&
      it.qty < 1e6
  );
}
