import type { HttpClient } from "../api/httpClient";

export async function listWishlist(http: HttpClient): Promise<unknown[]> {
  const res = await http.request("/v1/wishlists");
  if (!res.ok) throw new Error("wishlist_failed");
  return (await res.json()) as unknown[];
}

export async function getChart(http: HttpClient, productId: number): Promise<unknown> {
  const res = await http.request(`/v1/products/${productId}/chart`);
  if (!res.ok) throw new Error("chart_failed");
  return res.json();
}
