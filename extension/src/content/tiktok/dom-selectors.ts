export const CART_ROOT_SELECTORS = [
  '[data-testid="cart-container"]',
  '.cart-list-wrapper',
  '#cart-app',
];

export const CART_ITEM_SELECTORS = [
  '[data-testid="cart-item"]',
  '.cart-item',
  '[class*="item-container"]'
];

export const PRODUCT_ID_SELECTORS = [
  '[data-product-id]',
  'a[href*="/product/"]'
];

export const PRICE_SELECTORS = [
  '[data-testid="item-price"]',
  '.price-val',
  '[class*="current-price"]'
];

export const QTY_SELECTORS = [
  '[data-testid="item-qty"] input',
  '.qty-val',
  '[class*="quantity-input"]'
];
