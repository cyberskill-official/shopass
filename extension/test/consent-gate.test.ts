import { ensureConsent } from "../src/consent/consent-gate";
import { setConsent } from "../src/consent/consent-store";
import { fakeChromeStorage } from "./helpers";

test("chưa opt-in → gate chặn đọc giỏ + đồng bộ", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  expect(await ensureConsent("read_cart")).toBe(false);
  expect(await ensureConsent("sync_backend")).toBe(false);
});

test("granular: bật read_cart không tự bật sync_backend", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setConsent(["read_cart"]);
  expect(await ensureConsent("read_cart")).toBe(true);
  expect(await ensureConsent("sync_backend")).toBe(false);  // độc lập
});

test("rút consent có hiệu lực ngay", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setConsent(["read_cart"]);
  await setConsent([]); // tắt
  expect(await ensureConsent("read_cart")).toBe(false);
});
