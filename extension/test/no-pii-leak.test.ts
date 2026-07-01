import { minimize } from "../src/pipeline/minimize";

describe("no PII/credential leak (DEC-EXT-13)", () => {
  test("trường thừa (cookie/email/userId) bị loại khỏi OutboundPayload", () => {
    const msg: any = {
      type: "CART_READ",
      platform: "shopee",
      items: [
        {
          productId: "90112",
          price: 89000,
          qty: 1,
          cookie: "SPC_=abc",
          userId: 7,
          email: "a@b.vn",
        },
      ],
      vouchers: [],
    };
    const out = minimize(msg)!;
    const flat = JSON.stringify(out).toLowerCase();
    expect(flat).not.toMatch(/cookie|userid|email|@/);
    expect(out.items[0]).toEqual({ productId: "90112", price: 89000, qty: 1 });
  });

  test("credential nhồi vào productId bị từ chối (fail-closed)", () => {
    const msg: any = {
      type: "CART_READ",
      platform: "shopee",
      items: [
        {
          productId: "SPC_SESSION_eyJhbGciOi...verylong",
          price: 1,
          qty: 1,
        },
      ],
      vouchers: [],
    };
    // SPC_SESSION matches COOKIE_LIKE pattern → item filtered out → empty items → null or empty
    const out = minimize(msg);
    // Either null (fail-closed) or items empty
    if (out !== null) {
      expect(out.items.length).toBe(0);
    }
  });

  test("email nhồi vào productId bị loại", () => {
    const msg: any = {
      type: "CART_READ",
      platform: "shopee",
      items: [{ productId: "user@example.com", price: 100, qty: 1 }],
      vouchers: [],
    };
    const out = minimize(msg);
    // productId fails ID_RE (contains @) → validatePayload fails → null
    // OR looksLikeCredential filters it → empty items
    if (out !== null) {
      expect(out.items.length).toBe(0);
    }
  });

  test("long token nhồi vào productId bị loại", () => {
    const msg: any = {
      type: "CART_READ",
      platform: "shopee",
      items: [
        {
          productId: "a".repeat(65), // exceeds ID_RE max 64
          price: 100,
          qty: 1,
        },
      ],
      vouchers: [],
    };
    const out = minimize(msg);
    // LONG_TOKEN pattern filters the item → empty items list (still valid payload)
    if (out !== null) {
      expect(out.items.length).toBe(0);
    }
  });
});
