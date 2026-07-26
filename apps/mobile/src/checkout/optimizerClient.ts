import type { HttpClient } from "../api/httpClient";

export interface CartLine {
  product_id: number;
  qty: number;
  unit_price: number;
}

export interface OptimizeResult {
  suggestions: Array<{ code: string; save_vnd: number }>;
  total_save_vnd: number;
  message?: string;
  auto_apply?: boolean;
}

/**
 * User-initiated cart optimize only (DEC-MOBILE-11/12/15).
 * Payload is minimized: product_id/qty/unit_price — never cookies/tokens.
 */
export function buildOptimizePayload(lines: CartLine[], vouchers: string[] = []) {
  return {
    items: lines.map((l) => ({
      product_id: l.product_id,
      qty: l.qty,
      unit_price: l.unit_price,
    })),
    vouchers,
  };
}

export function assertMinimizedPayload(payload: ReturnType<typeof buildOptimizePayload>): void {
  const raw = JSON.stringify(payload).toLowerCase();
  if (raw.includes("cookie") || raw.includes("token") || raw.includes("session")) {
    throw new Error("cart_payload_not_minimized");
  }
}

export async function optimizeCart(
  http: HttpClient,
  lines: CartLine[],
  vouchers: string[] = [],
): Promise<OptimizeResult> {
  const payload = buildOptimizePayload(lines, vouchers);
  assertMinimizedPayload(payload);
  const res = await http.request("/v1/cart/optimize", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error("optimize_failed");
  const body = (await res.json()) as OptimizeResult;
  // Mobile never auto-applies (DEC-MOBILE-13).
  return { ...body, auto_apply: false };
}

export function displayOptimizeResult(r: OptimizeResult): string {
  if (!r.suggestions?.length) return "không có voucher áp dụng";
  return `Tiết kiệm ~${r.total_save_vnd}đ — tự áp mã trong app sàn`;
}
