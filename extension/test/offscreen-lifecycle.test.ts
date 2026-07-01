import { parseDomOffscreen, OFFSCREEN_PATH } from "../src/offscreen/manager";
import type { ParseDomRequest } from "../src/shared/types";

function fakeOffscreen(calls: string[], hasDocument = false) {
  return {
    offscreen: {
      hasDocument: jest.fn(async () => {
        return hasDocument;
      }),
      createDocument: jest.fn(async (opts: any) => {
        calls.push("create");
        // Validate reason is explicit (DEC-EXT-18)
        expect(opts.reasons).toContain("DOM_SCRAPING");
        expect(opts.justification).toBeTruthy();
      }),
      closeDocument: jest.fn(async () => {
        calls.push("close");
      }),
      Reason: { DOM_SCRAPING: "DOM_SCRAPING", CLIPBOARD: "CLIPBOARD" },
    },
    runtime: {
      sendMessage: jest.fn((_msg: any, callback: any) => {
        callback({ type: "PARSE_DOM_RESULT", items: [{ productId: "1", price: 100, qty: 1 }] });
      }),
      lastError: null as any,
    },
  };
}

const makeReq = (html = "<div></div>"): ParseDomRequest => ({
  target: "offscreen",
  type: "PARSE_DOM",
  html,
  platform: "shopee",
});

describe("offscreen lifecycle", () => {
  afterEach(() => {
    (globalThis as any).chrome = undefined;
  });

  test("tạo offscreen rồi đóng ngay sau khi xong (DEC-EXT-19)", async () => {
    const calls: string[] = [];
    (globalThis as any).chrome = fakeOffscreen(calls);
    await parseDomOffscreen(makeReq());
    expect(calls).toContain("create");
    expect(calls).toContain("close"); // đóng NGAY, không để mở thường trực
  });

  test("không tạo offscreen mới khi đã có tài liệu (§1 #2)", async () => {
    const calls: string[] = [];
    (globalThis as any).chrome = fakeOffscreen(calls, /*hasDocument*/ true);
    await parseDomOffscreen(makeReq());
    expect(calls.filter((c) => c === "create")).toHaveLength(0); // tái dùng
    expect(calls).toContain("close"); // vẫn đóng sau khi xong
  });

  test("offscreen tạo với reason DOM_SCRAPING tường minh (§1 #1)", async () => {
    const calls: string[] = [];
    (globalThis as any).chrome = fakeOffscreen(calls);
    await parseDomOffscreen(makeReq());
    const mock = (globalThis as any).chrome.offscreen.createDocument;
    expect(mock).toHaveBeenCalledWith(
      expect.objectContaining({
        url: OFFSCREEN_PATH,
        reasons: expect.arrayContaining(["DOM_SCRAPING"]),
      })
    );
  });

  test("offscreen đóng ngay cả khi task lỗi (finally)", async () => {
    const calls: string[] = [];
    const fake = fakeOffscreen(calls);
    fake.runtime.sendMessage = jest.fn((_msg: any, callback: any) => {
      // Simulate chrome.runtime.lastError being set BEFORE callback
      (fake.runtime as any).lastError = { message: "test error" };
      callback(undefined);
    });
    (globalThis as any).chrome = fake;
    // Should still close even on error
    await expect(parseDomOffscreen(makeReq())).rejects.toThrow("test error");
    expect(calls).toContain("close");
  });
});
