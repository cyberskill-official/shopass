export const lazadaSelectors = {
  embeddedState: 'script:has-text("window.pageData"), script:has-text("__moduleData__")',
  priceText: '.pdp-price--current, [data-price]',
  listPriceText: '.pdp-price--deleted, .origin-block .price',
  flashBadge: '.pdp-mod-flash-sale, [data-spm="flashsale"]',
  readyAnchor: '.pdp-price--current, [data-price]', // chờ giá render xong
};
