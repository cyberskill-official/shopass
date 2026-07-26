import type { HttpClient } from "../api/httpClient";

export interface CartLine {
  product_id: number;
  qty: number;
  unit_price: number;
}

/** Explicit user-tap only — never auto-run in background (MOBILE-002). */
export async function optimizeCart(http: HttpClient, lines: CartLine[]): Promise<unknown> {
  const res = await http.request("/v1/cart/optimize", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ items: lines }),
  });
  if (!res.ok) throw new Error("optimize_failed");
  return res.json();
}
