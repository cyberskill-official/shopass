import { CartItem } from "../../shared/types";
import { mapCart } from "../../shared/normalize";

const CART_ENDPOINT = "https://shopee.vn/api/v4/cart/get";

export async function fetchCartViaApi(): Promise<CartItem[] | null> {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), 25_000);
  try {
    const res = await fetch(CART_ENDPOINT, {
      method: "GET",
      credentials: "include",
      signal: ctrl.signal
    });
    if (!res.ok) return null;
    const json = await res.json();
    return mapCart(json);
  } catch {
    return null;
  } finally {
    clearTimeout(t);
  }
}
