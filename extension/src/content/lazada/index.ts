import { readLazadaCart } from "./cart-reader";

async function main() {
  const msg = await readLazadaCart();
  if (msg.items.length > 0) {
    chrome.runtime.sendMessage(msg).catch(() => {});
  }
}

main();
