import { assertSafeShareURL, buildShareURL } from "./shareLink";

describe("shareLink", () => {
  it("only carries product_id + ref", () => {
    const url = buildShareURL("https://shopass.app", 9, "ABC123");
    expect(url).toContain("product_id=9");
    expect(url).toContain("ref=ABC123");
    expect(url.toLowerCase()).not.toContain("token");
    expect(() => assertSafeShareURL(url)).not.toThrow();
  });

  it("rejects token-bearing urls", () => {
    expect(() => assertSafeShareURL("https://x/p?token=abc")).toThrow(/pii/);
  });
});
