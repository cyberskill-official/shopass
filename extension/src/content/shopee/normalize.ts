import { CartItem } from "../../shared/types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function mapCart(json: any): CartItem[] {
  const items: CartItem[] = [];
  if (json?.data?.shop_orders) {
    for (const shop of json.data.shop_orders) {
      if (shop.items) {
        for (const item of shop.items) {
          const productId = String(item.itemid || item.item_id || "");
          const price = item.price ? Number(item.price) / 100000 : 0; // Shopee uses price multiplied by 100000
          const qty = item.amount || 0;
          if (productId) {
            items.push({ productId, price, qty });
          }
        }
      }
    }
  }
  return items;
}
