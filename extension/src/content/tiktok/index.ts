import { onCartRendered } from "./spa-observer";
import { readTiktokCart } from "./cart-reader";

// Entry point for TikTok Shop content script
console.log("SănDeal: TikTok Shop content script loaded");

// Start SPA observer
const disconnect = onCartRendered(async () => {
  console.log("SănDeal: TikTok cart rendered, reading...");
  const msg = await readTiktokCart();
  if (msg.items.length > 0) {
    try {
      chrome.runtime.sendMessage(msg);
    } catch (err) {
      console.debug("SănDeal: failed to send cart message", err);
    }
  }
});

// Optionally, handle disconnect on page unload
window.addEventListener("unload", disconnect);
