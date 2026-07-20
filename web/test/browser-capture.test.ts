import {
  buildShopeeCaptureBookmarklet,
  canonicalShopeeProductURL,
  normalizeCapturedPrice,
} from "../lib/browser-capture";

describe("browser-assisted Shopee capture", () => {
  it("canonicalizes only valid Shopee VN product links", () => {
    expect(canonicalShopeeProductURL("https://shopee.vn/Ao-x-i.28863989.57657079476?sp_atk=private#review")).toBe(
      "https://shopee.vn/Ao-x-i.28863989.57657079476"
    );
    expect(canonicalShopeeProductURL("https://www.shopee.vn/Ao-x-i.28863989.57657079476")).toBe(
      "https://shopee.vn/Ao-x-i.28863989.57657079476"
    );
    expect(canonicalShopeeProductURL("https://example.com/x-i.28863989.57657079476")).toBeNull();
    expect(canonicalShopeeProductURL("https://shopee.vn/not-a-product")).toBeNull();
    expect(canonicalShopeeProductURL("http://shopee.vn/x-i.28863989.57657079476")).toBeNull();
    expect(canonicalShopeeProductURL("https://user:password@shopee.vn/x-i.28863989.57657079476")).toBeNull();
    expect(canonicalShopeeProductURL("https://shopee.vn:8443/x-i.28863989.57657079476")).toBeNull();
    expect(canonicalShopeeProductURL("https://shopee.vn/x-i.28863989.57657079476-extra")).toBeNull();
    expect(canonicalShopeeProductURL(`https://shopee.vn/x-i.28863989.57657079476${"x".repeat(4096)}`)).toBeNull();
  });

  it("normalizes a displayed Vietnamese đồng amount without accepting invalid values", () => {
    expect(normalizeCapturedPrice("6.490.000đ")).toBe(6_490_000);
    expect(normalizeCapturedPrice(" 59,000 ")).toBe(59_000);
    expect(normalizeCapturedPrice("0")).toBeNull();
    expect(normalizeCapturedPrice("price unknown")).toBeNull();
    expect(normalizeCapturedPrice("1".repeat(65))).toBeNull();
  });

  it("emits a local-only bookmarklet with no cookie or network access", () => {
    const bookmarklet = buildShopeeCaptureBookmarklet("https://shopass.cyberskill.world");
    expect(bookmarklet).toContain("javascript:");
    expect(bookmarklet).toContain("https://shopass.cyberskill.world/capture");
    expect(bookmarklet).toContain('searchParams.set("url"');
    expect(bookmarklet).toContain('searchParams.set("price"');
    expect(bookmarklet).toContain('a.referrerPolicy="no-referrer"');
    expect(bookmarklet).toContain('a.rel="noreferrer noopener"');
    expect(bookmarklet).not.toContain("location.assign");
    expect(bookmarklet).not.toMatch(/document\.cookie|fetch\(|XMLHttpRequest|WebSocket|sendBeacon|localStorage|sessionStorage/i);
  });

  it("does not permit an arbitrary configured bookmarklet destination", () => {
    const bookmarklet = buildShopeeCaptureBookmarklet("https://evil.example/capture?x=1");
    expect(bookmarklet).toContain("https://shopass.cyberskill.world/capture");
    expect(bookmarklet).not.toContain("evil.example");
  });
});
