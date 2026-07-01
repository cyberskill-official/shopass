import { getConsent, setConsent, POLICY_VERSION } from "../src/consent/consent-store";
import { fakeChromeStorage } from "./helpers";

test("setConsent ghi ConsentRecord có policyVersion + decidedAt + granted", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setConsent(["read_cart"]);
  const rec = await getConsent();
  expect(rec.policyVersion).toBe(POLICY_VERSION);
  expect(rec.decidedAt).toBeGreaterThan(0);
  expect(rec.granted).toEqual(["read_cart"]);
});

test("đổi policyVersion → yêu cầu lại consent", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setConsent(["read_cart"]);
  
  // mô phỏng policy cũ
  await chrome.storage.local.set({
    "sandeal:consent": {
      policyVersion: "old-version",
      decidedAt: 123,
      granted: ["read_cart"],
    }
  });

  const rec = await getConsent();
  expect(rec.granted).toEqual([]); // reset vì version khác
});

test("consent state bền qua SW kill", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setConsent(["read_cart"]);
  
  jest.resetModules();
  const { getConsent: getConsent2 } = await import("../src/consent/consent-store");
  const rec = await getConsent2();
  expect(rec.granted).toEqual(["read_cart"]);
});
