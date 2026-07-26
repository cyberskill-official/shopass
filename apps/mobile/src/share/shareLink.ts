/** Build share URL with product_id + ref only — no tokens/PII (MOBILE-003). */

export function buildShareURL(base: string, productId: number, ref: string): string {
  const u = new URL("/p", base.endsWith("/") ? base : base + "/");
  u.searchParams.set("product_id", String(productId));
  if (ref) u.searchParams.set("ref", ref);
  return u.toString();
}
