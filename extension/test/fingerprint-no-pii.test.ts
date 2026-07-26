import { TextEncoder } from "node:util";
import { webcrypto } from "node:crypto";

import { hashDeviceSignals } from "../src/fingerprint/hash";

const g = globalThis as typeof globalThis & {
  TextEncoder?: typeof TextEncoder;
  crypto?: Crypto;
};
if (typeof g.TextEncoder === "undefined") {
  g.TextEncoder = TextEncoder;
}
if (typeof g.crypto === "undefined" || typeof g.crypto.subtle === "undefined") {
  Object.defineProperty(g, "crypto", { value: webcrypto, configurable: true });
}

describe("fingerprint hash", () => {
  it("produces stable hex digest without embedding raw UA", async () => {
    const h = await hashDeviceSignals(["en-US", "1440x900", "Asia/Ho_Chi_Minh"]);
    expect(h).toMatch(/^[0-9a-f]{64}$/);
    expect(h.includes("Mozilla")).toBe(false);
    expect(h.includes("Asia")).toBe(false);
  });
});
