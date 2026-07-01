import { readCart } from "./cart-reader";
import { ensureConsent } from "../../consent/consent-gate";

async function main() {
  if (!(await ensureConsent("read_cart"))) {
    return; // FR-EXT-006: Do not read cart if not opted in
  }

  const msg = await readCart();
  
  if (msg.type === "CART_READ" && !(await ensureConsent("read_voucher"))) {
    msg.vouchers = []; // Remove vouchers if not opted in
  }

  if (!(await ensureConsent("sync_backend"))) {
    return; // Do not send message if sync_backend is not opted in
  }

  if (chrome && chrome.runtime && chrome.runtime.sendMessage) {
    chrome.runtime.sendMessage(msg).catch(() => {});
  }
}

// Start execution
main();
