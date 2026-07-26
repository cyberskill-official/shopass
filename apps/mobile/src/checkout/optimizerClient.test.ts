import {
  assertMinimizedPayload,
  buildOptimizePayload,
  displayOptimizeResult,
  optimizeCart,
} from "./optimizerClient";
import { HttpClient } from "../api/httpClient";
import { AuthClient } from "../auth/authClient";
import { MemoryAccessToken, MemorySecureStore } from "../auth/tokenStore";

describe("optimizerClient", () => {
  it("builds minimized payload without cookies/tokens", () => {
    const payload = buildOptimizePayload([{ product_id: 1, qty: 2, unit_price: 1000 }]);
    expect(() => assertMinimizedPayload(payload)).not.toThrow();
    expect(JSON.stringify(payload)).not.toMatch(/cookie|token|session/i);
  });

  it("never marks auto_apply true", async () => {
    const fetchImpl = (async () =>
      new Response(
        JSON.stringify({
          suggestions: [{ code: "SALE10", save_vnd: 10000 }],
          total_save_vnd: 10000,
          auto_apply: true,
        }),
        { status: 200 },
      )) as typeof fetch;
    const access = new MemoryAccessToken();
    access.set("t");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    const r = await optimizeCart(http, [{ product_id: 1, qty: 1, unit_price: 50000 }]);
    expect(r.auto_apply).toBe(false);
    expect(displayOptimizeResult(r)).toContain("tự áp mã");
  });

  it("empty suggestions message", () => {
    expect(displayOptimizeResult({ suggestions: [], total_save_vnd: 0 })).toBe(
      "không có voucher áp dụng",
    );
  });
});
