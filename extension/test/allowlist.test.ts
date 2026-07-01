import { minimize } from "../src/pipeline/minimize";

describe("allowlist filter (DEC-EXT-14)", () => {
  test("khóa lạ mới mặc nhiên bị loại (allowlist, không cần sửa code)", () => {
    const out = minimize({
      type: "CART_READ",
      platform: "shopee",
      items: [
        { productId: "1", price: 1, qty: 1, brandNewField: "x" } as any,
      ],
      vouchers: [],
    } as any)!;
    expect("brandNewField" in out.items[0]).toBe(false);
    expect(out.items[0]).toEqual({ productId: "1", price: 1, qty: 1 });
  });

  test("voucher trường thừa bị loại", () => {
    const out = minimize({
      type: "CART_READ",
      platform: "shopee",
      items: [{ productId: "1", price: 1, qty: 1 }],
      vouchers: [{ code: "FREE", internalId: 9, trackingRef: "abc" } as any],
    } as any)!;
    expect("internalId" in out.vouchers[0]).toBe(false);
    expect("trackingRef" in out.vouchers[0]).toBe(false);
    expect(out.vouchers[0].code).toBe("FREE");
  });
});
