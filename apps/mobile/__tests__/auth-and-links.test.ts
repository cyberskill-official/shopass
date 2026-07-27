import { MemoryAccessToken } from "../src/auth/tokenStore";
import { resolveScreen, APP_TABS } from "../src/app/RootNavigator";
import { buildShareURL } from "../src/share/shareLink";
import { parseDeepLink } from "../src/deeplink/linkHandler";
import { PendingReferral } from "../src/deeplink/pendingReferral";

describe("mobile scaffold integration", () => {
  it("keeps access token in memory only", () => {
    const mem = new MemoryAccessToken();
    expect(mem.get()).toBeNull();
    mem.set("abc");
    expect(mem.get()).toBe("abc");
    mem.set(null);
    expect(mem.get()).toBeNull();
  });

  it("gates navigation on session", () => {
    expect(resolveScreen(false)).toBe("Login");
    expect(resolveScreen(true)).toBe("Home");
    expect(resolveScreen(true, 42)).toBe("Product");
    expect(APP_TABS).toContain("Track");
    expect(APP_TABS).toContain("Checkout");
  });

  it("share links only carry product_id + ref", () => {
    const url = buildShareURL("https://shopass.app", 9, "ABC123");
    expect(url).toContain("product_id=9");
    expect(url).toContain("ref=ABC123");
    expect(url.toLowerCase()).not.toContain("token");
  });

  it("parses deeplinks and blocks self-referral client-side", () => {
    expect(parseDeepLink("https://shopass.app/p?product_id=3&ref=R1")).toEqual({
      productId: 3,
      ref: "R1",
    });
    const pending = new PendingReferral();
    pending.setIfEmpty("R1");
    expect(pending.consume("R1")).toBeNull();
    pending.setIfEmpty("R2");
    expect(pending.consume("ME")).toBe("R2");
  });
});
