import { safeNextPath } from "../lib/safe-next";

const origin = "https://sandeal.example";

describe("safeNextPath", () => {
  it("keeps a normalized same-origin destination", () => {
    expect(safeNextPath("/alerts?tab=active#rules", origin)).toBe("/alerts?tab=active#rules");
    expect(safeNextPath("https://sandeal.example/wishlist?sort=price", origin)).toBe(
      "/wishlist?sort=price",
    );
  });

  it.each([
    "https://evil.example/phish",
    "//evil.example/phish",
    "/\\evil.example/phish",
    "javascript:alert(1)",
  ])("rejects an external or unsafe destination: %s", (candidate) => {
    expect(safeNextPath(candidate, origin)).toBe("/dashboard");
  });
});
