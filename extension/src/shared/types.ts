export interface CartItem {
  productId: string;
  price: number;
  qty: number;
}

export type Message =
  | { type: "CART_READ"; platform: string; items: CartItem[] }
  | { type: "SYNC_REQUEST"; payload: Record<string, unknown> }
  | { type: "SYNC_RESPONSE"; ok: boolean };
