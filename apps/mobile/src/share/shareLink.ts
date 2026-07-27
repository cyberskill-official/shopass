import type { HttpClient } from "../api/httpClient";

/** Build share URL with product_id + ref only — no tokens/PII (DEC-MOBILE-25). */
export function buildShareURL(base: string, productId: number, ref: string): string {
  const root = base.endsWith("/") ? base.slice(0, -1) : base;
  const u = new URL(`${root}/p`);
  u.searchParams.set("product_id", String(productId));
  if (ref) u.searchParams.set("ref", ref);
  return u.toString();
}

export function assertSafeShareURL(url: string): void {
  const low = url.toLowerCase();
  if (low.includes("token") || low.includes("password") || low.includes("email=")) {
    throw new Error("share_url_contains_pii");
  }
}

/** Fetch referral_code from BILL-004 — mobile MUST NOT invent codes (DEC-MOBILE-21). */
export async function fetchMyReferralCode(http: HttpClient): Promise<string> {
  const res = await http.request("/v1/referral/me");
  if (!res.ok) throw new Error("referral_fetch_failed");
  const body = (await res.json()) as { code?: string; referral_code?: string };
  const code = body.code ?? body.referral_code;
  if (!code) throw new Error("referral_missing");
  return code;
}

export async function buildUserShareLink(
  http: HttpClient,
  base: string,
  productId: number,
): Promise<string> {
  const ref = await fetchMyReferralCode(http);
  const url = buildShareURL(base, productId, ref);
  assertSafeShareURL(url);
  return url;
}
