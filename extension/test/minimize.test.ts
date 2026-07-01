import { minimize, metrics } from "../src/pipeline/minimize";

beforeEach(() => metrics.reset());

describe("minimize pipeline", () => {
  test("happy path đúng tập tối thiểu", () => {
    const out = minimize({
      type: "CART_READ",
      platform: "shopee",
      items: [{ productId: "90112", price: 89000, qty: 1 }],
      vouchers: [{ code: "FREESHIP" }],
    } as any);
    expect(out).toEqual({
      platform: "shopee",
      items: [{ productId: "90112", price: 89000, qty: 1 }],
      vouchers: [{ code: "FREESHIP" }],
    });
    expect(metrics.passed).toBe(1);
  });

  test("schema lỗi → null (fail-closed) + metric", () => {
    expect(
      minimize({
        type: "CART_READ",
        platform: "vng" as any,
        items: [],
        vouchers: [],
      } as any)
    ).toBeNull();
    expect(metrics.rejectedSchema).toBe(1);
  });

  test("price âm → null (fail-closed)", () => {
    expect(
      minimize({
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "1", price: -100, qty: 1 }],
        vouchers: [],
      } as any)
    ).toBeNull();
  });

  test("price quá lớn → null", () => {
    expect(
      minimize({
        type: "CART_READ",
        platform: "shopee",
        items: [{ productId: "1", price: 2e12, qty: 1 }],
        vouchers: [],
      } as any)
    ).toBeNull();
  });

  test("productId thiếu → null", () => {
    expect(
      minimize({
        type: "CART_READ",
        platform: "shopee",
        items: [{ price: 100, qty: 1 }],
        vouchers: [],
      } as any)
    ).toBeNull();
  });

  test("voucher minSpend + discountText được giữ", () => {
    const out = minimize({
      type: "CART_READ",
      platform: "shopee",
      items: [{ productId: "1", price: 100, qty: 1 }],
      vouchers: [{ code: "FREESHIPXTRA", minSpend: 0, discountText: "đến 15k" }],
    } as any);
    expect(out!.vouchers[0]).toEqual({
      code: "FREESHIPXTRA",
      minSpend: 0,
      discountText: "đến 15k",
    });
  });
});
