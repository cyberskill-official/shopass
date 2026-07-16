/**
 * minimize.ts — van an toàn cuối trước khi dữ liệu rời máy client.
 * allowlist -> redact -> validate -> fail-closed (DEC-EXT-15/17).
 */
import type { OutboundPayload } from "./schema";
import { validatePayload } from "./schema";
import { pickItem, pickVoucher } from "./allowlist";
import { containsPiiOrCredential, looksLikeCredential } from "./redact";

interface CartReadMessage {
  type: "CART_READ";
  platform: "shopee" | "tiktok" | "lazada";
  items: Array<Record<string, unknown>>;
  vouchers: Array<Record<string, unknown>>;
}

// Simple metrics counter (in-memory, for TASK-TRUST-003 audit)
const metrics = {
  passed: 0,
  droppedFields: 0,
  rejectedSchema: 0,
  rejectedRedact: 0,
  reset() {
    this.passed = 0;
    this.droppedFields = 0;
    this.rejectedSchema = 0;
    this.rejectedRedact = 0;
  },
};

/**
 * minimize — nhận CartReadMessage (có thể "bẩn") và trả OutboundPayload (sạch)
 * hoặc null nếu payload không hợp lệ (fail-closed).
 *
 * 1. Allowlist filter: chỉ trường kê tên mới đi tiếp (DEC-EXT-14)
 * 2. Redact: loại item có giá trị nghi credential/PII (§1 #6)
 * 3. Validate schema: fail-closed nếu lệch (DEC-EXT-17)
 * 4. Final PII scan: từ chối cả payload nếu vẫn phát hiện PII (§1 #4)
 */
export function minimize(msg: CartReadMessage): OutboundPayload | null {
  // §1 #2: allowlist filter — chỉ trường được kê tên
  const items = msg.items
    .map((i) => pickItem(i as Record<string, unknown>))
    .filter((it) => !looksLikeCredential(it.productId)); // §1 #8: loại productId bất thường

  const vouchers = msg.vouchers
    .map((v) => pickVoucher(v as Record<string, unknown>))
    .filter((v) => !looksLikeCredential(v.code));

  const payload: OutboundPayload = {
    platform: msg.platform,
    items,
    vouchers,
  };

  // §1 #5: fail-closed nếu schema không khớp
  if (!validatePayload(payload)) {
    metrics.rejectedSchema++;
    return null;
  }

  // §1 #6: quét giá trị credential/PII lần cuối
  if (containsPiiOrCredential(payload as unknown as Record<string, unknown>)) {
    metrics.rejectedRedact++;
    return null;
  }

  metrics.passed++;
  return payload;
}

export { metrics };
