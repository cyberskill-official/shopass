/** @jest-environment jsdom */
import { readLazadaCart, readCartFromDom } from "../src/content/lazada/cart-reader";
import { setConsent } from "../src/shared/consent";

const lazadaCartFixtureVariant = `
<div class="item-content">
  <div data-item-id="888">Variant 1</div>
  <div class="cart-item-price">30.000 ₫</div>
  <div class="cart-item-qty">3</div>
</div>
`;

describe("Lazada DOM Fallback", () => {
  beforeEach(() => {
    let storageData: any = {};
    global.chrome = {
      runtime: { sendMessage: jest.fn() },
      storage: {
        local: {
          get: jest.fn(async (key) => {
            if (typeof key === "string") return { [key]: storageData[key] };
            return storageData;
          }),
          set: jest.fn(async (data) => {
            storageData = { ...storageData, ...data };
          })
        }
      }
    } as any;
  });

  test("selector chính trượt → dùng selector dự phòng (A/B variant)", () => {
    document.body.innerHTML = lazadaCartFixtureVariant;       // class đổi
    const items = readCartFromDom();
    expect(items && items.length).toBeGreaterThan(0);
    if (items) {
      expect(items[0].productId).toBe("888");
      expect(items[0].price).toBe(30000);
      expect(items[0].qty).toBe(3);
    }
  });

  test("parse hỏng hẳn → health signal + items rỗng", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = "<div>không khớp</div>";
    const msg = await readLazadaCart();
    expect(chrome.runtime.sendMessage).toHaveBeenCalledWith(expect.objectContaining({ type: "HEALTH_SIGNAL", broke: "cart", platform: "lazada" }));
    expect(msg.items).toEqual([]);
  });
});
