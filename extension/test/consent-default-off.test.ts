import { getConsent } from "../src/consent/consent-store";
import { fakeChromeStorage } from "./helpers";
import { promises as fs } from "fs";

test("cài mới: mọi consent TẮT (im lặng != đồng thuận)", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() }; // chưa ghi gì
  const rec = await getConsent();
  expect(rec.granted).toEqual([]); // KHÔNG mục nào bật
});

test("onboarding không tick sẵn mục nào", async () => {
  const html = await fs.readFile("src/ui/onboarding.html", "utf8");
  expect(html).not.toMatch(/<input[^>]*type=["']checkbox["'][^>]*checked/i); // không checked sẵn
});
