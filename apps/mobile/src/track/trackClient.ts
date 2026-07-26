import type { HttpClient } from "../api/httpClient";

export interface WishlistItem {
  product_id: number;
  title?: string;
  verdict?: string;
}

export interface PriceHistoryPoint {
  day: string;
  close_p: number;
}

export interface AlertRule {
  id: number;
  product_id: number;
  kind: string;
}

/** Thin client — never compute sale verdicts locally (DEC-MOBILE-10). */
export async function listWishlist(http: HttpClient): Promise<WishlistItem[]> {
  const res = await http.request("/v1/wishlists");
  if (!res.ok) throw new Error("wishlist_failed");
  return (await res.json()) as WishlistItem[];
}

export async function getPriceHistory(
  http: HttpClient,
  productId: number,
  range = "90d",
): Promise<PriceHistoryPoint[]> {
  const res = await http.request(`/v1/products/${productId}/price-history?range=${range}`);
  if (!res.ok) throw new Error("price_history_failed");
  const body = (await res.json()) as { points?: PriceHistoryPoint[] } | PriceHistoryPoint[];
  return Array.isArray(body) ? body : (body.points ?? []);
}

export async function listAlertRules(http: HttpClient): Promise<AlertRule[]> {
  const res = await http.request("/v1/alerts");
  if (!res.ok) throw new Error("alerts_failed");
  return (await res.json()) as AlertRule[];
}

/** @deprecated prefer getPriceHistory — kept for chart alias. */
export async function getChart(http: HttpClient, productId: number): Promise<unknown> {
  return getPriceHistory(http, productId);
}

export function displayVerdict(verdict: string | undefined): string {
  if (!verdict || verdict === "UNKNOWN") return "chưa đủ dữ liệu";
  return verdict;
}
