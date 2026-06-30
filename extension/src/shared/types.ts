export type CartItem = {
  productId: string;
  price: number;
  qty: number;
};

export type Message =
  | { type: "CART_READ"; platform: string; items: CartItem[] }
  | { type: "PING" }; // for testing
