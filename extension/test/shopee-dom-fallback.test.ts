/** @jest-environment jsdom */
import { readCartFromDom, readCart, setHealthReporter } from "../src/content/shopee/cart-reader";

beforeEach(() => {
  jest.resetAllMocks();
});

test("selector chính trượt → dùng selector dự phòng (A/B variant)", () => {
  document.body.innerHTML = `
    <div class="shopee-cart-item">
      <div class="item-id-hidden">67890</div>
      <div class="shopee-cart-item__price">200.000</div>
      <div class="shopee-cart-item__quantity"><input value="1" /></div>
    </div>
  `;
  const items = readCartFromDom();
  expect(items && items.length).toBeGreaterThan(0);
  expect(items![0]).toMatchObject({ productId: "67890", price: 200000, qty: 1 });
});

test("parse hỏng hẳn → health signal + items rỗng, không ném lỗi", async () => {
  global.fetch = jest.fn().mockResolvedValue({ ok: false });
  document.body.innerHTML = "<div>không khớp gì</div>";
  
  const spy = jest.fn();
  setHealthReporter(spy);
  
  const msg = await readCart();
  expect(spy).toHaveBeenCalledWith(expect.objectContaining({ broke: "cart" }));
  if (msg.type === "CART_READ") {
    expect(msg.items).toEqual([]);
  } else {
    fail("Expected CART_READ message");
  }
});
