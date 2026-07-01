/** @jest-environment jsdom */
import { readCart } from "../src/content/shopee/cart-reader";

beforeEach(() => {
  jest.resetAllMocks();
});

test("đọc giỏ từ internal JSON khi 200", async () => {
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    json: async () => ({
      data: {
        shop_orders: [
          {
            items: [
              { itemid: "90112", price: 8900000000, amount: 1 },
              { itemid: "90113", price: 10000000000, amount: 2 },
              { itemid: "90114", price: 5000000000, amount: 3 },
            ]
          }
        ]
      }
    })
  });
  
  const msg = await readCart();
  if (msg.type === "CART_READ") {
    expect(msg.items).toHaveLength(3);
    expect(msg.items[0]).toMatchObject({ productId: "90112", price: 89000, qty: 1 });
  } else {
    fail("Expected CART_READ message");
  }
});

test("fallback DOM khi JSON non-200", async () => {
  global.fetch = jest.fn().mockResolvedValue({
    ok: false
  });
  
  document.body.innerHTML = `
    <div class="cart-item">
      <input name="product_id" value="12345" />
      <div class="cart-item-price">₫150.000</div>
      <input class="shopee-sort-bar__input" value="2" />
    </div>
  `;
  
  const msg = await readCart();
  if (msg.type === "CART_READ") {
    expect(msg.items.length).toBeGreaterThan(0);
    expect(msg.items[0]).toMatchObject({ productId: "12345", price: 150000, qty: 2 });
  } else {
    fail("Expected CART_READ message");
  }
});
