import { PendingReferral } from "./pendingReferral";

describe("PendingReferral", () => {
  it("blocks self-referral client-side", () => {
    const pending = new PendingReferral();
    pending.setIfEmpty("R1");
    expect(pending.consume("R1")).toBeNull();
  });

  it("returns other refs once", () => {
    const pending = new PendingReferral();
    pending.setIfEmpty("R2");
    expect(pending.consume("ME")).toBe("R2");
    expect(pending.consume("ME")).toBeNull();
  });
});
