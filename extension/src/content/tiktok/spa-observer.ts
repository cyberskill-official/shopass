import { CART_ROOT_SELECTORS } from "./dom-selectors";

export function onCartRendered(cb: () => void): () => void {
  const obs = new MutationObserver(() => {
    if (document.querySelector(CART_ROOT_SELECTORS.find(s => document.querySelector(s)) ?? "x")) {
      cb();
    }
  });
  obs.observe(document.body, { childList: true, subtree: true });
  
  // Try to fire immediately if it's already rendered
  if (document.querySelector(CART_ROOT_SELECTORS.find(s => document.querySelector(s)) ?? "x")) {
    cb();
  }
  
  return () => obs.disconnect();
}
