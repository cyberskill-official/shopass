import { setConsent } from "../consent/consent-store";
import type { ConsentPurpose } from "../shared/types";

const PURPOSES: { id: string; purpose: ConsentPurpose }[] = [
  { id: "consent-read-cart", purpose: "read_cart" },
  { id: "consent-read-voucher", purpose: "read_voucher" },
  { id: "consent-sync-backend", purpose: "sync_backend" },
];

document.getElementById("btn-save")?.addEventListener("click", async () => {
  const granted: ConsentPurpose[] = PURPOSES.filter(
    (p) => (document.getElementById(p.id) as HTMLInputElement)?.checked
  ).map((p) => p.purpose);
  await setConsent(granted);
  window.close();
});

document.getElementById("btn-skip")?.addEventListener("click", () => {
  // DEC-EXT-29: bỏ qua = granted: [] — không có consent nào
  window.close();
});
