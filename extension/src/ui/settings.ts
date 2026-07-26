import { getConsent, setConsent } from "../consent/consent-store";
import { DSAR_URL } from "../shared/config";
import type { ConsentPurpose } from "../shared/types";

const TOGGLES: { id: string; purpose: ConsentPurpose }[] = [
  { id: "set-read-cart", purpose: "read_cart" },
  { id: "set-read-voucher", purpose: "read_voucher" },
  { id: "set-sync-backend", purpose: "sync_backend" },
];

async function load() {
  const rec = await getConsent();
  for (const t of TOGGLES) {
    const el = document.getElementById(t.id) as HTMLInputElement;
    if (el) el.checked = rec.granted.includes(t.purpose);
  }
}

async function save() {
  const granted: ConsentPurpose[] = TOGGLES.filter(
    (t) => (document.getElementById(t.id) as HTMLInputElement)?.checked
  ).map((t) => t.purpose);
  await setConsent(granted);
  // §1 #7: rút consent có hiệu lực ngay — gate re-reads state
  const status = document.getElementById("status");
  if (status) {
    status.style.display = "block";
    setTimeout(() => (status.style.display = "none"), 2000);
  }
}

// Wire up toggles
for (const t of TOGGLES) {
  document.getElementById(t.id)?.addEventListener("change", save);
}

// DSAR link (TASK-COMPLY-003)
document.getElementById("link-dsar")?.addEventListener("click", (e) => {
  e.preventDefault();
  chrome.tabs.create({ url: DSAR_URL });
});

load();
