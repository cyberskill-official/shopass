export const CART_ITEM_SELECTORS = [
  ".cart-item",
  "[data-tracking='cart-item']",
  ".item-content"
];

export const PRODUCT_ID_SELECTORS = [
  "[data-item-id]",
  "a[href*='-i']" // Lazada usually has -i123456.html in URL
];

export const PRICE_SELECTORS = [
  ".current-price",
  "[data-tracking='item-price']",
  ".cart-item-price"
];

export const QTY_SELECTORS = [
  ".quantity-input",
  "input.next-input",
  ".cart-item-qty"
];

export function firstMatch(root: ParentNode, sels: string[]): Element[] {
  for (const s of sels) {
    const found = root.querySelectorAll(s);
    if (found.length) return [...Array.from(found)];
  }
  return [];
}
