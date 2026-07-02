import { CartItem } from "../../shared/types";

// normalize.ts - chuẩn hóa ra CartItem{productId, price, qty} tối thiểu; bỏ mọi trường thừa
export function normalizeCartItem(raw: any): CartItem | null {
  if (!raw) return null;
  const productId = raw.productId || raw.itemid || raw.id;
  const price = raw.price || raw.item_price;
  const qty = raw.qty || raw.amount || raw.quantity;

  if (!productId || price == null || qty == null) return null;

  return {
    productId: String(productId),
    price: Number(price),
    qty: Number(qty)
  };
}
