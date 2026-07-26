import { hashDeviceSignals } from "../src/fingerprint/hash";

describe("fingerprint hash", () => {
  it("produces stable hex digest without embedding raw UA", async () => {
    const h = await hashDeviceSignals(["en-US", "1440x900", "Asia/Ho_Chi_Minh"]);
    expect(h).toMatch(/^[0-9a-f]{64}$/);
    expect(h.includes("Mozilla")).toBe(false);
    expect(h.includes("Asia")).toBe(false);
  });
});
