/**
 * Cart sync live path: CART_READ → minimize → enqueue → flush.
 * PIPELINE_DONE is intentionally unused (nothing ever sent it).
 */
import * as fs from "fs";
import * as path from "path";
import { fakeChromeStorage } from "./helpers";

const mockEnqueue = jest.fn(async (..._args: unknown[]) => undefined);
const mockFlushQueue = jest.fn(async (..._args: unknown[]) => undefined);
const mockOpenRealtime = jest.fn((..._args: unknown[]) => undefined);
const mockCloseRealtime = jest.fn((..._args: unknown[]) => undefined);

jest.mock("../src/sync/queue", () => ({
  enqueue: (...args: unknown[]) => mockEnqueue(...args),
}));
jest.mock("../src/sync/sender", () => ({
  flushQueue: (...args: unknown[]) => mockFlushQueue(...args),
}));
jest.mock("../src/sync/ws-client", () => ({
  openRealtime: (...args: unknown[]) => mockOpenRealtime(...args),
  closeRealtime: (...args: unknown[]) => mockCloseRealtime(...args),
}));

import { onMessage, isTrustedCartSender } from "../src/shared/messaging";
import { setConsent } from "../src/consent/consent-store";

function mockChrome() {
  (globalThis as any).chrome = {
    runtime: { id: "sandeal-ext" },
    storage: fakeChromeStorage(),
  };
}

function callOnMessage(
  message: unknown,
  sender: chrome.runtime.MessageSender
): Promise<any> {
  return new Promise((resolve, reject) => {
    let settled = false;
    const sendResponse = (r: any) => {
      settled = true;
      resolve(r);
    };
    const async = onMessage(message, sender, sendResponse);
    if (!async && !settled) {
      resolve(undefined);
    }
    if (async) {
      setTimeout(() => {
        if (!settled) reject(new Error("onMessage timed out"));
      }, 2000);
    }
  });
}

const shopSender = {
  id: "sandeal-ext",
  tab: { url: "https://shopee.vn/cart" } as chrome.tabs.Tab,
};

describe("CART_READ sync pipeline", () => {
  beforeEach(async () => {
    mockChrome();
    mockEnqueue.mockClear();
    mockFlushQueue.mockClear();
    await setConsent([]);
  });

  test("isTrustedCartSender accepts allowlisted shop tabs", () => {
    expect(
      isTrustedCartSender({
        id: "sandeal-ext",
        tab: { url: "https://shopee.vn/cart" } as chrome.tabs.Tab,
      })
    ).toBe(true);
    expect(
      isTrustedCartSender({
        id: "sandeal-ext",
        tab: { url: "https://shop.tiktok.com/cart" } as chrome.tabs.Tab,
      })
    ).toBe(true);
    expect(
      isTrustedCartSender({
        id: "sandeal-ext",
        tab: { url: "https://www.lazada.vn/cart" } as chrome.tabs.Tab,
      })
    ).toBe(true);
  });

  test("isTrustedCartSender rejects non-shop / missing tab", () => {
    expect(isTrustedCartSender({ id: "sandeal-ext" })).toBe(false);
    expect(
      isTrustedCartSender({
        id: "sandeal-ext",
        tab: { url: "https://evil.example/cart" } as chrome.tabs.Tab,
      })
    ).toBe(false);
    expect(
      isTrustedCartSender({
        id: "other-ext",
        tab: { url: "https://shopee.vn/cart" } as chrome.tabs.Tab,
      })
    ).toBe(false);
  });

  test("CART_READ without consent is rejected", async () => {
    await setConsent(["read_cart"]); // no sync_backend
    const res = await callOnMessage(
      {
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "90112", price: 89000, qty: 1 }],
        vouchers: [],
      },
      shopSender
    );
    expect(res).toEqual({ success: false, reason: "no_consent" });
    expect(mockEnqueue).not.toHaveBeenCalled();
  });

  test("CART_READ from untrusted sender is rejected", async () => {
    await setConsent(["sync_backend"]);
    const res = await callOnMessage(
      {
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "90112", price: 89000, qty: 1 }],
        vouchers: [],
      },
      { id: "sandeal-ext", tab: { url: "https://evil.example/" } as chrome.tabs.Tab }
    );
    expect(res).toEqual({ success: false, reason: "untrusted_sender" });
    expect(mockEnqueue).not.toHaveBeenCalled();
  });

  test("CART_READ → minimize → enqueue → flush (live path; no PIPELINE_DONE)", async () => {
    await setConsent(["sync_backend"]);
    const res = await callOnMessage(
      {
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "90112", price: 89000, qty: 1 }],
        vouchers: [{ code: "FREESHIP" }],
      },
      shopSender
    );
    expect(res).toEqual({ success: true });
    expect(mockEnqueue).toHaveBeenCalledTimes(1);
    const env = mockEnqueue.mock.calls[0]![0] as unknown as {
      payload: unknown;
      clientTs: number;
    };
    expect(env.payload).toEqual({
      platform: "shopee",
      items: [{ productId: "90112", price: 89000, qty: 1 }],
      vouchers: [{ code: "FREESHIP" }],
    });
    expect(typeof env.clientTs).toBe("number");
    expect(mockFlushQueue).toHaveBeenCalledTimes(1);
  });

  test("CART_READ with invalid payload is rejected by minimize", async () => {
    await setConsent(["sync_backend"]);
    const res = await callOnMessage(
      {
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "1", price: -100, qty: 1 }],
        vouchers: [],
      },
      shopSender
    );
    expect(res).toEqual({ success: false, reason: "minimize_rejected" });
    expect(mockEnqueue).not.toHaveBeenCalled();
  });

  test("service-worker source no longer listens for PIPELINE_DONE", () => {
    const sw = fs.readFileSync(
      path.join(__dirname, "..", "src", "background", "service-worker.ts"),
      "utf8"
    );
    expect(sw).not.toMatch(/PIPELINE_DONE/);
    expect(sw).toMatch(/onMessage/);
  });
});
