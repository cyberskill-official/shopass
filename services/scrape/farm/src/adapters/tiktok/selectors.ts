export const tiktokSelectors = {
  embeddedState: 'script#__NEXT_DATA__, script#SIGI_STATE',
  priceText: '[data-e2e="product-price"]',
  listPriceText: '[data-e2e="product-origin-price"]',
  flashBadge: '[data-e2e="flash-sale-badge"], [data-e2e="live-deal"]',
  readyAnchor: '[data-e2e="product-price"]', // chờ selector này trước khi trích (SPA hydrate)
};
